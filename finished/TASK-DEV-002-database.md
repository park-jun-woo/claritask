# TASK-DEV-002: 데이터베이스 레이어

## 파일
`internal/db/db.go`

## 목표
SQLite 연결, 스키마 마이그레이션, 기본 CRUD 함수 구현

## 작업 내용

### 1. DB 구조체 및 연결
```go
type DB struct {
    *sql.DB
    path string
}

// Open - DB 연결 (파일 없으면 생성)
func Open(path string) (*DB, error)

// Close - DB 연결 종료
func (db *DB) Close() error
```

### 2. 스키마 마이그레이션
```go
// Migrate - 테이블 생성 (멱등성 보장)
func (db *DB) Migrate() error
```

**생성할 테이블**:
- projects
- phases
- tasks
- context (싱글톤, id=1 고정)
- tech (싱글톤, id=1 고정)
- design (싱글톤, id=1 고정)
- state (key-value)
- memos

### 3. 유틸리티 함수
```go
// TimeNow - ISO 8601 포맷 현재 시간
func TimeNow() string

// ParseTime - 문자열을 time.Time으로 변환
func ParseTime(s string) (time.Time, error)
```

## 스키마 SQL
```sql
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'active',
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS phases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    order_num INTEGER,
    status TEXT DEFAULT 'pending'
        CHECK(status IN ('pending', 'active', 'done')),
    created_at TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    phase_id INTEGER NOT NULL,
    parent_id INTEGER DEFAULT NULL,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK(status IN ('pending', 'doing', 'done', 'failed')),
    title TEXT NOT NULL,
    level TEXT DEFAULT ''
        CHECK(level IN ('', 'node', 'leaf')),
    skill TEXT DEFAULT '',
    "references" TEXT DEFAULT '[]',
    content TEXT DEFAULT '',
    result TEXT DEFAULT '',
    error TEXT DEFAULT '',
    created_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    failed_at TEXT,
    FOREIGN KEY (phase_id) REFERENCES phases(id),
    FOREIGN KEY (parent_id) REFERENCES tasks(id)
);

CREATE TABLE IF NOT EXISTS context (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    data TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tech (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    data TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS design (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    data TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS state (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS memos (
    scope TEXT NOT NULL,
    scope_id TEXT NOT NULL,
    key TEXT NOT NULL,
    data TEXT NOT NULL,
    priority INTEGER DEFAULT 2
        CHECK(priority IN (1, 2, 3)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (scope, scope_id, key)
);
```

## 참조
- `specs/Talos.md` - 데이터베이스 스키마 섹션
- `internal/model/models.go` - 모델 정의
