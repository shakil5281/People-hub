package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count     int
	windowEnd time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (r *rateLimiter) cleanup() {
	for {
		time.Sleep(r.window)
		r.mu.Lock()
		now := time.Now()
		for ip, v := range r.visitors {
			if now.After(v.windowEnd) {
				delete(r.visitors, ip)
			}
		}
		r.mu.Unlock()
	}
}

func (r *rateLimiter) allow(ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	v, exists := r.visitors[ip]
	if !exists || now.After(v.windowEnd) {
		r.visitors[ip] = &visitor{count: 1, windowEnd: now.Add(r.window)}
		return true
	}
	if v.count >= r.limit {
		return false
	}
	v.count++
	return true
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	rl := newRateLimiter(limit, window)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.allow(ip) {
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests, please try again later."})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AuthRateLimit limits sensitive auth endpoints to prevent brute force.
// Default: 5 requests per minute per IP for login/forgot, 10 per minute for refresh.
func AuthRateLimit() gin.HandlerFunc {
	return RateLimit(5, time.Minute)
}
