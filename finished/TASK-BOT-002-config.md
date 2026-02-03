# TASK-BOT-002: Config 모듈

## 목표
환경변수 기반 설정 관리 구현

## 파일
`bot/internal/config/config.go`

## 작업 내용

### Config 구조체
```go
type Config struct {
    // 텔레그램 설정
    TelegramToken string  `env:"TELEGRAM_TOKEN,required"`
    AllowedUsers  []int64 `env:"ALLOWED_USERS" envSeparator:","`
    AdminUsers    []int64 `env:"ADMIN_USERS" envSeparator:","`

    // DB 설정
    DBPath string `env:"CLARITASK_DB" envDefault:"~/.claritask/db.clt"`

    // 알림 설정
    NotifyOnComplete bool `env:"NOTIFY_ON_COMPLETE" envDefault:"true"`
    NotifyOnFail     bool `env:"NOTIFY_ON_FAIL" envDefault:"true"`

    // 로깅
    LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

    // Rate Limiting
    RateLimit float64 `env:"RATE_LIMIT" envDefault:"1"`
    RateBurst int     `env:"RATE_BURST" envDefault:"5"`
}
```

### 함수
- `Load() (*Config, error)` - 환경변수에서 설정 로드
- `(c *Config) IsAllowed(userID int64) bool` - 사용자 허용 여부
- `(c *Config) IsAdmin(userID int64) bool` - 관리자 여부
- `(c *Config) GetDBPath() string` - DB 경로 (~ 확장)

## 완료 조건
- [ ] Config 구조체 정의
- [ ] Load 함수 구현
- [ ] 권한 검사 함수 구현
