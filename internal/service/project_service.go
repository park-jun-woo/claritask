package service

import (
	"encoding/json"
	"fmt"

	"parkjunwoo.com/talos/internal/db"
	"parkjunwoo.com/talos/internal/model"
)

// CreateProject creates a new project
func CreateProject(database *db.DB, id, name, description string) error {
	now := db.TimeNow()
	_, err := database.Exec(
		`INSERT INTO projects (id, name, description, status, created_at) VALUES (?, ?, ?, 'active', ?)`,
		id, name, description, now,
	)
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}
	return nil
}

// GetProject retrieves the project (single project per DB)
func GetProject(database *db.DB) (*model.Project, error) {
	row := database.QueryRow(`SELECT id, name, description, status, created_at FROM projects LIMIT 1`)
	var p model.Project
	var createdAt string
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Status, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	p.CreatedAt, _ = db.ParseTime(createdAt)
	return &p, nil
}

// UpdateProject updates an existing project
func UpdateProject(database *db.DB, p *model.Project) error {
	_, err := database.Exec(
		`UPDATE projects SET name = ?, description = ?, status = ? WHERE id = ?`,
		p.Name, p.Description, p.Status, p.ID,
	)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	return nil
}

// SetContext sets the project context (upsert)
func SetContext(database *db.DB, data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal context: %w", err)
	}
	now := db.TimeNow()
	_, err = database.Exec(
		`INSERT INTO context (id, data, created_at, updated_at) VALUES (1, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET data = ?, updated_at = ?`,
		string(jsonData), now, now, string(jsonData), now,
	)
	if err != nil {
		return fmt.Errorf("set context: %w", err)
	}
	return nil
}

// GetContext retrieves the project context
func GetContext(database *db.DB) (map[string]interface{}, error) {
	row := database.QueryRow(`SELECT data FROM context WHERE id = 1`)
	var data string
	err := row.Scan(&data)
	if err != nil {
		return nil, fmt.Errorf("get context: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("unmarshal context: %w", err)
	}
	return result, nil
}

// SetTech sets the tech stack (upsert)
func SetTech(database *db.DB, data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal tech: %w", err)
	}
	now := db.TimeNow()
	_, err = database.Exec(
		`INSERT INTO tech (id, data, created_at, updated_at) VALUES (1, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET data = ?, updated_at = ?`,
		string(jsonData), now, now, string(jsonData), now,
	)
	if err != nil {
		return fmt.Errorf("set tech: %w", err)
	}
	return nil
}

// GetTech retrieves the tech stack
func GetTech(database *db.DB) (map[string]interface{}, error) {
	row := database.QueryRow(`SELECT data FROM tech WHERE id = 1`)
	var data string
	err := row.Scan(&data)
	if err != nil {
		return nil, fmt.Errorf("get tech: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("unmarshal tech: %w", err)
	}
	return result, nil
}

// SetDesign sets the design decisions (upsert)
func SetDesign(database *db.DB, data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal design: %w", err)
	}
	now := db.TimeNow()
	_, err = database.Exec(
		`INSERT INTO design (id, data, created_at, updated_at) VALUES (1, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET data = ?, updated_at = ?`,
		string(jsonData), now, now, string(jsonData), now,
	)
	if err != nil {
		return fmt.Errorf("set design: %w", err)
	}
	return nil
}

// GetDesign retrieves the design decisions
func GetDesign(database *db.DB) (map[string]interface{}, error) {
	row := database.QueryRow(`SELECT data FROM design WHERE id = 1`)
	var data string
	err := row.Scan(&data)
	if err != nil {
		return nil, fmt.Errorf("get design: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("unmarshal design: %w", err)
	}
	return result, nil
}

// MissingField represents a missing required field
type MissingField struct {
	Field   string   `json:"field"`
	Prompt  string   `json:"prompt"`
	Options []string `json:"options,omitempty"`
}

// RequiredResult represents the result of required check
type RequiredResult struct {
	Ready           bool           `json:"ready"`
	MissingRequired []MissingField `json:"missing_required,omitempty"`
}

// CheckRequired checks if all required fields are set
func CheckRequired(database *db.DB) (*RequiredResult, error) {
	result := &RequiredResult{Ready: true}
	var missing []MissingField

	// Check context
	ctx, err := GetContext(database)
	if err != nil {
		// Context not set
		missing = append(missing, MissingField{
			Field:  "context.project_name",
			Prompt: "프로젝트 이름을 입력하세요",
		})
		missing = append(missing, MissingField{
			Field:  "context.description",
			Prompt: "프로젝트 설명을 입력하세요",
		})
	} else {
		if _, ok := ctx["project_name"]; !ok {
			missing = append(missing, MissingField{
				Field:  "context.project_name",
				Prompt: "프로젝트 이름을 입력하세요",
			})
		}
		if _, ok := ctx["description"]; !ok {
			missing = append(missing, MissingField{
				Field:  "context.description",
				Prompt: "프로젝트 설명을 입력하세요",
			})
		}
	}

	// Check tech
	tech, err := GetTech(database)
	if err != nil {
		missing = append(missing, MissingField{
			Field:   "tech.backend",
			Prompt:  "백엔드 기술을 선택하세요",
			Options: []string{"go", "node", "python", "java"},
		})
		missing = append(missing, MissingField{
			Field:   "tech.frontend",
			Prompt:  "프론트엔드 기술을 선택하세요",
			Options: []string{"react", "vue", "angular", "none"},
		})
		missing = append(missing, MissingField{
			Field:   "tech.database",
			Prompt:  "데이터베이스를 선택하세요",
			Options: []string{"postgresql", "mysql", "sqlite", "mongodb"},
		})
	} else {
		if _, ok := tech["backend"]; !ok {
			missing = append(missing, MissingField{
				Field:   "tech.backend",
				Prompt:  "백엔드 기술을 선택하세요",
				Options: []string{"go", "node", "python", "java"},
			})
		}
		if _, ok := tech["frontend"]; !ok {
			missing = append(missing, MissingField{
				Field:   "tech.frontend",
				Prompt:  "프론트엔드 기술을 선택하세요",
				Options: []string{"react", "vue", "angular", "none"},
			})
		}
		if _, ok := tech["database"]; !ok {
			missing = append(missing, MissingField{
				Field:   "tech.database",
				Prompt:  "데이터베이스를 선택하세요",
				Options: []string{"postgresql", "mysql", "sqlite", "mongodb"},
			})
		}
	}

	// Check design
	design, err := GetDesign(database)
	if err != nil {
		missing = append(missing, MissingField{
			Field:   "design.architecture",
			Prompt:  "아키텍처 패턴을 선택하세요",
			Options: []string{"monolith", "microservice", "serverless"},
		})
		missing = append(missing, MissingField{
			Field:   "design.auth_method",
			Prompt:  "인증 방식을 선택하세요",
			Options: []string{"jwt", "session", "oauth", "none"},
		})
		missing = append(missing, MissingField{
			Field:   "design.api_style",
			Prompt:  "API 스타일을 선택하세요",
			Options: []string{"rest", "graphql", "grpc"},
		})
	} else {
		if _, ok := design["architecture"]; !ok {
			missing = append(missing, MissingField{
				Field:   "design.architecture",
				Prompt:  "아키텍처 패턴을 선택하세요",
				Options: []string{"monolith", "microservice", "serverless"},
			})
		}
		if _, ok := design["auth_method"]; !ok {
			missing = append(missing, MissingField{
				Field:   "design.auth_method",
				Prompt:  "인증 방식을 선택하세요",
				Options: []string{"jwt", "session", "oauth", "none"},
			})
		}
		if _, ok := design["api_style"]; !ok {
			missing = append(missing, MissingField{
				Field:   "design.api_style",
				Prompt:  "API 스타일을 선택하세요",
				Options: []string{"rest", "graphql", "grpc"},
			})
		}
	}

	if len(missing) > 0 {
		result.Ready = false
		result.MissingRequired = missing
	}

	return result, nil
}

// ProjectSetInput represents input for SetProjectFull
type ProjectSetInput struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Context     map[string]interface{} `json:"context"`
	Tech        map[string]interface{} `json:"tech"`
	Design      map[string]interface{} `json:"design"`
}

// SetProjectFull sets all project data at once
func SetProjectFull(database *db.DB, input ProjectSetInput) error {
	// Update project
	project, err := GetProject(database)
	if err != nil {
		return fmt.Errorf("get project: %w", err)
	}

	if input.Name != "" {
		project.Name = input.Name
	}
	if input.Description != "" {
		project.Description = input.Description
	}

	if err := UpdateProject(database, project); err != nil {
		return fmt.Errorf("update project: %w", err)
	}

	// Set context if provided
	if input.Context != nil {
		if err := SetContext(database, input.Context); err != nil {
			return fmt.Errorf("set context: %w", err)
		}
	}

	// Set tech if provided
	if input.Tech != nil {
		if err := SetTech(database, input.Tech); err != nil {
			return fmt.Errorf("set tech: %w", err)
		}
	}

	// Set design if provided
	if input.Design != nil {
		if err := SetDesign(database, input.Design); err != nil {
			return fmt.Errorf("set design: %w", err)
		}
	}

	return nil
}
