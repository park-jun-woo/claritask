# Database: Migration

> **현재 버전**: v0.0.4 ([변경이력](../HISTORY.md))

---

## 마이그레이션 전략

### 자동 마이그레이션

Claritask 실행 시 DB 버전을 확인하고 자동 마이그레이션:

```go
func AutoMigrate(db *sql.DB) error {
    currentVersion := getDBVersion(db)

    for version := currentVersion + 1; version <= LatestVersion; version++ {
        if err := migrations[version](db); err != nil {
            return fmt.Errorf("migration v%d failed: %w", version, err)
        }
        setDBVersion(db, version)
    }
    return nil
}
```

### 버전 추적

```sql
CREATE TABLE _migrations (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL
);
```

---

## 마이그레이션 히스토리

### v1: 초기 스키마

```sql
-- Core tables
CREATE TABLE projects (...);
CREATE TABLE features (...);
CREATE TABLE tasks (...);
CREATE TABLE task_edges (...);
CREATE TABLE feature_edges (...);

-- Settings tables
CREATE TABLE context (...);
CREATE TABLE tech (...);
CREATE TABLE design (...);
CREATE TABLE state (...);

-- Content tables
CREATE TABLE memos (...);
CREATE TABLE skeletons (...);
```

### v2: Expert 지원

```sql
CREATE TABLE experts (...);
CREATE TABLE expert_assignments (...);

-- 인덱스
CREATE INDEX idx_experts_status ON experts(status);
```

### v3: 낙관적 잠금

```sql
ALTER TABLE features ADD COLUMN version INTEGER DEFAULT 1;
ALTER TABLE tasks ADD COLUMN version INTEGER DEFAULT 1;
```

### v4: Expert 백업

```sql
ALTER TABLE experts ADD COLUMN content_backup TEXT DEFAULT '';
```

---

## 롤백

마이그레이션 실패 시 이전 버전으로 롤백:

```go
func Rollback(db *sql.DB, targetVersion int) error {
    currentVersion := getDBVersion(db)

    for version := currentVersion; version > targetVersion; version-- {
        if err := rollbacks[version](db); err != nil {
            return fmt.Errorf("rollback v%d failed: %w", version, err)
        }
        setDBVersion(db, version-1)
    }
    return nil
}
```

---

## 백업

마이그레이션 전 자동 백업:

```go
func BackupDB(dbPath string) (string, error) {
    backupPath := fmt.Sprintf("%s.backup.%s", dbPath, time.Now().Format("20060102150405"))
    return backupPath, copyFile(dbPath, backupPath)
}
```

**백업 위치**:
```
.claritask/
├── db.clt
├── db.clt.backup.20260203100000
└── db.clt.backup.20260202090000
```

---

## CLI 명령어

```bash
# 현재 DB 버전 확인
clari db version

# 마이그레이션 실행
clari db migrate

# 특정 버전으로 롤백
clari db rollback --version 2

# 백업 생성
clari db backup
```

---

*Database Specification v0.0.4*
