package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/model"
)

// TaskCreateInput represents input for creating a task
type TaskCreateInput struct {
	FeatureID      int64
	SkeletonID     *int64
	Title          string
	Content        string
	TargetFile     string
	TargetLine     *int
	TargetFunction string
}

// CreateTask creates a new task
func CreateTask(database *db.DB, input TaskCreateInput) (int64, error) {
	now := db.TimeNow()

	result, err := database.Exec(
		`INSERT INTO tasks (feature_id, skeleton_id, title, content,
		 target_file, target_line, target_function, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', ?)`,
		input.FeatureID, input.SkeletonID, input.Title, input.Content,
		input.TargetFile, input.TargetLine, input.TargetFunction, now,
	)
	if err != nil {
		return 0, fmt.Errorf("create task: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}
	return id, nil
}

// GetTask retrieves a task by ID (accepts string or int64)
func GetTask(database *db.DB, id interface{}) (*model.Task, error) {
	var idInt int64
	switch v := id.(type) {
	case int64:
		idInt = v
	case string:
		var err error
		idInt, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid task ID: %v", id)
		}
	default:
		return nil, fmt.Errorf("invalid task ID type: %T", id)
	}

	row := database.QueryRow(
		`SELECT id, feature_id, skeleton_id, status, title, content,
		        target_file, target_line, target_function, result, error,
		        created_at, started_at, completed_at, failed_at
		 FROM tasks WHERE id = ?`, idInt,
	)
	return scanTask(row)
}

func scanTask(row *sql.Row) (*model.Task, error) {
	var t model.Task
	var idInt, featureIDInt int64
	var skeletonID sql.NullInt64
	var targetLine sql.NullInt64
	var createdAt string
	var startedAt, completedAt, failedAt sql.NullString

	err := row.Scan(
		&idInt, &featureIDInt, &skeletonID, &t.Status, &t.Title, &t.Content,
		&t.TargetFile, &targetLine, &t.TargetFunction, &t.Result, &t.Error,
		&createdAt, &startedAt, &completedAt, &failedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan task: %w", err)
	}

	t.ID = strconv.FormatInt(idInt, 10)
	t.FeatureID = featureIDInt
	if skeletonID.Valid {
		t.SkeletonID = &skeletonID.Int64
	}
	if targetLine.Valid {
		line := int(targetLine.Int64)
		t.TargetLine = &line
	}

	t.CreatedAt, _ = db.ParseTime(createdAt)
	if startedAt.Valid {
		ts, _ := db.ParseTime(startedAt.String)
		t.StartedAt = &ts
	}
	if completedAt.Valid {
		ts, _ := db.ParseTime(completedAt.String)
		t.CompletedAt = &ts
	}
	if failedAt.Valid {
		ts, _ := db.ParseTime(failedAt.String)
		t.FailedAt = &ts
	}

	return &t, nil
}

// scanTasks scans multiple tasks from rows
func scanTasks(rows *sql.Rows) ([]model.Task, error) {
	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		var idInt, featureIDInt int64
		var skeletonID sql.NullInt64
		var targetLine sql.NullInt64
		var createdAt string
		var startedAt, completedAt, failedAt sql.NullString

		err := rows.Scan(
			&idInt, &featureIDInt, &skeletonID, &t.Status, &t.Title, &t.Content,
			&t.TargetFile, &targetLine, &t.TargetFunction, &t.Result, &t.Error,
			&createdAt, &startedAt, &completedAt, &failedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}

		t.ID = strconv.FormatInt(idInt, 10)
		t.FeatureID = featureIDInt
		if skeletonID.Valid {
			t.SkeletonID = &skeletonID.Int64
		}
		if targetLine.Valid {
			line := int(targetLine.Int64)
			t.TargetLine = &line
		}

		t.CreatedAt, _ = db.ParseTime(createdAt)
		if startedAt.Valid {
			ts, _ := db.ParseTime(startedAt.String)
			t.StartedAt = &ts
		}
		if completedAt.Valid {
			ts, _ := db.ParseTime(completedAt.String)
			t.CompletedAt = &ts
		}
		if failedAt.Valid {
			ts, _ := db.ParseTime(failedAt.String)
			t.FailedAt = &ts
		}

		tasks = append(tasks, t)
	}
	return tasks, nil
}

// ListTasksByFeature lists all tasks for a feature
func ListTasksByFeature(database *db.DB, featureID int64) ([]model.Task, error) {
	rows, err := database.Query(
		`SELECT id, feature_id, skeleton_id, status, title, content,
		        target_file, target_line, target_function, result, error,
		        created_at, started_at, completed_at, failed_at
		 FROM tasks WHERE feature_id = ? ORDER BY id`, featureID,
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks by feature: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// ListAllTasks lists all tasks
func ListAllTasks(database *db.DB) ([]model.Task, error) {
	rows, err := database.Query(
		`SELECT id, feature_id, skeleton_id, status, title, content,
		        target_file, target_line, target_function, result, error,
		        created_at, started_at, completed_at, failed_at
		 FROM tasks ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list all tasks: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// StartTask starts a task (pending -> doing)
func StartTask(database *db.DB, id int64) error {
	task, err := GetTask(database, id)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}
	if task.Status != "pending" {
		return fmt.Errorf("task status must be 'pending' to start, current: %s", task.Status)
	}
	now := db.TimeNow()
	_, err = database.Exec(`UPDATE tasks SET status = 'doing', started_at = ? WHERE id = ?`, now, id)
	if err != nil {
		return fmt.Errorf("start task: %w", err)
	}
	return nil
}

// CompleteTask completes a task (doing -> done)
func CompleteTask(database *db.DB, id int64, result string) error {
	task, err := GetTask(database, id)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}
	if task.Status != "doing" {
		return fmt.Errorf("task status must be 'doing' to complete, current: %s", task.Status)
	}
	now := db.TimeNow()
	_, err = database.Exec(`UPDATE tasks SET status = 'done', result = ?, completed_at = ? WHERE id = ?`, result, now, id)
	if err != nil {
		return fmt.Errorf("complete task: %w", err)
	}
	return nil
}

// FailTask fails a task (doing -> failed)
func FailTask(database *db.DB, id int64, errMsg string) error {
	task, err := GetTask(database, id)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}
	if task.Status != "doing" {
		return fmt.Errorf("task status must be 'doing' to fail, current: %s", task.Status)
	}
	now := db.TimeNow()
	_, err = database.Exec(`UPDATE tasks SET status = 'failed', error = ?, failed_at = ? WHERE id = ?`, errMsg, now, id)
	if err != nil {
		return fmt.Errorf("fail task: %w", err)
	}
	return nil
}

// TaskPopResult represents the result of PopTask
type TaskPopResult struct {
	Task     *model.Task
	Manifest *model.Manifest
}

// PopTask pops the next pending task with manifest
func PopTask(database *db.DB) (*TaskPopResult, error) {
	// Get next pending task
	row := database.QueryRow(
		`SELECT id, feature_id, skeleton_id, status, title, content,
		        target_file, target_line, target_function, result, error,
		        created_at, started_at, completed_at, failed_at
		 FROM tasks WHERE status = 'pending' ORDER BY id LIMIT 1`,
	)
	task, err := scanTask(row)
	if err != nil {
		if err == sql.ErrNoRows || err.Error() == "scan task: sql: no rows in result set" {
			return &TaskPopResult{Task: nil, Manifest: nil}, nil
		}
		return nil, fmt.Errorf("pop task: %w", err)
	}

	// Start the task
	taskID, _ := strconv.ParseInt(task.ID, 10, 64)
	if err := StartTask(database, taskID); err != nil {
		return nil, fmt.Errorf("start task: %w", err)
	}

	// Update task status
	task.Status = "doing"

	// Build manifest
	manifest := &model.Manifest{
		Context: make(map[string]interface{}),
		Tech:    make(map[string]interface{}),
		Design:  make(map[string]interface{}),
		State:   make(map[string]string),
		Memos:   []model.MemoData{},
	}

	// Get context
	if ctx, err := GetContext(database); err == nil {
		manifest.Context = ctx
	}

	// Get tech
	if tech, err := GetTech(database); err == nil {
		manifest.Tech = tech
	}

	// Get design
	if design, err := GetDesign(database); err == nil {
		manifest.Design = design
	}

	// Get state
	if state, err := GetAllStates(database); err == nil {
		manifest.State = state
	}

	// Get high priority memos
	if memos, err := GetHighPriorityMemos(database); err == nil {
		for _, m := range memos {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(m.Data), &data); err == nil {
				manifest.Memos = append(manifest.Memos, model.MemoData{
					Scope:    m.Scope,
					ScopeID:  m.ScopeID,
					Key:      m.Key,
					Data:     data,
					Priority: m.Priority,
				})
			}
		}
	}

	// Update current state
	project, _ := GetProject(database)
	projectID := ""
	if project != nil {
		projectID = project.ID
	}

	// Find next task
	var nextTaskID int64
	row = database.QueryRow(`SELECT id FROM tasks WHERE status = 'pending' AND id > ? ORDER BY id LIMIT 1`, taskID)
	row.Scan(&nextTaskID)

	UpdateCurrentState(database, projectID, task.FeatureID, taskID, nextTaskID)

	// Start feature if pending
	feature, err := GetFeature(database, task.FeatureID)
	if err == nil && feature.Status == "pending" {
		StartFeature(database, task.FeatureID)
	}

	return &TaskPopResult{Task: task, Manifest: manifest}, nil
}

// TaskStatusResult represents task status summary
type TaskStatusResult struct {
	Total    int     `json:"total"`
	Pending  int     `json:"pending"`
	Doing    int     `json:"doing"`
	Done     int     `json:"done"`
	Failed   int     `json:"failed"`
	Progress float64 `json:"progress"`
}

// GetTaskStatus returns task status summary
func GetTaskStatus(database *db.DB) (*TaskStatusResult, error) {
	result := &TaskStatusResult{}

	row := database.QueryRow(`SELECT COUNT(*) FROM tasks`)
	row.Scan(&result.Total)

	row = database.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'pending'`)
	row.Scan(&result.Pending)

	row = database.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'doing'`)
	row.Scan(&result.Doing)

	row = database.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'done'`)
	row.Scan(&result.Done)

	row = database.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'failed'`)
	row.Scan(&result.Failed)

	if result.Total > 0 {
		result.Progress = float64(result.Done) / float64(result.Total) * 100
	}

	return result, nil
}

// PopTaskFull pops the next executable task with full context (FDL, Skeleton, Dependencies)
func PopTaskFull(database *db.DB) (*model.TaskPopResponse, error) {
	// Get next executable task (respecting dependencies)
	task, err := GetNextExecutableTask(database)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return &model.TaskPopResponse{Task: nil}, nil
	}

	// Start the task
	taskID, _ := strconv.ParseInt(task.ID, 10, 64)
	if err := StartTask(database, taskID); err != nil {
		return nil, fmt.Errorf("start task: %w", err)
	}
	task.Status = "doing"

	response := &model.TaskPopResponse{
		Task: task,
	}

	// Get FDL info
	fdlInfo, err := GetFDLInfoFromDB(database, task.FeatureID)
	if err == nil && fdlInfo != nil {
		response.FDL = fdlInfo
	}

	// Get Skeleton info if skeleton is associated
	if task.SkeletonID != nil {
		targetLine := 0
		if task.TargetLine != nil {
			targetLine = *task.TargetLine
		}
		skeletonInfo, err := GetSkeletonInfo(database, *task.SkeletonID, targetLine)
		if err == nil && skeletonInfo != nil {
			response.Skeleton = skeletonInfo
		}
	}

	// Get dependencies
	deps, _ := GetDependencyResults(database, task.ID)
	if len(deps) > 0 {
		response.Dependencies = deps
	}

	// Build manifest
	manifest := model.Manifest{
		Context: make(map[string]interface{}),
		Tech:    make(map[string]interface{}),
		Design:  make(map[string]interface{}),
		Feature: make(map[string]interface{}),
		Experts: []model.ExpertInfo{},
		State:   make(map[string]string),
		Memos:   []model.MemoData{},
	}

	if ctx, err := GetContext(database); err == nil {
		manifest.Context = ctx
	}
	if tech, err := GetTech(database); err == nil {
		manifest.Tech = tech
	}
	if design, err := GetDesign(database); err == nil {
		manifest.Design = design
	}

	// Add feature info to manifest
	feature, err := GetFeature(database, task.FeatureID)
	if err == nil && feature != nil {
		manifest.Feature = map[string]interface{}{
			"id":          feature.ID,
			"name":        feature.Name,
			"description": feature.Description,
			"spec":        feature.Spec,
			"status":      feature.Status,
		}
	}

	// Add assigned experts to manifest
	project, _ := GetProject(database)
	if project != nil {
		experts, err := GetAssignedExperts(database, project.ID)
		if err == nil {
			manifest.Experts = experts
		}
	}

	if state, err := GetAllStates(database); err == nil {
		manifest.State = state
	}
	if memos, err := GetHighPriorityMemos(database); err == nil {
		for _, m := range memos {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(m.Data), &data); err == nil {
				manifest.Memos = append(manifest.Memos, model.MemoData{
					Scope:    m.Scope,
					ScopeID:  m.ScopeID,
					Key:      m.Key,
					Data:     data,
					Priority: m.Priority,
				})
			}
		}
	}
	response.Manifest = manifest

	// Update current state
	projectID := ""
	if project != nil {
		projectID = project.ID
	}

	var nextTaskID int64
	row := database.QueryRow(`SELECT id FROM tasks WHERE status = 'pending' AND id > ? ORDER BY id LIMIT 1`, taskID)
	row.Scan(&nextTaskID)
	UpdateCurrentState(database, projectID, task.FeatureID, taskID, nextTaskID)

	// Start feature if pending
	if feature != nil && feature.Status == "pending" {
		StartFeature(database, task.FeatureID)
	}

	return response, nil
}

// GetNextExecutableTask returns the next pending task with all dependencies completed
func GetNextExecutableTask(database *db.DB) (*model.Task, error) {
	rows, err := database.Query(
		`SELECT id, feature_id, skeleton_id, status, title, content,
		        target_file, target_line, target_function, result, error,
		        created_at, started_at, completed_at, failed_at
		 FROM tasks
		 WHERE status = 'pending'
		 AND NOT EXISTS (
		     SELECT 1 FROM task_edges e
		     JOIN tasks dep ON e.to_task_id = dep.id
		     WHERE e.from_task_id = tasks.id AND dep.status != 'done'
		 )
		 ORDER BY id LIMIT 1`,
	)
	if err != nil {
		return nil, fmt.Errorf("get next executable task: %w", err)
	}
	defer rows.Close()

	tasks, err := scanTasks(rows)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	return &tasks[0], nil
}

// TaskListItem represents a task in list view with dependencies
type TaskListItem struct {
	ID             string   `json:"id"`
	FeatureID      int64    `json:"feature_id"`
	Title          string   `json:"title"`
	Status         string   `json:"status"`
	TargetFile     string   `json:"target_file,omitempty"`
	TargetFunction string   `json:"target_function,omitempty"`
	DependsOn      []string `json:"depends_on,omitempty"`
}

// ListTasksWithDependencies lists tasks with dependency information
func ListTasksWithDependencies(database *db.DB, featureID int64) ([]TaskListItem, error) {
	tasks, err := ListTasksByFeature(database, featureID)
	if err != nil {
		return nil, err
	}

	var items []TaskListItem
	for _, t := range tasks {
		item := TaskListItem{
			ID:             t.ID,
			FeatureID:      t.FeatureID,
			Title:          t.Title,
			Status:         t.Status,
			TargetFile:     t.TargetFile,
			TargetFunction: t.TargetFunction,
		}

		// Get dependencies
		deps, _ := GetTaskDependencies(database, t.ID)
		for _, dep := range deps {
			item.DependsOn = append(item.DependsOn, dep.ID)
		}

		items = append(items, item)
	}

	return items, nil
}

// ResetTaskToPending resets a task to pending status
func ResetTaskToPending(database *db.DB, id int64) error {
	_, err := database.Exec(`
		UPDATE tasks
		SET status = 'pending',
		    started_at = NULL,
		    completed_at = NULL,
		    failed_at = NULL,
		    result = '',
		    error = ''
		WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("reset task: %w", err)
	}
	return nil
}
