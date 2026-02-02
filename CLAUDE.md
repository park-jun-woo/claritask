# TALOS Project

## Role

Go 언어 CLI 개발 전문가. Cobra 라이브러리와 SQLite를 사용한 고성능 CLI 애플리케이션 개발.

## Tech Stack

- **Language**: Go 1.21+
- **CLI Framework**: Cobra
- **Database**: SQLite (mattn/go-sqlite3)

## Planning Rule

### Trigger 조건
- 사용자가 명시적으로 '계획 수립해'라고 말하면 계획 수립 절차를 시작한다.

### Planning 작업을 수행할 전문가
- 개발 계획수립 전문가(`/planner`)가 백그라운드에서 실행한다.

### Planning Process
1. 요구사항 명세서(`specs/*`)를 확인한다.
2. 파일 단위(.go)로 코딩해야할 TASK 목록을 나열한다.
3. `tasks/` 폴더에 TASK 단위로 작업지시를 요약해서 md 문서를 전부 생성한다. 파일명은 `TASK-DEV-<task-number>-<task-name>.md`로 한다.

## Development Rule

### Trigger 조건
- 사용자가 명시적으로 '개발 실행해'라고 말하면 개발 절차를 시작한다.

### Development 작업을 수행할 전문가
- 개발 전문가(`/developer`)가 백그라운드에서 실행한다.

### Development Process
1. Task 목록(`tasks/*`)을 확인한다.
2. Task를 하나씩 실행한다.
3. Task를 완료하면 해당 Task md 파일을 `finished/`로 이동한다.
4. '/clear'하여 컨텍스트를 초기화한다.

## Testing Rule

### Trigger 조건
- 사용자가 명시적으로 '테스트 실행해'라고 말하면 테스트 절차를 시작한다.

### Testing 작업을 수행할 전문가
- 테스트 전문가(`/tester`)가 백그라운드에서 실행한다.

### Testing Process
1. 완료한 개발 업무 목록(`finished/TASK-DEV-*.md`)을 확인한다.
2. `tasks/` 폴더에 TASK 단위로 테스트 작업지시를 요약해서 md 문서를 전부 생성한다. 파일명은 `TASK-TEST-<task-number>-<task-name>.md`로 한다. test 코드는 `test/`폴더에 생성하는 것으로 한다.
3. Task 목록(`tasks/*`)을 확인한다.
4. Task를 하나씩 실행한다.
5. Task를 완료하면 해당 Task md 파일을 `finished/`로 이동한다.
6. '/clear'하여 컨텍스트를 초기화한다.

## Project Structure

```
talos/
├── cmd/talos/           # 메인 진입점
│   ├── main.go
├── internal/
│   ├── cmd/             # Cobra 명령어
│   ├── db/              # 데이터베이스 레이어
│   ├── model/           # 데이터 모델
│   └── service/         # 비즈니스 로직
├── test/
├── go.mod
├── specs/
│   ├── Talos.md         # 프로젝트 스펙
│   └── Commands.md      # 명령어 레퍼런스
├── tasks/               # 해야할 Task 문서 폴더
└── finished/            # 완료한 Task 문서 폴더
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

- 위치: `.talos/db`
- 자동 생성: 첫 실행 시 또는 'talos init <project>'로 생성시
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
    dbPath := filepath.Join(home, ".talos", "db")
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