# Database: Content Tables

> **현재 버전**: v0.0.6 ([변경이력](../HISTORY.md))

---

## messages

사용자 수정 요청 메시지

```sql
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT NOT NULL,
    feature_id INTEGER,                -- 관련 Feature (선택)
    content TEXT NOT NULL,             -- 요청 내용
    response TEXT DEFAULT '',          -- Claude 분석 결과 (MD)
    status TEXT DEFAULT 'pending'
        CHECK(status IN ('pending', 'processing', 'completed', 'failed')),
    error TEXT DEFAULT '',             -- 에러 메시지 (failed 시)
    created_at TEXT NOT NULL,
    completed_at TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (feature_id) REFERENCES features(id)
);
```

**Status**:

| 상태 | 설명 |
|------|------|
| `pending` | 대기 중 |
| `processing` | Claude 분석 중 |
| `completed` | 완료 (Task 생성됨) |
| `failed` | 실패 |

**워크플로우**:
1. `clari message send` → INSERT (status: pending)
2. Claude 호출 → UPDATE (status: processing)
3. Task 생성 완료 → UPDATE (status: completed, response: 보고서)
4. 에러 발생 시 → UPDATE (status: failed, error: 메시지)

---

## message_tasks

Message와 생성된 Task 연결

```sql
CREATE TABLE message_tasks (
    message_id INTEGER NOT NULL,
    task_id INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (message_id, task_id),
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);
```

**용도**: Message로부터 생성된 Task 추적

---

## memos

메모 (scope 기반)

```sql
CREATE TABLE memos (
    scope TEXT NOT NULL,          -- 'project', 'feature', 'task'
    scope_id TEXT NOT NULL,       -- project_id, feature_id, task_id
    key TEXT NOT NULL,
    data TEXT NOT NULL,           -- JSON
    priority INTEGER DEFAULT 2
        CHECK(priority IN (1, 2, 3)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (scope, scope_id, key)
);
```

**Scope 종류**:

| Scope | scope_id | 용도 |
|-------|----------|------|
| `project` | project_id | 프로젝트 전역 메모 |
| `feature` | feature_id | 특정 Feature 메모 |
| `task` | task_id | 특정 Task 메모 |

**Priority**:

| 값 | 의미 | Manifest 포함 |
|----|------|--------------|
| 1 | 중요 | 자동 포함 |
| 2 | 보통 | 요청 시 포함 |
| 3 | 사소함 | 요청 시 포함 |

**JSON 포맷**:
```json
{
  "value": "Use httpOnly cookies for refresh tokens",
  "summary": "JWT 보안 정책",
  "tags": ["security", "jwt"]
}
```

**Key 포맷**:

| 포맷 | Scope | 예시 |
|------|-------|------|
| `key` | project | `jwt_security` |
| `<feature_id>:key` | feature | `1:api_decisions` |
| `<feature_id>:<task_id>:key` | task | `1:42:impl_notes` |

---

## skeletons

생성된 스켈레톤 파일 추적

```sql
CREATE TABLE skeletons (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feature_id INTEGER NOT NULL,
    file_path TEXT NOT NULL,           -- 생성된 파일 경로
    layer TEXT NOT NULL                -- model, service, api, ui
        CHECK(layer IN ('model', 'service', 'api', 'ui')),
    checksum TEXT NOT NULL,            -- 파일 변경 감지용 SHA256
    created_at TEXT NOT NULL,
    FOREIGN KEY (feature_id) REFERENCES features(id)
);
```

**Layer**:

| Layer | 설명 | 예시 경로 |
|-------|------|----------|
| `model` | 데이터 모델 | `models/comment.py` |
| `service` | 비즈니스 로직 | `services/comment_service.py` |
| `api` | API 엔드포인트 | `api/comment_api.py` |
| `ui` | UI 컴포넌트 | `components/CommentSection.tsx` |

**예시**:
```sql
INSERT INTO skeletons (feature_id, file_path, layer, checksum, created_at)
VALUES (1, 'services/comment_service.py', 'service', 'abc123...', '2026-02-03T10:00:00Z');
```

---

## experts

Expert 문서

```sql
CREATE TABLE experts (
    id TEXT PRIMARY KEY,               -- Expert ID (폴더명, e.g., "backend-go-gin")
    name TEXT NOT NULL,                -- Expert 이름
    version TEXT DEFAULT '1.0.0',      -- 버전
    domain TEXT DEFAULT '',            -- 도메인 설명
    language TEXT DEFAULT '',          -- 주 언어 (Go, Python 등)
    framework TEXT DEFAULT '',         -- 주 프레임워크 (GIN, FastAPI 등)
    path TEXT NOT NULL,                -- EXPERT.md 파일 경로
    description TEXT DEFAULT '',       -- 간단한 설명
    content TEXT DEFAULT '',           -- EXPERT.md 전체 내용
    content_hash TEXT DEFAULT '',      -- 내용 해시 (동기화 감지용)
    content_backup TEXT DEFAULT '',    -- VSCode 편집 전 백업
    status TEXT DEFAULT 'active'
        CHECK(status IN ('active', 'archived')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

**Expert 파일 위치**:
```
<project>/
├── .claritask/
│   └── experts/
│       ├── backend-go-gin/
│       │   └── EXPERT.md
│       ├── frontend-react/
│       │   └── EXPERT.md
│       └── devops-k8s/
│           └── EXPERT.md
```

**동기화 정책**:
1. CLI가 `.md` 파일 수정 → DB 자동 업데이트
2. VSCode가 DB 수정 → `.md` 파일 자동 업데이트
3. 충돌 시 최신 타임스탬프 우선

---

## project_experts

프로젝트-Expert 연결 (프로젝트 레벨 할당)

```sql
CREATE TABLE project_experts (
    project_id TEXT NOT NULL,
    expert_id TEXT NOT NULL,
    assigned_at TEXT NOT NULL,
    PRIMARY KEY (project_id, expert_id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (expert_id) REFERENCES experts(id)
);
```

**용도**: 프로젝트에 할당된 Expert. Task pop 시 manifest에 포함됨.

---

## expert_assignments

Expert-Feature 연결 (Feature 레벨 할당)

```sql
CREATE TABLE expert_assignments (
    expert_id TEXT NOT NULL,
    feature_id INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (expert_id, feature_id),
    FOREIGN KEY (expert_id) REFERENCES experts(id),
    FOREIGN KEY (feature_id) REFERENCES features(id)
);
```

**용도**: Feature 실행 시 해당 Expert 문서를 컨텍스트에 포함

---

## 인덱스

```sql
-- Memo 조회 최적화
CREATE INDEX idx_memos_scope ON memos(scope, scope_id);
CREATE INDEX idx_memos_priority ON memos(priority);

-- Skeleton 조회 최적화
CREATE INDEX idx_skeletons_feature ON skeletons(feature_id);
CREATE INDEX idx_skeletons_layer ON skeletons(layer);

-- Expert 조회 최적화
CREATE INDEX idx_experts_status ON experts(status);
CREATE INDEX idx_project_experts_project ON project_experts(project_id);
CREATE INDEX idx_expert_assignments_feature ON expert_assignments(feature_id);
```

---

*Database Specification v0.0.5*
