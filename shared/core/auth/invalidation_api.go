package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/szjason72/zervigo/shared/core/response"
)

// InvalidationAPI Token失效管理API
type InvalidationAPI struct {
	authSystem *UnifiedAuthSystem
}

// NewInvalidationAPI 创建Token失效管理API
func NewInvalidationAPI(authSystem *UnifiedAuthSystem) *InvalidationAPI {
	return &InvalidationAPI{
		authSystem: authSystem,
	}
}

// HandleInvalidateAllTokens 处理全局Token失效请求
// 需要管理员权限
func (api *InvalidationAPI) HandleInvalidateAllTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Method not allowed"))
		return
	}

	// 验证管理员权限（从Token中获取）
	token := extractTokenFromRequest(r)
	if token == "" {
		api.writeErrorResponse(w, response.Error(response.CodeUnauthorized, "Token required"))
		return
	}

	result, err := api.authSystem.ValidateJWT(token)
	if err != nil || !result.Success {
		api.writeErrorResponse(w, response.Error(response.CodeUnauthorized, "Invalid token"))
		return
	}

	// 检查是否为管理员
	if result.User.Role != "admin" && result.User.Role != "super_admin" {
		api.writeErrorResponse(w, response.Error(response.CodeForbidden, "Admin permission required"))
		return
	}

	// 使所有Token失效
	InvalidateAllTokens()

	// 记录操作日志
	api.authSystem.logAccess(
		result.User.ID,
		"invalidate_all_tokens",
		"auth",
		"success",
		getClientIP(r),
		getUserAgent(r),
	)

	api.writeSuccessResponse(w, response.Success("所有Token已失效", map[string]interface{}{
		"invalidation_time": GetTokenInvalidationTime(),
		"message":           "所有在此时间之前签发的Token都已失效",
	}))
}

// HandleGetInvalidationTime 获取Token失效时间
func (api *InvalidationAPI) HandleGetInvalidationTime(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Method not allowed"))
		return
	}

	api.writeSuccessResponse(w, response.Success("获取成功", map[string]interface{}{
		"invalidation_time": GetTokenInvalidationTime(),
	}))
}

// 辅助方法
func (api *InvalidationAPI) writeSuccessResponse(w http.ResponseWriter, resp *response.ApiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (api *InvalidationAPI) writeErrorResponse(w http.ResponseWriter, resp *response.ApiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// extractTokenFromRequest 从请求中提取token（复用unified_auth_api.go中的逻辑）
func extractTokenFromRequest(r *http.Request) string {
	// 从Authorization头获取
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// 兼容大小写并按空格分割
		fields := strings.Fields(authHeader)
		if len(fields) == 2 && strings.EqualFold(fields[0], "Bearer") {
			return fields[1]
		}
	}

	// 从查询参数获取
	token := r.URL.Query().Get("token")
	if token != "" {
		return token
	}

	// 从Cookie获取
	cookie, err := r.Cookie("access_token")
	if err == nil && cookie != nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

