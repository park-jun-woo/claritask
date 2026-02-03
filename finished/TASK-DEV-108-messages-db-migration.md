# TASK-DEV-108: Messages DB 마이그레이션

## 목표
`messages` 및 `message_tasks` 테이블 추가

## 변경 파일
- `cli/internal/db/db.go`

## 작업 내용

### 1. messages 테이블 생성 (migrations 추가)
```sql
CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT NOT NULL,
    feature_id INTEGER,
    content TEXT NOT NULL,
    response TEXT DEFAULT '',
    status TEXT DEFAULT 'pending'
        CHECK(status IN ('pending', 'processing', 'completed', 'failed')),
    error TEXT DEFAULT '',
    created_at TEXT NOT NULL,
    completed_at TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (feature_id) REFERENCES features(id)
);
```

### 2. message_tasks 테이블 생성
```sql
CREATE TABLE IF NOT EXISTS message_tasks (
    message_id INTEGER NOT NULL,
    task_id INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (message_id, task_id),
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);
```

### 3. 인덱스 추가
```sql
CREATE INDEX IF NOT EXISTS idx_messages_project ON messages(project_id);
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
CREATE INDEX IF NOT EXISTS idx_message_tasks_message ON message_tasks(message_id);
```

### 4. migrations 테이블에 버전 추가
- migration version 증가

## 테스트
- `go test ./test/db_test.go -v` 실행
- messages, message_tasks 테이블 존재 확인
