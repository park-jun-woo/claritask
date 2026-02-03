# TASK-EXT-008: 메시지 핸들러 완성

## 목표
Extension ↔ Webview 간 양방향 메시지 처리 완성.

## 파일
`src/CltEditorProvider.ts` 업데이트

## 구현 내용

### handleMessage 완성

```typescript
private handleMessage(
  message: any,
  database: Database,
  webview: vscode.Webview,
  sync: SyncManager
): void {
  switch (message.type) {
    case 'save':
      this.handleSave(message, database, webview, sync);
      break;

    case 'refresh':
      sync.refresh();
      break;

    case 'addEdge':
      this.handleAddEdge(message, database, webview, sync);
      break;

    case 'removeEdge':
      this.handleRemoveEdge(message, database, webview, sync);
      break;

    case 'createTask':
      this.handleCreateTask(message, database, webview, sync);
      break;

    case 'createFeature':
      this.handleCreateFeature(message, database, webview, sync);
      break;
  }
}

private handleSave(
  message: { table: string; id: number; data: any; version: number },
  database: Database,
  webview: vscode.Webview,
  sync: SyncManager
): void {
  let success = false;

  try {
    if (message.table === 'tasks') {
      success = database.updateTask(message.id, message.data, message.version);
    } else if (message.table === 'features') {
      success = database.updateFeature(message.id, message.data, message.version);
    }

    if (success) {
      webview.postMessage({
        type: 'saveResult',
        success: true,
        table: message.table,
        id: message.id,
      });
      sync.refresh();
    } else {
      // 버전 충돌
      webview.postMessage({
        type: 'conflict',
        table: message.table,
        id: message.id,
      });
    }
  } catch (err) {
    webview.postMessage({
      type: 'saveResult',
      success: false,
      error: String(err),
    });
  }
}

private handleAddEdge(
  message: { fromId: number; toId: number },
  database: Database,
  webview: vscode.Webview,
  sync: SyncManager
): void {
  try {
    database.addTaskEdge(message.fromId, message.toId);
    webview.postMessage({
      type: 'edgeResult',
      success: true,
      action: 'add',
    });
    sync.refresh();
  } catch (err) {
    webview.postMessage({
      type: 'edgeResult',
      success: false,
      error: String(err),
    });
  }
}

private handleRemoveEdge(
  message: { fromId: number; toId: number },
  database: Database,
  webview: vscode.Webview,
  sync: SyncManager
): void {
  try {
    database.removeTaskEdge(message.fromId, message.toId);
    webview.postMessage({
      type: 'edgeResult',
      success: true,
      action: 'remove',
    });
    sync.refresh();
  } catch (err) {
    webview.postMessage({
      type: 'edgeResult',
      success: false,
      error: String(err),
    });
  }
}

private handleCreateTask(
  message: { featureId: number; title: string; content: string },
  database: Database,
  webview: vscode.Webview,
  sync: SyncManager
): void {
  try {
    const id = database.createTask(message.featureId, message.title, message.content);
    webview.postMessage({
      type: 'createResult',
      success: true,
      table: 'tasks',
      id,
    });
    sync.refresh();
  } catch (err) {
    webview.postMessage({
      type: 'createResult',
      success: false,
      error: String(err),
    });
  }
}

private handleCreateFeature(
  message: { name: string; description: string },
  database: Database,
  webview: vscode.Webview,
  sync: SyncManager
): void {
  try {
    const id = database.createFeature(message.name, message.description);
    webview.postMessage({
      type: 'createResult',
      success: true,
      table: 'features',
      id,
    });
    sync.refresh();
  } catch (err) {
    webview.postMessage({
      type: 'createResult',
      success: false,
      error: String(err),
    });
  }
}
```

### Database 추가 메서드

```typescript
// database.ts에 추가

createTask(featureId: number, title: string, content: string): number {
  const now = new Date().toISOString();
  const result = this.db.prepare(`
    INSERT INTO tasks (feature_id, title, content, status, created_at)
    VALUES (?, ?, ?, 'pending', ?)
  `).run(featureId, title, content, now);
  return result.lastInsertRowid as number;
}

createFeature(name: string, description: string): number {
  const projectId = this.getProject()?.id ?? '';
  const now = new Date().toISOString();
  const result = this.db.prepare(`
    INSERT INTO features (project_id, name, description, status, created_at)
    VALUES (?, ?, ?, 'pending', ?)
  `).run(projectId, name, description, now);
  return result.lastInsertRowid as number;
}
```

## 메시지 프로토콜 요약

| Direction | Type | 설명 |
|-----------|------|------|
| W→E | save | 데이터 저장 (낙관적 잠금) |
| W→E | refresh | 수동 새로고침 |
| W→E | addEdge | Edge 추가 |
| W→E | removeEdge | Edge 삭제 |
| W→E | createTask | Task 생성 |
| W→E | createFeature | Feature 생성 |
| E→W | sync | 전체 데이터 동기화 |
| E→W | conflict | 버전 충돌 알림 |
| E→W | saveResult | 저장 결과 |
| E→W | edgeResult | Edge 작업 결과 |
| E→W | createResult | 생성 결과 |

## 완료 조건
- [ ] 모든 메시지 타입 핸들러 구현
- [ ] 낙관적 잠금 충돌 처리
- [ ] 에러 응답 처리
- [ ] Database 생성 메서드 추가
