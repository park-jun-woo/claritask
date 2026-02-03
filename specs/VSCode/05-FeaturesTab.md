# VSCode Extension Features 탭

> **현재 버전**: v0.0.7 ([변경이력](../HISTORY.md))

---

## Features 탭 레이아웃

```
┌─────────────────────────────────────────────────────────┐
│  [Project]  [Features]  [Tasks]                         │
├────────────┬────────────────────────────────────────────┤
│            │                                            │
│  Features  │      Feature Detail                        │
│  ──────    │      ──────────────                        │
│  ▸ user_auth │    Name: user_auth                       │
│  ▸ blog_post │    Status: active                        │
│  + Add...    │    Description: 사용자 인증 시스템        │
│              │                                          │
│              │    ┌─ FDL ────────────────────────────┐  │
│              │    │ feature: user_auth               │  │
│              │    │ version: 1.0.0                   │  │
│              │    │ ...                              │  │
│              │    │    [Open File] [Regenerate FDL]  │  │
│              │    └──────────────────────────────────┘  │
│              │                                          │
│              │    [Generate Tasks] [Generate Skeleton]  │
├────────────┴────────────────────────────────────────────┤
│  Status: Connected │ Last sync: 2s ago │ WAL mode: ON   │
└─────────────────────────────────────────────────────────┘
```

---

## Feature 관리 기능

- Feature 목록 트리 뷰
- Feature 추가/삭제
- FDL 코드 표시 (읽기 전용, 파일 열기로 편집)
- Task/Skeleton 생성 버튼

---

## Feature 생성 다이얼로그

`[+ Add...]` 버튼 클릭 시 Feature 생성 다이얼로그 표시:

```
┌─────────────────────────────────────────────────────────┐
│  Create New Feature                              [×]   │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Name:        [user_auth________________]               │
│                                                         │
│  Description:                                           │
│  ┌───────────────────────────────────────────────────┐  │
│  │ 사용자 인증 시스템.                               │  │
│  │ - JWT 기반 로그인/로그아웃                        │  │
│  │ - 회원가입, 비밀번호 찾기                         │  │
│  │ - OAuth 2.0 (Google, GitHub)                     │  │
│  └───────────────────────────────────────────────────┘  │
│                                                         │
│  ※ Description을 기반으로 Claude Code가 FDL을 생성합니다 │
│                                                         │
│                              [Cancel] [Create Feature]  │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### 동작 흐름

1. 사용자가 Name, Description 입력
2. [Create Feature] 버튼 클릭
3. **Extension이 `clari feature add` CLI 명령어 실행**
4. CLI가 TTY Handover로 Claude Code 호출
5. Claude Code가 FDL 생성 → `features/<name>.fdl.yaml` 저장
6. 완료 후 Feature 목록 새로고침

### CLI 호출 (Extension → CLI)

```typescript
// Extension Host에서 CLI 실행
import { spawn } from 'child_process';

async function createFeature(name: string, description: string) {
  const input = JSON.stringify({ name, description });

  // Terminal에서 실행 (TTY Handover 지원)
  const terminal = vscode.window.createTerminal('Claritask');
  terminal.show();
  terminal.sendText(`clari feature add '${input}'`);
}
```

### 메시지 프로토콜

**Webview → Extension:**
```typescript
// Feature 생성 요청 (CLI 호출)
{
  type: 'createFeature',
  data: {
    name: string,
    description: string
  }
}

// FDL 재생성 요청
{ type: 'regenerateFDL', featureId: number }

// Task 생성 요청
{ type: 'generateTasks', featureId: number }

// FDL 파일 열기
{ type: 'openFDLFile', featureId: number }
```

**Extension → Webview:**
```typescript
// CLI 실행 완료 알림
{
  type: 'cliCompleted',
  command: 'feature.add',
  success: boolean,
  featureId?: number,
  error?: string
}
```

---

## FDL 파일 관리

### 파일 구조

```
project/
├── features/
│   ├── user_auth.fdl.yaml     ← FDL 파일
│   └── blog_post.fdl.yaml
└── .claritask/
    └── db.clt
```

### FileSystemWatcher

FDL 파일 변경 감지 및 DB 동기화:

```typescript
// extension.ts
const fdlWatcher = vscode.workspace.createFileSystemWatcher(
  '**/features/*.fdl.yaml'
);

fdlWatcher.onDidChange(uri => syncFDLToDB(uri));
fdlWatcher.onDidCreate(uri => syncFDLToDB(uri));
fdlWatcher.onDidDelete(uri => clearFDLFromDB(uri));
```

### 버튼 동작

- **Open File**: `features/<name>.fdl.yaml` 파일을 VSCode 에디터에서 열기
- **Regenerate FDL**: `clari feature fdl <id>` 실행 (Claude Code 재호출)
- **Generate Tasks**: `clari fdl tasks <id>` 실행
- **Generate Skeleton**: `clari fdl skeleton <id>` 실행

---

*Claritask VSCode Extension Spec v0.0.7*
