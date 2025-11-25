package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	jobfirst "github.com/szjason72/zervigo/shared/core"
	"github.com/szjason72/zervigo/shared/core/auth"
	"github.com/szjason72/zervigo/shared/core/response"
)

func main() {
	// 设置进程名称
	if len(os.Args) > 0 {
		os.Args[0] = "permission-service"
	}

	// 初始化JobFirst核心包
	core, err := jobfirst.NewCore("")
	if err != nil {
		log.Fatalf("初始化JobFirst核心包失败: %v", err)
	}
	defer core.Close()

	// 初始化Go-Zervi认证系统
	jwtSecret := "zervigo-mvp-secret-key-2025"
	sqlDB, err := core.Database.GetPostgreSQL().GetSQLDB()
	if err != nil {
		log.Fatalf("获取PostgreSQL连接失败: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由器
	r := gin.Default()

	// 设置标准路由
	setupStandardRoutes(r, core)

	// 设置业务路由
	setupBusinessRoutes(r, core, sqlDB, jwtSecret)

	// 注册到Consul
	registerToConsul("permission-service", "127.0.0.1", 8086)

	// 启动服务
	log.Println("Starting Permission Service with Go-Zervi Framework on 0.0.0.0:8086")
	if err := r.Run(":8086"); err != nil {
		log.Fatalf("Permission Service启动失败: %v", err)
	}
}

// setupStandardRoutes 设置标准路由
func setupStandardRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		health := core.Health()
		c.JSON(http.StatusOK, gin.H{
			"service":     "permission-service",
			"status":      "healthy",
			"timestamp":   time.Now().Format(time.RFC3339),
			"version":     "3.1.0",
			"core_health": health,
		})
	})

	// 版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "permission-service",
			"version": "3.1.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 服务信息
	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":    "permission-service",
			"version":    "3.1.0",
			"port":       8086,
			"status":     "running",
			"started_at": time.Now().Format(time.RFC3339),
		})
	})
}

// setupBusinessRoutes 设置业务路由
func setupBusinessRoutes(r *gin.Engine, core *jobfirst.Core, sqlDB *sql.DB, jwtSecret string) {
	// 公开API路由（不需要认证）
	public := r.Group("/api/v1")
	{
		// 获取所有角色列表（公开）
		public.GET("/roles", func(c *gin.Context) {
			roles := getAllRoles(sqlDB)
			standardSuccessResponse(c, roles, "角色列表获取成功")
		})

		// 获取所有权限列表（公开）
		public.GET("/permissions", func(c *gin.Context) {
			permissions := getAllPermissions(sqlDB)
			standardSuccessResponse(c, permissions, "权限列表获取成功")
		})
	}

	// 需要认证的API路由
	zerviAuthAdapter := auth.NewZerviAuthAdapter(sqlDB, jwtSecret)
	authMiddleware := zerviAuthAdapter.RequireAuth()
	api := r.Group("/api/v1")
	api.Use(authMiddleware)
	{
		// 角色管理
		roles := api.Group("/roles")
		{
			// 创建角色
			roles.POST("/", func(c *gin.Context) {
				var req CreateRoleRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
					return
				}

				roleID := createRole(sqlDB, req)
				if roleID == "" {
					standardErrorResponse(c, http.StatusInternalServerError, "创建角色失败", "")
					return
				}

				result := gin.H{
					"roleId":  roleID,
					"message": "角色创建成功",
				}
				standardSuccessResponse(c, result, "角色创建成功")
			})

			// 更新角色
			roles.PUT("/:roleId", func(c *gin.Context) {
				roleID := c.Param("roleId")

				var req UpdateRoleRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
					return
				}

				if !updateRole(sqlDB, roleID, req) {
					standardErrorResponse(c, http.StatusNotFound, "角色不存在", "")
					return
				}

				standardSuccessResponse(c, "角色已更新", "角色更新成功")
			})

			// 删除角色
			roles.DELETE("/:roleId", func(c *gin.Context) {
				roleID := c.Param("roleId")

				if !deleteRole(sqlDB, roleID) {
					standardErrorResponse(c, http.StatusNotFound, "角色不存在", "")
					return
				}

				standardSuccessResponse(c, "角色已删除", "角色删除成功")
			})

			// 获取角色详情
			roles.GET("/:roleId", func(c *gin.Context) {
				roleID := c.Param("roleId")

				role := getRoleDetail(sqlDB, roleID)
				if role == nil {
					standardErrorResponse(c, http.StatusNotFound, "角色不存在", "")
					return
				}

				standardSuccessResponse(c, role, "角色详情获取成功")
			})
		}

		// 权限管理
		permissions := api.Group("/permissions")
		{
			// 创建权限
			permissions.POST("/", func(c *gin.Context) {
				var req CreatePermissionRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
					return
				}

				permissionID := createPermission(sqlDB, req)
				if permissionID == "" {
					standardErrorResponse(c, http.StatusInternalServerError, "创建权限失败", "")
					return
				}

				result := gin.H{
					"permissionId": permissionID,
					"message":      "权限创建成功",
				}
				standardSuccessResponse(c, result, "权限创建成功")
			})

			// 更新权限
			permissions.PUT("/:permissionId", func(c *gin.Context) {
				permissionID := c.Param("permissionId")

				var req UpdatePermissionRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
					return
				}

				if !updatePermission(sqlDB, permissionID, req) {
					standardErrorResponse(c, http.StatusNotFound, "权限不存在", "")
					return
				}

				standardSuccessResponse(c, "权限已更新", "权限更新成功")
			})

			// 删除权限
			permissions.DELETE("/:permissionId", func(c *gin.Context) {
				permissionID := c.Param("permissionId")

				if !deletePermission(sqlDB, permissionID) {
					standardErrorResponse(c, http.StatusNotFound, "权限不存在", "")
					return
				}

				standardSuccessResponse(c, "权限已删除", "权限删除成功")
			})

			// 获取权限详情
			permissions.GET("/:permissionId", func(c *gin.Context) {
				permissionID := c.Param("permissionId")

				permission := getPermissionDetail(sqlDB, permissionID)
				if permission == nil {
					standardErrorResponse(c, http.StatusNotFound, "权限不存在", "")
					return
				}

				standardSuccessResponse(c, permission, "权限详情获取成功")
			})
		}

		// 用户角色管理
		userRoles := api.Group("/users")
		{
			// 获取用户角色
			userRoles.GET("/:userId/roles", func(c *gin.Context) {
				userIDStr := c.Param("userId")
				userID, err := strconv.ParseUint(userIDStr, 10, 64)
				if err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "用户ID格式错误", err.Error())
					return
				}

				// 检查权限：只能查看自己的角色
				currentUserID := c.GetUint("user_id")
				if uint(userID) != currentUserID {
					standardErrorResponse(c, http.StatusForbidden, "无权限查看其他用户的角色", "")
					return
				}

				roles := getUserRoles(sqlDB, uint(userID))
				standardSuccessResponse(c, roles, "用户角色获取成功")
			})

			// 分配角色给用户
			userRoles.POST("/:userId/roles", func(c *gin.Context) {
				userIDStr := c.Param("userId")
				userID, err := strconv.ParseUint(userIDStr, 10, 64)
				if err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "用户ID格式错误", err.Error())
					return
				}

				var req AssignRoleRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
					return
				}

				if !assignRoleToUser(sqlDB, uint(userID), req.RoleID) {
					standardErrorResponse(c, http.StatusInternalServerError, "分配角色失败", "")
					return
				}

				standardSuccessResponse(c, "角色分配成功", "角色分配成功")
			})

			// 移除用户角色
			userRoles.DELETE("/:userId/roles/:roleId", func(c *gin.Context) {
				userIDStr := c.Param("userId")
				userID, err := strconv.ParseUint(userIDStr, 10, 64)
				if err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "用户ID格式错误", err.Error())
					return
				}

				roleID := c.Param("roleId")

				if !removeRoleFromUser(sqlDB, uint(userID), roleID) {
					standardErrorResponse(c, http.StatusNotFound, "用户角色不存在", "")
					return
				}

				standardSuccessResponse(c, "角色移除成功", "角色移除成功")
			})
		}

		// 角色权限管理
		rolePermissions := api.Group("/roles")
		{
			// 获取角色权限
			rolePermissions.GET("/:roleId/permissions", func(c *gin.Context) {
				roleID := c.Param("roleId")

				permissions := getRolePermissions(sqlDB, roleID)
				standardSuccessResponse(c, permissions, "角色权限获取成功")
			})

			// 分配权限给角色
			rolePermissions.POST("/:roleId/permissions", func(c *gin.Context) {
				roleID := c.Param("roleId")

				var req AssignPermissionRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
					return
				}

				if !assignPermissionToRole(sqlDB, roleID, req.PermissionID) {
					standardErrorResponse(c, http.StatusInternalServerError, "分配权限失败", "")
					return
				}

				standardSuccessResponse(c, "权限分配成功", "权限分配成功")
			})

			// 移除角色权限
			rolePermissions.DELETE("/:roleId/permissions/:permissionId", func(c *gin.Context) {
				roleID := c.Param("roleId")
				permissionID := c.Param("permissionId")

				if !removePermissionFromRole(sqlDB, roleID, permissionID) {
					standardErrorResponse(c, http.StatusNotFound, "角色权限不存在", "")
					return
				}

				standardSuccessResponse(c, "权限移除成功", "权限移除成功")
			})
		}
	}
}

// 数据模型定义
type CreateRoleRequest struct {
	RoleName        string `json:"roleName" binding:"required"`
	RoleDescription string `json:"roleDescription"`
	Level           int    `json:"level"`
}

type UpdateRoleRequest struct {
	RoleName        string `json:"roleName"`
	RoleDescription string `json:"roleDescription"`
	Level           int    `json:"level"`
}

type CreatePermissionRequest struct {
	PermissionName        string `json:"permissionName" binding:"required"`
	PermissionCode        string `json:"permissionCode" binding:"required"`
	ResourceType          string `json:"resourceType"`
	Action                string `json:"action"`
	PermissionDescription string `json:"permissionDescription"`
	ServiceName           string `json:"serviceName"`
}

type UpdatePermissionRequest struct {
	PermissionName        string `json:"permissionName"`
	PermissionCode        string `json:"permissionCode"`
	ResourceType          string `json:"resourceType"`
	Action                string `json:"action"`
	PermissionDescription string `json:"permissionDescription"`
	ServiceName           string `json:"serviceName"`
}

type AssignRoleRequest struct {
	RoleID string `json:"roleId" binding:"required"`
}

type AssignPermissionRequest struct {
	PermissionID string `json:"permissionId" binding:"required"`
}

// 业务逻辑函数
func getAllRoles(sqlDB *sql.DB) []gin.H {
	query := `
		SELECT id, role_name, role_description, level, created_at, updated_at
		FROM zervigo_auth_roles
		ORDER BY level DESC, created_at ASC
	`
	rows, err := sqlDB.Query(query)
	if err != nil {
		log.Printf("查询角色列表失败: %v", err)
		return []gin.H{}
	}
	defer rows.Close()

	var roles []gin.H
	for rows.Next() {
		var id int
		var roleName, roleDescription string
		var level int
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &roleName, &roleDescription, &level, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("扫描角色列表失败: %v", err)
			continue
		}

		role := gin.H{
			"id":              id,
			"roleName":        roleName,
			"roleDescription": roleDescription,
			"level":           level,
			"createdAt":       createdAt.UnixMilli(),
			"updatedAt":       updatedAt.UnixMilli(),
		}
		roles = append(roles, role)
	}

	return roles
}

func getAllPermissions(sqlDB *sql.DB) []gin.H {
	query := `
		SELECT id, permission_name, permission_code, resource_type, action, 
		       permission_description, service_name, created_at, updated_at
		FROM zervigo_auth_permissions
		ORDER BY service_name, resource_type, action
	`
	rows, err := sqlDB.Query(query)
	if err != nil {
		log.Printf("查询权限列表失败: %v", err)
		return []gin.H{}
	}
	defer rows.Close()

	var permissions []gin.H
	for rows.Next() {
		var id int
		var permissionName, permissionCode, resourceType, action, permissionDescription, serviceName string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &permissionName, &permissionCode, &resourceType, &action,
			&permissionDescription, &serviceName, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("扫描权限列表失败: %v", err)
			continue
		}

		permission := gin.H{
			"id":                    id,
			"permissionName":        permissionName,
			"permissionCode":        permissionCode,
			"resourceType":          resourceType,
			"action":                action,
			"permissionDescription": permissionDescription,
			"serviceName":           serviceName,
			"createdAt":             createdAt.UnixMilli(),
			"updatedAt":             updatedAt.UnixMilli(),
		}
		permissions = append(permissions, permission)
	}

	return permissions
}

func createRole(sqlDB *sql.DB, req CreateRoleRequest) string {
	query := `
		INSERT INTO zervigo_auth_roles (role_name, role_description, level, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`

	var roleID int
	err := sqlDB.QueryRow(query, req.RoleName, req.RoleDescription, req.Level).Scan(&roleID)
	if err != nil {
		log.Printf("创建角色失败: %v", err)
		return ""
	}

	return strconv.Itoa(roleID)
}

func updateRole(sqlDB *sql.DB, roleID string, req UpdateRoleRequest) bool {
	query := `
		UPDATE zervigo_auth_roles 
		SET role_name = $1, role_description = $2, level = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`

	result, err := sqlDB.Exec(query, req.RoleName, req.RoleDescription, req.Level, roleID)
	if err != nil {
		log.Printf("更新角色失败: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}

func deleteRole(sqlDB *sql.DB, roleID string) bool {
	query := `DELETE FROM zervigo_auth_roles WHERE id = $1`

	result, err := sqlDB.Exec(query, roleID)
	if err != nil {
		log.Printf("删除角色失败: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}

func getRoleDetail(sqlDB *sql.DB, roleID string) *gin.H {
	query := `
		SELECT id, role_name, role_description, level, created_at, updated_at
		FROM zervigo_auth_roles 
		WHERE id = $1
	`
	row := sqlDB.QueryRow(query, roleID)

	var id int
	var roleName, roleDescription string
	var level int
	var createdAt, updatedAt time.Time

	err := row.Scan(&id, &roleName, &roleDescription, &level, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Printf("查询角色详情失败: %v", err)
		return nil
	}

	role := &gin.H{
		"id":              id,
		"roleName":        roleName,
		"roleDescription": roleDescription,
		"level":           level,
		"createdAt":       createdAt.UnixMilli(),
		"updatedAt":       updatedAt.UnixMilli(),
	}

	return role
}

func createPermission(sqlDB *sql.DB, req CreatePermissionRequest) string {
	query := `
		INSERT INTO zervigo_auth_permissions (permission_name, permission_code, resource_type, action, 
		                                      permission_description, service_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`

	var permissionID int
	err := sqlDB.QueryRow(query, req.PermissionName, req.PermissionCode, req.ResourceType,
		req.Action, req.PermissionDescription, req.ServiceName).Scan(&permissionID)
	if err != nil {
		log.Printf("创建权限失败: %v", err)
		return ""
	}

	return strconv.Itoa(permissionID)
}

func updatePermission(sqlDB *sql.DB, permissionID string, req UpdatePermissionRequest) bool {
	query := `
		UPDATE zervigo_auth_permissions 
		SET permission_name = $1, permission_code = $2, resource_type = $3, action = $4, 
		    permission_description = $5, service_name = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
	`

	result, err := sqlDB.Exec(query, req.PermissionName, req.PermissionCode, req.ResourceType,
		req.Action, req.PermissionDescription, req.ServiceName, permissionID)
	if err != nil {
		log.Printf("更新权限失败: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}

func deletePermission(sqlDB *sql.DB, permissionID string) bool {
	query := `DELETE FROM zervigo_auth_permissions WHERE id = $1`

	result, err := sqlDB.Exec(query, permissionID)
	if err != nil {
		log.Printf("删除权限失败: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}

func getPermissionDetail(sqlDB *sql.DB, permissionID string) *gin.H {
	query := `
		SELECT id, permission_name, permission_code, resource_type, action, 
		       permission_description, service_name, created_at, updated_at
		FROM zervigo_auth_permissions 
		WHERE id = $1
	`
	row := sqlDB.QueryRow(query, permissionID)

	var id int
	var permissionName, permissionCode, resourceType, action, permissionDescription, serviceName string
	var createdAt, updatedAt time.Time

	err := row.Scan(&id, &permissionName, &permissionCode, &resourceType, &action,
		&permissionDescription, &serviceName, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Printf("查询权限详情失败: %v", err)
		return nil
	}

	permission := &gin.H{
		"id":                    id,
		"permissionName":        permissionName,
		"permissionCode":        permissionCode,
		"resourceType":          resourceType,
		"action":                action,
		"permissionDescription": permissionDescription,
		"serviceName":           serviceName,
		"createdAt":             createdAt.UnixMilli(),
		"updatedAt":             updatedAt.UnixMilli(),
	}

	return permission
}

func getUserRoles(sqlDB *sql.DB, userID uint) []gin.H {
	query := `
		SELECT r.id, r.role_name, r.role_description, r.level, ur.created_at
		FROM zervigo_auth_user_roles ur
		JOIN zervigo_auth_roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1
		ORDER BY r.level DESC, ur.created_at ASC
	`
	rows, err := sqlDB.Query(query, userID)
	if err != nil {
		log.Printf("查询用户角色失败: %v", err)
		return []gin.H{}
	}
	defer rows.Close()

	var roles []gin.H
	for rows.Next() {
		var id int
		var roleName, roleDescription string
		var level int
		var createdAt time.Time

		err := rows.Scan(&id, &roleName, &roleDescription, &level, &createdAt)
		if err != nil {
			log.Printf("扫描用户角色失败: %v", err)
			continue
		}

		role := gin.H{
			"id":              id,
			"roleName":        roleName,
			"roleDescription": roleDescription,
			"level":           level,
			"assignedAt":      createdAt.UnixMilli(),
		}
		roles = append(roles, role)
	}

	return roles
}

func assignRoleToUser(sqlDB *sql.DB, userID uint, roleID string) bool {
	query := `
		INSERT INTO zervigo_auth_user_roles (user_id, role_id, created_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`

	_, err := sqlDB.Exec(query, userID, roleID)
	if err != nil {
		log.Printf("分配角色失败: %v", err)
		return false
	}

	return true
}

func removeRoleFromUser(sqlDB *sql.DB, userID uint, roleID string) bool {
	query := `DELETE FROM zervigo_auth_user_roles WHERE user_id = $1 AND role_id = $2`

	result, err := sqlDB.Exec(query, userID, roleID)
	if err != nil {
		log.Printf("移除角色失败: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}

func getRolePermissions(sqlDB *sql.DB, roleID string) []gin.H {
	query := `
		SELECT p.id, p.permission_name, p.permission_code, p.resource_type, p.action, 
		       p.permission_description, p.service_name, rp.created_at
		FROM zervigo_auth_role_permissions rp
		JOIN zervigo_auth_permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.service_name, p.resource_type, p.action
	`
	rows, err := sqlDB.Query(query, roleID)
	if err != nil {
		log.Printf("查询角色权限失败: %v", err)
		return []gin.H{}
	}
	defer rows.Close()

	var permissions []gin.H
	for rows.Next() {
		var id int
		var permissionName, permissionCode, resourceType, action, permissionDescription, serviceName string
		var createdAt time.Time

		err := rows.Scan(&id, &permissionName, &permissionCode, &resourceType, &action,
			&permissionDescription, &serviceName, &createdAt)
		if err != nil {
			log.Printf("扫描角色权限失败: %v", err)
			continue
		}

		permission := gin.H{
			"id":                    id,
			"permissionName":        permissionName,
			"permissionCode":        permissionCode,
			"resourceType":          resourceType,
			"action":                action,
			"permissionDescription": permissionDescription,
			"serviceName":           serviceName,
			"assignedAt":            createdAt.UnixMilli(),
		}
		permissions = append(permissions, permission)
	}

	return permissions
}

func assignPermissionToRole(sqlDB *sql.DB, roleID, permissionID string) bool {
	query := `
		INSERT INTO zervigo_auth_role_permissions (role_id, permission_id, created_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`

	_, err := sqlDB.Exec(query, roleID, permissionID)
	if err != nil {
		log.Printf("分配权限失败: %v", err)
		return false
	}

	return true
}

func removePermissionFromRole(sqlDB *sql.DB, roleID, permissionID string) bool {
	query := `DELETE FROM zervigo_auth_role_permissions WHERE role_id = $1 AND permission_id = $2`

	result, err := sqlDB.Exec(query, roleID, permissionID)
	if err != nil {
		log.Printf("移除权限失败: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}

// 辅助函数
func registerToConsul(serviceName, serviceHost string, servicePort int) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Printf("创建Consul客户端失败: %v", err)
		return
	}

	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%d", serviceName, servicePort),
		Name:    serviceName,
		Tags:    []string{"permission", "rbac", "role", "auth", "jobfirst", "microservice", "version:3.1.0"},
		Port:    servicePort,
		Address: serviceHost,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", serviceHost, servicePort),
			Timeout:                        "3s",
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	if err := client.Agent().ServiceRegister(registration); err != nil {
		log.Printf("注册服务到Consul失败: %v", err)
	} else {
		log.Printf("%s registered with Consul successfully", serviceName)
	}
}

func standardSuccessResponse(c *gin.Context, data interface{}, message ...string) {
	msg := "操作成功"
	if len(message) > 0 {
		msg = message[0]
	}
	resp := response.Success(msg, data)
	c.JSON(http.StatusOK, resp)
}

func standardErrorResponse(c *gin.Context, statusCode int, message string, details ...string) {
	code := response.CodeInternalError
	switch statusCode {
	case http.StatusBadRequest:
		code = response.CodeInvalidParams
	case http.StatusUnauthorized:
		code = response.CodeUnauthorized
	case http.StatusForbidden:
		code = response.CodeForbidden
	case http.StatusNotFound:
		code = response.CodeNotFound
	}

	resp := response.Error(code, message)
	c.JSON(http.StatusOK, resp) // 使用200状态码，错误信息在响应体中
}
