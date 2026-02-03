# TASK-EXT-037: VSCode Config Service 구현

## 목표
VSCode Extension에서 `.claritask/config.yaml` 로드

## 변경 파일
- `vscode-extension/src/configService.ts` (신규)
- `vscode-extension/package.json` (yaml 의존성)

## 작업 내용

### 1. 타입 정의
```typescript
export interface TTYConfig {
  max_parallel_sessions: number;
  terminal_close_delay: number;
}

export interface VSCodeConfig {
  sync_interval: number;
  watch_feature_files: boolean;
}

export interface ClaritaskConfig {
  tty: TTYConfig;
  vscode: VSCodeConfig;
}
```

### 2. 기본값
```typescript
export const DEFAULT_CONFIG: ClaritaskConfig = {
  tty: {
    max_parallel_sessions: 3,
    terminal_close_delay: 1,
  },
  vscode: {
    sync_interval: 1000,
    watch_feature_files: true,
  },
};
```

### 3. loadConfig 함수
```typescript
import * as yaml from 'yaml';
import * as fs from 'fs';
import * as path from 'path';

export function loadConfig(workspacePath: string): ClaritaskConfig {
  const configPath = path.join(workspacePath, '.claritask', 'config.yaml');

  try {
    const content = fs.readFileSync(configPath, 'utf8');
    const loaded = yaml.parse(content) as Partial<ClaritaskConfig>;

    return {
      tty: { ...DEFAULT_CONFIG.tty, ...loaded?.tty },
      vscode: { ...DEFAULT_CONFIG.vscode, ...loaded?.vscode },
    };
  } catch {
    return DEFAULT_CONFIG;
  }
}
```

### 4. 의존성 추가
```json
{
  "dependencies": {
    "yaml": "^2.3.4"
  }
}
```

## 테스트
- 설정 파일 없을 때 기본값 반환
- 설정 파일 부분 오버라이드 동작

## 참고
- specs/CLI/16-Config.md
