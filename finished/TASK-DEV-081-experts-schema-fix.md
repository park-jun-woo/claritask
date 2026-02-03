# TASK-DEV-081: experts 테이블 스키마 수정

## 개요

specs와 현재 구현의 experts 테이블 불일치 해결

## 스펙 정의 (02-C-Content.md)

```sql
CREATE TABLE experts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    file_path TEXT NOT NULL,
    description TEXT,
    content TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    content_backup TEXT DEFAULT '',
    status TEXT DEFAULT 'active' CHECK(status IN ('active', 'archived')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

## 현재 구현

```sql
CREATE TABLE experts (
    id TEXT PRIMARY KEY,                    -- ← INTEGER로 변경?
    name TEXT NOT NULL,                     -- ← UNIQUE 추가 필요
    version TEXT DEFAULT '1.0.0',           -- ← 스펙에 없음 (유지)
    domain TEXT DEFAULT '',                 -- ← 스펙에 없음 (유지)
    language TEXT DEFAULT '',               -- ← 스펙에 없음 (유지)
    framework TEXT DEFAULT '',              -- ← 스펙에 없음 (유지)
    path TEXT NOT NULL,                     -- ← file_path로 변경?
    content TEXT DEFAULT '',
    content_hash TEXT DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
    -- description 누락
    -- content_backup 누락
    -- status 누락
);
```

## 결정 필요 사항

### ID 타입 문제
- 스펙: INTEGER AUTOINCREMENT
- 현재: TEXT (expert-id 문자열)
- **권장**: TEXT 유지 (expert-id가 폴더명으로 사용되므로)

### 추가 컬럼 (스펙에 없지만 유용)
- version, domain, language, framework
- **권장**: 유지 (Expert 메타데이터로 유용)

## 작업 내용

### 1. 컬럼 추가

```sql
ALTER TABLE experts ADD COLUMN description TEXT DEFAULT '';
ALTER TABLE experts ADD COLUMN content_backup TEXT DEFAULT '';
ALTER TABLE experts ADD COLUMN status TEXT DEFAULT 'active';
```

### 2. UNIQUE 제약 추가

```sql
-- SQLite는 ALTER TABLE로 UNIQUE 추가 불가
-- 새 테이블 생성 후 데이터 마이그레이션 필요
```

### 3. Model 수정 (models.go)

```go
type Expert struct {
    ID            string
    Name          string
    Version       string
    Domain        string
    Language      string
    Framework     string
    Path          string
    Description   string    // 추가
    Content       string
    ContentHash   string
    ContentBackup string    // 추가
    Status        string    // 추가: active, archived
    Assigned      bool
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

### 4. Service 수정

- AddExpert: 새 필드 초기화
- GetExpert: 새 필드 조회
- ArchiveExpert: status 변경 함수 추가

## 완료 조건

- [ ] description, content_backup, status 컬럼 추가
- [ ] Expert 구조체 업데이트
- [ ] AddExpert에서 새 컬럼 처리
- [ ] ArchiveExpert 함수 추가
- [ ] 테스트 통과
