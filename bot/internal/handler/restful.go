package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"parkjunwoo.com/claribot/internal/config"
	"parkjunwoo.com/claribot/internal/message"
	"parkjunwoo.com/claribot/internal/project"
	"parkjunwoo.com/claribot/internal/schedule"
	"parkjunwoo.com/claribot/internal/task"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/pagination"
)

// --- JSON helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeResult(w http.ResponseWriter, result types.Result) {
	status := http.StatusOK
	if !result.Success {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, result)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, types.Result{Success: false, Message: msg})
}

func decodeBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// parsePage extracts page and page_size from query parameters
func (r *Router) parsePage(req *http.Request) (int, int) {
	page := 1
	pageSize := r.pageSize

	if p := req.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if n := req.URL.Query().Get("page_size"); n != "" {
		if v, err := strconv.Atoi(n); err == nil && v > 0 {
			pageSize = v
		}
	}
	if req.URL.Query().Get("all") == "true" {
		pageSize = pagination.MaxPageSize
	}
	return page, pageSize
}

// --- Project handlers ---

// ProjectStats represents task statistics for a single project
type ProjectStats struct {
	ProjectID   string     `json:"project_id"`
	ProjectName string     `json:"project_name"`
	Stats       *task.Stats `json:"stats"`
}

// HandleProjectsStats handles GET /api/projects/stats
func (r *Router) HandleProjectsStats(w http.ResponseWriter, req *http.Request) {
	projects, err := project.ListAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list projects: %v", err))
		return
	}

	var result []ProjectStats
	for _, p := range projects {
		stats, err := task.GetStats(p.Path)
		if err != nil {
			// Skip projects with no task DB or errors
			stats = &task.Stats{}
		}
		result = append(result, ProjectStats{
			ProjectID:   p.ID,
			ProjectName: p.Name,
			Stats:       stats,
		})
	}

	writeJSON(w, http.StatusOK, types.Result{
		Success: true,
		Data:    result,
	})
}

// HandleListProjects handles GET /api/projects
func (r *Router) HandleListProjects(w http.ResponseWriter, req *http.Request) {
	page, pageSize := r.parsePage(req)
	writeResult(w, project.List(pagination.NewPageRequest(page, pageSize)))
}

// HandleCreateProject handles POST /api/projects
func (r *Router) HandleCreateProject(w http.ResponseWriter, req *http.Request) {
	var body struct {
		ID          string `json:"id"`
		Path        string `json:"path"`
		Type        string `json:"type"`
		Description string `json:"description"`
	}
	if err := decodeBody(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	var result types.Result
	if body.Path != "" {
		result = project.Add(body.Path, body.Type, body.Description)
	} else if body.ID != "" {
		result = project.Create(body.ID, body.Type, body.Description)
	} else {
		writeError(w, http.StatusBadRequest, "id or path required")
		return
	}

	if result.Success && !result.NeedsInput {
		if p, ok := result.Data.(*project.Project); ok {
			r.SetProject(p.ID, p.Path, p.Description)
		}
	}
	status := http.StatusCreated
	if !result.Success {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, result)
}

// HandleGetProject handles GET /api/projects/{id}
func (r *Router) HandleGetProject(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "project id required")
		return
	}
	writeResult(w, project.Get(id))
}

// HandleDeleteProject handles DELETE /api/projects/{id}
func (r *Router) HandleDeleteProject(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "project id required")
		return
	}
	result := project.Delete(id, true)
	ctx := r.SnapshotContext()
	if result.Success && ctx.ProjectID == id {
		r.SetProject("", "", "")
	}
	writeResult(w, result)
}

// HandleSwitchProject handles POST /api/projects/{id}/switch
func (r *Router) HandleSwitchProject(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "project id required")
		return
	}
	if id == "none" {
		r.SetProject("", project.DefaultPath, "글로벌 모드")
		writeJSON(w, http.StatusOK, types.Result{
			Success: true,
			Message: "프로젝트 선택 해제됨 (글로벌 모드)",
		})
		return
	}
	result := project.Switch(id)
	if result.Success {
		if p, ok := result.Data.(*project.Project); ok {
			r.SetProject(p.ID, p.Path, p.Description)
		}
	}
	writeResult(w, result)
}

// HandleSetProject handles PATCH /api/projects/{id}
func (r *Router) HandleSetProject(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "project id required")
		return
	}
	var body struct {
		Field string `json:"field"`
		Value string `json:"value"`
	}
	if err := decodeBody(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.Field == "" {
		writeError(w, http.StatusBadRequest, "field required")
		return
	}
	writeResult(w, project.Set(id, body.Field, body.Value))
}

// --- Task handlers ---

// HandleListTasks handles GET /api/tasks
func (r *Router) HandleListTasks(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}

	if req.URL.Query().Get("tree") == "true" {
		writeResult(w, task.ListTree(ctx.ProjectPath))
		return
	}

	var parentID *int
	if p := req.URL.Query().Get("parent_id"); p != "" {
		if pid, err := strconv.Atoi(p); err == nil {
			parentID = &pid
		}
	}
	page, pageSize := r.parsePage(req)
	writeResult(w, task.List(ctx.ProjectPath, parentID, pagination.NewPageRequest(page, pageSize)))
}

// HandleAddTask handles POST /api/tasks
func (r *Router) HandleAddTask(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	var body struct {
		Title    string `json:"title"`
		ParentID *int   `json:"parent_id"`
		Spec     string `json:"spec"`
	}
	if err := decodeBody(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title required")
		return
	}
	result := task.Add(ctx.ProjectPath, body.Title, body.ParentID, body.Spec)
	status := http.StatusCreated
	if !result.Success {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, result)
}

// HandleGetTask handles GET /api/tasks/{id}
func (r *Router) HandleGetTask(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "task id required")
		return
	}
	writeResult(w, task.Get(ctx.ProjectPath, id))
}

// HandleUpdateTask handles PATCH /api/tasks/{id}
func (r *Router) HandleUpdateTask(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "task id required")
		return
	}
	var body struct {
		Field string `json:"field"`
		Value string `json:"value"`
	}
	if err := decodeBody(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.Field == "" {
		writeError(w, http.StatusBadRequest, "field required")
		return
	}
	writeResult(w, task.Set(ctx.ProjectPath, id, body.Field, body.Value))
}

// HandleDeleteTask handles DELETE /api/tasks/{id}
func (r *Router) HandleDeleteTask(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "task id required")
		return
	}
	writeResult(w, task.Delete(ctx.ProjectPath, id, true))
}

// HandlePlanTask handles POST /api/tasks/{id}/plan
func (r *Router) HandlePlanTask(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	id := req.PathValue("id")
	writeResult(w, task.Plan(ctx.ProjectPath, id))
}

// HandleRunTask handles POST /api/tasks/{id}/run
func (r *Router) HandleRunTask(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	id := req.PathValue("id")
	writeResult(w, task.Run(ctx.ProjectPath, id))
}

// HandlePlanAllTasks handles POST /api/tasks/plan-all
func (r *Router) HandlePlanAllTasks(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	writeResult(w, task.PlanAll(ctx.ProjectPath))
}

// HandleRunAllTasks handles POST /api/tasks/run-all
func (r *Router) HandleRunAllTasks(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	writeResult(w, task.RunAll(ctx.ProjectPath))
}

// HandleCycleTasks handles POST /api/tasks/cycle
func (r *Router) HandleCycleTasks(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	writeResult(w, task.Cycle(ctx.ProjectPath))
}

// HandleStopTask handles POST /api/tasks/stop
func (r *Router) HandleStopTask(w http.ResponseWriter, req *http.Request) {
	msg, running := task.Stop()
	writeResult(w, types.Result{Success: running, Message: msg})
}

// --- Message handlers ---

// HandleListMessages handles GET /api/messages
func (r *Router) HandleListMessages(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	projectPath := ctx.ProjectPath
	if projectPath == "" {
		projectPath = project.DefaultPath
	}
	if projectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트 경로가 설정되지 않았습니다")
		return
	}
	page, pageSize := r.parsePage(req)
	writeResult(w, message.List(projectPath, pagination.NewPageRequest(page, pageSize)))
}

// HandleSendMessage handles POST /api/messages
func (r *Router) HandleSendMessage(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	projectPath := ctx.ProjectPath
	if projectPath == "" {
		projectPath = project.DefaultPath
	}
	if projectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트 경로가 설정되지 않았습니다")
		return
	}
	var body struct {
		Content string `json:"content"`
		Source  string `json:"source"`
	}
	if err := decodeBody(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.Content == "" {
		writeError(w, http.StatusBadRequest, "content required")
		return
	}
	if body.Source == "" {
		body.Source = "api"
	}
	result := message.Send(projectPath, body.Content, body.Source)
	status := http.StatusCreated
	if !result.Success {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, result)
}

// HandleGetMessage handles GET /api/messages/{id}
func (r *Router) HandleGetMessage(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	projectPath := ctx.ProjectPath
	if projectPath == "" {
		projectPath = project.DefaultPath
	}
	if projectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트 경로가 설정되지 않았습니다")
		return
	}
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "message id required")
		return
	}
	writeResult(w, message.Get(projectPath, id))
}

// HandleMessageStatus handles GET /api/messages/status
func (r *Router) HandleMessageStatus(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	projectPath := ctx.ProjectPath
	if projectPath == "" {
		projectPath = project.DefaultPath
	}
	if projectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트 경로가 설정되지 않았습니다")
		return
	}
	writeResult(w, message.Status(projectPath))
}

// HandleMessageProcessing handles GET /api/messages/processing
func (r *Router) HandleMessageProcessing(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	projectPath := ctx.ProjectPath
	if projectPath == "" {
		projectPath = project.DefaultPath
	}
	if projectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트 경로가 설정되지 않았습니다")
		return
	}
	writeResult(w, message.Processing(projectPath))
}

// --- Config handlers ---

// HandleListConfigs handles GET /api/configs
func (r *Router) HandleListConfigs(w http.ResponseWriter, req *http.Request) {
	page, pageSize := r.parsePage(req)
	writeResult(w, config.ListDBConfig(pagination.NewPageRequest(page, pageSize)))
}

// HandleGetConfig handles GET /api/configs/{key}
func (r *Router) HandleGetConfig(w http.ResponseWriter, req *http.Request) {
	key := req.PathValue("key")
	if key == "" {
		writeError(w, http.StatusBadRequest, "config key required")
		return
	}
	writeResult(w, config.GetDBConfig(key))
}

// HandleSetConfig handles PUT /api/configs/{key}
func (r *Router) HandleSetConfig(w http.ResponseWriter, req *http.Request) {
	key := req.PathValue("key")
	if key == "" {
		writeError(w, http.StatusBadRequest, "config key required")
		return
	}
	var body struct {
		Value string `json:"value"`
	}
	if err := decodeBody(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	writeResult(w, config.SetDBConfig(key, body.Value))
}

// HandleDeleteConfig handles DELETE /api/configs/{key}
func (r *Router) HandleDeleteConfig(w http.ResponseWriter, req *http.Request) {
	key := req.PathValue("key")
	if key == "" {
		writeError(w, http.StatusBadRequest, "config key required")
		return
	}
	writeResult(w, config.DeleteDBConfig(key, true))
}

// --- Schedule handlers ---

// HandleListSchedules handles GET /api/schedules
func (r *Router) HandleListSchedules(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	showAll := req.URL.Query().Get("all") == "true"
	page, pageSize := r.parsePage(req)

	var projectID *string
	if !showAll && ctx.ProjectID != "" {
		projectID = &ctx.ProjectID
	}
	// Override with explicit query param
	if pid := req.URL.Query().Get("project_id"); pid != "" {
		if pid == "none" {
			projectID = nil
		} else {
			projectID = &pid
		}
	}
	writeResult(w, schedule.List(projectID, showAll, pagination.NewPageRequest(page, pageSize)))
}

// HandleAddSchedule handles POST /api/schedules
func (r *Router) HandleAddSchedule(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	var body struct {
		CronExpr  string  `json:"cron_expr"`
		Message   string  `json:"message"`
		Type      string  `json:"type"`
		ProjectID *string `json:"project_id"`
		RunOnce   bool    `json:"run_once"`
	}
	if err := decodeBody(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.CronExpr == "" || body.Message == "" {
		writeError(w, http.StatusBadRequest, "cron_expr and message required")
		return
	}
	if body.ProjectID == nil && ctx.ProjectID != "" {
		body.ProjectID = &ctx.ProjectID
	}
	result := schedule.Add(body.CronExpr, body.Message, body.ProjectID, body.RunOnce, body.Type)
	status := http.StatusCreated
	if !result.Success {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, result)
}

// HandleGetSchedule handles GET /api/schedules/{id}
func (r *Router) HandleGetSchedule(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "schedule id required")
		return
	}
	writeResult(w, schedule.Get(id))
}

// HandleDeleteSchedule handles DELETE /api/schedules/{id}
func (r *Router) HandleDeleteSchedule(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "schedule id required")
		return
	}
	writeResult(w, schedule.Delete(id, true))
}

// HandleEnableSchedule handles POST /api/schedules/{id}/enable
func (r *Router) HandleEnableSchedule(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "schedule id required")
		return
	}
	writeResult(w, schedule.Enable(id))
}

// HandleDisableSchedule handles POST /api/schedules/{id}/disable
func (r *Router) HandleDisableSchedule(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "schedule id required")
		return
	}
	writeResult(w, schedule.Disable(id))
}

// HandleUpdateSchedule handles PATCH /api/schedules/{id}
func (r *Router) HandleUpdateSchedule(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "schedule id required")
		return
	}
	var body struct {
		Field string `json:"field"`
		Value string `json:"value"`
	}
	if err := decodeBody(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.Field == "project" {
		var projectID *string
		if body.Value != "" && body.Value != "none" {
			projectID = &body.Value
		}
		writeResult(w, schedule.SetProject(id, projectID))
		return
	}
	writeError(w, http.StatusBadRequest, fmt.Sprintf("unsupported field: %s", body.Field))
}

// HandleScheduleRuns handles GET /api/schedules/{id}/runs
func (r *Router) HandleScheduleRuns(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "schedule id required")
		return
	}
	page, pageSize := r.parsePage(req)
	writeResult(w, schedule.Runs(id, pagination.NewPageRequest(page, pageSize)))
}

// HandleScheduleRunDetail handles GET /api/schedules/runs/{runId}
func (r *Router) HandleScheduleRunDetail(w http.ResponseWriter, req *http.Request) {
	runID := req.PathValue("runId")
	if runID == "" {
		writeError(w, http.StatusBadRequest, "run id required")
		return
	}
	writeResult(w, schedule.Run(runID))
}

// --- Status handler ---

// HandleStatus handles GET /api/status
func (r *Router) HandleStatus(w http.ResponseWriter, req *http.Request) {
	ctx := r.SnapshotContext()
	writeResult(w, r.handleStatus(ctx))
}

// RegisterRESTfulRoutes registers all RESTful API endpoints on the given mux.
// The handlers are methods on Router so they share the same project context.
func (r *Router) RegisterRESTfulRoutes(mux *http.ServeMux) {
	// Status
	mux.HandleFunc("GET /api/status", r.HandleStatus)

	// Projects - specific routes before parameterized
	mux.HandleFunc("GET /api/projects/stats", r.HandleProjectsStats)
	mux.HandleFunc("GET /api/projects", r.HandleListProjects)
	mux.HandleFunc("POST /api/projects", r.HandleCreateProject)
	mux.HandleFunc("GET /api/projects/{id}", r.HandleGetProject)
	mux.HandleFunc("PATCH /api/projects/{id}", r.HandleSetProject)
	mux.HandleFunc("DELETE /api/projects/{id}", r.HandleDeleteProject)
	mux.HandleFunc("POST /api/projects/{id}/switch", r.HandleSwitchProject)

	// Tasks - action routes MUST be registered before parameterized routes
	mux.HandleFunc("POST /api/tasks/plan-all", r.HandlePlanAllTasks)
	mux.HandleFunc("POST /api/tasks/run-all", r.HandleRunAllTasks)
	mux.HandleFunc("POST /api/tasks/cycle", r.HandleCycleTasks)
	mux.HandleFunc("POST /api/tasks/stop", r.HandleStopTask)
	mux.HandleFunc("GET /api/tasks", r.HandleListTasks)
	mux.HandleFunc("POST /api/tasks", r.HandleAddTask)
	mux.HandleFunc("GET /api/tasks/{id}", r.HandleGetTask)
	mux.HandleFunc("PATCH /api/tasks/{id}", r.HandleUpdateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", r.HandleDeleteTask)
	mux.HandleFunc("POST /api/tasks/{id}/plan", r.HandlePlanTask)
	mux.HandleFunc("POST /api/tasks/{id}/run", r.HandleRunTask)

	// Messages - specific routes before parameterized
	mux.HandleFunc("GET /api/messages/status", r.HandleMessageStatus)
	mux.HandleFunc("GET /api/messages/processing", r.HandleMessageProcessing)
	mux.HandleFunc("GET /api/messages", r.HandleListMessages)
	mux.HandleFunc("POST /api/messages", r.HandleSendMessage)
	mux.HandleFunc("GET /api/messages/{id}", r.HandleGetMessage)

	// Configs
	mux.HandleFunc("GET /api/configs", r.HandleListConfigs)
	mux.HandleFunc("GET /api/configs/{key}", r.HandleGetConfig)
	mux.HandleFunc("PUT /api/configs/{key}", r.HandleSetConfig)
	mux.HandleFunc("DELETE /api/configs/{key}", r.HandleDeleteConfig)

	// Schedules
	mux.HandleFunc("GET /api/schedules", r.HandleListSchedules)
	mux.HandleFunc("POST /api/schedules", r.HandleAddSchedule)
	mux.HandleFunc("GET /api/schedules/{id}", r.HandleGetSchedule)
	mux.HandleFunc("PATCH /api/schedules/{id}", r.HandleUpdateSchedule)
	mux.HandleFunc("DELETE /api/schedules/{id}", r.HandleDeleteSchedule)
	mux.HandleFunc("POST /api/schedules/{id}/enable", r.HandleEnableSchedule)
	mux.HandleFunc("POST /api/schedules/{id}/disable", r.HandleDisableSchedule)
	mux.HandleFunc("GET /api/schedules/{id}/runs", r.HandleScheduleRuns)

	// Schedule runs (separate path to avoid conflict with /api/schedules/{id}/runs)
	mux.HandleFunc("GET /api/schedule-runs/{runId}", r.HandleScheduleRunDetail)
}

