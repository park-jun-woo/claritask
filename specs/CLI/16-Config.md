# Config 설정

> **현재 버전**: v0.0.1 ([변경이력](../HISTORY.md))

---

## 개요

Claritask 런타임 설정을 관리하는 YAML 설정 파일입니다.

**파일 위치**: `.claritask/config.yaml`

---

## 설정 파일 구조

```yaml
# .claritask/config.yaml

# Claude Code TTY 세션 설정
tty:
  # 동시 실행 가능한 Claude Code 세션 최대 개수
  max_parallel_sessions: 3

  # 세션 완료 후 터미널 자동 종료 대기 시간 (초)
  # 0: 즉시 종료, -1: 자동 종료 안함
  terminal_close_delay: 1

# VSCode Extension 설정
vscode:
  # DB 변경 감지 폴링 간격 (ms)
  sync_interval: 1000

  # Feature 파일 자동 감시
  watch_feature_files: true
```

---

## 설정 항목

### tty.max_parallel_sessions

| 속성 | 값 |
|------|-----|
| 타입 | integer |
| 기본값 | 3 |
| 최소값 | 1 |
| 최대값 | 10 |

동시에 실행 가능한 Claude Code TTY 세션 수를 제한합니다.

**동작 방식**:
- clari CLI 자체는 무제한 실행 가능
- Claude Code (`claude` 명령) 실행만 제한
- 먼저 시작된 세션이 우선권 (FIFO)
- 제한 초과 시 대기 (무한 대기, 취소 가능)

**예시 시나리오**:
```
[세션 1] clari feature fdl 1  → Claude Code 실행 중
[세션 2] clari feature fdl 2  → Claude Code 실행 중
[세션 3] clari feature fdl 3  → Claude Code 실행 중
[세션 4] clari task run 1     → 대기 중... (max_parallel_sessions=3)
[세션 1 완료]
[세션 4] clari task run 1     → Claude Code 실행 시작
```

### tty.terminal_close_delay

| 속성 | 값 |
|------|-----|
| 타입 | integer |
| 기본값 | 1 |
| 단위 | 초 |

Claude Code 세션 완료 후 터미널 자동 종료 전 대기 시간입니다.

| 값 | 동작 |
|----|------|
| 0 | 즉시 종료 |
| 1~N | N초 후 종료 (결과 확인 시간) |
| -1 | 자동 종료 안함 (수동 종료) |

### vscode.sync_interval

| 속성 | 값 |
|------|-----|
| 타입 | integer |
| 기본값 | 1000 |
| 단위 | ms |

VSCode Extension의 DB 변경 감지 폴링 간격입니다.

### vscode.watch_feature_files

| 속성 | 값 |
|------|-----|
| 타입 | boolean |
| 기본값 | true |

`features/*.fdl.yaml` 파일 변경 시 자동 동기화 여부입니다.

---

## 설정 파일 로딩

### 우선순위

1. `.claritask/config.yaml` (프로젝트별)
2. 기본값 (코드 내장)

### CLI에서 로딩

```go
// config_service.go
type Config struct {
    TTY struct {
        MaxParallelSessions int `yaml:"max_parallel_sessions"`
        TerminalCloseDelay  int `yaml:"terminal_close_delay"`
    } `yaml:"tty"`
    VSCode struct {
        SyncInterval      int  `yaml:"sync_interval"`
        WatchFeatureFiles bool `yaml:"watch_feature_files"`
    } `yaml:"vscode"`
}

func LoadConfig() (*Config, error) {
    config := &Config{
        TTY: struct {
            MaxParallelSessions int `yaml:"max_parallel_sessions"`
            TerminalCloseDelay  int `yaml:"terminal_close_delay"`
        }{
            MaxParallelSessions: 3,
            TerminalCloseDelay:  1,
        },
        VSCode: struct {
            SyncInterval      int  `yaml:"sync_interval"`
            WatchFeatureFiles bool `yaml:"watch_feature_files"`
        }{
            SyncInterval:      1000,
            WatchFeatureFiles: true,
        },
    }

    // Load from file if exists
    data, err := os.ReadFile(".claritask/config.yaml")
    if err == nil {
        yaml.Unmarshal(data, config)
    }

    return config, nil
}
```

### VSCode Extension에서 로딩

```typescript
// configService.ts
import * as yaml from 'yaml';
import * as fs from 'fs';
import * as path from 'path';

interface ClaritaskConfig {
  tty: {
    max_parallel_sessions: number;
    terminal_close_delay: number;
  };
  vscode: {
    sync_interval: number;
    watch_feature_files: boolean;
  };
}

const DEFAULT_CONFIG: ClaritaskConfig = {
  tty: {
    max_parallel_sessions: 3,
    terminal_close_delay: 1,
  },
  vscode: {
    sync_interval: 1000,
    watch_feature_files: true,
  },
};

export function loadConfig(workspacePath: string): ClaritaskConfig {
  const configPath = path.join(workspacePath, '.claritask', 'config.yaml');

  try {
    const content = fs.readFileSync(configPath, 'utf8');
    const loaded = yaml.parse(content);
    return { ...DEFAULT_CONFIG, ...loaded };
  } catch {
    return DEFAULT_CONFIG;
  }
}
```

---

## 관련 문서

| 문서 | 내용 |
|------|------|
| [VSCode/14-CLICompatibility.md](../VSCode/14-CLICompatibility.md) | TTY 세션 관리 |
| [TTY/01-Overview.md](../TTY/01-Overview.md) | TTY Handover 아키텍처 |

---

*Claritask Config Specification v0.0.1*
