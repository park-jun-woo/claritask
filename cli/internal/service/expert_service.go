package service

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/model"
)

const (
	ExpertsDir     = ".claritask/experts"
	ExpertFileName = "EXPERT.md"
)

var expertIDRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`)

// AddExpert creates a new expert with template
func AddExpert(database *db.DB, expertID string) (*model.Expert, error) {
	// Validate expert ID
	if !expertIDRegex.MatchString(expertID) {
		return nil, fmt.Errorf("invalid expert ID: must be lowercase letters, numbers, and hyphens only")
	}

	// Check if expert already exists
	expertDir := filepath.Join(ExpertsDir, expertID)
	expertPath := filepath.Join(expertDir, ExpertFileName)

	if _, err := os.Stat(expertPath); err == nil {
		return nil, fmt.Errorf("expert '%s' already exists", expertID)
	}

	// Create directory
	if err := os.MkdirAll(expertDir, 0755); err != nil {
		return nil, fmt.Errorf("create expert directory: %w", err)
	}

	// Create EXPERT.md template
	template := getExpertTemplate(expertID)
	if err := os.WriteFile(expertPath, []byte(template), 0644); err != nil {
		return nil, fmt.Errorf("create expert file: %w", err)
	}

	// Calculate content hash
	hash := sha256.Sum256([]byte(template))
	contentHash := hex.EncodeToString(hash[:])

	// Save to database with content
	now := db.TimeNow()
	_, err := database.Exec(
		`INSERT INTO experts (id, name, version, domain, language, framework, path, description, content, content_hash, content_backup, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		expertID, expertID, "1.0.0", "", "", "", expertPath, "", template, contentHash, "", "active", now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("save expert to database: %w", err)
	}

	return &model.Expert{
		ID:      expertID,
		Name:    expertID,
		Version: "1.0.0",
		Path:    expertPath,
	}, nil
}

// ListExperts returns all experts with optional filter
func ListExperts(database *db.DB, filter string) ([]model.Expert, error) {
	// Get current project for assigned check
	project, _ := GetProject(database)
	projectID := ""
	if project != nil {
		projectID = project.ID
	}

	// Get assigned expert IDs
	assignedMap := make(map[string]bool)
	if projectID != "" {
		rows, err := database.Query(
			`SELECT expert_id FROM project_experts WHERE project_id = ?`,
			projectID,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id string
				rows.Scan(&id)
				assignedMap[id] = true
			}
		}
	}

	// Scan filesystem for experts
	var experts []model.Expert

	entries, err := os.ReadDir(ExpertsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return experts, nil
		}
		return nil, fmt.Errorf("read experts directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		expertID := entry.Name()
		expertPath := filepath.Join(ExpertsDir, expertID, ExpertFileName)

		if _, err := os.Stat(expertPath); os.IsNotExist(err) {
			continue
		}

		expert, err := parseExpertMetadata(expertPath)
		if err != nil {
			// Use basic info if parsing fails
			expert = &model.Expert{
				ID:   expertID,
				Name: expertID,
				Path: expertPath,
			}
		}
		expert.ID = expertID
		expert.Path = expertPath
		expert.Assigned = assignedMap[expertID]

		// Apply filter
		switch filter {
		case "assigned":
			if !expert.Assigned {
				continue
			}
		case "available":
			if expert.Assigned {
				continue
			}
		}

		experts = append(experts, *expert)
	}

	return experts, nil
}

// GetExpert returns a single expert by ID
func GetExpert(database *db.DB, expertID string) (*model.Expert, error) {
	expertPath := filepath.Join(ExpertsDir, expertID, ExpertFileName)

	if _, err := os.Stat(expertPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("expert '%s' not found", expertID)
	}

	expert, err := parseExpertMetadata(expertPath)
	if err != nil {
		expert = &model.Expert{
			ID:   expertID,
			Name: expertID,
		}
	}
	expert.ID = expertID
	expert.Path = expertPath

	// Check if assigned
	project, _ := GetProject(database)
	if project != nil {
		var count int
		database.QueryRow(
			`SELECT COUNT(*) FROM project_experts WHERE project_id = ? AND expert_id = ?`,
			project.ID, expertID,
		).Scan(&count)
		expert.Assigned = count > 0
	}

	return expert, nil
}

// RemoveExpert removes an expert
func RemoveExpert(database *db.DB, expertID string, force bool) error {
	expertDir := filepath.Join(ExpertsDir, expertID)

	if _, err := os.Stat(expertDir); os.IsNotExist(err) {
		return fmt.Errorf("expert '%s' not found", expertID)
	}

	// Check if assigned
	var count int
	database.QueryRow(
		`SELECT COUNT(*) FROM project_experts WHERE expert_id = ?`,
		expertID,
	).Scan(&count)

	if count > 0 && !force {
		return fmt.Errorf("expert '%s' is assigned to project. Use --force to remove", expertID)
	}

	// Remove from project_experts
	database.Exec(`DELETE FROM project_experts WHERE expert_id = ?`, expertID)

	// Remove from expert_assignments (feature-level)
	database.Exec(`DELETE FROM expert_assignments WHERE expert_id = ?`, expertID)

	// Remove from experts table
	database.Exec(`DELETE FROM experts WHERE id = ?`, expertID)

	// Remove directory
	if err := os.RemoveAll(expertDir); err != nil {
		return fmt.Errorf("remove expert directory: %w", err)
	}

	return nil
}

// AssignExpert assigns an expert to a project
func AssignExpert(database *db.DB, projectID, expertID string) error {
	// Check if expert exists
	expertPath := filepath.Join(ExpertsDir, expertID, ExpertFileName)
	if _, err := os.Stat(expertPath); os.IsNotExist(err) {
		return fmt.Errorf("expert '%s' not found", expertID)
	}

	// Check if already assigned
	var count int
	database.QueryRow(
		`SELECT COUNT(*) FROM project_experts WHERE project_id = ? AND expert_id = ?`,
		projectID, expertID,
	).Scan(&count)

	if count > 0 {
		return fmt.Errorf("expert '%s' is already assigned to project '%s'", expertID, projectID)
	}

	// Insert assignment
	now := db.TimeNow()
	_, err := database.Exec(
		`INSERT INTO project_experts (project_id, expert_id, assigned_at) VALUES (?, ?, ?)`,
		projectID, expertID, now,
	)
	if err != nil {
		return fmt.Errorf("assign expert: %w", err)
	}

	return nil
}

// UnassignExpert removes an expert from a project
func UnassignExpert(database *db.DB, projectID, expertID string) error {
	result, err := database.Exec(
		`DELETE FROM project_experts WHERE project_id = ? AND expert_id = ?`,
		projectID, expertID,
	)
	if err != nil {
		return fmt.Errorf("unassign expert: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("expert '%s' is not assigned to project '%s'", expertID, projectID)
	}

	return nil
}

// AssignExpertToFeature assigns an expert to a feature
func AssignExpertToFeature(database *db.DB, expertID string, featureID int64) error {
	// Check if expert exists
	expertPath := filepath.Join(ExpertsDir, expertID, ExpertFileName)
	if _, err := os.Stat(expertPath); os.IsNotExist(err) {
		return fmt.Errorf("expert '%s' not found", expertID)
	}

	// Check if feature exists
	var featureCount int
	err := database.QueryRow(`SELECT COUNT(*) FROM features WHERE id = ?`, featureID).Scan(&featureCount)
	if err != nil || featureCount == 0 {
		return fmt.Errorf("feature %d not found", featureID)
	}

	// Check if already assigned
	var count int
	database.QueryRow(
		`SELECT COUNT(*) FROM expert_assignments WHERE expert_id = ? AND feature_id = ?`,
		expertID, featureID,
	).Scan(&count)

	if count > 0 {
		return fmt.Errorf("expert '%s' is already assigned to feature %d", expertID, featureID)
	}

	// Insert assignment
	now := db.TimeNow()
	_, err = database.Exec(
		`INSERT INTO expert_assignments (expert_id, feature_id, created_at) VALUES (?, ?, ?)`,
		expertID, featureID, now,
	)
	if err != nil {
		return fmt.Errorf("assign expert to feature: %w", err)
	}

	return nil
}

// UnassignExpertFromFeature removes an expert from a feature
func UnassignExpertFromFeature(database *db.DB, expertID string, featureID int64) error {
	result, err := database.Exec(
		`DELETE FROM expert_assignments WHERE expert_id = ? AND feature_id = ?`,
		expertID, featureID,
	)
	if err != nil {
		return fmt.Errorf("unassign expert from feature: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("expert '%s' is not assigned to feature %d", expertID, featureID)
	}

	return nil
}

// GetAssignedExpertsByFeature returns experts assigned to a feature with full content
func GetAssignedExpertsByFeature(database *db.DB, featureID int64) ([]model.ExpertInfo, error) {
	rows, err := database.Query(
		`SELECT expert_id FROM expert_assignments WHERE feature_id = ?`,
		featureID,
	)
	if err != nil {
		return nil, fmt.Errorf("query assigned experts by feature: %w", err)
	}
	defer rows.Close()

	var experts []model.ExpertInfo
	for rows.Next() {
		var expertID string
		if err := rows.Scan(&expertID); err != nil {
			continue
		}

		expertPath := filepath.Join(ExpertsDir, expertID, ExpertFileName)
		content, err := os.ReadFile(expertPath)
		if err != nil {
			continue
		}

		// Get name from metadata
		expert, _ := parseExpertMetadata(expertPath)
		name := expertID
		if expert != nil && expert.Name != "" {
			name = expert.Name
		}

		experts = append(experts, model.ExpertInfo{
			ID:      expertID,
			Name:    name,
			Content: string(content),
		})
	}

	return experts, nil
}

// GetAssignedExperts returns experts assigned to a project with full content
func GetAssignedExperts(database *db.DB, projectID string) ([]model.ExpertInfo, error) {
	rows, err := database.Query(
		`SELECT expert_id FROM project_experts WHERE project_id = ?`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("query assigned experts: %w", err)
	}
	defer rows.Close()

	var experts []model.ExpertInfo
	for rows.Next() {
		var expertID string
		if err := rows.Scan(&expertID); err != nil {
			continue
		}

		expertPath := filepath.Join(ExpertsDir, expertID, ExpertFileName)
		content, err := os.ReadFile(expertPath)
		if err != nil {
			continue
		}

		// Get name from metadata
		expert, _ := parseExpertMetadata(expertPath)
		name := expertID
		if expert != nil && expert.Name != "" {
			name = expert.Name
		}

		experts = append(experts, model.ExpertInfo{
			ID:      expertID,
			Name:    name,
			Content: string(content),
		})
	}

	return experts, nil
}

// getExpertTemplate returns the EXPERT.md template
func getExpertTemplate(expertID string) string {
	return fmt.Sprintf(`# Expert: [Expert Name]

## Metadata

| Field       | Value                          |
|-------------|--------------------------------|
| ID          | %s                             |
| Name        | Expert Name                    |
| Version     | 1.0.0                          |
| Domain      | Domain Description             |
| Language    | Language Version               |
| Framework   | Framework Name                 |

## Role Definition

[전문가 역할 설명 - 한 문장]

## Tech Stack

### Core
- **Language**:
- **Framework**:
- **Database**:

### Supporting
- **Auth**:
- **Validation**:
- **Logging**:
- **Testing**:

## Architecture Pattern

[디렉토리 구조]

## Coding Rules

[패턴별 코드 템플릿]

## Error Handling

[에러 처리 규칙]

## Testing Rules

[테스트 코드 규칙]

## Security Checklist

- [ ] 보안 항목들

## References

- [문서 링크]
`, "`"+expertID+"`")
}

// parseExpertMetadata parses EXPERT.md metadata table
func parseExpertMetadata(filePath string) (*model.Expert, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	expert := &model.Expert{}
	scanner := bufio.NewScanner(file)
	inMetadata := false

	for scanner.Scan() {
		line := scanner.Text()

		// Check for Metadata section
		if strings.HasPrefix(line, "## Metadata") {
			inMetadata = true
			continue
		}

		// Stop at next section
		if inMetadata && strings.HasPrefix(line, "## ") {
			break
		}

		// Parse table rows
		if inMetadata && strings.HasPrefix(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				key := strings.TrimSpace(parts[1])
				value := strings.TrimSpace(parts[2])
				value = strings.Trim(value, "`")

				switch key {
				case "ID":
					expert.ID = value
				case "Name":
					expert.Name = value
				case "Version":
					expert.Version = value
				case "Domain":
					expert.Domain = value
				case "Language":
					expert.Language = value
				case "Framework":
					expert.Framework = value
				}
			}
		}
	}

	return expert, nil
}

// SyncExpert syncs expert file with DB
func SyncExpert(database *db.DB, expertID string) error {
	expertPath := filepath.Join(ExpertsDir, expertID, ExpertFileName)

	// Check if file exists
	content, err := os.ReadFile(expertPath)
	if os.IsNotExist(err) {
		// File deleted - restore from DB backup
		return restoreExpertFromDB(database, expertID, expertPath)
	}
	if err != nil {
		return fmt.Errorf("read expert file: %w", err)
	}

	// Calculate hash
	hash := sha256.Sum256(content)
	contentHash := hex.EncodeToString(hash[:])

	// Check if content changed
	var dbHash string
	database.QueryRow("SELECT content_hash FROM experts WHERE id = ?", expertID).Scan(&dbHash)

	if dbHash != contentHash {
		// File modified - update DB backup
		now := db.TimeNow()
		_, err := database.Exec(
			`UPDATE experts SET content = ?, content_hash = ?, updated_at = ? WHERE id = ?`,
			string(content), contentHash, now, expertID,
		)
		if err != nil {
			return fmt.Errorf("update expert content: %w", err)
		}
	}

	return nil
}

// restoreExpertFromDB restores expert file from DB backup
func restoreExpertFromDB(database *db.DB, expertID, expertPath string) error {
	var content string
	err := database.QueryRow("SELECT content FROM experts WHERE id = ?", expertID).Scan(&content)
	if err != nil {
		return fmt.Errorf("expert '%s' not found in DB", expertID)
	}

	if content == "" {
		return fmt.Errorf("no backup content for expert '%s'", expertID)
	}

	// Create directory
	dir := filepath.Dir(expertPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(expertPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write expert file: %w", err)
	}

	return nil
}

// SyncAllExperts syncs all experts
func SyncAllExperts(database *db.DB) error {
	// Get all expert IDs from DB
	rows, err := database.Query("SELECT id FROM experts")
	if err != nil {
		return err
	}
	defer rows.Close()

	var expertIDs []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		expertIDs = append(expertIDs, id)
	}

	// Sync each expert
	for _, id := range expertIDs {
		SyncExpert(database, id) // Ignore errors, continue with others
	}

	return nil
}
