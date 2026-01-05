package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/szjason72/zervigo/shared/core/context"
	"github.com/szjason72/zervigo/shared/core/auth"
)

const (
	// TenantIDHeader 租户ID请求头名称
	TenantIDHeader = "X-Tenant-ID"
)

// TenantMiddleware 租户ID中间件
// 从JWT Token或请求头获取租户ID，并设置到context中
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tenantID int64

		// 1. 优先从JWT Token获取租户ID
		// 从gin context获取用户信息（由AuthMiddleware设置）
		if userValue, exists := c.Get("user"); exists {
			if user, ok := userValue.(*auth.User); ok {
				// 如果User结构体有TenantID字段，使用它
				if user.TenantID > 0 {
					tenantID = user.TenantID
				} else if user.LastTenantID != nil && *user.LastTenantID > 0 {
					// 如果没有当前租户ID，使用最后使用的租户ID
					tenantID = *user.LastTenantID
				}
			}
		}

		// 2. 如果JWT中没有，尝试从请求头获取（用于租户切换）
		if tenantID == 0 {
			tenantIDHeader := c.GetHeader(TenantIDHeader)
			if tenantIDHeader != "" {
				// 解析租户ID
				// 这里需要验证用户是否有该租户的权限
				// 暂时先解析，后续添加权限验证
				// tenantID = parseTenantID(tenantIDHeader)
			}
		}

		// 3. 如果获取到租户ID，设置到context
		if tenantID > 0 {
			ctx := context.SetTenantID(c.Request.Context(), tenantID)
			c.Request = c.Request.WithContext(ctx)
			c.Set("tenant_id", tenantID)
		}

		c.Next()
	}
}

// RequireTenant 需要租户的中间件
// 确保请求中必须包含有效的租户ID
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, err := context.GetTenantID(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "租户ID未找到，请先选择租户",
			})
			c.Abort()
			return
		}

		// 验证租户ID有效性（可以添加数据库查询验证）
		// 暂时只检查是否大于0
		if tenantID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "无效的租户ID",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

