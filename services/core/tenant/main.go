package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	jobfirst "github.com/szjason72/zervigo/shared/core"
	"github.com/szjason72/zervigo/shared/core/auth"
	"github.com/szjason72/zervigo/shared/core/middleware"
)

func main() {
	// 设置进程名称
	if len(os.Args) > 0 {
		os.Args[0] = "tenant-service"
	}

	// 初始化JobFirst核心包
	core, err := jobfirst.NewCore("")
	if err != nil {
		log.Fatalf("初始化JobFirst核心包失败: %v", err)
	}
	defer core.Close()

	// 获取数据库连接
	db := core.GetDB()
	if db == nil {
		log.Fatal("未能获取数据库连接，无法启动 tenant-service")
	}

	// 创建租户服务
	tenantService := NewTenantService(db)

	// 初始化默认租户（如果不存在）
	if err := ensureDefaultTenant(context.Background(), tenantService); err != nil {
		log.Printf("WARN: 初始化默认租户失败: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.Default()

	// 设置标准路由
	setupStandardRoutes(r, core)

	// 创建认证客户端
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8207"
	}
	authClient := auth.NewAuthClient(authServiceURL)

	// 设置业务路由
	setupBusinessRoutes(r, core, authClient, tenantService)

	// 注册到Consul
	port := 8088
	if portEnv := os.Getenv("TENANT_SERVICE_PORT"); portEnv != "" {
		fmt.Sscanf(portEnv, "%d", &port)
	}
	registerToConsul("tenant-service", "127.0.0.1", port)

	// 启动服务器
	log.Printf("Starting Tenant Service on 0.0.0.0:%d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Tenant Service启动失败: %v", err)
	}
}

// setupStandardRoutes 设置标准路由
func setupStandardRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		health := core.Health()
		c.JSON(http.StatusOK, gin.H{
			"service":     "tenant-service",
			"status":      "healthy",
			"timestamp":   time.Now().Format(time.RFC3339),
			"version":     "1.0.0",
			"core_health": health,
		})
	})

	// 版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "tenant-service",
			"version": "1.0.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 服务信息
	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":    "tenant-service",
			"version":    "1.0.0",
			"port":       8088,
			"status":     "running",
			"started_at": time.Now().Format(time.RFC3339),
		})
	})
}

// setupBusinessRoutes 设置业务路由
func setupBusinessRoutes(r *gin.Engine, core *jobfirst.Core, authClient *auth.AuthClient, tenantService *TenantService) {
	// 创建认证中间件
	authMiddleware := createAuthMiddleware(authClient)

	// 创建租户中间件
	tenantMiddleware := middleware.TenantMiddleware()

	// API路由组
	api := r.Group("/api/v1")
	api.Use(authMiddleware) // 需要认证
	api.Use(tenantMiddleware) // 租户中间件（可选，某些API可能需要）

	// 创建租户API处理器
	tenantAPI := NewTenantAPI(tenantService)
	tenantAPI.RegisterRoutes(api)
}

// createAuthMiddleware 创建认证中间件
func createAuthMiddleware(authClient *auth.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头提取token
		token := extractTokenFromRequest(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "未登录",
			})
			c.Abort()
			return
		}

		// 调用auth-service验证token
		result, err := authClient.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "认证服务不可用",
			})
			c.Abort()
			return
		}

		if !result.Success {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   result.Error,
			})
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", result.User.ID)
		c.Set("username", result.User.Username)
		c.Set("role", result.User.Role)
		c.Set("email", result.User.Email)
		c.Set("user", result.User)

		c.Next()
	}
}

// extractTokenFromRequest 从请求中提取token
func extractTokenFromRequest(c *gin.Context) string {
	// 从Authorization头获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// 兼容大小写并按空格分割
		fields := strings.Fields(authHeader)
		if len(fields) == 2 && strings.EqualFold(fields[0], "Bearer") {
			return fields[1]
		}
	}

	// 从查询参数获取
	token := c.Query("token")
	if token != "" {
		return token
	}

	// 从Cookie获取
	cookie, err := c.Cookie("token")
	if err == nil && cookie != "" {
		return cookie
	}

	return ""
}

// ensureDefaultTenant 确保默认租户存在
func ensureDefaultTenant(ctx context.Context, service *TenantService) error {
	// 检查默认租户是否存在
	_, err := service.GetTenant(ctx, 1)
	if err == nil {
		// 默认租户已存在
		return nil
	}

	if err != ErrTenantNotFound {
		return err
	}

	// 创建默认租户（如果不存在）
	// 注意：这里需要系统管理员权限，暂时跳过
	log.Println("INFO: 默认租户不存在，请通过API创建或运行数据库迁移脚本")
	return nil
}

// registerToConsul 注册服务到Consul
func registerToConsul(serviceName, serviceHost string, servicePort int) {
	config := api.DefaultConfig()
	config.Address = "localhost:8500"
	client, err := api.NewClient(config)
	if err != nil {
		log.Printf("创建Consul客户端失败: %v", err)
		return
	}

	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%d", serviceName, servicePort),
		Name:    serviceName,
		Tags:    []string{"tenant", "saas", "microservice", "version:1.0.0"},
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

