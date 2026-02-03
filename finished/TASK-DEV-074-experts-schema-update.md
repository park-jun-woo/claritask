# TASK-DEV-074: experts 테이블 스키마 업데이트

## 개요

experts 테이블에 누락된 컬럼 추가 및 마이그레이션

## 대상 파일

- `cli/internal/db/db.go`
- `cli/internal/model/models.go`
- `cli/internal/service/expert_service.go`

## 작업 내용

### 1. db.go 수정

experts 테이블 스키마 업데이트:

```sql
CREATE TABLE IF NOT EXISTS experts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT DEFAULT '1.0.0',
    domain TEXT DEFAULT '',
    language TEXT DEFAULT '',
    framework TEXT DEFAULT '',
    path TEXT NOT NULL,
    content TEXT DEFAULT '',       -- 추가: EXPERT.md 백업
    content_hash TEXT DEFAULT '',  -- 추가: 변경 감지용
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL       -- 추가: 수정 시각
);
```

### 2. models.go 수정

Expert 구조체에 필드 추가:

```go
type Expert struct {
    ID          string
    Name        string
    Version     string
    Domain      string
    Language    string
    Framework   string
    Path        string
    Content     string    // 추가
    ContentHash string    // 추가
    Assigned    bool
    CreatedAt   time.Time
    UpdatedAt   time.Time // 추가
}
```

### 3. expert_service.go 수정

AddExpert 함수에서 content, content_hash, updated_at 컬럼 처리:

```go
// content 읽기
content, _ := os.ReadFile(expertPath)
hash := sha256.Sum256(content)
contentHash := hex.EncodeToString(hash[:])

// INSERT 쿼리 수정
_, err := database.Exec(
    `INSERT INTO experts (id, name, version, domain, language, framework, path, content, content_hash, created_at, updated_at)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    expertID, expertID, "1.0.0", "", "", "", expertPath, string(content), contentHash, now, now,
)
```

### 4. 마이그레이션 쿼리

기존 DB에 컬럼 추가:

```go
migrations := []string{
    "ALTER TABLE experts ADD COLUMN content TEXT DEFAULT ''",
    "ALTER TABLE experts ADD COLUMN content_hash TEXT DEFAULT ''",
    "ALTER TABLE experts ADD COLUMN updated_at TEXT DEFAULT ''",
}
```

## 완료 조건

- [ ] experts 테이블에 content, content_hash, updated_at 컬럼 추가
- [ ] Expert 모델 구조체 업데이트
- [ ] AddExpert 함수에서 새 컬럼 처리
- [ ] 기존 테스트 통과
