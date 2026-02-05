package task

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/claude"
)

// Run runs a task (2회차 순회: plan_ready → done)
// If id is empty, runs next plan_ready task
func Run(projectPath, id string) types.Result {
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
		// Get next plan_ready leaf task
		err = localDB.QueryRow(`
			SELECT id, title, spec, plan, status FROM tasks
			WHERE status = 'plan_ready' AND is_leaf = 1
			ORDER BY depth DESC, id ASC LIMIT 1
		`).Scan(&t.ID, &t.Title, &t.Spec, &t.Plan, &t.Status)
		if err == sql.ErrNoRows {
			return types.Result{
				Success: true,
				Message: "실행할 작업이 없습니다. (plan_ready 상태 leaf 작업 없음)\n[작업 목록:task list]",
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

	if t.Status != "plan_ready" {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업 #%d은(는) %s 상태입니다. (plan_ready 상태만 실행 가능)", t.ID, t.Status),
		}
	}

	if t.Plan == "" {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업 #%d에 Plan이 없습니다. 먼저 'task plan %d'를 실행하세요.", t.ID, t.ID),
		}
	}

	// Get related tasks' plans
	relatedTasks, err := GetRelatedPlans(localDB, t.ID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("연관 작업 조회 실패: %v", err),
		}
	}

	// Build report path
	reportPath := filepath.Join(projectPath, ".claribot", fmt.Sprintf("task-run-%d-report.md", t.ID))
	// Ensure .claribot directory exists
	os.MkdirAll(filepath.Dir(reportPath), 0755)

	// Build prompt with report path
	prompt := BuildExecutePrompt(&t, relatedTasks, reportPath)

	// Run Claude Code
	opts := claude.Options{
		UserPrompt: prompt,
		WorkDir:    projectPath,
		ReportPath: reportPath,
	}

	result, err := claude.Run(opts)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("Claude 실행 오류: %v", err),
		}
	}

	now := db.TimeNow()

	if result.ExitCode != 0 {
		// Save error and mark as failed
		if _, err := localDB.Exec(`UPDATE tasks SET error = ?, status = 'failed', updated_at = ? WHERE id = ?`, result.Output, now, t.ID); err != nil {
			log.Printf("[Task] Run 에러 저장 실패 (task #%d): %v", t.ID, err)
		}
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업 실행 실패: %s", result.Output),
		}
	}

	// Save report and update status to done
	_, err = localDB.Exec(`UPDATE tasks SET report = ?, status = 'done', updated_at = ? WHERE id = ?`, result.Output, now, t.ID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("Report 저장 실패: %v", err),
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("✅ 작업 #%d 완료: %s\n[조회:task get %d]", t.ID, t.Title, t.ID),
		Data:    &t,
	}
}

// RunAll runs all plan_ready tasks (2회차 순회 전체 실행)
func RunAll(projectPath string) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	// Get all plan_ready leaf tasks (deepest first)
	rows, err := localDB.Query(`
		SELECT id, title FROM tasks
		WHERE status = 'plan_ready' AND is_leaf = 1
		ORDER BY depth DESC, id ASC
	`)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title); err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("스캔 실패: %v", err),
			}
		}
		tasks = append(tasks, t)
	}

	if len(tasks) == 0 {
		return types.Result{
			Success: true,
			Message: "실행할 작업이 없습니다. (plan_ready 상태 작업 없음)\n[작업 목록:task list]",
		}
	}

	// Run each task
	var success, failed int
	var messages []string

	for _, t := range tasks {
		result := Run(projectPath, fmt.Sprintf("%d", t.ID))
		if result.Success {
			success++
			messages = append(messages, fmt.Sprintf("✅ #%d %s", t.ID, t.Title))
		} else {
			failed++
			messages = append(messages, fmt.Sprintf("❌ #%d %s: %s", t.ID, t.Title, result.Message))
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
