package test

import (
	"testing"

	"parkjunwoo.com/claritask/internal/service"
)

func TestCreateFeature(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Create project first
	service.CreateProject(database, "test-project", "Test Project", "Description")

	featureID, err := service.CreateFeature(database, "test-project", "Login", "User authentication feature")
	if err != nil {
		t.Fatalf("CreateFeature failed: %v", err)
	}

	if featureID <= 0 {
		t.Errorf("Expected positive feature ID, got %d", featureID)
	}
}

func TestGetFeature(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")

	featureID, err := service.CreateFeature(database, "test-project", "Login", "User authentication feature")
	if err != nil {
		t.Fatalf("CreateFeature failed: %v", err)
	}

	feature, err := service.GetFeature(database, featureID)
	if err != nil {
		t.Fatalf("GetFeature failed: %v", err)
	}

	if feature.Name != "Login" {
		t.Errorf("Expected name 'Login', got '%s'", feature.Name)
	}
	if feature.Description != "User authentication feature" {
		t.Errorf("Expected description 'User authentication feature', got '%s'", feature.Description)
	}
	if feature.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", feature.Status)
	}
}

func TestGetFeatureNotFound(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := service.GetFeature(database, 999)
	if err == nil {
		t.Error("Expected error for non-existent feature")
	}
}

func TestListFeatures(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")

	service.CreateFeature(database, "test-project", "Feature1", "Description 1")
	service.CreateFeature(database, "test-project", "Feature2", "Description 2")

	features, err := service.ListFeatures(database, "test-project")
	if err != nil {
		t.Fatalf("ListFeatures failed: %v", err)
	}

	if len(features) != 2 {
		t.Errorf("Expected 2 features, got %d", len(features))
	}
}

func TestUpdateFeature(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")

	featureID, _ := service.CreateFeature(database, "test-project", "Login", "Original description")
	feature, _ := service.GetFeature(database, featureID)

	feature.Description = "Updated description"
	err := service.UpdateFeature(database, feature)
	if err != nil {
		t.Fatalf("UpdateFeature failed: %v", err)
	}

	updated, _ := service.GetFeature(database, featureID)
	if updated.Description != "Updated description" {
		t.Errorf("Expected 'Updated description', got '%s'", updated.Description)
	}
}

func TestDeleteFeature(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")

	featureID, _ := service.CreateFeature(database, "test-project", "Login", "Description")

	err := service.DeleteFeature(database, featureID)
	if err != nil {
		t.Fatalf("DeleteFeature failed: %v", err)
	}

	_, err = service.GetFeature(database, featureID)
	if err == nil {
		t.Error("Expected error when getting deleted feature")
	}
}

func TestStartFeature(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")

	featureID, _ := service.CreateFeature(database, "test-project", "Login", "Description")

	err := service.StartFeature(database, featureID)
	if err != nil {
		t.Fatalf("StartFeature failed: %v", err)
	}

	feature, _ := service.GetFeature(database, featureID)
	if feature.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", feature.Status)
	}
}

func TestCompleteFeature(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")

	featureID, _ := service.CreateFeature(database, "test-project", "Login", "Description")
	service.StartFeature(database, featureID)

	err := service.CompleteFeature(database, featureID)
	if err != nil {
		t.Fatalf("CompleteFeature failed: %v", err)
	}

	feature, _ := service.GetFeature(database, featureID)
	if feature.Status != "done" {
		t.Errorf("Expected status 'done', got '%s'", feature.Status)
	}
}

func TestSetAndGetFeatureFDL(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")

	featureID, _ := service.CreateFeature(database, "test-project", "Login", "Description")

	fdl := "feature: login\ndescription: User authentication"
	err := service.SetFeatureFDL(database, featureID, fdl)
	if err != nil {
		t.Fatalf("SetFeatureFDL failed: %v", err)
	}

	retrieved, err := service.GetFeatureFDL(database, featureID)
	if err != nil {
		t.Fatalf("GetFeatureFDL failed: %v", err)
	}

	if retrieved != fdl {
		t.Errorf("FDL mismatch: expected '%s', got '%s'", fdl, retrieved)
	}
}

func TestCalculateFDLHashFeature(t *testing.T) {
	fdl := "feature: login"
	hash1 := service.CalculateFDLHash(fdl)
	hash2 := service.CalculateFDLHash(fdl)

	if hash1 != hash2 {
		t.Error("Same input should produce same hash")
	}

	hash3 := service.CalculateFDLHash("feature: different")
	if hash1 == hash3 {
		t.Error("Different input should produce different hash")
	}
}

func TestFeatureEdge(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")

	feature1, _ := service.CreateFeature(database, "test-project", "Feature1", "Description 1")
	feature2, _ := service.CreateFeature(database, "test-project", "Feature2", "Description 2")

	// Add edge: feature2 depends on feature1
	err := service.AddFeatureEdge(database, feature2, feature1)
	if err != nil {
		t.Fatalf("AddFeatureEdge failed: %v", err)
	}

	// Check dependencies
	deps, err := service.GetFeatureDependencies(database, feature2)
	if err != nil {
		t.Fatalf("GetFeatureDependencies failed: %v", err)
	}

	if len(deps) != 1 || deps[0].ID != feature1 {
		t.Errorf("Expected feature2 to depend on feature1")
	}

	// Remove edge
	err = service.RemoveFeatureEdge(database, feature2, feature1)
	if err != nil {
		t.Fatalf("RemoveFeatureEdge failed: %v", err)
	}

	deps, _ = service.GetFeatureDependencies(database, feature2)
	if len(deps) != 0 {
		t.Error("Expected no dependencies after removal")
	}
}

func TestFeatureCycleDetection(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")

	feature1, _ := service.CreateFeature(database, "test-project", "Feature1", "")
	feature2, _ := service.CreateFeature(database, "test-project", "Feature2", "")
	feature3, _ := service.CreateFeature(database, "test-project", "Feature3", "")

	// Create chain: feature3 -> feature2 -> feature1
	service.AddFeatureEdge(database, feature2, feature1)
	service.AddFeatureEdge(database, feature3, feature2)

	// Try to add cycle: feature1 -> feature3 would create cycle
	hasCycle, _, _ := service.CheckFeatureCycle(database, feature1, feature3)
	if !hasCycle {
		t.Error("Expected cycle detection when adding feature1 -> feature3")
	}
}
