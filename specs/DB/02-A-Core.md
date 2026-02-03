# Database: Core Tables

> **현재 버전**: v0.0.5 ([변경이력](../HISTORY.md))

---

## projects

프로젝트 정보

```sql
CREATE TABLE projects (
    id TEXT PRIMARY KEY,              -- 영문 소문자, 숫자, 하이픈, 언더스코어
    name TEXT NOT NULL,               -- 프로젝트 표시명
    description TEXT,                 -- 프로젝트 설명
    status TEXT DEFAULT 'active'      -- active, archived
        CHECK(status IN ('active', 'archived')),
    created_at TEXT NOT NULL          -- ISO8601
);
```

**예시**:
```sql
INSERT INTO projects VALUES ('blog-api', 'Blog Platform', 'Developer blogging', 'active', '2026-02-03T10:00:00Z');
```

---

## features

Feature (기능 단위)

```sql
CREATE TABLE features (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,                   -- Feature 이름 (snake_case)
    description TEXT,
    spec TEXT DEFAULT '',                 -- Feature 상세 스펙
    fdl TEXT DEFAULT '',                  -- FDL YAML 원문
    fdl_hash TEXT DEFAULT '',             -- FDL 변경 감지용 SHA256
    skeleton_generated INTEGER DEFAULT 0, -- 스켈레톤 생성 완료 여부
    file_path TEXT DEFAULT '',            -- Feature md 파일 경로 (features/<name>.md)
    content TEXT DEFAULT '',              -- md 파일 전체 내용 (동기화용)
    content_hash TEXT DEFAULT '',         -- 내용 해시 (변경 감지용)
    status TEXT DEFAULT 'pending'
        CHECK(status IN ('pending', 'active', 'done')),
    version INTEGER DEFAULT 1,            -- 낙관적 잠금용
    created_at TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id)
);
```

**파일 동기화 필드**:
- `file_path`: `features/<name>.md` 파일 경로
- `content`: md 파일 전체 내용 (양방향 동기화용)
- `content_hash`: 변경 감지용 SHA256 해시

**Status 전이**:
```
pending → active → done
```

---

## tasks

Task (실행 단위)

```sql
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feature_id INTEGER NOT NULL,
    parent_id INTEGER,                    -- 부모 Task ID (계층 구조용, nullable)
    skeleton_id INTEGER,                  -- 연결된 스켈레톤 (nullable)
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK(status IN ('pending', 'doing', 'done', 'failed')),
    title TEXT NOT NULL,
    level TEXT DEFAULT 'leaf',            -- leaf, parent (계층 레벨)
    skill TEXT DEFAULT '',                -- 필요 기술 (sql, python, go 등)
    refs TEXT DEFAULT '',                 -- 참조 목록 (JSON 배열)
    content TEXT DEFAULT '',              -- Task 상세 내용
    target_file TEXT DEFAULT '',          -- 구현 대상 파일 경로
    target_line INTEGER,                  -- 구현 대상 라인 번호
    target_function TEXT DEFAULT '',      -- 구현 대상 함수명
    result TEXT DEFAULT '',               -- Task 완료 시 결과 요약
    error TEXT DEFAULT '',                -- 실패 시 에러 메시지
    version INTEGER DEFAULT 1,            -- 낙관적 잠금용
    created_at TEXT NOT NULL,
    started_at TEXT,                      -- doing 전환 시각
    completed_at TEXT,                    -- done 전환 시각
    failed_at TEXT,                       -- failed 전환 시각
    FOREIGN KEY (feature_id) REFERENCES features(id),
    FOREIGN KEY (parent_id) REFERENCES tasks(id),
    FOREIGN KEY (skeleton_id) REFERENCES skeletons(id)
);
```

**추가 필드 설명**:
- `parent_id`: Task 계층 구조를 위한 부모 Task 참조
- `level`: `leaf`(실행 가능) 또는 `parent`(하위 Task 있음)
- `skill`: Task 실행에 필요한 기술 스택
- `refs`: 참조할 파일/문서 목록 (JSON 배열)

**Status 전이**:
```
pending → doing → done
              └─→ failed
```

**result 필드**: Task 완료 시 결과 요약. 의존하는 Task에 컨텍스트로 전달됨.

---

## task_edges

Task 간 의존성 (DAG)

```sql
CREATE TABLE task_edges (
    from_task_id INTEGER NOT NULL,        -- 의존하는 Task
    to_task_id INTEGER NOT NULL,          -- 의존되는 Task
    created_at TEXT NOT NULL,
    PRIMARY KEY (from_task_id, to_task_id),
    FOREIGN KEY (from_task_id) REFERENCES tasks(id),
    FOREIGN KEY (to_task_id) REFERENCES tasks(id)
);
```

**의미**: `from_task_id`가 `to_task_id`에 의존
- from이 실행되려면 to가 먼저 완료되어야 함
- from 실행 시 to의 `result`가 컨텍스트에 포함됨

**예시**:
```sql
-- Task 3(user_model)이 Task 2(user_table_sql)에 의존
INSERT INTO task_edges VALUES (3, 2, '2026-02-03T10:00:00Z');
```

**제약**: 순환 의존성 불가 (DAG)

---

## feature_edges

Feature 간 의존성

```sql
CREATE TABLE feature_edges (
    from_feature_id INTEGER NOT NULL,     -- 의존하는 Feature
    to_feature_id INTEGER NOT NULL,       -- 의존되는 Feature
    created_at TEXT NOT NULL,
    PRIMARY KEY (from_feature_id, to_feature_id),
    FOREIGN KEY (from_feature_id) REFERENCES features(id),
    FOREIGN KEY (to_feature_id) REFERENCES features(id)
);
```

**의미**: `from_feature_id`가 `to_feature_id`에 의존
- from Feature의 Task들이 실행되려면 to Feature가 먼저 완료되어야 함

**예시**:
```sql
-- 결제 Feature가 로그인 Feature에 의존
INSERT INTO feature_edges VALUES (2, 1, '2026-02-03T10:00:00Z');
```

---

## 인덱스

```sql
-- Feature 조회 최적화
CREATE INDEX idx_features_project ON features(project_id);
CREATE INDEX idx_features_status ON features(status);

-- Task 조회 최적화
CREATE INDEX idx_tasks_feature ON tasks(feature_id);
CREATE INDEX idx_tasks_status ON tasks(status);

-- Edge 조회 최적화
CREATE INDEX idx_task_edges_to ON task_edges(to_task_id);
CREATE INDEX idx_feature_edges_to ON feature_edges(to_feature_id);
```

---

*Database Specification v0.0.5*
