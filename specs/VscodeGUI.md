# Claritask VSCode Extension GUI

> **버전**: v0.0.3

## 변경이력
| 버전 | 날짜 | 내용 |
|------|------|------|
| v0.0.3 | 2026-02-03 | Execution Status 섹션 추가 |
| v0.0.2 | 2026-02-03 | Project 탭 설계 추가 |
| v0.0.1 | 2026-02-03 | 최초 작성 |

## 개요

`.clt` 확장자의 SQLite 데이터베이스 파일을 VSCode 내에서 시각적으로 편집하는 GUI 확장 프로그램.

**목표**: CLI와 GUI가 동일한 DB 파일을 실시간으로 공유하며 편집

---

## 아키텍처

### Phase 1: MVP (WAL + Polling)

```
┌─────────────────┐         ┌─────────────────┐
│  Claude Code    │         │  VSCode GUI     │
│  (clari cli)    │         │  (Webview)      │
└────────┬────────┘         └────────┬────────┘
         │                           │
         │ write                     │ read (1초 polling)
         ▼                           ▼
      ┌─────────────────────────────────┐
      │     db.clt (SQLite + WAL)       │
      └─────────────────────────────────┘
```

### Phase 2: File Watcher

```
db.clt 변경 감지 (fs.watch)
    ↓
즉시 SQLite 재읽기
    ↓
GUI 업데이트
```

### Phase 3: Daemon (향후)

```
┌──────────┐     ┌──────────┐
│  CLI     │     │   GUI    │
└────┬─────┘     └────┬─────┘
     │  WebSocket     │
     ▼                ▼
┌─────────────────────────┐
│   clari daemon          │
│   - SQLite 단독 접근    │
│   - 변경 시 broadcast   │
└─────────────────────────┘
```

---

## 파일 확장자

| 확장자 | 설명 |
|--------|------|
| `.clt` | Claritask SQLite 데이터베이스 |

**기존 경로 변경:**
```
Before: .claritask/db
After:  .claritask/db.clt
```

VSCode에서 `.clt` 파일 열면 Custom Editor가 활성화됨.

---

## 기술 스택

### Extension Host (Node.js)
- TypeScript
- VSCode Extension API
- better-sqlite3 (동기 SQLite 바인딩)

### Webview (브라우저)
- React 18
- TypeScript
- TailwindCSS
- React Flow (드래그앤드롭 캔버스)
- @vscode/webview-ui-toolkit

---

## UI 레이아웃

### 탭 구조

```
┌─────────────────────────────────────────────────────────┐
│  Claritask: my-project                            [⟳]  │
├─────────────────────────────────────────────────────────┤
│  [Project]  [Features]  [Tasks]                         │
├─────────────────────────────────────────────────────────┤
```

3개의 메인 탭으로 구성:
- **Project**: 프로젝트 정보, Context, Tech, Design 조회/편집
- **Features**: Feature 목록 및 관리
- **Tasks**: Task 목록 및 Canvas 뷰

---

### Project 탭 레이아웃

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

### Project 탭 섹션 구성

| 섹션 | 내용 | 편집 |
|------|------|------|
| **Project Info** | id, name, status, created_at | 읽기 전용 |
| **Context** | project_name, description, target_users, deadline | 편집 가능 |
| **Tech Stack** | backend, frontend, database + 사용자 정의 필드 | 편집 가능 |
| **Design Decisions** | architecture, auth_method, api_style + 사용자 정의 필드 | 편집 가능 |
| **Execution Status** | 진행도, 실행 상태, 최근 Task 로그 | 읽기 전용 |

---

### Features 탭 레이아웃

```
┌─────────────────────────────────────────────────────────┐
│  [Project]  [Features]  [Tasks]                         │
├────────────┬────────────────────────────────────────────┤
│            │                                            │
│  Features  │      Feature Detail                        │
│  ──────    │      ──────────────                        │
│  ▸ user_auth │    Name: user_auth                       │
│  ▸ blog_post │    Status: active                        │
│  + Add...    │    Description: ...                      │
│              │                                          │
│              │    ┌─ Spec ───────────────────────────┐  │
│              │    │ # User Authentication            │  │
│              │    │ ...                              │  │
│              │    └──────────────────────────────────┘  │
│              │                                          │
│              │    ┌─ FDL ────────────────────────────┐  │
│              │    │ feature: user_auth               │  │
│              │    │ ...                              │  │
│              │    └──────────────────────────────────┘  │
├────────────┴────────────────────────────────────────────┤
│  Status: Connected │ Last sync: 2s ago │ WAL mode: ON   │
└─────────────────────────────────────────────────────────┘
```

---

### Tasks 탭 레이아웃 (기존)

```
┌─────────────────────────────────────────────────────────┐
│  [Project]  [Features]  [Tasks]                         │
├────────────┬────────────────────────┬───────────────────┤
│            │                        │                   │
│  Features  │      Canvas            │    Inspector      │
│  ──────    │      (드래그앤드롭)     │    ──────────     │
│  □ user_auth │                      │    Feature:       │
│    ├ task1   │   ┌─────┐  ┌─────┐   │    - name         │
│    ├ task2   │   │ T1  │→→│ T2  │   │    - spec         │
│    └ task3   │   └─────┘  └─────┘   │    - status       │
│  □ blog_post │        ↓             │                   │
│              │   ┌─────┐            │    Task:          │
│              │   │ T3  │            │    - title        │
│              │   └─────┘            │    - content      │
│              │                      │    - skill        │
│              │                      │    - status       │
├────────────┴────────────────────────┴───────────────────┤
│  Status: Connected │ Last sync: 2s ago │ WAL mode: ON   │
└─────────────────────────────────────────────────────────┘
```

### 패널 구성

| 패널 | 기능 |
|------|------|
| **Left: Tree View** | Feature/Task 계층 구조 |
| **Center: Canvas** | Task 노드 시각화, 드래그앤드롭으로 Edge 연결 |
| **Right: Inspector** | 선택된 항목 속성 편집 |
| **Bottom: Status Bar** | 동기화 상태, WAL 모드, 마지막 업데이트 |

---

## 핵심 기능

### 0. Project 정보 관리

Project 탭에서 프로젝트 메타데이터와 설정을 조회/편집.

#### Project Info (읽기 전용)
- **id**: 프로젝트 고유 식별자
- **name**: 프로젝트 이름
- **status**: 프로젝트 상태 (active, paused, completed)
- **created_at**: 생성 일시

#### Context 편집
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

#### Tech Stack 편집
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

#### Design Decisions 편집
필수 필드와 사용자 정의 필드 지원:
```typescript
interface Design {
  architecture: string;  // 필수
  auth_method: string;   // 필수
  api_style: string;     // 필수
  [key: string]: any;    // db_schema, rate_limiting, etc.
}
```

#### Execution Status (읽기 전용)

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

#### Project 탭 컴포넌트 구조

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

#### 상태 표시
- **● active**: 초록색 점
- **● paused**: 노란색 점
- **● completed**: 파란색 점

---

### 1. Feature 관리

- Feature 목록 트리 뷰
- Feature 추가/삭제/편집
- Feature 스펙 (Markdown) 편집
- FDL 코드 편집 (코드 에디터 내장)

### 2. Task 관리

- Task 노드 시각화
- 드래그앤드롭으로 Task 생성
- Task 상태 표시 (색상 구분)
  - `pending`: 회색
  - `doing`: 파란색
  - `done`: 녹색
  - `failed`: 빨간색

### 3. Edge (의존성) 관리

- 노드 간 드래그로 Edge 생성
- Edge 클릭으로 삭제
- 순환 의존성 감지 및 경고

### 4. 실시간 동기화

- CLI 변경 사항 자동 반영
- 충돌 감지 및 알림
- 수동 새로고침 버튼

### 5. Context/Tech/Design 편집

- JSON 에디터 또는 폼 기반 UI
- 스키마 검증

---

## SQLite 동기화 전략

### WAL 모드 활성화

```sql
PRAGMA journal_mode=WAL;
PRAGMA busy_timeout=5000;
```

### Polling 메커니즘

```typescript
// Extension Host
setInterval(async () => {
  const mtime = fs.statSync(dbPath).mtimeMs;
  if (mtime !== lastMtime) {
    lastMtime = mtime;
    const data = readDatabase();
    webview.postMessage({ type: 'sync', data });
  }
}, 1000);
```

### 충돌 방지: 낙관적 잠금

**스키마 변경:**
```sql
ALTER TABLE tasks ADD COLUMN version INTEGER DEFAULT 1;
ALTER TABLE features ADD COLUMN version INTEGER DEFAULT 1;
```

**업데이트 로직:**
```sql
UPDATE tasks
SET title = ?, status = ?, version = version + 1
WHERE id = ? AND version = ?;
-- affected_rows = 0 이면 충돌
```

**GUI 처리:**
```
저장 실패 → "외부에서 수정됨" 다이얼로그
         → [새로고침] [강제 저장] [취소]
```

---

## Extension 구조

```
claritask-vscode/
├── package.json              # Extension manifest
├── src/
│   ├── extension.ts          # Extension 진입점
│   ├── CltEditorProvider.ts  # Custom Editor Provider
│   ├── database.ts           # SQLite 읽기/쓰기
│   └── sync.ts               # Polling 로직
├── webview-ui/
│   ├── package.json
│   ├── src/
│   │   ├── App.tsx
│   │   ├── components/
│   │   │   ├── FeatureList.tsx      # Feature 목록
│   │   │   ├── TaskPanel.tsx        # Task 관리 패널
│   │   │   ├── StatusBar.tsx        # 하단 상태 바
│   │   │   ├── ProjectPanel.tsx     # Project 탭 메인
│   │   │   ├── ProjectInfo.tsx      # 프로젝트 기본 정보 (읽기 전용)
│   │   │   ├── ContextSection.tsx   # Context 편집 섹션
│   │   │   ├── TechSection.tsx      # Tech Stack 편집 섹션
│   │   │   ├── DesignSection.tsx    # Design Decisions 편집 섹션
│   │   │   ├── ExecutionStatus.tsx  # 실행 상태 표시
│   │   │   ├── ProgressBar.tsx      # 진행률 바
│   │   │   ├── RecentTaskList.tsx   # 최근 Task 로그
│   │   │   ├── EditableField.tsx    # Key-Value 편집 컴포넌트
│   │   │   └── SectionCard.tsx      # 섹션 카드 래퍼
│   │   ├── hooks/
│   │   │   └── useSync.ts
│   │   └── stores/
│   │       └── store.ts
│   └── vite.config.ts
└── README.md
```

---

## package.json (Extension Manifest)

```json
{
  "name": "claritask",
  "displayName": "Claritask",
  "description": "Visual editor for Claritask projects",
  "version": "0.1.0",
  "engines": {
    "vscode": "^1.85.0"
  },
  "categories": ["Other"],
  "activationEvents": [],
  "main": "./out/extension.js",
  "contributes": {
    "customEditors": [
      {
        "viewType": "claritask.cltEditor",
        "displayName": "Claritask Editor",
        "selector": [
          {
            "filenamePattern": "*.clt"
          }
        ],
        "priority": "default"
      }
    ]
  }
}
```

---

## 메시지 프로토콜

### Extension → Webview

```typescript
// 전체 데이터 동기화
{ type: 'sync', data: ProjectData }

// 부분 업데이트
{ type: 'update', table: 'tasks', id: 3, data: TaskData }

// 충돌 알림
{ type: 'conflict', table: 'tasks', id: 3 }

// Context/Tech/Design 저장 결과
{ type: 'settingSaveResult', section: 'context' | 'tech' | 'design', success: boolean, error?: string }
```

### Webview → Extension

```typescript
// 데이터 저장 요청
{ type: 'save', table: 'tasks', id: 3, data: TaskData, version: 5 }

// Edge 생성
{ type: 'addEdge', fromId: 2, toId: 1 }

// 새로고침 요청
{ type: 'refresh' }

// Context 저장
{ type: 'saveContext', data: ContextData }

// Tech 저장
{ type: 'saveTech', data: TechData }

// Design 저장
{ type: 'saveDesign', data: DesignData }
```

---

## 데이터 타입

```typescript
interface ProjectData {
  project: Project;
  features: Feature[];
  tasks: Task[];
  taskEdges: Edge[];
  featureEdges: Edge[];
  context: Record<string, any>;
  tech: Record<string, any>;
  design: Record<string, any>;
  state: Record<string, string>;
  memos: Memo[];
}

interface Feature {
  id: number;
  name: string;
  spec: string;
  fdl: string;
  fdl_hash: string;
  status: string;
  version: number;
}

interface Task {
  id: number;
  feature_id: number;
  parent_id: number | null;
  title: string;
  content: string;
  level: string;
  skill: string;
  status: 'pending' | 'doing' | 'done' | 'failed';
  version: number;
}

interface Edge {
  from_id: number;
  to_id: number;
}
```

---

## 로드맵

### Phase 1: MVP
- [x] Custom Editor Provider 구현
- [x] SQLite 읽기 (sql.js)
- [x] React Webview 기본 구조
- [x] Feature/Task 트리 뷰
- [x] 1초 polling 동기화
- [x] WAL 모드 활성화

### Phase 1.5: Project 탭
- [ ] Project 탭 UI 구현
- [ ] ProjectPanel 컴포넌트
- [ ] ProjectInfo (읽기 전용) 컴포넌트
- [ ] ContextSection 편집 컴포넌트
- [ ] TechSection 편집 컴포넌트
- [ ] DesignSection 편집 컴포넌트
- [ ] ExecutionStatus 컴포넌트
  - [ ] ProgressBar (진행률 시각화)
  - [ ] RecentTaskList (최근 Task 로그)
  - [ ] 상태 표시 (Running/Idle/Completed/Has Failures)
- [ ] Context/Tech/Design 저장 메시지 핸들러
- [ ] 필수 필드 검증 (required indicator)
- [ ] 사용자 정의 필드 추가/삭제

### Phase 2: Canvas
- [ ] React Flow 통합
- [ ] Task 노드 시각화
- [ ] 드래그앤드롭 Edge 생성
- [ ] 상태별 색상 표시

### Phase 3: 편집 기능
- [ ] Inspector 패널
- [ ] Task 속성 편집
- [ ] Feature 스펙 편집
- [ ] FDL 코드 편집기

### Phase 4: 동기화 강화
- [ ] File watcher 추가
- [ ] 낙관적 잠금 구현
- [ ] 충돌 해결 UI

### Phase 5: Daemon (선택)
- [ ] clari daemon 명령어
- [ ] WebSocket 서버
- [ ] 실시간 push 동기화

---

## CLI 호환성

### 확장자 변경 마이그레이션

```bash
# 기존 프로젝트 마이그레이션
mv .claritask/db .claritask/db.clt
```

### clari CLI 수정 사항

1. DB 경로 변경: `.claritask/db` → `.claritask/db.clt`
2. WAL 모드 기본 활성화
3. version 컬럼 마이그레이션 추가

---

## 참고 자료

- [VSCode Custom Editor API](https://code.visualstudio.com/api/extension-guides/custom-editors)
- [VSCode Webview UI Toolkit](https://github.com/microsoft/vscode-webview-ui-toolkit)
- [React Flow](https://reactflow.dev/)
- [better-sqlite3](https://github.com/WiseLibs/better-sqlite3)
- [SQLite WAL Mode](https://www.sqlite.org/wal.html)

---

*Claritask VSCode Extension Spec v1.2 - 2026-02-03*
