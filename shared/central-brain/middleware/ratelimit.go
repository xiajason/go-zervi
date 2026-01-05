package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter 限流器
type RateLimiter struct {
	limiter *rate.Limiter
	enabled bool
}

// NewRateLimiter 创建限流器
// rps: 每秒请求数
// burst: 突发请求数
func NewRateLimiter(rps int, burst int, enabled bool) *RateLimiter {
	if !enabled {
		return &RateLimiter{enabled: false}
	}

	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
		enabled: enabled,
	}
}

// Middleware 限流中间件
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.enabled {
			c.Next()
			return
		}

		// 检查是否允许请求
		if !rl.limiter.Allow() {
			c.JSON(429, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPRateLimiter IP限流器（每个IP独立限流）
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      int
	burst    int
	enabled  bool
}

// NewIPRateLimiter 创建IP限流器
func NewIPRateLimiter(rps int, burst int, enabled bool) *IPRateLimiter {
	if !enabled {
		return &IPRateLimiter{enabled: false}
	}

	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
		enabled:  enabled,
	}
}

// getLimiter 获取IP对应的限流器
func (rl *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		defer rl.mu.Unlock()

		// 双重检查
		limiter, exists = rl.limiters[ip]
		if !exists {
			limiter = rate.NewLimiter(rate.Limit(rl.rps), rl.burst)
			rl.limiters[ip] = limiter
		}
	}

	return limiter
}

// Middleware IP限流中间件
func (rl *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.enabled {
			c.Next()
			return
		}

		ip := c.ClientIP()
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			c.JSON(429, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
