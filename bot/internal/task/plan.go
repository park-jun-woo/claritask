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

// Plan generates plan for a task (1회차 순회: spec_ready → plan_ready/subdivided)
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
		// Get next spec_ready task (root level first)
		err = localDB.QueryRow(`
			SELECT id, parent_id, title, spec, plan, status, is_leaf, depth FROM tasks
			WHERE status = 'spec_ready'
			ORDER BY depth ASC, id ASC LIMIT 1
		`).Scan(&t.ID, &t.ParentID, &t.Title, &t.Spec, &t.Plan, &t.Status, &t.IsLeaf, &t.Depth)
		if err == sql.ErrNoRows {
			return types.Result{
				Success: true,
				Message: "Plan을 생성할 작업이 없습니다. (spec_ready 상태 작업 없음)\n[작업 목록:task list]",
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

	// Execute plan recursively
	return planRecursive(localDB, projectPath, &t)
}

// planRecursive executes plan for a task and its children recursively
func planRecursive(localDB *db.DB, projectPath string, t *Task) types.Result {
	// Check depth limit - force plan if at max depth
	forceLeaf := t.Depth >= MaxDepth

	// Get related tasks' specs
	relatedTasks, err := GetRelatedSpecs(localDB, t.ID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("연관 작업 조회 실패: %v", err),
		}
	}

	// Build report path
	reportPath := filepath.Join(projectPath, ".claribot", fmt.Sprintf("task-plan-%d-report.md", t.ID))
	// Ensure .claribot directory exists
	os.MkdirAll(filepath.Dir(reportPath), 0755)

	// Build prompt with report path
	prompt := BuildPlanPrompt(t, relatedTasks, reportPath)

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

	result, err := claude.Run(opts)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("Claude 실행 오류: %v", err),
		}
	}

	if result.ExitCode != 0 {
		// Save error
		now := db.TimeNow()
		if _, err := localDB.Exec(`UPDATE tasks SET error = ?, updated_at = ? WHERE id = ?`, result.Output, now, t.ID); err != nil {
			log.Printf("[Task] Plan 에러 저장 실패 (task #%d): %v", t.ID, err)
		}
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("Plan 생성 실패: %s", result.Output),
		}
	}

	// Parse output
	planResult := ParsePlanOutput(result.Output)
	now := db.TimeNow()

	if planResult.Type == "subdivided" && !forceLeaf {
		// Mark as subdivided, not a leaf
		_, err = localDB.Exec(`
			UPDATE tasks SET status = 'subdivided', is_leaf = 0, updated_at = ? WHERE id = ?
		`, now, t.ID)
		if err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("상태 업데이트 실패: %v", err),
			}
		}

		// Get newly created children and recursively plan them
		rows, err := localDB.Query(`
			SELECT id, parent_id, title, spec, plan, status, is_leaf, depth FROM tasks
			WHERE parent_id = ? AND status = 'spec_ready'
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

		// Recursively plan children
		var childResults []string
		for _, child := range children {
			childResult := planRecursive(localDB, projectPath, &child)
			if childResult.Success {
				childResults = append(childResults, fmt.Sprintf("✅ #%d %s", child.ID, child.Title))
			} else {
				childResults = append(childResults, fmt.Sprintf("❌ #%d %s: %s", child.ID, child.Title, childResult.Message))
			}
		}

		msg := fmt.Sprintf("📂 작업 #%d 분할됨: %s\n", t.ID, t.Title)
		for _, r := range childResults {
			msg += "  " + r + "\n"
		}

		return types.Result{
			Success: true,
			Message: msg,
			Data:    t,
		}
	}

	// Planned (leaf)
	_, err = localDB.Exec(`
		UPDATE tasks SET plan = ?, status = 'plan_ready', is_leaf = 1, updated_at = ? WHERE id = ?
	`, planResult.Plan, now, t.ID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("Plan 저장 실패: %v", err),
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("📋 작업 #%d Plan 생성 완료: %s\n[조회:task get %d][실행:task run %d]", t.ID, t.Title, t.ID, t.ID),
		Data:    t,
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

	// Get all root spec_ready tasks (no parent or parent is subdivided)
	rows, err := localDB.Query(`
		SELECT id, parent_id, title, spec, plan, status, is_leaf, depth FROM tasks
		WHERE status = 'spec_ready' AND (parent_id IS NULL OR parent_id IN (
			SELECT id FROM tasks WHERE status = 'subdivided'
		))
		ORDER BY depth ASC, id ASC
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
		if err := rows.Scan(&t.ID, &t.ParentID, &t.Title, &t.Spec, &t.Plan, &t.Status, &t.IsLeaf, &t.Depth); err != nil {
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

	// Plan each task (will recursively handle children)
	var success, failed int
	var messages []string

	for _, t := range tasks {
		// Re-check status (might have been planned as child of previous task)
		var currentStatus string
		err := localDB.QueryRow(`SELECT status FROM tasks WHERE id = ?`, t.ID).Scan(&currentStatus)
		if err != nil || currentStatus != "spec_ready" {
			continue
		}

		result := planRecursive(localDB, projectPath, &t)
		if result.Success {
			success++
			messages = append(messages, result.Message)
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
