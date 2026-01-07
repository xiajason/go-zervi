package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/szjason72/zervigo/shared/core/response"
)

func main() {
	// 设置进程名称
	if len(os.Args) > 0 {
		os.Args[0] = "permission-service"
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由器
	r := gin.Default()

	// 设置标准路由
	setupStandardRoutes(r)

	// 设置业务路由
	setupBusinessRoutes(r)

	// 注册到Consul
	registerToConsul("permission-service", "127.0.0.1", 8086)

	// 启动服务
	log.Println("Starting Permission Service with Go-Zervi Framework on 0.0.0.0:8086")
	if err := r.Run(":8086"); err != nil {
		log.Fatalf("Permission Service启动失败: %v", err)
	}
}

// setupStandardRoutes 设置标准路由
func setupStandardRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "permission-service",
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "3.1.0",
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
func setupBusinessRoutes(r *gin.Engine) {
	// 公开API路由（不需要认证）
	public := r.Group("/api/v1")
	{
		// 获取所有角色列表（公开）
		public.GET("/roles", func(c *gin.Context) {
			roles := getAllRoles()
			standardSuccessResponse(c, roles, "角色列表获取成功")
		})

		// 获取所有权限列表（公开）
		public.GET("/permissions", func(c *gin.Context) {
			permissions := getAllPermissions()
			standardSuccessResponse(c, permissions, "权限列表获取成功")
		})

		// 获取角色详情
		public.GET("/roles/:roleId", func(c *gin.Context) {
			roleID := c.Param("roleId")
			role := getRoleDetail(roleID)
			if role == nil {
				standardErrorResponse(c, http.StatusNotFound, "角色不存在", "")
				return
			}
			standardSuccessResponse(c, role, "角色详情获取成功")
		})

		// 获取权限详情
		public.GET("/permissions/:permissionId", func(c *gin.Context) {
			permissionID := c.Param("permissionId")
			permission := getPermissionDetail(permissionID)
			if permission == nil {
				standardErrorResponse(c, http.StatusNotFound, "权限不存在", "")
				return
			}
			standardSuccessResponse(c, permission, "权限详情获取成功")
		})

		// 获取用户角色
		public.GET("/users/:userId/roles", func(c *gin.Context) {
			userID := c.Param("userId")
			roles := getUserRoles(userID)
			standardSuccessResponse(c, roles, "用户角色获取成功")
		})

		// 获取角色权限
		public.GET("/roles/:roleId/permissions", func(c *gin.Context) {
			roleID := c.Param("roleId")
			permissions := getRolePermissions(roleID)
			standardSuccessResponse(c, permissions, "角色权限获取成功")
		})
	}
}

// 业务逻辑函数
func getAllRoles() []gin.H {
	// 模拟数据
	roles := []gin.H{
		{
			"id":              1,
			"roleName":        "super_admin",
			"roleDescription": "超级管理员",
			"level":           4,
			"createdAt":       time.Now().UnixMilli(),
			"updatedAt":       time.Now().UnixMilli(),
		},
		{
			"id":              2,
			"roleName":        "admin",
			"roleDescription": "管理员",
			"level":           3,
			"createdAt":       time.Now().UnixMilli(),
			"updatedAt":       time.Now().UnixMilli(),
		},
		{
			"id":              3,
			"roleName":        "user",
			"roleDescription": "普通用户",
			"level":           1,
			"createdAt":       time.Now().UnixMilli(),
			"updatedAt":       time.Now().UnixMilli(),
		},
		{
			"id":              4,
			"roleName":        "enterprise",
			"roleDescription": "企业用户",
			"level":           2,
			"createdAt":       time.Now().UnixMilli(),
			"updatedAt":       time.Now().UnixMilli(),
		},
	}
	return roles
}

func getAllPermissions() []gin.H {
	// 模拟数据
	permissions := []gin.H{
		{
			"id":                    1,
			"permissionName":        "用户管理",
			"permissionCode":        "user:manage",
			"resourceType":          "user",
			"action":                "manage",
			"permissionDescription": "管理用户信息",
			"serviceName":           "user-service",
			"createdAt":             time.Now().UnixMilli(),
			"updatedAt":             time.Now().UnixMilli(),
		},
		{
			"id":                    2,
			"permissionName":        "简历查看",
			"permissionCode":        "resume:view",
			"resourceType":          "resume",
			"action":                "view",
			"permissionDescription": "查看简历信息",
			"serviceName":           "resume-service",
			"createdAt":             time.Now().UnixMilli(),
			"updatedAt":             time.Now().UnixMilli(),
		},
		{
			"id":                    3,
			"permissionName":        "简历下载",
			"permissionCode":        "resume:download",
			"resourceType":          "resume",
			"action":                "download",
			"permissionDescription": "下载简历文件",
			"serviceName":           "resume-service",
			"createdAt":             time.Now().UnixMilli(),
			"updatedAt":             time.Now().UnixMilli(),
		},
		{
			"id":                    4,
			"permissionName":        "权限管理",
			"permissionCode":        "permission:manage",
			"resourceType":          "permission",
			"action":                "manage",
			"permissionDescription": "管理权限和角色",
			"serviceName":           "permission-service",
			"createdAt":             time.Now().UnixMilli(),
			"updatedAt":             time.Now().UnixMilli(),
		},
		{
			"id":                    5,
			"permissionName":        "系统管理",
			"permissionCode":        "system:manage",
			"resourceType":          "system",
			"action":                "manage",
			"permissionDescription": "管理系统配置",
			"serviceName":           "system-service",
			"createdAt":             time.Now().UnixMilli(),
			"updatedAt":             time.Now().UnixMilli(),
		},
	}
	return permissions
}

func getRoleDetail(roleID string) *gin.H {
	// 模拟数据
	roles := map[string]gin.H{
		"1": {
			"id":              1,
			"roleName":        "super_admin",
			"roleDescription": "超级管理员",
			"level":           4,
			"createdAt":       time.Now().UnixMilli(),
			"updatedAt":       time.Now().UnixMilli(),
		},
		"2": {
			"id":              2,
			"roleName":        "admin",
			"roleDescription": "管理员",
			"level":           3,
			"createdAt":       time.Now().UnixMilli(),
			"updatedAt":       time.Now().UnixMilli(),
		},
		"3": {
			"id":              3,
			"roleName":        "user",
			"roleDescription": "普通用户",
			"level":           1,
			"createdAt":       time.Now().UnixMilli(),
			"updatedAt":       time.Now().UnixMilli(),
		},
		"4": {
			"id":              4,
			"roleName":        "enterprise",
			"roleDescription": "企业用户",
			"level":           2,
			"createdAt":       time.Now().UnixMilli(),
			"updatedAt":       time.Now().UnixMilli(),
		},
	}

	if role, exists := roles[roleID]; exists {
		return &role
	}
	return nil
}

func getPermissionDetail(permissionID string) *gin.H {
	// 模拟数据
	permissions := map[string]gin.H{
		"1": {
			"id":                    1,
			"permissionName":        "用户管理",
			"permissionCode":        "user:manage",
			"resourceType":          "user",
			"action":                "manage",
			"permissionDescription": "管理用户信息",
			"serviceName":           "user-service",
			"createdAt":             time.Now().UnixMilli(),
			"updatedAt":             time.Now().UnixMilli(),
		},
		"2": {
			"id":                    2,
			"permissionName":        "简历查看",
			"permissionCode":        "resume:view",
			"resourceType":          "resume",
			"action":                "view",
			"permissionDescription": "查看简历信息",
			"serviceName":           "resume-service",
			"createdAt":             time.Now().UnixMilli(),
			"updatedAt":             time.Now().UnixMilli(),
		},
		"3": {
			"id":                    3,
			"permissionName":        "简历下载",
			"permissionCode":        "resume:download",
			"resourceType":          "resume",
			"action":                "download",
			"permissionDescription": "下载简历文件",
			"serviceName":           "resume-service",
			"createdAt":             time.Now().UnixMilli(),
			"updatedAt":             time.Now().UnixMilli(),
		},
		"4": {
			"id":                    4,
			"permissionName":        "权限管理",
			"permissionCode":        "permission:manage",
			"resourceType":          "permission",
			"action":                "manage",
			"permissionDescription": "管理权限和角色",
			"serviceName":           "permission-service",
			"createdAt":             time.Now().UnixMilli(),
			"updatedAt":             time.Now().UnixMilli(),
		},
		"5": {
			"id":                    5,
			"permissionName":        "系统管理",
			"permissionCode":        "system:manage",
			"resourceType":          "system",
			"action":                "manage",
			"permissionDescription": "管理系统配置",
			"serviceName":           "system-service",
			"createdAt":             time.Now().UnixMilli(),
			"updatedAt":             time.Now().UnixMilli(),
		},
	}

	if permission, exists := permissions[permissionID]; exists {
		return &permission
	}
	return nil
}

func getUserRoles(userID string) []gin.H {
	// 模拟数据 - 根据用户ID返回不同角色
	userRoles := map[string][]gin.H{
		"1": { // admin用户
			{
				"id":              1,
				"roleName":        "super_admin",
				"roleDescription": "超级管理员",
				"level":           4,
				"assignedAt":      time.Now().UnixMilli(),
			},
		},
		"2": { // 普通用户
			{
				"id":              3,
				"roleName":        "user",
				"roleDescription": "普通用户",
				"level":           1,
				"assignedAt":      time.Now().UnixMilli(),
			},
		},
		"3": { // 企业用户
			{
				"id":              4,
				"roleName":        "enterprise",
				"roleDescription": "企业用户",
				"level":           2,
				"assignedAt":      time.Now().UnixMilli(),
			},
		},
	}

	if roles, exists := userRoles[userID]; exists {
		return roles
	}

	// 默认返回普通用户角色
	return []gin.H{
		{
			"id":              3,
			"roleName":        "user",
			"roleDescription": "普通用户",
			"level":           1,
			"assignedAt":      time.Now().UnixMilli(),
		},
	}
}

func getRolePermissions(roleID string) []gin.H {
	// 模拟数据 - 根据角色ID返回不同权限
	rolePermissions := map[string][]gin.H{
		"1": { // super_admin - 所有权限
			{
				"id":                    1,
				"permissionName":        "用户管理",
				"permissionCode":        "user:manage",
				"resourceType":          "user",
				"action":                "manage",
				"permissionDescription": "管理用户信息",
				"serviceName":           "user-service",
				"assignedAt":            time.Now().UnixMilli(),
			},
			{
				"id":                    2,
				"permissionName":        "简历查看",
				"permissionCode":        "resume:view",
				"resourceType":          "resume",
				"action":                "view",
				"permissionDescription": "查看简历信息",
				"serviceName":           "resume-service",
				"assignedAt":            time.Now().UnixMilli(),
			},
			{
				"id":                    3,
				"permissionName":        "简历下载",
				"permissionCode":        "resume:download",
				"resourceType":          "resume",
				"action":                "download",
				"permissionDescription": "下载简历文件",
				"serviceName":           "resume-service",
				"assignedAt":            time.Now().UnixMilli(),
			},
			{
				"id":                    4,
				"permissionName":        "权限管理",
				"permissionCode":        "permission:manage",
				"resourceType":          "permission",
				"action":                "manage",
				"permissionDescription": "管理权限和角色",
				"serviceName":           "permission-service",
				"assignedAt":            time.Now().UnixMilli(),
			},
			{
				"id":                    5,
				"permissionName":        "系统管理",
				"permissionCode":        "system:manage",
				"resourceType":          "system",
				"action":                "manage",
				"permissionDescription": "管理系统配置",
				"serviceName":           "system-service",
				"assignedAt":            time.Now().UnixMilli(),
			},
		},
		"2": { // admin - 大部分权限
			{
				"id":                    1,
				"permissionName":        "用户管理",
				"permissionCode":        "user:manage",
				"resourceType":          "user",
				"action":                "manage",
				"permissionDescription": "管理用户信息",
				"serviceName":           "user-service",
				"assignedAt":            time.Now().UnixMilli(),
			},
			{
				"id":                    2,
				"permissionName":        "简历查看",
				"permissionCode":        "resume:view",
				"resourceType":          "resume",
				"action":                "view",
				"permissionDescription": "查看简历信息",
				"serviceName":           "resume-service",
				"assignedAt":            time.Now().UnixMilli(),
			},
			{
				"id":                    3,
				"permissionName":        "简历下载",
				"permissionCode":        "resume:download",
				"resourceType":          "resume",
				"action":                "download",
				"permissionDescription": "下载简历文件",
				"serviceName":           "resume-service",
				"assignedAt":            time.Now().UnixMilli(),
			},
		},
		"3": { // user - 基本权限
			{
				"id":                    2,
				"permissionName":        "简历查看",
				"permissionCode":        "resume:view",
				"resourceType":          "resume",
				"action":                "view",
				"permissionDescription": "查看简历信息",
				"serviceName":           "resume-service",
				"assignedAt":            time.Now().UnixMilli(),
			},
		},
		"4": { // enterprise - 企业权限
			{
				"id":                    2,
				"permissionName":        "简历查看",
				"permissionCode":        "resume:view",
				"resourceType":          "resume",
				"action":                "view",
				"permissionDescription": "查看简历信息",
				"serviceName":           "resume-service",
				"assignedAt":            time.Now().UnixMilli(),
			},
			{
				"id":                    3,
				"permissionName":        "简历下载",
				"permissionCode":        "resume:download",
				"resourceType":          "resume",
				"action":                "download",
				"permissionDescription": "下载简历文件",
				"serviceName":           "resume-service",
				"assignedAt":            time.Now().UnixMilli(),
			},
		},
	}

	if permissions, exists := rolePermissions[roleID]; exists {
		return permissions
	}

	// 默认返回空权限
	return []gin.H{}
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
