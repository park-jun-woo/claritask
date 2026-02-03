import * as vscode from 'vscode';
import { loadConfig, ClaritaskConfig } from './configService';

interface PendingSession {
  id: string;
  command: string;
  resolve: (terminal: vscode.Terminal) => void;
  reject: (error: Error) => void;
}

interface SessionStatus {
  active: number;
  waiting: number;
  max: number;
}

/**
 * Manages TTY sessions with parallel execution limits
 */
export class TTYSessionManager {
  private activeSessions: Map<string, vscode.Terminal> = new Map();
  private waitingQueue: PendingSession[] = [];
  private maxParallel: number;
  private terminalCloseDelay: number;
  private disposables: vscode.Disposable[] = [];

  private _onStatusChange = new vscode.EventEmitter<SessionStatus>();
  readonly onStatusChange = this._onStatusChange.event;

  constructor(private workspacePath: string) {
    const config = loadConfig(workspacePath);
    this.maxParallel = config.tty.max_parallel_sessions;
    this.terminalCloseDelay = config.tty.terminal_close_delay;

    // Watch for terminal close events
    this.disposables.push(
      vscode.window.onDidCloseTerminal((terminal) => {
        this.onTerminalClosed(terminal);
      })
    );
  }

  /**
   * Reload configuration
   */
  reloadConfig(): void {
    const config = loadConfig(this.workspacePath);
    this.maxParallel = config.tty.max_parallel_sessions;
    this.terminalCloseDelay = config.tty.terminal_close_delay;
  }

  /**
   * Start a new TTY session
   * If max parallel sessions reached, waits in queue
   */
  async startSession(id: string, command: string): Promise<vscode.Terminal> {
    // Check if session with this ID already exists
    if (this.activeSessions.has(id)) {
      const existing = this.activeSessions.get(id)!;
      existing.show();
      return existing;
    }

    // If under limit, start immediately
    if (this.activeSessions.size < this.maxParallel) {
      return this.createTerminal(id, command);
    }

    // Over limit: add to queue and wait
    vscode.window.showInformationMessage(
      `Claude Code 세션 대기 중... (${this.activeSessions.size}/${this.maxParallel} 실행 중)`
    );

    return new Promise((resolve, reject) => {
      this.waitingQueue.push({ id, command, resolve, reject });
      this.emitStatusChange();
    });
  }

  /**
   * Create and track a new terminal
   */
  private createTerminal(id: string, command: string): vscode.Terminal {
    const terminal = vscode.window.createTerminal({
      name: `Claritask: ${id}`,
    });

    this.activeSessions.set(id, terminal);
    terminal.show();
    terminal.sendText(command);

    this.emitStatusChange();

    return terminal;
  }

  /**
   * Handle terminal close event
   */
  private onTerminalClosed(terminal: vscode.Terminal): void {
    // Find and remove from active sessions
    for (const [id, t] of this.activeSessions) {
      if (t === terminal) {
        this.activeSessions.delete(id);
        break;
      }
    }

    this.emitStatusChange();

    // Process next in queue
    this.processQueue();
  }

  /**
   * Process the waiting queue
   */
  private processQueue(): void {
    if (this.waitingQueue.length === 0) {
      return;
    }
    if (this.activeSessions.size >= this.maxParallel) {
      return;
    }

    const next = this.waitingQueue.shift()!;
    const terminal = this.createTerminal(next.id, next.command);
    next.resolve(terminal);

    vscode.window.showInformationMessage(
      `대기 중이던 세션 시작: ${next.id}`
    );
  }

  /**
   * Emit status change event
   */
  private emitStatusChange(): void {
    this._onStatusChange.fire(this.getStatus());
  }

  /**
   * Get current session status
   */
  getStatus(): SessionStatus {
    return {
      active: this.activeSessions.size,
      waiting: this.waitingQueue.length,
      max: this.maxParallel,
    };
  }

  /**
   * Cancel a waiting session
   */
  cancelWaiting(id: string): boolean {
    const index = this.waitingQueue.findIndex((s) => s.id === id);
    if (index >= 0) {
      const session = this.waitingQueue.splice(index, 1)[0];
      session.reject(new Error('Session cancelled by user'));
      this.emitStatusChange();
      return true;
    }
    return false;
  }

  /**
   * Check if a session is active
   */
  isSessionActive(id: string): boolean {
    return this.activeSessions.has(id);
  }

  /**
   * Get active session IDs
   */
  getActiveSessionIds(): string[] {
    return Array.from(this.activeSessions.keys());
  }

  /**
   * Get waiting session IDs
   */
  getWaitingSessionIds(): string[] {
    return this.waitingQueue.map((s) => s.id);
  }

  /**
   * Dispose resources
   */
  dispose(): void {
    this.disposables.forEach((d) => d.dispose());
    this._onStatusChange.dispose();
  }
}
