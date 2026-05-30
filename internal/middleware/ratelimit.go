package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter holds per-IP token bucket limiters with automatic cleanup.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	r        rate.Limit
	burst    int
}

func newRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		r:        r,
		burst:    burst,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) get(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	v, ok := rl.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(rl.r, rl.burst)}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanupLoop() {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for range t.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns an Echo middleware that blocks IPs exceeding the rate limit.
func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if !rl.get(c.RealIP()).Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "too many requests — please slow down")
			}
			return next(c)
		}
	}
}

// LoginRateLimiter returns a strict limiter for authentication endpoints.
// Allows 5 requests/min per IP (1 token every 12s), burst of 3.
func LoginRateLimiter() *RateLimiter {
	return newRateLimiter(rate.Every(12*time.Second), 3)
}

// APIRateLimiter returns a general limiter for all API endpoints.
// Allows 120 requests/min per IP (2/s), burst of 20.
func APIRateLimiter() *RateLimiter {
	return newRateLimiter(rate.Every(500*time.Millisecond), 20)
}
