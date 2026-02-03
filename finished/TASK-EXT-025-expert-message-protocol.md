# TASK-EXT-025: Expert 메시지 프로토콜

## 개요
Extension과 Webview 간 Expert 관련 메시지 프로토콜 구현

## 배경
- **스펙**: specs/VSCode/11-MessageProtocol.md
- **현재 상태**: Expert 관련 메시지 타입 없음

## 작업 내용

### 1. 메시지 타입 정의
**파일**: `vscode-extension/src/types.ts`

```typescript
// Webview -> Extension
export type WebviewMessage =
  | { type: 'assignExpert'; expertId: string }
  | { type: 'unassignExpert'; expertId: string }
  | { type: 'openExpertFile'; expertId: string }
  | { type: 'createExpert'; expertId: string }
  | { type: 'deleteExpert'; expertId: string }
  // 기존 메시지들...

// Extension -> Webview
export type ExtensionMessage =
  | { type: 'expertsUpdated'; experts: Expert[] }
  | { type: 'expertAssignResult'; success: boolean; error?: string }
  // 기존 메시지들...
```

### 2. CltEditorProvider.ts 핸들러 추가
**파일**: `vscode-extension/src/CltEditorProvider.ts`

```typescript
private handleMessage(message: WebviewMessage) {
  switch (message.type) {
    // 기존 케이스들...

    case 'assignExpert':
      this.handleAssignExpert(message.expertId);
      break;

    case 'unassignExpert':
      this.handleUnassignExpert(message.expertId);
      break;

    case 'openExpertFile':
      this.handleOpenExpertFile(message.expertId);
      break;

    case 'createExpert':
      this.handleCreateExpert(message.expertId);
      break;

    case 'deleteExpert':
      this.handleDeleteExpert(message.expertId);
      break;
  }
}

private async handleAssignExpert(expertId: string) {
  // DB에 할당 저장
  // 결과 전송
}

private async handleOpenExpertFile(expertId: string) {
  // EXPERT.md 파일을 VSCode 에디터에서 열기
  const expert = this.findExpert(expertId);
  if (expert) {
    const uri = vscode.Uri.file(expert.path);
    await vscode.window.showTextDocument(uri);
  }
}
```

### 3. Webview 메시지 전송 함수
**파일**: `vscode-extension/webview-ui/src/vscode.ts`

```typescript
export function assignExpert(expertId: string) {
  vscode.postMessage({ type: 'assignExpert', expertId });
}

export function unassignExpert(expertId: string) {
  vscode.postMessage({ type: 'unassignExpert', expertId });
}

export function openExpertFile(expertId: string) {
  vscode.postMessage({ type: 'openExpertFile', expertId });
}

export function createExpert(expertId: string) {
  vscode.postMessage({ type: 'createExpert', expertId });
}

export function deleteExpert(expertId: string) {
  vscode.postMessage({ type: 'deleteExpert', expertId });
}
```

## 완료 기준
- [ ] 메시지 타입 정의
- [ ] CltEditorProvider에 핸들러 추가
- [ ] Webview 메시지 전송 함수 추가
- [ ] 양방향 통신 테스트
