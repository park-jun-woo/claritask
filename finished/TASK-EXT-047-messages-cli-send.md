# TASK-EXT-047: Messages CLI Send 연동

## 목표
New Message Send 시 `clari message send` CLI 명령어를 터미널에서 실행

## 작업 내용

### 1. types.ts 수정
```typescript
// MessageFromWebview 추가
| { type: 'sendMessageCLI'; content: string; featureId?: number }
```

### 2. CltEditorProvider.ts 수정

#### handleMessage switch 케이스 추가
```typescript
case 'sendMessageCLI':
  this.handleSendMessageCLI(message, webview, sync);
  break;
```

#### handleSendMessageCLI 메서드 추가
```typescript
private async handleSendMessageCLI(
  message: { content: string; featureId?: number },
  webview: vscode.Webview,
  sync: SyncManager
): Promise<void> {
  // Escape content for shell
  const escapedContent = message.content.replace(/'/g, "'\\''");

  // Build command
  let command = `~/bin/clari message send '${escapedContent}'`;
  if (message.featureId) {
    command += ` --feature ${message.featureId}`;
  }

  // Build full command with cd for WSL
  const isWindows = process.platform === 'win32';
  const workspacePath = vscode.workspace.workspaceFolders?.[0].uri.fsPath;
  let fullCommand = command;

  if (isWindows && workspacePath) {
    const wslPath = windowsToWslPath(workspacePath);
    fullCommand = `cd '${wslPath}' && ${command}`;
  }

  // Use TTY Session Manager
  if (ttySessionManager) {
    const sessionId = `message-send-${Date.now()}`;
    await ttySessionManager.startSession(sessionId, fullCommand);
  } else {
    // Fallback
    const terminal = vscode.window.createTerminal({
      name: 'Claritask - Send Message',
      shellPath: isWindows ? 'wsl.exe' : undefined,
    });
    terminal.show();
    terminal.sendText(fullCommand);
  }

  webview.postMessage({
    type: 'cliStarted',
    command: 'message.send',
    message: 'Claude Code will analyze your request...',
  });
}
```

### 3. useSync.ts 수정
```typescript
export function sendMessageCLI(content: string, featureId?: number): void {
  postMessage({
    type: 'sendMessageCLI',
    content,
    featureId,
  });
}
```

### 4. MessagesPanel.tsx 수정
- handleSend에서 sendMessageCLI 호출
- DB 직접 저장 대신 CLI 명령어 실행

## 완료 조건
- [ ] sendMessageCLI 메시지 타입 추가
- [ ] CltEditorProvider 핸들러 추가
- [ ] TTY 세션 매니저 연동
- [ ] MessagesPanel에서 CLI 호출
