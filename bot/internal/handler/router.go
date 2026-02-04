package handler

import (
	"fmt"
	"strings"

	"parkjunwoo.com/claribot/internal/edge"
	"parkjunwoo.com/claribot/internal/project"
	"parkjunwoo.com/claribot/internal/task"
	"parkjunwoo.com/claribot/internal/types"
)

// Context holds the current state for command execution
type Context struct {
	ProjectID   string
	ProjectPath string
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
func (r *Router) SetProject(id, path string) {
	r.ctx.ProjectID = id
	r.ctx.ProjectPath = path
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
		return types.Result{Success: false, Message: fmt.Sprintf("unknown category: %s", category)}
	}
}

func (r *Router) handleProject(cmd string, args []string) types.Result {
	switch cmd {
	case "create":
		if len(args) < 1 {
			return types.Result{Success: false, Message: "usage: project create <id>"}
		}
		return project.Create(args[0])
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
		return project.Delete(args[0])
	case "switch":
		if len(args) < 1 {
			return types.Result{Success: false, Message: "usage: project switch <id>"}
		}
		result := project.Switch(args[0])
		if result.Success {
			if p, ok := result.Data.(*project.Project); ok {
				r.SetProject(p.ID, p.Path)
			}
		}
		return result
	default:
		return types.Result{Success: false, Message: fmt.Sprintf("unknown project command: %s", cmd)}
	}
}

func (r *Router) handleTask(cmd string, args []string) types.Result {
	if r.ctx.ProjectPath == "" {
		return types.Result{Success: false, Message: "no project selected"}
	}

	switch cmd {
	case "add":
		if len(args) < 1 {
			return types.Result{Success: false, Message: "usage: task add <title>"}
		}
		title := strings.Join(args, " ")
		return task.Add(r.ctx.ProjectPath, title)
	case "list":
		return task.List(r.ctx.ProjectPath)
	case "get":
		if len(args) < 1 {
			return types.Result{Success: false, Message: "usage: task get <id>"}
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
		return task.Delete(r.ctx.ProjectPath, args[0])
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
	if r.ctx.ProjectPath == "" {
		return types.Result{Success: false, Message: "no project selected"}
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
	return types.Result{
		Success: true,
		Message: fmt.Sprintf("Project: %s\nPath: %s", r.ctx.ProjectID, r.ctx.ProjectPath),
	}
}
