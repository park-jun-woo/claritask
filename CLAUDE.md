# Claritask Project

## Role

Go 언어 CLI 개발 전문가. Cobra 라이브러리와 SQLite를 사용한 고성능 CLI 애플리케이션 개발.

## Tech Stack

- **Language**: Go 1.21+
- **CLI Framework**: Cobra
- **Database**: SQLite (mattn/go-sqlite3)

## Summary Rule

### Trigger 조건
- 사용자가 명시적으로 '대화 저장해'라고 말하면 대화 요약 저장 절차를 시작한다.

### Summary Process
1. 대화 내용 전체를 `talks/` 일단 md로 저장한다.
2. 적당한 라인씩 검토하며 불필요한 내용은 제거한다.
3. 남은 내용 중에서 중요한 내용은 최대한 보존하고, 요약해도 괜찮은 덜 중요한 내용은 요약한다.

## Planning Rule

### Planning Trigger 조건
- 사용자가 명시적으로 '계획 수립해'라고 말하면 계획 수립 절차를 시작한다.

### Planning Process
1. 요구사항 명세서(`specs/*`)를 확인한다.
2. 파일 단위(.go)로 코딩해야할 TASK 목록을 나열한다.
3. `tasks/` 폴더에 TASK 단위로 작업지시를 요약해서 md 문서를 전부 생성한다. 파일명은 `TASK-DEV-<task-number>-<task-name>.md`로 한다.

## Development Rule

### Development Trigger 조건
- 사용자가 명시적으로 '개발 실행해'라고 말하면 개발 절차를 시작한다.

### Development Process
1. Task 목록(`tasks/*`)을 확인한다.
2. Task를 하나씩 실행한다.
3. Task를 완료하면 해당 Task md 파일을 `finished/`로 이동한다.

## Report Rule

### Report Trigger 조건
- 사용자가 명시적으로 '코드 파악해'라고 말하면 개발 절차를 시작한다.
- 
### Report Process
1. 코드 파일 목록(`cli/cmd/*, cli/internal/*, cli/test/*`)을 확인한다.
2. 파일을 하나씩 열어 분석하여 요약하고 `reports/<0000-00-00>/<파일명>-report.md`를 작성한다.
3. 모든 파일을 분석하면 최종 전체 보고서(`reports/<0000-00-00>-report.md`)를 작성한다.

## Project Structure

```
claritask/
├── cli/                     # Go CLI 소스코드
│   ├── cmd/claritask/       # 메인 진입점
│   │   └── main.go
│   ├── internal/
│   │   ├── cmd/             # Cobra 명령어
│   │   ├── db/              # 데이터베이스 레이어
│   │   ├── model/           # 데이터 모델
│   │   └── service/         # 비즈니스 로직
│   ├── test/
│   ├── scripts/
│   ├── go.mod
│   ├── go.sum
│   └── Makefile
├── vscode-extension/        # VSCode Extension 소스코드
├── specs/                   # 요구사항 명세서
│   ├── Claritask.md
│   ├── CLI/
│   ├── DB/
│   ├── FDL/
│   ├── TTY/
│   └── VSCode/
├── tasks/                   # 해야할 Task 문서 폴더
├── finished/                # 완료한 Task 문서 폴더
└── talks/                   # 요약 저장한 대화내역
```

## Coding Conventions

### Go Style
- `gofmt` 스타일 준수
- 에러는 즉시 처리 (early return)
- 인터페이스는 사용처에서 정의
- 패키지명은 단수형, 소문자

### Naming
- 파일명: snake_case (`task_service.go`)
- 변수/함수: camelCase (`taskService`)
- 타입/상수: PascalCase (`TaskService`)
- 약어는 대문자 유지 (`ID`, `URL`, `JSON`)

### Error Handling
```go
if err != nil {
    return fmt.Errorf("context: %w", err)
}
```

### JSON Output
모든 CLI 출력은 JSON 형식:
```go
type Response struct {
    Success bool   `json:"success"`
    Error   string `json:"error,omitempty"`
    // ... data fields
}
```

## Database

- 위치: `.claritask/db.clt`
- 자동 생성: 첫 실행 시 또는 'clari init <project>'로 생성시
- 마이그레이션: 앱 시작 시 자동

### Tables
- `projects` - 프로젝트 (목록, 영문숫자 ID 지정)
- `phases` - 작업 단계 (auto increment)
- `tasks` - 실행 단위 (auto increment)
- `context` - 프로젝트 컨텍스트 (프로젝트별)
- `tech` - 기술 스택 (프로젝트별)
- `design` - 설계 결정 (프로젝트별)
- `state` - 현재 상태 (key-value)
- `memos` - 메모 (scope 기반)

### Cobra Command Pattern
```go
var exampleCmd = &cobra.Command{
    Use:   "example",
    Short: "Short description",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Parse input
        // 2. Open DB
        // 3. Execute logic
        // 4. Output JSON
        return nil
    },
}
```

### DB Connection Pattern
```go
func getDB() (*db.DB, error) {
    home, _ := os.UserHomeDir()
    dbPath := filepath.Join(home, ".claritask", "db")
    return db.Open(dbPath)
}
```

### JSON Input Handling
```go
func parseJSON(jsonStr string, v interface{}) error {
    return json.Unmarshal([]byte(jsonStr), v)
}
```

### JSON Output
```go
func outputJSON(v interface{}) {
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    enc.Encode(v)
}
```

### ID 규칙
- **Project**: 영문 소문자, 숫자, 하이픈(-), 언더스코어(_) - 예: `blog`, `api-server`
- **Phase/Task**: 정수 (auto increment)

## References
- specs/* - 전체 요구사항 명세서
- tasks/* - 구현 계획 Task 파일들
- talks/* - 사용자와 클로드 코드의 과거 대화 내용

## 버전 표기 규칙
- vX.X.N 형식이며 테스트하며 수정할때 N 숫자만 올려라. 10이 넘어도 vX.X.11로 표기하라.