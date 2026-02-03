# TASK-DEV-075: _migrations 테이블 및 버전 관리 시스템

## 개요

DB 버전 추적을 위한 _migrations 테이블 및 자동 마이그레이션 시스템 구현

## 대상 파일

- `cli/internal/db/db.go`

## 작업 내용

### 1. _migrations 테이블 생성

```sql
CREATE TABLE IF NOT EXISTS _migrations (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL
);
```

### 2. 버전 관리 함수 추가

```go
// LatestVersion is the current schema version
const LatestVersion = 4

// getDBVersion returns current DB version
func (db *DB) getDBVersion() int {
    var version int
    err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM _migrations").Scan(&version)
    if err != nil {
        return 0
    }
    return version
}

// setDBVersion records migration version
func (db *DB) setDBVersion(version int) error {
    now := TimeNow()
    _, err := db.Exec("INSERT INTO _migrations (version, applied_at) VALUES (?, ?)", version, now)
    return err
}
```

### 3. 마이그레이션 함수 정의

```go
type MigrationFunc func(db *DB) error

var migrations = map[int]MigrationFunc{
    1: migrateV1, // 초기 스키마
    2: migrateV2, // Expert 지원
    3: migrateV3, // 낙관적 잠금
    4: migrateV4, // Expert 백업 필드
}
```

### 4. AutoMigrate 함수 구현

```go
func (db *DB) AutoMigrate() error {
    // _migrations 테이블 생성
    db.Exec(`CREATE TABLE IF NOT EXISTS _migrations (
        version INTEGER PRIMARY KEY,
        applied_at TEXT NOT NULL
    )`)

    currentVersion := db.getDBVersion()

    for version := currentVersion + 1; version <= LatestVersion; version++ {
        if migrateFn, ok := migrations[version]; ok {
            if err := migrateFn(db); err != nil {
                return fmt.Errorf("migration v%d failed: %w", version, err)
            }
            if err := db.setDBVersion(version); err != nil {
                return fmt.Errorf("set version v%d failed: %w", version, err)
            }
        }
    }
    return nil
}
```

### 5. 기존 Migrate() 함수 리팩토링

기존 Migrate() 함수를 migrateV1()으로 분리하고, Migrate()에서 AutoMigrate() 호출

## 완료 조건

- [ ] _migrations 테이블 생성
- [ ] getDBVersion(), setDBVersion() 함수 구현
- [ ] AutoMigrate() 함수 구현
- [ ] 기존 스키마를 버전별 마이그레이션 함수로 분리
- [ ] 기존 테스트 통과
