# File-Centric Architecture

> Claribot v0.3 — 파일 중심 데이터 관리 설계

---

## Overview

모든 콘텐츠를 파일로 관리하고, DB에는 메타데이터만 저장한다. Git이 백업/이력/복구 인프라 역할을 한다.

**3층 구조**:

```
┌─────────────┐
│   Files     │  Source of truth. 모든 콘텐츠
├─────────────┤
│   Git       │  백업, 이력, 복구
├─────────────┤
│   DB (meta) │  인덱스, 트리 관계, 상태 조회
└─────────────┘
```

**핵심 원칙**:
- 파일이 source of truth
- DB는 메타 인덱스 (목록 조회, 트리 탐색용)
- DB가 날아가도 파일 + git에서 전체 복원 가능
- Claribot은 포맷 검증자 (validator)

---

## 설계 배경

### 기존 방식의 문제

| 문제 | 원인 |
|------|------|
| CLI 파싱 실패 (title/spec 분리 오류) | 인라인 멀티라인 전달 한계 |
| AI의 CLI 명령어 비준수 | `--spec-file` 강제해도 인라인 사용 |
| 검증 부재 | 결과물 형식 확인 없이 DB 저장 |
| 디버깅 어려움 | DB 쿼리 필요, 내용 직접 확인 불가 |

### 파일 중심의 이점

- **AI 친화적**: 파일 쓰기는 AI가 가장 안정적으로 수행하는 작업
- **검증 용이**: 파일 존재 + frontmatter 파싱으로 구조 검증
- **git 네이티브**: diff/blame/history 기본 제공
- **디버깅**: 파일 열어보면 끝

---

## 파일 구조

```
<project>/
└── .claribot/
    ├── db.clt              # 메타 DB (인덱스)
    ├── tasks/
    │   ├── 1.md            # Task #1 (frontmatter + spec)
    │   ├── 1.plan.md       # Task #1 계획서
    │   ├── 1.report.md     # Task #1 완료 보고서
    │   ├── 2.md
    │   ├── 2.plan.md
    │   └── ...
    └── specs/
        ├── auth.md         # Spec 문서
        └── refactoring.md
```

### 파일 명명 규칙

| 파일 | 용도 |
|------|------|
| `tasks/{id}.md` | Task 본문 (spec) |
| `tasks/{id}.plan.md` | 계획서 (1회차 순회 결과) |
| `tasks/{id}.report.md` | 완료 보고서 (2회차 순회 결과) |
| `specs/{name}.md` | 독립 Spec 문서 |

---

## Task 파일 포맷

### tasks/{id}.md

```markdown
---
status: todo
parent: 1
priority: 0
---
# API 엔드포인트 리팩토링

## Spec
RESTful 규칙에 맞게 API 경로 체계를 변경한다.

## 변경 파일
- `bot/internal/handler/router.go` - 라우트 정의 변경
- `cli/cmd/clari/main.go` - CLI 경로 매칭

## 세부사항
- GET /tasks → GET /tasks
- POST /tasks → POST /tasks
- GET /tasks/:id → GET /tasks/:id
```

### Frontmatter 필드

| 필드 | 필수 | 기본값 | 설명 |
|------|------|--------|------|
| `status` | Y | `todo` | `todo`, `planned`, `split`, `done`, `failed` |
| `parent` | N | (없음) | 상위 Task ID |
| `priority` | N | `0` | 실행 우선순위 (높을수록 먼저) |

### 포맷 룰

1. `---`로 시작하는 YAML frontmatter 필수
2. `status` 필드 필수, 허용 값만 사용
3. Frontmatter 다음에 `# 제목` (H1) 필수 — DB의 title로 추출
4. 본문 비어있지 않을 것

### tasks/{id}.plan.md

```markdown
## 구현 방향
Express 스타일 라우터를 RESTful 규칙에 맞게 변경

## 변경 파일
- `router.go` - 라우트 재정의

## 구현 순서
1. 기존 라우트 매핑 정리
2. RESTful 규칙 적용
3. CLI 경로 매칭 수정

## 검증 방법
- go build 통과
- 기존 API 호출 정상 동작
```

### tasks/{id}.report.md

```markdown
## 결과
API 엔드포인트 RESTful 규칙 적용 완료

## 변경 사항
- `router.go`: 12개 라우트 변경
- `main.go`: CLI 경로 매칭 수정

## 검증
- go build 통과
- 테스트 실행 정상
```

---

## DB 스키마 (메타만)

```sql
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY,
    parent_id INTEGER,
    title TEXT,              -- H1에서 추출
    status TEXT DEFAULT 'todo'
        CHECK(status IN ('todo', 'split', 'planned', 'done', 'failed')),
    priority INTEGER DEFAULT 0,
    is_leaf INTEGER DEFAULT 1,
    depth INTEGER DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX idx_tasks_parent ON tasks(parent_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_leaf ON tasks(is_leaf);
```

**제거된 필드** (파일로 이동):
- `spec` → `tasks/{id}.md` 본문
- `plan` → `tasks/{id}.plan.md`
- `report` → `tasks/{id}.report.md`
- `error` → `tasks/{id}.report.md` 또는 frontmatter

---

## 데이터 흐름

### 파일 생성/수정 (단방향)

```
파일 생성 또는 수정
    ↓
Claribot 감지 (fswatch 또는 작업 완료 후 스캔)
    ↓
포맷 검증 (frontmatter 필수 필드, H1 존재, 허용 status 값)
    ↓
  ✅ 통과 → DB 메타 갱신 (frontmatter에서 추출) → git add + commit
  ❌ 위반 → 에러 기록 + 알림
```

### 파일 생성 경로 (2가지, 결과 동일)

```
경로 A: CLI 경유
  clari task add --spec-file spec.md
    → ID 채번 → tasks/{id}.md 생성 → DB 메타 INSERT → git commit

경로 B: 직접 파일 작성 (Claude PTY)
  tasks/5.md 직접 생성
    → Claribot 감지 → 포맷 검증 → DB 메타 INSERT → git commit
```

### 파일 삭제 (보호)

```
직접 삭제 감지 (rm tasks/4.md)
    → git checkout -- tasks/4.md (즉시 복구)
    → 경고 알림

정식 삭제 (clari task delete 4)
    → DB 메타 삭제 → git rm tasks/4.md → git commit
```

### 조회

```
clari task list     → DB 메타에서 빠르게 (id, status, title, parent)
clari task get 4    → 파일 직접 읽기 (tasks/4.md 전체 내용)
```

---

## Git 전략

### 의존성

Git은 이 설계의 인프라 레이어. 필수 의존성.
- `project create` 시 `git init` (기존)
- `.claribot/tasks/` 디렉토리를 git 추적 대상에 포함

### Commit (로컬, 자주)

상태 변경 단위로 자동 커밋:

| 시점 | 커밋 메시지 |
|------|-------------|
| Task 생성 | `task(#4): created` |
| Plan 완료 | `task(#4): planned` |
| Run 완료 | `task(#4): done` |
| Task 삭제 | `task(#4): deleted` |
| Cycle 완료 | `cycle: 3 done, 1 failed` |

### Push (리모트, 크게)

Push는 수동 또는 설정에 따라:

```yaml
# .claribot/config (project-level)
git:
  auto_commit: true      # 상태 변경 시 자동 커밋 (기본값)
  auto_push: false       # push는 수동 (기본값)
  # auto_push: cycle     # cycle 완료 시 push
  # auto_push: always    # 커밋마다 push
```

```bash
clari push               # 수동 push
```

### 복구

```bash
# DB 재구축 (DB 손실 시)
clari rebuild            # .claribot/tasks/*.md 스캔 → DB 메타 재생성

# 파일 복구 (실수로 삭제 시)
# → Claribot이 자동으로 git checkout 실행
```

---

## Claribot 역할 변화

### Before (v0.2): 데이터 관리자

```
CLI → Claribot → DB (전체 데이터 저장)
Claude PTY → CLI → Claribot → DB
```

### After (v0.3): 포맷 검증자

```
CLI → 파일 생성 + DB 메타
Claude PTY → 파일 직접 생성
Claribot → 파일 감지 → 검증 → DB 메타 갱신 → git commit
```

| 기능 | v0.2 | v0.3 |
|------|------|------|
| 데이터 저장 | DB | 파일 |
| 메타 인덱스 | DB | DB (메타만) |
| 백업/이력 | 없음 | git |
| 복구 | 불가 | git checkout |
| 검증 | 없음 | frontmatter 포맷 검증 |
| AI 데이터 입력 | CLI 명령어 | 파일 직접 작성 |

---

## 마이그레이션 (v0.2 → v0.3)

```bash
clari migrate
```

1. 기존 DB에서 tasks 조회
2. 각 task를 `tasks/{id}.md` 파일로 변환 (frontmatter + spec)
3. plan이 있으면 `tasks/{id}.plan.md` 생성
4. report가 있으면 `tasks/{id}.report.md` 생성
5. DB를 메타 전용 스키마로 재생성
6. git add + commit

---

*Claribot File-Centric Architecture v0.3*
