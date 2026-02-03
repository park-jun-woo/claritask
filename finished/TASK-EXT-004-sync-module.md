# TASK-EXT-004: Sync 모듈 (Polling 로직)

## 목표
1초 polling으로 DB 변경 감지 및 Webview 동기화.

## 파일
`src/sync.ts`

## 구현 내용

```typescript
import * as fs from 'fs';
import * as vscode from 'vscode';
import { Database, ProjectData } from './database';

export class SyncManager {
  private intervalId: NodeJS.Timeout | null = null;
  private lastMtime: number = 0;
  private dbPath: string;

  constructor(
    private database: Database,
    private webview: vscode.Webview
  ) {
    this.dbPath = (database as any).db.name; // better-sqlite3 내부 경로
  }

  start(): void {
    // 초기 데이터 전송
    this.sendFullSync();

    // 1초마다 polling
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

  // 수동 새로고침
  refresh(): void {
    this.sendFullSync();
  }

  // 부분 업데이트 알림
  notifyUpdate(table: string, id: number): void {
    this.sendFullSync(); // MVP에서는 전체 동기화
  }
}
```

## 메시지 타입

```typescript
// Extension → Webview
interface SyncMessage {
  type: 'sync';
  data: ProjectData;
  timestamp: number;
}

interface ErrorMessage {
  type: 'error';
  message: string;
}

interface ConflictMessage {
  type: 'conflict';
  table: string;
  id: number;
}
```

## 완료 조건
- [ ] SyncManager 클래스 구현
- [ ] 1초 polling 구현
- [ ] 파일 mtime 변경 감지
- [ ] 전체 데이터 동기화
- [ ] start/stop 생명주기 관리
