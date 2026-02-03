package test

import (
	"strings"
	"testing"

	"parkjunwoo.com/claritask/internal/service"
)

func TestParseFDL(t *testing.T) {
	validYAML := `
feature: login
description: User authentication feature
models:
  - name: User
    table: users
    fields:
      - name: id
        type: uuid
        constraints: pk
      - name: email
        type: string
service:
  - name: authenticate
    desc: Authenticate user
    input:
      email: string
      password: string
    steps:
      - Validate input
      - Check credentials
api:
  - path: /api/auth/login
    method: POST
    use: service.authenticate
    response:
      200: { token: string }
ui:
  - component: LoginForm
    type: Organism
    state:
      - email
      - password
`

	spec, err := service.ParseFDL(validYAML)
	if err != nil {
		t.Fatalf("ParseFDL failed: %v", err)
	}

	if spec.Feature != "login" {
		t.Errorf("Expected feature 'login', got '%s'", spec.Feature)
	}
	if spec.Description != "User authentication feature" {
		t.Errorf("Expected description 'User authentication feature', got '%s'", spec.Description)
	}
	if len(spec.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(spec.Models))
	}
	if len(spec.Service) != 1 {
		t.Errorf("Expected 1 service, got %d", len(spec.Service))
	}
	if len(spec.API) != 1 {
		t.Errorf("Expected 1 API, got %d", len(spec.API))
	}
	if len(spec.UI) != 1 {
		t.Errorf("Expected 1 UI, got %d", len(spec.UI))
	}
}

func TestParseFDLInvalidYAML(t *testing.T) {
	invalidYAML := `
feature: login
  invalid: yaml
    - broken
`

	_, err := service.ParseFDL(invalidYAML)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestValidateFDLValid(t *testing.T) {
	spec := &service.FDLSpec{
		Feature:     "login",
		Description: "User authentication",
		Models: []service.FDLModel{
			{
				Name:  "User",
				Table: "users",
				Fields: []service.FDLField{
					{Name: "id", Type: "uuid"},
				},
			},
		},
		Service: []service.FDLService{
			{
				Name:  "authenticate",
				Desc:  "Auth service",
				Steps: []string{"Step 1"},
			},
		},
		API: []service.FDLAPI{
			{
				Path:   "/api/login",
				Method: "POST",
				Use:    "service.authenticate",
				Response: map[string]interface{}{
					"200": map[string]interface{}{"token": "string"},
				},
			},
		},
		UI: []service.FDLUI{
			{
				Component: "LoginForm",
				Type:      "Organism",
			},
		},
	}

	err := service.ValidateFDL(spec)
	if err != nil {
		t.Errorf("ValidateFDL failed for valid spec: %v", err)
	}
}

func TestValidateFDLInvalidFeatureName(t *testing.T) {
	tests := []struct {
		name        string
		featureName string
	}{
		{"empty name", ""},
		{"starts with number", "123login"},
		{"contains spaces", "login form"},
		{"special characters", "login@form"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &service.FDLSpec{Feature: tt.featureName}
			err := service.ValidateFDL(spec)
			if err == nil {
				t.Errorf("Expected error for feature name '%s'", tt.featureName)
			}
		})
	}
}

func TestValidateFDLDuplicateModel(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "test",
		Models: []service.FDLModel{
			{Name: "User", Table: "users", Fields: []service.FDLField{{Name: "id", Type: "uuid"}}},
			{Name: "User", Table: "users2", Fields: []service.FDLField{{Name: "id", Type: "uuid"}}},
		},
	}

	err := service.ValidateFDL(spec)
	if err == nil {
		t.Error("Expected error for duplicate model name")
	}
	if !strings.Contains(err.Error(), "duplicate model") {
		t.Errorf("Expected 'duplicate model' error, got: %v", err)
	}
}

func TestValidateFDLModelMissingTable(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "test",
		Models: []service.FDLModel{
			{Name: "User", Table: "", Fields: []service.FDLField{{Name: "id", Type: "uuid"}}},
		},
	}

	err := service.ValidateFDL(spec)
	if err == nil {
		t.Error("Expected error for missing table name")
	}
}

func TestValidateFDLModelEmptyFields(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "test",
		Models: []service.FDLModel{
			{Name: "User", Table: "users", Fields: []service.FDLField{}},
		},
	}

	err := service.ValidateFDL(spec)
	if err == nil {
		t.Error("Expected error for empty fields")
	}
}

func TestValidateFDLServiceEmptySteps(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "test",
		Service: []service.FDLService{
			{Name: "auth", Steps: []string{}},
		},
	}

	err := service.ValidateFDL(spec)
	if err == nil {
		t.Error("Expected error for empty steps")
	}
	if !strings.Contains(err.Error(), "at least one step") {
		t.Errorf("Expected 'at least one step' error, got: %v", err)
	}
}

func TestValidateFDLDuplicateService(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "test",
		Service: []service.FDLService{
			{Name: "auth", Steps: []string{"Step 1"}},
			{Name: "auth", Steps: []string{"Step 2"}},
		},
	}

	err := service.ValidateFDL(spec)
	if err == nil {
		t.Error("Expected error for duplicate service name")
	}
}

func TestValidateFDLAPIInvalidMethod(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "test",
		API: []service.FDLAPI{
			{Path: "/api/test", Method: "INVALID"},
		},
	}

	err := service.ValidateFDL(spec)
	if err == nil {
		t.Error("Expected error for invalid HTTP method")
	}
}

func TestValidateFDLAPINonExistentService(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "test",
		Service: []service.FDLService{
			{Name: "existingService", Steps: []string{"Step 1"}},
		},
		API: []service.FDLAPI{
			{Path: "/api/test", Method: "POST", Use: "service.nonExistentService"},
		},
	}

	err := service.ValidateFDL(spec)
	if err == nil {
		t.Error("Expected error for non-existent service reference")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestValidateFDLAPIMissingPath(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "test",
		API: []service.FDLAPI{
			{Path: "", Method: "POST"},
		},
	}

	err := service.ValidateFDL(spec)
	if err == nil {
		t.Error("Expected error for missing API path")
	}
}

func TestValidateFDLUIDuplicateComponent(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "test",
		UI: []service.FDLUI{
			{Component: "LoginForm", Type: "Organism"},
			{Component: "LoginForm", Type: "Molecule"},
		},
	}

	err := service.ValidateFDL(spec)
	if err == nil {
		t.Error("Expected error for duplicate component name")
	}
}

func TestValidateFDLUIInvalidType(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "test",
		UI: []service.FDLUI{
			{Component: "LoginForm", Type: "InvalidType"},
		},
	}

	err := service.ValidateFDL(spec)
	if err == nil {
		t.Error("Expected error for invalid UI type")
	}
}

func TestCalculateFDLHash(t *testing.T) {
	fdl1 := "feature: login"
	fdl2 := "feature: login"
	fdl3 := "feature: different"

	hash1 := service.CalculateFDLHashFromSpec(fdl1)
	hash2 := service.CalculateFDLHashFromSpec(fdl2)
	hash3 := service.CalculateFDLHashFromSpec(fdl3)

	if hash1 != hash2 {
		t.Error("Same input should produce same hash")
	}

	if hash1 == hash3 {
		t.Error("Different input should produce different hash")
	}

	// Check hash format (SHA256 = 64 hex characters)
	if len(hash1) != 64 {
		t.Errorf("Expected 64 character hash, got %d", len(hash1))
	}
}

func TestGenerateFDLTemplate(t *testing.T) {
	template := service.GenerateFDLTemplate("login")

	if !strings.Contains(template, "feature: login") {
		t.Error("Template should contain feature name")
	}

	// Verify template is valid YAML
	spec, err := service.ParseFDL(template)
	if err != nil {
		t.Fatalf("Generated template should be valid YAML: %v", err)
	}

	if spec.Feature != "login" {
		t.Errorf("Expected feature 'login', got '%s'", spec.Feature)
	}
}

func TestExtractTaskMappings(t *testing.T) {
	spec := &service.FDLSpec{
		Feature:     "login",
		Description: "User authentication",
		Models: []service.FDLModel{
			{Name: "User", Table: "users", Fields: []service.FDLField{{Name: "id", Type: "uuid"}}},
		},
		Service: []service.FDLService{
			{Name: "authenticate", Desc: "Auth service", Steps: []string{"Step 1"}},
		},
		API: []service.FDLAPI{
			{Path: "/api/login", Method: "POST", Use: "service.authenticate"},
		},
		UI: []service.FDLUI{
			{Component: "LoginForm", Type: "Organism"},
		},
	}

	tech := map[string]interface{}{
		"backend":  "go",
		"frontend": "react",
	}

	mappings, err := service.ExtractTaskMappings(spec, tech)
	if err != nil {
		t.Fatalf("ExtractTaskMappings failed: %v", err)
	}

	// Should have 4 tasks: 1 model + 1 service + 1 API + 1 UI
	if len(mappings) != 4 {
		t.Errorf("Expected 4 mappings, got %d", len(mappings))
	}

	// Check model task
	modelTask := mappings[0]
	if modelTask.Layer != "model" {
		t.Errorf("First task should be model layer, got %s", modelTask.Layer)
	}
	if !strings.Contains(modelTask.TargetFile, ".go") {
		t.Error("Go backend should generate .go files")
	}

	// Check service task dependencies
	serviceTask := mappings[1]
	if serviceTask.Layer != "service" {
		t.Errorf("Second task should be service layer, got %s", serviceTask.Layer)
	}
	if len(serviceTask.Dependencies) == 0 {
		t.Error("Service task should depend on model tasks")
	}

	// Check API task
	apiTask := mappings[2]
	if apiTask.Layer != "api" {
		t.Errorf("Third task should be api layer, got %s", apiTask.Layer)
	}

	// Check UI task
	uiTask := mappings[3]
	if uiTask.Layer != "ui" {
		t.Errorf("Fourth task should be ui layer, got %s", uiTask.Layer)
	}
	if !strings.Contains(uiTask.TargetFile, ".tsx") {
		t.Error("React frontend should generate .tsx files")
	}
}

func TestExtractTaskMappingsPython(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "login",
		Models: []service.FDLModel{
			{Name: "User", Table: "users", Fields: []service.FDLField{{Name: "id", Type: "uuid"}}},
		},
		Service: []service.FDLService{
			{Name: "authenticate", Desc: "Auth service", Steps: []string{"Step 1"}},
		},
	}

	tech := map[string]interface{}{
		"backend": "python",
	}

	mappings, err := service.ExtractTaskMappings(spec, tech)
	if err != nil {
		t.Fatalf("ExtractTaskMappings failed: %v", err)
	}

	// Check that Python backend generates .py files
	for _, m := range mappings {
		if !strings.Contains(m.TargetFile, ".py") {
			t.Errorf("Python backend should generate .py files, got %s", m.TargetFile)
		}
	}
}

func TestExtractTaskMappingsVue(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "login",
		UI: []service.FDLUI{
			{Component: "LoginForm", Type: "Organism"},
		},
	}

	tech := map[string]interface{}{
		"frontend": "vue",
	}

	mappings, err := service.ExtractTaskMappings(spec, tech)
	if err != nil {
		t.Fatalf("ExtractTaskMappings failed: %v", err)
	}

	if len(mappings) != 1 {
		t.Fatalf("Expected 1 mapping, got %d", len(mappings))
	}

	if !strings.Contains(mappings[0].TargetFile, ".vue") {
		t.Errorf("Vue frontend should generate .vue files, got %s", mappings[0].TargetFile)
	}
}

func TestGetFDLInfo(t *testing.T) {
	spec := &service.FDLSpec{
		Feature:     "login",
		Description: "User authentication",
		Models: []service.FDLModel{
			{Name: "User", Table: "users", Fields: []service.FDLField{{Name: "id", Type: "uuid"}}},
		},
		Service: []service.FDLService{
			{Name: "authenticate", Desc: "Auth service", Input: map[string]interface{}{"email": "string"}, Steps: []string{"Step 1"}},
		},
		API: []service.FDLAPI{
			{Path: "/api/login", Method: "POST", Use: "service.authenticate", Response: map[string]interface{}{"200": "ok"}},
		},
		UI: []service.FDLUI{
			{Component: "LoginForm", Type: "Organism", Props: map[string]interface{}{"onSubmit": "function"}},
		},
	}

	info := service.GetFDLInfo(spec)

	if info.Feature != "login" {
		t.Errorf("Expected feature 'login', got '%s'", info.Feature)
	}

	if info.Models == nil || info.Models["User"] == nil {
		t.Error("Expected User model in info")
	}

	if info.Service == nil || info.Service["authenticate"] == nil {
		t.Error("Expected authenticate service in info")
	}

	if info.API == nil || info.API["POST /api/login"] == nil {
		t.Error("Expected POST /api/login in info")
	}

	if info.UI == nil || info.UI["LoginForm"] == nil {
		t.Error("Expected LoginForm in info")
	}
}

func TestGetFDLInfoEmptySpec(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "empty",
	}

	info := service.GetFDLInfo(spec)

	if info.Feature != "empty" {
		t.Errorf("Expected feature 'empty', got '%s'", info.Feature)
	}

	if info.Models != nil {
		t.Error("Expected nil models for empty spec")
	}
	if info.Service != nil {
		t.Error("Expected nil service for empty spec")
	}
	if info.API != nil {
		t.Error("Expected nil API for empty spec")
	}
	if info.UI != nil {
		t.Error("Expected nil UI for empty spec")
	}
}

func TestValidateFDLValidMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "get", "post"}

	for _, method := range methods {
		spec := &service.FDLSpec{
			Feature: "test",
			API: []service.FDLAPI{
				{Path: "/api/test", Method: method},
			},
		}

		err := service.ValidateFDL(spec)
		if err != nil {
			t.Errorf("Method %s should be valid: %v", method, err)
		}
	}
}

func TestValidateFDLUIValidTypes(t *testing.T) {
	types := []string{"Page", "Organism", "Molecule", "Atom", ""}

	for _, typ := range types {
		spec := &service.FDLSpec{
			Feature: "test",
			UI: []service.FDLUI{
				{Component: "TestComponent", Type: typ},
			},
		}

		err := service.ValidateFDL(spec)
		if err != nil {
			t.Errorf("UI type '%s' should be valid: %v", typ, err)
		}
	}
}

func TestValidateFDLModelDuplicateField(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "test",
		Models: []service.FDLModel{
			{
				Name:  "User",
				Table: "users",
				Fields: []service.FDLField{
					{Name: "id", Type: "uuid"},
					{Name: "id", Type: "string"}, // duplicate
				},
			},
		},
	}

	err := service.ValidateFDL(spec)
	if err == nil {
		t.Error("Expected error for duplicate field name")
	}
	if !strings.Contains(err.Error(), "duplicate field") {
		t.Errorf("Expected 'duplicate field' error, got: %v", err)
	}
}

func TestValidateFDLModelFieldMissingType(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "test",
		Models: []service.FDLModel{
			{
				Name:  "User",
				Table: "users",
				Fields: []service.FDLField{
					{Name: "id", Type: ""},
				},
			},
		},
	}

	err := service.ValidateFDL(spec)
	if err == nil {
		t.Error("Expected error for missing field type")
	}
}

func TestValidateFDLAPIUseFormat(t *testing.T) {
	spec := &service.FDLSpec{
		Feature: "test",
		API: []service.FDLAPI{
			{Path: "/api/test", Method: "POST", Use: "invalidFormat"},
		},
	}

	err := service.ValidateFDL(spec)
	if err == nil {
		t.Error("Expected error for invalid use format")
	}
	if !strings.Contains(err.Error(), "service.FunctionName") {
		t.Errorf("Expected hint about correct format, got: %v", err)
	}
}
