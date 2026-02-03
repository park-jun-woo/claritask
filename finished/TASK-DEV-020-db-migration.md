# TASK-DEV-020: DB 마이그레이션

## 개요
- **파일**: `internal/db/db.go`
- **유형**: 수정
- **우선순위**: High
- **Phase**: 1 (핵심 기초 구조)
- **예상 LOC**: +80

## 목적
Feature, FeatureEdge, TaskEdge, Skeleton 테이블 추가 및 Task 테이블 확장

## 작업 내용

### 1. features 테이블 추가
```sql
CREATE TABLE IF NOT EXISTS features (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    spec TEXT DEFAULT '',
    fdl TEXT DEFAULT '',
    fdl_hash TEXT DEFAULT '',
    skeleton_generated INTEGER DEFAULT 0,
    status TEXT DEFAULT 'pending'
        CHECK(status IN ('pending', 'active', 'done')),
    created_at TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id),
    UNIQUE(project_id, name)
);
```

### 2. feature_edges 테이블 추가
```sql
CREATE TABLE IF NOT EXISTS feature_edges (
    from_feature_id INTEGER NOT NULL,
    to_feature_id INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (from_feature_id, to_feature_id),
    FOREIGN KEY (from_feature_id) REFERENCES features(id),
    FOREIGN KEY (to_feature_id) REFERENCES features(id)
);
```

### 3. task_edges 테이블 추가
```sql
CREATE TABLE IF NOT EXISTS task_edges (
    from_task_id INTEGER NOT NULL,
    to_task_id INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (from_task_id, to_task_id),
    FOREIGN KEY (from_task_id) REFERENCES tasks(id),
    FOREIGN KEY (to_task_id) REFERENCES tasks(id)
);
```

### 4. skeletons 테이블 추가
```sql
CREATE TABLE IF NOT EXISTS skeletons (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feature_id INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    layer TEXT NOT NULL
        CHECK(layer IN ('model', 'service', 'api', 'ui')),
    checksum TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (feature_id) REFERENCES features(id)
);
```

### 5. tasks 테이블 확장
기존 tasks 테이블 스키마에 필드 추가:
```sql
-- 기존 tasks 테이블 정의에 추가
feature_id INTEGER DEFAULT NULL,
skeleton_id INTEGER DEFAULT NULL,
target_file TEXT DEFAULT '',
target_line INTEGER,
target_function TEXT DEFAULT '',
FOREIGN KEY (feature_id) REFERENCES features(id),
FOREIGN KEY (skeleton_id) REFERENCES skeletons(id)
```

### 6. Migrate 함수 수정
`Migrate()` 함수의 schema 문자열에 위 테이블들 추가

## 의존성
- TASK-DEV-019 (모델 확장) 완료 필요

## 완료 기준
- [ ] 모든 테이블이 스키마에 추가됨
- [ ] Migrate() 함수 실행 시 테이블 생성됨
- [ ] 기존 DB에 영향 없음 (IF NOT EXISTS)
- [ ] go build 성공
- [ ] 테스트 통과
