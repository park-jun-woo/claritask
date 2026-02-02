# TASK-DEV-009: Root 명령어

## 목적
Cobra CLI Root 명령어 및 공통 유틸리티 구현

## 구현 파일
- `internal/cmd/root.go` - Root 명령어 및 공통 함수

## 상세 요구사항

### 1. Root Command
```go
var rootCmd = &cobra.Command{
    Use:   "talos",
    Short: "Task And LLM Operating System",
    Long:  "Claude Code를 위한 장시간 자동 실행 시스템",
}

func Execute() error {
    return rootCmd.Execute()
}
```

### 2. 공통 유틸리티
```go
// getDB - 현재 디렉토리의 .talos/db 연결
func getDB() (*db.DB, error) {
    cwd, _ := os.Getwd()
    dbPath := filepath.Join(cwd, ".talos", "db")
    return db.Open(dbPath)
}

// outputJSON - JSON 응답 출력
func outputJSON(v interface{}) {
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    enc.Encode(v)
}

// outputError - 에러 JSON 출력
func outputError(err error) {
    outputJSON(map[string]interface{}{
        "success": false,
        "error":   err.Error(),
    })
}

// parseJSON - JSON 문자열 파싱
func parseJSON(jsonStr string, v interface{}) error {
    return json.Unmarshal([]byte(jsonStr), v)
}
```

### 3. 서브커맨드 등록
```go
func init() {
    rootCmd.AddCommand(initCmd)
    rootCmd.AddCommand(projectCmd)
    rootCmd.AddCommand(contextCmd)
    rootCmd.AddCommand(techCmd)
    rootCmd.AddCommand(designCmd)
    rootCmd.AddCommand(requiredCmd)
    rootCmd.AddCommand(phaseCmd)
    rootCmd.AddCommand(taskCmd)
    rootCmd.AddCommand(memoCmd)
}
```

## 의존성
- 선행 Task: TASK-DEV-002 (database)
- 필요 패키지: github.com/spf13/cobra, encoding/json, os, path/filepath

## 완료 기준
- [ ] Root command 정의
- [ ] Execute() 함수 구현
- [ ] 공통 유틸리티 함수 구현 (getDB, outputJSON, outputError, parseJSON)
- [ ] 서브커맨드 등록 구조 준비
