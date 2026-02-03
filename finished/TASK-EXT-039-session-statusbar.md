# TASK-EXT-039: 세션 상태 StatusBar 표시

## 목표
StatusBar에 현재 Claude Code 세션 상태 표시

## 변경 파일
- `vscode-extension/src/extension.ts`
- `vscode-extension/src/ttySessionManager.ts` (이벤트 추가)

## 작업 내용

### 1. StatusBar 아이템 생성
```typescript
// extension.ts
let sessionStatusBar: vscode.StatusBarItem;

export function activate(context: vscode.ExtensionContext) {
  // StatusBar 생성
  sessionStatusBar = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Right,
    100
  );
  sessionStatusBar.command = 'claritask.showSessionStatus';
  context.subscriptions.push(sessionStatusBar);

  // 초기 상태 표시
  updateSessionStatusBar(0, 0, 3);
}

function updateSessionStatusBar(active: number, waiting: number, max: number) {
  if (active === 0 && waiting === 0) {
    sessionStatusBar.hide();
    return;
  }

  sessionStatusBar.text = `$(terminal) Claude: ${active}/${max}`;
  if (waiting > 0) {
    sessionStatusBar.text += ` (${waiting} 대기)`;
  }
  sessionStatusBar.tooltip = `활성 세션: ${active}\n대기 중: ${waiting}\n최대: ${max}`;
  sessionStatusBar.show();
}
```

### 2. TTYSessionManager 이벤트 추가
```typescript
// ttySessionManager.ts
export class TTYSessionManager {
  private _onStatusChange = new vscode.EventEmitter<{
    active: number;
    waiting: number;
    max: number;
  }>();
  readonly onStatusChange = this._onStatusChange.event;

  private emitStatusChange(): void {
    this._onStatusChange.fire(this.getStatus());
  }

  // createTerminal, onTerminalClosed, processQueue에서 호출
}
```

### 3. 이벤트 연결
```typescript
// CltEditorProvider.ts
this.ttyManager.onStatusChange(status => {
  updateSessionStatusBar(status.active, status.waiting, status.max);
});
```

### 4. 세션 상태 명령어
```typescript
// extension.ts
context.subscriptions.push(
  vscode.commands.registerCommand('claritask.showSessionStatus', () => {
    const status = ttyManager.getStatus();
    vscode.window.showInformationMessage(
      `Claude Code 세션 상태\n` +
      `활성: ${status.active}/${status.max}\n` +
      `대기: ${status.waiting}`
    );
  })
);
```

## 표시 형식

| 상태 | StatusBar 표시 |
|------|---------------|
| 세션 없음 | (숨김) |
| 2개 실행 중 | `$(terminal) Claude: 2/3` |
| 3개 실행, 1개 대기 | `$(terminal) Claude: 3/3 (1 대기)` |

## 테스트
- 세션 시작 시 StatusBar 표시 확인
- 세션 종료 시 카운트 감소 확인
- 대기 세션 있을 때 표시 확인
- 모든 세션 종료 시 StatusBar 숨김 확인

## 참고
- specs/VSCode/14-CLICompatibility.md
