package middleware

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int32

const (
	StateClosed   CircuitBreakerState = iota // 关闭状态（正常）
	StateOpen                                // 打开状态（熔断）
	StateHalfOpen                            // 半开状态（尝试恢复）
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	mu sync.RWMutex

	// 配置
	failureThreshold int           // 失败阈值
	resetTimeout     time.Duration // 重置超时时间
	successThreshold int           // 半开状态成功阈值

	// 状态
	state         int32     // 当前状态（原子操作）
	failureCount  int64     // 失败计数（原子操作）
	successCount  int64     // 成功计数（原子操作）
	lastFailTime  time.Time // 上次失败时间
	lastResetTime time.Time // 上次重置时间

	// 服务状态
	serviceStats map[string]*ServiceStats
}

// ServiceStats 服务统计
type ServiceStats struct {
	FailureCount  int64
	SuccessCount  int64
	LastFailTime  time.Time
	LastResetTime time.Time
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration, successThreshold int) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		successThreshold: successThreshold,
		state:            int32(StateClosed),
		serviceStats:     make(map[string]*ServiceStats),
	}
}

// Middleware 熔断器中间件
func (cb *CircuitBreaker) Middleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查熔断器状态
		state := CircuitBreakerState(atomic.LoadInt32(&cb.state))

		// 根据状态决定是否允许请求
		if state == StateOpen {
			// 检查是否可以尝试恢复
			if time.Since(cb.lastResetTime) >= cb.resetTimeout {
				// 转换为半开状态
				atomic.CompareAndSwapInt32(&cb.state, int32(StateOpen), int32(StateHalfOpen))
				atomic.StoreInt64(&cb.successCount, 0)
			} else {
				// 拒绝请求
				c.JSON(503, gin.H{
					"code":    503,
					"message": "服务暂时不可用（熔断器打开）",
					"data":    nil,
				})
				c.Abort()
				return
			}
		}

		// 执行请求
		c.Next()

		// 记录结果
		statusCode := c.Writer.Status()
		if statusCode >= 500 {
			cb.recordFailure(serviceName)
		} else {
			cb.recordSuccess(serviceName)
		}
	}
}

// recordFailure 记录失败
func (cb *CircuitBreaker) recordFailure(serviceName string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// 更新服务统计
	if cb.serviceStats[serviceName] == nil {
		cb.serviceStats[serviceName] = &ServiceStats{}
	}
	stats := cb.serviceStats[serviceName]
	stats.FailureCount++
	stats.LastFailTime = time.Now()

	// 更新全局计数
	failureCount := atomic.AddInt64(&cb.failureCount, 1)
	cb.lastFailTime = time.Now()

	// 检查是否达到阈值
	state := CircuitBreakerState(atomic.LoadInt32(&cb.state))
	if state == StateHalfOpen {
		// 半开状态：只要失败就立即打开
		atomic.StoreInt32(&cb.state, int32(StateOpen))
		atomic.StoreInt64(&cb.failureCount, 0)
		cb.lastResetTime = time.Now()
	} else if state == StateClosed && failureCount >= int64(cb.failureThreshold) {
		// 关闭状态：达到失败阈值，打开熔断器
		atomic.StoreInt32(&cb.state, int32(StateOpen))
		atomic.StoreInt64(&cb.failureCount, 0)
		cb.lastResetTime = time.Now()
	}
}

// recordSuccess 记录成功
func (cb *CircuitBreaker) recordSuccess(serviceName string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// 更新服务统计
	if cb.serviceStats[serviceName] == nil {
		cb.serviceStats[serviceName] = &ServiceStats{}
	}
	stats := cb.serviceStats[serviceName]
	stats.SuccessCount++

	// 更新全局计数
	state := CircuitBreakerState(atomic.LoadInt32(&cb.state))
	if state == StateHalfOpen {
		// 半开状态：检查是否达到成功阈值
		successCount := atomic.AddInt64(&cb.successCount, 1)
		if successCount >= int64(cb.successThreshold) {
			// 恢复为关闭状态
			atomic.StoreInt32(&cb.state, int32(StateClosed))
			atomic.StoreInt64(&cb.successCount, 0)
			atomic.StoreInt64(&cb.failureCount, 0)
		}
	} else if state == StateClosed {
		// 关闭状态：重置失败计数
		atomic.StoreInt64(&cb.failureCount, 0)
	}
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	return CircuitBreakerState(atomic.LoadInt32(&cb.state))
}

// GetStats 获取统计信息
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	state := CircuitBreakerState(atomic.LoadInt32(&cb.state))
	stateStr := "closed"
	switch state {
	case StateOpen:
		stateStr = "open"
	case StateHalfOpen:
		stateStr = "half-open"
	}

	return map[string]interface{}{
		"state":         stateStr,
		"failure_count": atomic.LoadInt64(&cb.failureCount),
		"success_count": atomic.LoadInt64(&cb.successCount),
		"service_stats": cb.serviceStats,
	}
}
