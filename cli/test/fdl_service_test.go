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
				Steps: []interface{}{"Step 1"},
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
			{Name: "auth", Steps: []interface{}{}},
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
			{Name: "auth", Steps: []interface{}{"Step 1"}},
			{Name: "auth", Steps: []interface{}{"Step 2"}},
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
			{Name: "existingService", Steps: []interface{}{"Step 1"}},
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
			{Name: "authenticate", Desc: "Auth service", Steps: []interface{}{"Step 1"}},
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
			{Name: "authenticate", Desc: "Auth service", Steps: []interface{}{"Step 1"}},
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
			{Name: "authenticate", Desc: "Auth service", Input: map[string]interface{}{"email": "string"}, Steps: []interface{}{"Step 1"}},
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

func TestParseAndValidateField(t *testing.T) {
	tests := []struct {
		name        string
		field       service.FDLField
		wantErr     bool
		wantType    string
		wantOption  string
		wantPK      bool
		wantFK      bool
		wantFKRef   string
	}{
		{
			name:       "simple uuid",
			field:      service.FDLField{Name: "id", Type: "uuid", Constraints: "pk"},
			wantType:   "uuid",
			wantOption: "",
			wantPK:     true,
		},
		{
			name:       "string with length",
			field:      service.FDLField{Name: "name", Type: "string(50)", Constraints: "required"},
			wantType:   "string",
			wantOption: "50",
		},
		{
			name:       "decimal with precision",
			field:      service.FDLField{Name: "amount", Type: "decimal(10,2)", Constraints: "required"},
			wantType:   "decimal",
			wantOption: "10,2",
		},
		{
			name:       "enum with values",
			field:      service.FDLField{Name: "role", Type: "enum(admin,user,guest)", Constraints: "required"},
			wantType:   "enum",
			wantOption: "admin,user,guest",
		},
		{
			name:      "foreign key",
			field:     service.FDLField{Name: "user_id", Type: "uuid", Constraints: "fk(users.id), onDelete: cascade"},
			wantType:  "uuid",
			wantFK:    true,
			wantFKRef: "users.id",
		},
		{
			name:    "invalid type",
			field:   service.FDLField{Name: "foo", Type: "invalidtype"},
			wantErr: true,
		},
		{
			name:    "decimal without precision",
			field:   service.FDLField{Name: "amount", Type: "decimal(10)"},
			wantErr: true,
		},
		{
			name:    "enum without values",
			field:   service.FDLField{Name: "role", Type: "enum"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ParseAndValidateField(&tt.field)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAndValidateField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if tt.field.ParsedType != tt.wantType {
				t.Errorf("ParsedType = %v, want %v", tt.field.ParsedType, tt.wantType)
			}
			if tt.field.TypeOption != tt.wantOption {
				t.Errorf("TypeOption = %v, want %v", tt.field.TypeOption, tt.wantOption)
			}
			if tt.field.ParsedConstraint.IsPK != tt.wantPK {
				t.Errorf("IsPK = %v, want %v", tt.field.ParsedConstraint.IsPK, tt.wantPK)
			}
			if tt.field.ParsedConstraint.IsFK != tt.wantFK {
				t.Errorf("IsFK = %v, want %v", tt.field.ParsedConstraint.IsFK, tt.wantFK)
			}
			if tt.wantFKRef != "" && tt.field.ParsedConstraint.FKRef != tt.wantFKRef {
				t.Errorf("FKRef = %v, want %v", tt.field.ParsedConstraint.FKRef, tt.wantFKRef)
			}
		})
	}
}

func TestValidateModelAdvanced(t *testing.T) {
	allModels := []service.FDLModel{
		{
			Name:  "User",
			Table: "users",
			Fields: []service.FDLField{
				{Name: "id", Type: "uuid", Constraints: "pk"},
			},
		},
		{
			Name:  "Post",
			Table: "posts",
			Fields: []service.FDLField{
				{Name: "id", Type: "uuid", Constraints: "pk"},
				{Name: "author_id", Type: "uuid", Constraints: "fk(users.id)"},
			},
			Relations: []service.FDLRelation{
				{BelongsTo: "User", ForeignKey: "author_id"},
			},
		},
	}

	// Test valid model
	errors := service.ValidateModelAdvanced(&allModels[1], allModels)
	if len(errors) > 0 {
		t.Errorf("Expected no errors for valid model, got: %v", errors)
	}

	// Test invalid FK reference
	invalidModel := service.FDLModel{
		Name:  "Comment",
		Table: "comments",
		Fields: []service.FDLField{
			{Name: "id", Type: "uuid", Constraints: "pk"},
			{Name: "post_id", Type: "uuid", Constraints: "fk(nonexistent.id)"},
		},
	}
	errors = service.ValidateModelAdvanced(&invalidModel, allModels)
	if len(errors) == 0 {
		t.Error("Expected error for invalid FK reference")
	}

	// Test invalid relation target
	invalidRelModel := service.FDLModel{
		Name:  "Like",
		Table: "likes",
		Fields: []service.FDLField{
			{Name: "id", Type: "uuid", Constraints: "pk"},
		},
		Relations: []service.FDLRelation{
			{BelongsTo: "NonExistentModel"},
		},
	}
	errors = service.ValidateModelAdvanced(&invalidRelModel, allModels)
	if len(errors) == 0 {
		t.Error("Expected error for invalid relation target")
	}
}

func TestValidFieldTypes(t *testing.T) {
	expectedTypes := []string{
		"uuid", "string", "text", "int", "bigint", "float", "decimal", "boolean",
		"datetime", "date", "time", "json", "blob", "enum",
	}

	for _, typ := range expectedTypes {
		found := false
		for _, valid := range service.ValidFieldTypes {
			if valid == typ {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Type %s should be in ValidFieldTypes", typ)
		}
	}
}

func TestParseAndValidateService(t *testing.T) {
	svc := &service.FDLService{
		Name: "createUser",
		Desc: "Create a new user",
		Input: map[string]interface{}{
			"email": map[string]interface{}{
				"type":      "string",
				"required":  true,
				"minLength": 5,
				"maxLength": 255,
				"pattern":   "^[a-zA-Z0-9._%+-]+@",
			},
			"password": map[string]interface{}{
				"type":      "string",
				"required":  true,
				"minLength": 8,
			},
			"name": "string", // Simple format
		},
		Output:      "User",
		Throws:      []string{"USER_EXISTS", "INVALID_EMAIL"},
		Transaction: true,
		Auth:        "required",
		Roles:       []string{"admin"},
		Steps:       []interface{}{"validate: email format", "db: insert User", "return: user"},
	}

	err := service.ParseAndValidateService(svc)
	if err != nil {
		t.Fatalf("ParseAndValidateService failed: %v", err)
	}

	// Check parsed input
	if len(svc.ParsedInput) != 3 {
		t.Errorf("Expected 3 parsed inputs, got %d", len(svc.ParsedInput))
	}

	emailParam := svc.ParsedInput["email"]
	if emailParam.Type != "string" {
		t.Errorf("Expected email type 'string', got '%s'", emailParam.Type)
	}
	if !emailParam.Required {
		t.Error("Expected email to be required")
	}
	if emailParam.MinLength != 5 {
		t.Errorf("Expected email minLength 5, got %d", emailParam.MinLength)
	}

	nameParam := svc.ParsedInput["name"]
	if nameParam.Type != "string" {
		t.Errorf("Expected name type 'string', got '%s'", nameParam.Type)
	}

	// Check parsed output
	if svc.ParsedOutput.Type != "User" {
		t.Errorf("Expected output type 'User', got '%s'", svc.ParsedOutput.Type)
	}
	if svc.ParsedOutput.IsArray {
		t.Error("Expected output to not be array")
	}

	// Check parsed steps
	if len(svc.ParsedSteps) != 3 {
		t.Errorf("Expected 3 parsed steps, got %d", len(svc.ParsedSteps))
	}
	if svc.ParsedSteps[0].Type != "validate" {
		t.Errorf("Expected first step type 'validate', got '%s'", svc.ParsedSteps[0].Type)
	}
	if svc.ParsedSteps[1].Type != "db" {
		t.Errorf("Expected second step type 'db', got '%s'", svc.ParsedSteps[1].Type)
	}
	if svc.ParsedSteps[2].Type != "return" {
		t.Errorf("Expected third step type 'return', got '%s'", svc.ParsedSteps[2].Type)
	}
}

func TestParseOutputArray(t *testing.T) {
	svc := &service.FDLService{
		Name:   "listUsers",
		Desc:   "List all users",
		Output: "Array<User>",
		Steps:  []interface{}{"db: select all User"},
	}

	err := service.ParseAndValidateService(svc)
	if err != nil {
		t.Fatalf("ParseAndValidateService failed: %v", err)
	}

	if !svc.ParsedOutput.IsArray {
		t.Error("Expected output to be array")
	}
	if svc.ParsedOutput.Type != "User" {
		t.Errorf("Expected output type 'User', got '%s'", svc.ParsedOutput.Type)
	}
}

func TestParseOutputComplex(t *testing.T) {
	svc := &service.FDLService{
		Name:   "login",
		Desc:   "User login",
		Output: "{ user: User, token: string }",
		Steps:  []interface{}{"return: auth result"},
	}

	err := service.ParseAndValidateService(svc)
	if err != nil {
		t.Fatalf("ParseAndValidateService failed: %v", err)
	}

	if !svc.ParsedOutput.IsComplex {
		t.Error("Expected output to be complex")
	}
	if len(svc.ParsedOutput.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(svc.ParsedOutput.Fields))
	}
	if svc.ParsedOutput.Fields["user"] != "User" {
		t.Errorf("Expected field user=User, got %s", svc.ParsedOutput.Fields["user"])
	}
	if svc.ParsedOutput.Fields["token"] != "string" {
		t.Errorf("Expected field token=string, got %s", svc.ParsedOutput.Fields["token"])
	}
}

func TestParseOutputVoid(t *testing.T) {
	svc := &service.FDLService{
		Name:   "deleteUser",
		Desc:   "Delete a user",
		Output: "void",
		Steps:  []interface{}{"db: delete User"},
	}

	err := service.ParseAndValidateService(svc)
	if err != nil {
		t.Fatalf("ParseAndValidateService failed: %v", err)
	}

	if svc.ParsedOutput.Type != "void" {
		t.Errorf("Expected output type 'void', got '%s'", svc.ParsedOutput.Type)
	}
}

func TestParseServiceInvalidAuth(t *testing.T) {
	svc := &service.FDLService{
		Name:  "test",
		Auth:  "invalid_auth",
		Steps: []interface{}{"return: test"},
	}

	err := service.ParseAndValidateService(svc)
	if err == nil {
		t.Error("Expected error for invalid auth value")
	}
}

func TestParseStepsMap(t *testing.T) {
	svc := &service.FDLService{
		Name: "test",
		Steps: []interface{}{
			map[string]interface{}{
				"validate": "email format",
				"message":  "Invalid email",
			},
			map[string]interface{}{
				"db":    "select User",
				"table": "users",
			},
		},
	}

	err := service.ParseAndValidateService(svc)
	if err != nil {
		t.Fatalf("ParseAndValidateService failed: %v", err)
	}

	if len(svc.ParsedSteps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(svc.ParsedSteps))
	}

	if svc.ParsedSteps[0].Type != "validate" {
		t.Errorf("Expected first step type 'validate', got '%s'", svc.ParsedSteps[0].Type)
	}
	if svc.ParsedSteps[0].Params["message"] != "Invalid email" {
		t.Error("Expected message param to be preserved")
	}
}

func TestParseAndValidateAPI(t *testing.T) {
	api := &service.FDLAPI{
		Path:    "/api/users/{id}",
		Method:  "GET",
		Summary: "Get user by ID",
		Use:     "service.getUser",
		Request: map[string]interface{}{
			"params": map[string]interface{}{
				"id": "int (required)",
			},
			"query": map[string]interface{}{
				"include": map[string]interface{}{
					"type":    "string",
					"default": "basic",
				},
			},
			"headers": map[string]interface{}{
				"Authorization": "string (required)",
			},
		},
		Response: map[string]interface{}{
			"200": map[string]interface{}{"id": "int", "name": "string"},
			"404": map[string]interface{}{"error": "Not Found"},
		},
		Auth:  "required",
		Roles: []string{"user", "admin"},
		Tags:  []string{"users"},
		RateLimit: map[string]interface{}{
			"limit":  100,
			"window": 60,
			"by":     "ip",
		},
	}

	services := []service.FDLService{
		{Name: "getUser"},
	}

	err := service.ParseAndValidateAPI(api, services)
	if err != nil {
		t.Fatalf("ParseAndValidateAPI failed: %v", err)
	}

	// Check parsed request params
	if len(api.ParsedRequest.Params) != 1 {
		t.Errorf("Expected 1 param, got %d", len(api.ParsedRequest.Params))
	}
	idParam := api.ParsedRequest.Params["id"]
	if idParam.Type != "int" {
		t.Errorf("Expected param type 'int', got '%s'", idParam.Type)
	}
	if !idParam.Required {
		t.Error("Expected param to be required")
	}

	// Check parsed query
	if len(api.ParsedRequest.Query) != 1 {
		t.Errorf("Expected 1 query param, got %d", len(api.ParsedRequest.Query))
	}
	includeParam := api.ParsedRequest.Query["include"]
	if includeParam.Default != "basic" {
		t.Errorf("Expected default 'basic', got '%v'", includeParam.Default)
	}

	// Check parsed response
	if len(api.ParsedResponse) != 2 {
		t.Errorf("Expected 2 response codes, got %d", len(api.ParsedResponse))
	}
	if _, ok := api.ParsedResponse[200]; !ok {
		t.Error("Expected 200 response")
	}
	if _, ok := api.ParsedResponse[404]; !ok {
		t.Error("Expected 404 response")
	}

	// Check parsed rate limit
	if api.ParsedRateLimit == nil {
		t.Fatal("Expected rate limit to be parsed")
	}
	if api.ParsedRateLimit.Limit != 100 {
		t.Errorf("Expected rate limit 100, got %d", api.ParsedRateLimit.Limit)
	}
	if api.ParsedRateLimit.Window != 60 {
		t.Errorf("Expected window 60, got %d", api.ParsedRateLimit.Window)
	}
	if api.ParsedRateLimit.By != "ip" {
		t.Errorf("Expected by 'ip', got '%s'", api.ParsedRateLimit.By)
	}
}

func TestParseAndValidateAPIInvalidMethod(t *testing.T) {
	api := &service.FDLAPI{
		Path:   "/api/test",
		Method: "INVALID",
	}

	err := service.ParseAndValidateAPI(api, nil)
	if err == nil {
		t.Error("Expected error for invalid method")
	}
}

func TestParseAndValidateAPIInvalidPath(t *testing.T) {
	api := &service.FDLAPI{
		Path:   "api/test", // Missing leading /
		Method: "GET",
	}

	err := service.ParseAndValidateAPI(api, nil)
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestParseAndValidateAPIInvalidAuth(t *testing.T) {
	api := &service.FDLAPI{
		Path:   "/api/test",
		Method: "GET",
		Auth:   "invalid_auth",
	}

	err := service.ParseAndValidateAPI(api, nil)
	if err == nil {
		t.Error("Expected error for invalid auth")
	}
}

func TestParseTransform(t *testing.T) {
	api := &service.FDLAPI{
		Path:   "/api/test",
		Method: "GET",
		Transform: map[string]interface{}{
			"exclude": []interface{}{"password", "secret"},
			"rename": map[string]interface{}{
				"created_at": "createdAt",
				"updated_at": "updatedAt",
			},
		},
	}

	err := service.ParseAndValidateAPI(api, nil)
	if err != nil {
		t.Fatalf("ParseAndValidateAPI failed: %v", err)
	}

	if api.ParsedTransform == nil {
		t.Fatal("Expected transform to be parsed")
	}
	if len(api.ParsedTransform.Exclude) != 2 {
		t.Errorf("Expected 2 excluded fields, got %d", len(api.ParsedTransform.Exclude))
	}
	if len(api.ParsedTransform.Rename) != 2 {
		t.Errorf("Expected 2 renamed fields, got %d", len(api.ParsedTransform.Rename))
	}
	if api.ParsedTransform.Rename["created_at"] != "createdAt" {
		t.Error("Expected rename created_at to createdAt")
	}
}

// ============================================================================
// UI Parsing Tests (Presentation Layer)
// ============================================================================

func TestParseAndValidateUI(t *testing.T) {
	ui := &service.FDLUI{
		Component:   "UserProfile",
		Type:        "Organism",
		Description: "User profile component",
		Props: map[string]interface{}{
			"userId": map[string]interface{}{
				"type":     "string",
				"required": true,
			},
			"onClose": map[string]interface{}{
				"type":     "function",
				"optional": true,
			},
			"title": "string", // Simple format
		},
		State: []interface{}{
			"user: User | null",
			"isLoading: boolean = false",
			"error: string = \"\"",
		},
		Computed: []interface{}{
			"fullName: user.firstName + \" \" + user.lastName",
			"canEdit: user.role === \"admin\"",
		},
		Init: []interface{}{
			"call: api.getUser($userId)",
		},
		Methods: map[string]interface{}{
			"handleSubmit": []interface{}{
				"validate: form",
				"call: api.updateUser($userId, $form)",
				"navigate: /users/$userId",
			},
		},
		View: []map[string]interface{}{
			{"Text": "Hello World"},
		},
	}

	allUIs := []service.FDLUI{}

	errors := service.ParseAndValidateUI(ui, allUIs)
	if len(errors) > 0 {
		t.Errorf("ParseAndValidateUI returned errors: %v", errors)
	}

	// Check parsed props
	if len(ui.ParsedProps) != 3 {
		t.Errorf("Expected 3 props, got %d", len(ui.ParsedProps))
	}
	userIdProp := ui.ParsedProps["userId"]
	if userIdProp.Type != "string" {
		t.Errorf("Expected userId type 'string', got '%s'", userIdProp.Type)
	}
	if !userIdProp.Required {
		t.Error("Expected userId to be required")
	}
	titleProp := ui.ParsedProps["title"]
	if titleProp.Type != "string" {
		t.Errorf("Expected title type 'string', got '%s'", titleProp.Type)
	}

	// Check parsed state
	if len(ui.ParsedState) != 3 {
		t.Errorf("Expected 3 state items, got %d", len(ui.ParsedState))
	}
	userState := ui.ParsedState[0]
	if userState.Name != "user" {
		t.Errorf("Expected state name 'user', got '%s'", userState.Name)
	}
	if userState.Type != "User | null" {
		t.Errorf("Expected state type 'User | null', got '%s'", userState.Type)
	}
	isLoadingState := ui.ParsedState[1]
	if isLoadingState.Name != "isLoading" {
		t.Errorf("Expected state name 'isLoading', got '%s'", isLoadingState.Name)
	}
	if isLoadingState.Default != false {
		t.Errorf("Expected isLoading default false, got '%v'", isLoadingState.Default)
	}

	// Check parsed computed
	if len(ui.ParsedComputed) != 2 {
		t.Errorf("Expected 2 computed items, got %d", len(ui.ParsedComputed))
	}
	fullNameComputed := ui.ParsedComputed[0]
	if fullNameComputed.Name != "fullName" {
		t.Errorf("Expected computed name 'fullName', got '%s'", fullNameComputed.Name)
	}

	// Check parsed init
	if len(ui.ParsedInit) != 1 {
		t.Errorf("Expected 1 init action, got %d", len(ui.ParsedInit))
	}
	if ui.ParsedInit[0].Type != "call" {
		t.Errorf("Expected init action type 'call', got '%s'", ui.ParsedInit[0].Type)
	}

	// Check parsed methods
	if len(ui.ParsedMethods) != 1 {
		t.Errorf("Expected 1 method, got %d", len(ui.ParsedMethods))
	}
	handleSubmit := ui.ParsedMethods["handleSubmit"]
	if len(handleSubmit) != 3 {
		t.Errorf("Expected 3 actions in handleSubmit, got %d", len(handleSubmit))
	}

	// Check parsed view
	if len(ui.ParsedView) != 1 {
		t.Errorf("Expected 1 view element, got %d", len(ui.ParsedView))
	}
	if ui.ParsedView[0].Type != "Text" {
		t.Errorf("Expected view element type 'Text', got '%s'", ui.ParsedView[0].Type)
	}
}

func TestParseAndValidateUIInvalidType(t *testing.T) {
	ui := &service.FDLUI{
		Component: "TestComponent",
		Type:      "InvalidType",
	}

	errors := service.ParseAndValidateUI(ui, []service.FDLUI{})
	if len(errors) == 0 {
		t.Error("Expected error for invalid UI type")
	}

	foundTypeError := false
	for _, err := range errors {
		if strings.Contains(err.Error(), "invalid type") {
			foundTypeError = true
			break
		}
	}
	if !foundTypeError {
		t.Error("Expected 'invalid type' error")
	}
}

func TestParseAndValidateUIInvalidParent(t *testing.T) {
	ui := &service.FDLUI{
		Component: "ChildComponent",
		Type:      "Molecule",
		Parent:    "NonExistentParent",
	}

	errors := service.ParseAndValidateUI(ui, []service.FDLUI{})
	if len(errors) == 0 {
		t.Error("Expected error for invalid parent")
	}

	foundParentError := false
	for _, err := range errors {
		if strings.Contains(err.Error(), "parent component not found") {
			foundParentError = true
			break
		}
	}
	if !foundParentError {
		t.Error("Expected 'parent component not found' error")
	}
}

func TestParseAndValidateUIWithParent(t *testing.T) {
	parentUI := service.FDLUI{
		Component: "ParentComponent",
		Type:      "Organism",
	}

	childUI := &service.FDLUI{
		Component: "ChildComponent",
		Type:      "Molecule",
		Parent:    "ParentComponent",
	}

	errors := service.ParseAndValidateUI(childUI, []service.FDLUI{parentUI})
	if len(errors) > 0 {
		t.Errorf("Expected no errors with valid parent, got: %v", errors)
	}
}

func TestParseUIPropsEmptyTypeError(t *testing.T) {
	ui := &service.FDLUI{
		Component: "TestComponent",
		Type:      "Atom",
		Props: map[string]interface{}{
			"invalidProp": map[string]interface{}{
				"required": true,
				// Missing type
			},
		},
	}

	errors := service.ParseAndValidateUI(ui, []service.FDLUI{})
	if len(errors) == 0 {
		t.Error("Expected error for prop without type")
	}

	foundPropError := false
	for _, err := range errors {
		if strings.Contains(err.Error(), "has no type") {
			foundPropError = true
			break
		}
	}
	if !foundPropError {
		t.Error("Expected 'has no type' error for prop")
	}
}

func TestParseUIStateFormats(t *testing.T) {
	ui := &service.FDLUI{
		Component: "TestComponent",
		Type:      "Organism",
		State: []interface{}{
			"count: number = 0",
			"name: string = \"default\"",
			"active: boolean = true",
			"data: object",
			"items",
		},
	}

	errors := service.ParseAndValidateUI(ui, []service.FDLUI{})
	if len(errors) > 0 {
		t.Errorf("ParseAndValidateUI returned errors: %v", errors)
	}

	if len(ui.ParsedState) != 5 {
		t.Errorf("Expected 5 state items, got %d", len(ui.ParsedState))
	}

	// Check count state with default 0
	countState := ui.ParsedState[0]
	if countState.Name != "count" {
		t.Errorf("Expected name 'count', got '%s'", countState.Name)
	}
	if countState.Type != "number" {
		t.Errorf("Expected type 'number', got '%s'", countState.Type)
	}
	if countState.Default != 0 {
		t.Errorf("Expected default 0, got '%v'", countState.Default)
	}

	// Check active state with default true
	activeState := ui.ParsedState[2]
	if activeState.Name != "active" {
		t.Errorf("Expected name 'active', got '%s'", activeState.Name)
	}
	if activeState.Default != true {
		t.Errorf("Expected default true, got '%v'", activeState.Default)
	}

	// Check simple name-only state
	itemsState := ui.ParsedState[4]
	if itemsState.Name != "items" {
		t.Errorf("Expected name 'items', got '%s'", itemsState.Name)
	}
}

func TestParseUIComputedFormats(t *testing.T) {
	ui := &service.FDLUI{
		Component: "TestComponent",
		Type:      "Organism",
		Computed: []interface{}{
			"fullName: firstName + \" \" + lastName",
			map[string]interface{}{
				"isValid": "count > 0 && name !== \"\"",
			},
		},
	}

	errors := service.ParseAndValidateUI(ui, []service.FDLUI{})
	if len(errors) > 0 {
		t.Errorf("ParseAndValidateUI returned errors: %v", errors)
	}

	if len(ui.ParsedComputed) != 2 {
		t.Errorf("Expected 2 computed items, got %d", len(ui.ParsedComputed))
	}

	// Check string format
	fullName := ui.ParsedComputed[0]
	if fullName.Name != "fullName" {
		t.Errorf("Expected name 'fullName', got '%s'", fullName.Name)
	}
	if !strings.Contains(fullName.Expression, "firstName") {
		t.Error("Expected expression to contain 'firstName'")
	}

	// Check map format
	isValid := ui.ParsedComputed[1]
	if isValid.Name != "isValid" {
		t.Errorf("Expected name 'isValid', got '%s'", isValid.Name)
	}
}

func TestParseUIActionsStringFormat(t *testing.T) {
	ui := &service.FDLUI{
		Component: "TestComponent",
		Type:      "Organism",
		Init: []interface{}{
			"call: api.getData()",
			"set: loading = true",
			"navigate: /home",
			"show: toast \"Success\"",
			"validate: form",
			"emit: dataLoaded",
		},
	}

	errors := service.ParseAndValidateUI(ui, []service.FDLUI{})
	if len(errors) > 0 {
		t.Errorf("ParseAndValidateUI returned errors: %v", errors)
	}

	if len(ui.ParsedInit) != 6 {
		t.Errorf("Expected 6 init actions, got %d", len(ui.ParsedInit))
	}

	expectedTypes := []string{"call", "set", "navigate", "show", "validate", "emit"}
	for i, action := range ui.ParsedInit {
		if action.Type != expectedTypes[i] {
			t.Errorf("Expected action %d type '%s', got '%s'", i, expectedTypes[i], action.Type)
		}
	}
}

func TestParseUIActionsMapFormat(t *testing.T) {
	ui := &service.FDLUI{
		Component: "TestComponent",
		Type:      "Organism",
		Init: []interface{}{
			map[string]interface{}{
				"call": "api.getData()",
				"set":  "data",
				"onSuccess": []interface{}{
					"show: toast \"Loaded\"",
				},
				"onError": []interface{}{
					"set: error = $error.message",
				},
			},
		},
	}

	errors := service.ParseAndValidateUI(ui, []service.FDLUI{})
	if len(errors) > 0 {
		t.Errorf("ParseAndValidateUI returned errors: %v", errors)
	}

	if len(ui.ParsedInit) != 1 {
		t.Errorf("Expected 1 init action, got %d", len(ui.ParsedInit))
	}

	action := ui.ParsedInit[0]
	if action.Type != "call" {
		t.Errorf("Expected action type 'call', got '%s'", action.Type)
	}
	if len(action.OnSuccess) != 1 {
		t.Errorf("Expected 1 onSuccess action, got %d", len(action.OnSuccess))
	}
	if len(action.OnError) != 1 {
		t.Errorf("Expected 1 onError action, got %d", len(action.OnError))
	}
}

func TestParseUIViewSimple(t *testing.T) {
	ui := &service.FDLUI{
		Component: "TestComponent",
		Type:      "Atom",
		View: []map[string]interface{}{
			{"Text": "Hello"},
			{"Button": map[string]interface{}{
				"label":   "Click me",
				"onClick": "handleClick",
			}},
		},
	}

	errors := service.ParseAndValidateUI(ui, []service.FDLUI{})
	if len(errors) > 0 {
		t.Errorf("ParseAndValidateUI returned errors: %v", errors)
	}

	if len(ui.ParsedView) != 2 {
		t.Errorf("Expected 2 view elements, got %d", len(ui.ParsedView))
	}

	// Check Text element
	textEl := ui.ParsedView[0]
	if textEl.Type != "Text" {
		t.Errorf("Expected type 'Text', got '%s'", textEl.Type)
	}
	if textEl.Props["text"] != "Hello" {
		t.Errorf("Expected text prop 'Hello', got '%v'", textEl.Props["text"])
	}

	// Check Button element
	buttonEl := ui.ParsedView[1]
	if buttonEl.Type != "Button" {
		t.Errorf("Expected type 'Button', got '%s'", buttonEl.Type)
	}
	if buttonEl.Props["label"] != "Click me" {
		t.Errorf("Expected label 'Click me', got '%v'", buttonEl.Props["label"])
	}
}

func TestParseUIViewNested(t *testing.T) {
	ui := &service.FDLUI{
		Component: "TestComponent",
		Type:      "Organism",
		View: []map[string]interface{}{
			{"Flex": map[string]interface{}{
				"direction": "column",
				"children": []interface{}{
					map[string]interface{}{"Text": "Header"},
					map[string]interface{}{"Input": map[string]interface{}{
						"placeholder": "Enter name",
					}},
				},
			}},
		},
	}

	errors := service.ParseAndValidateUI(ui, []service.FDLUI{})
	if len(errors) > 0 {
		t.Errorf("ParseAndValidateUI returned errors: %v", errors)
	}

	if len(ui.ParsedView) != 1 {
		t.Errorf("Expected 1 view element, got %d", len(ui.ParsedView))
	}

	flexEl := ui.ParsedView[0]
	if flexEl.Type != "Flex" {
		t.Errorf("Expected type 'Flex', got '%s'", flexEl.Type)
	}
	if flexEl.Props["direction"] != "column" {
		t.Errorf("Expected direction 'column', got '%v'", flexEl.Props["direction"])
	}

	if len(flexEl.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(flexEl.Children))
	}

	textChild := flexEl.Children[0]
	if textChild.Type != "Text" {
		t.Errorf("Expected child type 'Text', got '%s'", textChild.Type)
	}

	inputChild := flexEl.Children[1]
	if inputChild.Type != "Input" {
		t.Errorf("Expected child type 'Input', got '%s'", inputChild.Type)
	}
}

func TestParseUIViewConditional(t *testing.T) {
	ui := &service.FDLUI{
		Component: "TestComponent",
		Type:      "Organism",
		View: []map[string]interface{}{
			{
				"if":   "isLoading",
				"then": []interface{}{map[string]interface{}{"Text": "Loading..."}},
				"else": []interface{}{map[string]interface{}{"Text": "Content"}},
			},
		},
	}

	errors := service.ParseAndValidateUI(ui, []service.FDLUI{})
	if len(errors) > 0 {
		t.Errorf("ParseAndValidateUI returned errors: %v", errors)
	}

	if len(ui.ParsedView) != 1 {
		t.Errorf("Expected 1 view element, got %d", len(ui.ParsedView))
	}

	condEl := ui.ParsedView[0]
	if condEl.Condition == nil {
		t.Fatal("Expected condition to be set")
	}
	if condEl.Condition.If != "isLoading" {
		t.Errorf("Expected condition 'isLoading', got '%s'", condEl.Condition.If)
	}
	if len(condEl.Condition.Then) != 1 {
		t.Errorf("Expected 1 then element, got %d", len(condEl.Condition.Then))
	}
	if len(condEl.Condition.Else) != 1 {
		t.Errorf("Expected 1 else element, got %d", len(condEl.Condition.Else))
	}
}

func TestParseUIMethodsMultiple(t *testing.T) {
	ui := &service.FDLUI{
		Component: "TestComponent",
		Type:      "Organism",
		Methods: map[string]interface{}{
			"handleSubmit": []interface{}{
				"validate: form",
				"call: api.submit($form)",
			},
			"handleCancel": []interface{}{
				"navigate: /home",
			},
			"handleReset": []interface{}{
				"set: form = {}",
				"set: error = null",
			},
		},
	}

	errors := service.ParseAndValidateUI(ui, []service.FDLUI{})
	if len(errors) > 0 {
		t.Errorf("ParseAndValidateUI returned errors: %v", errors)
	}

	if len(ui.ParsedMethods) != 3 {
		t.Errorf("Expected 3 methods, got %d", len(ui.ParsedMethods))
	}

	submitMethod := ui.ParsedMethods["handleSubmit"]
	if len(submitMethod) != 2 {
		t.Errorf("Expected 2 actions in handleSubmit, got %d", len(submitMethod))
	}

	cancelMethod := ui.ParsedMethods["handleCancel"]
	if len(cancelMethod) != 1 {
		t.Errorf("Expected 1 action in handleCancel, got %d", len(cancelMethod))
	}

	resetMethod := ui.ParsedMethods["handleReset"]
	if len(resetMethod) != 2 {
		t.Errorf("Expected 2 actions in handleReset, got %d", len(resetMethod))
	}
}

func TestParseUIValidTypes(t *testing.T) {
	validTypes := []string{"Page", "Template", "Organism", "Molecule", "Atom"}

	for _, typ := range validTypes {
		ui := &service.FDLUI{
			Component: "TestComponent",
			Type:      typ,
		}

		errors := service.ParseAndValidateUI(ui, []service.FDLUI{})
		if len(errors) > 0 {
			t.Errorf("Type '%s' should be valid, got errors: %v", typ, errors)
		}
	}
}

func TestParseUIEmptyType(t *testing.T) {
	// Empty type should be allowed (defaults to no type validation)
	ui := &service.FDLUI{
		Component: "TestComponent",
		Type:      "",
	}

	errors := service.ParseAndValidateUI(ui, []service.FDLUI{})
	if len(errors) > 0 {
		t.Errorf("Empty type should be valid, got errors: %v", errors)
	}
}
