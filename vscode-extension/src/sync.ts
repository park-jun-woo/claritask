import * as fs from 'fs';
import * as vscode from 'vscode';
import { Database } from './database';

export class SyncManager {
  private intervalId: NodeJS.Timeout | null = null;
  private lastMtime: number = 0;

  constructor(
    private database: Database,
    private webview: vscode.Webview,
    private dbPath: string
  ) {}

  start(): void {
    this.sendFullSync();

    this.intervalId = setInterval(() => {
      this.checkForChanges();
    }, 1000);
  }

  stop(): void {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }
  }

  private checkForChanges(): void {
    try {
      const stat = fs.statSync(this.dbPath);
      const mtime = stat.mtimeMs;

      if (mtime !== this.lastMtime) {
        this.lastMtime = mtime;
        this.sendFullSync();
      }
    } catch (err) {
      console.error('Failed to check file changes:', err);
    }
  }

  private sendFullSync(): void {
    try {
      const data = this.database.readAll();
      this.webview.postMessage({
        type: 'sync',
        data,
        timestamp: Date.now(),
      });
    } catch (err) {
      console.error('Failed to sync:', err);
      this.webview.postMessage({
        type: 'error',
        message: 'Failed to read database',
      });
    }
  }

  refresh(): void {
    this.sendFullSync();
  }
}
