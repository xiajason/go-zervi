package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TenantAPI 租户管理API处理器
type TenantAPI struct {
	service *TenantService
}

// NewTenantAPI 创建租户API处理器
func NewTenantAPI(service *TenantService) *TenantAPI {
	return &TenantAPI{service: service}
}

// RegisterRoutes 注册路由
func (api *TenantAPI) RegisterRoutes(r *gin.RouterGroup) {
	tenants := r.Group("/tenants")
	{
		// 创建租户
		tenants.POST("", api.CreateTenant)
		// 获取租户列表
		tenants.GET("", api.ListTenants)
		// 获取租户详情
		tenants.GET("/:id", api.GetTenant)
		// 更新租户
		tenants.PUT("/:id", api.UpdateTenant)
		// 删除租户
		tenants.DELETE("/:id", api.DeleteTenant)
		// 切换租户
		tenants.POST("/:id/switch", api.SwitchTenant)
		// 获取当前用户的租户列表
		tenants.GET("/my/list", api.GetMyTenants)
		// 获取租户的用户列表
		tenants.GET("/:id/users", api.GetTenantUsers)
		// 添加用户到租户
		tenants.POST("/:id/users", api.AddUserToTenant)
		// 从租户移除用户
		tenants.DELETE("/:id/users/:user_id", api.RemoveUserFromTenant)
	}
}

// CreateTenant 创建租户
func (api *TenantAPI) CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request payload: " + err.Error(),
		})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User ID not found",
		})
		return
	}

	userIDInt64 := int64(userID.(int))
	if userIDInt64 == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	// 创建租户
	tenant, err := api.service.CreateTenant(c.Request.Context(), req, userIDInt64)
	if err != nil {
		if err == ErrTenantCodeExists {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Tenant code already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create tenant: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tenant created successfully",
		"data":    tenant,
	})
}

// ListTenants 获取租户列表
func (api *TenantAPI) ListTenants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	// 检查管理员权限（可选，根据业务需求）
	role := c.GetString("role")
	if role != "admin" && role != "super_admin" {
		// 非管理员只能查看自己所属的租户
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "User ID not found",
			})
			return
		}
		userIDInt64 := int64(userID.(int))
		tenants, err := api.service.GetUserTenants(c.Request.Context(), userIDInt64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to get user tenants: " + err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"items":       tenants,
				"total":       len(tenants),
				"page":        1,
				"page_size":   len(tenants),
				"total_pages": 1,
			},
		})
		return
	}

	// 管理员可以查看所有租户
	result, err := api.service.ListTenants(c.Request.Context(), page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to list tenants: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetTenant 获取租户详情
func (api *TenantAPI) GetTenant(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid tenant ID",
		})
		return
	}

	// 检查用户是否有权限访问该租户
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User ID not found",
		})
		return
	}
	userIDInt64 := int64(userID.(int))

	// 检查管理员权限
	role := c.GetString("role")
	if role != "admin" && role != "super_admin" {
		// 非管理员需要验证是否属于该租户
		_, err := api.service.CheckUserTenantPermission(c.Request.Context(), userIDInt64, tenantID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "You don't have permission to access this tenant",
			})
			return
		}
	}

	tenant, err := api.service.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		if err == ErrTenantNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tenant not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get tenant: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tenant,
	})
}

// UpdateTenant 更新租户
func (api *TenantAPI) UpdateTenant(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid tenant ID",
		})
		return
	}

	var req UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request payload: " + err.Error(),
		})
		return
	}

	// 检查用户权限（必须是租户所有者或管理员）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User ID not found",
		})
		return
	}
	userIDInt64 := int64(userID.(int))

	// 检查管理员权限
	role := c.GetString("role")
	if role != "admin" && role != "super_admin" {
		// 非管理员需要验证是否是租户所有者
		userTenant, err := api.service.CheckUserTenantPermission(c.Request.Context(), userIDInt64, tenantID)
		if err != nil || userTenant.Role != TenantRoleOwner {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Only tenant owner can update tenant",
			})
			return
		}
	}

	tenant, err := api.service.UpdateTenant(c.Request.Context(), tenantID, req)
	if err != nil {
		if err == ErrTenantNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tenant not found",
			})
			return
		}
		if err == ErrInvalidTenantStatus {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid tenant status",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update tenant: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tenant updated successfully",
		"data":    tenant,
	})
}

// DeleteTenant 删除租户
func (api *TenantAPI) DeleteTenant(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid tenant ID",
		})
		return
	}

	// 检查用户权限（必须是租户所有者或超级管理员）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User ID not found",
		})
		return
	}
	userIDInt64 := int64(userID.(int))

	// 检查管理员权限
	role := c.GetString("role")
	if role != "admin" && role != "super_admin" {
		// 非管理员需要验证是否是租户所有者
		userTenant, err := api.service.CheckUserTenantPermission(c.Request.Context(), userIDInt64, tenantID)
		if err != nil || userTenant.Role != TenantRoleOwner {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Only tenant owner can delete tenant",
			})
			return
		}
	}

	if err := api.service.DeleteTenant(c.Request.Context(), tenantID); err != nil {
		if err == ErrTenantNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tenant not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete tenant: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tenant deleted successfully",
	})
}

// SwitchTenant 切换租户
func (api *TenantAPI) SwitchTenant(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid tenant ID",
		})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User ID not found",
		})
		return
	}
	userIDInt64 := int64(userID.(int))

	// 切换租户
	if err := api.service.SwitchTenant(c.Request.Context(), userIDInt64, tenantID); err != nil {
		if err == ErrTenantNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tenant not found",
			})
			return
		}
		if err == ErrUserNotInTenant {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "You are not a member of this tenant",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to switch tenant: " + err.Error(),
		})
		return
	}

	// 获取租户信息
	tenant, err := api.service.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get tenant info: " + err.Error(),
		})
		return
	}

	// 更新用户的当前租户ID（需要更新auth服务中的用户信息）
	// 这里返回租户信息，前端可以更新JWT Token或本地存储
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tenant switched successfully",
		"data":    tenant,
	})
}

// GetMyTenants 获取当前用户的租户列表
func (api *TenantAPI) GetMyTenants(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User ID not found",
		})
		return
	}
	userIDInt64 := int64(userID.(int))

	tenants, err := api.service.GetUserTenants(c.Request.Context(), userIDInt64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get user tenants: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tenants,
	})
}

// GetTenantUsers 获取租户的用户列表
func (api *TenantAPI) GetTenantUsers(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid tenant ID",
		})
		return
	}

	// 检查用户权限
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User ID not found",
		})
		return
	}
	userIDInt64 := int64(userID.(int))

	// 检查管理员权限或租户成员权限
	role := c.GetString("role")
	if role != "admin" && role != "super_admin" {
		_, err := api.service.CheckUserTenantPermission(c.Request.Context(), userIDInt64, tenantID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "You don't have permission to access this tenant",
			})
			return
		}
	}

	users, err := api.service.GetTenantUsers(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get tenant users: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
	})
}

// AddUserToTenant 添加用户到租户
func (api *TenantAPI) AddUserToTenant(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid tenant ID",
		})
		return
	}

	var req struct {
		UserID int64  `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request payload: " + err.Error(),
		})
		return
	}

	// 检查用户权限（必须是租户所有者或管理员）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User ID not found",
		})
		return
	}
	currentUserID := int64(userID.(int))

	// 检查管理员权限
	role := c.GetString("role")
	if role != "admin" && role != "super_admin" {
		// 非管理员需要验证是否是租户所有者或管理员
		userTenant, err := api.service.CheckUserTenantPermission(c.Request.Context(), currentUserID, tenantID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "You don't have permission to manage this tenant",
			})
			return
		}
		if userTenant.Role != TenantRoleOwner && userTenant.Role != TenantRoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Only tenant owner or admin can add users",
			})
			return
		}
	}

	if err := api.service.AddUserToTenant(c.Request.Context(), tenantID, req.UserID, req.Role); err != nil {
		if err == ErrInvalidTenantRole {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid tenant role",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to add user to tenant: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User added to tenant successfully",
	})
}

// RemoveUserFromTenant 从租户移除用户
func (api *TenantAPI) RemoveUserFromTenant(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid tenant ID",
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	// 检查用户权限（必须是租户所有者或管理员）
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User ID not found",
		})
		return
	}
	currentUserIDInt64 := int64(currentUserID.(int))

	// 检查管理员权限
	role := c.GetString("role")
	if role != "admin" && role != "super_admin" {
		// 非管理员需要验证是否是租户所有者或管理员
		userTenant, err := api.service.CheckUserTenantPermission(c.Request.Context(), currentUserIDInt64, tenantID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "You don't have permission to manage this tenant",
			})
			return
		}
		if userTenant.Role != TenantRoleOwner && userTenant.Role != TenantRoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Only tenant owner or admin can remove users",
			})
			return
		}
	}

	if err := api.service.RemoveUserFromTenant(c.Request.Context(), tenantID, userID); err != nil {
		if err == ErrUserNotInTenant {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "User is not in this tenant",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to remove user from tenant: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User removed from tenant successfully",
	})
}

