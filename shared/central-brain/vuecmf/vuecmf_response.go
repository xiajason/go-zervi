package vuecmf

// VueCMFResponse VueCMF 统一响应格式
type VueCMFResponse struct {
	Code int         `json:"code"` // 0=成功, 非0=失败
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Success 成功响应
func Success(data interface{}) VueCMFResponse {
	return VueCMFResponse{
		Code: 0,
		Msg:  "success",
		Data: data,
	}
}

// Error 错误响应
func Error(code int, msg string) VueCMFResponse {
	return VueCMFResponse{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}

// 统一的错误码
const (
	CodeSuccess      = 0    // 成功
	CodeError        = 1    // 通用错误
	CodeUnauthorized = 401  // 未授权
	CodeForbidden    = 403  // 禁止访问
	CodeNotFound     = 404  // 未找到
	CodeValidation   = 422  // 验证失败
	CodeInternal     = 500  // 内部错误
)

