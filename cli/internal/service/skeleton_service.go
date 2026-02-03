package service

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/model"
)

// TechConfig holds tech stack configuration
type TechConfig struct {
	Backend  string // go, python, node, java
	Frontend string // react, vue, angular, svelte
}

// ParseTechConfig parses tech map into TechConfig
func ParseTechConfig(tech map[string]interface{}) TechConfig {
	config := TechConfig{
		Backend:  "python",
		Frontend: "react",
	}
	if tech == nil {
		return config
	}
	if b, ok := tech["backend"].(string); ok {
		config.Backend = strings.ToLower(b)
	}
	if f, ok := tech["frontend"].(string); ok {
		config.Frontend = strings.ToLower(f)
	}
	return config
}

// toSnakeCase converts PascalCase to snake_case
func toSnakeCase(s string) string {
	var result bytes.Buffer
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// toPascalCase converts snake_case to PascalCase
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	var result bytes.Buffer
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(string(part[0])))
			if len(part) > 1 {
				result.WriteString(part[1:])
			}
		}
	}
	return result.String()
}

// toCamelCase converts snake_case to camelCase
func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if len(pascal) == 0 {
		return ""
	}
	return strings.ToLower(string(pascal[0])) + pascal[1:]
}

// GetModelPath returns model file path based on tech config
func GetModelPath(tech TechConfig, modelName string) string {
	snake := toSnakeCase(modelName)
	switch tech.Backend {
	case "go", "golang":
		return fmt.Sprintf("internal/model/%s.go", snake)
	case "python", "fastapi", "django":
		return fmt.Sprintf("models/%s.py", snake)
	case "node", "typescript", "express":
		return fmt.Sprintf("src/models/%s.ts", toCamelCase(modelName))
	case "java", "spring":
		return fmt.Sprintf("src/main/java/models/%s.java", toPascalCase(modelName))
	default:
		return fmt.Sprintf("models/%s.py", snake)
	}
}

// GetServicePath returns service file path based on tech config
func GetServicePath(tech TechConfig, serviceName, featureName string) string {
	snake := toSnakeCase(featureName)
	switch tech.Backend {
	case "go", "golang":
		return fmt.Sprintf("internal/service/%s_service.go", snake)
	case "python", "fastapi", "django":
		return fmt.Sprintf("services/%s_service.py", snake)
	case "node", "typescript", "express":
		return fmt.Sprintf("src/services/%sService.ts", toCamelCase(featureName))
	case "java", "spring":
		return fmt.Sprintf("src/main/java/services/%sService.java", toPascalCase(featureName))
	default:
		return fmt.Sprintf("services/%s_service.py", snake)
	}
}

// GetAPIPath returns API handler file path based on tech config
func GetAPIPath(tech TechConfig, featureName string) string {
	snake := toSnakeCase(featureName)
	switch tech.Backend {
	case "go", "golang":
		return fmt.Sprintf("internal/api/%s_handler.go", snake)
	case "python", "fastapi":
		return fmt.Sprintf("api/%s_router.py", snake)
	case "django":
		return fmt.Sprintf("api/views/%s.py", snake)
	case "node", "typescript", "express":
		return fmt.Sprintf("src/routes/%sRoutes.ts", toCamelCase(featureName))
	case "java", "spring":
		return fmt.Sprintf("src/main/java/controllers/%sController.java", toPascalCase(featureName))
	default:
		return fmt.Sprintf("api/%s_router.py", snake)
	}
}

// GetUIPath returns UI component file path based on tech config
func GetUIPath(tech TechConfig, componentName string) string {
	pascal := toPascalCase(componentName)
	switch tech.Frontend {
	case "react":
		return fmt.Sprintf("src/components/%s.tsx", pascal)
	case "vue":
		return fmt.Sprintf("src/components/%s.vue", pascal)
	case "angular":
		return fmt.Sprintf("src/app/components/%s/%s.component.ts", toSnakeCase(componentName), toSnakeCase(componentName))
	case "svelte":
		return fmt.Sprintf("src/lib/components/%s.svelte", pascal)
	default:
		return fmt.Sprintf("components/%s.tsx", pascal)
	}
}

// goType converts FDL type to Go type
func goType(fdlType string) string {
	// Parse base type
	baseType, _ := parseFieldType(fdlType)
	switch baseType {
	case "uuid":
		return "string"
	case "string":
		return "string"
	case "text":
		return "string"
	case "int":
		return "int"
	case "bigint":
		return "int64"
	case "float":
		return "float64"
	case "decimal":
		return "float64"
	case "boolean":
		return "bool"
	case "datetime":
		return "time.Time"
	case "date":
		return "time.Time"
	case "time":
		return "string"
	case "json":
		return "json.RawMessage"
	case "blob":
		return "[]byte"
	case "enum":
		return "string"
	default:
		return "interface{}"
	}
}

// pythonType converts FDL type to Python type hint
func pythonType(fdlType string) string {
	baseType, _ := parseFieldType(fdlType)
	switch baseType {
	case "uuid":
		return "str"
	case "string":
		return "str"
	case "text":
		return "str"
	case "int":
		return "int"
	case "bigint":
		return "int"
	case "float":
		return "float"
	case "decimal":
		return "Decimal"
	case "boolean":
		return "bool"
	case "datetime":
		return "datetime"
	case "date":
		return "date"
	case "time":
		return "time"
	case "json":
		return "dict"
	case "blob":
		return "bytes"
	case "enum":
		return "str"
	default:
		return "Any"
	}
}

// tsType converts FDL type to TypeScript type
func tsType(fdlType string) string {
	baseType, _ := parseFieldType(fdlType)
	switch baseType {
	case "uuid":
		return "string"
	case "string":
		return "string"
	case "text":
		return "string"
	case "int", "bigint":
		return "number"
	case "float", "decimal":
		return "number"
	case "boolean":
		return "boolean"
	case "datetime", "date", "time":
		return "Date"
	case "json":
		return "Record<string, any>"
	case "blob":
		return "Buffer"
	case "enum":
		return "string"
	default:
		return "any"
	}
}

// GenerateGoModel generates Go model skeleton
func GenerateGoModel(m *FDLModel) string {
	var buf bytes.Buffer
	buf.WriteString("package model\n\n")

	// Check if we need time import
	needsTime := false
	needsJSON := false
	for _, f := range m.Fields {
		baseType, _ := parseFieldType(f.Type)
		if baseType == "datetime" || baseType == "date" {
			needsTime = true
		}
		if baseType == "json" {
			needsJSON = true
		}
	}

	if needsTime || needsJSON {
		buf.WriteString("import (\n")
		if needsJSON {
			buf.WriteString("\t\"encoding/json\"\n")
		}
		if needsTime {
			buf.WriteString("\t\"time\"\n")
		}
		buf.WriteString(")\n\n")
	}

	buf.WriteString(fmt.Sprintf("// %s represents %s entity\n", toPascalCase(m.Name), m.Name))
	buf.WriteString(fmt.Sprintf("type %s struct {\n", toPascalCase(m.Name)))
	for _, field := range m.Fields {
		fieldName := toPascalCase(field.Name)
		fieldType := goType(field.Type)
		jsonTag := toSnakeCase(field.Name)
		buf.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", fieldName, fieldType, jsonTag))
	}
	buf.WriteString("}\n")
	return buf.String()
}

// GeneratePythonModel generates Python model skeleton
func GeneratePythonModel(m *FDLModel) string {
	var buf bytes.Buffer
	buf.WriteString("from dataclasses import dataclass\n")
	buf.WriteString("from typing import Optional, Any\n")
	buf.WriteString("from datetime import datetime, date, time\n")
	buf.WriteString("from decimal import Decimal\n\n")

	buf.WriteString("@dataclass\n")
	buf.WriteString(fmt.Sprintf("class %s:\n", toPascalCase(m.Name)))
	buf.WriteString(fmt.Sprintf("    \"\"\"Represents %s entity\"\"\"\n", m.Name))
	for _, field := range m.Fields {
		pyType := pythonType(field.Type)
		buf.WriteString(fmt.Sprintf("    %s: %s\n", toSnakeCase(field.Name), pyType))
	}
	buf.WriteString("\n")
	return buf.String()
}

// GenerateTSModel generates TypeScript model skeleton
func GenerateTSModel(m *FDLModel) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("// %s entity\n", m.Name))
	buf.WriteString(fmt.Sprintf("export interface %s {\n", toPascalCase(m.Name)))
	for _, field := range m.Fields {
		fieldName := toCamelCase(field.Name)
		fieldType := tsType(field.Type)
		buf.WriteString(fmt.Sprintf("  %s: %s;\n", fieldName, fieldType))
	}
	buf.WriteString("}\n")
	return buf.String()
}

// GenerateModelSkeleton generates model skeleton based on tech config
func GenerateModelSkeleton(tech TechConfig, m *FDLModel) string {
	switch tech.Backend {
	case "go", "golang":
		return GenerateGoModel(m)
	case "python", "fastapi", "django":
		return GeneratePythonModel(m)
	case "node", "typescript", "express":
		return GenerateTSModel(m)
	default:
		return GeneratePythonModel(m)
	}
}

// GenerateGoService generates Go service skeleton
func GenerateGoService(featureName string, services []FDLService) string {
	var buf bytes.Buffer
	buf.WriteString("package service\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"fmt\"\n")
	buf.WriteString("\n")
	buf.WriteString("\t\"parkjunwoo.com/claritask/internal/db\"\n")
	buf.WriteString(")\n\n")

	for _, svc := range services {
		funcName := toPascalCase(svc.Name)
		buf.WriteString(fmt.Sprintf("// %s %s\n", funcName, svc.Desc))
		buf.WriteString(fmt.Sprintf("func %s(database *db.DB) error {\n", funcName))
		buf.WriteString("\t// TODO: Implement service logic\n")
		for _, step := range svc.Steps {
			if s, ok := step.(string); ok {
				buf.WriteString(fmt.Sprintf("\t// Step: %s\n", s))
			}
		}
		buf.WriteString("\treturn fmt.Errorf(\"not implemented\")\n")
		buf.WriteString("}\n\n")
	}
	return buf.String()
}

// GeneratePythonService generates Python service skeleton
func GeneratePythonService(featureName string, services []FDLService) string {
	var buf bytes.Buffer
	buf.WriteString("from typing import Any, Optional\n\n\n")

	for _, svc := range services {
		funcName := toSnakeCase(svc.Name)
		buf.WriteString(fmt.Sprintf("def %s():\n", funcName))
		buf.WriteString(fmt.Sprintf("    \"\"\"%s\"\"\"\n", svc.Desc))
		buf.WriteString("    # TODO: Implement service logic\n")
		for _, step := range svc.Steps {
			if s, ok := step.(string); ok {
				buf.WriteString(fmt.Sprintf("    # Step: %s\n", s))
			}
		}
		buf.WriteString("    raise NotImplementedError()\n\n\n")
	}
	return buf.String()
}

// GenerateTSService generates TypeScript service skeleton
func GenerateTSService(featureName string, services []FDLService) string {
	var buf bytes.Buffer
	buf.WriteString("// Service functions\n\n")

	for _, svc := range services {
		funcName := toCamelCase(svc.Name)
		buf.WriteString(fmt.Sprintf("/**\n * %s\n */\n", svc.Desc))
		buf.WriteString(fmt.Sprintf("export async function %s(): Promise<void> {\n", funcName))
		buf.WriteString("  // TODO: Implement service logic\n")
		for _, step := range svc.Steps {
			if s, ok := step.(string); ok {
				buf.WriteString(fmt.Sprintf("  // Step: %s\n", s))
			}
		}
		buf.WriteString("  throw new Error('Not implemented');\n")
		buf.WriteString("}\n\n")
	}
	return buf.String()
}

// GenerateServiceSkeleton generates service skeleton based on tech config
func GenerateServiceSkeleton(tech TechConfig, featureName string, services []FDLService) string {
	switch tech.Backend {
	case "go", "golang":
		return GenerateGoService(featureName, services)
	case "python", "fastapi", "django":
		return GeneratePythonService(featureName, services)
	case "node", "typescript", "express":
		return GenerateTSService(featureName, services)
	default:
		return GeneratePythonService(featureName, services)
	}
}

// GenerateGoAPI generates Go API handler skeleton
func GenerateGoAPI(featureName string, apis []FDLAPI) string {
	var buf bytes.Buffer
	buf.WriteString("package api\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"net/http\"\n")
	buf.WriteString(")\n\n")

	for _, api := range apis {
		methodName := methodToHandler(api.Method, api.Path)
		buf.WriteString(fmt.Sprintf("// %s handles %s %s\n", methodName, api.Method, api.Path))
		buf.WriteString(fmt.Sprintf("func %s(w http.ResponseWriter, r *http.Request) {\n", methodName))
		buf.WriteString(fmt.Sprintf("\t// TODO: Implement %s %s\n", api.Method, api.Path))
		if api.Use != "" {
			buf.WriteString(fmt.Sprintf("\t// Uses: %s\n", api.Use))
		}
		buf.WriteString("\tw.WriteHeader(http.StatusNotImplemented)\n")
		buf.WriteString("}\n\n")
	}
	return buf.String()
}

// GeneratePythonAPI generates Python FastAPI router skeleton
func GeneratePythonAPI(featureName string, apis []FDLAPI) string {
	var buf bytes.Buffer
	buf.WriteString("from fastapi import APIRouter, HTTPException\n")
	buf.WriteString("from typing import Any\n\n")
	buf.WriteString(fmt.Sprintf("router = APIRouter(prefix=\"/api/%s\", tags=[\"%s\"])\n\n\n", toSnakeCase(featureName), featureName))

	for _, api := range apis {
		method := strings.ToLower(api.Method)
		buf.WriteString(fmt.Sprintf("@router.%s(\"%s\")\n", method, api.Path))
		funcName := pathToFuncName(api.Method, api.Path)
		buf.WriteString(fmt.Sprintf("async def %s():\n", funcName))
		if api.Summary != "" {
			buf.WriteString(fmt.Sprintf("    \"\"\"%s\"\"\"\n", api.Summary))
		}
		buf.WriteString(fmt.Sprintf("    # TODO: Implement %s %s\n", api.Method, api.Path))
		if api.Use != "" {
			buf.WriteString(fmt.Sprintf("    # Uses: %s\n", api.Use))
		}
		buf.WriteString("    raise HTTPException(status_code=501, detail=\"Not implemented\")\n\n\n")
	}
	return buf.String()
}

// GenerateTSAPI generates TypeScript Express router skeleton
func GenerateTSAPI(featureName string, apis []FDLAPI) string {
	var buf bytes.Buffer
	buf.WriteString("import { Router, Request, Response } from 'express';\n\n")
	buf.WriteString("const router = Router();\n\n")

	for _, api := range apis {
		method := strings.ToLower(api.Method)
		buf.WriteString(fmt.Sprintf("// %s\n", api.Summary))
		buf.WriteString(fmt.Sprintf("router.%s('%s', async (req: Request, res: Response) => {\n", method, api.Path))
		buf.WriteString(fmt.Sprintf("  // TODO: Implement %s %s\n", api.Method, api.Path))
		if api.Use != "" {
			buf.WriteString(fmt.Sprintf("  // Uses: %s\n", api.Use))
		}
		buf.WriteString("  res.status(501).json({ error: 'Not implemented' });\n")
		buf.WriteString("});\n\n")
	}

	buf.WriteString("export default router;\n")
	return buf.String()
}

// GenerateAPISkeleton generates API skeleton based on tech config
func GenerateAPISkeleton(tech TechConfig, featureName string, apis []FDLAPI) string {
	switch tech.Backend {
	case "go", "golang":
		return GenerateGoAPI(featureName, apis)
	case "python", "fastapi":
		return GeneratePythonAPI(featureName, apis)
	case "node", "typescript", "express":
		return GenerateTSAPI(featureName, apis)
	default:
		return GeneratePythonAPI(featureName, apis)
	}
}

// GenerateReactComponent generates React component skeleton
func GenerateReactComponent(ui *FDLUI) string {
	var buf bytes.Buffer
	componentName := toPascalCase(ui.Component)

	buf.WriteString("import React, { useState, useEffect } from 'react';\n\n")

	// Props interface
	if ui.Props != nil && len(ui.Props) > 0 {
		buf.WriteString(fmt.Sprintf("interface %sProps {\n", componentName))
		for name, spec := range ui.Props {
			propType := "any"
			required := ""
			if specMap, ok := spec.(map[string]interface{}); ok {
				if t, ok := specMap["type"].(string); ok {
					propType = tsType(t)
				}
				if opt, ok := specMap["optional"].(bool); ok && opt {
					required = "?"
				}
			} else if t, ok := spec.(string); ok {
				propType = tsType(t)
			}
			buf.WriteString(fmt.Sprintf("  %s%s: %s;\n", name, required, propType))
		}
		buf.WriteString("}\n\n")
	}

	// Component
	propsArg := ""
	if ui.Props != nil && len(ui.Props) > 0 {
		propsArg = fmt.Sprintf("props: %sProps", componentName)
	}
	buf.WriteString(fmt.Sprintf("export const %s: React.FC<%sProps> = (%s) => {\n", componentName, componentName, propsArg))

	// State
	if ui.State != nil {
		for _, state := range ui.State {
			if s, ok := state.(string); ok {
				// Parse "name: type = default" format
				parts := strings.SplitN(s, ":", 2)
				if len(parts) >= 1 {
					stateName := strings.TrimSpace(parts[0])
					defaultVal := "null"
					if strings.Contains(s, "=") {
						eqIdx := strings.Index(s, "=")
						defaultVal = strings.TrimSpace(s[eqIdx+1:])
					}
					buf.WriteString(fmt.Sprintf("  const [%s, set%s] = useState(%s);\n", stateName, toPascalCase(stateName), defaultVal))
				}
			}
		}
		buf.WriteString("\n")
	}

	// Init useEffect
	if ui.Init != nil && len(ui.Init) > 0 {
		buf.WriteString("  useEffect(() => {\n")
		buf.WriteString("    // TODO: Implement initialization\n")
		for _, init := range ui.Init {
			if s, ok := init.(string); ok {
				buf.WriteString(fmt.Sprintf("    // %s\n", s))
			}
		}
		buf.WriteString("  }, []);\n\n")
	}

	// Methods
	if ui.Methods != nil {
		for name, steps := range ui.Methods {
			buf.WriteString(fmt.Sprintf("  const %s = () => {\n", name))
			buf.WriteString("    // TODO: Implement method\n")
			if stepsList, ok := steps.([]interface{}); ok {
				for _, step := range stepsList {
					if s, ok := step.(string); ok {
						buf.WriteString(fmt.Sprintf("    // %s\n", s))
					}
				}
			}
			buf.WriteString("  };\n\n")
		}
	}

	// Render
	buf.WriteString("  return (\n")
	buf.WriteString("    <div>\n")
	buf.WriteString(fmt.Sprintf("      {/* TODO: Implement %s UI */}\n", componentName))
	buf.WriteString("    </div>\n")
	buf.WriteString("  );\n")
	buf.WriteString("};\n\n")
	buf.WriteString(fmt.Sprintf("export default %s;\n", componentName))

	return buf.String()
}

// GenerateVueComponent generates Vue component skeleton
func GenerateVueComponent(ui *FDLUI) string {
	var buf bytes.Buffer
	componentName := toPascalCase(ui.Component)

	buf.WriteString("<template>\n")
	buf.WriteString("  <div>\n")
	buf.WriteString(fmt.Sprintf("    <!-- TODO: Implement %s UI -->\n", componentName))
	buf.WriteString("  </div>\n")
	buf.WriteString("</template>\n\n")

	buf.WriteString("<script setup lang=\"ts\">\n")
	buf.WriteString("import { ref, onMounted } from 'vue';\n\n")

	// Props
	if ui.Props != nil && len(ui.Props) > 0 {
		buf.WriteString("const props = defineProps<{\n")
		for name, spec := range ui.Props {
			propType := "any"
			if specMap, ok := spec.(map[string]interface{}); ok {
				if t, ok := specMap["type"].(string); ok {
					propType = tsType(t)
				}
			} else if t, ok := spec.(string); ok {
				propType = tsType(t)
			}
			buf.WriteString(fmt.Sprintf("  %s?: %s;\n", name, propType))
		}
		buf.WriteString("}>;\n\n")
	}

	// State
	if ui.State != nil {
		for _, state := range ui.State {
			if s, ok := state.(string); ok {
				parts := strings.SplitN(s, ":", 2)
				if len(parts) >= 1 {
					stateName := strings.TrimSpace(parts[0])
					defaultVal := "null"
					if strings.Contains(s, "=") {
						eqIdx := strings.Index(s, "=")
						defaultVal = strings.TrimSpace(s[eqIdx+1:])
					}
					buf.WriteString(fmt.Sprintf("const %s = ref(%s);\n", stateName, defaultVal))
				}
			}
		}
		buf.WriteString("\n")
	}

	// Init
	if ui.Init != nil && len(ui.Init) > 0 {
		buf.WriteString("onMounted(() => {\n")
		buf.WriteString("  // TODO: Implement initialization\n")
		buf.WriteString("});\n\n")
	}

	// Methods
	if ui.Methods != nil {
		for name := range ui.Methods {
			buf.WriteString(fmt.Sprintf("const %s = () => {\n", name))
			buf.WriteString("  // TODO: Implement method\n")
			buf.WriteString("};\n\n")
		}
	}

	buf.WriteString("</script>\n\n")
	buf.WriteString("<style scoped>\n")
	buf.WriteString("</style>\n")

	return buf.String()
}

// GenerateUISkeleton generates UI skeleton based on tech config
func GenerateUISkeleton(tech TechConfig, ui *FDLUI) string {
	switch tech.Frontend {
	case "react":
		return GenerateReactComponent(ui)
	case "vue":
		return GenerateVueComponent(ui)
	default:
		return GenerateReactComponent(ui)
	}
}

// methodToHandler converts HTTP method and path to handler function name
func methodToHandler(method, path string) string {
	// Extract resource name from path
	parts := strings.Split(path, "/")
	resourceName := ""
	for _, p := range parts {
		if p != "" && !strings.HasPrefix(p, "{") && !strings.HasPrefix(p, ":") {
			resourceName = p
		}
	}
	return fmt.Sprintf("Handle%s%s", toPascalCase(strings.ToLower(method)), toPascalCase(resourceName))
}

// pathToFuncName converts HTTP method and path to Python function name
func pathToFuncName(method, path string) string {
	parts := strings.Split(path, "/")
	resourceName := ""
	for _, p := range parts {
		if p != "" && !strings.HasPrefix(p, "{") && !strings.HasPrefix(p, ":") {
			resourceName = p
		}
	}
	return fmt.Sprintf("%s_%s", strings.ToLower(method), toSnakeCase(resourceName))
}

// SkeletonFile represents a skeleton file to be generated
type SkeletonFile struct {
	Path    string `json:"path"`
	Layer   string `json:"layer"`
	Content string `json:"content,omitempty"`
}

// SkeletonResult represents the result of skeleton generation
type SkeletonResult struct {
	FeatureID int64          `json:"feature_id"`
	Files     []SkeletonFile `json:"files"`
	Errors    []string       `json:"errors,omitempty"`
}

// GenerateSkeletons generates all skeleton files for a feature from FDL
func GenerateSkeletons(database *db.DB, featureID int64, dryRun bool) (*SkeletonResult, error) {
	result := &SkeletonResult{
		FeatureID: featureID,
		Files:     []SkeletonFile{},
	}

	// Get feature
	feature, err := GetFeature(database, featureID)
	if err != nil {
		return nil, fmt.Errorf("get feature: %w", err)
	}

	if feature.FDL == "" {
		return nil, fmt.Errorf("no FDL defined for feature %d", featureID)
	}

	// Parse FDL
	spec, err := ParseFDL(feature.FDL)
	if err != nil {
		return nil, fmt.Errorf("parse FDL: %w", err)
	}

	// Get tech stack
	tech, _ := GetTech(database)
	techConfig := ParseTechConfig(tech)

	// Generate model skeletons
	for _, m := range spec.Models {
		path := GetModelPath(techConfig, m.Name)
		content := GenerateModelSkeleton(techConfig, &m)
		result.Files = append(result.Files, SkeletonFile{
			Path:    path,
			Layer:   "model",
			Content: content,
		})
	}

	// Generate service skeletons (one file for all services)
	if len(spec.Service) > 0 {
		path := GetServicePath(techConfig, "", spec.Feature)
		content := GenerateServiceSkeleton(techConfig, spec.Feature, spec.Service)
		result.Files = append(result.Files, SkeletonFile{
			Path:    path,
			Layer:   "service",
			Content: content,
		})
	}

	// Generate API skeletons (one file for all APIs)
	if len(spec.API) > 0 {
		path := GetAPIPath(techConfig, spec.Feature)
		content := GenerateAPISkeleton(techConfig, spec.Feature, spec.API)
		result.Files = append(result.Files, SkeletonFile{
			Path:    path,
			Layer:   "api",
			Content: content,
		})
	}

	// Generate UI skeletons
	for _, ui := range spec.UI {
		path := GetUIPath(techConfig, ui.Component)
		content := GenerateUISkeleton(techConfig, &ui)
		result.Files = append(result.Files, SkeletonFile{
			Path:    path,
			Layer:   "ui",
			Content: content,
		})
	}

	// If not dry-run, write files and save to DB
	if !dryRun {
		for _, f := range result.Files {
			// Create directory if needed
			dir := filepath.Dir(f.Path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("create dir %s: %v", dir, err))
				continue
			}

			// Write file
			if err := os.WriteFile(f.Path, []byte(f.Content), 0644); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("write file %s: %v", f.Path, err))
				continue
			}

			// Save to database
			if _, err := CreateSkeleton(database, featureID, f.Path, f.Layer); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("save skeleton %s: %v", f.Path, err))
			}
		}

		// Mark feature as skeleton generated
		feature.SkeletonGenerated = true
		UpdateFeature(database, feature)
	}

	return result, nil
}

// CreateSkeleton creates a skeleton record
func CreateSkeleton(database *db.DB, featureID int64, filePath, layer string) (int64, error) {
	checksum, _ := CalculateFileChecksum(filePath)
	now := db.TimeNow()

	result, err := database.Exec(
		`INSERT INTO skeletons (feature_id, file_path, layer, checksum, created_at) VALUES (?, ?, ?, ?, ?)`,
		featureID, filePath, layer, checksum, now,
	)
	if err != nil {
		return 0, fmt.Errorf("create skeleton: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}
	return id, nil
}

// GetSkeleton retrieves a skeleton by ID
func GetSkeleton(database *db.DB, id int64) (*model.Skeleton, error) {
	row := database.QueryRow(
		`SELECT id, feature_id, file_path, layer, checksum, created_at FROM skeletons WHERE id = ?`, id,
	)

	var s model.Skeleton
	var createdAt string
	err := row.Scan(&s.ID, &s.FeatureID, &s.FilePath, &s.Layer, &s.Checksum, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("get skeleton: %w", err)
	}
	s.CreatedAt, _ = db.ParseTime(createdAt)
	return &s, nil
}

// ListSkeletonsByFeature lists all skeletons for a feature
func ListSkeletonsByFeature(database *db.DB, featureID int64) ([]model.Skeleton, error) {
	rows, err := database.Query(
		`SELECT id, feature_id, file_path, layer, checksum, created_at FROM skeletons WHERE feature_id = ? ORDER BY id`,
		featureID,
	)
	if err != nil {
		return nil, fmt.Errorf("list skeletons: %w", err)
	}
	defer rows.Close()

	var skeletons []model.Skeleton
	for rows.Next() {
		var s model.Skeleton
		var createdAt string
		if err := rows.Scan(&s.ID, &s.FeatureID, &s.FilePath, &s.Layer, &s.Checksum, &createdAt); err != nil {
			return nil, fmt.Errorf("scan skeleton: %w", err)
		}
		s.CreatedAt, _ = db.ParseTime(createdAt)
		skeletons = append(skeletons, s)
	}
	return skeletons, nil
}

// DeleteSkeleton deletes a skeleton
func DeleteSkeleton(database *db.DB, id int64) error {
	_, err := database.Exec(`DELETE FROM skeletons WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete skeleton: %w", err)
	}
	return nil
}

// DeleteSkeletonsByFeature deletes all skeletons for a feature
func DeleteSkeletonsByFeature(database *db.DB, featureID int64) error {
	_, err := database.Exec(`DELETE FROM skeletons WHERE feature_id = ?`, featureID)
	if err != nil {
		return fmt.Errorf("delete skeletons by feature: %w", err)
	}
	return nil
}

// CalculateFileChecksum calculates SHA256 checksum of a file
func CalculateFileChecksum(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash), nil
}

// UpdateSkeletonChecksum updates skeleton checksum
func UpdateSkeletonChecksum(database *db.DB, id int64, checksum string) error {
	_, err := database.Exec(`UPDATE skeletons SET checksum = ? WHERE id = ?`, checksum, id)
	if err != nil {
		return fmt.Errorf("update skeleton checksum: %w", err)
	}
	return nil
}

// HasSkeletonChanged checks if skeleton file has changed
func HasSkeletonChanged(database *db.DB, id int64) (bool, error) {
	skeleton, err := GetSkeleton(database, id)
	if err != nil {
		return false, err
	}

	currentChecksum, err := CalculateFileChecksum(skeleton.FilePath)
	if err != nil {
		return true, nil // File doesn't exist, consider it changed
	}

	return currentChecksum != skeleton.Checksum, nil
}

// ReadSkeletonContent reads skeleton file content
func ReadSkeletonContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read skeleton file: %w", err)
	}
	return string(content), nil
}

// GetSkeletonAtLine reads code around a specific line
func GetSkeletonAtLine(filePath string, line, contextLines int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var lines []string

	startLine := line - contextLines
	if startLine < 1 {
		startLine = 1
	}
	endLine := line + contextLines

	for scanner.Scan() {
		lineNum++
		if lineNum >= startLine && lineNum <= endLine {
			lines = append(lines, fmt.Sprintf("%4d: %s", lineNum, scanner.Text()))
		}
		if lineNum > endLine {
			break
		}
	}

	return strings.Join(lines, "\n"), nil
}

// TODOLocation represents a TODO location in code
type TODOLocation struct {
	Line     int    `json:"line"`
	Function string `json:"function"`
	Content  string `json:"content"`
}

// ExtractTODOLocations extracts TODO locations from a file
func ExtractTODOLocations(filePath string) ([]TODOLocation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	var locations []TODOLocation
	scanner := bufio.NewScanner(file)
	lineNum := 0
	todoPattern := regexp.MustCompile(`(?i)#\s*TODO|//\s*TODO|/\*\s*TODO|\*\s*TODO`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if todoPattern.MatchString(line) {
			funcName, _ := GetFunctionAtLine(filePath, lineNum)
			locations = append(locations, TODOLocation{
				Line:     lineNum,
				Function: funcName,
				Content:  strings.TrimSpace(line),
			})
		}
	}

	return locations, nil
}

// GetFunctionAtLine extracts function name at a specific line
func GetFunctionAtLine(filePath string, targetLine int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// Patterns for different languages
	pythonDefPattern := regexp.MustCompile(`^\s*(async\s+)?def\s+(\w+)\s*\(`)
	goFuncPattern := regexp.MustCompile(`^\s*func\s+(?:\(\w+\s+\*?\w+\)\s+)?(\w+)\s*\(`)
	jsFuncPattern := regexp.MustCompile(`^\s*(async\s+)?function\s+(\w+)|^\s*(const|let|var)\s+(\w+)\s*=\s*(async\s+)?(\(|function)`)

	scanner := bufio.NewScanner(file)
	lineNum := 0
	lastFuncName := ""

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check Python def
		if matches := pythonDefPattern.FindStringSubmatch(line); len(matches) > 2 {
			lastFuncName = matches[2]
		}
		// Check Go func
		if matches := goFuncPattern.FindStringSubmatch(line); len(matches) > 1 {
			lastFuncName = matches[1]
		}
		// Check JavaScript/TypeScript function
		if matches := jsFuncPattern.FindStringSubmatch(line); len(matches) > 0 {
			for i := len(matches) - 1; i >= 0; i-- {
				if matches[i] != "" && matches[i] != "const" && matches[i] != "let" && matches[i] != "var" && matches[i] != "async" && matches[i] != "(" && matches[i] != "function" {
					lastFuncName = matches[i]
					break
				}
			}
		}

		if lineNum >= targetLine {
			break
		}
	}

	return lastFuncName, nil
}

// SkeletonGeneratorResult represents skeleton generator result
type SkeletonGeneratorResult struct {
	GeneratedFiles []GeneratedFile `json:"generated_files"`
	Errors         []string        `json:"errors,omitempty"`
}

// GeneratedFile represents a generated skeleton file
type GeneratedFile struct {
	Path     string `json:"path"`
	Layer    string `json:"layer"`
	Checksum string `json:"checksum"`
}

// RunSkeletonGenerator runs Python skeleton generator
func RunSkeletonGenerator(fdlPath, outputDir, backend, frontend string, force bool) (*SkeletonGeneratorResult, error) {
	args := []string{
		"scripts/skeleton_generator.py",
		"--fdl", fdlPath,
		"--output-dir", outputDir,
		"--backend", backend,
		"--frontend", frontend,
	}
	if force {
		args = append(args, "--force")
	}

	cmd := exec.Command("python3", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run skeleton generator: %w\nOutput: %s", err, string(output))
	}

	var result SkeletonGeneratorResult
	if err := json.Unmarshal(output, &result); err != nil {
		// If not JSON, assume success with file list in output
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				result.GeneratedFiles = append(result.GeneratedFiles, GeneratedFile{
					Path: strings.TrimSpace(line),
				})
			}
		}
	}

	return &result, nil
}

// RunSkeletonGeneratorDryRun returns list of files that would be generated
func RunSkeletonGeneratorDryRun(fdlPath, outputDir, backend, frontend string) (*SkeletonGeneratorResult, error) {
	args := []string{
		"scripts/skeleton_generator.py",
		"--fdl", fdlPath,
		"--output-dir", outputDir,
		"--backend", backend,
		"--frontend", frontend,
		"--dry-run",
	}

	cmd := exec.Command("python3", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run skeleton generator dry-run: %w\nOutput: %s", err, string(output))
	}

	var result SkeletonGeneratorResult
	if err := json.Unmarshal(output, &result); err != nil {
		// If not JSON, assume success with file list in output
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				result.GeneratedFiles = append(result.GeneratedFiles, GeneratedFile{
					Path: strings.TrimSpace(line),
				})
			}
		}
	}

	return &result, nil
}

// RunSkeletonGeneratorForFeature runs skeleton generator for a feature
func RunSkeletonGeneratorForFeature(database *db.DB, featureID int64, outputDir string, force bool) (*SkeletonGeneratorResult, error) {
	// Get feature
	feature, err := GetFeature(database, featureID)
	if err != nil {
		return nil, fmt.Errorf("get feature: %w", err)
	}

	if feature.FDL == "" {
		return nil, fmt.Errorf("no FDL defined for feature %d", featureID)
	}

	// Get tech stack
	tech, _ := GetTech(database)
	backend := "python"
	frontend := "none"
	if tech != nil {
		if b, ok := tech["backend"].(string); ok {
			backend = strings.ToLower(b)
		}
		if f, ok := tech["frontend"].(string); ok {
			frontend = strings.ToLower(f)
		}
	}

	// Write FDL to temp file
	tmpFile, err := os.CreateTemp("", "fdl-*.yaml")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(feature.FDL); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("write FDL: %w", err)
	}
	tmpFile.Close()

	// Run generator
	result, err := RunSkeletonGenerator(tmpFile.Name(), outputDir, backend, frontend, force)
	if err != nil {
		return nil, err
	}

	// Record skeletons in database
	for _, f := range result.GeneratedFiles {
		_, err := CreateSkeleton(database, featureID, f.Path, f.Layer)
		if err != nil {
			// Log warning but continue
			result.Errors = append(result.Errors, fmt.Sprintf("failed to record skeleton: %v", err))
		}
	}

	// Mark feature as skeleton generated
	feature.SkeletonGenerated = true
	UpdateFeature(database, feature)

	return result, nil
}

// GetSkeletonInfo builds SkeletonInfo for task pop response
func GetSkeletonInfo(database *db.DB, skeletonID int64, targetLine int) (*model.SkeletonInfo, error) {
	skeleton, err := GetSkeleton(database, skeletonID)
	if err != nil {
		return nil, err
	}

	content := ""
	if targetLine > 0 {
		content, _ = GetSkeletonAtLine(skeleton.FilePath, targetLine, 10)
	} else {
		content, _ = ReadSkeletonContent(skeleton.FilePath)
		// Truncate if too long
		if len(content) > 2000 {
			content = content[:2000] + "\n... (truncated)"
		}
	}

	line := targetLine
	if line == 0 {
		line = 1
	}

	return &model.SkeletonInfo{
		File:    skeleton.FilePath,
		Line:    line,
		Content: content,
	}, nil
}
