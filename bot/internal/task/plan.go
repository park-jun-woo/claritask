package task

import (
	"database/sql"
	"fmt"
	"log"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/claude"
)

// Plan generates plan for a task (1회차 순회: spec_ready → plan_ready)
// If id is empty, plans next spec_ready task
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
		// Get next spec_ready task
		err = localDB.QueryRow(`
			SELECT id, title, spec, plan, status FROM tasks
			WHERE status = 'spec_ready' AND parent_id IS NULL
			ORDER BY id ASC LIMIT 1
		`).Scan(&t.ID, &t.Title, &t.Spec, &t.Plan, &t.Status)
		if err == sql.ErrNoRows {
			return types.Result{
				Success: true,
				Message: "Plan을 생성할 작업이 없습니다. (spec_ready 상태 작업 없음)\n[작업 목록:task list]",
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

	if t.Status != "spec_ready" {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업 #%d은(는) %s 상태입니다. (spec_ready 상태만 Plan 생성 가능)", t.ID, t.Status),
		}
	}

	if t.Spec == "" {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업 #%d에 Spec이 없습니다. 먼저 'task set %d spec <내용>'으로 명세서를 작성하세요.", t.ID, t.ID),
		}
	}

	// Get related tasks' specs
	relatedTasks, err := GetRelatedSpecs(localDB, t.ID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("연관 작업 조회 실패: %v", err),
		}
	}

	// Build prompt
	prompt := BuildPlanPrompt(&t, relatedTasks)

	// Run Claude Code
	opts := claude.Options{
		UserPrompt: prompt,
		WorkDir:    projectPath,
	}

	result, err := claude.Run(opts)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("Claude 실행 오류: %v", err),
		}
	}

	if result.ExitCode != 0 {
		// Save error and mark as failed
		now := db.TimeNow()
		if _, err := localDB.Exec(`UPDATE tasks SET error = ?, updated_at = ? WHERE id = ?`, result.Output, now, t.ID); err != nil {
			log.Printf("[Task] Plan 에러 저장 실패 (task #%d): %v", t.ID, err)
		}
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("Plan 생성 실패: %s", result.Output),
		}
	}

	// Save plan and update status
	now := db.TimeNow()
	_, err = localDB.Exec(`UPDATE tasks SET plan = ?, status = 'plan_ready', updated_at = ? WHERE id = ?`, result.Output, now, t.ID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("Plan 저장 실패: %v", err),
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("📋 작업 #%d Plan 생성 완료: %s\n[조회:task get %d][실행:task run %d]", t.ID, t.Title, t.ID, t.ID),
		Data:    &t,
	}
}

// PlanAll generates plans for all spec_ready tasks (1회차 순회 전체 실행)
func PlanAll(projectPath string) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	// Get all spec_ready tasks
	rows, err := localDB.Query(`
		SELECT id, title FROM tasks
		WHERE status = 'spec_ready'
		ORDER BY id ASC
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
			Message: "Plan을 생성할 작업이 없습니다. (spec_ready 상태 작업 없음)\n[작업 목록:task list]",
		}
	}

	// Plan each task
	var success, failed int
	var messages []string

	for _, t := range tasks {
		result := Plan(projectPath, fmt.Sprintf("%d", t.ID))
		if result.Success {
			success++
			messages = append(messages, fmt.Sprintf("✅ #%d %s", t.ID, t.Title))
		} else {
			failed++
			messages = append(messages, fmt.Sprintf("❌ #%d %s: %s", t.ID, t.Title, result.Message))
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
