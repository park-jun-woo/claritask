# Gap Analysis: specs vs 현재 구현

> **분석일**: 2026-02-03
> **현재 버전**: v0.0.4

---

## 분석 요약

현재 코드와 최종 specs를 비교한 결과 **5개 영역**에서 구현이 필요합니다.

| 영역 | 상태 | 우선순위 |
|------|------|----------|
| experts 테이블 스키마 | 컬럼 누락 | 높음 |
| _migrations 테이블 | 미구현 | 높음 |
| DB 인덱스 | 미구현 | 중간 |
| clari db 명령어 | 미구현 | 중간 |
| Expert 동기화 로직 | 미구현 | 낮음 |

---

## 1. experts 테이블 스키마 차이

### 현재 코드 (db.go:175-184)
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

### specs/CLI/11-Expert.md 정의
```sql
CREATE TABLE IF NOT EXISTS experts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT DEFAULT '1.0.0',
    domain TEXT DEFAULT '',
    language TEXT DEFAULT '',
    framework TEXT DEFAULT '',
    path TEXT NOT NULL,
    content TEXT DEFAULT '',       -- 누락
    content_hash TEXT DEFAULT '',  -- 누락
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL       -- 누락
);
```

### 누락 컬럼
- `content TEXT DEFAULT ''`: EXPERT.md 내용 백업
- `content_hash TEXT DEFAULT ''`: 변경 감지용 해시
- `updated_at TEXT NOT NULL`: 수정 시각

---

## 2. _migrations 테이블 미구현

### specs/DB/03-Migration.md 정의
```sql
CREATE TABLE _migrations (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL
);
```

### 필요 기능
- `getDBVersion()`: 현재 DB 버전 조회
- `setDBVersion()`: DB 버전 설정
- `AutoMigrate()`: 자동 마이그레이션 실행
- `Rollback()`: 롤백 기능
- `BackupDB()`: 백업 생성

---

## 3. DB 인덱스 미구현

### specs 정의 인덱스 (미구현)

```sql
-- Core Tables
CREATE INDEX idx_features_project ON features(project_id);
CREATE INDEX idx_features_status ON features(status);
CREATE INDEX idx_tasks_feature ON tasks(feature_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_task_edges_to ON task_edges(to_task_id);
CREATE INDEX idx_feature_edges_to ON feature_edges(to_feature_id);

-- Content Tables
CREATE INDEX idx_memos_scope ON memos(scope, scope_id);
CREATE INDEX idx_memos_priority ON memos(priority);
CREATE INDEX idx_skeletons_feature ON skeletons(feature_id);
CREATE INDEX idx_skeletons_layer ON skeletons(layer);
CREATE INDEX idx_experts_status ON experts(status);
```

---

## 4. clari db 명령어 미구현

### specs/DB/03-Migration.md 정의

| 명령어 | 설명 |
|--------|------|
| `clari db version` | 현재 DB 버전 확인 |
| `clari db migrate` | 마이그레이션 실행 |
| `clari db rollback --version <n>` | 특정 버전으로 롤백 |
| `clari db backup` | 백업 생성 |

---

## 5. Expert 동기화 로직 미구현

### specs/CLI/11-Expert.md 정의

**동기화 정책**:
1. 파일 수정 → DB content 컬럼에 자동 백업
2. 파일 삭제 → DB 백업에서 자동 복구
3. UI에서 삭제 → DB + 파일 모두 삭제

### 현재 구현
- expert_service.go에서 파일 기반으로 Expert 관리
- DB에 메타데이터만 저장
- content 백업 로직 없음

---

## 구현 계획

### Task 목록

| Task ID | 파일 | 내용 |
|---------|------|------|
| TASK-DEV-074 | db.go | experts 테이블 스키마 업데이트 |
| TASK-DEV-075 | db.go | _migrations 테이블 및 버전 관리 |
| TASK-DEV-076 | db.go | 인덱스 추가 |
| TASK-DEV-077 | cmd/db.go, service/db_service.go | clari db 명령어 구현 |
| TASK-DEV-078 | expert_service.go | Expert 동기화 로직 구현 |

---

*Gap Analysis Report v0.0.4*
