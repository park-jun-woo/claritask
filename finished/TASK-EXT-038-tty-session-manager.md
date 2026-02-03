# TASK-EXT-038: TTY Session Manager 구현

## 목표
VSCode에서 Claude Code TTY 세션 수 제한 및 대기열 관리

## 변경 파일
- `vscode-extension/src/ttySessionManager.ts` (신규)
- `vscode-extension/src/CltEditorProvider.ts` (TTYSessionManager 사용)

## 작업 내용

### 1. TTYSessionManager 클래스
```typescript
import * as vscode from 'vscode';
import { loadConfig } from './configService';

interface PendingSession {
  id: string;
  command: string;
  resolve: (terminal: vscode.Terminal) => void;
  reject: (error: Error) => void;
}

export class TTYSessionManager {
  private activeSessions: Map<string, vscode.Terminal> = new Map();
  private waitingQueue: PendingSession[] = [];
  private maxParallel: number;
  private terminalCloseDelay: number;
  private disposables: vscode.Disposable[] = [];

  constructor(workspacePath: string) {
    const config = loadConfig(workspacePath);
    this.maxParallel = config.tty.max_parallel_sessions;
    this.terminalCloseDelay = config.tty.terminal_close_delay;

    // 터미널 종료 감지
    this.disposables.push(
      vscode.window.onDidCloseTerminal(terminal => {
        this.onTerminalClosed(terminal);
      })
    );
  }

  async startSession(id: string, command: string): Promise<vscode.Terminal> {
    if (this.activeSessions.size < this.maxParallel) {
      return this.createTerminal(id, command);
    }

    vscode.window.showInformationMessage(
      `Claude Code 세션 대기 중... (${this.activeSessions.size}/${this.maxParallel} 실행 중)`
    );

    return new Promise((resolve, reject) => {
      this.waitingQueue.push({ id, command, resolve, reject });
    });
  }

  private createTerminal(id: string, command: string): vscode.Terminal {
    const terminal = vscode.window.createTerminal({
      name: `Claritask: ${id}`,
    });

    this.activeSessions.set(id, terminal);
    terminal.show();
    terminal.sendText(command);

    return terminal;
  }

  private onTerminalClosed(terminal: vscode.Terminal): void {
    for (const [id, t] of this.activeSessions) {
      if (t === terminal) {
        this.activeSessions.delete(id);
        break;
      }
    }
    this.processQueue();
  }

  private processQueue(): void {
    if (this.waitingQueue.length === 0) return;
    if (this.activeSessions.size >= this.maxParallel) return;

    const next = this.waitingQueue.shift()!;
    const terminal = this.createTerminal(next.id, next.command);
    next.resolve(terminal);
  }

  getStatus(): { active: number; waiting: number; max: number } {
    return {
      active: this.activeSessions.size,
      waiting: this.waitingQueue.length,
      max: this.maxParallel,
    };
  }

  cancelWaiting(id: string): boolean {
    const index = this.waitingQueue.findIndex(s => s.id === id);
    if (index >= 0) {
      const session = this.waitingQueue.splice(index, 1)[0];
      session.reject(new Error('Session cancelled'));
      return true;
    }
    return false;
  }

  dispose(): void {
    this.disposables.forEach(d => d.dispose());
  }
}
```

### 2. CltEditorProvider 수정
```typescript
// 생성자에 추가
private ttyManager: TTYSessionManager;

constructor(context: vscode.ExtensionContext) {
  // ...
  const workspacePath = vscode.workspace.workspaceFolders?.[0].uri.fsPath || '';
  this.ttyManager = new TTYSessionManager(workspacePath);
}

// CLI 호출 메서드 수정
private async runCLIWithTTY(id: string, command: string) {
  try {
    await this.ttyManager.startSession(id, command);
  } catch (error) {
    vscode.window.showErrorMessage(`세션 시작 실패: ${error}`);
  }
}
```

## 테스트
- 3개 세션 실행 중 4번째 세션 대기 확인
- 세션 종료 시 대기 세션 자동 시작 확인
- 대기 세션 취소 기능 확인

## 참고
- specs/VSCode/14-CLICompatibility.md
