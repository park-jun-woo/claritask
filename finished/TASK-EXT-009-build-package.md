# TASK-EXT-009: 빌드 및 패키징

## 목표
Extension 빌드 및 패키징 설정 완료.

## 파일

### 1. 빌드 스크립트 완성

**package.json scripts 업데이트:**

```json
{
  "scripts": {
    "vscode:prepublish": "npm run compile && npm run build:webview",
    "compile": "tsc -p ./",
    "watch": "tsc -watch -p ./",
    "build:webview": "cd webview-ui && npm install && npm run build",
    "package": "vsce package",
    "lint": "eslint src --ext ts",
    "test": "node ./out/test/runTest.js"
  }
}
```

### 2. .vscodeignore 완성

```
.vscode/**
node_modules/**
webview-ui/node_modules/**
webview-ui/src/**
webview-ui/*.json
webview-ui/*.js
webview-ui/*.ts
webview-ui/index.html
src/**
!webview-ui/dist/**
*.map
.gitignore
tsconfig.json
**/*.md
```

### 3. README.md

```markdown
# Claritask VSCode Extension

Visual editor for Claritask projects (`.clt` files).

## Features

- View Feature/Task hierarchy
- Real-time sync with CLI changes
- WAL mode for concurrent access

## Installation

1. Install from VSIX:
   ```bash
   code --install-extension claritask-0.1.0.vsix
   ```

2. Or build from source:
   ```bash
   npm install
   npm run build:webview
   npm run compile
   ```

## Usage

1. Open a `.clt` file in VSCode
2. The visual editor will open automatically

## Development

```bash
# Install dependencies
npm install
cd webview-ui && npm install && cd ..

# Watch mode
npm run watch

# Build webview
npm run build:webview

# Package
npm run package
```

## Requirements

- VSCode 1.85.0+
- Node.js 18+

## Known Issues

- Phase 1 MVP: Canvas and Inspector not yet implemented
- Conflict resolution UI pending

## Release Notes

### 0.1.0

- Initial release
- Feature/Task tree view
- 1-second polling sync
- WAL mode support
```

### 4. 타입 정의 파일

**src/types.ts:**

```typescript
// VSCode API types are from @types/vscode

// Webview Messages: Extension → Webview
export interface SyncMessage {
  type: 'sync';
  data: ProjectData;
  timestamp: number;
}

export interface ConflictMessage {
  type: 'conflict';
  table: 'tasks' | 'features';
  id: number;
}

export interface ErrorMessage {
  type: 'error';
  message: string;
}

export interface SaveResultMessage {
  type: 'saveResult';
  success: boolean;
  table?: string;
  id?: number;
  error?: string;
}

export type ExtensionMessage =
  | SyncMessage
  | ConflictMessage
  | ErrorMessage
  | SaveResultMessage;

// Webview Messages: Webview → Extension
export interface SaveMessage {
  type: 'save';
  table: 'tasks' | 'features';
  id: number;
  data: Record<string, any>;
  version: number;
}

export interface RefreshMessage {
  type: 'refresh';
}

export interface AddEdgeMessage {
  type: 'addEdge';
  fromId: number;
  toId: number;
}

export interface RemoveEdgeMessage {
  type: 'removeEdge';
  fromId: number;
  toId: number;
}

export type WebviewMessage =
  | SaveMessage
  | RefreshMessage
  | AddEdgeMessage
  | RemoveEdgeMessage;
```

## 빌드 순서

```bash
# 1. 의존성 설치
npm install
cd webview-ui && npm install && cd ..

# 2. Webview 빌드
npm run build:webview

# 3. Extension 컴파일
npm run compile

# 4. 패키지 생성
npm run package
# → claritask-0.1.0.vsix 생성
```

## 테스트 방법

1. VSCode에서 F5 (Extension Development Host)
2. 테스트용 `.clt` 파일 열기
3. GUI 확인

## 완료 조건
- [ ] 빌드 스크립트 완성
- [ ] .vscodeignore 완성
- [ ] README.md 작성
- [ ] types.ts 작성
- [ ] npm run build:webview 성공
- [ ] npm run compile 성공
- [ ] npm run package 성공 (.vsix 생성)
- [ ] F5로 테스트 실행
