package task

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/claude"
)

// Plan generates plan for a task (1회차 순회: todo → planned/split)
// If id is empty, plans next todo task
func Plan(projectPath, id string) types.Result {
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
		// Get next todo task (root level first)
		err = localDB.QueryRow(`
			SELECT id, parent_id, title, spec, plan, status, is_leaf, depth FROM tasks
			WHERE status = 'todo'
			ORDER BY depth ASC, id ASC LIMIT 1
		`).Scan(&t.ID, &t.ParentID, &t.Title, &t.Spec, &t.Plan, &t.Status, &t.IsLeaf, &t.Depth)
		if err == sql.ErrNoRows {
			return types.Result{
				Success: true,
				Message: "Plan을 생성할 작업이 없습니다. (todo 상태 작업 없음)\n[작업 목록:task list]",
			}
		}
	} else {
		err = localDB.QueryRow(`
			SELECT id, parent_id, title, spec, plan, status, is_leaf, depth FROM tasks WHERE id = ?
		`, id).Scan(&t.ID, &t.ParentID, &t.Title, &t.Spec, &t.Plan, &t.Status, &t.IsLeaf, &t.Depth)
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

	if t.Status != "todo" {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업 #%d은(는) %s 상태입니다. (todo 상태만 Plan 생성 가능)", t.ID, t.Status),
		}
	}

	if t.Spec == "" {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업 #%d에 Spec이 없습니다. 먼저 'task set %d spec <내용>'으로 명세서를 작성하세요.", t.ID, t.ID),
		}
	}

	// Insert traversal record
	travID, travErr := insertTraversal(localDB, "plan", &t.ID, "")
	if travErr != nil {
		log.Printf("[Task] traversal INSERT 실패: %v", travErr)
	}

	result := planRecursive(context.Background(), localDB, projectPath, &t)

	// Update traversal record
	if travErr == nil {
		status := "done"
		total, success, failed := 1, 0, 0
		if result.Success {
			success = 1
		} else {
			status = "failed"
			failed = 1
		}
		finishTraversal(localDB, travID, status, total, success, failed)
	}

	return result
}

// planRecursive executes plan for a task and its children recursively
func planRecursive(ctx context.Context, localDB *db.DB, projectPath string, t *Task) types.Result {
	UpdateCurrentTask(t.ID)

	// Check depth limit - force plan if at max depth
	forceLeaf := t.Depth >= MaxDepth

	// Build context map (lightweight task tree summary)
	contextMap, err := BuildContextMap(localDB)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("Context Map 생성 실패: %v", err),
		}
	}

	// Build report path
	reportPath := filepath.Join(projectPath, ".claribot", fmt.Sprintf("task-plan-%d-report.md", t.ID))
	// Ensure .claribot directory exists
	if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("report 디렉토리 생성 실패: %v", err),
		}
	}

	// Build prompt with report path
	prompt := BuildPlanPrompt(t, contextMap, reportPath)

	// Add force leaf instruction if at max depth
	if forceLeaf {
		prompt += "\n\n⚠️ 최대 깊이에 도달했습니다. 반드시 [PLANNED] 형식으로 계획을 작성하세요. 분할은 불가능합니다."
	}

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
		return ret
	}

	if result.ExitCode != 0 {
		// Check for authentication error
		authError := claude.IsAuthError(result)
		if authError {
			log.Printf("[Task] Plan 인증 오류 감지 (task #%d)", t.ID)
		}

		// Save error
		now := db.TimeNow()
		if _, err := localDB.Exec(`UPDATE tasks SET error = ?, updated_at = ? WHERE id = ?`, result.Output, now, t.ID); err != nil {
			log.Printf("[Task] Plan 에러 저장 실패 (task #%d): %v", t.ID, err)
		}

		ret := types.Result{
			Success: false,
			Message: fmt.Sprintf("Plan 생성 실패: %s", result.Output),
		}
		if authError {
			ret.ErrorType = "auth_error"
		}
		return ret
	}

	// Parse output
	planResult := ParsePlanOutput(result.Output)
	now := db.TimeNow()

	if planResult.Type == "split" && !forceLeaf {
		// Mark as split, not a leaf
		_, err = localDB.Exec(`
			UPDATE tasks SET status = 'split', is_leaf = 0, updated_at = ? WHERE id = ?
		`, now, t.ID)
		if err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("상태 업데이트 실패: %v", err),
			}
		}

		// Clean up report file after DB save
		if err := os.Remove(reportPath); err != nil && !os.IsNotExist(err) {
			log.Printf("[Task] Plan report 파일 삭제 실패 (task #%d): %v", t.ID, err)
		}

		// Get newly created children and recursively plan them
		rows, err := localDB.Query(`
			SELECT id, parent_id, title, spec, plan, status, is_leaf, depth FROM tasks
			WHERE parent_id = ? AND status = 'todo'
			ORDER BY id ASC
		`, t.ID)
		if err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("하위 작업 조회 실패: %v", err),
			}
		}
		defer rows.Close()

		var children []Task
		for rows.Next() {
			var child Task
			if err := rows.Scan(&child.ID, &child.ParentID, &child.Title, &child.Spec, &child.Plan, &child.Status, &child.IsLeaf, &child.Depth); err != nil {
				continue
			}
			children = append(children, child)
		}
		if err := rows.Err(); err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("행 순회 오류: %v", err),
			}
		}

		// Recursively plan children
		var childResults []string
		var authErrorDetected bool
		for _, child := range children {
			// Check cancel flag before each child
			if IsCancelled() || ctx.Err() != nil {
				childResults = append(childResults, fmt.Sprintf("🛑 중단 요청으로 나머지 하위 작업 건너뜀"))
				break
			}
			if child.Spec == "" {
				log.Printf("[Task] Plan skip: child #%d (%s) has empty spec, keeping todo", child.ID, child.Title)
				childResults = append(childResults, fmt.Sprintf("⚠️ #%d %s: Spec이 비어있어 건너뜀 (todo 유지)", child.ID, child.Title))
				continue
			}
			childResult := planRecursive(ctx, localDB, projectPath, &child)
			if childResult.Success {
				childResults = append(childResults, fmt.Sprintf("✅ #%d %s", child.ID, child.Title))
			} else {
				childResults = append(childResults, fmt.Sprintf("❌ #%d %s: %s", child.ID, child.Title, childResult.Message))
				// Auth error from child: propagate upward
				if childResult.ErrorType == "auth_error" {
					authErrorDetected = true
					childResults = append(childResults, "🔐 인증 오류로 나머지 하위 작업 건너뜀")
					break
				}
			}
		}

		msg := fmt.Sprintf("📂 작업 #%d 분할됨: %s\n", t.ID, t.Title)
		for _, r := range childResults {
			msg += "  " + r + "\n"
		}

		ret := types.Result{
			Success: !authErrorDetected,
			Message: msg,
			Data:    t,
		}
		if authErrorDetected {
			ret.ErrorType = "auth_error"
		}
		return ret
	}

	// Planned (leaf)
	_, err = localDB.Exec(`
		UPDATE tasks SET plan = ?, status = 'planned', is_leaf = 1, updated_at = ? WHERE id = ?
	`, planResult.Plan, now, t.ID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("Plan 저장 실패: %v", err),
		}
	}

	// Clean up report file after DB save
	if err := os.Remove(reportPath); err != nil && !os.IsNotExist(err) {
		log.Printf("[Task] Plan report 파일 삭제 실패 (task #%d): %v", t.ID, err)
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("📋 작업 #%d Plan 생성 완료: %s\n[조회:task get %d][실행:task run %d]", t.ID, t.Title, t.ID, t.ID),
		Data:    t,
	}
}

// PlanAll generates plans for all todo tasks (1회차 순회 전체 실행)
func PlanAll(projectPath string) types.Result {
	ResetCancel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startTime := time.Now()
	SetCycleState(CycleState{
		Running:     true,
		Type:        "plan",
		StartedAt:   startTime,
		ProjectPath: projectPath,
	})
	SetCycleCancel(cancel)
	defer ClearCycleState()

	// Insert traversal record
	localDB, travErr := db.OpenLocal(projectPath)
	var travID int64
	if travErr == nil {
		travID, travErr = insertTraversal(localDB, "plan", nil, "")
		if travErr != nil {
			log.Printf("[Task] traversal INSERT 실패: %v", travErr)
		}
		localDB.Close()
	}

	result := planAllInternal(ctx, projectPath)

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
		notification := fmt.Sprintf("📋 Plan 순회 완료\n소요: %s\n%s", formatDuration(time.Since(startTime)), result.Message)
		globalNotifier(nil, notification)
	}

	return result
}

// planResult holds the result of a single task plan for channel communication
type planResult struct {
	TaskID  int
	Title   string
	Success bool
	Message string
	IsAuth  bool
}

// planAllInternal is the internal implementation of PlanAll without CycleState management.
// Used by Cycle() to avoid overwriting the cycle type.
// Supports parallel execution based on project's parallel config.
func planAllInternal(ctx context.Context, projectPath string) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}

	// Read parallel config
	parallel := getParallel(localDB)

	// Get all root todo tasks (no parent or parent is split)
	rows, err := localDB.Query(`
		SELECT id, parent_id, title, spec, plan, status, is_leaf, depth FROM tasks
		WHERE status = 'todo' AND (parent_id IS NULL OR parent_id IN (
			SELECT id FROM tasks WHERE status = 'split'
		))
		ORDER BY depth ASC, id ASC
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
		if err := rows.Scan(&t.ID, &t.ParentID, &t.Title, &t.Spec, &t.Plan, &t.Status, &t.IsLeaf, &t.Depth); err != nil {
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
			Message: "Plan을 생성할 작업이 없습니다. (todo 상태 작업 없음)\n[작업 목록:task list]",
		}
	}

	log.Printf("[Task] PlanAll: %d tasks, parallel=%d", len(tasks), parallel)

	// Sequential execution when parallel=1 (original behavior)
	if parallel <= 1 {
		return planAllSequential(ctx, projectPath, tasks)
	}

	// Parallel execution
	return planAllParallel(ctx, projectPath, tasks, parallel)
}

// planAllSequential plans tasks one by one (original behavior)
func planAllSequential(ctx context.Context, projectPath string, tasks []Task) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	var success, failed int
	var messages []string

	for _, t := range tasks {
		if IsCancelled() || ctx.Err() != nil {
			skipped := len(tasks) - success - failed
			messages = append(messages, fmt.Sprintf("🛑 중단 요청으로 %d개 작업 건너뜀", skipped))
			break
		}

		// Re-check status (might have been planned as child of previous task)
		var currentStatus string
		err := localDB.QueryRow(`SELECT status FROM tasks WHERE id = ?`, t.ID).Scan(&currentStatus)
		if err != nil || currentStatus != "todo" {
			continue
		}

		result := planRecursive(ctx, localDB, projectPath, &t)
		if result.Success {
			success++
			messages = append(messages, result.Message)
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

	summary := fmt.Sprintf("📋 Plan 생성 완료: 성공 %d개, 실패 %d개\n", success, failed)
	for _, msg := range messages {
		summary += msg + "\n"
	}

	return types.Result{
		Success: failed == 0,
		Message: summary,
	}
}

// planAllParallel plans tasks concurrently with a worker pool.
// Each goroutine opens its own DB connection for SQLite concurrency safety.
func planAllParallel(parentCtx context.Context, projectPath string, tasks []Task, parallel int) types.Result {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	ResetActiveWorkers()
	defer ResetActiveWorkers()

	sem := make(chan struct{}, parallel)
	resultCh := make(chan planResult, len(tasks))
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
				resultCh <- planResult{TaskID: t.ID, Title: t.Title, Success: false, Message: "컨텍스트 취소됨"}
				return
			}
			defer func() { <-sem }()

			// Re-check after acquiring semaphore
			if IsCancelled() || ctx.Err() != nil {
				resultCh <- planResult{TaskID: t.ID, Title: t.Title, Success: false, Message: "중단됨"}
				return
			}

			// Each goroutine opens its own DB connection (SQLite concurrency)
			workerDB, err := db.OpenLocal(projectPath)
			if err != nil {
				resultCh <- planResult{TaskID: t.ID, Title: t.Title, Success: false, Message: fmt.Sprintf("DB 열기 실패: %v", err)}
				return
			}
			defer workerDB.Close()

			// Re-check status (might have been planned as child of another task)
			var currentStatus string
			err = workerDB.QueryRow(`SELECT status FROM tasks WHERE id = ?`, t.ID).Scan(&currentStatus)
			if err != nil || currentStatus != "todo" {
				resultCh <- planResult{TaskID: t.ID, Title: t.Title, Success: true, Message: "이미 처리됨"}
				return
			}

			UpdateActiveWorkers(+1)
			log.Printf("[Task] Worker started plan task #%d (%s)", t.ID, t.Title)
			defer func() {
				UpdateActiveWorkers(-1)
				log.Printf("[Task] Worker finished plan task #%d (%s)", t.ID, t.Title)
			}()

			result := planRecursive(ctx, workerDB, projectPath, &t)
			pr := planResult{
				TaskID:  t.ID,
				Title:   t.Title,
				Success: result.Success,
				Message: result.Message,
				IsAuth:  result.ErrorType == "auth_error",
			}

			// Auth error: cancel all other workers
			if pr.IsAuth {
				log.Printf("[Task] Auth error detected in plan task #%d, cancelling remaining workers", t.ID)
				cancel()
			}

			resultCh <- pr
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

	for pr := range resultCh {
		if pr.Message == "중단됨" || pr.Message == "컨텍스트 취소됨" {
			skipped++
			continue
		}
		if pr.Message == "이미 처리됨" {
			continue
		}
		if pr.Success {
			success++
			messages = append(messages, pr.Message)
		} else {
			failed++
			messages = append(messages, fmt.Sprintf("❌ #%d %s: %s", pr.TaskID, pr.Title, pr.Message))
			if pr.IsAuth {
				authDetected = true
			}
		}
	}

	if authDetected {
		messages = append(messages, fmt.Sprintf("🔐 인증 오류로 순회 중단, %d개 작업 건너뜀", skipped))
	} else if IsCancelled() && skipped > 0 {
		messages = append(messages, fmt.Sprintf("🛑 중단 요청으로 %d개 작업 건너뜀", skipped))
	}

	summary := fmt.Sprintf("📋 Plan 생성 완료: 성공 %d개, 실패 %d개\n", success, failed)
	for _, msg := range messages {
		summary += msg + "\n"
	}

	return types.Result{
		Success: failed == 0,
		Message: summary,
	}
}
