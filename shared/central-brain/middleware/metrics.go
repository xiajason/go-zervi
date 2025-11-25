package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics 性能指标收集器
type Metrics struct {
	mu sync.RWMutex

	// 请求统计
	totalRequests   int64
	successRequests int64
	failedRequests  int64

	// 响应时间统计
	totalDuration time.Duration
	minDuration   time.Duration
	maxDuration   time.Duration

	// 状态码统计
	statusCodes map[int]int64

	// 路径统计
	pathStats map[string]*PathStats
}

// PathStats 路径统计
type PathStats struct {
	Count       int64
	TotalTime   time.Duration
	MinTime     time.Duration
	MaxTime     time.Duration
	SuccessRate float64
}

// NewMetrics 创建指标收集器
func NewMetrics() *Metrics {
	return &Metrics{
		statusCodes: make(map[int]int64),
		pathStats:   make(map[string]*PathStats),
		minDuration: time.Hour, // 初始化为最大值
	}
}

// Middleware 指标收集中间件
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		// 记录指标
		m.recordMetrics(c.Request.URL.Path, statusCode, duration)
	}
}

// recordMetrics 记录指标
func (m *Metrics) recordMetrics(path string, statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 更新请求统计
	m.totalRequests++
	if statusCode >= 200 && statusCode < 400 {
		m.successRequests++
	} else {
		m.failedRequests++
	}

	// 更新响应时间统计
	m.totalDuration += duration
	if duration < m.minDuration {
		m.minDuration = duration
	}
	if duration > m.maxDuration {
		m.maxDuration = duration
	}

	// 更新状态码统计
	m.statusCodes[statusCode]++

	// 更新路径统计
	if m.pathStats[path] == nil {
		m.pathStats[path] = &PathStats{
			MinTime: duration,
			MaxTime: duration,
		}
	}
	stats := m.pathStats[path]
	stats.Count++
	stats.TotalTime += duration
	if duration < stats.MinTime {
		stats.MinTime = duration
	}
	if duration > stats.MaxTime {
		stats.MaxTime = duration
	}
	if statusCode >= 200 && statusCode < 400 {
		successCount := float64(stats.Count) - (float64(stats.Count) - float64(m.successRequests)/float64(m.totalRequests)*float64(stats.Count))
		stats.SuccessRate = successCount / float64(stats.Count) * 100
	}
}

// GetStats 获取统计信息
func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgDuration := time.Duration(0)
	if m.totalRequests > 0 {
		avgDuration = m.totalDuration / time.Duration(m.totalRequests)
	}

	successRate := float64(0)
	if m.totalRequests > 0 {
		successRate = float64(m.successRequests) / float64(m.totalRequests) * 100
	}

	return map[string]interface{}{
		"total_requests":   m.totalRequests,
		"success_requests": m.successRequests,
		"failed_requests":  m.failedRequests,
		"success_rate":     successRate,
		"avg_duration_ms":  avgDuration.Milliseconds(),
		"min_duration_ms":  m.minDuration.Milliseconds(),
		"max_duration_ms":  m.maxDuration.Milliseconds(),
		"status_codes":     m.statusCodes,
		"path_stats":       m.pathStats,
	}
}

// Reset 重置统计
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRequests = 0
	m.successRequests = 0
	m.failedRequests = 0
	m.totalDuration = 0
	m.minDuration = time.Hour
	m.maxDuration = 0
	m.statusCodes = make(map[int]int64)
	m.pathStats = make(map[string]*PathStats)
}
