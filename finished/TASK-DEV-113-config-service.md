# TASK-DEV-113: Config Service 구현

## 목표
`.claritask/config.yaml` 파일을 로드하는 서비스 구현

## 변경 파일
- `cli/internal/service/config_service.go` (신규)

## 작업 내용

### 1. Config 구조체 정의
```go
type TTYConfig struct {
    MaxParallelSessions int `yaml:"max_parallel_sessions"`
    TerminalCloseDelay  int `yaml:"terminal_close_delay"`
}

type VSCodeConfig struct {
    SyncInterval      int  `yaml:"sync_interval"`
    WatchFeatureFiles bool `yaml:"watch_feature_files"`
}

type Config struct {
    TTY    TTYConfig    `yaml:"tty"`
    VSCode VSCodeConfig `yaml:"vscode"`
}
```

### 2. 기본값 설정
```go
func DefaultConfig() *Config {
    return &Config{
        TTY: TTYConfig{
            MaxParallelSessions: 3,
            TerminalCloseDelay:  1,
        },
        VSCode: VSCodeConfig{
            SyncInterval:      1000,
            WatchFeatureFiles: true,
        },
    }
}
```

### 3. LoadConfig 함수
```go
func LoadConfig() (*Config, error) {
    config := DefaultConfig()

    data, err := os.ReadFile(".claritask/config.yaml")
    if err != nil {
        // 파일 없으면 기본값 반환
        return config, nil
    }

    if err := yaml.Unmarshal(data, config); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }

    // 유효성 검증
    if config.TTY.MaxParallelSessions < 1 {
        config.TTY.MaxParallelSessions = 1
    }
    if config.TTY.MaxParallelSessions > 10 {
        config.TTY.MaxParallelSessions = 10
    }

    return config, nil
}
```

### 4. 의존성 추가
```go
import "gopkg.in/yaml.v3"
```

## 테스트
- 설정 파일 없을 때 기본값 반환
- 설정 파일 있을 때 값 오버라이드
- 유효하지 않은 값 보정

## 참고
- specs/CLI/16-Config.md
