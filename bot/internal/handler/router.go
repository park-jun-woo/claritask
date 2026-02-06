package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"parkjunwoo.com/claribot/internal/config"
	"parkjunwoo.com/claribot/internal/message"
	"parkjunwoo.com/claribot/internal/project"
	"parkjunwoo.com/claribot/internal/schedule"
	"parkjunwoo.com/claribot/internal/task"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/claude"
	"parkjunwoo.com/claribot/pkg/pagination"
)

// Context holds the current state for command execution
type Context struct {
	ProjectID          string
	ProjectPath        string
	ProjectDescription string
}

// Router handles command routing
type Router struct {
	ctx      *Context
	mu       sync.RWMutex // protects ctx for concurrent access
	pageSize int          // 페이지당 항목 수
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		ctx:      &Context{},
		pageSize: pagination.DefaultPageSize,
	}
}

// SetPageSize sets the default page size for list operations
func (r *Router) SetPageSize(size int) {
	if size > 0 {
		r.pageSize = size
	}
}

// SetProject sets the current project context and persists the selection
func (r *Router) SetProject(id, path, desc string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ctx.ProjectID = id
	r.ctx.ProjectPath = path
	r.ctx.ProjectDescription = desc
	saveLastProject(id)
}

// RestoreProject restores the last selected project from disk
func (r *Router) RestoreProject() {
	id := loadLastProject()
	if id == "" {
		return
	}
	result := project.Get(id)
	if !result.Success {
		return
	}
	if p, ok := result.Data.(*project.Project); ok {
		r.mu.Lock()
		r.ctx.ProjectID = p.ID
		r.ctx.ProjectPath = p.Path
		r.ctx.ProjectDescription = p.Description
		r.mu.Unlock()
	}
}

func lastProjectPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claribot", "last_project")
}

func saveLastProject(id string) {
	os.WriteFile(lastProjectPath(), []byte(id), 0644)
}

func loadLastProject() string {
	data, err := os.ReadFile(lastProjectPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// GetProject returns the current project
func (r *Router) GetProject() (string, string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ctx.ProjectID, r.ctx.ProjectPath
}

// SnapshotContext returns a copy of the current context (thread-safe)
func (r *Router) SnapshotContext() *Context {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return &Context{
		ProjectID:          r.ctx.ProjectID,
		ProjectPath:        r.ctx.ProjectPath,
		ProjectDescription: r.ctx.ProjectDescription,
	}
}

// parseArgs splits input respecting quoted strings
func parseArgs(input string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, ch := range input {
		if !inQuote && (ch == '"' || ch == '\'') {
			inQuote = true
			quoteChar = ch
		} else if inQuote && ch == quoteChar {
			inQuote = false
			quoteChar = 0
		} else if !inQuote && (ch == ' ' || ch == '\t') {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

// Execute parses and executes a command with the given context
func (r *Router) Execute(ctx *Context, input string) types.Result {
	parts := parseArgs(input)
	if len(parts) == 0 {
		return types.Result{Success: false, Message: "empty command"}
	}

	category := parts[0]
	var cmd string
	var args []string

	if len(parts) > 1 {
		cmd = parts[1]
	}
	if len(parts) > 2 {
		args = parts[2:]
	}

	switch category {
	case "project":
		return r.handleProject(ctx, cmd, args)
	case "task":
		return r.handleTask(ctx, cmd, args)
	case "message":
		return r.handleMessage(ctx, cmd, args)
	case "config":
		return r.handleConfig(ctx, cmd, args)
	case "schedule":
		return r.handleSchedule(ctx, cmd, args)
	case "send":
		// "send <content>" → message send <content>
		content := strings.TrimSpace(strings.TrimPrefix(input, "send"))
		return r.handleMessage(ctx, "send", []string{content})
	case "status":
		return r.handleStatus(ctx)
	default:
		return r.handleClaude(ctx, input)
	}
}

// handleClaude sends the input to Claude Code TTY
func (r *Router) handleClaude(ctx *Context, input string) types.Result {
	opts := claude.Options{
		UserPrompt:   input,
		SystemPrompt: "",
		WorkDir:      ctx.ProjectPath,
	}

	result, err := claude.Run(opts)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("Claude 실행 오류: %v", err),
		}
	}

	return types.Result{
		Success: result.ExitCode == 0,
		Message: result.Output,
	}
}

func (r *Router) handleProject(ctx *Context, cmd string, args []string) types.Result {
	switch cmd {
	case "":
		return types.Result{
			Success: true,
			Message: "project 명령어:\n  [목록:project list]\n  [생성:project create]\n  [추가:project add]\n  [설정:project set]",
		}
	case "add":
		var path, projType, desc string
		if len(args) > 0 {
			path = args[0]
		}
		if len(args) > 1 {
			projType = args[1]
		}
		if len(args) > 2 {
			desc = strings.Join(args[2:], " ")
		}
		result := project.Add(path, projType, desc)
		// Auto-switch to added project
		if result.Success && !result.NeedsInput {
			if p, ok := result.Data.(*project.Project); ok {
				r.SetProject(p.ID, p.Path, p.Description)
			}
		}
		return result
	case "create":
		if len(args) < 1 {
			return types.Result{
				Success:    true,
				Message:    "프로젝트 ID를 입력하세요:",
				NeedsInput: true,
				Prompt:     "ID: ",
				Context:    "project create",
			}
		}
		var projType, desc string
		if len(args) > 1 {
			projType = args[1]
		}
		if len(args) > 2 {
			desc = strings.Join(args[2:], " ")
		}
		result := project.Create(args[0], projType, desc)
		// Auto-switch to created project
		if result.Success && !result.NeedsInput {
			if p, ok := result.Data.(*project.Project); ok {
				r.SetProject(p.ID, p.Path, p.Description)
			}
		}
		return result
	case "list":
		page, pageSize := r.parsePagination(args)
		return project.List(pagination.NewPageRequest(page, pageSize))
	case "get":
		id := ctx.ProjectID
		if len(args) > 0 {
			id = args[0]
		}
		if id == "" {
			return types.Result{Success: false, Message: "no project selected"}
		}
		return project.Get(id)
	case "delete":
		if len(args) < 1 {
			return types.Result{Success: false, Message: "usage: project delete <id>"}
		}
		confirmed := len(args) > 1 && args[1] == "yes"
		if len(args) > 1 && args[1] == "no" {
			return types.Result{Success: true, Message: "삭제 취소됨"}
		}
		result := project.Delete(args[0], confirmed)
		// Clear context if deleted project was selected
		if result.Success && confirmed && ctx.ProjectID == args[0] {
			r.SetProject("", "", "")
		}
		return result
	case "set":
		// project set <id> <field> <value>
		if len(args) < 3 {
			return types.Result{Success: false, Message: "usage: project set <id> <field> <value>"}
		}
		value := strings.Join(args[2:], " ")
		return project.Set(args[0], args[1], value)
	case "switch":
		if len(args) < 1 {
			return project.List(pagination.NewPageRequest(1, r.pageSize)) // show list with switch buttons
		}
		// Handle deselect
		if args[0] == "none" {
			r.SetProject("", project.DefaultPath, "글로벌 모드")
			return types.Result{
				Success: true,
				Message: "프로젝트 선택 해제됨 (글로벌 모드)\nPath: " + project.DefaultPath,
			}
		}
		result := project.Switch(args[0])
		if result.Success {
			if p, ok := result.Data.(*project.Project); ok {
				r.SetProject(p.ID, p.Path, p.Description)
			}
		}
		return result
	default:
		return types.Result{Success: false, Message: fmt.Sprintf("unknown project command: %s", cmd)}
	}
}

func (r *Router) handleTask(ctx *Context, cmd string, args []string) types.Result {
	// Show help even without project selected
	if cmd == "" {
		return types.Result{
			Success: true,
			Message: "task 명령어:\n[목록:task list] [추가:task add]\n[플랜순회:task plan --all]\n[실행순회:task run --all]\n[전체순회:task cycle]\n[중단:task stop]",
		}
	}

	// stop doesn't require project selection
	if cmd == "stop" {
		msg, running := task.Stop()
		return types.Result{Success: running, Message: msg}
	}

	if ctx.ProjectPath == "" {
		return types.Result{Success: false, Message: "프로젝트를 먼저 선택하세요: /project switch <id>"}
	}

	switch cmd {
	case "add":
		if len(args) < 1 {
			return types.Result{
				Success:    true,
				Message:    "작업 제목을 입력하세요:",
				NeedsInput: true,
				Prompt:     "Title: ",
				Context:    "task add",
			}
		}
		// Parse --parent, --spec, --spec-file options
		var parentID *int
		var spec, specFile string
		var titleParts []string
		for i := 0; i < len(args); i++ {
			if args[i] == "--parent" && i+1 < len(args) {
				pid, err := strconv.Atoi(args[i+1])
				if err != nil {
					return types.Result{Success: false, Message: "잘못된 parent ID: " + args[i+1]}
				}
				parentID = &pid
				i++ // skip next arg
			} else if args[i] == "--spec-file" && i+1 < len(args) {
				specFile = args[i+1]
				i++ // skip next arg
			} else if args[i] == "--spec" && i+1 < len(args) {
				spec = args[i+1]
				i++ // skip next arg
			} else {
				titleParts = append(titleParts, args[i])
			}
		}
		// --spec-file takes priority over --spec
		if specFile != "" {
			data, err := os.ReadFile(specFile)
			if err != nil {
				return types.Result{Success: false, Message: fmt.Sprintf("spec 파일 읽기 실패: %v", err)}
			}
			spec = string(data)
		}
		title := strings.Join(titleParts, " ")
		if title == "" {
			return types.Result{
				Success:    true,
				Message:    "작업 제목을 입력하세요:",
				NeedsInput: true,
				Prompt:     "Title: ",
				Context:    "task add",
			}
		}
		return task.Add(ctx.ProjectPath, title, parentID, spec)
	case "list":
		// task list [parent_id] [-p page] [-n pageSize] [--tree]
		// Check for --tree flag
		for _, arg := range args {
			if arg == "--tree" {
				return task.ListTree(ctx.ProjectPath)
			}
		}
		var parentID *int
		page, pageSize := r.parsePagination(args)
		// Check first positional arg for parent_id (skip -p/-n and their values)
		for i := 0; i < len(args); i++ {
			arg := args[i]
			if arg == "-p" || arg == "-n" {
				i++ // skip next value
				continue
			}
			if strings.HasPrefix(arg, "-") {
				continue
			}
			pid, err := strconv.Atoi(arg)
			if err == nil {
				parentID = &pid
				break
			}
		}
		return task.List(ctx.ProjectPath, parentID, pagination.NewPageRequest(page, pageSize))
	case "get":
		if len(args) < 1 {
			return task.List(ctx.ProjectPath, nil, pagination.NewPageRequest(1, r.pageSize)) // show list if no id
		}
		return task.Get(ctx.ProjectPath, args[0])
	case "set":
		if len(args) < 3 {
			return types.Result{Success: false, Message: "usage: task set <id> <field> <value>"}
		}
		value := strings.Join(args[2:], " ")
		return task.Set(ctx.ProjectPath, args[0], args[1], value)
	case "delete":
		if len(args) < 1 {
			return types.Result{Success: false, Message: "usage: task delete <id>"}
		}
		confirmed := len(args) > 1 && args[1] == "yes"
		if len(args) > 1 && args[1] == "no" {
			return types.Result{Success: true, Message: "삭제 취소됨"}
		}
		return task.Delete(ctx.ProjectPath, args[0], confirmed)
	case "plan":
		// task plan [id] [--all]
		if len(args) > 0 && args[0] == "--all" {
			return task.PlanAll(ctx.ProjectPath)
		}
		var id string
		if len(args) > 0 {
			id = args[0]
		}
		return task.Plan(ctx.ProjectPath, id)
	case "run":
		// task run [id] [--all]
		if len(args) > 0 && args[0] == "--all" {
			return task.RunAll(ctx.ProjectPath)
		}
		var id string
		if len(args) > 0 {
			id = args[0]
		}
		return task.Run(ctx.ProjectPath, id)
	case "cycle":
		// task cycle - 1회차 + 2회차 자동 실행
		return task.Cycle(ctx.ProjectPath)
	default:
		return types.Result{Success: false, Message: fmt.Sprintf("unknown task command: %s", cmd)}
	}
}

func (r *Router) handleMessage(ctx *Context, cmd string, args []string) types.Result {
	// Show help even without project selected
	if cmd == "" {
		return types.Result{
			Success: true,
			Message: "message 명령어:\n  [목록:message list]\n  [전송:message send]\n  [상태:message status]",
		}
	}

	// Use default path if no project selected (global mode)
	projectPath := ctx.ProjectPath
	if projectPath == "" {
		projectPath = project.DefaultPath
	}
	if projectPath == "" {
		return types.Result{Success: false, Message: "프로젝트 경로가 설정되지 않았습니다"}
	}

	switch cmd {
	case "send":
		if len(args) < 1 || args[0] == "" {
			return types.Result{
				Success:    true,
				Message:    "메시지를 입력하세요:",
				NeedsInput: true,
				Prompt:     "Message: ",
				Context:    "send",
			}
		}
		// Check if first arg is source (telegram/cli)
		source := "cli"
		content := strings.Join(args, " ")
		if len(args) > 1 && (args[0] == "telegram" || args[0] == "cli") {
			source = args[0]
			content = strings.Join(args[1:], " ")
		}
		return message.Send(projectPath, content, source)
	case "list":
		page, pageSize := r.parsePagination(args)
		return message.List(projectPath, pagination.NewPageRequest(page, pageSize))
	case "get":
		if len(args) < 1 {
			return message.List(projectPath, pagination.NewPageRequest(1, r.pageSize))
		}
		return message.Get(projectPath, args[0])
	case "status":
		return message.Status(projectPath)
	case "processing":
		return message.Processing(projectPath)
	default:
		return types.Result{Success: false, Message: fmt.Sprintf("unknown message command: %s", cmd)}
	}
}

func (r *Router) handleStatus(ctx *Context) types.Result {
	var sb strings.Builder

	// Claude status
	claudeStatus := claude.GetStatus()
	sb.WriteString(fmt.Sprintf("🤖 Claude: %d/%d 사용중", claudeStatus.Used, claudeStatus.Max))
	if claudeStatus.Available == 0 {
		sb.WriteString(" (대기열 가득)")
	}
	sb.WriteString("\n")

	// Project status
	if ctx.ProjectID == "" {
		sb.WriteString("\n📁 프로젝트: 선택 안됨 (글로벌 모드)\n")
		sb.WriteString("[선택:project switch]")
	} else {
		sb.WriteString(fmt.Sprintf("\n📁 프로젝트: %s\n", ctx.ProjectID))
		sb.WriteString(fmt.Sprintf("   설명: %s\n", ctx.ProjectDescription))

		// Cycle status
		cycleStatus := task.GetCycleStatus()
		if cycleStatus.Status != "idle" {
			sb.WriteString("\n🔄 순회 상태:\n")
			typeLabel := map[string]string{"cycle": "전체순회", "plan": "플랜순회", "run": "실행순회"}[cycleStatus.Type]
			if typeLabel == "" {
				typeLabel = cycleStatus.Type
			}
			switch cycleStatus.Status {
			case "running":
				sb.WriteString(fmt.Sprintf("   ▶️ %s 진행 중", typeLabel))
				if cycleStatus.CurrentTaskID > 0 {
					sb.WriteString(fmt.Sprintf(" (Task #%d)", cycleStatus.CurrentTaskID))
				}
				sb.WriteString("\n")
			case "interrupted":
				sb.WriteString(fmt.Sprintf("   ⚠️ %s 중단됨", typeLabel))
				if cycleStatus.CurrentTaskID > 0 {
					sb.WriteString(fmt.Sprintf(" (Task #%d에서 중단)", cycleStatus.CurrentTaskID))
				}
				sb.WriteString("\n")
				sb.WriteString(fmt.Sprintf("[순회 재개:resume:%s]", cycleStatus.Type))
				sb.WriteString("\n")
			}
			elapsed := time.Since(cycleStatus.StartedAt).Truncate(time.Second)
			sb.WriteString(fmt.Sprintf("   경과: %s\n", elapsed))
		}

		// Task stats
		if stats, err := task.GetStats(ctx.ProjectPath); err == nil && stats.Total > 0 {
			sb.WriteString("\n📊 Task 현황:\n")

			// 통계
			remaining := stats.Todo + stats.Planned
			sb.WriteString(fmt.Sprintf("   전체: %d개 (실행대상: %d개)\n", stats.Total, stats.Leaf))
			sb.WriteString(fmt.Sprintf("   ✅ 완료: %d개", stats.Done))
			if stats.Failed > 0 {
				sb.WriteString(fmt.Sprintf(" / ❌ 실패: %d개", stats.Failed))
			}
			sb.WriteString("\n")
			if remaining > 0 {
				sb.WriteString(fmt.Sprintf("   ⏳ 대기: %d개 (todo:%d, planned:%d)\n", remaining, stats.Todo, stats.Planned))
			}

			// 진행률
			if stats.Leaf > 0 {
				progress := float64(stats.Done) / float64(stats.Leaf) * 100
				sb.WriteString(fmt.Sprintf("   진행률: %.0f%%", progress))
			}
		}
	}

	return types.Result{
		Success: true,
		Message: sb.String(),
		Data:    claudeStatus,
	}
}

// parsePagination extracts -p (page), -n (pageSize), --all from args
func (r *Router) parsePagination(args []string) (page, pageSize int) {
	page = 1
	pageSize = r.pageSize

	for i := 0; i < len(args); i++ {
		if args[i] == "-p" && i+1 < len(args) {
			if p, err := strconv.Atoi(args[i+1]); err == nil && p > 0 {
				page = p
			}
			i++
		} else if args[i] == "-n" && i+1 < len(args) {
			if n, err := strconv.Atoi(args[i+1]); err == nil && n > 0 {
				pageSize = n
			}
			i++
		} else if args[i] == "--all" {
			pageSize = pagination.MaxPageSize
		}
	}
	return
}

func (r *Router) handleConfig(ctx *Context, cmd string, args []string) types.Result {
	if cmd == "" {
		return types.Result{
			Success: true,
			Message: "config 명령어:\n  [목록:config list]\n  [조회:config get]\n  [설정:config set]\n  [삭제:config delete]",
		}
	}

	switch cmd {
	case "set":
		if len(args) < 2 {
			return types.Result{Success: false, Message: "usage: config set <key> <value>"}
		}
		value := strings.Join(args[1:], " ")
		return config.SetDBConfig(args[0], value)
	case "get":
		if len(args) < 1 {
			return types.Result{Success: false, Message: "usage: config get <key>"}
		}
		return config.GetDBConfig(args[0])
	case "list":
		page, pageSize := r.parsePagination(args)
		return config.ListDBConfig(pagination.NewPageRequest(page, pageSize))
	case "delete":
		if len(args) < 1 {
			return types.Result{Success: false, Message: "usage: config delete <key>"}
		}
		confirmed := len(args) > 1 && args[1] == "yes"
		if len(args) > 1 && args[1] == "no" {
			return types.Result{Success: true, Message: "삭제 취소됨"}
		}
		return config.DeleteDBConfig(args[0], confirmed)
	default:
		return types.Result{Success: false, Message: fmt.Sprintf("unknown config command: %s", cmd)}
	}
}

func (r *Router) handleSchedule(ctx *Context, cmd string, args []string) types.Result {
	if cmd == "" {
		return types.Result{
			Success: true,
			Message: "schedule 명령어:\n  [목록:schedule list]\n  [추가:schedule add]\n  [조회:schedule get]\n  [수정:schedule set]\n  [실행기록:schedule runs]",
		}
	}

	switch cmd {
	case "add":
		// schedule add "cron" "message" [--project id] [--once]
		if len(args) < 2 {
			return types.Result{
				Success: false,
				Message: "usage: schedule add <cron_expr> <message> [--project <id>] [--once]",
			}
		}

		cronExpr := args[0]
		var messageParts []string
		var projectID *string
		runOnce := false

		for i := 1; i < len(args); i++ {
			if args[i] == "--project" && i+1 < len(args) {
				projectID = &args[i+1]
				i++
			} else if args[i] == "--once" {
				runOnce = true
			} else {
				messageParts = append(messageParts, args[i])
			}
		}

		message := strings.Join(messageParts, " ")
		if message == "" {
			return types.Result{
				Success: false,
				Message: "메시지를 입력하세요",
			}
		}

		// Use current project if not specified
		if projectID == nil && ctx.ProjectID != "" {
			projectID = &ctx.ProjectID
		}

		return schedule.Add(cronExpr, message, projectID, runOnce)

	case "list":
		// schedule list [--all] [-p page]
		showAll := false
		for _, arg := range args {
			if arg == "--all" {
				showAll = true
				break
			}
		}
		page, pageSize := r.parsePagination(args)

		var projectID *string
		if !showAll && ctx.ProjectID != "" {
			projectID = &ctx.ProjectID
		}

		return schedule.List(projectID, showAll, pagination.NewPageRequest(page, pageSize))

	case "get":
		if len(args) < 1 {
			return schedule.List(nil, true, pagination.NewPageRequest(1, r.pageSize))
		}
		return schedule.Get(args[0])

	case "delete":
		if len(args) < 1 {
			return types.Result{Success: false, Message: "usage: schedule delete <id>"}
		}
		confirmed := len(args) > 1 && args[1] == "yes"
		if len(args) > 1 && args[1] == "no" {
			return types.Result{Success: true, Message: "삭제 취소됨"}
		}
		return schedule.Delete(args[0], confirmed)

	case "enable":
		if len(args) < 1 {
			return types.Result{Success: false, Message: "usage: schedule enable <id>"}
		}
		return schedule.Enable(args[0])

	case "disable":
		if len(args) < 1 {
			return types.Result{Success: false, Message: "usage: schedule disable <id>"}
		}
		return schedule.Disable(args[0])

	case "runs":
		// schedule runs <schedule_id> [-p page]
		if len(args) < 1 {
			return types.Result{Success: false, Message: "usage: schedule runs <schedule_id>"}
		}
		page, pageSize := r.parsePagination(args)
		return schedule.Runs(args[0], pagination.NewPageRequest(page, pageSize))

	case "run":
		// schedule run <run_id>
		if len(args) < 1 {
			return types.Result{Success: false, Message: "usage: schedule run <run_id>"}
		}
		return schedule.Run(args[0])

	case "set":
		// schedule set <id> project <project_id|none>
		if len(args) < 3 {
			return types.Result{Success: false, Message: "usage: schedule set <id> project <project_id|none>"}
		}
		if args[1] != "project" {
			return types.Result{Success: false, Message: "usage: schedule set <id> project <project_id|none>"}
		}
		var projectID *string
		if args[2] != "none" {
			projectID = &args[2]
		}
		return schedule.SetProject(args[0], projectID)

	default:
		return types.Result{Success: false, Message: fmt.Sprintf("unknown schedule command: %s", cmd)}
	}
}
