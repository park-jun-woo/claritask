package task

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/claude"
)

// defaultParallel is the default parallel execution count
const defaultParallel = 3

// getParallel reads the parallel config from project local DB
func getParallel(localDB *db.DB) int {
	var val string
	err := localDB.QueryRow("SELECT value FROM config WHERE key = 'parallel'").Scan(&val)
	if err != nil {
		return defaultParallel
	}
	n, err := strconv.Atoi(val)
	if err != nil || n < 1 {
		return defaultParallel
	}
	return n
}

// Run runs a task (2회차 순회: planned → done)
// If id is empty, runs next planned task
func Run(projectPath, id string) types.Result {
	return RunWithContext(context.Background(), projectPath, id)
}

// RunWithContext runs a task with context for cancellation support
func RunWithContext(ctx context.Context, projectPath, id string) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	var t Task

	if id == "" {
		// Get next planned leaf task
		err = localDB.QueryRow(`
			SELECT id, title, spec, plan, status FROM tasks
			WHERE status = 'planned' AND is_leaf = 1
			ORDER BY priority DESC, depth DESC, id ASC LIMIT 1
		`).Scan(&t.ID, &t.Title, &t.Spec, &t.Plan, &t.Status)
		if err == sql.ErrNoRows {
			return types.Result{
				Success: true,
				Message: "실행할 작업이 없습니다. (planned 상태 leaf 작업 없음)\n[작업 목록:task list]",
			}
		}
	} else {
		err = localDB.QueryRow(`
			SELECT id, title, spec, plan, status FROM tasks WHERE id = ?
		`, id).Scan(&t.ID, &t.Title, &t.Spec, &t.Plan, &t.Status)
		if err == sql.ErrNoRows {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("작업을 찾을 수 없습니다: #%s", id),
			}
		}
	}

	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}

	if t.Status != "planned" {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업 #%d은(는) %s 상태입니다. (planned 상태만 실행 가능)", t.ID, t.Status),
		}
	}

	if t.Plan == "" {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업 #%d에 Plan이 없습니다. 먼저 'task plan %d'를 실행하세요.", t.ID, t.ID),
		}
	}

	// Insert traversal record
	travID, travErr := insertTraversal(localDB, "run", &t.ID, "")
	if travErr != nil {
		log.Printf("[Task] traversal INSERT 실패: %v", travErr)
	}

	// Build context map (lightweight task tree summary)
	contextMap, err := BuildContextMap(localDB)
	if err != nil {
		ret := types.Result{
			Success: false,
			Message: fmt.Sprintf("Context Map 생성 실패: %v", err),
		}
		if travErr == nil {
			finishTraversal(localDB, travID, "failed", 1, 0, 1)
		}
		return ret
	}

	// Build report path
	reportPath := filepath.Join(projectPath, ".claribot", fmt.Sprintf("task-run-%d-report.md", t.ID))
	// Ensure .claribot directory exists
	if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err != nil {
		ret := types.Result{
			Success: false,
			Message: fmt.Sprintf("report 디렉토리 생성 실패: %v", err),
		}
		if travErr == nil {
			finishTraversal(localDB, travID, "failed", 1, 0, 1)
		}
		return ret
	}

	// Build prompt with report path
	prompt := BuildExecutePrompt(&t, contextMap, reportPath)

	// Run Claude Code
	opts := claude.Options{
		UserPrompt: prompt,
		WorkDir:    projectPath,
		ReportPath: reportPath,
	}

	result, err := claude.RunContext(ctx, opts)
	if err != nil {
		ret := types.Result{
			Success: false,
			Message: fmt.Sprintf("Claude 실행 오류: %v", err),
		}
		if ctx.Err() != nil {
			ret.ErrorType = "cancelled"
		}
		if travErr == nil {
			finishTraversal(localDB, travID, "failed", 1, 0, 1)
		}
		return ret
	}

	now := db.TimeNow()

	if result.ExitCode != 0 {
		// Check for authentication error
		authError := claude.IsAuthError(result)
		if authError {
			log.Printf("[Task] Run 인증 오류 감지 (task #%d)", t.ID)
		}

		// Save error and mark as failed
		if _, err := localDB.Exec(`UPDATE tasks SET error = ?, status = 'failed', updated_at = ? WHERE id = ?`, result.Output, now, t.ID); err != nil {
			log.Printf("[Task] Run 에러 저장 실패 (task #%d): %v", t.ID, err)
		}
		// Clean up report file
		if err := os.Remove(reportPath); err != nil && !os.IsNotExist(err) {
			log.Printf("[Task] Run report 파일 삭제 실패 (task #%d): %v", t.ID, err)
		}
		if travErr == nil {
			finishTraversal(localDB, travID, "failed", 1, 0, 1)
		}

		ret := types.Result{
			Success: false,
			Message: fmt.Sprintf("작업 실행 실패: %s", result.Output),
		}
		if authError {
			ret.ErrorType = "auth_error"
		}
		return ret
	}

	// Save report and update status to done
	_, err = localDB.Exec(`UPDATE tasks SET report = ?, status = 'done', updated_at = ? WHERE id = ?`, result.Output, now, t.ID)
	if err != nil {
		if travErr == nil {
			finishTraversal(localDB, travID, "failed", 1, 0, 1)
		}
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("Report 저장 실패: %v", err),
		}
	}

	// Clean up report file after DB save
	if err := os.Remove(reportPath); err != nil && !os.IsNotExist(err) {
		log.Printf("[Task] Run report 파일 삭제 실패 (task #%d): %v", t.ID, err)
	}

	// Update traversal record
	if travErr == nil {
		finishTraversal(localDB, travID, "done", 1, 1, 0)
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("✅ 작업 #%d 완료: %s\n[조회:task get %d]", t.ID, t.Title, t.ID),
		Data:    &t,
	}
}

// RunAll runs all planned tasks (2회차 순회 전체 실행)
func RunAll(projectPath string) types.Result {
	// Check if already running for this project
	if IsCycleRunning(projectPath) {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("이 프로젝트는 이미 순회 중입니다: %s", getProjectID(projectPath)),
		}
	}

	ResetCancel()
	ctx, cancel := context.WithCancel(context.Background())

	startTime := time.Now()
	SetCycleState(projectPath, CycleState{
		Running:     true,
		Type:        "run",
		StartedAt:   startTime,
		ProjectPath: projectPath,
		Phase:       "run",
	})
	SetCycleCancel(projectPath, cancel)
	defer func() {
		cancel()
		ClearCycleState(projectPath)
	}()

	// Insert traversal record
	localDB, travErr := db.OpenLocal(projectPath)
	var travID int64
	if travErr == nil {
		travID, travErr = insertTraversal(localDB, "run", nil, "")
		if travErr != nil {
			log.Printf("[Task] traversal INSERT 실패: %v", travErr)
		}
		localDB.Close()
	}

	result := runAllInternal(ctx, projectPath)

	// Update traversal record
	if travErr == nil {
		localDB, err := db.OpenLocal(projectPath)
		if err == nil {
			status := "done"
			if !result.Success {
				status = "failed"
			}
			if IsCancelled() {
				status = "cancelled"
			}
			total, success, failed := countFromMessage(result.Message)
			finishTraversal(localDB, travID, status, total, success, failed)
			localDB.Close()
		}
	}

	if globalNotifier != nil {
		notification := fmt.Sprintf("🔄 Run 순회 완료\n소요: %s\n%s", formatDuration(time.Since(startTime)), result.Message)
		globalNotifier(nil, notification)
	}

	return result
}

// runResult holds the result of a single task run for channel communication
type runResult struct {
	TaskID  int
	Title   string
	Success bool
	Message string
	IsAuth  bool
}

// runAllInternal is the internal implementation of RunAll without CycleState management.
// Used by Cycle() to avoid overwriting the cycle type.
// Supports parallel execution based on project's parallel config.
func runAllInternal(ctx context.Context, projectPath string) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}

	// Read parallel config
	parallel := getParallel(localDB)

	// Get all planned leaf tasks (priority first, then deepest)
	rows, err := localDB.Query(`
		SELECT id, title FROM tasks
		WHERE status = 'planned' AND is_leaf = 1
		ORDER BY priority DESC, depth DESC, id ASC
	`)
	if err != nil {
		localDB.Close()
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title); err != nil {
			rows.Close()
			localDB.Close()
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("스캔 실패: %v", err),
			}
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		localDB.Close()
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("행 순회 오류: %v", err),
		}
	}
	rows.Close()
	localDB.Close()

	if len(tasks) == 0 {
		return types.Result{
			Success: true,
			Message: "실행할 작업이 없습니다. (planned 상태 작업 없음)\n[작업 목록:task list]",
		}
	}

	UpdatePhase(projectPath, "run", len(tasks))
	log.Printf("[Task] RunAll: %d tasks, parallel=%d", len(tasks), parallel)

	// Sequential execution when parallel=1 (original behavior)
	if parallel <= 1 {
		return runAllSequential(ctx, projectPath, tasks)
	}

	// Parallel execution
	return runAllParallel(ctx, projectPath, tasks, parallel)
}

// runAllSequential runs tasks one by one (original behavior)
func runAllSequential(ctx context.Context, projectPath string, tasks []Task) types.Result {
	var success, failed int
	var messages []string

	for _, t := range tasks {
		if IsCancelled() || ctx.Err() != nil {
			skipped := len(tasks) - success - failed
			messages = append(messages, fmt.Sprintf("🛑 중단 요청으로 %d개 작업 건너뜀", skipped))
			break
		}

		UpdateCurrentTask(projectPath, t.ID)
		result := RunWithContext(ctx, projectPath, fmt.Sprintf("%d", t.ID))
		IncrementCompleted(projectPath)
		if result.Success {
			success++
			messages = append(messages, fmt.Sprintf("✅ #%d %s", t.ID, t.Title))
		} else {
			failed++
			messages = append(messages, fmt.Sprintf("❌ #%d %s: %s", t.ID, t.Title, result.Message))
			if result.ErrorType == "auth_error" {
				skipped := len(tasks) - success - failed
				messages = append(messages, fmt.Sprintf("🔐 인증 오류로 순회 중단, %d개 작업 건너뜀", skipped))
				break
			}
			if result.ErrorType == "cancelled" {
				skipped := len(tasks) - success - failed
				messages = append(messages, fmt.Sprintf("🛑 중단 요청으로 %d개 작업 건너뜀", skipped))
				break
			}
		}
	}

	summary := fmt.Sprintf("✅ 작업 실행 완료: 성공 %d개, 실패 %d개\n", success, failed)
	for _, msg := range messages {
		summary += msg + "\n"
	}

	return types.Result{
		Success: failed == 0,
		Message: summary,
	}
}

// runAllParallel runs tasks concurrently with a worker pool
func runAllParallel(parentCtx context.Context, projectPath string, tasks []Task, parallel int) types.Result {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	ResetActiveWorkers()
	defer ResetActiveWorkers()

	sem := make(chan struct{}, parallel)
	resultCh := make(chan runResult, len(tasks))
	var wg sync.WaitGroup

	for _, t := range tasks {
		// Check cancel/stop before dispatching
		if IsCancelled() {
			break
		}
		select {
		case <-ctx.Done():
			break
		default:
		}
		if ctx.Err() != nil {
			break
		}

		wg.Add(1)
		go func(t Task) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				// acquired
			case <-ctx.Done():
				resultCh <- runResult{TaskID: t.ID, Title: t.Title, Success: false, Message: "컨텍스트 취소됨"}
				return
			}
			defer func() { <-sem }()

			// Re-check after acquiring semaphore
			if IsCancelled() || ctx.Err() != nil {
				resultCh <- runResult{TaskID: t.ID, Title: t.Title, Success: false, Message: "중단됨"}
				return
			}

			UpdateActiveWorkers(projectPath, +1)
			log.Printf("[Task] Worker started task #%d (%s)", t.ID, t.Title)
			defer func() {
				UpdateActiveWorkers(projectPath, -1)
				log.Printf("[Task] Worker finished task #%d (%s)", t.ID, t.Title)
			}()

			result := RunWithContext(ctx, projectPath, fmt.Sprintf("%d", t.ID))
			rr := runResult{
				TaskID:  t.ID,
				Title:   t.Title,
				Success: result.Success,
				Message: result.Message,
				IsAuth:  result.ErrorType == "auth_error",
			}

			// Auth error: cancel all other workers
			if rr.IsAuth {
				log.Printf("[Task] Auth error detected in task #%d, cancelling remaining workers", t.ID)
				cancel()
			}

			resultCh <- rr
		}(t)
	}

	// Wait for all goroutines to finish, then close results channel
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	var success, failed, skipped int
	var messages []string
	var authDetected bool

	for rr := range resultCh {
		if rr.Message == "중단됨" || rr.Message == "컨텍스트 취소됨" {
			skipped++
			continue
		}
		IncrementCompleted(projectPath)
		if rr.Success {
			success++
			messages = append(messages, fmt.Sprintf("✅ #%d %s", rr.TaskID, rr.Title))
		} else {
			failed++
			messages = append(messages, fmt.Sprintf("❌ #%d %s: %s", rr.TaskID, rr.Title, rr.Message))
			if rr.IsAuth {
				authDetected = true
			}
		}
	}

	if authDetected {
		messages = append(messages, fmt.Sprintf("🔐 인증 오류로 순회 중단, %d개 작업 건너뜀", skipped))
	} else if IsCancelled() && skipped > 0 {
		messages = append(messages, fmt.Sprintf("🛑 중단 요청으로 %d개 작업 건너뜀", skipped))
	}

	summary := fmt.Sprintf("✅ 작업 실행 완료: 성공 %d개, 실패 %d개\n", success, failed)
	for _, msg := range messages {
		summary += msg + "\n"
	}

	return types.Result{
		Success: failed == 0,
		Message: summary,
	}
}
