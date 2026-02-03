# TASK-EXT-034: VSCode Terminal CLI 호출

## 목표
VSCode Extension에서 Feature 생성 시 Terminal을 통해 CLI 호출 (TTY Handover 지원)

## 변경 파일
- `vscode-extension/src/CltEditorProvider.ts`
- `vscode-extension/webview-ui/src/components/FeatureList.tsx` (또는 관련 컴포넌트)

## 작업 내용

### 1. CltEditorProvider.ts 수정

`handleCreateFeature` 수정 - spawn 대신 Terminal 사용:

```typescript
private async handleCreateFeature(
  message: { name: string; description: string },
  database: Database,
  webview: vscode.Webview,
  sync: SyncManager
): Promise<void> {
  const input = JSON.stringify({
    name: message.name,
    description: message.description
  });

  // Terminal을 사용하여 CLI 실행 (TTY Handover 지원)
  const terminal = vscode.window.createTerminal({
    name: 'Claritask - Create Feature',
    cwd: vscode.workspace.workspaceFolders?.[0].uri.fsPath
  });
  terminal.show();
  terminal.sendText(`clari feature add '${input.replace(/'/g, "'\\''")}'`);

  // 알림 전송
  webview.postMessage({
    type: 'cliStarted',
    command: 'feature.add',
    message: 'Claude Code will generate FDL for the feature...'
  });
}
```

### 2. Webview 수정

Feature 생성 버튼 클릭 시:
- 다이얼로그에서 name, description만 입력
- FDL 입력 필드 제거
- "Claude Code가 FDL을 생성합니다" 안내 메시지 표시

### 3. FDL 파일 Watcher 추가

```typescript
// features/*.fdl.yaml 파일 감시
const fdlWatcher = vscode.workspace.createFileSystemWatcher(
  '**/features/*.fdl.yaml'
);

fdlWatcher.onDidChange(uri => {
  // FDL 파일 변경 시 DB 동기화
  syncFDLToDB(uri, database);
  sync.refresh();
});
```

## 테스트
- VSCode에서 Feature 추가 버튼 클릭
- Terminal이 열리고 `clari feature add` 명령어 실행 확인
- Claude Code가 FDL 생성 후 파일 저장 확인
- DB 자동 동기화 확인
