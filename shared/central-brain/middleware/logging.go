package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestLogger 请求日志中间件
type RequestLogger struct {
	enabled bool
}

// NewRequestLogger 创建请求日志中间件
func NewRequestLogger(enabled bool) *RequestLogger {
	return &RequestLogger{enabled: enabled}
}

// Middleware 日志中间件处理函数
func (rl *RequestLogger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.enabled {
			c.Next()
			return
		}

		// 生成请求追踪ID
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)

		// 记录开始时间
		startTime := time.Now()

		// 读取请求体（用于日志记录）
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建响应写入器（用于记录响应）
		responseWriter := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(startTime)

		// 记录日志
		rl.logRequest(c, traceID, requestBody, responseWriter, duration)
	}
}

// logRequest 记录请求日志
func (rl *RequestLogger) logRequest(c *gin.Context, traceID string, requestBody []byte, rw *responseWriter, duration time.Duration) {
	logEntry := map[string]interface{}{
		"timestamp":     time.Now().Format(time.RFC3339),
		"trace_id":      traceID,
		"method":        c.Request.Method,
		"path":          c.Request.URL.Path,
		"query":         c.Request.URL.RawQuery,
		"remote_addr":   c.ClientIP(),
		"user_agent":    c.Request.UserAgent(),
		"status_code":   rw.Status(),
		"duration_ms":   duration.Milliseconds(),
		"request_size":  len(requestBody),
		"response_size": rw.body.Len(),
	}

	// 添加错误信息（如果有）
	if len(c.Errors) > 0 {
		errors := make([]string, len(c.Errors))
		for i, err := range c.Errors {
			errors[i] = err.Error()
		}
		logEntry["errors"] = errors
	}

	// 输出结构化日志（JSON格式）
	logJSON, _ := json.Marshal(logEntry)
	if rw.Status() >= 400 {
		// 错误日志
		println("ERROR:", string(logJSON))
	} else {
		// 正常日志
		println("INFO:", string(logJSON))
	}
}

// responseWriter 响应写入器包装器
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}
