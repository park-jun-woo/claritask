# TASK-BOT-007: Rate Limiter

## 목표
사용자별 요청 제한 구현

## 파일
`bot/internal/bot/ratelimit.go`

## 작업 내용

### RateLimiter 구조체
```go
type RateLimiter struct {
    mu       sync.Mutex
    limiters map[int64]*rate.Limiter
    rate     rate.Limit
    burst    int
}
```

### 메서드
```go
func NewRateLimiter(r float64, burst int) *RateLimiter
func (rl *RateLimiter) Allow(userID int64) bool
func (rl *RateLimiter) Cleanup() // 오래된 limiter 정리
```

### 의존성
```go
import "golang.org/x/time/rate"
```

## 완료 조건
- [ ] RateLimiter 구조체 정의
- [ ] Allow 메서드 구현
- [ ] 정리 로직 구현
