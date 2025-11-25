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
	"github.com/szjason72/zervigo/shared/core/database"
	"github.com/szjason72/zervigo/shared/core/response"
)

func main() {
	// 设置进程名称
	if len(os.Args) > 0 {
		os.Args[0] = "rls-demo-service"
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
	registerToConsul("rls-demo-service", "127.0.0.1", 8088)

	// 启动服务
	log.Println("Starting RLS Demo Service with Go-Zervi Framework on 0.0.0.0:8088")
	if err := r.Run(":8088"); err != nil {
		log.Fatalf("RLS Demo Service启动失败: %v", err)
	}
}

// setupStandardRoutes 设置标准路由
func setupStandardRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		health := core.Health()
		c.JSON(http.StatusOK, gin.H{
			"service":     "rls-demo-service",
			"status":      "healthy",
			"timestamp":   time.Now().Format(time.RFC3339),
			"version":     "3.1.0",
			"core_health": health,
		})
	})

	// 版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "rls-demo-service",
			"version": "3.1.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})
}

// setupBusinessRoutes 设置业务路由
func setupBusinessRoutes(r *gin.Engine, core *jobfirst.Core, sqlDB *sql.DB, jwtSecret string) {
	// 公开API路由（不需要认证）
	public := r.Group("/api/v1/rls")
	{
		// 测试RLS策略
		public.GET("/test-policies", func(c *gin.Context) {
			policies := testRLSPolicies(sqlDB)
			standardSuccessResponse(c, policies, "RLS策略测试完成")
		})

		// 测试权限检查
		public.GET("/test-permissions/:user_id", func(c *gin.Context) {
			userIDStr := c.Param("user_id")
			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				standardErrorResponse(c, http.StatusBadRequest, "无效的用户ID", err.Error())
				return
			}

			permissions := testUserPermissions(sqlDB, uint(userID))
			standardSuccessResponse(c, permissions, "用户权限测试完成")
		})
	}

	// 需要认证的API路由
	zerviAuthAdapter := auth.NewZerviAuthAdapter(sqlDB, jwtSecret)
	authMiddleware := zerviAuthAdapter.RequireAuth()
	api := r.Group("/api/v1/rls")
	api.Use(authMiddleware)
	{
		// 演示权限感知的数据库查询
		api.GET("/demo/permission-aware-query", func(c *gin.Context) {
			userID := c.GetUint("user_id")
			if userID == 0 {
				standardErrorResponse(c, http.StatusUnauthorized, "未登录", "")
				return
			}

			// 创建用户上下文
			userCtx := &database.UserContext{
				UserID:       userID,
				Username:     c.GetString("username"),
				Roles:        []string{"user"},                         // 简化处理
				Permissions:  []string{"resume:view", "resume:create"}, // 简化处理
				DatabaseRole: "zervigo_user",
			}

			// 创建权限感知的数据库管理器
			pam := core.Database.NewPermissionAwareManager(userCtx)

			// 设置用户上下文到数据库
			if err := setUserContext(sqlDB, userID, userCtx.Username); err != nil {
				standardErrorResponse(c, http.StatusInternalServerError, "设置用户上下文失败", err.Error())
				return
			}

			// 执行权限感知查询
			results := performPermissionAwareQueries(pam)
			standardSuccessResponse(c, results, "权限感知查询完成")
		})

		// 演示RLS策略效果
		api.GET("/demo/rls-effect", func(c *gin.Context) {
			userID := c.GetUint("user_id")
			if userID == 0 {
				standardErrorResponse(c, http.StatusUnauthorized, "未登录", "")
				return
			}

			// 设置用户上下文
			if err := setUserContext(sqlDB, userID, c.GetString("username")); err != nil {
				standardErrorResponse(c, http.StatusInternalServerError, "设置用户上下文失败", err.Error())
				return
			}

			// 演示RLS效果
			results := demonstrateRLSEffect(sqlDB, userID)
			standardSuccessResponse(c, results, "RLS效果演示完成")
		})
	}
}

// testRLSPolicies 测试RLS策略
func testRLSPolicies(sqlDB *sql.DB) map[string]interface{} {
	// 查询RLS策略状态
	query := `
		SELECT schemaname, tablename, policyname, permissive
		FROM pg_policies 
		WHERE schemaname = 'public'
		ORDER BY tablename, policyname
	`

	rows, err := sqlDB.Query(query)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}
	defer rows.Close()

	var policies []map[string]interface{}
	for rows.Next() {
		var schemaName, tableName, policyName string
		var permissive bool

		if err := rows.Scan(&schemaName, &tableName, &policyName, &permissive); err != nil {
			continue
		}

		policies = append(policies, map[string]interface{}{
			"schema":     schemaName,
			"table":      tableName,
			"policy":     policyName,
			"permissive": permissive,
		})
	}

	return map[string]interface{}{
		"total_policies": len(policies),
		"policies":       policies,
		"status":         "active",
	}
}

// testUserPermissions 测试用户权限
func testUserPermissions(sqlDB *sql.DB, userID uint) map[string]interface{} {
	// 设置用户上下文
	if err := setUserContext(sqlDB, userID, "test_user"); err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	// 测试权限检查函数
	var hasResumeView bool
	err := sqlDB.QueryRow("SELECT has_permission('resume:view')").Scan(&hasResumeView)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	var hasUserRole bool
	err = sqlDB.QueryRow("SELECT has_role('user')").Scan(&hasUserRole)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	// 获取用户权限列表
	permissionsQuery := `
		SELECT p.permission_code, p.permission_name
		FROM zervigo_auth_user_roles ur
		JOIN zervigo_auth_role_permissions rp ON ur.role_id = rp.role_id
		JOIN zervigo_auth_permissions p ON rp.permission_id = p.id
		WHERE ur.user_id = $1
		ORDER BY p.permission_code
	`

	rows, err := sqlDB.Query(permissionsQuery, userID)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}
	defer rows.Close()

	var permissions []map[string]interface{}
	for rows.Next() {
		var code, name string
		if err := rows.Scan(&code, &name); err != nil {
			continue
		}
		permissions = append(permissions, map[string]interface{}{
			"code": code,
			"name": name,
		})
	}

	return map[string]interface{}{
		"user_id":          userID,
		"has_resume_view":  hasResumeView,
		"has_user_role":    hasUserRole,
		"permissions":      permissions,
		"permission_count": len(permissions),
	}
}

// setUserContext 设置用户上下文
func setUserContext(sqlDB *sql.DB, userID uint, username string) error {
	query := "SELECT set_user_context($1, $2)"
	_, err := sqlDB.Exec(query, userID, username)
	return err
}

// performPermissionAwareQueries 执行权限感知查询
func performPermissionAwareQueries(pam *database.PermissionAwareManager) map[string]interface{} {
	results := make(map[string]interface{})

	// 创建权限感知查询
	paq, err := pam.NewPermissionAwareQuery()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	// 查询用户信息
	var userCount int
	err = paq.db.Table("zervigo_auth_users").Count(&userCount).Error
	if err != nil {
		results["user_query_error"] = err.Error()
	} else {
		results["user_count"] = userCount
	}

	// 查询简历权限
	var permissionCount int
	err = paq.db.Table("resume_permission").Count(&permissionCount).Error
	if err != nil {
		results["permission_query_error"] = err.Error()
	} else {
		results["permission_count"] = permissionCount
	}

	// 查询审批记录
	var approvalCount int
	err = paq.db.Table("approve_record").Count(&approvalCount).Error
	if err != nil {
		results["approval_query_error"] = err.Error()
	} else {
		results["approval_count"] = approvalCount
	}

	// 查询积分记录
	var pointsCount int
	err = paq.db.Table("points_bill").Count(&pointsCount).Error
	if err != nil {
		results["points_query_error"] = err.Error()
	} else {
		results["points_count"] = pointsCount
	}

	results["database_type"] = paq.GetDatabaseType()
	results["capabilities"] = paq.GetCapabilities()
	results["current_user"] = paq.GetCurrentUser()

	return results
}

// demonstrateRLSEffect 演示RLS效果
func demonstrateRLSEffect(sqlDB *sql.DB, userID uint) map[string]interface{} {
	results := make(map[string]interface{})

	// 设置用户上下文
	if err := setUserContext(sqlDB, userID, "test_user"); err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	// 查询不同表的数据量（受RLS影响）
	tables := []string{
		"zervigo_auth_users",
		"resume_permission",
		"approve_record",
		"points_bill",
		"view_history",
	}

	tableCounts := make(map[string]int)
	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		err := sqlDB.QueryRow(query).Scan(&count)
		if err != nil {
			tableCounts[table] = -1 // 表示查询失败
		} else {
			tableCounts[table] = count
		}
	}

	results["user_id"] = userID
	results["table_counts"] = tableCounts
	results["rls_active"] = true
	results["note"] = "数据量受RLS策略影响，不同用户看到的数据可能不同"

	return results
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
		Tags:    []string{"rls", "demo", "permission", "postgresql", "jobfirst", "microservice", "version:3.1.0"},
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
