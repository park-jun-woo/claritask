package service

import (
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/model"
)

// FDLSpec - FDL 전체 구조
type FDLSpec struct {
	Feature     string       `yaml:"feature"`
	Description string       `yaml:"description"`
	Models      []FDLModel   `yaml:"models,omitempty"`
	Service     []FDLService `yaml:"service,omitempty"`
	API         []FDLAPI     `yaml:"api,omitempty"`
	UI          []FDLUI      `yaml:"ui,omitempty"`
}

// FDLModel - 데이터 모델 정의
type FDLModel struct {
	Name   string     `yaml:"name"`
	Table  string     `yaml:"table"`
	Fields []FDLField `yaml:"fields"`
}

// FDLField - 필드 정의
type FDLField struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Constraints string `yaml:"constraints,omitempty"`
}

// FDLService - 서비스 함수 정의
type FDLService struct {
	Name   string                 `yaml:"name"`
	Desc   string                 `yaml:"desc"`
	Input  map[string]interface{} `yaml:"input"`
	Output string                 `yaml:"output,omitempty"`
	Steps  []string               `yaml:"steps"`
}

// FDLAPI - API 엔드포인트 정의
type FDLAPI struct {
	Path     string                 `yaml:"path"`
	Method   string                 `yaml:"method"`
	Summary  string                 `yaml:"summary,omitempty"`
	Use      string                 `yaml:"use"` // service.FunctionName
	Request  map[string]interface{} `yaml:"request,omitempty"`
	Response map[string]interface{} `yaml:"response"`
}

// FDLUI - UI 컴포넌트 정의
type FDLUI struct {
	Component string                   `yaml:"component"`
	Type      string                   `yaml:"type"` // Page, Organism, Molecule
	Props     map[string]interface{}   `yaml:"props,omitempty"`
	State     []string                 `yaml:"state,omitempty"`
	Init      []string                 `yaml:"init,omitempty"`
	View      []map[string]interface{} `yaml:"view,omitempty"`
	Parent    string                   `yaml:"parent,omitempty"`
}

// ParseFDL parses YAML string to FDLSpec
func ParseFDL(yamlStr string) (*FDLSpec, error) {
	var spec FDLSpec
	err := yaml.Unmarshal([]byte(yamlStr), &spec)
	if err != nil {
		return nil, fmt.Errorf("parse FDL: %w", err)
	}
	return &spec, nil
}

// ParseFDLFile parses FDL from file
func ParseFDLFile(filePath string) (*FDLSpec, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read FDL file: %w", err)
	}
	return ParseFDL(string(content))
}

// ValidateFDL validates FDL specification
func ValidateFDL(spec *FDLSpec) error {
	if err := validateFeatureName(spec.Feature); err != nil {
		return err
	}

	if err := validateModels(spec.Models); err != nil {
		return err
	}

	if err := validateServices(spec.Service); err != nil {
		return err
	}

	if err := validateAPIs(spec.API, spec.Service); err != nil {
		return err
	}

	if err := validateUIs(spec.UI); err != nil {
		return err
	}

	return nil
}

// validateFeatureName validates feature name
func validateFeatureName(name string) error {
	if name == "" {
		return fmt.Errorf("feature name is required")
	}
	// Feature name should be alphanumeric with underscores or hyphens
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_-]*$`, name)
	if !matched {
		return fmt.Errorf("invalid feature name: %s (must start with letter, contain only letters, numbers, underscores, hyphens)", name)
	}
	return nil
}

// validateModels validates models section
func validateModels(models []FDLModel) error {
	names := make(map[string]bool)
	for _, m := range models {
		if m.Name == "" {
			return fmt.Errorf("model name is required")
		}
		if names[m.Name] {
			return fmt.Errorf("duplicate model name: %s", m.Name)
		}
		names[m.Name] = true

		if m.Table == "" {
			return fmt.Errorf("model %s: table name is required", m.Name)
		}

		if len(m.Fields) == 0 {
			return fmt.Errorf("model %s: at least one field is required", m.Name)
		}

		fieldNames := make(map[string]bool)
		for _, f := range m.Fields {
			if f.Name == "" {
				return fmt.Errorf("model %s: field name is required", m.Name)
			}
			if fieldNames[f.Name] {
				return fmt.Errorf("model %s: duplicate field name: %s", m.Name, f.Name)
			}
			fieldNames[f.Name] = true

			if f.Type == "" {
				return fmt.Errorf("model %s: field %s type is required", m.Name, f.Name)
			}
		}
	}
	return nil
}

// validateServices validates services section
func validateServices(services []FDLService) error {
	names := make(map[string]bool)
	for _, s := range services {
		if s.Name == "" {
			return fmt.Errorf("service name is required")
		}
		if names[s.Name] {
			return fmt.Errorf("duplicate service name: %s", s.Name)
		}
		names[s.Name] = true

		if len(s.Steps) == 0 {
			return fmt.Errorf("service %s: at least one step is required", s.Name)
		}
	}
	return nil
}

// validateAPIs validates API section and service.use connections
func validateAPIs(apis []FDLAPI, services []FDLService) error {
	// Build service name set
	serviceNames := make(map[string]bool)
	for _, s := range services {
		serviceNames[s.Name] = true
	}

	for _, a := range apis {
		if a.Path == "" {
			return fmt.Errorf("API path is required")
		}
		if a.Method == "" {
			return fmt.Errorf("API %s: method is required", a.Path)
		}

		// Validate method
		method := strings.ToUpper(a.Method)
		validMethods := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
		}
		if !validMethods[method] {
			return fmt.Errorf("API %s: invalid method: %s", a.Path, a.Method)
		}

		// Validate use reference
		if a.Use != "" {
			// Expected format: service.FunctionName
			if !strings.HasPrefix(a.Use, "service.") {
				return fmt.Errorf("API %s: use must reference service (service.FunctionName)", a.Path)
			}
			funcName := strings.TrimPrefix(a.Use, "service.")
			if !serviceNames[funcName] {
				return fmt.Errorf("API %s: referenced service '%s' not found", a.Path, funcName)
			}
		}
	}
	return nil
}

// validateUIs validates UI section
func validateUIs(uis []FDLUI) error {
	names := make(map[string]bool)
	validTypes := map[string]bool{
		"Page": true, "Organism": true, "Molecule": true, "Atom": true,
	}

	for _, u := range uis {
		if u.Component == "" {
			return fmt.Errorf("UI component name is required")
		}
		if names[u.Component] {
			return fmt.Errorf("duplicate UI component name: %s", u.Component)
		}
		names[u.Component] = true

		if u.Type != "" && !validTypes[u.Type] {
			return fmt.Errorf("UI %s: invalid type: %s (must be Page, Organism, Molecule, or Atom)", u.Component, u.Type)
		}
	}
	return nil
}

// CalculateFDLHashFromSpec calculates SHA256 hash of FDL spec
func CalculateFDLHashFromSpec(yamlStr string) string {
	hash := sha256.Sum256([]byte(yamlStr))
	return fmt.Sprintf("%x", hash)
}

// FDL 템플릿
const fdlTemplate = `feature: %s
description: "TODO: Feature description"

models:
  - name: TODO
    table: todos
    fields:
      - name: id
        type: uuid
        constraints: pk
      - name: created_at
        type: datetime
        constraints: "default: now"

service:
  - name: createTODO
    desc: "TODO: Service description"
    input: {}
    steps:
      - "TODO: Step 1"

api:
  - path: /api/todos
    method: POST
    use: service.createTODO
    response:
      201: { id: uuid }

ui:
  - component: TODOComponent
    type: Organism
    state:
      - items
`

// GenerateFDLTemplate generates an empty FDL template
func GenerateFDLTemplate(featureName string) string {
	return fmt.Sprintf(fdlTemplate, featureName)
}

// FDLTaskMapping - FDL에서 Task 생성 정보 추출
type FDLTaskMapping struct {
	Title          string   // Task 제목
	Content        string   // Task 내용
	TargetFile     string   // 대상 파일 경로
	TargetFunction string   // 대상 함수명
	Layer          string   // model, service, api, ui
	Dependencies   []string // 의존 Task 힌트
}

// ExtractTaskMappings extracts task mappings from FDL
func ExtractTaskMappings(spec *FDLSpec, tech map[string]interface{}) ([]FDLTaskMapping, error) {
	var mappings []FDLTaskMapping

	// Determine file paths based on tech stack
	backend := "python"
	if b, ok := tech["backend"].(string); ok {
		backend = strings.ToLower(b)
	}

	// Model tasks
	for _, m := range spec.Models {
		mapping := FDLTaskMapping{
			Title:   fmt.Sprintf("Create %s model", m.Name),
			Content: fmt.Sprintf("Create model %s with table %s and fields", m.Name, m.Table),
			Layer:   "model",
		}

		// Set target file based on backend
		switch backend {
		case "go", "golang":
			mapping.TargetFile = fmt.Sprintf("internal/model/%s.go", strings.ToLower(m.Name))
		case "python", "fastapi", "django":
			mapping.TargetFile = fmt.Sprintf("models/%s.py", strings.ToLower(m.Name))
		default:
			mapping.TargetFile = fmt.Sprintf("models/%s.py", strings.ToLower(m.Name))
		}

		mappings = append(mappings, mapping)
	}

	// Service tasks (depend on models)
	modelTasks := make([]string, len(spec.Models))
	for i, m := range spec.Models {
		modelTasks[i] = fmt.Sprintf("Create %s model", m.Name)
	}

	for _, s := range spec.Service {
		mapping := FDLTaskMapping{
			Title:          fmt.Sprintf("Implement %s service", s.Name),
			Content:        fmt.Sprintf("Implement service function %s: %s", s.Name, s.Desc),
			TargetFunction: s.Name,
			Layer:          "service",
			Dependencies:   modelTasks,
		}

		switch backend {
		case "go", "golang":
			mapping.TargetFile = fmt.Sprintf("internal/service/%s_service.go", strings.ToLower(spec.Feature))
		case "python", "fastapi", "django":
			mapping.TargetFile = fmt.Sprintf("services/%s_service.py", strings.ToLower(spec.Feature))
		default:
			mapping.TargetFile = fmt.Sprintf("services/%s_service.py", strings.ToLower(spec.Feature))
		}

		mappings = append(mappings, mapping)
	}

	// API tasks (depend on services)
	serviceTasks := make([]string, len(spec.Service))
	for i, s := range spec.Service {
		serviceTasks[i] = fmt.Sprintf("Implement %s service", s.Name)
	}

	for _, a := range spec.API {
		mapping := FDLTaskMapping{
			Title:   fmt.Sprintf("Create %s %s endpoint", a.Method, a.Path),
			Content: fmt.Sprintf("Create API endpoint %s %s using %s", a.Method, a.Path, a.Use),
			Layer:   "api",
		}

		// Only add dependency if the API uses a service
		if a.Use != "" {
			funcName := strings.TrimPrefix(a.Use, "service.")
			mapping.Dependencies = []string{fmt.Sprintf("Implement %s service", funcName)}
		}

		switch backend {
		case "go", "golang":
			mapping.TargetFile = fmt.Sprintf("internal/api/%s_handler.go", strings.ToLower(spec.Feature))
		case "python", "fastapi", "django":
			mapping.TargetFile = fmt.Sprintf("api/%s_router.py", strings.ToLower(spec.Feature))
		default:
			mapping.TargetFile = fmt.Sprintf("api/%s_router.py", strings.ToLower(spec.Feature))
		}

		mappings = append(mappings, mapping)
	}

	// UI tasks (depend on API)
	apiTasks := make([]string, len(spec.API))
	for i, a := range spec.API {
		apiTasks[i] = fmt.Sprintf("Create %s %s endpoint", a.Method, a.Path)
	}

	frontend := ""
	if f, ok := tech["frontend"].(string); ok {
		frontend = strings.ToLower(f)
	}

	for _, u := range spec.UI {
		mapping := FDLTaskMapping{
			Title:        fmt.Sprintf("Create %s component", u.Component),
			Content:      fmt.Sprintf("Create UI component %s of type %s", u.Component, u.Type),
			Layer:        "ui",
			Dependencies: apiTasks,
		}

		switch frontend {
		case "react":
			mapping.TargetFile = fmt.Sprintf("src/components/%s.tsx", u.Component)
		case "vue":
			mapping.TargetFile = fmt.Sprintf("src/components/%s.vue", u.Component)
		default:
			mapping.TargetFile = fmt.Sprintf("components/%s.tsx", u.Component)
		}

		mappings = append(mappings, mapping)
	}

	return mappings, nil
}

// GetFDLInfo converts FDLSpec to FDLInfo for task pop response
func GetFDLInfo(spec *FDLSpec) *model.FDLInfo {
	info := &model.FDLInfo{
		Feature: spec.Feature,
	}

	// Convert models
	if len(spec.Models) > 0 {
		info.Models = make(map[string]interface{})
		for _, m := range spec.Models {
			info.Models[m.Name] = map[string]interface{}{
				"table":  m.Table,
				"fields": m.Fields,
			}
		}
	}

	// Convert services
	if len(spec.Service) > 0 {
		info.Service = make(map[string]interface{})
		for _, s := range spec.Service {
			info.Service[s.Name] = map[string]interface{}{
				"desc":   s.Desc,
				"input":  s.Input,
				"output": s.Output,
				"steps":  s.Steps,
			}
		}
	}

	// Convert API
	if len(spec.API) > 0 {
		info.API = make(map[string]interface{})
		for _, a := range spec.API {
			key := fmt.Sprintf("%s %s", a.Method, a.Path)
			info.API[key] = map[string]interface{}{
				"use":      a.Use,
				"request":  a.Request,
				"response": a.Response,
			}
		}
	}

	// Convert UI
	if len(spec.UI) > 0 {
		info.UI = make(map[string]interface{})
		for _, u := range spec.UI {
			info.UI[u.Component] = map[string]interface{}{
				"type":  u.Type,
				"props": u.Props,
				"state": u.State,
			}
		}
	}

	return info
}

// GetFDLInfoFromDB retrieves FDL info for a feature
func GetFDLInfoFromDB(database *db.DB, featureID int64) (*model.FDLInfo, error) {
	fdl, err := GetFeatureFDL(database, featureID)
	if err != nil {
		return nil, err
	}
	if fdl == "" {
		return nil, nil
	}

	spec, err := ParseFDL(fdl)
	if err != nil {
		return nil, err
	}

	return GetFDLInfo(spec), nil
}

// VerifyResult contains the result of FDL implementation verification
type VerifyResult struct {
	Valid             bool              `json:"valid"`
	Errors            []string          `json:"errors,omitempty"`
	Warnings          []string          `json:"warnings,omitempty"`
	FunctionsMissing  []string          `json:"functions_missing,omitempty"`
	FunctionsExtra    []string          `json:"functions_extra,omitempty"`
	FilesMissing      []string          `json:"files_missing,omitempty"`
	SignatureMismatch []SignatureDiff   `json:"signature_mismatch,omitempty"`
	ModelsMissing     []string          `json:"models_missing,omitempty"`
	APIsMissing       []string          `json:"apis_missing,omitempty"`
}

// SignatureDiff represents a function signature mismatch
type SignatureDiff struct {
	Function string `json:"function"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

// VerifyFDLImplementation checks if code matches FDL spec
func VerifyFDLImplementation(database *db.DB, featureID int64) (*VerifyResult, error) {
	result := &VerifyResult{Valid: true}

	// Get feature and FDL
	feature, err := GetFeature(database, featureID)
	if err != nil {
		return nil, fmt.Errorf("get feature: %w", err)
	}

	if feature.FDL == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "No FDL defined for this feature")
		return result, nil
	}

	// Parse FDL
	spec, err := ParseFDL(feature.FDL)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse FDL: %v", err))
		return result, nil
	}

	// Get tech stack for file path determination
	tech, _ := GetTech(database)
	if tech == nil {
		tech = make(map[string]interface{})
	}

	// Get skeletons for this feature
	skeletons, err := ListSkeletonsByFeature(database, featureID)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not list skeletons: %v", err))
	}

	// Build expected file paths from FDL
	mappings, _ := ExtractTaskMappings(spec, tech)
	expectedFiles := make(map[string]bool)
	for _, m := range mappings {
		if m.TargetFile != "" {
			expectedFiles[m.TargetFile] = true
		}
	}

	// Check if skeleton files exist
	skeletonFiles := make(map[string]bool)
	for _, s := range skeletons {
		skeletonFiles[s.FilePath] = true
	}

	// Check for missing files
	for file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			if !skeletonFiles[file] {
				result.FilesMissing = append(result.FilesMissing, file)
				result.Valid = false
			}
		}
	}

	// Check models
	for _, m := range spec.Models {
		// For now, just verify that model files would be created
		// A full implementation would parse the actual code files
		found := false
		for _, mapping := range mappings {
			if mapping.Layer == "model" && strings.Contains(mapping.Title, m.Name) {
				found = true
				break
			}
		}
		if !found {
			result.ModelsMissing = append(result.ModelsMissing, m.Name)
		}
	}

	// Check services (function definitions)
	for _, s := range spec.Service {
		// Verify service function exists
		expectedFunc := s.Name
		funcFound := false

		// Check in skeleton files
		for _, skel := range skeletons {
			if skel.Layer == "service" {
				content, err := os.ReadFile(skel.FilePath)
				if err == nil {
					// Simple check: look for function name in content
					if strings.Contains(string(content), expectedFunc) {
						funcFound = true
						break
					}
				}
			}
		}

		if !funcFound {
			// Check if skeleton was generated
			if !feature.SkeletonGenerated {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Skeleton not generated, cannot verify function: %s", expectedFunc))
			} else {
				result.FunctionsMissing = append(result.FunctionsMissing, expectedFunc)
				result.Valid = false
			}
		}
	}

	// Check APIs
	for _, a := range spec.API {
		apiKey := fmt.Sprintf("%s %s", a.Method, a.Path)
		apiFound := false

		// Check in skeleton files
		for _, skel := range skeletons {
			if skel.Layer == "api" {
				content, err := os.ReadFile(skel.FilePath)
				if err == nil {
					// Simple check: look for path in content
					if strings.Contains(string(content), a.Path) {
						apiFound = true
						break
					}
				}
			}
		}

		if !apiFound && feature.SkeletonGenerated {
			result.APIsMissing = append(result.APIsMissing, apiKey)
			result.Valid = false
		}
	}

	// Add summary error if not valid
	if !result.Valid {
		errCount := len(result.FilesMissing) + len(result.FunctionsMissing) + len(result.APIsMissing) + len(result.ModelsMissing) + len(result.SignatureMismatch)
		result.Errors = append(result.Errors, fmt.Sprintf("Found %d verification issues", errCount))
	}

	return result, nil
}

// DiffResult contains the differences between FDL and actual implementation
type DiffResult struct {
	FeatureID    int64      `json:"feature_id"`
	FeatureName  string     `json:"feature_name"`
	Differences  []FileDiff `json:"differences"`
	TotalChanges int        `json:"total_changes"`
}

// FileDiff represents differences in a single file
type FileDiff struct {
	FilePath string   `json:"file_path"`
	Layer    string   `json:"layer"` // model, service, api, ui
	Changes  []Change `json:"changes"`
}

// Change represents a single change/difference
type Change struct {
	Type     string `json:"type"` // "missing", "modified", "extra"
	Element  string `json:"element"`
	Expected string `json:"expected,omitempty"`
	Actual   string `json:"actual,omitempty"`
}

// DiffFDLImplementation shows differences between FDL and actual code
func DiffFDLImplementation(database *db.DB, featureID int64) (*DiffResult, error) {
	result := &DiffResult{
		FeatureID:   featureID,
		Differences: []FileDiff{},
	}

	// Get feature and FDL
	feature, err := GetFeature(database, featureID)
	if err != nil {
		return nil, fmt.Errorf("get feature: %w", err)
	}
	result.FeatureName = feature.Name

	if feature.FDL == "" {
		return result, nil
	}

	// Parse FDL
	spec, err := ParseFDL(feature.FDL)
	if err != nil {
		return nil, fmt.Errorf("parse FDL: %w", err)
	}

	// Get tech stack for file path determination
	tech, _ := GetTech(database)
	if tech == nil {
		tech = make(map[string]interface{})
	}

	// Get skeletons for this feature
	skeletons, _ := ListSkeletonsByFeature(database, featureID)
	skeletonMap := make(map[string]*model.Skeleton)
	for i := range skeletons {
		skeletonMap[skeletons[i].FilePath] = &skeletons[i]
	}

	// Extract expected mappings
	mappings, _ := ExtractTaskMappings(spec, tech)

	// Group mappings by file
	fileGroups := make(map[string][]FDLTaskMapping)
	for _, m := range mappings {
		if m.TargetFile != "" {
			fileGroups[m.TargetFile] = append(fileGroups[m.TargetFile], m)
		}
	}

	// Check each expected file
	for filePath, fileMappings := range fileGroups {
		fileDiff := FileDiff{
			FilePath: filePath,
			Layer:    fileMappings[0].Layer,
			Changes:  []Change{},
		}

		// Check if file exists
		content, err := os.ReadFile(filePath)
		if os.IsNotExist(err) {
			// File missing
			fileDiff.Changes = append(fileDiff.Changes, Change{
				Type:    "missing",
				Element: "file",
			})

			// Check skeleton
			if skel, ok := skeletonMap[filePath]; ok {
				fileDiff.Changes[0].Expected = fmt.Sprintf("Generated skeleton at: %s", skel.FilePath)
			}
		} else if err != nil {
			// Other error reading file
			fileDiff.Changes = append(fileDiff.Changes, Change{
				Type:    "error",
				Element: "file",
				Actual:  err.Error(),
			})
		} else {
			// File exists, check content
			contentStr := string(content)

			for _, mapping := range fileMappings {
				if mapping.TargetFunction != "" {
					// Check if function exists
					if !strings.Contains(contentStr, mapping.TargetFunction) {
						fileDiff.Changes = append(fileDiff.Changes, Change{
							Type:     "missing",
							Element:  "function",
							Expected: mapping.TargetFunction,
						})
					}
				}
			}
		}

		if len(fileDiff.Changes) > 0 {
			result.Differences = append(result.Differences, fileDiff)
			result.TotalChanges += len(fileDiff.Changes)
		}
	}

	// Check for service function signatures
	for _, s := range spec.Service {
		for filePath := range fileGroups {
			if !strings.Contains(filePath, "service") {
				continue
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			contentStr := string(content)
			if strings.Contains(contentStr, s.Name) {
				// Function exists, could check signature here
				// For now, just note if input parameters seem different
				if s.Input != nil && len(s.Input) > 0 {
					inputParams := make([]string, 0, len(s.Input))
					for param := range s.Input {
						inputParams = append(inputParams, param)
					}

					allFound := true
					for _, param := range inputParams {
						if !strings.Contains(contentStr, param) {
							allFound = false
							break
						}
					}

					if !allFound {
						// Find or create file diff
						var fileDiff *FileDiff
						for i := range result.Differences {
							if result.Differences[i].FilePath == filePath {
								fileDiff = &result.Differences[i]
								break
							}
						}
						if fileDiff == nil {
							result.Differences = append(result.Differences, FileDiff{
								FilePath: filePath,
								Layer:    "service",
								Changes:  []Change{},
							})
							fileDiff = &result.Differences[len(result.Differences)-1]
						}

						fileDiff.Changes = append(fileDiff.Changes, Change{
							Type:     "modified",
							Element:  "signature",
							Expected: fmt.Sprintf("%s(%v)", s.Name, inputParams),
							Actual:   "parameters may differ",
						})
						result.TotalChanges++
					}
				}
			}
		}
	}

	return result, nil
}
