# TASK-DEV-095: Required 필드 검증 확인

## 개요

specs에 정의된 필수 필드가 올바르게 검증되는지 확인

## 스펙 요구사항 (06-Settings.md)

### 필수 필드 정의

**Context**:
- `project_name` (필수)
- `description` (필수)

**Tech**:
- `backend` (필수)
- `frontend` (필수)
- `database` (필수)

**Design**:
- `architecture` (필수)
- `auth_method` (필수)
- `api_style` (필수)

## 작업 내용

### 1. project_service.go CheckRequired 확인

```go
func CheckRequired(db *db.DB) (bool, []map[string]interface{}, error) {
    missing := []map[string]interface{}{}

    // Context 필수 필드
    context, _ := GetContext(db)
    if context == nil {
        missing = append(missing, map[string]interface{}{
            "field":  "context.project_name",
            "prompt": "프로젝트 이름을 입력하세요",
        })
        missing = append(missing, map[string]interface{}{
            "field":  "context.description",
            "prompt": "프로젝트 설명을 입력하세요",
        })
    } else {
        var data map[string]interface{}
        json.Unmarshal([]byte(context.Data), &data)

        if _, ok := data["project_name"]; !ok {
            missing = append(missing, map[string]interface{}{
                "field":  "context.project_name",
                "prompt": "프로젝트 이름을 입력하세요",
            })
        }
        if _, ok := data["description"]; !ok {
            missing = append(missing, map[string]interface{}{
                "field":  "context.description",
                "prompt": "프로젝트 설명을 입력하세요",
            })
        }
    }

    // Tech 필수 필드
    tech, _ := GetTech(db)
    if tech == nil {
        // backend, frontend, database 모두 누락
        missing = append(missing, map[string]interface{}{
            "field":   "tech.backend",
            "prompt":  "백엔드 기술을 선택하세요",
            "options": []string{"go", "python", "node", "java"},
        })
        // ... frontend, database
    } else {
        var data map[string]interface{}
        json.Unmarshal([]byte(tech.Data), &data)
        // 각 필드 확인
    }

    // Design 필수 필드
    design, _ := GetDesign(db)
    if design == nil {
        // architecture, auth_method, api_style 모두 누락
    } else {
        // 각 필드 확인
    }

    return len(missing) == 0, missing, nil
}
```

### 2. 응답 스키마 확인

```json
{
  "success": true,
  "ready": false,
  "missing_required": [
    {"field": "context.project_name", "prompt": "프로젝트 이름을 입력하세요"},
    {"field": "tech.backend", "prompt": "백엔드 기술을 선택하세요", "options": ["go", "python", "node", "java"]},
    {"field": "design.architecture", "prompt": "아키텍처를 선택하세요", "options": ["monolithic", "microservices", "serverless"]}
  ],
  "total_missing": 3
}
```

## 완료 조건

- [ ] CheckRequired 함수에서 모든 필수 필드 검사
- [ ] 누락 필드에 prompt 메시지 포함
- [ ] 선택형 필드에 options 포함
- [ ] 응답에 total_missing 포함
- [ ] 테스트 작성
