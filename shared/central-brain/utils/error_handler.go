package utils

import (
	"encoding/json"
	"net/http"
	"time"
)

// ErrorResponse 统一错误响应格式
type ErrorResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	TraceID   string      `json:"trace_id,omitempty"`
}

// SuccessResponse 统一成功响应格式
type SuccessResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	TraceID   string      `json:"trace_id,omitempty"`
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string, traceID string) *ErrorResponse {
	return &ErrorResponse{
		Code:      code,
		Message:   message,
		Data:      nil,
		Timestamp: time.Now().Unix(),
		TraceID:   traceID,
	}
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(message string, data interface{}, traceID string) *SuccessResponse {
	return &SuccessResponse{
		Code:      0,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		TraceID:   traceID,
	}
}

// WriteErrorResponse 写入错误响应
func WriteErrorResponse(w http.ResponseWriter, code int, message string, traceID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := NewErrorResponse(code, message, traceID)
	json.NewEncoder(w).Encode(response)
}

// WriteSuccessResponse 写入成功响应
func WriteSuccessResponse(w http.ResponseWriter, message string, data interface{}, traceID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := NewSuccessResponse(message, data, traceID)
	json.NewEncoder(w).Encode(response)
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []int // 可重试的HTTP状态码
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []int{
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
	}
}

// ShouldRetry 判断是否应该重试
func (rc *RetryConfig) ShouldRetry(statusCode int) bool {
	for _, code := range rc.RetryableErrors {
		if statusCode == code {
			return true
		}
	}
	return false
}

// CalculateDelay 计算重试延迟（指数退避）
func (rc *RetryConfig) CalculateDelay(attempt int) time.Duration {
	delay := time.Duration(float64(rc.InitialDelay) * pow(rc.BackoffFactor, float64(attempt)))
	if delay > rc.MaxDelay {
		return rc.MaxDelay
	}
	return delay
}

// pow 计算幂
func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}
