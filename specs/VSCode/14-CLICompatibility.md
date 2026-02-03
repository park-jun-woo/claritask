# VSCode Extension CLI 호환성

> **현재 버전**: v0.0.7 ([변경이력](../HISTORY.md))

---

## 확장자 변경 마이그레이션

```bash
# 기존 프로젝트 마이그레이션
mv .claritask/db .claritask/db.clt
```

---

## clari CLI 수정 사항

1. DB 경로 변경: `.claritask/db` → `.claritask/db.clt`
2. WAL 모드 기본 활성화
3. version 컬럼 마이그레이션 추가

---

## CLI 호출 아키텍처

VSCode Extension은 복잡한 작업 시 직접 DB 조작 대신 CLI를 호출합니다.

```
┌─────────────────────────────────────────────────────────┐
│  Webview (React)                                        │
│  ┌───────────────────────────────────────────────────┐  │
│  │  Feature 추가 버튼 클릭                            │  │
│  │       ↓                                           │  │
│  │  { type: 'createFeature', data: {...} }           │  │
│  └───────────────────────────────────────────────────┘  │
│                        ↓ postMessage                    │
├─────────────────────────────────────────────────────────┤
│  Extension Host                                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │  executeCLI('feature', 'create', jsonData)        │  │
│  │       ↓                                           │  │
│  │  child_process.spawn('clari', args)               │  │
│  │       ↓                                           │  │
│  │  JSON 응답 파싱 → Webview에 결과 전달              │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### CLI 실행 서비스

```typescript
// cliService.ts
import { spawn } from 'child_process';

interface CLIResult {
  success: boolean;
  data?: any;
  error?: string;
}

export async function executeCLI(
  command: string,
  subcommand: string,
  jsonArg?: object
): Promise<CLIResult> {
  return new Promise((resolve) => {
    const args = [command, subcommand];
    if (jsonArg) {
      args.push(JSON.stringify(jsonArg));
    }

    const proc = spawn('clari', args, {
      cwd: vscode.workspace.workspaceFolders?.[0].uri.fsPath
    });

    let stdout = '';
    let stderr = '';

    proc.stdout.on('data', (data) => { stdout += data; });
    proc.stderr.on('data', (data) => { stderr += data; });

    proc.on('close', (code) => {
      try {
        const result = JSON.parse(stdout);
        resolve(result);
      } catch {
        resolve({ success: false, error: stderr || 'CLI execution failed' });
      }
    });
  });
}
```

---

## CLI 호출 대상 명령어

| 작업 | CLI 명령어 | 직접 DB 조작 |
|------|-----------|-------------|
| Feature 생성 (통합) | `clari feature create` | ❌ |
| FDL 검증 | `clari fdl validate` | ❌ |
| Task 생성 | `clari fdl tasks` | ❌ |
| 스켈레톤 생성 | `clari fdl skeleton` | ❌ |
| Expert 생성 | `clari expert add` | ❌ |
| 단순 필드 수정 | - | ✅ |
| 상태 변경 | - | ✅ |

**원칙**: 비즈니스 로직이 있는 작업은 CLI 호출, 단순 CRUD는 직접 DB 조작

---

## 메시지 프로토콜 (CLI 호출)

### Webview → Extension

```typescript
// Feature 통합 생성 요청
{
  type: 'createFeature',
  data: {
    name: string,
    description: string,
    fdl?: string,
    generateTasks?: boolean,
    generateSkeleton?: boolean
  }
}

// FDL 검증 요청
{ type: 'validateFDL', featureId: number }

// Task 생성 요청
{ type: 'generateTasks', featureId: number }

// 스켈레톤 생성 요청
{ type: 'generateSkeleton', featureId: number, dryRun?: boolean }
```

### Extension → Webview

```typescript
// CLI 실행 결과
{
  type: 'cliResult',
  command: 'feature.create',
  success: boolean,
  data?: any,
  error?: string
}

// 진행 상태 (긴 작업용)
{
  type: 'cliProgress',
  command: 'feature.create',
  step: 'validating_fdl' | 'creating_tasks' | 'generating_skeleton',
  message: string
}
```

---

## Context/Tech/Design 편집

- JSON 에디터 또는 폼 기반 UI
- 스키마 검증

---

## TTY 세션 관리

VSCode Extension에서 Claude Code를 호출하는 CLI 명령어 실행 시 세션 관리 정책입니다.

### 설계 원칙

1. **clari CLI**: 무제한 실행 가능
2. **Claude Code 세션**: `max_parallel_sessions` 설정에 따라 제한 (기본값: 3)
3. **우선순위**: 먼저 시작된 세션이 우선권 (FIFO)
4. **대기 정책**: 제한 초과 시 무한 대기 (앞선 세션 완료 시 자동 시작)

### 세션 관리 아키텍처

```
┌─────────────────────────────────────────────────────────────┐
│  VSCode Extension Host                                      │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  TTYSessionManager                                    │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │  activeSessions: Map<string, Terminal>          │  │  │
│  │  │  waitingQueue: Queue<PendingSession>            │  │  │
│  │  │  maxParallel: number (from config.yaml)         │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────┘  │
│                           │                                  │
│      ┌────────────────────┼────────────────────┐            │
│      ▼                    ▼                    ▼            │
│  [Terminal 1]        [Terminal 2]        [Terminal 3]       │
│  Claude Code         Claude Code         Claude Code        │
│  (feature fdl 1)     (task run 5)        (feature fdl 2)    │
│                                                             │
│  [Waiting Queue]                                            │
│  → Session 4 (feature fdl 3) - 대기 중                       │
│  → Session 5 (task run 7) - 대기 중                          │
└─────────────────────────────────────────────────────────────┘
```

### TTYSessionManager 구현

```typescript
// ttySessionManager.ts
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

  constructor(workspacePath: string) {
    const config = loadConfig(workspacePath);
    this.maxParallel = config.tty.max_parallel_sessions;
    this.terminalCloseDelay = config.tty.terminal_close_delay;

    // 터미널 종료 감지
    vscode.window.onDidCloseTerminal(terminal => {
      this.onTerminalClosed(terminal);
    });
  }

  async startSession(id: string, command: string): Promise<vscode.Terminal> {
    // 현재 활성 세션이 제한 미만이면 즉시 시작
    if (this.activeSessions.size < this.maxParallel) {
      return this.createTerminal(id, command);
    }

    // 제한 초과: 대기열에 추가
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
      shellPath: '/bin/bash',
    });

    this.activeSessions.set(id, terminal);
    terminal.show();
    terminal.sendText(command);

    return terminal;
  }

  private onTerminalClosed(terminal: vscode.Terminal): void {
    // 활성 세션에서 제거
    for (const [id, t] of this.activeSessions) {
      if (t === terminal) {
        this.activeSessions.delete(id);
        break;
      }
    }

    // 대기열에서 다음 세션 시작
    this.processQueue();
  }

  private processQueue(): void {
    if (this.waitingQueue.length === 0) return;
    if (this.activeSessions.size >= this.maxParallel) return;

    const next = this.waitingQueue.shift()!;
    const terminal = this.createTerminal(next.id, next.command);
    next.resolve(terminal);
  }

  // 현재 상태 조회
  getStatus(): { active: number; waiting: number; max: number } {
    return {
      active: this.activeSessions.size,
      waiting: this.waitingQueue.length,
      max: this.maxParallel,
    };
  }

  // 대기 중인 세션 취소
  cancelWaiting(id: string): boolean {
    const index = this.waitingQueue.findIndex(s => s.id === id);
    if (index >= 0) {
      const session = this.waitingQueue.splice(index, 1)[0];
      session.reject(new Error('Session cancelled'));
      return true;
    }
    return false;
  }
}
```

### 사용 예시

```typescript
// CltEditorProvider.ts
private ttyManager: TTYSessionManager;

constructor(context: vscode.ExtensionContext) {
  const workspacePath = vscode.workspace.workspaceFolders?.[0].uri.fsPath || '';
  this.ttyManager = new TTYSessionManager(workspacePath);
}

private async handleCreateFeatureFDL(featureId: number, featureName: string) {
  const wslPath = windowsToWslPath(this.workspacePath);
  const command = `cd "${wslPath}" && clari feature fdl ${featureId}`;

  try {
    const terminal = await this.ttyManager.startSession(
      `fdl-${featureId}`,
      command
    );
    // 세션 시작됨
  } catch (error) {
    // 세션 취소됨 또는 에러
    vscode.window.showErrorMessage(`세션 시작 실패: ${error}`);
  }
}
```

### 설정

`.claritask/config.yaml`:

```yaml
tty:
  max_parallel_sessions: 3  # 동시 실행 가능한 Claude Code 세션 수
  terminal_close_delay: 1   # 완료 후 터미널 종료 대기 시간 (초)
```

### UI 표시

StatusBar에 현재 세션 상태 표시:

```
[Claude: 2/3 실행 중, 1 대기]
```

---

## 관련 문서

| 문서 | 내용 |
|------|------|
| [CLI/16-Config.md](../CLI/16-Config.md) | Config 설정 파일 |
| [TTY/01-Overview.md](../TTY/01-Overview.md) | TTY Handover 아키텍처 |

---

*Claritask VSCode Extension Spec v0.0.7*
