package handler

import (
	"fmt"
	"strings"

	"parkjunwoo.com/claribot/internal/edge"
	"parkjunwoo.com/claribot/internal/project"
	"parkjunwoo.com/claribot/internal/task"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/claude"
)

// Context holds the current state for command execution
type Context struct {
	ProjectID          string
	ProjectPath        string
	ProjectDescription string
}

// Router handles command routing
type Router struct {
	ctx *Context
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		ctx: &Context{},
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
			Message: "project 명령어:\n  [목록:project list]\n  [생성:project create]",
		}
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
		return project.List()
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
			return project.List() // show list with switch buttons
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
		title := strings.Join(args, " ")
		return task.Add(r.ctx.ProjectPath, title)
	case "list":
		return task.List(r.ctx.ProjectPath)
	case "get":
		if len(args) < 1 {
			return task.List(r.ctx.ProjectPath) // show list if no id
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
			Message: "edge 명령어:\n  [목록:edge list]\n  [추가:edge add]",
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
		if len(args) > 0 {
			taskID = args[0]
		}
		return edge.List(r.ctx.ProjectPath, taskID)
	case "delete":
		if len(args) < 2 {
			return types.Result{Success: false, Message: "usage: edge delete <from_id> <to_id>"}
		}
		return edge.Delete(r.ctx.ProjectPath, args[0], args[1])
	default:
		return types.Result{Success: false, Message: fmt.Sprintf("unknown edge command: %s", cmd)}
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
