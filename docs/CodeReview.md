# Claribot 프로젝트 코드 리뷰

> 작성일: 2025-02-05

## 프로젝트 개요
- **총 Go 코드**: 7,668줄
- **모듈**: bot (서비스 데몬), cli (클라이언트), vsx (VSCode 확장)
- **Go 버전**: 1.21+
- **주요 의존성**: telegram-bot-api, sqlite3, pty, cron/v3

---

## 1. 코드 구조 및 아키텍처

### 강점
- **계층별 분리**: internal/ 패키지로 명확한 모듈화
  - `config/` - 설정 관리
  - `db/` - 데이터베이스 추상화
  - `handler/` - HTTP 라우팅
  - `project/`, `task/`, `message/`, `schedule/`, `edge/` - 도메인 기능
  - `pkg/` - 재사용 가능한 라이브러리 (claude, logger, telegram, etc.)

- **멀티 DB 전략**: 전역 DB와 로컬 DB 분리로 프로젝트별 독립성 확보
  - 전역: 프로젝트 목록, 스케줄, 메시지
  - 로컬: 작업, 엣지 (git 관리 가능)

- **PRAGMA 최적화**: SQLite 설정 (foreign_keys, WAL, busy_timeout) 적용

### 개선점
- **테스트 커버리지**: 8개 테스트 파일만 있음 (테스트 부족)
- **인터페이스 활용**: 현재는 구체적 구조체만 사용, 의존성 주입 부재

---

## 2. 코딩 컨벤션 준수 여부

### 준수 사항
- **파일명**: snake_case 준수 (`config.go`, `task_service.go`)
- **변수/함수**: camelCase 준수
- **타입/상수**: PascalCase 준수
- **에러 처리**: Early return 패턴 일관성 있음
- **패키지명**: 단수형 소문자 준수

### 미준수 사항
- **패키지 주석**: 많은 파일에서 패키지 주석 부재
- **공개 함수 문서화**: 많은 exported 함수에 doc comment 부재
- **상수 문서화**: 상수들이 충분히 문서화되지 않음

---

## 3. 에러 처리 방식

### 긍정적 패턴
- **Early Return**: 명확한 early return 원칙 준수
- **Defer Close**: 데이터베이스, 파일 리소스 proper cleanup
- **구조화된 에러**: `errors` 패키지에서 에러 코드 정의

### 문제점
- **에러 추가 정보 부족**: 많은 곳에서 로그 기록 없이 에러 반환
- **에러 래핑 미흡**: `fmt.Errorf(...%w...)` 사용하지만, 구조화된 에러 패키지의 Wrap 함수 미사용

---

## 4. 잠재적 버그 및 개선점

### CRITICAL BUG

#### Bug #1: Type Conversion Error (scheduler.go:198)
```go
errorText = "exit code: " + string(rune(claudeResult.ExitCode))
```
**문제**: `claudeResult.ExitCode` (int)를 `rune`으로 변환 후 `string()`으로 변환하면 문자가 됨
- ExitCode = 1 → rune(1) = '\x01' → string('\x01') = "\x01"

**수정**:
```go
errorText = fmt.Sprintf("exit code: %d", claudeResult.ExitCode)
```

---

### MAJOR ISSUES

#### Issue #2: Race Condition in Telegram Handler
```go
// tghandler.go
type Handler struct {
    pendingContext map[int64]string  // No mutex!
}
```

**문제**: 여러 Telegram 메시지가 동시에 도착할 때 race condition 발생

**수정**:
```go
type Handler struct {
    pendingContext map[int64]string
    mu             sync.RWMutex  // Add mutex
}
```

---

#### Issue #3: Rollback Logic Incomplete (project/add.go)
```go
if err := localDB.MigrateLocal(); err != nil {
    globalDB.Exec(`DELETE FROM projects WHERE id = ?`, id)  // 에러 확인 없음
    return types.Result{...}
}
```

**수정**: globalDB.Exec() 에러 처리 추가

---

### MEDIUM ISSUES

#### Issue #4: Missing Input Validation
CLI 명령어 인자 검증 부재 - 안전한 인덱스 접근 필요

#### Issue #5: Inconsistent Error Message Display
한글/영문 혼재 메시지

#### Issue #6: Timeout Bug in Claude Session
Hard-coded 2초 타임아웃이 manager 설정 무시

#### Issue #7: YAML Unmarshal Error Ignored
```go
yaml.Unmarshal(data, &cfg)  // err 체크 없음
```

#### Issue #8: No Pagination Bounds Check
page, pageSize 음수/0 체크 없음

---

## 5. 보안 관련 이슈

### 좋은 점
- **SQL Injection 방지**: 모든 쿼리에서 매개변수화된 쿼리 사용
- **Claude 권한 제한**: `--dangerously-skip-permissions` 플래그 사용
- **경로 검증**: `filepath.Abs()` 사용

### 보안 우려사항

#### Security #1: Insecure Default Configuration
포트 충돌 시 실패 처리 필요

#### Security #2: Insufficient Input Validation
프로젝트 ID에 특수문자나 매우 긴 이름 체크 없음

**수정**:
```go
func isValidProjectID(id string) bool {
    if len(id) == 0 || len(id) > 100 {
        return false
    }
    for _, ch := range id {
        if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
            return false
        }
    }
    return true
}
```

#### Security #3: Telegram Token in Logs
설정 로그 시 민감 정보 마스킹 필요

#### Security #4: Working Directory Not Validated
임의의 경로에서 Claude Code 실행 가능 - 절대 경로 검증 필요

---

## 6. 테스트 커버리지

### 현황
- **테스트 파일**: 8개
  - `config_test.go`, `task_test.go`, `related_test.go`
  - `claude_test.go`, `logger_test.go`, `telegram_test.go`
  - `render_test.go`, `errors_test.go`

### 문제점
- `handler/router.go` 테스트 없음 (핵심 라우팅)
- `project/*`, `task/*`, `message/*` 실제 기능 테스트 없음
- `db/*` 마이그레이션 테스트 없음
- `schedule/*` 스케줄 실행 테스트 없음

---

## 7. 문서화 상태

### 있는 것
- `CLAUDE.md` - 프로젝트 개요, 기술 스택
- `docs/Claribot.md` - 아키텍처, DB 스키마
- `docs/Task.md` - Task 시스템 설계
- `docs/Schedule.md` - Schedule 시스템 설계
- `README.md` - 기능, 아키텍처, 설치 가이드

### 부족한 것
- API 문서화 (HTTP API 명세 없음)
- 환경 변수 문서화
- 에러 코드 참고
- 배포 가이드 (systemd, Docker, Kubernetes)

---

## 종합 평가

| 항목 | 등급 | 설명 |
|------|------|------|
| **아키텍처** | A- | 계층 분리 명확, DB 전략 좋음, 테스트 부족 |
| **코딩 스타일** | A | 컨벤션 준수 일관성, 문서화 미흡 |
| **에러 처리** | B+ | Early return 좋음, 에러 정보 추가 권장 |
| **보안** | B+ | SQL injection 방지, 입력 검증 강화 필요 |
| **테스트** | D | 8개 파일만 있음, 핵심 기능 테스트 부재 |
| **문서화** | B- | 설계 문서 좋음, API/환경변수 문서 부재 |
| **성능** | A | SQLite WAL, 세마포어 기반 동시성 제어 |
| **유지보수성** | B | 모듈화 좋음, 로깅 일관성 부족 |

---

## 우선순위 개선 항목

### 즉시 수정 필요
1. **Bug #1**: Type conversion error in scheduler.go:198
2. **Issue #2**: Race condition in pendingContext (add mutex)
3. **Issue #4**: Input validation in CLI argument handling

### 주간 내 수정 권장
4. **Issue #3**: Incomplete rollback logic
5. **Issue #6**: Hard-coded timeout in claude.go
6. **Issue #7**: YAML unmarshal error handling
7. **Security #2**: Project ID validation

### 월간 내 개선 권장
8. **테스트 추가**: Handler, Project, Task, Message 모듈
9. **문서화**: API 명세, 환경 변수, 에러 코드
10. **로깅 일관성**: 한글/영문 통일, 에러 컨텍스트 추가
11. **설정 검증**: Telegram token, 포트 등 보안 개선

---

## 결론

Claribot은 **잘 설계된 아키텍처**를 가진 프로젝트이며, **코딩 스타일도 일관성 있게 유지**되고 있습니다. 다만, **테스트 커버리지 부족**, **일부 race condition**, **타입 변환 버그**, **문서화 미흡** 등이 개선이 필요합니다.

특히 **Bug #1(타입 변환), Issue #2(race condition), Issue #4(입력 검증)**은 즉시 수정을 권장하며, 이후 **테스트 추가**와 **문서화 강화**를 통해 프로덕션 레벨의 품질을 확보할 수 있을 것으로 예상됩니다.
