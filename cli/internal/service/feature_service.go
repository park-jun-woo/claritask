package service

import (
	"crypto/sha256"
	"fmt"

	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/model"
)

// CreateFeature creates a new feature
func CreateFeature(database *db.DB, projectID, name, description string) (int64, error) {
	now := db.TimeNow()
	result, err := database.Exec(
		`INSERT INTO features (project_id, name, description, status, created_at) VALUES (?, ?, ?, 'pending', ?)`,
		projectID, name, description, now,
	)
	if err != nil {
		return 0, fmt.Errorf("create feature: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}
	return id, nil
}

// GetFeature retrieves a feature by ID
func GetFeature(database *db.DB, id int64) (*model.Feature, error) {
	row := database.QueryRow(
		`SELECT id, project_id, name, description, spec, fdl, fdl_hash, skeleton_generated, status, version, created_at
		 FROM features WHERE id = ?`, id,
	)
	var f model.Feature
	var createdAt string
	var skeletonGenerated int
	err := row.Scan(&f.ID, &f.ProjectID, &f.Name, &f.Description, &f.Spec, &f.FDL, &f.FDLHash, &skeletonGenerated, &f.Status, &f.Version, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("get feature: %w", err)
	}
	f.SkeletonGenerated = skeletonGenerated == 1
	f.CreatedAt, _ = db.ParseTime(createdAt)
	return &f, nil
}

// ListFeatures retrieves all features for a project
func ListFeatures(database *db.DB, projectID string) ([]model.Feature, error) {
	rows, err := database.Query(
		`SELECT id, project_id, name, description, spec, fdl, fdl_hash, skeleton_generated, status, version, created_at
		 FROM features WHERE project_id = ? ORDER BY id`, projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list features: %w", err)
	}
	defer rows.Close()

	var features []model.Feature
	for rows.Next() {
		var f model.Feature
		var createdAt string
		var skeletonGenerated int
		if err := rows.Scan(&f.ID, &f.ProjectID, &f.Name, &f.Description, &f.Spec, &f.FDL, &f.FDLHash, &skeletonGenerated, &f.Status, &f.Version, &createdAt); err != nil {
			return nil, fmt.Errorf("scan feature: %w", err)
		}
		f.SkeletonGenerated = skeletonGenerated == 1
		f.CreatedAt, _ = db.ParseTime(createdAt)
		features = append(features, f)
	}
	return features, nil
}

// UpdateFeature updates an existing feature
func UpdateFeature(database *db.DB, feature *model.Feature) error {
	skeletonGenerated := 0
	if feature.SkeletonGenerated {
		skeletonGenerated = 1
	}
	_, err := database.Exec(
		`UPDATE features SET name = ?, description = ?, spec = ?, fdl = ?, fdl_hash = ?, skeleton_generated = ?, status = ? WHERE id = ?`,
		feature.Name, feature.Description, feature.Spec, feature.FDL, feature.FDLHash, skeletonGenerated, feature.Status, feature.ID,
	)
	if err != nil {
		return fmt.Errorf("update feature: %w", err)
	}
	return nil
}

// DeleteFeature deletes a feature
func DeleteFeature(database *db.DB, id int64) error {
	// Delete related edges first
	_, err := database.Exec(`DELETE FROM feature_edges WHERE from_feature_id = ? OR to_feature_id = ?`, id, id)
	if err != nil {
		return fmt.Errorf("delete feature edges: %w", err)
	}

	_, err = database.Exec(`DELETE FROM features WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete feature: %w", err)
	}
	return nil
}

// StartFeature changes feature status to active
func StartFeature(database *db.DB, id int64) error {
	_, err := database.Exec(`UPDATE features SET status = 'active' WHERE id = ? AND status = 'pending'`, id)
	if err != nil {
		return fmt.Errorf("start feature: %w", err)
	}
	return nil
}

// CompleteFeature changes feature status to done
func CompleteFeature(database *db.DB, id int64) error {
	_, err := database.Exec(`UPDATE features SET status = 'done' WHERE id = ? AND status = 'active'`, id)
	if err != nil {
		return fmt.Errorf("complete feature: %w", err)
	}
	return nil
}

// SetFeatureSpec sets the feature spec
func SetFeatureSpec(database *db.DB, id int64, spec string) error {
	_, err := database.Exec(`UPDATE features SET spec = ? WHERE id = ?`, spec, id)
	if err != nil {
		return fmt.Errorf("set feature spec: %w", err)
	}
	return nil
}

// GetFeatureSpec retrieves the feature spec
func GetFeatureSpec(database *db.DB, id int64) (string, error) {
	row := database.QueryRow(`SELECT spec FROM features WHERE id = ?`, id)
	var spec string
	err := row.Scan(&spec)
	if err != nil {
		return "", fmt.Errorf("get feature spec: %w", err)
	}
	return spec, nil
}

// SetFeatureFDL sets the feature FDL
func SetFeatureFDL(database *db.DB, id int64, fdl string) error {
	hash := CalculateFDLHash(fdl)
	_, err := database.Exec(`UPDATE features SET fdl = ?, fdl_hash = ? WHERE id = ?`, fdl, hash, id)
	if err != nil {
		return fmt.Errorf("set feature fdl: %w", err)
	}
	return nil
}

// GetFeatureFDL retrieves the feature FDL
func GetFeatureFDL(database *db.DB, id int64) (string, error) {
	row := database.QueryRow(`SELECT fdl FROM features WHERE id = ?`, id)
	var fdl string
	err := row.Scan(&fdl)
	if err != nil {
		return "", fmt.Errorf("get feature fdl: %w", err)
	}
	return fdl, nil
}

// CalculateFDLHash calculates SHA256 hash of FDL
func CalculateFDLHash(fdl string) string {
	hash := sha256.Sum256([]byte(fdl))
	return fmt.Sprintf("%x", hash)
}

// AddFeatureEdge adds a dependency edge between features
// from depends on to (to must be completed before from can start)
func AddFeatureEdge(database *db.DB, fromID, toID int64) error {
	// Check for cycle
	hasCycle, _, err := CheckFeatureCycle(database, fromID, toID)
	if err != nil {
		return fmt.Errorf("check cycle: %w", err)
	}
	if hasCycle {
		return fmt.Errorf("adding edge would create a cycle")
	}

	now := db.TimeNow()
	_, err = database.Exec(
		`INSERT INTO feature_edges (from_feature_id, to_feature_id, created_at) VALUES (?, ?, ?)`,
		fromID, toID, now,
	)
	if err != nil {
		return fmt.Errorf("add feature edge: %w", err)
	}
	return nil
}

// RemoveFeatureEdge removes a dependency edge between features
func RemoveFeatureEdge(database *db.DB, fromID, toID int64) error {
	_, err := database.Exec(
		`DELETE FROM feature_edges WHERE from_feature_id = ? AND to_feature_id = ?`,
		fromID, toID,
	)
	if err != nil {
		return fmt.Errorf("remove feature edge: %w", err)
	}
	return nil
}

// GetFeatureEdges retrieves all edges for a feature
func GetFeatureEdges(database *db.DB, featureID int64) ([]model.FeatureEdge, error) {
	rows, err := database.Query(
		`SELECT from_feature_id, to_feature_id, created_at FROM feature_edges
		 WHERE from_feature_id = ? OR to_feature_id = ?`, featureID, featureID,
	)
	if err != nil {
		return nil, fmt.Errorf("get feature edges: %w", err)
	}
	defer rows.Close()

	var edges []model.FeatureEdge
	for rows.Next() {
		var e model.FeatureEdge
		var createdAt string
		if err := rows.Scan(&e.FromFeatureID, &e.ToFeatureID, &createdAt); err != nil {
			return nil, fmt.Errorf("scan edge: %w", err)
		}
		e.CreatedAt, _ = db.ParseTime(createdAt)
		edges = append(edges, e)
	}
	return edges, nil
}

// GetFeatureDependencies retrieves features that this feature depends on
func GetFeatureDependencies(database *db.DB, featureID int64) ([]model.Feature, error) {
	rows, err := database.Query(
		`SELECT f.id, f.project_id, f.name, f.description, f.spec, f.fdl, f.fdl_hash, f.skeleton_generated, f.status, f.version, f.created_at
		 FROM features f
		 JOIN feature_edges e ON f.id = e.to_feature_id
		 WHERE e.from_feature_id = ?`, featureID,
	)
	if err != nil {
		return nil, fmt.Errorf("get feature dependencies: %w", err)
	}
	defer rows.Close()

	var features []model.Feature
	for rows.Next() {
		var f model.Feature
		var createdAt string
		var skeletonGenerated int
		if err := rows.Scan(&f.ID, &f.ProjectID, &f.Name, &f.Description, &f.Spec, &f.FDL, &f.FDLHash, &skeletonGenerated, &f.Status, &f.Version, &createdAt); err != nil {
			return nil, fmt.Errorf("scan feature: %w", err)
		}
		f.SkeletonGenerated = skeletonGenerated == 1
		f.CreatedAt, _ = db.ParseTime(createdAt)
		features = append(features, f)
	}
	return features, nil
}

// GetFeatureDependents retrieves features that depend on this feature
func GetFeatureDependents(database *db.DB, featureID int64) ([]model.Feature, error) {
	rows, err := database.Query(
		`SELECT f.id, f.project_id, f.name, f.description, f.spec, f.fdl, f.fdl_hash, f.skeleton_generated, f.status, f.version, f.created_at
		 FROM features f
		 JOIN feature_edges e ON f.id = e.from_feature_id
		 WHERE e.to_feature_id = ?`, featureID,
	)
	if err != nil {
		return nil, fmt.Errorf("get feature dependents: %w", err)
	}
	defer rows.Close()

	var features []model.Feature
	for rows.Next() {
		var f model.Feature
		var createdAt string
		var skeletonGenerated int
		if err := rows.Scan(&f.ID, &f.ProjectID, &f.Name, &f.Description, &f.Spec, &f.FDL, &f.FDLHash, &skeletonGenerated, &f.Status, &f.Version, &createdAt); err != nil {
			return nil, fmt.Errorf("scan feature: %w", err)
		}
		f.SkeletonGenerated = skeletonGenerated == 1
		f.CreatedAt, _ = db.ParseTime(createdAt)
		features = append(features, f)
	}
	return features, nil
}

// CheckFeatureCycle checks if adding an edge would create a cycle
// Uses DFS to detect if there's a path from toID to fromID
func CheckFeatureCycle(database *db.DB, fromID, toID int64) (bool, []int64, error) {
	// If adding edge from -> to, check if there's already a path to -> from
	visited := make(map[int64]bool)
	path := []int64{}

	var dfs func(current int64) bool
	dfs = func(current int64) bool {
		if current == fromID {
			return true
		}
		if visited[current] {
			return false
		}
		visited[current] = true
		path = append(path, current)

		// Get dependencies of current (features that current depends on)
		deps, err := GetFeatureDependencies(database, current)
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

// FeatureListItem represents a feature in list view with stats
type FeatureListItem struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Spec        string  `json:"spec,omitempty"`
	Status      string  `json:"status"`
	TasksTotal  int     `json:"tasks_total"`
	TasksDone   int     `json:"tasks_done"`
	DependsOn   []int64 `json:"depends_on,omitempty"`
}

// ListFeaturesWithStats retrieves features with task statistics
func ListFeaturesWithStats(database *db.DB, projectID string) ([]FeatureListItem, error) {
	features, err := ListFeatures(database, projectID)
	if err != nil {
		return nil, err
	}

	var items []FeatureListItem
	for _, f := range features {
		item := FeatureListItem{
			ID:          f.ID,
			Name:        f.Name,
			Description: f.Description,
			Spec:        f.Spec,
			Status:      f.Status,
		}

		// Get task counts
		row := database.QueryRow(
			`SELECT COUNT(*), COALESCE(SUM(CASE WHEN status = 'done' THEN 1 ELSE 0 END), 0)
			 FROM tasks WHERE feature_id = ?`, f.ID,
		)
		err := row.Scan(&item.TasksTotal, &item.TasksDone)
		if err != nil {
			item.TasksTotal = 0
			item.TasksDone = 0
		}

		// Get dependencies
		deps, err := GetFeatureDependencies(database, f.ID)
		if err == nil && len(deps) > 0 {
			for _, d := range deps {
				item.DependsOn = append(item.DependsOn, d.ID)
			}
		}

		items = append(items, item)
	}

	return items, nil
}
