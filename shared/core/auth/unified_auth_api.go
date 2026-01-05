package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/szjason72/zervigo/shared/core/response"
)

// UnifiedAuthAPI 统一认证API服务器
type UnifiedAuthAPI struct {
	authSystem         *UnifiedAuthSystem
	serviceAuthService *ServiceAuthService
	port               int
}

// NewUnifiedAuthAPI 创建统一认证API服务器
func NewUnifiedAuthAPI(authSystem *UnifiedAuthSystem, port int) *UnifiedAuthAPI {
	// 创建服务认证服务（使用zervigo-2025密钥）
	serviceJWTSecret := "zervigo-mvp-secret-key-2025"
	serviceAuthService := NewServiceAuthService(authSystem.db, serviceJWTSecret)

	return &UnifiedAuthAPI{
		authSystem:         authSystem,
		serviceAuthService: serviceAuthService,
		port:               port,
	}
}

// Start 启动API服务器
func (api *UnifiedAuthAPI) Start() error {
	// 用户认证路由（使用jobfirst-2024密钥）
	http.HandleFunc("/api/v1/auth/login", api.handleLogin)
	http.HandleFunc("/api/v1/auth/validate", api.handleValidateJWT)
	http.HandleFunc("/api/v1/auth/permission", api.handleCheckPermission)
	http.HandleFunc("/api/v1/auth/user", api.handleGetUser)
	http.HandleFunc("/api/v1/auth/access", api.handleValidateAccess)
	http.HandleFunc("/api/v1/auth/log", api.handleLogAccess)
	http.HandleFunc("/api/v1/auth/roles", api.handleGetRoles)
	http.HandleFunc("/api/v1/auth/permissions", api.handleGetPermissions)

	// 服务认证路由（使用zervigo-2025密钥）
	http.HandleFunc("/api/v1/auth/service/login", api.handleServiceLogin)
	http.HandleFunc("/api/v1/auth/service/validate", api.handleServiceValidate)
	http.HandleFunc("/api/v1/auth/service/permission", api.handleServicePermission)

	http.HandleFunc("/health", api.handleHealth)

	// 启动服务器
	addr := fmt.Sprintf(":%d", api.port)
	fmt.Printf("统一认证服务启动在端口 %d\n", api.port)
	return http.ListenAndServe(addr, nil)
}

// handleLogin 处理登录请求
func (api *UnifiedAuthAPI) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Method not allowed"))
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Invalid JSON"))
		return
	}

	if req.Username == "" || req.Password == "" {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Username and password are required"))
		return
	}

	result, err := api.authSystem.Authenticate(req.Username, req.Password)
	if err != nil {
		api.writeErrorResponse(w, response.Error(response.CodeInternalError, err.Error()))
		return
	}

	// 调试信息
	fmt.Printf("DEBUG: Authentication result - Success: %v, User: %v, Error: %s\n",
		result.Success, result.User != nil, result.Error)

	// 记录访问日志
	api.authSystem.logAccess(0, "login", "auth",
		map[bool]string{true: "success", false: "failed"}[result.Success],
		getClientIP(r), getUserAgent(r))

	// 构建标准响应格式
	if result.Success && result.User != nil {
		loginData := map[string]interface{}{
			"userId":       result.User.ID,
			"userName":     result.User.Username,
			"userPhone":    result.User.Phone,
			"userAvatar":   nil, // 暂时设为nil，后续可以从资源服务获取
			"userStatus":   result.User.Status,
			"loginStatus":  api.calculateLoginStatus(api.getUserStatusInt(result.User.Status)),
			"accessToken":  result.Token,
			"refreshToken": "", // 暂时为空，后续可以生成
		}
		api.writeSuccessResponse(w, response.Success("登录成功", loginData))
	} else {
		// 使用result中的错误信息
		errorCode := response.CodeUserNotFound
		if result.ErrorCode == "INVALID_PASSWORD" {
			errorCode = response.CodeUnauthorized
		} else if result.ErrorCode == "USER_DISABLED" {
			errorCode = response.CodeForbidden
		}
		api.writeErrorResponse(w, response.Error(errorCode, result.Error))
	}
}

// handleValidateJWT 处理JWT验证请求
func (api *UnifiedAuthAPI) handleValidateJWT(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Method not allowed"))
		return
	}

	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Invalid JSON"))
		return
	}

	if req.Token == "" {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Token is required"))
		return
	}

	result, err := api.authSystem.ValidateJWT(req.Token)
	if err != nil {
		api.writeErrorResponse(w, response.Error(response.CodeInternalError, err.Error()))
		return
	}

	if result.Success {
		api.writeSuccessResponse(w, response.Success("Token验证成功", result))
	} else {
		errorCode := response.CodeUnauthorized
		if result.ErrorCode == "TOKEN_EXPIRED" {
			errorCode = response.CodeUnauthorized
		}
		api.writeErrorResponse(w, response.Error(errorCode, result.Error))
	}
}

// handleCheckPermission 处理权限检查请求
func (api *UnifiedAuthAPI) handleCheckPermission(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	permission := r.URL.Query().Get("permission")

	if userIDStr == "" || permission == "" {
		http.Error(w, "user_id and permission are required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	hasPermission, err := api.authSystem.CheckPermission(userID, permission)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user_id":        userID,
		"permission":     permission,
		"has_permission": hasPermission,
		"timestamp":      time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetUser 处理获取用户信息请求
func (api *UnifiedAuthAPI) handleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	user, err := api.authSystem.getUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// handleValidateAccess 处理访问验证请求
func (api *UnifiedAuthAPI) handleValidateAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID   int    `json:"user_id"`
		Resource string `json:"resource"`
		Action   string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.UserID == 0 || req.Resource == "" || req.Action == "" {
		http.Error(w, "user_id, resource, and action are required", http.StatusBadRequest)
		return
	}

	// 构建权限字符串
	permission := fmt.Sprintf("%s:%s", req.Action, req.Resource)
	hasPermission, err := api.authSystem.CheckPermission(req.UserID, permission)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user_id":        req.UserID,
		"resource":       req.Resource,
		"action":         req.Action,
		"permission":     permission,
		"has_permission": hasPermission,
		"timestamp":      time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleLogAccess 处理访问日志记录请求
func (api *UnifiedAuthAPI) handleLogAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID    int    `json:"user_id"`
		Action    string `json:"action"`
		Resource  string `json:"resource"`
		Result    string `json:"result"`
		IPAddress string `json:"ip_address"`
		UserAgent string `json:"user_agent"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.UserID == 0 || req.Action == "" || req.Resource == "" || req.Result == "" {
		http.Error(w, "user_id, action, resource, and result are required", http.StatusBadRequest)
		return
	}

	api.authSystem.logAccess(req.UserID, req.Action, req.Resource, req.Result, req.IPAddress, req.UserAgent)

	response := map[string]interface{}{
		"success":   true,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetRoles 处理获取角色列表请求
func (api *UnifiedAuthAPI) handleGetRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.authSystem.roleConfig.Roles)
}

// handleGetPermissions 处理获取权限列表请求
func (api *UnifiedAuthAPI) handleGetPermissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role := r.URL.Query().Get("role")
	if role == "" {
		http.Error(w, "role is required", http.StatusBadRequest)
		return
	}

	permissions, err := api.authSystem.getUserPermissions(role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"role":        role,
		"permissions": permissions,
		"timestamp":   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth 处理健康检查请求
func (api *UnifiedAuthAPI) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Method not allowed"))
		return
	}

	healthData := map[string]interface{}{
		"status":  "healthy",
		"service": "unified-auth-service",
		"version": "2.0.0",
		"features": []string{
			"unified_role_system",
			"complete_jwt_validation",
			"permission_management",
			"access_logging",
			"database_optimization",
		},
	}

	api.writeSuccessResponse(w, response.Success("服务健康", healthData))
}

// 辅助方法
func (api *UnifiedAuthAPI) writeSuccessResponse(w http.ResponseWriter, resp *response.ApiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (api *UnifiedAuthAPI) writeErrorResponse(w http.ResponseWriter, resp *response.ApiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 使用200状态码，错误信息在响应体中
	json.NewEncoder(w).Encode(resp)
}

// calculateLoginStatus 计算登录状态
func (api *UnifiedAuthAPI) calculateLoginStatus(userStatus int) int {
	switch userStatus {
	case 0:
		return 1 // 需要手机号
	case 1:
		return 0 // 正常
	case 2:
		return 2 // 需要实名认证
	default:
		return 0
	}
}

// getUserStatusInt 将用户状态字符串转换为整数
func (api *UnifiedAuthAPI) getUserStatusInt(status string) int {
	switch status {
	case "active":
		return 1
	case "inactive":
		return 0
	case "suspended":
		return 2
	default:
		return 0
	}
}

// 辅助函数
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

func getUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// ==================== 服务认证API ====================

// handleServiceLogin 处理服务登录请求
func (api *UnifiedAuthAPI) handleServiceLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Method not allowed"))
		return
	}

	var req struct {
		ServiceID     string `json:"service_id"`
		ServiceSecret string `json:"service_secret"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Invalid JSON"))
		return
	}

	if req.ServiceID == "" || req.ServiceSecret == "" {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Service ID and secret are required"))
		return
	}

	result, err := api.serviceAuthService.AuthenticateService(req.ServiceID, req.ServiceSecret)
	if err != nil {
		api.writeErrorResponse(w, response.Error(response.CodeInternalError, err.Error()))
		return
	}

	if result.Success {
		serviceData := map[string]interface{}{
			"service_id":    result.Service.ServiceID,
			"service_name":  result.Service.ServiceName,
			"service_type":  result.Service.ServiceType,
			"service_token": result.Token,
			"expires_in":    result.ExpiresIn,
		}
		api.writeSuccessResponse(w, response.Success("服务认证成功", serviceData))
	} else {
		errorCode := response.CodeUnauthorized
		if result.ErrorCode == "SERVICE_NOT_FOUND" {
			errorCode = response.CodeNotFound
		}
		api.writeErrorResponse(w, response.Error(errorCode, result.Error))
	}
}

// handleServiceValidate 处理服务token验证请求
func (api *UnifiedAuthAPI) handleServiceValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Method not allowed"))
		return
	}

	var req struct {
		ServiceToken string `json:"service_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Invalid JSON"))
		return
	}

	if req.ServiceToken == "" {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Service token is required"))
		return
	}

	result, err := api.serviceAuthService.ValidateServiceToken(req.ServiceToken)
	if err != nil {
		api.writeErrorResponse(w, response.Error(response.CodeInternalError, err.Error()))
		return
	}

	if result.Success {
		serviceData := map[string]interface{}{
			"valid":        true,
			"service_id":   result.Service.ServiceID,
			"service_name": result.Service.ServiceName,
			"service_type": result.Service.ServiceType,
			"allowed_apis": result.Service.AllowedAPIs,
		}
		api.writeSuccessResponse(w, response.Success("服务token验证成功", serviceData))
	} else {
		errorCode := response.CodeUnauthorized
		if result.ErrorCode == "SERVICE_NOT_FOUND" {
			errorCode = response.CodeNotFound
		}
		api.writeErrorResponse(w, response.Error(errorCode, result.Error))
	}
}

// handleServicePermission 处理服务权限检查请求
func (api *UnifiedAuthAPI) handleServicePermission(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Method not allowed"))
		return
	}

	var req struct {
		ServiceToken string `json:"service_token"`
		APIPath      string `json:"api_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Invalid JSON"))
		return
	}

	if req.ServiceToken == "" || req.APIPath == "" {
		api.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Service token and API path are required"))
		return
	}

	// 验证服务token
	result, err := api.serviceAuthService.ValidateServiceToken(req.ServiceToken)
	if err != nil {
		api.writeErrorResponse(w, response.Error(response.CodeInternalError, err.Error()))
		return
	}

	if !result.Success {
		api.writeErrorResponse(w, response.Error(response.CodeUnauthorized, result.Error))
		return
	}

	// 检查服务权限
	hasPermission, err := api.serviceAuthService.CheckServicePermission(result.Service.ServiceID, req.APIPath)
	if err != nil {
		api.writeErrorResponse(w, response.Error(response.CodeInternalError, err.Error()))
		return
	}

	permissionData := map[string]interface{}{
		"service_id":     result.Service.ServiceID,
		"api_path":       req.APIPath,
		"has_permission": hasPermission,
	}

	if hasPermission {
		api.writeSuccessResponse(w, response.Success("服务有权限访问该API", permissionData))
	} else {
		api.writeErrorResponse(w, response.Error(response.CodeForbidden, "服务无权限访问该API"))
	}
}
