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
	Name        string        `yaml:"name"`
	Table       string        `yaml:"table"`
	Description string        `yaml:"description,omitempty"`
	Fields      []FDLField    `yaml:"fields"`
	Indexes     []FDLIndex    `yaml:"indexes,omitempty"`
	Relations   []FDLRelation `yaml:"relations,omitempty"`
	Patterns    []string      `yaml:"patterns,omitempty"`
}

// FDLField - 필드 정의
type FDLField struct {
	Name             string `yaml:"name"`
	Type             string `yaml:"type"`
	Constraints      string `yaml:"constraints,omitempty"`
	Description      string `yaml:"description,omitempty"`
	ParsedType       string // 기본 타입 (파싱 후)
	TypeOption       string // 타입 옵션 (50, 10,2, admin,user 등)
	ParsedConstraint FDLFieldConstraint
}

// FDLFieldConstraint - 파싱된 제약조건
type FDLFieldConstraint struct {
	IsPK       bool
	IsFK       bool
	FKRef      string // 참조 테이블.컬럼
	IsRequired bool
	IsUnique   bool
	IsAuto     bool
	IsIndex    bool
	IsNullable bool
	Default    string
	Check      string
	OnDelete   string
	OnUpdate   string
}

// FDLIndex - 인덱스 정의
type FDLIndex struct {
	Fields []string `yaml:"fields"`
	Unique bool     `yaml:"unique,omitempty"`
	Name   string   `yaml:"name,omitempty"`
	Where  string   `yaml:"where,omitempty"` // 부분 인덱스
}

// FDLRelation - 관계 정의
type FDLRelation struct {
	Type       string `yaml:"-"`        // hasOne, hasMany, belongsTo, belongsToMany
	Target     string `yaml:"-"`        // 대상 모델
	HasOne     string `yaml:"hasOne,omitempty"`
	HasMany    string `yaml:"hasMany,omitempty"`
	BelongsTo  string `yaml:"belongsTo,omitempty"`
	ForeignKey string `yaml:"foreignKey,omitempty"`
	As         string `yaml:"as,omitempty"` // 별칭
	Through    string `yaml:"through,omitempty"` // 중간 테이블
}

// FDLService - 서비스 함수 정의
type FDLService struct {
	Name        string                 `yaml:"name"`
	Desc        string                 `yaml:"desc"`
	Input       map[string]interface{} `yaml:"input"`
	Output      string                 `yaml:"output,omitempty"`
	Throws      []string               `yaml:"throws,omitempty"`
	Transaction bool                   `yaml:"transaction,omitempty"`
	Auth        string                 `yaml:"auth,omitempty"` // required, optional, none
	Roles       []string               `yaml:"roles,omitempty"`
	Ownership   string                 `yaml:"ownership,omitempty"`
	Steps       []interface{}          `yaml:"steps"`
	// Parsed fields (populated after parsing)
	ParsedInput  map[string]FDLInputParam `yaml:"-"`
	ParsedOutput FDLOutput                `yaml:"-"`
	ParsedSteps  []FDLStep                `yaml:"-"`
}

// FDLInputParam - 입력 파라미터 검증 규칙
type FDLInputParam struct {
	Type      string      `yaml:"type"`
	Required  bool        `yaml:"required,omitempty"`
	Optional  bool        `yaml:"optional,omitempty"`
	Default   interface{} `yaml:"default,omitempty"`
	MinLength int         `yaml:"minLength,omitempty"`
	MaxLength int         `yaml:"maxLength,omitempty"`
	Min       float64     `yaml:"min,omitempty"`
	Max       float64     `yaml:"max,omitempty"`
	Pattern   string      `yaml:"pattern,omitempty"`
	Enum      []string    `yaml:"enum,omitempty"`
}

// FDLOutput - 출력 타입 정의
type FDLOutput struct {
	Type      string            // User, void
	IsArray   bool              // Array<User>
	IsComplex bool              // { user: User, token: string }
	Fields    map[string]string // 복합 타입의 필드들
}

// FDLStep - 서비스 단계 정의
type FDLStep struct {
	Type      string                 // validate, db, event, call, cache, log, transform, condition, loop, return
	Operation string                 // 세부 동작
	Params    map[string]interface{} // 추가 파라미터
	Raw       string                 // 원본 문자열
}

// FDLAPI - API 엔드포인트 정의
type FDLAPI struct {
	Path      string                 `yaml:"path"`
	Method    string                 `yaml:"method"`
	Summary   string                 `yaml:"summary,omitempty"`
	Use       string                 `yaml:"use"` // service.FunctionName
	Request   map[string]interface{} `yaml:"request,omitempty"`
	Response  map[string]interface{} `yaml:"response"`
	Auth      string                 `yaml:"auth,omitempty"`      // required, optional, none, apiKey
	Roles     []string               `yaml:"roles,omitempty"`
	Tags      []string               `yaml:"tags,omitempty"`
	RateLimit map[string]interface{} `yaml:"rateLimit,omitempty"`
	Mapping   map[string]string      `yaml:"mapping,omitempty"`   // 파라미터 매핑
	Transform map[string]interface{} `yaml:"transform,omitempty"` // 응답 변환
	// Parsed fields
	ParsedRequest   FDLAPIRequest   `yaml:"-"`
	ParsedResponse  map[int]interface{} `yaml:"-"`
	ParsedRateLimit *FDLRateLimit   `yaml:"-"`
	ParsedTransform *FDLTransform   `yaml:"-"`
}

// FDLAPIRequest - API 요청 구조
type FDLAPIRequest struct {
	Params  map[string]FDLRequestParam // 경로 파라미터
	Query   map[string]FDLRequestParam // 쿼리스트링
	Headers map[string]FDLRequestParam
	Body    map[string]FDLRequestParam
}

// FDLRequestParam - 요청 파라미터 정의
type FDLRequestParam struct {
	Type      string
	Required  bool
	Default   interface{}
	Min       float64
	Max       float64
	MinLength int
	MaxLength int
	Pattern   string
	Enum      []string
}

// FDLRateLimit - Rate limiting 설정
type FDLRateLimit struct {
	Limit  int    // 요청 수
	Window int    // 시간 (초)
	By     string // ip, user, apiKey
}

// FDLTransform - 응답 변환 설정
type FDLTransform struct {
	Exclude []string
	Rename  map[string]string
}

// FDLUI - UI 컴포넌트 정의
type FDLUI struct {
	Component   string                   `yaml:"component"`
	Type        string                   `yaml:"type"` // Page, Template, Organism, Molecule, Atom
	Description string                   `yaml:"description,omitempty"`
	Parent      string                   `yaml:"parent,omitempty"`
	Props       map[string]interface{}   `yaml:"props,omitempty"`
	State       []interface{}            `yaml:"state,omitempty"`
	Computed    []interface{}            `yaml:"computed,omitempty"`
	Init        []interface{}            `yaml:"init,omitempty"`
	Methods     map[string]interface{}   `yaml:"methods,omitempty"`
	View        []map[string]interface{} `yaml:"view,omitempty"`
	Styles      map[string]interface{}   `yaml:"styles,omitempty"`
	// Parsed fields
	ParsedProps    map[string]FDLUIProp     `yaml:"-"`
	ParsedState    []FDLUIState             `yaml:"-"`
	ParsedComputed []FDLUIComputed          `yaml:"-"`
	ParsedInit     []FDLUIAction            `yaml:"-"`
	ParsedMethods  map[string][]FDLUIAction `yaml:"-"`
	ParsedView     []FDLUIElement           `yaml:"-"`
}

// FDLUIProp - UI 프로퍼티 정의
type FDLUIProp struct {
	Type     string // string, number, boolean, function, object, array
	Required bool
	Optional bool
	Default  interface{}
}

// FDLUIState - UI 상태 정의
type FDLUIState struct {
	Name    string
	Type    string
	Default interface{}
}

// FDLUIComputed - 계산된 속성 정의
type FDLUIComputed struct {
	Name       string
	Expression string
}

// FDLUIAction - UI 액션 정의
type FDLUIAction struct {
	Type      string // call, set, navigate, show, validate, confirm, emit, parallel, redirect
	Target    string
	Params    map[string]interface{}
	OnSuccess []FDLUIAction
	OnError   []FDLUIAction
	Raw       string
}

// FDLUIElement - UI 요소 정의
type FDLUIElement struct {
	Type      string // Text, Input, Button, Image, Flex, Grid, Stack, etc.
	Props     map[string]interface{}
	Children  []FDLUIElement
	Condition *FDLUICondition // if/else
}

// FDLUICondition - UI 조건부 렌더링
type FDLUICondition struct {
	If   string
	Then []FDLUIElement
	Else []FDLUIElement
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

// ValidFieldTypes is the list of valid FDL field types
var ValidFieldTypes = []string{
	"uuid", "string", "text", "int", "bigint", "float", "decimal", "boolean",
	"datetime", "date", "time", "json", "blob", "enum",
}

// parseFieldType parses field type string and extracts base type and option
// Examples: "string(50)" -> "string", "50"
//           "decimal(10,2)" -> "decimal", "10,2"
//           "enum(admin,user,guest)" -> "enum", "admin,user,guest"
func parseFieldType(typeStr string) (baseType, option string) {
	typeStr = strings.TrimSpace(typeStr)

	// Find parentheses
	parenStart := strings.Index(typeStr, "(")
	if parenStart == -1 {
		return typeStr, ""
	}

	parenEnd := strings.LastIndex(typeStr, ")")
	if parenEnd == -1 || parenEnd <= parenStart {
		return typeStr, ""
	}

	baseType = strings.TrimSpace(typeStr[:parenStart])
	option = strings.TrimSpace(typeStr[parenStart+1 : parenEnd])
	return baseType, option
}

// parseConstraints parses constraint string into FDLFieldConstraint struct
// Examples: "pk, required, unique" -> FDLFieldConstraint{IsPK: true, IsRequired: true, IsUnique: true}
//           "fk(users.id), onDelete: cascade" -> FDLFieldConstraint{IsFK: true, FKRef: "users.id", OnDelete: "cascade"}
func parseConstraints(constraintStr string) FDLFieldConstraint {
	result := FDLFieldConstraint{}
	if constraintStr == "" {
		return result
	}

	// Split by comma but handle parentheses
	parts := splitConstraints(constraintStr)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		partLower := strings.ToLower(part)

		switch {
		case partLower == "pk":
			result.IsPK = true
		case partLower == "required":
			result.IsRequired = true
		case partLower == "unique":
			result.IsUnique = true
		case partLower == "auto":
			result.IsAuto = true
		case partLower == "index":
			result.IsIndex = true
		case partLower == "nullable":
			result.IsNullable = true
		case strings.HasPrefix(partLower, "fk("):
			result.IsFK = true
			// Extract reference: fk(table.column)
			if idx := strings.Index(part, "("); idx != -1 {
				ref := part[idx+1:]
				if endIdx := strings.Index(ref, ")"); endIdx != -1 {
					result.FKRef = strings.TrimSpace(ref[:endIdx])
				}
			}
		case strings.HasPrefix(partLower, "default:"):
			result.Default = strings.TrimSpace(part[8:])
		case strings.HasPrefix(partLower, "check:"):
			result.Check = strings.TrimSpace(part[6:])
		case strings.HasPrefix(partLower, "ondelete:"):
			result.OnDelete = strings.TrimSpace(part[9:])
		case strings.HasPrefix(partLower, "onupdate:"):
			result.OnUpdate = strings.TrimSpace(part[9:])
		}
	}

	return result
}

// splitConstraints splits constraint string by comma, handling parentheses
func splitConstraints(s string) []string {
	var result []string
	var current strings.Builder
	depth := 0

	for _, ch := range s {
		switch ch {
		case '(':
			depth++
			current.WriteRune(ch)
		case ')':
			depth--
			current.WriteRune(ch)
		case ',':
			if depth == 0 {
				if str := strings.TrimSpace(current.String()); str != "" {
					result = append(result, str)
				}
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if str := strings.TrimSpace(current.String()); str != "" {
		result = append(result, str)
	}

	return result
}

// isValidFieldType checks if the type is a valid FDL type
func isValidFieldType(fieldType string) bool {
	for _, t := range ValidFieldTypes {
		if t == fieldType {
			return true
		}
	}
	return false
}

// ParseAndValidateField parses field type and constraints, returns errors if invalid
func ParseAndValidateField(field *FDLField) error {
	// Parse type
	baseType, option := parseFieldType(field.Type)
	field.ParsedType = baseType
	field.TypeOption = option

	// Validate type
	if !isValidFieldType(baseType) {
		return fmt.Errorf("invalid field type: %s (valid: %v)", baseType, ValidFieldTypes)
	}

	// Validate type-specific constraints
	switch baseType {
	case "string":
		// Option should be a number (max length)
		if option != "" {
			// Allow just a number
			if _, err := fmt.Sscanf(option, "%d", new(int)); err != nil {
				return fmt.Errorf("string type option must be a number (length), got: %s", option)
			}
		}
	case "decimal":
		// Option should be precision,scale
		if option != "" {
			parts := strings.Split(option, ",")
			if len(parts) != 2 {
				return fmt.Errorf("decimal type requires precision,scale format, got: %s", option)
			}
		}
	case "enum":
		// Option should be comma-separated values
		if option == "" {
			return fmt.Errorf("enum type requires values in parentheses")
		}
	}

	// Parse constraints
	field.ParsedConstraint = parseConstraints(field.Constraints)

	return nil
}

// ParseModelRelations extracts relation type and target from FDLRelation
func ParseModelRelations(model *FDLModel) {
	for i := range model.Relations {
		rel := &model.Relations[i]
		if rel.HasOne != "" {
			rel.Type = "hasOne"
			rel.Target = rel.HasOne
		} else if rel.HasMany != "" {
			rel.Type = "hasMany"
			rel.Target = rel.HasMany
		} else if rel.BelongsTo != "" {
			rel.Type = "belongsTo"
			rel.Target = rel.BelongsTo
		}
	}
}

// ValidateModelAdvanced performs advanced validation on a model with FK and relation checks
func ValidateModelAdvanced(model *FDLModel, allModels []FDLModel) []error {
	var errors []error

	// Parse relations first
	ParseModelRelations(model)

	// Validate each field
	for i := range model.Fields {
		if err := ParseAndValidateField(&model.Fields[i]); err != nil {
			errors = append(errors, fmt.Errorf("model %s, field %s: %w", model.Name, model.Fields[i].Name, err))
		}
	}

	// Validate FK references
	for _, field := range model.Fields {
		if field.ParsedConstraint.IsFK && field.ParsedConstraint.FKRef != "" {
			if !fkRefExists(field.ParsedConstraint.FKRef, allModels) {
				errors = append(errors, fmt.Errorf("model %s, field %s: FK reference not found: %s", model.Name, field.Name, field.ParsedConstraint.FKRef))
			}
		}
	}

	// Validate relations
	for _, rel := range model.Relations {
		if rel.Target != "" && !modelExistsByName(rel.Target, allModels) {
			errors = append(errors, fmt.Errorf("model %s: relation target not found: %s", model.Name, rel.Target))
		}
	}

	return errors
}

// fkRefExists checks if a FK reference exists in the models
// FK ref format: "table.column" or "ModelName.field"
func fkRefExists(fkRef string, allModels []FDLModel) bool {
	parts := strings.Split(fkRef, ".")
	if len(parts) != 2 {
		return false
	}

	tableName := parts[0]
	fieldName := parts[1]

	for _, m := range allModels {
		if m.Table == tableName || m.Name == tableName {
			for _, f := range m.Fields {
				if f.Name == fieldName {
					return true
				}
			}
		}
	}
	return false
}

// modelExistsByName checks if a model with the given name exists
func modelExistsByName(name string, allModels []FDLModel) bool {
	for _, m := range allModels {
		if m.Name == name {
			return true
		}
	}
	return false
}

// parseInputParams parses input parameters from raw map
func parseInputParams(raw map[string]interface{}) map[string]FDLInputParam {
	result := make(map[string]FDLInputParam)
	if raw == nil {
		return result
	}

	for name, spec := range raw {
		param := FDLInputParam{}

		switch v := spec.(type) {
		case string:
			// Simple format: email: string
			param.Type = v
		case map[string]interface{}:
			// Detailed format with validation
			if t, ok := v["type"].(string); ok {
				param.Type = t
			}
			if r, ok := v["required"].(bool); ok {
				param.Required = r
			}
			if o, ok := v["optional"].(bool); ok {
				param.Optional = o
			}
			if d, ok := v["default"]; ok {
				param.Default = d
			}
			if ml, ok := v["minLength"].(int); ok {
				param.MinLength = ml
			} else if ml, ok := v["minLength"].(float64); ok {
				param.MinLength = int(ml)
			}
			if ml, ok := v["maxLength"].(int); ok {
				param.MaxLength = ml
			} else if ml, ok := v["maxLength"].(float64); ok {
				param.MaxLength = int(ml)
			}
			if m, ok := v["min"].(float64); ok {
				param.Min = m
			} else if m, ok := v["min"].(int); ok {
				param.Min = float64(m)
			}
			if m, ok := v["max"].(float64); ok {
				param.Max = m
			} else if m, ok := v["max"].(int); ok {
				param.Max = float64(m)
			}
			if p, ok := v["pattern"].(string); ok {
				param.Pattern = p
			}
			if e, ok := v["enum"].([]interface{}); ok {
				for _, item := range e {
					if s, ok := item.(string); ok {
						param.Enum = append(param.Enum, s)
					}
				}
			}
		}

		result[name] = param
	}

	return result
}

// parseOutput parses output type string
func parseOutput(raw string) FDLOutput {
	output := FDLOutput{}

	if raw == "" || raw == "void" {
		output.Type = "void"
		return output
	}

	raw = strings.TrimSpace(raw)

	// Check for Array<Type>
	if strings.HasPrefix(raw, "Array<") && strings.HasSuffix(raw, ">") {
		output.IsArray = true
		output.Type = strings.TrimSuffix(strings.TrimPrefix(raw, "Array<"), ">")
		return output
	}

	// Check for complex type { field: Type, ... }
	if strings.HasPrefix(raw, "{") && strings.HasSuffix(raw, "}") {
		output.IsComplex = true
		output.Fields = make(map[string]string)

		// Parse inner content
		inner := strings.TrimSuffix(strings.TrimPrefix(raw, "{"), "}")
		inner = strings.TrimSpace(inner)

		// Split by comma and parse each field
		fields := strings.Split(inner, ",")
		for _, field := range fields {
			field = strings.TrimSpace(field)
			parts := strings.SplitN(field, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				output.Fields[key] = value
			}
		}
		return output
	}

	// Simple type
	output.Type = raw
	return output
}

// ValidStepTypes is the list of valid step types
var ValidStepTypes = []string{
	"validate", "db", "event", "call", "cache", "log",
	"transform", "condition", "loop", "return", "if", "throw",
}

// parseSteps parses steps from raw slice
func parseSteps(raw []interface{}) []FDLStep {
	var steps []FDLStep

	for _, s := range raw {
		switch v := s.(type) {
		case string:
			steps = append(steps, parseStepString(v))
		case map[string]interface{}:
			steps = append(steps, parseStepMap(v))
		}
	}

	return steps
}

// parseStepString parses a step from string format
// Examples: "validate: email format" -> Step{Type: "validate", Operation: "email format"}
//           "db: select User where email = $email" -> Step{Type: "db", Operation: "select User..."}
func parseStepString(s string) FDLStep {
	step := FDLStep{Raw: s}

	// Look for "type: operation" format
	if idx := strings.Index(s, ":"); idx > 0 {
		possibleType := strings.TrimSpace(s[:idx])
		// Check if it's a valid step type
		for _, t := range ValidStepTypes {
			if strings.EqualFold(possibleType, t) {
				step.Type = strings.ToLower(possibleType)
				step.Operation = strings.TrimSpace(s[idx+1:])
				return step
			}
		}
	}

	// If no type prefix, try to infer from keywords
	lower := strings.ToLower(s)
	switch {
	case strings.HasPrefix(lower, "return"):
		step.Type = "return"
		step.Operation = strings.TrimSpace(strings.TrimPrefix(lower, "return"))
	case strings.HasPrefix(lower, "throw"):
		step.Type = "throw"
		step.Operation = strings.TrimSpace(s[5:])
	case strings.HasPrefix(lower, "if "):
		step.Type = "condition"
		step.Operation = s
	case strings.Contains(lower, "emit"):
		step.Type = "event"
		step.Operation = s
	default:
		step.Type = "unknown"
		step.Operation = s
	}

	return step
}

// parseStepMap parses a step from map format
func parseStepMap(m map[string]interface{}) FDLStep {
	step := FDLStep{
		Params: make(map[string]interface{}),
	}

	// Look for step type key
	for _, t := range ValidStepTypes {
		if val, ok := m[t]; ok {
			step.Type = t
			if s, ok := val.(string); ok {
				step.Operation = s
			}
			break
		}
	}

	// Copy remaining params
	for k, v := range m {
		if k != step.Type {
			step.Params[k] = v
		}
	}

	return step
}

// ParseAndValidateService parses and validates a service definition
func ParseAndValidateService(svc *FDLService) error {
	// Parse input
	svc.ParsedInput = parseInputParams(svc.Input)

	// Parse output
	svc.ParsedOutput = parseOutput(svc.Output)

	// Parse steps
	if svc.Steps != nil {
		svc.ParsedSteps = parseSteps(svc.Steps)
	}

	// Validate auth value
	if svc.Auth != "" {
		validAuth := map[string]bool{
			"required": true,
			"optional": true,
			"none":     true,
		}
		if !validAuth[strings.ToLower(svc.Auth)] {
			return fmt.Errorf("service %s: invalid auth value: %s (must be required, optional, or none)", svc.Name, svc.Auth)
		}
	}

	return nil
}

// parseAPIRequest parses API request from raw map
func parseAPIRequest(raw map[string]interface{}) FDLAPIRequest {
	req := FDLAPIRequest{
		Params:  make(map[string]FDLRequestParam),
		Query:   make(map[string]FDLRequestParam),
		Headers: make(map[string]FDLRequestParam),
		Body:    make(map[string]FDLRequestParam),
	}

	if raw == nil {
		return req
	}

	if params, ok := raw["params"].(map[string]interface{}); ok {
		for name, spec := range params {
			req.Params[name] = parseRequestParam(spec)
		}
	}

	if query, ok := raw["query"].(map[string]interface{}); ok {
		for name, spec := range query {
			req.Query[name] = parseRequestParam(spec)
		}
	}

	if headers, ok := raw["headers"].(map[string]interface{}); ok {
		for name, spec := range headers {
			req.Headers[name] = parseRequestParam(spec)
		}
	}

	if body, ok := raw["body"].(map[string]interface{}); ok {
		for name, spec := range body {
			req.Body[name] = parseRequestParam(spec)
		}
	}

	return req
}

// parseRequestParam parses a single request parameter
func parseRequestParam(spec interface{}) FDLRequestParam {
	param := FDLRequestParam{}

	switch v := spec.(type) {
	case string:
		// Simple format: "int (required)" or "string"
		param.Type, param.Required = parseParamTypeString(v)
	case map[string]interface{}:
		// Detailed format
		if t, ok := v["type"].(string); ok {
			param.Type = t
		}
		if r, ok := v["required"].(bool); ok {
			param.Required = r
		}
		if d, ok := v["default"]; ok {
			param.Default = d
		}
		if m, ok := v["min"].(float64); ok {
			param.Min = m
		} else if m, ok := v["min"].(int); ok {
			param.Min = float64(m)
		}
		if m, ok := v["max"].(float64); ok {
			param.Max = m
		} else if m, ok := v["max"].(int); ok {
			param.Max = float64(m)
		}
		if ml, ok := v["minLength"].(int); ok {
			param.MinLength = ml
		} else if ml, ok := v["minLength"].(float64); ok {
			param.MinLength = int(ml)
		}
		if ml, ok := v["maxLength"].(int); ok {
			param.MaxLength = ml
		} else if ml, ok := v["maxLength"].(float64); ok {
			param.MaxLength = int(ml)
		}
		if p, ok := v["pattern"].(string); ok {
			param.Pattern = p
		}
		if e, ok := v["enum"].([]interface{}); ok {
			for _, item := range e {
				if s, ok := item.(string); ok {
					param.Enum = append(param.Enum, s)
				}
			}
		}
	}

	return param
}

// parseParamTypeString parses "int (required)" format
func parseParamTypeString(s string) (typeName string, required bool) {
	s = strings.TrimSpace(s)

	// Check for (required) or (default: ...) suffix
	if idx := strings.Index(s, "("); idx > 0 {
		typeName = strings.TrimSpace(s[:idx])
		suffix := strings.ToLower(s[idx:])
		required = strings.Contains(suffix, "required")
	} else {
		typeName = s
	}

	return typeName, required
}

// parseAPIResponse parses API response from raw map
func parseAPIResponse(raw map[string]interface{}) map[int]interface{} {
	result := make(map[int]interface{})
	if raw == nil {
		return result
	}

	for codeStr, schema := range raw {
		var code int
		if _, err := fmt.Sscanf(codeStr, "%d", &code); err == nil {
			result[code] = schema
		}
	}

	return result
}

// parseRateLimit parses rate limit configuration
func parseRateLimit(raw map[string]interface{}) *FDLRateLimit {
	if raw == nil {
		return nil
	}

	rl := &FDLRateLimit{}

	if limit, ok := raw["limit"].(int); ok {
		rl.Limit = limit
	} else if limit, ok := raw["limit"].(float64); ok {
		rl.Limit = int(limit)
	}

	if window, ok := raw["window"].(int); ok {
		rl.Window = window
	} else if window, ok := raw["window"].(float64); ok {
		rl.Window = int(window)
	}

	if by, ok := raw["by"].(string); ok {
		rl.By = by
	}

	return rl
}

// parseTransform parses transform configuration
func parseTransform(raw map[string]interface{}) *FDLTransform {
	if raw == nil {
		return nil
	}

	tr := &FDLTransform{
		Rename: make(map[string]string),
	}

	if exclude, ok := raw["exclude"].([]interface{}); ok {
		for _, item := range exclude {
			if s, ok := item.(string); ok {
				tr.Exclude = append(tr.Exclude, s)
			}
		}
	}

	if rename, ok := raw["rename"].(map[string]interface{}); ok {
		for k, v := range rename {
			if s, ok := v.(string); ok {
				tr.Rename[k] = s
			}
		}
	}

	return tr
}

// ValidHTTPStatusCodes is the list of valid HTTP status codes
var ValidHTTPStatusCodes = []int{
	200, 201, 202, 204,
	400, 401, 403, 404, 405, 409, 422, 429,
	500, 501, 502, 503,
}

// ParseAndValidateAPI parses and validates an API definition
func ParseAndValidateAPI(api *FDLAPI, services []FDLService) error {
	// Parse request
	api.ParsedRequest = parseAPIRequest(api.Request)

	// Parse response
	api.ParsedResponse = parseAPIResponse(api.Response)

	// Parse rate limit
	api.ParsedRateLimit = parseRateLimit(api.RateLimit)

	// Parse transform
	api.ParsedTransform = parseTransform(api.Transform)

	// Validate method
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true, "HEAD": true, "OPTIONS": true,
	}
	if !validMethods[strings.ToUpper(api.Method)] {
		return fmt.Errorf("API %s: invalid method: %s", api.Path, api.Method)
	}

	// Validate path
	if !strings.HasPrefix(api.Path, "/") {
		return fmt.Errorf("API path must start with /: %s", api.Path)
	}

	// Validate auth
	if api.Auth != "" {
		validAuth := map[string]bool{
			"required": true, "optional": true, "none": true, "apikey": true, "bearer": true,
		}
		if !validAuth[strings.ToLower(api.Auth)] {
			return fmt.Errorf("API %s: invalid auth value: %s", api.Path, api.Auth)
		}
	}

	// Validate use reference
	if api.Use != "" && !strings.HasPrefix(api.Use, "service.") {
		return fmt.Errorf("API %s: use must reference service (service.FunctionName)", api.Path)
	}

	return nil
}

// CalculateFDLHashFromSpec calculates SHA256 hash of FDL spec
func CalculateFDLHashFromSpec(yamlStr string) string {
	hash := sha256.Sum256([]byte(yamlStr))
	return fmt.Sprintf("%x", hash)
}

// parseUIProps parses UI props from raw map
func parseUIProps(raw map[string]interface{}) map[string]FDLUIProp {
	result := make(map[string]FDLUIProp)
	if raw == nil {
		return result
	}

	for name, spec := range raw {
		prop := FDLUIProp{}

		switch v := spec.(type) {
		case string:
			// Simple format: "userId: string"
			prop.Type = v
		case map[string]interface{}:
			// Detailed format with type, required, optional, default
			if t, ok := v["type"].(string); ok {
				prop.Type = t
			}
			if r, ok := v["required"].(bool); ok {
				prop.Required = r
			}
			if o, ok := v["optional"].(bool); ok {
				prop.Optional = o
			}
			if d, ok := v["default"]; ok {
				prop.Default = d
			}
		}

		result[name] = prop
	}

	return result
}

// parseUIState parses UI state from raw slice
// Format: "- user: User | null" or "- isLoading: boolean = false"
func parseUIState(raw []interface{}) []FDLUIState {
	var states []FDLUIState
	if raw == nil {
		return states
	}

	for _, item := range raw {
		state := FDLUIState{}

		switch v := item.(type) {
		case string:
			// Parse "name: type = default" format
			state = parseStateString(v)
		case map[string]interface{}:
			// Parse map format
			for name, spec := range v {
				state.Name = name
				if specMap, ok := spec.(map[string]interface{}); ok {
					if t, ok := specMap["type"].(string); ok {
						state.Type = t
					}
					if d, ok := specMap["default"]; ok {
						state.Default = d
					}
				} else if specStr, ok := spec.(string); ok {
					// "name: type" format within map
					state.Type = specStr
				}
				break // Only first key-value pair
			}
		}

		if state.Name != "" {
			states = append(states, state)
		}
	}

	return states
}

// parseStateString parses state string like "user: User | null" or "isLoading: boolean = false"
func parseStateString(s string) FDLUIState {
	state := FDLUIState{}
	s = strings.TrimSpace(s)

	// Find name: type part
	colonIdx := strings.Index(s, ":")
	if colonIdx == -1 {
		state.Name = s
		return state
	}

	state.Name = strings.TrimSpace(s[:colonIdx])
	rest := strings.TrimSpace(s[colonIdx+1:])

	// Check for default value
	eqIdx := strings.Index(rest, "=")
	if eqIdx != -1 {
		state.Type = strings.TrimSpace(rest[:eqIdx])
		defaultStr := strings.TrimSpace(rest[eqIdx+1:])
		// Parse default value
		state.Default = parseDefaultValue(defaultStr)
	} else {
		state.Type = rest
	}

	return state
}

// parseDefaultValue parses default value string
func parseDefaultValue(s string) interface{} {
	s = strings.TrimSpace(s)

	// Boolean
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	// Null
	if s == "null" || s == "nil" {
		return nil
	}

	// Number
	if num, err := fmt.Sscanf(s, "%f", new(float64)); err == nil && num == 1 {
		var f float64
		fmt.Sscanf(s, "%f", &f)
		// Return as int if whole number
		if f == float64(int(f)) {
			return int(f)
		}
		return f
	}

	// String (remove quotes if present)
	if (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		return s[1 : len(s)-1]
	}

	// Empty string
	if s == `""` || s == "''" {
		return ""
	}

	return s
}

// parseUIComputed parses UI computed properties from raw slice
// Format: "- fullName: user.firstName + \" \" + user.lastName"
func parseUIComputed(raw []interface{}) []FDLUIComputed {
	var computed []FDLUIComputed
	if raw == nil {
		return computed
	}

	for _, item := range raw {
		comp := FDLUIComputed{}

		switch v := item.(type) {
		case string:
			// Parse "name: expression" format
			colonIdx := strings.Index(v, ":")
			if colonIdx != -1 {
				comp.Name = strings.TrimSpace(v[:colonIdx])
				comp.Expression = strings.TrimSpace(v[colonIdx+1:])
			} else {
				comp.Name = v
			}
		case map[string]interface{}:
			for name, expr := range v {
				comp.Name = name
				if exprStr, ok := expr.(string); ok {
					comp.Expression = exprStr
				}
				break
			}
		}

		if comp.Name != "" {
			computed = append(computed, comp)
		}
	}

	return computed
}

// parseUIInit parses UI init actions from raw slice
func parseUIInit(raw []interface{}) []FDLUIAction {
	return parseUIActions(raw)
}

// parseUIMethods parses UI methods from raw map
func parseUIMethods(raw map[string]interface{}) map[string][]FDLUIAction {
	result := make(map[string][]FDLUIAction)
	if raw == nil {
		return result
	}

	for name, spec := range raw {
		if steps, ok := spec.([]interface{}); ok {
			result[name] = parseUIActions(steps)
		}
	}

	return result
}

// parseUIActions parses a slice of UI actions
func parseUIActions(raw []interface{}) []FDLUIAction {
	var actions []FDLUIAction
	if raw == nil {
		return actions
	}

	for _, item := range raw {
		action := parseUIAction(item)
		if action.Type != "" || action.Raw != "" {
			actions = append(actions, action)
		}
	}

	return actions
}

// parseUIAction parses a single UI action
func parseUIAction(item interface{}) FDLUIAction {
	action := FDLUIAction{
		Params: make(map[string]interface{}),
	}

	switch v := item.(type) {
	case string:
		// Parse string format: "call: api.getUser($userId)" or "navigate: /users"
		action.Raw = v
		action = parseUIActionString(v)
	case map[string]interface{}:
		// Parse map format
		action = parseUIActionMap(v)
	}

	return action
}

// ValidUIActionTypes is the list of valid UI action types
var ValidUIActionTypes = []string{
	"call", "set", "navigate", "show", "validate", "confirm",
	"emit", "parallel", "redirect", "if", "throw", "return",
}

// parseUIActionString parses action string format
func parseUIActionString(s string) FDLUIAction {
	action := FDLUIAction{
		Raw:    s,
		Params: make(map[string]interface{}),
	}

	// Look for "type: target" format
	colonIdx := strings.Index(s, ":")
	if colonIdx > 0 {
		possibleType := strings.TrimSpace(s[:colonIdx])
		for _, t := range ValidUIActionTypes {
			if strings.EqualFold(possibleType, t) {
				action.Type = strings.ToLower(possibleType)
				action.Target = strings.TrimSpace(s[colonIdx+1:])
				return action
			}
		}
	}

	// Try to infer type from keywords
	lower := strings.ToLower(s)
	switch {
	case strings.HasPrefix(lower, "call "):
		action.Type = "call"
		action.Target = strings.TrimSpace(s[5:])
	case strings.HasPrefix(lower, "set "):
		action.Type = "set"
		action.Target = strings.TrimSpace(s[4:])
	case strings.HasPrefix(lower, "navigate "):
		action.Type = "navigate"
		action.Target = strings.TrimSpace(s[9:])
	case strings.HasPrefix(lower, "show "):
		action.Type = "show"
		action.Target = strings.TrimSpace(s[5:])
	case strings.HasPrefix(lower, "validate "):
		action.Type = "validate"
		action.Target = strings.TrimSpace(s[9:])
	case strings.HasPrefix(lower, "emit "):
		action.Type = "emit"
		action.Target = strings.TrimSpace(s[5:])
	default:
		action.Type = "unknown"
		action.Target = s
	}

	return action
}

// parseUIActionMap parses action map format
func parseUIActionMap(m map[string]interface{}) FDLUIAction {
	action := FDLUIAction{
		Params: make(map[string]interface{}),
	}

	// Look for action type key
	for _, t := range ValidUIActionTypes {
		if val, ok := m[t]; ok {
			action.Type = t
			if s, ok := val.(string); ok {
				action.Target = s
			}
			break
		}
	}

	// Parse additional params
	for k, v := range m {
		switch k {
		case action.Type:
			// Already processed
		case "set":
			if action.Type == "" {
				action.Type = "set"
				if s, ok := v.(string); ok {
					action.Target = s
				}
			} else {
				action.Params["set"] = v
			}
		case "onSuccess":
			if steps, ok := v.([]interface{}); ok {
				action.OnSuccess = parseUIActions(steps)
			}
		case "onError":
			if steps, ok := v.([]interface{}); ok {
				action.OnError = parseUIActions(steps)
			}
		default:
			action.Params[k] = v
		}
	}

	return action
}

// parseUIView parses UI view from raw slice
func parseUIView(raw []map[string]interface{}) []FDLUIElement {
	var elements []FDLUIElement
	if raw == nil {
		return elements
	}

	for _, item := range raw {
		element := parseUIElement(item)
		elements = append(elements, element)
	}

	return elements
}

// parseUIElement parses a single UI element from map
func parseUIElement(raw map[string]interface{}) FDLUIElement {
	element := FDLUIElement{
		Props: make(map[string]interface{}),
	}

	// Check for conditional rendering first (before iterating, since map order is random)
	if _, ok := raw["if"]; ok {
		element.Condition = parseUICondition(raw)
		return element
	}

	for key, value := range raw {
		// Skip condition-related keys (handled separately)
		if key == "then" || key == "else" {
			continue
		}

		// The key is the element type
		element.Type = key

		switch v := value.(type) {
		case string:
			// Simple element: "Text: Hello"
			element.Props["text"] = v
		case map[string]interface{}:
			// Element with props
			for propKey, propValue := range v {
				if propKey == "children" {
					// Recursive children parsing
					if children, ok := propValue.([]interface{}); ok {
						element.Children = parseUIViewInterface(children)
					}
				} else {
					element.Props[propKey] = propValue
				}
			}
		case []interface{}:
			// Element with array children
			element.Children = parseUIViewInterface(v)
		}

		break // Process only first key-value pair
	}

	return element
}

// parseUIViewInterface parses view from interface slice (for recursive parsing)
func parseUIViewInterface(raw []interface{}) []FDLUIElement {
	var elements []FDLUIElement
	for _, item := range raw {
		if itemMap, ok := item.(map[string]interface{}); ok {
			elements = append(elements, parseUIElement(itemMap))
		}
	}
	return elements
}

// parseUICondition parses conditional rendering
func parseUICondition(raw map[string]interface{}) *FDLUICondition {
	condition := &FDLUICondition{}

	if ifExpr, ok := raw["if"].(string); ok {
		condition.If = ifExpr
	}

	if thenItems, ok := raw["then"].([]interface{}); ok {
		condition.Then = parseUIViewInterface(thenItems)
	}

	if elseItems, ok := raw["else"].([]interface{}); ok {
		condition.Else = parseUIViewInterface(elseItems)
	}

	return condition
}

// ValidUITypes is the list of valid UI component types
var ValidUITypes = []string{"Page", "Template", "Organism", "Molecule", "Atom"}

// ParseAndValidateUI parses and validates a UI component definition
func ParseAndValidateUI(ui *FDLUI, allUIs []FDLUI) []error {
	var errors []error

	// Parse props
	ui.ParsedProps = parseUIProps(ui.Props)

	// Parse state
	ui.ParsedState = parseUIState(ui.State)

	// Parse computed
	ui.ParsedComputed = parseUIComputed(ui.Computed)

	// Parse init
	ui.ParsedInit = parseUIInit(ui.Init)

	// Parse methods
	ui.ParsedMethods = parseUIMethods(ui.Methods)

	// Parse view
	ui.ParsedView = parseUIView(ui.View)

	// Validate type
	if ui.Type != "" {
		validType := false
		for _, t := range ValidUITypes {
			if ui.Type == t {
				validType = true
				break
			}
		}
		if !validType {
			errors = append(errors, fmt.Errorf("UI %s: invalid type: %s (must be Page, Template, Organism, Molecule, or Atom)", ui.Component, ui.Type))
		}
	}

	// Validate parent reference
	if ui.Parent != "" {
		parentFound := false
		for _, other := range allUIs {
			if other.Component == ui.Parent {
				parentFound = true
				break
			}
		}
		if !parentFound {
			errors = append(errors, fmt.Errorf("UI %s: parent component not found: %s", ui.Component, ui.Parent))
		}
	}

	// Validate props have types
	for name, prop := range ui.ParsedProps {
		if prop.Type == "" {
			errors = append(errors, fmt.Errorf("UI %s: prop '%s' has no type", ui.Component, name))
		}
	}

	return errors
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
