# TASK-DEV-106: Feature FDL 재생성 명령어

## 목표
`clari feature fdl <id>` 명령어 구현 - 기존 Feature의 FDL 재생성

## 변경 파일
- `cli/internal/cmd/feature.go`

## 작업 내용

### 1. featureFdlCmd 추가
```go
var featureFdlCmd = &cobra.Command{
    Use:   "fdl <id>",
    Short: "Regenerate FDL for a feature using Claude Code",
    Args:  cobra.ExactArgs(1),
    RunE:  runFeatureFdl,
}

func init() {
    featureCmd.AddCommand(featureFdlCmd)
}
```

### 2. runFeatureFdl 함수
```go
func runFeatureFdl(cmd *cobra.Command, args []string) error {
    // 1. Get feature from DB
    // 2. Prepare FDL regeneration prompt
    // 3. Call TTY Handover
    // 4. Update DB after completion
}
```

## 테스트
- `clari feature fdl 1` 실행
- Claude Code가 열리고 기존 Feature 정보와 함께 FDL 재생성 요청 표시
