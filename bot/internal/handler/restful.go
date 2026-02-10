package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"parkjunwoo.com/claribot/internal/config"
	"parkjunwoo.com/claribot/internal/message"
	"parkjunwoo.com/claribot/internal/project"
	"parkjunwoo.com/claribot/internal/schedule"
	"parkjunwoo.com/claribot/internal/spec"
	"parkjunwoo.com/claribot/internal/task"
	"parkjunwoo.com/claribot/internal/terminal"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/claude"
	"parkjunwoo.com/claribot/pkg/logger"
	"parkjunwoo.com/claribot/pkg/pagination"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Auth is handled by middleware before upgrade
	},
}

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

// getContextFromRequest returns context based on X-Clari-Cwd header or falls back to global context.
// If cwd is inside a registered project path, use that project.
func (r *Router) getContextFromRequest(req *http.Request) *Context {
	cwd := req.Header.Get("X-Clari-Cwd")
	if cwd == "" {
		return r.SnapshotContext()
	}

	// Find project by cwd
	projects, err := project.ListAll()
	if err != nil {
		return r.SnapshotContext()
	}

	// Check if cwd starts with any project path
	for _, p := range projects {
		if cwd == p.Path || (len(cwd) > len(p.Path) && cwd[:len(p.Path)+1] == p.Path+string(filepath.Separator)) {
			return &Context{
				ProjectID:          p.ID,
				ProjectPath:        p.Path,
				ProjectDescription: p.Description,
			}
		}
	}

	return r.SnapshotContext()
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
	ProjectID          string      `json:"project_id"`
	ProjectName        string      `json:"project_name"`
	ProjectDescription string      `json:"project_description"`
	Stats              *task.Stats `json:"stats"`
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
			ProjectID:          p.ID,
			ProjectName:        p.Name,
			ProjectDescription: p.Description,
			Stats:              stats,
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
		Description string `json:"description"`
	}
	if err := decodeBody(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	var result types.Result
	if body.Path != "" {
		result = project.Add(body.Path, body.Description)
	} else if body.ID != "" {
		result = project.Create(body.ID, body.Description)
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
	ctx := r.getContextFromRequest(req)
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
// Supports two formats:
// 1. Single field: {"field": "parallel", "value": "2"}
// 2. Multiple fields: {"description": "...", "parallel": 2}
func (r *Router) HandleSetProject(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "project id required")
		return
	}
	var body struct {
		// Single field format
		Field string `json:"field"`
		Value string `json:"value"`
		// Multiple fields format
		Description *string `json:"description"`
		Parallel    *int    `json:"parallel"`
	}
	if err := decodeBody(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	// Handle single field format
	if body.Field != "" {
		writeResult(w, project.Set(id, body.Field, body.Value))
		return
	}

	// Handle multiple fields format
	var results []string
	var lastErr error

	if body.Description != nil {
		result := project.Set(id, "description", *body.Description)
		if !result.Success {
			lastErr = fmt.Errorf("%s", result.Message)
		} else {
			results = append(results, "description")
		}
	}

	if body.Parallel != nil {
		result := project.Set(id, "parallel", fmt.Sprintf("%d", *body.Parallel))
		if !result.Success {
			lastErr = fmt.Errorf("%s", result.Message)
		} else {
			results = append(results, "parallel")
		}
	}

	if len(results) == 0 && lastErr == nil {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	if lastErr != nil {
		writeJSON(w, http.StatusBadRequest, types.Result{
			Success: false,
			Message: lastErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, types.Result{
		Success: true,
		Message: fmt.Sprintf("Updated: %v", results),
	})
}

// --- Task handlers ---

// HandleListTasks handles GET /api/tasks
func (r *Router) HandleListTasks(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)
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
	ctx := r.getContextFromRequest(req)
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
	if body.Title == "" && body.Spec == "" {
		writeError(w, http.StatusBadRequest, "title or spec required")
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
	ctx := r.getContextFromRequest(req)
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
	ctx := r.getContextFromRequest(req)
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
	ctx := r.getContextFromRequest(req)
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
	ctx := r.getContextFromRequest(req)
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	id := req.PathValue("id")
	writeResult(w, task.Plan(ctx.ProjectPath, id))
}

// HandleRunTask handles POST /api/tasks/{id}/run
func (r *Router) HandleRunTask(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	id := req.PathValue("id")
	writeResult(w, task.Run(ctx.ProjectPath, id))
}

// HandlePlanAllTasks handles POST /api/tasks/plan-all
func (r *Router) HandlePlanAllTasks(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	writeResult(w, task.PlanAll(ctx.ProjectPath))
}

// HandleRunAllTasks handles POST /api/tasks/run-all
func (r *Router) HandleRunAllTasks(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	writeResult(w, task.RunAll(ctx.ProjectPath))
}

// HandleCycleTasks handles POST /api/tasks/cycle
func (r *Router) HandleCycleTasks(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)

	// Accept optional project_id from body
	var body struct {
		ProjectID string `json:"project_id"`
	}
	if req.Body != nil {
		decodeBody(req, &body)
	}

	projectPath := ctx.ProjectPath
	if body.ProjectID != "" {
		result := project.Get(body.ProjectID)
		if !result.Success {
			writeError(w, http.StatusBadRequest, "프로젝트를 찾을 수 없습니다: "+body.ProjectID)
			return
		}
		if p, ok := result.Data.(*project.Project); ok {
			projectPath = p.Path
			r.SetProject(p.ID, p.Path, p.Description)
		}
	}

	if projectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	writeResult(w, task.Cycle(projectPath))
}

// HandleStopTask handles POST /api/tasks/stop
func (r *Router) HandleStopTask(w http.ResponseWriter, req *http.Request) {
	msg, running := task.Stop()
	writeResult(w, types.Result{Success: running, Message: msg})
}

// --- Message handlers ---

// HandleListMessages handles GET /api/messages
func (r *Router) HandleListMessages(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)
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
	writeResult(w, message.List(projectID, showAll, pagination.NewPageRequest(page, pageSize)))
}

// HandleSendMessage handles POST /api/messages
func (r *Router) HandleSendMessage(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)

	var body struct {
		Content   string  `json:"content"`
		Source    string  `json:"source"`
		ProjectID *string `json:"project_id"`
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

	// Resolve project path: body.project_id takes priority over context
	projectPath := ctx.ProjectPath
	var projectID *string
	if body.ProjectID != nil && *body.ProjectID != "" {
		result := project.Get(*body.ProjectID)
		if !result.Success {
			writeError(w, http.StatusBadRequest, "프로젝트를 찾을 수 없습니다: "+*body.ProjectID)
			return
		}
		if p, ok := result.Data.(*project.Project); ok {
			projectPath = p.Path
			projectID = &p.ID
		}
	} else if ctx.ProjectID != "" {
		projectID = &ctx.ProjectID
	}

	if projectPath == "" {
		projectPath = project.DefaultPath
	}
	// Fallback to home directory for GLOBAL mode
	if projectPath == "" {
		if home, err := os.UserHomeDir(); err == nil {
			projectPath = filepath.Join(home, ".claribot")
		}
	}
	if projectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트 경로가 설정되지 않았습니다")
		return
	}

	// PTY path: if termManager exists and project has a terminal session
	usePTY := false
	if r.termManager != nil && projectID != nil && *projectID != "" {
		session := r.termManager.GetSession(*projectID)
		if session == nil {
			// Auto-create terminal session with claude
			initialCmd := fmt.Sprintf("cd %s && claude --dangerously-skip-permissions", projectPath)
			var createErr error
			session, _, createErr = r.termManager.GetOrCreate(*projectID, 120, 40, projectPath, initialCmd)
			if createErr != nil {
				logger.Debug("[message] PTY session creation failed, falling back to sync: %v", createErr)
			} else {
				// Wait for claude to start
				time.Sleep(2 * time.Second)
				usePTY = true
			}
		} else {
			usePTY = true
		}
		if usePTY {
			result := message.SendViaPTY(projectID, projectPath, body.Content, body.Source, session)
			status := http.StatusCreated
			if !result.Success {
				status = http.StatusBadRequest
			}
			writeJSON(w, status, result)
			return
		}
	}

	// Sync path (claude.Run)
	result := message.SendWithProject(projectID, projectPath, body.Content, body.Source)
	status := http.StatusCreated
	if !result.Success {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, result)
}

// HandleGetMessage handles GET /api/messages/{id}
func (r *Router) HandleGetMessage(w http.ResponseWriter, req *http.Request) {
	// Messages are stored in global DB, no project path required
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "message id required")
		return
	}
	writeResult(w, message.Get("", id))
}

// HandleMessageStatus handles GET /api/messages/status
func (r *Router) HandleMessageStatus(w http.ResponseWriter, req *http.Request) {
	// Messages are stored in global DB, no project path required
	writeResult(w, message.Status(""))
}

// HandleMessageProcessing handles GET /api/messages/processing
func (r *Router) HandleMessageProcessing(w http.ResponseWriter, req *http.Request) {
	// Messages are stored in global DB, no project path required
	writeResult(w, message.Processing(""))
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

// HandleGetConfigYaml handles GET /api/config-yaml
func (r *Router) HandleGetConfigYaml(w http.ResponseWriter, req *http.Request) {
	content, err := config.ReadRaw()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, types.Result{
		Success: true,
		Data:    content,
	})
}

// HandleSetConfigYaml handles PUT /api/config-yaml
func (r *Router) HandleSetConfigYaml(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Content string `json:"content"`
	}
	if err := decodeBody(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := config.WriteRaw(body.Content); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, types.Result{
		Success: true,
		Message: "Config saved. Restart claribot to apply changes.",
	})
}

// --- Schedule handlers ---

// HandleListSchedules handles GET /api/schedules
func (r *Router) HandleListSchedules(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)
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
	ctx := r.getContextFromRequest(req)
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

// --- Usage handler ---

// UsageResponse represents the usage API response
type UsageResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Live      string `json:"live,omitempty"`       // Real-time rate limit from /usage
	UpdatedAt string `json:"updated_at,omitempty"` // When live data was last fetched
}

// HandleGetUsage handles GET /api/usage
func (r *Router) HandleGetUsage(w http.ResponseWriter, req *http.Request) {
	// Get stats from stats-cache.json
	stats, err := claude.GetUsage()
	var statsOutput string
	if err == nil {
		statsOutput = claude.FormatUsage(stats)
	}

	// Get cached live usage from ~/.claribot/claude-usage.txt
	liveOutput, updatedAt, _ := claude.GetUsageLive()

	resp := UsageResponse{
		Success: true,
		Message: statsOutput,
		Live:    liveOutput,
	}
	if !updatedAt.IsZero() {
		resp.UpdatedAt = updatedAt.Format(time.RFC3339)
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleRefreshUsage handles POST /api/usage/refresh
func (r *Router) HandleRefreshUsage(w http.ResponseWriter, req *http.Request) {
	// Trigger async refresh
	claude.RefreshUsageLiveAsync()
	writeJSON(w, http.StatusAccepted, types.Result{
		Success: true,
		Message: "Usage refresh started",
	})
}

// --- Status handler ---

// CycleStatusJSON represents cycle status as structured JSON for the API
type CycleStatusJSON struct {
	Status        string `json:"status"`
	Type          string `json:"type,omitempty"`
	ProjectID     string `json:"project_id,omitempty"`
	StartedAt     string `json:"started_at,omitempty"`
	CurrentTaskID int    `json:"current_task_id,omitempty"`
	ActiveWorkers int    `json:"active_workers,omitempty"`
	Phase         string `json:"phase,omitempty"`
	TargetTotal   int    `json:"target_total,omitempty"`
	Completed     int    `json:"completed,omitempty"`
	ElapsedSec    int    `json:"elapsed_sec,omitempty"`
}

// HandleStatus handles GET /api/status
func (r *Router) HandleStatus(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)
	result := r.handleStatus(ctx)

	// Build cycle_status JSON (backward compatibility - first active cycle)
	cycleInfo := task.GetCycleStatus()
	csJSON := CycleStatusJSON{
		Status: cycleInfo.Status,
	}
	if cycleInfo.Status != "idle" {
		csJSON.Type = cycleInfo.Type
		csJSON.ProjectID = cycleInfo.ProjectID
		if !cycleInfo.StartedAt.IsZero() {
			csJSON.StartedAt = cycleInfo.StartedAt.Format(time.RFC3339)
			csJSON.ElapsedSec = int(time.Since(cycleInfo.StartedAt).Seconds())
		}
		csJSON.CurrentTaskID = cycleInfo.CurrentTaskID
		csJSON.ActiveWorkers = cycleInfo.ActiveWorkers
		csJSON.Phase = cycleInfo.Phase
		csJSON.TargetTotal = cycleInfo.TargetTotal
		csJSON.Completed = cycleInfo.Completed
	}

	// Build all cycle statuses
	allCycles := task.GetAllCycleStatuses()
	var cycleStatusesJSON []CycleStatusJSON
	for _, cs := range allCycles {
		cj := CycleStatusJSON{
			Status:        cs.Status,
			Type:          cs.Type,
			ProjectID:     cs.ProjectID,
			CurrentTaskID: cs.CurrentTaskID,
			ActiveWorkers: cs.ActiveWorkers,
			Phase:         cs.Phase,
			TargetTotal:   cs.TargetTotal,
			Completed:     cs.Completed,
		}
		if !cs.StartedAt.IsZero() {
			cj.StartedAt = cs.StartedAt.Format(time.RFC3339)
			cj.ElapsedSec = int(time.Since(cs.StartedAt).Seconds())
		}
		cycleStatusesJSON = append(cycleStatusesJSON, cj)
	}

	// Build task_stats for current project
	var taskStats *task.Stats
	if ctx.ProjectPath != "" {
		if s, err := task.GetStats(ctx.ProjectPath); err == nil {
			taskStats = s
		}
	}

	// Wrap into enriched response
	resp := map[string]interface{}{
		"success":        result.Success,
		"message":        result.Message,
		"data":           result.Data,
		"project_id":     ctx.ProjectID,
		"cycle_status":   csJSON,
		"cycle_statuses": cycleStatusesJSON,
	}
	if taskStats != nil {
		resp["task_stats"] = taskStats
	}

	writeJSON(w, http.StatusOK, resp)
}

// --- Spec handlers ---

// HandleListSpecs handles GET /api/specs
func (r *Router) HandleListSpecs(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	page, pageSize := r.parsePage(req)
	writeResult(w, spec.List(ctx.ProjectPath, pagination.NewPageRequest(page, pageSize)))
}

// HandleAddSpec handles POST /api/specs
func (r *Router) HandleAddSpec(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := decodeBody(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title required")
		return
	}
	result := spec.Add(ctx.ProjectPath, body.Title, body.Content)
	status := http.StatusCreated
	if !result.Success {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, result)
}

// HandleGetSpec handles GET /api/specs/{id}
func (r *Router) HandleGetSpec(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "spec id required")
		return
	}
	writeResult(w, spec.Get(ctx.ProjectPath, id))
}

// HandleUpdateSpec handles PATCH /api/specs/{id}
func (r *Router) HandleUpdateSpec(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "spec id required")
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
	writeResult(w, spec.Set(ctx.ProjectPath, id, body.Field, body.Value))
}

// HandleDeleteSpec handles DELETE /api/specs/{id}
func (r *Router) HandleDeleteSpec(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}
	id := req.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "spec id required")
		return
	}
	writeResult(w, spec.Delete(ctx.ProjectPath, id, true))
}

// --- File handlers ---

// fileItem represents a file or directory entry in the API response
type fileItem struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // "file" or "dir"
	Size     int64  `json:"size"`
	Ext      string `json:"ext,omitempty"`
	Modified string `json:"modified"`
}

// sensitiveExts contains file extensions that should be blocked from reading
var sensitiveExts = map[string]bool{
	".env": true, ".key": true, ".pem": true, ".p12": true, ".pfx": true,
	".jks": true, ".keystore": true, ".secret": true, ".credentials": true,
}

// skipDirs contains directory names to skip when listing files
var skipDirs = map[string]bool{
	".git": true, "node_modules": true, ".next": true, "dist": true,
	"build": true, "__pycache__": true, ".cache": true, "vendor": true,
	".idea": true, ".vscode": true, "bin": true, ".claribot": true,
}

// validateFilePath checks for path traversal attacks and returns the safe absolute path.
// Rejects: absolute paths, ".." components, symlinks escaping project root.
func validateFilePath(projectPath, relPath string) (string, error) {
	if relPath == "" {
		relPath = "."
	}
	// Reject absolute paths
	if filepath.IsAbs(relPath) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}
	// Reject ".." components
	cleaned := filepath.Clean(relPath)
	for _, part := range filepath.SplitList(cleaned) {
		if part == ".." {
			return "", fmt.Errorf("path traversal is not allowed")
		}
	}
	// filepath.Clean may still produce ".." — check explicitly
	if cleaned == ".." || len(cleaned) >= 3 && cleaned[:3] == ".."+string(filepath.Separator) {
		return "", fmt.Errorf("path traversal is not allowed")
	}
	// Also check within the path
	if cleaned != "." {
		parts := splitPath(cleaned)
		for _, p := range parts {
			if p == ".." {
				return "", fmt.Errorf("path traversal is not allowed")
			}
		}
	}

	absProject, err := filepath.Abs(projectPath)
	if err != nil {
		return "", fmt.Errorf("invalid project path")
	}

	var absTarget string
	if cleaned == "." {
		absTarget = absProject
	} else {
		absTarget = filepath.Join(absProject, cleaned)
	}

	// Resolve symlinks and verify still within project
	realTarget, err := filepath.EvalSymlinks(absTarget)
	if err != nil {
		// If the target doesn't exist, check the parent for symlink escape
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path not found")
		}
		return "", fmt.Errorf("invalid path")
	}
	realProject, _ := filepath.EvalSymlinks(absProject)
	if realTarget != realProject && !isSubPath(realProject, realTarget) {
		return "", fmt.Errorf("path is outside project directory")
	}

	return realTarget, nil
}

// splitPath splits a filepath into its components
func splitPath(p string) []string {
	var parts []string
	for {
		dir, file := filepath.Split(p)
		if file != "" {
			parts = append([]string{file}, parts...)
		}
		if dir == "" || dir == p {
			break
		}
		p = filepath.Clean(dir)
	}
	return parts
}

// isSubPath checks if child is a subdirectory/file of parent
func isSubPath(parent, child string) bool {
	return len(child) > len(parent) && child[:len(parent)+1] == parent+string(filepath.Separator)
}

// isBinaryContent detects if content is binary by checking for null bytes in first 512 bytes
func isBinaryContent(content []byte) bool {
	checkLen := 512
	if len(content) < checkLen {
		checkLen = len(content)
	}
	for i := 0; i < checkLen; i++ {
		if content[i] == 0 {
			return true
		}
	}
	return false
}

// isSensitiveFile checks if a filename has a sensitive extension
func isSensitiveFile(name string) bool {
	ext := filepath.Ext(name)
	if sensitiveExts[ext] {
		return true
	}
	// Also check for .env.* patterns (e.g. .env.local, .env.production)
	if name == ".env" || (len(name) > 4 && name[:4] == ".env") {
		return true
	}
	return false
}

// HandleListFiles handles GET /api/files?path=<relative_path>
// Returns 1-depth directory listing (lazy loading) for the current project.
func (r *Router) HandleListFiles(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}

	relPath := req.URL.Query().Get("path")
	if relPath == "" {
		relPath = "."
	}

	absDir, err := validateFilePath(ctx.ProjectPath, relPath)
	if err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	info, err := os.Stat(absDir)
	if err != nil {
		writeError(w, http.StatusNotFound, "directory not found")
		return
	}
	if !info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is not a directory")
		return
	}

	dirEntries, err := os.ReadDir(absDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read directory: "+err.Error())
		return
	}

	// Build items: 1 depth only, sorted folders first then by name
	var dirs []fileItem
	var files []fileItem
	for _, entry := range dirEntries {
		name := entry.Name()

		// Skip .git and other noisy directories
		if entry.IsDir() && skipDirs[name] {
			continue
		}
		// Skip hidden files/dirs starting with dot (except known config files)
		if len(name) > 0 && name[0] == '.' && skipDirs[name] {
			continue
		}

		entryInfo, err := entry.Info()
		if err != nil {
			continue
		}

		item := fileItem{
			Name:     name,
			Size:     entryInfo.Size(),
			Modified: entryInfo.ModTime().Format(time.RFC3339),
		}

		if entry.IsDir() {
			item.Type = "dir"
			dirs = append(dirs, item)
		} else {
			item.Type = "file"
			item.Ext = filepath.Ext(name)
			files = append(files, item)
		}
	}

	// Sort each group by name
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name < dirs[j].Name })
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })

	// Folders first, then files
	items := append(dirs, files...)

	writeJSON(w, http.StatusOK, types.Result{
		Success: true,
		Data: map[string]interface{}{
			"path":  relPath,
			"items": items,
		},
	})
}

// HandleFileContent handles GET /api/files/content?path=<relative_path>
// Returns the text content of a file within the project.
func (r *Router) HandleFileContent(w http.ResponseWriter, req *http.Request) {
	ctx := r.getContextFromRequest(req)
	if ctx.ProjectPath == "" {
		writeError(w, http.StatusBadRequest, "프로젝트를 먼저 선택하세요")
		return
	}

	filePath := req.URL.Query().Get("path")
	if filePath == "" {
		writeError(w, http.StatusBadRequest, "path parameter required")
		return
	}

	absPath, err := validateFilePath(ctx.ProjectPath, filePath)
	if err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}
	if info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is a directory, not a file")
		return
	}

	// Block sensitive files
	if isSensitiveFile(info.Name()) {
		writeError(w, http.StatusForbidden, "access to sensitive files is not allowed")
		return
	}

	ext := filepath.Ext(info.Name())

	// Check file size (limit to 1MB)
	const maxFileSize = 1 << 20
	if info.Size() > maxFileSize {
		writeJSON(w, http.StatusOK, types.Result{
			Success: true,
			Data: map[string]interface{}{
				"path":    filePath,
				"content": "",
				"size":    info.Size(),
				"ext":     ext,
				"binary":  true,
			},
		})
		return
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read file: "+err.Error())
		return
	}

	binary := isBinaryContent(content)
	var contentStr string
	if !binary {
		contentStr = string(content)
	}

	writeJSON(w, http.StatusOK, types.Result{
		Success: true,
		Data: map[string]interface{}{
			"path":    filePath,
			"content": contentStr,
			"size":    info.Size(),
			"ext":     ext,
			"binary":  binary,
		},
	})
}

// HandleTerminalWS handles GET /api/terminal/ws → WebSocket upgrade
func (r *Router) HandleTerminalWS(w http.ResponseWriter, req *http.Request) {
	if r.termManager == nil {
		writeError(w, http.StatusServiceUnavailable, "terminal not available")
		return
	}

	cols := uint16(80)
	rows := uint16(24)
	if c := req.URL.Query().Get("cols"); c != "" {
		if v, err := strconv.ParseUint(c, 10, 16); err == nil && v > 0 {
			cols = uint16(v)
		}
	}
	if rr := req.URL.Query().Get("rows"); rr != "" {
		if v, err := strconv.ParseUint(rr, 10, 16); err == nil && v > 0 {
			rows = uint16(v)
		}
	}

	// Determine session key and work dir
	sessionKey := "__global__"
	var workDir, initialCmd string
	if pid := req.URL.Query().Get("project_id"); pid != "" {
		result := project.Get(pid)
		if result.Success {
			if p, ok := result.Data.(*project.Project); ok {
				sessionKey = p.ID
				workDir = p.Path
				initialCmd = fmt.Sprintf("cd %s && claude --dangerously-skip-permissions", p.Path)
			}
		}
	}

	ws, err := wsUpgrader.Upgrade(w, req, nil)
	if err != nil {
		logger.Error("[terminal] websocket upgrade failed: %v", err)
		return
	}

	// GetOrCreate: reuse existing PTY or create new one
	session, isNew, err := r.termManager.GetOrCreate(sessionKey, cols, rows, workDir, initialCmd)
	if err != nil {
		logger.Error("[terminal] session create failed: %v", err)
		ws.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, err.Error()))
		ws.Close()
		return
	}

	// Attach WS to PTY session (replays ring buffer)
	replayed, err := session.Attach(ws)
	if err != nil {
		logger.Error("[terminal] attach failed: %v", err)
		ws.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, err.Error()))
		ws.Close()
		return
	}

	// Send attached notification (through bridge to avoid concurrent write)
	attachMsg, _ := json.Marshal(map[string]interface{}{
		"type":     "attached",
		"key":      sessionKey,
		"replayed": replayed,
		"is_new":   isNew,
	})
	session.SendMessage(websocket.TextMessage, attachMsg)

	if !isNew {
		logger.Info("[terminal] session %s reattached (replayed=%d bytes)", sessionKey, replayed)
	}
}

// HandleListTerminalSessions handles GET /api/terminal/sessions
func (r *Router) HandleListTerminalSessions(w http.ResponseWriter, req *http.Request) {
	if r.termManager == nil {
		writeError(w, http.StatusServiceUnavailable, "terminal not available")
		return
	}
	sessions := r.termManager.ListSessions()
	if sessions == nil {
		sessions = []terminal.SessionInfo{}
	}
	writeJSON(w, http.StatusOK, types.Result{
		Success: true,
		Data:    sessions,
	})
}

// HandleDeleteTerminalSession handles DELETE /api/terminal/sessions/{key}
func (r *Router) HandleDeleteTerminalSession(w http.ResponseWriter, req *http.Request) {
	if r.termManager == nil {
		writeError(w, http.StatusServiceUnavailable, "terminal not available")
		return
	}
	key := req.PathValue("key")
	if key == "" {
		writeError(w, http.StatusBadRequest, "session key required")
		return
	}
	r.termManager.Remove(key)
	writeJSON(w, http.StatusOK, types.Result{
		Success: true,
		Message: fmt.Sprintf("session %s closed", key),
	})
}

// SetTerminalManager sets the terminal session manager.
func (r *Router) SetTerminalManager(tm *terminal.Manager) {
	r.termManager = tm
}

// RegisterRESTfulRoutes registers all RESTful API endpoints on the given mux.
// The handlers are methods on Router so they share the same project context.
func (r *Router) RegisterRESTfulRoutes(mux *http.ServeMux) {
	// Status & Usage
	mux.HandleFunc("GET /api/status", r.HandleStatus)
	mux.HandleFunc("GET /api/usage", r.HandleGetUsage)
	mux.HandleFunc("POST /api/usage/refresh", r.HandleRefreshUsage)

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

	// Config YAML (raw file)
	mux.HandleFunc("GET /api/config-yaml", r.HandleGetConfigYaml)
	mux.HandleFunc("PUT /api/config-yaml", r.HandleSetConfigYaml)

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

	// Specs
	mux.HandleFunc("GET /api/specs", r.HandleListSpecs)
	mux.HandleFunc("POST /api/specs", r.HandleAddSpec)
	mux.HandleFunc("GET /api/specs/{id}", r.HandleGetSpec)
	mux.HandleFunc("PATCH /api/specs/{id}", r.HandleUpdateSpec)
	mux.HandleFunc("DELETE /api/specs/{id}", r.HandleDeleteSpec)

	// Files
	mux.HandleFunc("GET /api/files", r.HandleListFiles)
	mux.HandleFunc("GET /api/files/content", r.HandleFileContent)

	// Terminal
	mux.HandleFunc("GET /api/terminal/ws", r.HandleTerminalWS)
	mux.HandleFunc("GET /api/terminal/sessions", r.HandleListTerminalSessions)
	mux.HandleFunc("DELETE /api/terminal/sessions/{key}", r.HandleDeleteTerminalSession)
}

