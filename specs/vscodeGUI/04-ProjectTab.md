# VSCode Extension Project 탭

> **버전**: v0.0.4

## Project 탭 레이아웃

```
┌─────────────────────────────────────────────────────────┐
│  [Project]  [Features]  [Tasks]                         │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌─ Project Info ─────────────────────────────────────┐ │
│  │  ID:      claritask                                │ │
│  │  Name:    Claritask                                │ │
│  │  Status:  ● active                                 │ │
│  │  Created: 2026-02-03                               │ │
│  └────────────────────────────────────────────────────┘ │
│                                                         │
│  ┌─ Context ──────────────────────────────────[Edit]──┐ │
│  │  Project Name:  Claritask                          │ │
│  │  Description:   AI 에이전트를 위한 작업 관리 CLI...  │ │
│  │  Target Users:  AI 에이전트, 개발자                 │ │
│  │  Deadline:      -                                  │ │
│  └────────────────────────────────────────────────────┘ │
│                                                         │
│  ┌─ Tech Stack ───────────────────────────────[Edit]──┐ │
│  │  backend:   Go 1.21+                               │ │
│  │  frontend:  CLI (Cobra)                            │ │
│  │  database:  SQLite (mattn/go-sqlite3)              │ │
│  └────────────────────────────────────────────────────┘ │
│                                                         │
│  ┌─ Design Decisions ─────────────────────────[Edit]──┐ │
│  │  architecture:  CLI Application                    │ │
│  │  auth_method:   None (로컬 도구)                    │ │
│  │  api_style:     JSON 입출력                        │ │
│  └────────────────────────────────────────────────────┘ │
│                                                         │
│  ┌─ Execution Status ────────────────────────────────┐  │
│  │                                                    │  │
│  │  Progress: ████████░░░░░░░░░░░░  42% (17/40)      │  │
│  │  Status:   ● Running                               │  │
│  │                                                    │  │
│  │  Recent Tasks:                                     │  │
│  │  ✓ #15 user_table_sql    "테이블 생성 완료"        │  │
│  │  ✓ #16 user_model        "모델 구현 완료"          │  │
│  │  ● #17 auth_service      "JWT 구현 중..."          │  │
│  │  ○ #18 login_endpoint    (pending)                 │  │
│  │                                                    │  │
│  └────────────────────────────────────────────────────┘ │
│                                                         │
├─────────────────────────────────────────────────────────┤
│  Status: Connected │ Last sync: 2s ago │ WAL mode: ON   │
└─────────────────────────────────────────────────────────┘
```

## Project 탭 섹션 구성

| 섹션 | 내용 | 편집 |
|------|------|------|
| **Project Info** | id, name, status, created_at | 읽기 전용 |
| **Context** | project_name, description, target_users, deadline | 편집 가능 |
| **Tech Stack** | backend, frontend, database + 사용자 정의 필드 | 편집 가능 |
| **Design Decisions** | architecture, auth_method, api_style + 사용자 정의 필드 | 편집 가능 |
| **Execution Status** | 진행도, 실행 상태, 최근 Task 로그 | 읽기 전용 |

---

## Project Info (읽기 전용)

- **id**: 프로젝트 고유 식별자
- **name**: 프로젝트 이름
- **status**: 프로젝트 상태 (active, paused, completed)
- **created_at**: 생성 일시

### 상태 표시
- **● active**: 초록색 점
- **● paused**: 노란색 점
- **● completed**: 파란색 점

---

## Context 편집

필수 필드와 사용자 정의 필드 지원:
```typescript
interface Context {
  project_name: string;  // 필수
  description: string;   // 필수
  target_users?: string;
  deadline?: string;
  [key: string]: any;    // 사용자 정의 필드
}
```

**편집 모드:**
- 인라인 편집 (필드 클릭 시)
- 폼 다이얼로그 (Edit 버튼 클릭 시)
- 필수 필드 표시 (*)

---

## Tech Stack 편집

필수 필드와 사용자 정의 필드 지원:
```typescript
interface Tech {
  backend: string;   // 필수
  frontend: string;  // 필수
  database: string;  // 필수
  [key: string]: any; // cache, deployment, etc.
}
```

**UI 특징:**
- Key-Value 목록 형식
- 새 필드 추가 버튼 (+)
- 필드 삭제 (휴지통 아이콘)

---

## Design Decisions 편집

필수 필드와 사용자 정의 필드 지원:
```typescript
interface Design {
  architecture: string;  // 필수
  auth_method: string;   // 필수
  api_style: string;     // 필수
  [key: string]: any;    // db_schema, rate_limiting, etc.
}
```

---

## Execution Status (읽기 전용)

Task 실행 진행 상황을 실시간으로 표시. CLI에서 `clari project start` 실행 시 DB 변경을 polling으로 감지하여 UI 갱신.

```typescript
interface ExecutionStatus {
  total: number;      // 전체 Task 수
  pending: number;    // 대기 중
  doing: number;      // 실행 중
  done: number;       // 완료
  failed: number;     // 실패
  progress: number;   // 진행률 (0-100)
}

interface RecentTask {
  id: number;
  title: string;
  status: 'pending' | 'doing' | 'done' | 'failed';
  result?: string;    // 완료 시 한줄평
  error?: string;     // 실패 시 에러
}
```

**UI 요소:**
- **Progress Bar**: 시각적 진행률 표시
- **Status Indicator**: Running / Idle / Completed / Has Failures
- **Recent Tasks**: 최근 4-5개 Task 로그
  - `✓` 완료 (녹색)
  - `●` 실행 중 (파란색, 애니메이션)
  - `✗` 실패 (빨간색)
  - `○` 대기 중 (회색)

**상태 표시:**
- **● Running**: 파란색 - doing Task가 있음
- **● Idle**: 회색 - 대기 중 (pending만 있음)
- **● Completed**: 녹색 - 모든 Task 완료
- **● Has Failures**: 빨간색 - 실패한 Task 있음

---

## Project 탭 컴포넌트 구조

```
ProjectPanel/
├── ProjectInfo.tsx        # 읽기 전용 프로젝트 기본 정보
├── ContextSection.tsx     # Context 편집 섹션
├── TechSection.tsx        # Tech Stack 편집 섹션
├── DesignSection.tsx      # Design Decisions 편집 섹션
├── ExecutionStatus.tsx    # 실행 상태 및 진행률 표시
├── ProgressBar.tsx        # 진행률 바 컴포넌트
├── RecentTaskList.tsx     # 최근 Task 로그 목록
└── EditableField.tsx      # 공통 필드 편집 컴포넌트
```

---

*Claritask VSCode Extension Spec v0.0.4 - 2026-02-03*
