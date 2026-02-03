# TASK-DEV-069: Expert DB 마이그레이션

## 목표
`internal/db/db.go`에 Expert 관련 테이블 추가

## 작업 내용

### 1. experts 테이블 추가
```sql
CREATE TABLE IF NOT EXISTS experts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT DEFAULT '1.0.0',
    domain TEXT DEFAULT '',
    language TEXT DEFAULT '',
    framework TEXT DEFAULT '',
    path TEXT NOT NULL,
    created_at TEXT NOT NULL
);
```

### 2. project_experts 테이블 추가
```sql
CREATE TABLE IF NOT EXISTS project_experts (
    project_id TEXT NOT NULL,
    expert_id TEXT NOT NULL,
    assigned_at TEXT NOT NULL,
    PRIMARY KEY (project_id, expert_id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (expert_id) REFERENCES experts(id)
);
```

### 3. Migrate 함수에 스키마 추가
- 기존 schema 문자열에 위 테이블 정의 추가

## 완료 조건
- [ ] experts 테이블 스키마 추가
- [ ] project_experts 테이블 스키마 추가
- [ ] 마이그레이션 실행 확인
- [ ] 컴파일 성공
