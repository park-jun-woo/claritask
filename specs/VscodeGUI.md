# Claritask VSCode Extension GUI

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

```
┌─────────────────────────────────────────────────────────┐
│  Claritask: my-project                            [⟳]  │
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
│   │   │   ├── FeatureTree.tsx
│   │   │   ├── TaskCanvas.tsx
│   │   │   ├── Inspector.tsx
│   │   │   └── StatusBar.tsx
│   │   ├── hooks/
│   │   │   └── useSync.ts
│   │   └── stores/
│   │       └── projectStore.ts
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
```

### Webview → Extension

```typescript
// 데이터 저장 요청
{ type: 'save', table: 'tasks', id: 3, data: TaskData, version: 5 }

// Edge 생성
{ type: 'addEdge', fromId: 2, toId: 1 }

// 새로고침 요청
{ type: 'refresh' }
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
- [ ] Custom Editor Provider 구현
- [ ] SQLite 읽기 (better-sqlite3)
- [ ] React Webview 기본 구조
- [ ] Feature/Task 트리 뷰
- [ ] 1초 polling 동기화
- [ ] WAL 모드 활성화

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

*Claritask VSCode Extension Spec v1.0 - 2026-02-03*
