package handler

import (
	"fmt"
	"strconv"
	"strings"

	"parkjunwoo.com/claribot/internal/edge"
	"parkjunwoo.com/claribot/internal/message"
	"parkjunwoo.com/claribot/internal/project"
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
	pageSize int // 페이지당 항목 수
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

// SetProject sets the current project context
func (r *Router) SetProject(id, path, desc string) {
	r.ctx.ProjectID = id
	r.ctx.ProjectPath = path
	r.ctx.ProjectDescription = desc
}

// GetProject returns the current project
func (r *Router) GetProject() (string, string) {
	return r.ctx.ProjectID, r.ctx.ProjectPath
}

// Execute parses and executes a command
func (r *Router) Execute(input string) types.Result {
	parts := strings.Fields(input)
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
		return r.handleProject(cmd, args)
	case "task":
		return r.handleTask(cmd, args)
	case "edge":
		return r.handleEdge(cmd, args)
	case "message":
		return r.handleMessage(cmd, args)
	case "send":
		// "send <content>" → message send <content>
		content := strings.TrimSpace(strings.TrimPrefix(input, "send"))
		return r.handleMessage("send", []string{content})
	case "status":
		return r.handleStatus()
	default:
		return r.handleClaude(input)
	}
}

// handleClaude sends the input to Claude Code TTY
func (r *Router) handleClaude(input string) types.Result {
	opts := claude.Options{
		UserPrompt:   input,
		SystemPrompt: "",
		WorkDir:      r.ctx.ProjectPath,
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

func (r *Router) handleProject(cmd string, args []string) types.Result {
	switch cmd {
	case "":
		return types.Result{
			Success: true,
			Message: "project 명령어:\n  [목록:project list]\n  [생성:project create]\n  [추가:project add]",
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
		id := r.ctx.ProjectID
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
		if result.Success && confirmed && r.ctx.ProjectID == args[0] {
			r.SetProject("", "", "")
		}
		return result
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

func (r *Router) handleTask(cmd string, args []string) types.Result {
	// Show help even without project selected
	if cmd == "" {
		return types.Result{
			Success: true,
			Message: "task 명령어:\n  [목록:task list]\n  [추가:task add]\n  [실행:task run]",
		}
	}

	if r.ctx.ProjectPath == "" {
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
		// Parse --parent option
		var parentID *int
		var titleParts []string
		for i := 0; i < len(args); i++ {
			if args[i] == "--parent" && i+1 < len(args) {
				pid, err := strconv.Atoi(args[i+1])
				if err != nil {
					return types.Result{Success: false, Message: "잘못된 parent ID: " + args[i+1]}
				}
				parentID = &pid
				i++ // skip next arg
			} else {
				titleParts = append(titleParts, args[i])
			}
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
		return task.Add(r.ctx.ProjectPath, title, parentID)
	case "list":
		// task list [parent_id] [-p page] [-n pageSize]
		var parentID *int
		page, pageSize := r.parsePagination(args)
		// Check first positional arg for parent_id
		for _, arg := range args {
			if arg != "-p" && arg != "-n" && !strings.HasPrefix(arg, "-") {
				pid, err := strconv.Atoi(arg)
				if err == nil {
					parentID = &pid
					break
				}
			}
		}
		return task.List(r.ctx.ProjectPath, parentID, pagination.NewPageRequest(page, pageSize))
	case "get":
		if len(args) < 1 {
			return task.List(r.ctx.ProjectPath, nil, pagination.NewPageRequest(1, r.pageSize)) // show list if no id
		}
		return task.Get(r.ctx.ProjectPath, args[0])
	case "set":
		if len(args) < 3 {
			return types.Result{Success: false, Message: "usage: task set <id> <field> <value>"}
		}
		value := strings.Join(args[2:], " ")
		return task.Set(r.ctx.ProjectPath, args[0], args[1], value)
	case "delete":
		if len(args) < 1 {
			return types.Result{Success: false, Message: "usage: task delete <id>"}
		}
		confirmed := len(args) > 1 && args[1] == "yes"
		if len(args) > 1 && args[1] == "no" {
			return types.Result{Success: true, Message: "삭제 취소됨"}
		}
		return task.Delete(r.ctx.ProjectPath, args[0], confirmed)
	case "run":
		var id string
		if len(args) > 0 {
			id = args[0]
		}
		return task.Run(r.ctx.ProjectPath, id)
	default:
		return types.Result{Success: false, Message: fmt.Sprintf("unknown task command: %s", cmd)}
	}
}

func (r *Router) handleEdge(cmd string, args []string) types.Result {
	// Show help even without project selected
	if cmd == "" {
		return types.Result{
			Success: true,
			Message: "edge 명령어:\n  [목록:edge list]\n  [추가:edge add]\n  [조회:edge get]",
		}
	}

	if r.ctx.ProjectPath == "" {
		return types.Result{Success: false, Message: "프로젝트를 먼저 선택하세요: /project switch <id>"}
	}

	switch cmd {
	case "add":
		if len(args) < 2 {
			return types.Result{Success: false, Message: "usage: edge add <from_id> <to_id>"}
		}
		return edge.Add(r.ctx.ProjectPath, args[0], args[1])
	case "list":
		var taskID string
		page, pageSize := r.parsePagination(args)
		for _, arg := range args {
			if arg != "-p" && arg != "-n" && !strings.HasPrefix(arg, "-") {
				taskID = arg
				break
			}
		}
		return edge.List(r.ctx.ProjectPath, taskID, pagination.NewPageRequest(page, pageSize))
	case "get":
		if len(args) < 2 {
			return types.Result{Success: false, Message: "usage: edge get <from_id> <to_id>"}
		}
		return edge.Get(r.ctx.ProjectPath, args[0], args[1])
	case "delete":
		if len(args) < 2 {
			return types.Result{Success: false, Message: "usage: edge delete <from_id> <to_id>"}
		}
		confirmed := len(args) > 2 && args[2] == "yes"
		if len(args) > 2 && args[2] == "no" {
			return types.Result{Success: true, Message: "삭제 취소됨"}
		}
		return edge.Delete(r.ctx.ProjectPath, args[0], args[1], confirmed)
	default:
		return types.Result{Success: false, Message: fmt.Sprintf("unknown edge command: %s", cmd)}
	}
}

func (r *Router) handleMessage(cmd string, args []string) types.Result {
	// Show help even without project selected
	if cmd == "" {
		return types.Result{
			Success: true,
			Message: "message 명령어:\n  [목록:message list]\n  [전송:message send]",
		}
	}

	if r.ctx.ProjectPath == "" {
		return types.Result{Success: false, Message: "프로젝트를 먼저 선택하세요: /project switch <id>"}
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
		return message.Send(r.ctx.ProjectPath, content, source)
	case "list":
		page, pageSize := r.parsePagination(args)
		return message.List(r.ctx.ProjectPath, pagination.NewPageRequest(page, pageSize))
	case "get":
		if len(args) < 1 {
			return message.List(r.ctx.ProjectPath, pagination.NewPageRequest(1, r.pageSize))
		}
		return message.Get(r.ctx.ProjectPath, args[0])
	default:
		return types.Result{Success: false, Message: fmt.Sprintf("unknown message command: %s", cmd)}
	}
}

func (r *Router) handleStatus() types.Result {
	if r.ctx.ProjectID == "" {
		return types.Result{
			Success: true,
			Message: "선택된 프로젝트 없음\n[선택:project switch]",
		}
	}
	return types.Result{
		Success: true,
		Message: fmt.Sprintf("프로젝트: %s\n설명: %s", r.ctx.ProjectID, r.ctx.ProjectDescription),
	}
}

// parsePagination extracts -p (page) and -n (pageSize) from args
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
		}
	}
	return
}
