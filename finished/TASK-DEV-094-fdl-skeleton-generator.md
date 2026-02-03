# TASK-DEV-094: FDL 스켈레톤 생성 개선

## 개요

FDL에서 스켈레톤 코드 생성 시 tech 설정 기반으로 적절한 파일 생성

## 스펙 요구사항 (09-FDL.md)

```json
{
  "generated_files": [
    {"path": "models/user.py", "layer": "model"},
    {"path": "services/user_auth_service.py", "layer": "service"},
    {"path": "api/user_auth_handler.py", "layer": "api"},
    {"path": "components/LoginForm.tsx", "layer": "ui"}
  ]
}
```

## 현재 상태

skeleton_service.go 존재하지만 기본적인 파일 경로만 생성

## 작업 내용

### 1. Tech 기반 파일 경로 결정

```go
type TechConfig struct {
    Backend  string  // go, python, node, java
    Frontend string  // react, vue, angular, svelte
}

func getModelPath(tech TechConfig, modelName string) string {
    switch tech.Backend {
    case "go":
        return fmt.Sprintf("internal/model/%s.go", toSnakeCase(modelName))
    case "python":
        return fmt.Sprintf("models/%s.py", toSnakeCase(modelName))
    case "node":
        return fmt.Sprintf("src/models/%s.ts", toCamelCase(modelName))
    default:
        return fmt.Sprintf("models/%s.go", toSnakeCase(modelName))
    }
}

func getServicePath(tech TechConfig, serviceName string) string {
    // 유사하게 구현
}

func getAPIPath(tech TechConfig, handlerName string) string {
    // 유사하게 구현
}

func getUIPath(tech TechConfig, componentName string) string {
    switch tech.Frontend {
    case "react":
        return fmt.Sprintf("src/components/%s.tsx", toPascalCase(componentName))
    case "vue":
        return fmt.Sprintf("src/components/%s.vue", toPascalCase(componentName))
    default:
        return fmt.Sprintf("components/%s.tsx", toPascalCase(componentName))
    }
}
```

### 2. 스켈레톤 내용 생성

```go
func generateModelSkeleton(tech TechConfig, model *FDLModel) string {
    switch tech.Backend {
    case "go":
        return generateGoModel(model)
    case "python":
        return generatePythonModel(model)
    default:
        return generateGoModel(model)
    }
}

func generateGoModel(model *FDLModel) string {
    var buf bytes.Buffer
    buf.WriteString(fmt.Sprintf("package model\n\n"))
    buf.WriteString(fmt.Sprintf("type %s struct {\n", toPascalCase(model.Name)))
    for _, field := range model.Fields {
        buf.WriteString(fmt.Sprintf("\t%s %s\n", toPascalCase(field.Name), goType(field.Type)))
    }
    buf.WriteString("}\n")
    return buf.String()
}
```

### 3. GenerateSkeleton 수정

```go
func GenerateSkeleton(db *db.DB, featureID int64, dryRun bool) (*SkeletonResult, error) {
    // 1. Feature의 FDL 가져오기
    feature, err := GetFeature(db, featureID)
    if err != nil {
        return nil, err
    }

    // 2. FDL 파싱
    spec, err := ParseFDL(feature.FDL)
    if err != nil {
        return nil, err
    }

    // 3. Tech 설정 가져오기
    tech, err := GetTech(db)
    if err != nil {
        return nil, err
    }
    techConfig := parseTechConfig(tech)

    // 4. 각 레이어별 스켈레톤 생성
    files := []SkeletonFile{}

    for _, model := range spec.Models {
        path := getModelPath(techConfig, model.Name)
        content := generateModelSkeleton(techConfig, model)
        files = append(files, SkeletonFile{Path: path, Layer: "model", Content: content})
    }

    // ... service, api, ui 동일

    // 5. dry-run이 아니면 파일 생성
    if !dryRun {
        for _, f := range files {
            writeSkeletonFile(f)
            saveSkeletonToDB(db, featureID, f)
        }
    }

    return &SkeletonResult{Files: files}, nil
}
```

## 완료 조건

- [ ] Tech 기반 파일 경로 결정 함수
- [ ] 각 언어별 스켈레톤 템플릿
- [ ] Model 스켈레톤 생성
- [ ] Service 스켈레톤 생성
- [ ] API 스켈레톤 생성
- [ ] UI 스켈레톤 생성
- [ ] DB에 skeleton 정보 저장
- [ ] 테스트 작성
