package auth

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/szjason72/zervigo/shared/core/response"
)

// ServiceAuthMiddleware 服务认证中间件
// 验证服务间通信的token（使用zervigo-2025密钥）
type ServiceAuthMiddleware struct {
	serviceAuthService *ServiceAuthService
}

// NewServiceAuthMiddleware 创建服务认证中间件
func NewServiceAuthMiddleware(db *sql.DB) *ServiceAuthMiddleware {
	serviceJWTSecret := "zervigo-mvp-secret-key-2025"
	serviceAuthService := NewServiceAuthService(db, serviceJWTSecret)

	return &ServiceAuthMiddleware{
		serviceAuthService: serviceAuthService,
	}
}

// RequireServiceAuth 需要服务认证的中间件
func (sam *ServiceAuthMiddleware) RequireServiceAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取服务token
		serviceToken := sam.extractServiceToken(c)

		if serviceToken == "" {
			sam.writeErrorResponse(c, response.Error(response.CodeUnauthorized, "缺少服务token"))
			c.Abort()
			return
		}

		// 验证服务token
		result, err := sam.serviceAuthService.ValidateServiceToken(serviceToken)
		if err != nil {
			sam.writeErrorResponse(c, response.Error(response.CodeInternalError, err.Error()))
			c.Abort()
			return
		}

		if !result.Success {
			sam.writeErrorResponse(c, response.Error(response.CodeUnauthorized, result.Error))
			c.Abort()
			return
		}

		// 设置服务信息到上下文
		c.Set("service_id", result.Service.ServiceID)
		c.Set("service_name", result.Service.ServiceName)
		c.Set("service_type", result.Service.ServiceType)
		c.Set("allowed_apis", result.Service.AllowedAPIs)

		c.Next()
	}
}

// CheckServicePermission 检查服务是否有权限访问特定API
func (sam *ServiceAuthMiddleware) CheckServicePermission(apiPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先验证服务token
		sam.RequireServiceAuth()(c)
		if c.IsAborted() {
			return
		}

		serviceID := c.GetString("service_id")

		// 检查服务权限
		hasPermission, err := sam.serviceAuthService.CheckServicePermission(serviceID, apiPath)
		if err != nil {
			sam.writeErrorResponse(c, response.Error(response.CodeInternalError, err.Error()))
			c.Abort()
			return
		}

		if !hasPermission {
			sam.writeErrorResponse(c, response.Error(response.CodeForbidden, "服务无权限访问该API"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractServiceToken 提取服务token
func (sam *ServiceAuthMiddleware) extractServiceToken(c *gin.Context) string {
	// 从X-Service-Token头提取
	serviceToken := c.GetHeader("X-Service-Token")
	if serviceToken != "" {
		return serviceToken
	}

	// 从Authorization头提取（兼容）
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Service " {
		return authHeader[7:]
	}

	return ""
}

// writeErrorResponse 写入错误响应
func (sam *ServiceAuthMiddleware) writeErrorResponse(c *gin.Context, errResp *response.ApiResponse) {
	c.JSON(http.StatusOK, errResp)
}
