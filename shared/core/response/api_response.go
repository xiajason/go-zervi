package response

import (
	"encoding/json"
	"time"
)

// ApiResponse 标准API响应格式
type ApiResponse struct {
	Code      int         `json:"code"`                // 响应码：0表示成功
	Message   string      `json:"message"`             // 响应消息
	Data      interface{} `json:"data"`                // 业务数据
	Timestamp int64       `json:"timestamp"`           // 时间戳
}

// Success 成功响应
func Success(message string, data interface{}) *ApiResponse {
	return &ApiResponse{
		Code:      0,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
}

// Error 错误响应
func Error(code int, message string) *ApiResponse {
	return &ApiResponse{
		Code:      code,
		Message:   message,
		Data:      nil,
		Timestamp: time.Now().UnixMilli(),
	}
}

// ToJSON 转换为JSON字符串
func (r *ApiResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// 常用错误码
const (
	CodeSuccess           = 0    // 成功
	CodeInvalidParams     = 400  // 参数错误
	CodeUnauthorized      = 401  // 未授权
	CodeForbidden         = 403  // 禁止访问
	CodeNotFound          = 404  // 未找到
	CodeInternalError     = 500  // 内部错误
	CodeUserNotFound      = 1001 // 用户不存在
	CodeInvalidToken      = 1002 // 无效令牌
	CodePermissionDenied  = 1003 // 权限不足
	CodeResourceNotFound  = 1004 // 资源不存在
)

// 常用错误消息
const (
	MsgSuccess           = "操作成功"
	MsgInvalidParams     = "参数错误"
	MsgUnauthorized      = "未登录"
	MsgForbidden         = "禁止访问"
	MsgNotFound          = "资源不存在"
	MsgInternalError     = "内部错误"
	MsgUserNotFound      = "用户不存在"
	MsgInvalidToken      = "无效的token"
	MsgPermissionDenied  = "权限不足"
	MsgResourceNotFound  = "资源不存在"
)
