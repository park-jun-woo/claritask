# TASK-DEV-105: Feature Add TTY Handover 구현

## 목표
`clari feature add` 실행 시 Claude Code를 TTY Handover로 호출하여 FDL 생성

## 변경 파일
- `cli/internal/service/feature_service.go`
- `cli/internal/service/tty_service.go`
- `cli/internal/cmd/feature.go`

## 작업 내용

### 1. feature_service.go 수정
- `CreateFeature` 함수 수정: DB 레코드만 생성
- 새 함수 `GenerateFDLPrompt(featureID int64) string` 추가

### 2. feature.go 수정
- `runFeatureAdd` 수정:
  1. DB에 Feature 생성
  2. FDL 생성 프롬프트 준비
  3. TTY Handover로 Claude Code 호출
  4. Claude Code 완료 후 FDL 파일 저장

```go
func runFeatureAdd(cmd *cobra.Command, args []string) error {
    // 1. Parse input
    // 2. Create feature in DB
    // 3. Prepare prompt for FDL generation
    // 4. Call TTY Handover
    prompt := service.GenerateFDLPrompt(featureID, input.Name, input.Description)
    service.HandoverToClaudeCode(prompt, func(result string) {
        // Save FDL to features/<name>.fdl.yaml
        // Update DB with FDL content
    })
}
```

### 3. FDL 생성 프롬프트 템플릿
```go
const fdlGenerationPrompt = `다음 Feature에 대한 FDL YAML을 생성해주세요.

Feature Name: %s
Description: %s

FDL 스펙:
- feature, version, description 필수
- layers: data, logic, interface, presentation
- 파일 경로: features/%s.fdl.yaml

생성된 FDL을 위 경로에 저장해주세요.`
```

## 테스트
- `clari feature add '{"name":"test","description":"테스트"}'` 실행
- Claude Code가 열리고 FDL 생성 요청 표시 확인
- FDL 생성 후 `features/test.fdl.yaml` 파일 생성 확인
