# clari expert - Expert 관리

> **버전**: v0.0.3

## 개요

Expert는 프로젝트에서 사용하는 전문가 역할 정의입니다. 기술 스택별 코딩 규칙, 아키텍처 패턴, 테스트 규칙 등을 정의합니다.

**저장 구조:**
```
.claritask/
├── db.clt              # 메타데이터 + 백업
└── experts/
    ├── backend-go-gin/
    │   └── EXPERT.md   # 실제 내용 (파일 시스템)
    ├── frontend-react/
    │   └── EXPERT.md
    └── devops-k8s/
        └── EXPERT.md
```

**DB 스키마:**
```sql
CREATE TABLE IF NOT EXISTS experts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT DEFAULT '1.0.0',
    domain TEXT DEFAULT '',
    language TEXT DEFAULT '',
    framework TEXT DEFAULT '',
    path TEXT NOT NULL,
    content TEXT DEFAULT '',       -- EXPERT.md 백업
    content_hash TEXT DEFAULT '',  -- 변경 감지용
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS project_experts (
    project_id TEXT NOT NULL,
    expert_id TEXT NOT NULL,
    assigned_at TEXT NOT NULL,
    PRIMARY KEY (project_id, expert_id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (expert_id) REFERENCES experts(id)
);
```

**동기화 정책:**
- **파일 수정** → DB content 컬럼에 자동 백업
- **파일 삭제** → DB 백업에서 자동 복구
- **UI에서 삭제** → DB + 파일 모두 삭제

---

## clari expert add

새 Expert 생성 (폴더 + 템플릿 파일)

```bash
clari expert add <expert-id>
```

**인자:**
- `expert-id` (필수): Expert ID
  - 규칙: 영문 소문자, 숫자, 하이픈(`-`)만 허용

**동작:**
1. `.claritask/experts/<expert-id>/` 폴더 생성
2. `EXPERT.md` 템플릿 파일 생성
3. (옵션) 에디터로 파일 열기

**응답:**
```json
{
  "success": true,
  "expert_id": "backend-go-gin",
  "path": ".claritask/experts/backend-go-gin/EXPERT.md",
  "message": "Expert created. Edit the file to define the expert."
}
```

**에러:**
```json
{
  "success": false,
  "error": "expert 'backend-go-gin' already exists"
}
```

---

## clari expert list

Expert 목록 조회

```bash
clari expert list
clari expert list --assigned          # 현재 프로젝트에 할당된 것만
clari expert list --available         # 할당 가능한 것만
```

**플래그:**
- `--assigned`: 현재 프로젝트에 할당된 Expert만 표시
- `--available`: 할당되지 않은 Expert만 표시

**응답:**
```json
{
  "success": true,
  "experts": [
    {
      "id": "backend-go-gin",
      "name": "Backend Go GIN Developer",
      "domain": "Backend API Development",
      "assigned": true
    },
    {
      "id": "frontend-react",
      "name": "Frontend React Developer",
      "domain": "Frontend Development",
      "assigned": false
    }
  ],
  "total": 2
}
```

---

## clari expert get

Expert 상세 정보 조회

```bash
clari expert get <expert-id>
```

**응답:**
```json
{
  "success": true,
  "expert": {
    "id": "backend-go-gin",
    "name": "Backend Go GIN Developer",
    "version": "1.0.0",
    "domain": "Backend API Development",
    "language": "Go 1.21+",
    "framework": "GIN Web Framework",
    "path": ".claritask/experts/backend-go-gin/EXPERT.md",
    "assigned": true
  }
}
```

**에러:**
```json
{
  "success": false,
  "error": "expert 'backend-go-gin' not found"
}
```

---

## clari expert edit

Expert 파일 편집 (에디터 실행)

```bash
clari expert edit <expert-id>
```

**동작:**
1. `$EDITOR` 환경변수로 지정된 에디터 실행
2. 없으면 `vi` 또는 `notepad` (OS에 따라)

**응답:**
```json
{
  "success": true,
  "expert_id": "backend-go-gin",
  "path": ".claritask/experts/backend-go-gin/EXPERT.md",
  "message": "Opening editor..."
}
```

---

## clari expert remove

Expert 삭제

```bash
clari expert remove <expert-id>
clari expert remove <expert-id> --force
```

**플래그:**
- `--force`: 확인 없이 삭제

**동작:**
1. 프로젝트에서 할당 해제
2. Expert 폴더 전체 삭제

**응답:**
```json
{
  "success": true,
  "expert_id": "backend-go-gin",
  "message": "Expert removed successfully"
}
```

**에러:**
```json
{
  "success": false,
  "error": "expert 'backend-go-gin' is assigned to project. Use --force to remove"
}
```

---

## clari expert assign

프로젝트에 Expert 할당

```bash
clari expert assign <expert-id>
clari expert assign <expert-id> --project <project-id>
```

**플래그:**
- `--project <project-id>`: 특정 프로젝트에 할당 (기본값: 현재 프로젝트)

**동작:**
1. DB의 `project_experts` 테이블에 관계 추가
2. Task pop 시 해당 Expert 내용이 manifest에 포함됨

**응답:**
```json
{
  "success": true,
  "expert_id": "backend-go-gin",
  "project_id": "my-api",
  "message": "Expert assigned to project"
}
```

**에러:**
```json
{
  "success": false,
  "error": "expert 'backend-go-gin' is already assigned to project 'my-api'"
}
```

---

## clari expert unassign

프로젝트에서 Expert 할당 해제

```bash
clari expert unassign <expert-id>
clari expert unassign <expert-id> --project <project-id>
```

**응답:**
```json
{
  "success": true,
  "expert_id": "backend-go-gin",
  "project_id": "my-api",
  "message": "Expert unassigned from project"
}
```

---

*Claritask Commands Reference v0.0.3 - 2026-02-03*
