# TASK-EXT-030: VSCode CLI Service 구현

## 목표
Extension에서 clari CLI를 호출하는 서비스 구현

## 변경 파일
- `vscode-extension/src/cliService.ts` (신규)
- `vscode-extension/src/extension.ts`

## 작업 내용

### 1. cliService.ts 생성
```typescript
import * as vscode from 'vscode';
import { spawn } from 'child_process';

export interface CLIResult {
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

    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      resolve({ success: false, error: 'No workspace folder' });
      return;
    }

    const proc = spawn('clari', args, {
      cwd: workspaceFolder.uri.fsPath
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
        resolve({
          success: false,
          error: stderr || stdout || 'CLI execution failed'
        });
      }
    });

    proc.on('error', (err) => {
      resolve({ success: false, error: err.message });
    });
  });
}

// 편의 함수들
export const createFeature = (data: any) => executeCLI('feature', 'create', data);
export const validateFDL = (featureId: number) => executeCLI('fdl', 'validate', { id: featureId });
export const generateTasks = (featureId: number) => executeCLI('fdl', 'tasks', { id: featureId });
export const generateSkeleton = (featureId: number) => executeCLI('fdl', 'skeleton', { id: featureId });
```

### 2. extension.ts에서 import 및 사용
메시지 핸들러에서 CLI 호출 처리

## 테스트
- Extension에서 Feature 생성 시 CLI 호출 확인
- 에러 핸들링 확인

## 관련 스펙
- specs/VSCode/14-CLICompatibility.md (v0.0.6)
