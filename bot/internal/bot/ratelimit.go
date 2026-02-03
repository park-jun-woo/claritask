package bot

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter manages per-user rate limiting
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[int64]*userLimiter
	rate     rate.Limit
	burst    int
	ttl      time.Duration
}

type userLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(r float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[int64]*userLimiter),
		rate:     rate.Limit(r),
		burst:    burst,
		ttl:      time.Hour,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Allow checks if a request from a user should be allowed
func (rl *RateLimiter) Allow(userID int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	ul, exists := rl.limiters[userID]
	if !exists {
		ul = &userLimiter{
			limiter: rate.NewLimiter(rl.rate, rl.burst),
		}
		rl.limiters[userID] = ul
	}
	ul.lastSeen = time.Now()

	return ul.limiter.Allow()
}

// cleanupLoop periodically removes inactive limiters
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes limiters that haven't been used recently
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for userID, ul := range rl.limiters {
		if now.Sub(ul.lastSeen) > rl.ttl {
			delete(rl.limiters, userID)
		}
	}
}
