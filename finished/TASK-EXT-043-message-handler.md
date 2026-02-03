# TASK-EXT-043: Message Handler 추가

## 목표
CltEditorProvider에 Message 관련 메시지 핸들러 추가

## 작업 내용

### 1. src/CltEditorProvider.ts 수정

#### handleMessage switch 케이스 추가
```typescript
case 'sendMessage':
  this.handleSendMessage(message, database, webview, sync);
  break;

case 'deleteMessage':
  this.handleDeleteMessage(message, database, webview, sync);
  break;
```

#### handleSendMessage 메서드 추가
```typescript
private async handleSendMessage(
  message: { content: string; featureId?: number },
  database: Database,
  webview: vscode.Webview,
  sync: SyncManager
): Promise<void> {
  try {
    const project = database.getProject();
    if (!project) {
      webview.postMessage({
        type: 'messageResult',
        success: false,
        action: 'send',
        error: 'No project found',
      });
      return;
    }

    const messageId = database.createMessage(
      project.id,
      message.content,
      message.featureId
    );

    webview.postMessage({
      type: 'messageResult',
      success: true,
      action: 'send',
      messageId,
    });
    sync.refresh();

    // Optional: Start CLI processing in background
    // this.processMessageWithCLI(messageId, message.content);
  } catch (err) {
    webview.postMessage({
      type: 'messageResult',
      success: false,
      action: 'send',
      error: String(err),
    });
  }
}
```

#### handleDeleteMessage 메서드 추가
```typescript
private handleDeleteMessage(
  message: { messageId: number },
  database: Database,
  webview: vscode.Webview,
  sync: SyncManager
): void {
  try {
    database.deleteMessage(message.messageId);
    webview.postMessage({
      type: 'messageResult',
      success: true,
      action: 'delete',
      messageId: message.messageId,
    });
    sync.refresh();
  } catch (err) {
    webview.postMessage({
      type: 'messageResult',
      success: false,
      action: 'delete',
      error: String(err),
    });
  }
}
```

## 완료 조건
- [ ] sendMessage 핸들러 추가
- [ ] deleteMessage 핸들러 추가
- [ ] 에러 처리 구현
- [ ] sync.refresh() 호출
