package service

import (
	"fmt"
	"strconv"

	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/model"
)

// AddTaskEdge adds a dependency edge between tasks
// from depends on to (to must be completed before from can start)
func AddTaskEdge(database *db.DB, fromID, toID string) error {
	// Check for cycle
	hasCycle, _, err := CheckTaskCycle(database, fromID, toID)
	if err != nil {
		return fmt.Errorf("check cycle: %w", err)
	}
	if hasCycle {
		return fmt.Errorf("adding edge would create a cycle")
	}

	now := db.TimeNow()
	_, err = database.Exec(
		`INSERT INTO task_edges (from_task_id, to_task_id, created_at) VALUES (?, ?, ?)`,
		fromID, toID, now,
	)
	if err != nil {
		return fmt.Errorf("add task edge: %w", err)
	}
	return nil
}

// RemoveTaskEdge removes a dependency edge between tasks
func RemoveTaskEdge(database *db.DB, fromID, toID string) error {
	_, err := database.Exec(
		`DELETE FROM task_edges WHERE from_task_id = ? AND to_task_id = ?`,
		fromID, toID,
	)
	if err != nil {
		return fmt.Errorf("remove task edge: %w", err)
	}
	return nil
}

// GetTaskEdges retrieves all task edges
func GetTaskEdges(database *db.DB) ([]model.TaskEdge, error) {
	rows, err := database.Query(
		`SELECT from_task_id, to_task_id, created_at FROM task_edges`,
	)
	if err != nil {
		return nil, fmt.Errorf("get task edges: %w", err)
	}
	defer rows.Close()

	var edges []model.TaskEdge
	for rows.Next() {
		var e model.TaskEdge
		var createdAt string
		if err := rows.Scan(&e.FromTaskID, &e.ToTaskID, &createdAt); err != nil {
			return nil, fmt.Errorf("scan edge: %w", err)
		}
		e.CreatedAt, _ = db.ParseTime(createdAt)
		edges = append(edges, e)
	}
	return edges, nil
}

// GetTaskEdgesByFeature retrieves task edges within a feature
func GetTaskEdgesByFeature(database *db.DB, featureID int64) ([]model.TaskEdge, error) {
	rows, err := database.Query(
		`SELECT e.from_task_id, e.to_task_id, e.created_at
		 FROM task_edges e
		 JOIN tasks t1 ON e.from_task_id = t1.id
		 JOIN tasks t2 ON e.to_task_id = t2.id
		 WHERE t1.feature_id = ? AND t2.feature_id = ?`, featureID, featureID,
	)
	if err != nil {
		return nil, fmt.Errorf("get task edges by feature: %w", err)
	}
	defer rows.Close()

	var edges []model.TaskEdge
	for rows.Next() {
		var e model.TaskEdge
		var createdAt string
		if err := rows.Scan(&e.FromTaskID, &e.ToTaskID, &createdAt); err != nil {
			return nil, fmt.Errorf("scan edge: %w", err)
		}
		e.CreatedAt, _ = db.ParseTime(createdAt)
		edges = append(edges, e)
	}
	return edges, nil
}

// GetTaskDependencies retrieves tasks that this task depends on
func GetTaskDependencies(database *db.DB, taskID string) ([]model.Task, error) {
	rows, err := database.Query(
		`SELECT t.id, t.feature_id, t.skeleton_id, t.status, t.title, t.content,
		        t.target_file, t.target_line, t.target_function, t.result, t.error,
		        t.created_at, t.started_at, t.completed_at, t.failed_at
		 FROM tasks t
		 JOIN task_edges e ON t.id = e.to_task_id
		 WHERE e.from_task_id = ?`, taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("get task dependencies: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// GetTaskDependents retrieves tasks that depend on this task
func GetTaskDependents(database *db.DB, taskID string) ([]model.Task, error) {
	rows, err := database.Query(
		`SELECT t.id, t.feature_id, t.skeleton_id, t.status, t.title, t.content,
		        t.target_file, t.target_line, t.target_function, t.result, t.error,
		        t.created_at, t.started_at, t.completed_at, t.failed_at
		 FROM tasks t
		 JOIN task_edges e ON t.id = e.from_task_id
		 WHERE e.to_task_id = ?`, taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("get task dependents: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// GetDependencyResults retrieves results of tasks that this task depends on
func GetDependencyResults(database *db.DB, taskID string) ([]model.Dependency, error) {
	deps, err := GetTaskDependencies(database, taskID)
	if err != nil {
		return nil, err
	}

	var results []model.Dependency
	for _, t := range deps {
		results = append(results, model.Dependency{
			ID:     t.ID,
			Title:  t.Title,
			Result: t.Result,
			File:   t.TargetFile,
		})
	}
	return results, nil
}

// CheckTaskCycle checks if adding an edge would create a cycle
// Uses DFS to detect if there's a path from toID to fromID
func CheckTaskCycle(database *db.DB, fromID, toID string) (bool, []string, error) {
	visited := make(map[string]bool)
	path := []string{}

	var dfs func(current string) bool
	dfs = func(current string) bool {
		if current == fromID {
			return true
		}
		if visited[current] {
			return false
		}
		visited[current] = true
		path = append(path, current)

		deps, err := GetTaskDependencies(database, current)
		if err != nil {
			return false
		}

		for _, dep := range deps {
			if dfs(dep.ID) {
				return true
			}
		}

		path = path[:len(path)-1]
		return false
	}

	hasCycle := dfs(toID)
	if hasCycle {
		path = append(path, fromID)
	}
	return hasCycle, path, nil
}

// DetectAllCycles detects all cycles in the task graph
func DetectAllCycles(database *db.DB) ([][]string, error) {
	edges, err := GetTaskEdges(database)
	if err != nil {
		return nil, err
	}

	// Build adjacency list
	graph := make(map[string][]string)
	nodes := make(map[string]bool)
	for _, e := range edges {
		graph[e.FromTaskID] = append(graph[e.FromTaskID], e.ToTaskID)
		nodes[e.FromTaskID] = true
		nodes[e.ToTaskID] = true
	}

	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				dfs(neighbor)
			} else if recStack[neighbor] {
				// Found cycle
				cycleStart := -1
				for i, n := range path {
					if n == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]string, len(path)-cycleStart)
					copy(cycle, path[cycleStart:])
					cycles = append(cycles, cycle)
				}
			}
		}

		path = path[:len(path)-1]
		recStack[node] = false
	}

	for node := range nodes {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles, nil
}

// TopologicalSortTasks sorts tasks by dependency order using Kahn's algorithm
func TopologicalSortTasks(database *db.DB, featureID int64) ([]model.Task, error) {
	tasks, err := ListTasksByFeature(database, featureID)
	if err != nil {
		return nil, err
	}

	edges, err := GetTaskEdgesByFeature(database, featureID)
	if err != nil {
		return nil, err
	}

	// Build in-degree map and adjacency list
	inDegree := make(map[string]int)
	outEdges := make(map[string][]string)
	taskMap := make(map[string]model.Task)

	for _, t := range tasks {
		inDegree[t.ID] = 0
		taskMap[t.ID] = t
	}

	for _, e := range edges {
		inDegree[e.FromTaskID]++
		outEdges[e.ToTaskID] = append(outEdges[e.ToTaskID], e.FromTaskID)
	}

	// Find nodes with no incoming edges
	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	var sorted []model.Task
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if task, ok := taskMap[current]; ok {
			sorted = append(sorted, task)
		}

		for _, neighbor := range outEdges[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return sorted, nil
}

// TopologicalSortFeatures sorts features by dependency order
func TopologicalSortFeatures(database *db.DB, projectID string) ([]model.Feature, error) {
	features, err := ListFeatures(database, projectID)
	if err != nil {
		return nil, err
	}

	// Build in-degree map and adjacency list
	inDegree := make(map[int64]int)
	outEdges := make(map[int64][]int64)
	featureMap := make(map[int64]model.Feature)

	for _, f := range features {
		inDegree[f.ID] = 0
		featureMap[f.ID] = f
	}

	for _, f := range features {
		deps, err := GetFeatureDependencies(database, f.ID)
		if err != nil {
			continue
		}
		inDegree[f.ID] = len(deps)
		for _, dep := range deps {
			outEdges[dep.ID] = append(outEdges[dep.ID], f.ID)
		}
	}

	// Find nodes with no incoming edges
	var queue []int64
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	var sorted []model.Feature
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if feature, ok := featureMap[current]; ok {
			sorted = append(sorted, feature)
		}

		for _, neighbor := range outEdges[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return sorted, nil
}

// GetExecutableTasks retrieves pending tasks with all dependencies completed
func GetExecutableTasks(database *db.DB) ([]model.Task, error) {
	rows, err := database.Query(
		`SELECT t.id, t.feature_id, t.skeleton_id, t.status, t.title, t.content,
		        t.target_file, t.target_line, t.target_function, t.result, t.error,
		        t.created_at, t.started_at, t.completed_at, t.failed_at
		 FROM tasks t
		 WHERE t.status = 'pending'
		 AND NOT EXISTS (
		     SELECT 1 FROM task_edges e
		     JOIN tasks dep ON e.to_task_id = dep.id
		     WHERE e.from_task_id = t.id AND dep.status != 'done'
		 )
		 ORDER BY t.id`,
	)
	if err != nil {
		return nil, fmt.Errorf("get executable tasks: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// IsTaskExecutable checks if a task can be executed
func IsTaskExecutable(database *db.DB, taskID string) (bool, []model.Task, error) {
	task, err := GetTask(database, taskID)
	if err != nil {
		return false, nil, err
	}

	if task.Status != "pending" {
		return false, nil, fmt.Errorf("task is not pending")
	}

	deps, err := GetTaskDependencies(database, taskID)
	if err != nil {
		return false, nil, err
	}

	var blocking []model.Task
	for _, dep := range deps {
		if dep.Status != "done" {
			blocking = append(blocking, dep)
		}
	}

	return len(blocking) == 0, blocking, nil
}

// EdgeListResult represents the result of listing all edges
type EdgeListResult struct {
	FeatureEdges []FeatureEdgeItem `json:"feature_edges"`
	TaskEdges    []TaskEdgeItem    `json:"task_edges"`
	TotalFeature int               `json:"total_feature_edges"`
	TotalTask    int               `json:"total_task_edges"`
}

// FeatureEdgeItem represents a feature edge with names
type FeatureEdgeItem struct {
	From FeatureRef `json:"from"`
	To   FeatureRef `json:"to"`
}

// TaskEdgeItem represents a task edge with titles
type TaskEdgeItem struct {
	From TaskRef `json:"from"`
	To   TaskRef `json:"to"`
}

// FeatureRef represents a feature reference
type FeatureRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// TaskRef represents a task reference
type TaskRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ListAllEdges retrieves all edges with feature/task names
func ListAllEdges(database *db.DB) (*EdgeListResult, error) {
	result := &EdgeListResult{}

	// Get feature edges
	rows, err := database.Query(
		`SELECT e.from_feature_id, f1.name, e.to_feature_id, f2.name
		 FROM feature_edges e
		 JOIN features f1 ON e.from_feature_id = f1.id
		 JOIN features f2 ON e.to_feature_id = f2.id`,
	)
	if err != nil {
		return nil, fmt.Errorf("get feature edges: %w", err)
	}

	for rows.Next() {
		var item FeatureEdgeItem
		if err := rows.Scan(&item.From.ID, &item.From.Name, &item.To.ID, &item.To.Name); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan feature edge: %w", err)
		}
		result.FeatureEdges = append(result.FeatureEdges, item)
	}
	rows.Close()
	result.TotalFeature = len(result.FeatureEdges)

	// Get task edges
	rows, err = database.Query(
		`SELECT e.from_task_id, t1.title, e.to_task_id, t2.title
		 FROM task_edges e
		 JOIN tasks t1 ON e.from_task_id = t1.id
		 JOIN tasks t2 ON e.to_task_id = t2.id`,
	)
	if err != nil {
		return nil, fmt.Errorf("get task edges: %w", err)
	}

	for rows.Next() {
		var item TaskEdgeItem
		var fromID, toID int64
		if err := rows.Scan(&fromID, &item.From.Title, &toID, &item.To.Title); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan task edge: %w", err)
		}
		item.From.ID = strconv.FormatInt(fromID, 10)
		item.To.ID = strconv.FormatInt(toID, 10)
		result.TaskEdges = append(result.TaskEdges, item)
	}
	rows.Close()
	result.TotalTask = len(result.TaskEdges)

	return result, nil
}

// InferredEdge represents an inferred dependency edge
type InferredEdge struct {
	FromID     int64   `json:"from_id"`
	FromName   string  `json:"from_name"`
	ToID       int64   `json:"to_id"`
	ToName     string  `json:"to_name"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"`
}

// InferenceContext provides context for LLM edge inference
type InferenceContext struct {
	Type     string                 `json:"type"` // "task" or "feature"
	Items    []InferenceItem        `json:"items"`
	Existing []InferenceExisting    `json:"existing_edges"`
	Prompt   string                 `json:"prompt"`
}

// InferenceItem represents a task or feature for inference
type InferenceItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Layer       string `json:"layer,omitempty"` // for tasks: model, service, api, ui
}

// InferenceExisting represents an existing edge
type InferenceExisting struct {
	FromID   int64  `json:"from_id"`
	FromName string `json:"from_name"`
	ToID     int64  `json:"to_id"`
	ToName   string `json:"to_name"`
}

// PrepareTaskEdgeInference prepares context for LLM to infer task edges
func PrepareTaskEdgeInference(database *db.DB, featureID int64) (*InferenceContext, error) {
	ctx := &InferenceContext{
		Type:  "task",
		Items: []InferenceItem{},
	}

	// Get feature
	feature, err := GetFeature(database, featureID)
	if err != nil {
		return nil, fmt.Errorf("get feature: %w", err)
	}

	// Get tasks for this feature
	rows, err := database.Query(
		`SELECT id, title, content, target_file FROM tasks WHERE feature_id = ?`,
		featureID,
	)
	if err != nil {
		return nil, fmt.Errorf("get tasks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item InferenceItem
		var content, targetFile *string
		if err := rows.Scan(&item.ID, &item.Name, &content, &targetFile); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		if content != nil {
			item.Description = *content
		}
		if targetFile != nil {
			// Determine layer from file path
			item.Layer = inferLayerFromPath(*targetFile)
		}
		ctx.Items = append(ctx.Items, item)
	}

	// Get existing edges
	edges, err := GetTaskEdgesByFeature(database, featureID)
	if err == nil {
		for _, e := range edges {
			fromTask, _ := GetTask(database, e.FromTaskID)
			toTask, _ := GetTask(database, e.ToTaskID)
			if fromTask != nil && toTask != nil {
				fromID, _ := strconv.ParseInt(e.FromTaskID, 10, 64)
				toID, _ := strconv.ParseInt(e.ToTaskID, 10, 64)
				ctx.Existing = append(ctx.Existing, InferenceExisting{
					FromID:   fromID,
					FromName: fromTask.Title,
					ToID:     toID,
					ToName:   toTask.Title,
				})
			}
		}
	}

	// Build prompt
	ctx.Prompt = buildTaskInferencePrompt(feature.Name, ctx.Items, ctx.Existing)

	return ctx, nil
}

// PrepareFeatureEdgeInference prepares context for LLM to infer feature edges
func PrepareFeatureEdgeInference(database *db.DB) (*InferenceContext, error) {
	ctx := &InferenceContext{
		Type:  "feature",
		Items: []InferenceItem{},
	}

	// Get project
	project, err := GetProject(database)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}

	// Get features
	features, err := ListFeatures(database, project.ID)
	if err != nil {
		return nil, fmt.Errorf("list features: %w", err)
	}

	for _, f := range features {
		item := InferenceItem{
			ID:          f.ID,
			Name:        f.Name,
			Description: f.Spec,
		}
		ctx.Items = append(ctx.Items, item)
	}

	// Get existing edges
	rows, err := database.Query(
		`SELECT e.from_feature_id, f1.name, e.to_feature_id, f2.name
		 FROM feature_edges e
		 JOIN features f1 ON e.from_feature_id = f1.id
		 JOIN features f2 ON e.to_feature_id = f2.id`,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var existing InferenceExisting
			if err := rows.Scan(&existing.FromID, &existing.FromName, &existing.ToID, &existing.ToName); err == nil {
				ctx.Existing = append(ctx.Existing, existing)
			}
		}
	}

	// Build prompt
	ctx.Prompt = buildFeatureInferencePrompt(ctx.Items, ctx.Existing)

	return ctx, nil
}

// inferLayerFromPath determines the layer type from file path
func inferLayerFromPath(path string) string {
	lowerPath := path
	if contains(lowerPath, "model") || contains(lowerPath, "entity") {
		return "model"
	}
	if contains(lowerPath, "service") {
		return "service"
	}
	if contains(lowerPath, "api") || contains(lowerPath, "router") || contains(lowerPath, "handler") {
		return "api"
	}
	if contains(lowerPath, "component") || contains(lowerPath, "view") || contains(lowerPath, "page") {
		return "ui"
	}
	return "unknown"
}

// contains checks if a string contains a substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFold(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalFold(s, t string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if lower(s[i]) != lower(t[i]) {
			return false
		}
	}
	return true
}

func lower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}

// buildTaskInferencePrompt creates the LLM prompt for task edge inference
func buildTaskInferencePrompt(featureName string, items []InferenceItem, existing []InferenceExisting) string {
	prompt := fmt.Sprintf(`Analyze the following tasks for feature "%s" and infer dependency edges.

Tasks:
`, featureName)

	for _, item := range items {
		prompt += fmt.Sprintf("- ID %d: %s", item.ID, item.Name)
		if item.Layer != "" {
			prompt += fmt.Sprintf(" [%s]", item.Layer)
		}
		if item.Description != "" {
			prompt += fmt.Sprintf("\n  Description: %s", item.Description)
		}
		prompt += "\n"
	}

	if len(existing) > 0 {
		prompt += "\nExisting edges (do not duplicate):\n"
		for _, e := range existing {
			prompt += fmt.Sprintf("- %s (ID %d) depends on %s (ID %d)\n", e.FromName, e.FromID, e.ToName, e.ToID)
		}
	}

	prompt += `
Infer additional dependency edges based on:
1. Model tasks should be completed before service tasks
2. Service tasks should be completed before API tasks
3. API tasks should be completed before UI tasks
4. Tasks referencing the same entity may have dependencies

Return JSON array of inferred edges:
[{"from_id": <task_id>, "to_id": <task_id>, "reason": "<why this dependency exists>", "confidence": <0.0-1.0>}]`

	return prompt
}

// buildFeatureInferencePrompt creates the LLM prompt for feature edge inference
func buildFeatureInferencePrompt(items []InferenceItem, existing []InferenceExisting) string {
	prompt := "Analyze the following features and infer dependency edges.\n\nFeatures:\n"

	for _, item := range items {
		prompt += fmt.Sprintf("- ID %d: %s", item.ID, item.Name)
		if item.Description != "" {
			prompt += fmt.Sprintf("\n  Spec: %s", item.Description)
		}
		prompt += "\n"
	}

	if len(existing) > 0 {
		prompt += "\nExisting edges (do not duplicate):\n"
		for _, e := range existing {
			prompt += fmt.Sprintf("- %s (ID %d) depends on %s (ID %d)\n", e.FromName, e.FromID, e.ToName, e.ToID)
		}
	}

	prompt += `
Infer additional dependency edges based on:
1. Features providing data models should be completed before features using them
2. Core/foundation features should be completed before features building on them
3. Authentication features typically come before protected features

Return JSON array of inferred edges:
[{"from_id": <feature_id>, "to_id": <feature_id>, "reason": "<why this dependency exists>", "confidence": <0.0-1.0>}]`

	return prompt
}
