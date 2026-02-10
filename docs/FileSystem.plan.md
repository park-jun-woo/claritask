# FileSystem Architecture - Development Plan

> docs/FileSystem.md 구현 계획

---

## Phase 1: 기반 (파일 I/O + Frontmatter)

DB는 그대로 두고, 파일 읽기/쓰기 레이어만 추가한다. 기존 기능 깨지지 않음.

### 1-1. 디렉토리 구조 생성

**파일**: `bot/internal/task/file.go` (신규)

- `.claribot/tasks/` 디렉토리 자동 생성 (ensureDir)
- 파일 경로 헬퍼: `taskFilePath(projectPath, id)` → `.claribot/tasks/{id}.md`
- plan/report 경로: `{id}.plan.md`, `{id}.report.md`

### 1-2. Frontmatter 파서

**파일**: `bot/internal/task/frontmatter.go` (신규)

```go
type Frontmatter struct {
    Status   string `yaml:"status"`
    Parent   *int   `yaml:"parent,omitempty"`
    Priority int    `yaml:"priority,omitempty"`
}

// ParseTaskFile: .md 파일 → Frontmatter + title(H1) + body
func ParseTaskFile(filePath string) (Frontmatter, string, string, error)

// WriteTaskFile: Frontmatter + title + body → .md 파일
func WriteTaskFile(filePath string, fm Frontmatter, title, body string) error
```

의존성: `gopkg.in/yaml.v3` (이미 config에서 사용 중)

### 1-3. 포맷 검증

**파일**: `bot/internal/task/validate.go` (신규)

```go
// ValidateTaskFile: 포맷 룰 검증
// - frontmatter 존재 여부
// - status 필수, 허용 값만
// - H1 제목 존재
// - 본문 비어있지 않음
func ValidateTaskFile(filePath string) error
```

---

## Phase 2: 이중 쓰기 (DB + 파일 동시)

기존 DB 로직 유지하면서, 파일도 함께 생성. 안전한 전환 준비.

### 2-1. task add → 파일 동시 생성

**파일**: `bot/internal/task/add.go` 수정

- 기존 DB INSERT 유지
- INSERT 후 `WriteTaskFile()` 호출하여 `tasks/{id}.md` 생성
- 실패 시 로그만 (DB가 아직 primary)

### 2-2. plan 완료 → plan 파일 동시 생성

**파일**: `bot/internal/task/plan.go` 수정

- 기존 `UPDATE tasks SET plan = ?` 유지
- plan 내용을 `tasks/{id}.plan.md`에도 저장

### 2-3. run 완료 → report 파일 동시 생성

**파일**: `bot/internal/task/run.go` 수정

- 기존 `UPDATE tasks SET report = ?` 유지
- report 내용을 `tasks/{id}.report.md`에도 저장

### 2-4. task get → 파일 우선 읽기

**파일**: `bot/internal/task/get.go` 수정

- 파일 존재 시 파일에서 읽기
- 파일 없으면 DB fallback (마이그레이션 전 호환)

---

## Phase 3: Git 연동

### 3-1. Git 헬퍼

**파일**: `bot/internal/task/git.go` (신규)

```go
// GitAdd: git add <paths>
func GitAdd(projectPath string, paths ...string) error

// GitCommit: git commit -m <message>
func GitCommit(projectPath, message string) error

// GitRestore: git checkout -- <path> (삭제 복구)
func GitRestore(projectPath, filePath string) error
```

- `exec.Command("git", ...)` 기반
- 에러 시 로그만 (git 실패가 task 실패로 이어지지 않음)

### 3-2. 자동 커밋 연동

Phase 2의 각 지점에 git 커밋 추가:

| 시점 | 커밋 메시지 |
|------|-------------|
| add.go — task 생성 후 | `task(#{id}): created` |
| plan.go — plan 완료 후 | `task(#{id}): planned` |
| run.go — run 완료 후 | `task(#{id}): done` |
| delete.go — 삭제 후 | `task(#{id}): deleted` |
| cycle.go — cycle 완료 후 | `cycle: N done, M failed` |

### 3-3. 삭제 보호

**파일**: `bot/internal/task/file.go` 추가

- 파일 감지 (작업 완료 후 스캔 방식, fswatch는 나중에)
- DB에 있는데 파일 없음 → `GitRestore()` 호출
- `clari task delete`로만 정식 삭제 (git rm + DB 삭제)

---

## Phase 4: DB 축소 + 파일 전환

DB에서 콘텐츠 필드를 제거하고, 파일이 완전한 source of truth가 된다.

### 4-1. 마이그레이션 커맨드

**파일**: `bot/internal/task/migrate.go` (신규)

```bash
clari migrate
```

1. DB에서 전체 tasks 조회
2. 각 task → `tasks/{id}.md` 생성 (frontmatter + spec)
3. plan 있으면 → `tasks/{id}.plan.md`
4. report 있으면 → `tasks/{id}.report.md`
5. git add + commit "migrate: DB → files"

### 4-2. DB 스키마 변경

**파일**: `bot/internal/db/db.go` 수정

```sql
-- 제거 컬럼
spec, plan, report, error

-- 유지 컬럼
id, parent_id, title, status, priority, is_leaf, depth, created_at, updated_at
```

ALTER TABLE로 컬럼 삭제 (SQLite는 지원 안하므로 테이블 재생성).

### 4-3. Task 구조체 변경

**파일**: `bot/internal/task/task.go` 수정

```go
type Task struct {
    ID        int    `json:"id"`
    ParentID  *int   `json:"parent_id,omitempty"`
    Title     string `json:"title"`
    Status    string `json:"status"`
    Priority  int    `json:"priority"`
    IsLeaf    bool   `json:"is_leaf"`
    Depth     int    `json:"depth"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}

// TaskDetail: API 응답용 (파일 내용 포함)
type TaskDetail struct {
    Task
    Spec   string `json:"spec,omitempty"`
    Plan   string `json:"plan,omitempty"`
    Report string `json:"report,omitempty"`
}
```

### 4-4. 읽기 경로 전환

- `Get()` → 파일에서만 읽기 (DB fallback 제거)
- `List()` → DB 메타만 (변경 없음)
- 프롬프트 빌드 (`prompt.go`) → 파일에서 spec/plan 읽기

### 4-5. 쓰기 경로 전환

- `Add()` → 파일 생성 + DB 메타만 INSERT
- `plan.go` → plan 파일 생성 + DB status만 UPDATE
- `run.go` → report 파일 생성 + DB status만 UPDATE

---

## Phase 5: DB 재구축 + 파일 감지

### 5-1. DB 재구축 커맨드

**파일**: `bot/internal/task/rebuild.go` (신규)

```bash
clari rebuild
```

- `.claribot/tasks/*.md` 스캔
- 각 파일 frontmatter 파싱 → DB INSERT
- 기존 DB 드롭 + 재생성

### 5-2. 파일 변경 감지 (sync)

**파일**: `bot/internal/task/sync.go` (신규)

```bash
clari sync
```

- `.claribot/tasks/` 디렉토리 스캔
- 파일 ↔ DB 비교
- 불일치 시 파일 기준으로 DB 갱신
- Cycle/Run 완료 후 자동 호출

### 5-3. Claude PTY 직접 파일 작성 경로

- Claude가 `tasks/{id}.md`를 직접 생성/수정
- 작업 완료 후 `sync` 자동 실행
- 포맷 검증 통과 → DB 반영 → git commit
- 포맷 위반 → 에러 기록 + 알림

---

## Phase 6: 프롬프트 개정

### 6-1. task.md (1회차 프롬프트) 수정

**파일**: `bot/internal/prompts/common/task.md`

- `clari task add` 대신 파일 직접 작성 옵션 추가
- 출력 형식: 파일 경로 명시

### 6-2. task_run.md (2회차 프롬프트) 수정

**파일**: `bot/internal/prompts/common/task_run.md`

- report를 파일로 작성하도록 지침 변경

---

## 구현 순서 요약

```
Phase 1 → 기반 모듈 (file.go, frontmatter.go, validate.go)
Phase 2 → 이중 쓰기 (기존 깨지지 않음, 안전)
Phase 3 → Git 연동 (auto-commit, 삭제 보호)
Phase 4 → DB 축소 (source of truth 전환)
Phase 5 → 재구축/감지 (rebuild, sync)
Phase 6 → 프롬프트 개정 (AI에게 새 규칙 전달)
```

각 Phase는 독립 배포 가능. Phase 2까지는 기존 기능에 영향 없음.

---

## 신규 파일 목록

| 파일 | Phase | 설명 |
|------|-------|------|
| `bot/internal/task/file.go` | 1 | 파일 경로 헬퍼, 디렉토리 생성 |
| `bot/internal/task/frontmatter.go` | 1 | YAML frontmatter 파서/라이터 |
| `bot/internal/task/validate.go` | 1 | 포맷 검증 |
| `bot/internal/task/git.go` | 3 | Git 헬퍼 (add, commit, restore) |
| `bot/internal/task/migrate.go` | 4 | v0.2 → v0.3 마이그레이션 |
| `bot/internal/task/rebuild.go` | 5 | DB 재구축 |
| `bot/internal/task/sync.go` | 5 | 파일 ↔ DB 동기화 |

## 수정 파일 목록

| 파일 | Phase | 변경 |
|------|-------|------|
| `bot/internal/task/add.go` | 2, 4 | 파일 생성 추가, DB 축소 |
| `bot/internal/task/plan.go` | 2, 4 | plan 파일 저장, DB 축소 |
| `bot/internal/task/run.go` | 2, 4 | report 파일 저장, DB 축소 |
| `bot/internal/task/get.go` | 2, 4 | 파일 우선 읽기, DB fallback 제거 |
| `bot/internal/task/task.go` | 4 | 구조체 변경 (콘텐츠 필드 제거) |
| `bot/internal/task/prompt.go` | 4 | 파일에서 spec/plan 읽기 |
| `bot/internal/task/delete.go` | 3 | git rm 추가 |
| `bot/internal/task/cycle.go` | 3 | cycle 완료 시 git commit |
| `bot/internal/db/db.go` | 4 | 스키마 변경 |
| `bot/internal/prompts/common/task.md` | 6 | 파일 작성 지침 |
| `bot/internal/prompts/common/task_run.md` | 6 | report 파일 지침 |

---

*FileSystem Architecture Development Plan v1.0*
