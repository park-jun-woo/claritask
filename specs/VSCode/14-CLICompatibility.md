# VSCode Extension CLI 호환성

> **현재 버전**: v0.0.6 ([변경이력](../HISTORY.md))

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

*Claritask VSCode Extension Spec v0.0.6*
