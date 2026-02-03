# Claribot Security

> **현재 버전**: v0.0.1

---

## 보안 원칙

1. **최소 권한**: 필요한 최소한의 권한만 부여
2. **허용 목록**: 허용된 사용자만 접근 가능
3. **토큰 보호**: API 토큰 노출 방지
4. **감사 로그**: 모든 명령 실행 기록

---

## 인증 체계

### 사용자 허용 목록

```go
// internal/config/config.go
type Config struct {
    AllowedUsers []int64 `env:"ALLOWED_USERS"`  // 일반 사용자
    AdminUsers   []int64 `env:"ADMIN_USERS"`    // 관리자
}
```

### 권한 레벨

| 레벨 | 대상 | 권한 |
|------|------|------|
| Admin | ADMIN_USERS | 모든 기능 + 설정 변경 |
| User | ALLOWED_USERS | 조회 + 일부 수정 |
| Guest | 기타 | 접근 불가 |

### 미들웨어 구현

```go
// internal/bot/middleware.go
func AuthMiddleware(cfg *config.Config) telebot.MiddlewareFunc {
    return func(next telebot.HandlerFunc) telebot.HandlerFunc {
        return func(c telebot.Context) error {
            userID := c.Sender().ID

            // 허용 목록 확인
            if !isAllowed(userID, cfg.AllowedUsers, cfg.AdminUsers) {
                // 로그 기록 (의심스러운 접근)
                log.Warn().
                    Int64("user_id", userID).
                    Str("username", c.Sender().Username).
                    Msg("unauthorized access attempt")

                return c.Send("❌ 권한이 없습니다.")
            }

            // 컨텍스트에 권한 정보 추가
            c.Set("is_admin", isAdmin(userID, cfg.AdminUsers))

            return next(c)
        }
    }
}

func AdminOnlyMiddleware() telebot.MiddlewareFunc {
    return func(next telebot.HandlerFunc) telebot.HandlerFunc {
        return func(c telebot.Context) error {
            isAdmin, _ := c.Get("is_admin").(bool)
            if !isAdmin {
                return c.Send("❌ 관리자 권한이 필요합니다.")
            }
            return next(c)
        }
    }
}
```

---

## 토큰 관리

### 환경 변수 사용

```bash
# 올바른 방법
export TELEGRAM_TOKEN=your-token
./claribot

# 또는 환경 파일
EnvironmentFile=/etc/claribot/claribot.env
```

### 절대 하지 말 것

```go
// ❌ 코드에 토큰 하드코딩 금지
const token = "123456789:ABCdef..."

// ❌ 명령줄 인자로 전달 금지 (ps에 노출)
./claribot --token=123456789:ABCdef...

// ❌ 로그에 토큰 출력 금지
log.Info("token: %s", token)
```

### 토큰 로테이션

1. BotFather에서 `/revoke` 실행
2. 새 토큰 발급
3. 환경 파일 업데이트
4. 서비스 재시작

```bash
# 토큰 변경 시
sudo vim /etc/claribot/claribot.env
sudo systemctl restart claribot
```

---

## 명령어별 권한

### 권한 매트릭스

| 명령어 | User | Admin |
|--------|------|-------|
| /start, /help | ✓ | ✓ |
| /project list, status | ✓ | ✓ |
| /project switch | ✓ | ✓ |
| /task list, get | ✓ | ✓ |
| /task add | ✓ | ✓ |
| /task start, done, fail | ✓ | ✓ |
| /msg list, get | ✓ | ✓ |
| /msg send | ✓ | ✓ |
| /settings notify | ✓ | ✓ |
| /settings admin | ✗ | ✓ |
| 사용자 추가/제거 | ✗ | ✓ |

### 구현 예시

```go
// internal/bot/commands.go
func SetupCommands(b *telebot.Bot, h *Handlers, cfg *config.Config) {
    // 공통 미들웨어
    authMw := AuthMiddleware(cfg)
    adminMw := AdminOnlyMiddleware()

    // 일반 명령어 (User 이상)
    b.Handle("/start", h.Start, authMw)
    b.Handle("/project", h.Project, authMw)
    b.Handle("/task", h.Task, authMw)

    // 관리자 전용
    b.Handle("/admin", h.Admin, authMw, adminMw)
}
```

---

## 입력 검증

### 명령어 인자 검증

```go
// internal/bot/validation.go
func ValidateProjectID(id string) error {
    // 영문, 숫자, 하이픈, 언더스코어만 허용
    matched, _ := regexp.MatchString(`^[a-z0-9_-]+$`, id)
    if !matched {
        return errors.New("invalid project ID format")
    }
    if len(id) > 50 {
        return errors.New("project ID too long")
    }
    return nil
}

func ValidateTaskID(id string) error {
    _, err := strconv.Atoi(id)
    if err != nil {
        return errors.New("task ID must be a number")
    }
    return nil
}
```

### SQL Injection 방지

```go
// ✓ 올바른 방법: Prepared Statement
func (s *TaskService) GetByID(id int) (*Task, error) {
    row := s.db.QueryRow("SELECT * FROM tasks WHERE id = ?", id)
    // ...
}

// ❌ 절대 금지: 문자열 연결
func (s *TaskService) GetByID(id string) (*Task, error) {
    query := "SELECT * FROM tasks WHERE id = " + id  // 위험!
    // ...
}
```

### 메시지 내용 검증

```go
func ValidateMessageContent(content string) error {
    // 길이 제한
    if len(content) > 4096 {  // 텔레그램 메시지 제한
        return errors.New("message too long")
    }
    // 빈 내용 방지
    if strings.TrimSpace(content) == "" {
        return errors.New("message cannot be empty")
    }
    return nil
}
```

---

## 감사 로그

### 로그 구조

```go
// internal/bot/audit.go
type AuditLog struct {
    Timestamp time.Time `json:"timestamp"`
    UserID    int64     `json:"user_id"`
    Username  string    `json:"username"`
    Command   string    `json:"command"`
    Args      []string  `json:"args"`
    Result    string    `json:"result"`
    Error     string    `json:"error,omitempty"`
}

func LogCommand(c telebot.Context, result string, err error) {
    log := AuditLog{
        Timestamp: time.Now(),
        UserID:    c.Sender().ID,
        Username:  c.Sender().Username,
        Command:   c.Text(),
        Result:    result,
    }
    if err != nil {
        log.Error = err.Error()
    }

    // 구조화된 로그 출력
    zerolog.Info().
        Int64("user_id", log.UserID).
        Str("username", log.Username).
        Str("command", log.Command).
        Str("result", log.Result).
        Err(err).
        Msg("command executed")
}
```

### 로그 미들웨어

```go
func AuditMiddleware() telebot.MiddlewareFunc {
    return func(next telebot.HandlerFunc) telebot.HandlerFunc {
        return func(c telebot.Context) error {
            start := time.Now()

            err := next(c)

            duration := time.Since(start)
            result := "success"
            if err != nil {
                result = "error"
            }

            LogCommand(c, result, err)

            // 메트릭 기록 (선택적)
            commandDuration.WithLabelValues(c.Text()).Observe(duration.Seconds())

            return err
        }
    }
}
```

---

## Rate Limiting

### 사용자별 제한

```go
// internal/bot/ratelimit.go
type RateLimiter struct {
    mu       sync.Mutex
    limiters map[int64]*rate.Limiter
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
    return &RateLimiter{
        limiters: make(map[int64]*rate.Limiter),
        rate:     r,
        burst:    b,
    }
}

func (r *RateLimiter) Allow(userID int64) bool {
    r.mu.Lock()
    defer r.mu.Unlock()

    limiter, exists := r.limiters[userID]
    if !exists {
        limiter = rate.NewLimiter(r.rate, r.burst)
        r.limiters[userID] = limiter
    }

    return limiter.Allow()
}
```

### 미들웨어 적용

```go
func RateLimitMiddleware(rl *RateLimiter) telebot.MiddlewareFunc {
    return func(next telebot.HandlerFunc) telebot.HandlerFunc {
        return func(c telebot.Context) error {
            if !rl.Allow(c.Sender().ID) {
                return c.Send("⏳ 잠시 후 다시 시도해주세요.")
            }
            return next(c)
        }
    }
}
```

### 권장 제한값

| 대상 | Rate | Burst | 설명 |
|------|------|-------|------|
| 일반 사용자 | 1/sec | 5 | 초당 1개, 버스트 5개 |
| 관리자 | 5/sec | 10 | 초당 5개, 버스트 10개 |
| 전체 | 30/sec | 50 | 텔레그램 API 제한 준수 |

---

## 네트워크 보안

### systemd 보안 옵션

```ini
[Service]
# 새 권한 획득 방지
NoNewPrivileges=true

# 파일시스템 보호
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/home/youruser/.claritask

# 네트워크 제한 (텔레그램 API만)
# IPAddressAllow를 사용하면 DNS가 안되므로 주의

# 개인 temp 디렉토리
PrivateTmp=true

# 커널 모듈 로드 방지
ProtectKernelModules=true

# 커널 튜닝 방지
ProtectKernelTunables=true
```

### Docker 보안

```yaml
services:
  claribot:
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp
    cap_drop:
      - ALL
```

---

## 비상 대응

### 봇 즉시 중지

```bash
# systemctl
sudo systemctl stop claribot

# Docker
docker stop claribot
```

### 토큰 무효화

1. @BotFather 접속
2. `/mybots` 선택
3. 해당 봇 선택
4. `API Token` → `Revoke current token`

### 의심 활동 확인

```bash
# 최근 로그 확인
sudo journalctl -u claribot --since "1 hour ago" | grep -i "unauthorized\|error\|fail"

# 특정 사용자 활동
sudo journalctl -u claribot | grep "user_id=123456"
```

---

## 보안 체크리스트

### 배포 전

- [ ] 토큰이 환경 변수로 설정됨
- [ ] 환경 파일 권한 600
- [ ] ALLOWED_USERS 설정됨
- [ ] ADMIN_USERS 설정됨
- [ ] 테스트 환경에서 권한 검증 완료

### 운영 중

- [ ] 정기적인 로그 검토 (주 1회)
- [ ] 토큰 로테이션 (분기 1회)
- [ ] 허용 사용자 목록 검토 (월 1회)
- [ ] 의존성 보안 업데이트

### 인시던트 발생 시

- [ ] 봇 즉시 중지
- [ ] 토큰 무효화
- [ ] 로그 분석
- [ ] 원인 파악 및 조치
- [ ] 새 토큰으로 재시작

---

## 관련 문서

| 문서 | 내용 |
|------|------|
| [01-Overview.md](01-Overview.md) | 전체 개요 |
| [04-Deployment.md](04-Deployment.md) | 배포 |

---

*Claribot Security v0.0.1*
