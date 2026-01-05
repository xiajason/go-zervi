package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/szjason72/zervigo/shared/core/shared"

	"github.com/szjason72/zervigo/shared/central-brain/client"
	"github.com/szjason72/zervigo/shared/central-brain/middleware"
	"github.com/szjason72/zervigo/shared/central-brain/permission"
	"github.com/szjason72/zervigo/shared/central-brain/router"
	"github.com/szjason72/zervigo/shared/central-brain/utils"
)

// CentralBrain 中央大脑服务
type CentralBrain struct {
	config           *shared.Config
	httpClient       *http.Client
	clientPool       *client.HTTPClientPool // HTTP客户端连接池
	router           *gin.Engine
	authServiceURL   string                       // Auth Service的URL
	routerClient     *router.RouterClient         // Router Service客户端
	permissionClient *permission.PermissionClient // Permission Service客户端

	// 中间件组件
	requestLogger   *middleware.RequestLogger
	metrics         *middleware.Metrics
	rateLimiter     *middleware.RateLimiter
	circuitBreakers map[string]*middleware.CircuitBreaker

	// 服务token相关（带互斥锁保护）
	tokenMu                sync.RWMutex // 保护serviceToken和serviceTokenExp的并发访问
	serviceToken           string       // 缓存的服务token
	serviceTokenExp        time.Time    // token过期时间
	tokenRefreshInProgress bool         // 标记是否正在刷新token（防止并发刷新）
}

// ServiceProxy 服务代理配置
type ServiceProxy struct {
	ServiceName       string
	BaseURL           string
	PathPrefix        string
	TargetPrefix      string
	CircuitBreakerKey string
	Rewrite           map[string]string
}

// NewCentralBrain 创建中央大脑服务
func NewCentralBrain(config *shared.Config) *CentralBrain {
	// 数据库连接检查
	if config.DatabaseCheck.Enabled {
		fmt.Println("🔍 检查数据库连接...")
		checker := shared.NewDatabaseChecker(config)
		result, err := checker.CheckDatabase()

		if err != nil {
			errorMsg := shared.FormatDatabaseError(result)
			if config.DatabaseCheck.Required {
				// 必需模式：失败时阻止启动
				fmt.Printf("❌ 数据库连接检查失败（必需）:\n%s", errorMsg)
				panic(fmt.Sprintf("数据库连接检查失败（必需）: %v", err))
			} else {
				// 可选模式：失败时记录警告但继续启动
				fmt.Printf("⚠️ 数据库连接检查失败（可选）:\n%s", errorMsg)
			}
		} else {
			if result.Status == "connected" {
				fmt.Printf("✅ 数据库连接检查成功: %s (耗时: %v)\n", result.Message, result.Duration)
			} else if result.Status == "not_configured" {
				fmt.Printf("ℹ️  数据库未配置: %s\n", result.Message)
			}
			if len(result.Warnings) > 0 {
				for _, warning := range result.Warnings {
					fmt.Printf("   ⚠️ %s\n", warning)
				}
			}
		}
	}

	// 构建服务URL（使用配置的服务主机，支持Docker网络）
	serviceHost := config.ServiceDiscovery.ServiceHost
	authServiceURL := fmt.Sprintf("http://%s:%d", serviceHost, config.AuthServicePort)
	routerServiceURL := fmt.Sprintf("http://%s:%d", serviceHost, config.RouterServicePort)
	permissionServiceURL := fmt.Sprintf("http://%s:%d", serviceHost, config.PermissionServicePort)

	// 初始化HTTP客户端连接池
	clientPool := client.NewHTTPClientPool()

	// 初始化Router Service客户端
	routerClient := router.NewRouterClient(routerServiceURL)

	// 初始化Permission Service客户端
	permissionClient := permission.NewPermissionClient(permissionServiceURL)

	// 初始化中间件
	requestLogger := middleware.NewRequestLogger(true) // 启用日志
	metrics := middleware.NewMetrics()
	rateLimiter := middleware.NewRateLimiter(100, 200, true) // 100 RPS, 200 burst

	// 为每个服务创建熔断器
	circuitBreakers := make(map[string]*middleware.CircuitBreaker)
	serviceNames := []string{"auth", "ai", "blockchain", "user", "job", "resume", "company"}
	for _, serviceName := range serviceNames {
		circuitBreakers[serviceName] = middleware.NewCircuitBreaker(
			5,              // 失败阈值：5次
			60*time.Second, // 重置超时：60秒
			3,              // 半开状态成功阈值：3次
		)
	}

	cb := &CentralBrain{
		config:           config,
		httpClient:       clientPool.GetDefaultClient(), // 使用连接池的默认客户端
		clientPool:       clientPool,
		router:           gin.Default(),
		authServiceURL:   authServiceURL,
		routerClient:     routerClient,
		permissionClient: permissionClient,
		requestLogger:    requestLogger,
		metrics:          metrics,
		rateLimiter:      rateLimiter,
		circuitBreakers:  circuitBreakers,
	}

	// 启动时获取服务token（带重试机制）
	go cb.initializeServiceTokenWithRetry()

	return cb
}

// Start 启动中央大脑服务
func (cb *CentralBrain) Start() error {
	// 配置CORS（必须在最前面）
	cb.router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, accessToken")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 注册基础设施中间件（按顺序）
	cb.router.Use(cb.requestLogger.Middleware()) // 请求日志（第一层）
	cb.router.Use(cb.metrics.Middleware())       // 性能指标（第二层）
	cb.router.Use(cb.rateLimiter.Middleware())   // 限流（第三层）

	// 注册服务代理（带熔断器保护）
	cb.registerServiceProxies()

	// 注册管理API（健康检查、指标查询）
	cb.registerManagementRoutes()

	return cb.router.Run(fmt.Sprintf(":%d", cb.config.CentralBrainPort))
}

// registerManagementRoutes 注册管理API路由
func (cb *CentralBrain) registerManagementRoutes() {
	// 健康检查
	cb.router.GET("/health", cb.healthCheck)

	// 指标查询
	cb.router.GET("/api/v1/metrics", cb.getMetrics)

	// 熔断器状态
	cb.router.GET("/api/v1/circuit-breakers", cb.getCircuitBreakers)

	// Router和Permission服务通过代理提供API，不需要单独注册管理路由
}

// registerServiceProxies 注册服务代理
func (cb *CentralBrain) registerServiceProxies() {
	serviceHost := cb.config.ServiceDiscovery.ServiceHost

	services := []ServiceProxy{
		{
			ServiceName:       "auth-service",
			CircuitBreakerKey: "auth",
			BaseURL:           fmt.Sprintf("http://%s:%d", serviceHost, cb.config.AuthServicePort),
			PathPrefix:        "/api/v1/auth",
		},
		{
			ServiceName:       "auth-service",
			CircuitBreakerKey: "auth",
			BaseURL:           fmt.Sprintf("http://%s:%d", serviceHost, cb.config.AuthServicePort),
			PathPrefix:        "/api/auth",
			TargetPrefix:      "/api/v1/auth",
		},
		{
			ServiceName:       "router-service",
			CircuitBreakerKey: "router",
			BaseURL:           fmt.Sprintf("http://%s:%d", serviceHost, cb.config.RouterServicePort),
			PathPrefix:        "/api/v1/router",
		},
		{
			ServiceName:       "permission-service",
			CircuitBreakerKey: "permission",
			BaseURL:           fmt.Sprintf("http://%s:%d", serviceHost, cb.config.PermissionServicePort),
			PathPrefix:        "/api/v1/permission",
		},
		{
			ServiceName:       "ai-service",
			CircuitBreakerKey: "ai",
			BaseURL:           fmt.Sprintf("http://%s:%d", serviceHost, cb.config.AIServicePort),
			PathPrefix:        "/api/v1/ai",
		},
		{
			ServiceName:       "blockchain-service",
			CircuitBreakerKey: "blockchain",
			BaseURL:           fmt.Sprintf("http://%s:%d", serviceHost, cb.config.BlockchainServicePort),
			PathPrefix:        "/api/v1/blockchain",
		},
		{
			ServiceName:       "user-service",
			CircuitBreakerKey: "user",
			BaseURL:           fmt.Sprintf("http://%s:%d", serviceHost, cb.config.UserServicePort),
			PathPrefix:        "/api/v1/users",
		},
		{
			ServiceName:       "user-service",
			CircuitBreakerKey: "user",
			BaseURL:           fmt.Sprintf("http://%s:%d", serviceHost, cb.config.UserServicePort),
			PathPrefix:        "/api/user",
			TargetPrefix:      "/api/v1/users",
			Rewrite: map[string]string{
				"/current": "/api/v1/users/profile",
			},
		},
		{
			ServiceName:       "job-service",
			CircuitBreakerKey: "job",
			BaseURL:           fmt.Sprintf("http://%s:%d", serviceHost, cb.config.JobServicePort),
			PathPrefix:        "/api/v1/job",
		},
		{
			ServiceName:       "resume-service",
			CircuitBreakerKey: "resume",
			BaseURL:           fmt.Sprintf("http://%s:%d", serviceHost, cb.config.ResumeServicePort),
			PathPrefix:        "/api/v1/resume",
		},
		{
			ServiceName:       "resume-service",
			CircuitBreakerKey: "resume",
			BaseURL:           fmt.Sprintf("http://%s:%d", serviceHost, cb.config.ResumeServicePort),
			PathPrefix:        "/api/resume",
			TargetPrefix:      "/api/v1/resume",
		},
		{
			ServiceName:       "company-service",
			CircuitBreakerKey: "company",
			BaseURL:           fmt.Sprintf("http://%s:%d", serviceHost, cb.config.CompanyServicePort),
			PathPrefix:        "/api/v1/company",
		},
	}

	for _, service := range services {
		cb.registerServiceProxy(service)
	}
}

// registerServiceProxy 注册单个服务代理
func (cb *CentralBrain) registerServiceProxy(service ServiceProxy) {
	proxyGroup := cb.router.Group(service.PathPrefix)

	circuitKey := service.CircuitBreakerKey
	if circuitKey == "" {
		circuitKey = service.ServiceName
	}

	if circuitBreaker, exists := cb.circuitBreakers[circuitKey]; exists {
		proxyGroup.Use(circuitBreaker.Middleware(service.ServiceName))
	}

	proxyGroup.Any("/*path", func(sp ServiceProxy) gin.HandlerFunc {
		return func(c *gin.Context) {
			cb.proxyRequest(c, sp)
		}
	}(service))

	targetPrefix := service.TargetPrefix
	if targetPrefix == "" {
		targetPrefix = service.PathPrefix
	}

	fmt.Printf("✅ 注册服务代理: %s -> %s%s\n", service.PathPrefix, service.BaseURL, targetPrefix)
}

// proxyRequest 代理请求
func (cb *CentralBrain) proxyRequest(c *gin.Context, service ServiceProxy) {
	originalPath := c.Request.URL.Path
	relativePath := strings.TrimPrefix(originalPath, service.PathPrefix)
	if relativePath == "" {
		relativePath = "/"
	} else if !strings.HasPrefix(relativePath, "/") {
		relativePath = "/" + relativePath
	}

	targetPath := ""
	if service.Rewrite != nil {
		if rewrite, ok := service.Rewrite[relativePath]; ok {
			targetPath = rewrite
		}
	}

	if targetPath == "" {
		targetPrefix := service.TargetPrefix
		if targetPrefix == "" {
			targetPrefix = service.PathPrefix
		}

		if relativePath == "/" {
			targetPath = targetPrefix
		} else {
			targetPath = path.Join(targetPrefix, relativePath)
		}
	}

	baseURL := strings.TrimSuffix(service.BaseURL, "/")
	targetURL := baseURL + targetPath
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	fmt.Printf("🔄 代理请求: %s -> %s\n", originalPath, targetURL)

	// 3. 读取请求体
	var body []byte
	if c.Request.Body != nil {
		body, _ = io.ReadAll(c.Request.Body)
	}

	// 4. 创建HTTP请求
	req, err := http.NewRequestWithContext(c.Request.Context(),
		c.Request.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		cb.handleError(c, fmt.Errorf("创建请求失败: %v", err))
		return
	}

	// 调试：记录Authorization头的透传情况
	incomingAuth := c.Request.Header.Get("Authorization")
	fmt.Printf("DEBUG Gateway: incoming Authorization: %s\n", func() string {
		if incomingAuth == "" {
			return "<empty>"
		}
		if len(incomingAuth) > 60 {
			return incomingAuth[:60] + "..."
		}
		return incomingAuth
	}())

	// 5. 复制请求头（保留用户token）
	for key, values := range c.Request.Header {
		// 跳过某些内部头
		if strings.EqualFold(key, "X-Service-Token") || strings.EqualFold(key, "X-Service-ID") {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 5.1 验证用户token（如果存在）- 使用jobfirst-2024密钥
	userToken := cb.extractUserToken(c.Request)
	if userToken != "" {
		// 这里可以验证用户token，但为了性能，我们直接转发给目标服务验证
		// 目标服务会验证用户token（jobfirst-2024）
		req.Header.Set("Authorization", "Bearer "+userToken)
	}

	// 调试：记录下游请求Authorization
	outgoingAuth := req.Header.Get("Authorization")
	fmt.Printf("DEBUG Gateway: outgoing Authorization: %s\n", func() string {
		if outgoingAuth == "" {
			return "<empty>"
		}
		if len(outgoingAuth) > 60 {
			return outgoingAuth[:60] + "..."
		}
		return outgoingAuth
	}())

	// 5.2 添加服务token（zervigo-2025）- 用于服务间认证
	serviceToken := cb.getServiceToken()
	if serviceToken != "" {
		req.Header.Set("X-Service-Token", serviceToken)
		req.Header.Set("X-Service-ID", "central-brain")
		req.Header.Set("X-Service-Name", "Central Brain")
	}

	// 6. 发送请求（使用连接池的客户端）
	client := cb.clientPool.GetClient(service.ServiceName)
	resp, err := client.Do(req)
	if err != nil {
		cb.handleError(c, fmt.Errorf("请求失败: %v", err))
		return
	}
	defer resp.Body.Close()

	// 7. 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		cb.handleError(c, fmt.Errorf("读取响应失败: %v", err))
		return
	}

	// 8. 复制响应头（过滤冲突头）
	for key, values := range resp.Header {
		if !cb.isFilteredHeader(key) {
			for _, value := range values {
				c.Header(key, value)
			}
		}
	}

	// 9. 返回响应
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// isFilteredHeader 检查是否为需要过滤的响应头
func (cb *CentralBrain) isFilteredHeader(key string) bool {
	filteredHeaders := []string{
		"Transfer-Encoding",
		"Content-Length",
		"Connection",
		"Server",
	}

	for _, filtered := range filteredHeaders {
		if strings.EqualFold(key, filtered) {
			return true
		}
	}
	return false
}

// handleError 处理错误
func (cb *CentralBrain) handleError(c *gin.Context, err error) {
	fmt.Printf("❌ 代理错误: %v\n", err)

	// 获取追踪ID
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	// 使用统一的错误响应格式
	utils.WriteErrorResponse(c.Writer, http.StatusInternalServerError,
		fmt.Sprintf("中央大脑代理失败: %v", err), traceID)
}

// healthCheck 健康检查
func (cb *CentralBrain) healthCheck(c *gin.Context) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	utils.WriteSuccessResponse(c.Writer, "中央大脑服务健康", gin.H{
		"service":   "central-brain",
		"status":    "UP",
		"version":   "1.0.0",
		"timestamp": time.Now().Unix(),
	}, traceID)
}

// getMetrics 获取性能指标
func (cb *CentralBrain) getMetrics(c *gin.Context) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	stats := cb.metrics.GetStats()
	utils.WriteSuccessResponse(c.Writer, "指标获取成功", stats, traceID)
}

// getCircuitBreakers 获取熔断器状态
func (cb *CentralBrain) getCircuitBreakers(c *gin.Context) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	breakerStats := make(map[string]interface{})
	for serviceName, breaker := range cb.circuitBreakers {
		breakerStats[serviceName] = breaker.GetStats()
	}

	utils.WriteSuccessResponse(c.Writer, "熔断器状态获取成功", breakerStats, traceID)
}

// registerRouterRoutes 注册Router Service路由管理API
func (cb *CentralBrain) registerRouterRoutes() {
	// 公开API：获取所有路由配置
	cb.router.GET("/api/v1/router/routes", cb.getAllRoutes)

	// 公开API：获取所有页面配置
	cb.router.GET("/api/v1/router/pages", cb.getAllPages)

	// 需要认证的API：获取用户可访问的路由
	cb.router.GET("/api/v1/router/user-routes", cb.getUserRoutes)

	// 需要认证的API：获取用户可访问的页面
	cb.router.GET("/api/v1/router/user-pages", cb.getUserPages)
}

// getAllRoutes 获取所有路由配置（公开）
func (cb *CentralBrain) getAllRoutes(c *gin.Context) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	routes, err := cb.routerClient.GetAllRoutes()
	if err != nil {
		utils.WriteErrorResponse(c.Writer, http.StatusInternalServerError,
			fmt.Sprintf("获取路由配置失败: %v", err), traceID)
		return
	}

	utils.WriteSuccessResponse(c.Writer, "路由配置获取成功", routes, traceID)
}

// getAllPages 获取所有页面配置（公开）
func (cb *CentralBrain) getAllPages(c *gin.Context) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	pages, err := cb.routerClient.GetAllPages()
	if err != nil {
		utils.WriteErrorResponse(c.Writer, http.StatusInternalServerError,
			fmt.Sprintf("获取页面配置失败: %v", err), traceID)
		return
	}

	utils.WriteSuccessResponse(c.Writer, "页面配置获取成功", pages, traceID)
}

// getUserRoutes 获取用户可访问的路由（需要认证）
func (cb *CentralBrain) getUserRoutes(c *gin.Context) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	// 提取用户token
	userToken := cb.extractUserToken(c.Request)
	if userToken == "" {
		utils.WriteErrorResponse(c.Writer, http.StatusUnauthorized,
			"未提供认证token", traceID)
		return
	}

	routes, err := cb.routerClient.GetUserRoutes(userToken)
	if err != nil {
		utils.WriteErrorResponse(c.Writer, http.StatusInternalServerError,
			fmt.Sprintf("获取用户路由失败: %v", err), traceID)
		return
	}

	utils.WriteSuccessResponse(c.Writer, "获取用户路由成功", routes, traceID)
}

// getUserPages 获取用户可访问的页面（需要认证）
func (cb *CentralBrain) getUserPages(c *gin.Context) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	// 提取用户token
	userToken := cb.extractUserToken(c.Request)
	if userToken == "" {
		utils.WriteErrorResponse(c.Writer, http.StatusUnauthorized,
			"未提供认证token", traceID)
		return
	}

	pages, err := cb.routerClient.GetUserPages(userToken)
	if err != nil {
		utils.WriteErrorResponse(c.Writer, http.StatusInternalServerError,
			fmt.Sprintf("获取用户页面失败: %v", err), traceID)
		return
	}

	utils.WriteSuccessResponse(c.Writer, "获取用户页面成功", pages, traceID)
}

// registerPermissionRoutes 注册Permission Service权限管理API
func (cb *CentralBrain) registerPermissionRoutes() {
	// 公开API：获取所有角色列表
	cb.router.GET("/api/v1/permission/roles", cb.getAllRoles)

	// 公开API：获取所有权限列表
	cb.router.GET("/api/v1/permission/permissions", cb.getAllPermissions)

	// 需要认证的API：获取用户角色
	cb.router.GET("/api/v1/permission/user/:userId/roles", cb.getUserRoles)

	// 需要认证的API：获取用户权限
	cb.router.GET("/api/v1/permission/user/:userId/permissions", cb.getUserPermissions)

	// 需要认证的API：获取角色权限
	cb.router.GET("/api/v1/permission/role/:roleId/permissions", cb.getRolePermissions)
}

// getAllRoles 获取所有角色列表（公开）
func (cb *CentralBrain) getAllRoles(c *gin.Context) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	roles, err := cb.permissionClient.GetAllRoles()
	if err != nil {
		utils.WriteErrorResponse(c.Writer, http.StatusInternalServerError,
			fmt.Sprintf("获取角色列表失败: %v", err), traceID)
		return
	}

	utils.WriteSuccessResponse(c.Writer, "角色列表获取成功", roles, traceID)
}

// getAllPermissions 获取所有权限列表（公开）
func (cb *CentralBrain) getAllPermissions(c *gin.Context) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	permissions, err := cb.permissionClient.GetAllPermissions()
	if err != nil {
		utils.WriteErrorResponse(c.Writer, http.StatusInternalServerError,
			fmt.Sprintf("获取权限列表失败: %v", err), traceID)
		return
	}

	utils.WriteSuccessResponse(c.Writer, "权限列表获取成功", permissions, traceID)
}

// getUserRoles 获取用户角色列表（需要认证）
func (cb *CentralBrain) getUserRoles(c *gin.Context) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	// 提取用户token
	userToken := cb.extractUserToken(c.Request)
	if userToken == "" {
		utils.WriteErrorResponse(c.Writer, http.StatusUnauthorized,
			"未提供认证token", traceID)
		return
	}

	// 从路径参数获取用户ID
	userID := c.Param("userId")
	if userID == "" {
		utils.WriteErrorResponse(c.Writer, http.StatusBadRequest,
			"用户ID不能为空", traceID)
		return
	}

	roles, err := cb.permissionClient.GetUserRoles(userID)
	if err != nil {
		utils.WriteErrorResponse(c.Writer, http.StatusInternalServerError,
			fmt.Sprintf("获取用户角色失败: %v", err), traceID)
		return
	}

	utils.WriteSuccessResponse(c.Writer, "获取用户角色成功", gin.H{
		"user_id": userID,
		"roles":   roles,
	}, traceID)
}

// getUserPermissions 获取用户权限列表（需要认证）
func (cb *CentralBrain) getUserPermissions(c *gin.Context) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	// 提取用户token
	userToken := cb.extractUserToken(c.Request)
	if userToken == "" {
		utils.WriteErrorResponse(c.Writer, http.StatusUnauthorized,
			"未提供认证token", traceID)
		return
	}

	// 从路径参数获取用户ID
	userID := c.Param("userId")
	if userID == "" {
		utils.WriteErrorResponse(c.Writer, http.StatusBadRequest,
			"用户ID不能为空", traceID)
		return
	}

	permissions, err := cb.permissionClient.GetUserPermissions(userID)
	if err != nil {
		utils.WriteErrorResponse(c.Writer, http.StatusInternalServerError,
			fmt.Sprintf("获取用户权限失败: %v", err), traceID)
		return
	}

	utils.WriteSuccessResponse(c.Writer, "获取用户权限成功", permissions, traceID)
}

// getRolePermissions 获取角色权限列表（需要认证）
func (cb *CentralBrain) getRolePermissions(c *gin.Context) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	// 提取用户token
	userToken := cb.extractUserToken(c.Request)
	if userToken == "" {
		utils.WriteErrorResponse(c.Writer, http.StatusUnauthorized,
			"未提供认证token", traceID)
		return
	}

	// 从路径参数获取角色ID
	roleID := c.Param("roleId")
	if roleID == "" {
		utils.WriteErrorResponse(c.Writer, http.StatusBadRequest,
			"角色ID不能为空", traceID)
		return
	}

	permissions, err := cb.permissionClient.GetRolePermissions(roleID)
	if err != nil {
		utils.WriteErrorResponse(c.Writer, http.StatusInternalServerError,
			fmt.Sprintf("获取角色权限失败: %v", err), traceID)
		return
	}

	utils.WriteSuccessResponse(c.Writer, "获取角色权限成功", gin.H{
		"role_id":     roleID,
		"permissions": permissions,
	}, traceID)
}

// initializeServiceTokenWithRetry 初始化服务token（带重试机制）
func (cb *CentralBrain) initializeServiceTokenWithRetry() {
	maxRetries := 5
	retryDelay := 3 * time.Second

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			fmt.Printf("🔄 重试获取服务token (%d/%d)...\n", i+1, maxRetries)
			time.Sleep(retryDelay)
			retryDelay *= 2 // 指数退避
		} else {
			// 第一次等待3秒让Auth Service启动
			time.Sleep(3 * time.Second)
		}

		// 检查Auth Service是否可用
		if !cb.checkAuthServiceHealth() {
			fmt.Printf("⚠️ Auth Service不可用，继续重试...\n")
			continue
		}

		// 获取服务token
		token, err := cb.requestServiceToken()
		if err != nil {
			fmt.Printf("⚠️ 获取服务token失败: %v\n", err)
			continue
		}

		// 保存token（带锁保护）
		cb.tokenMu.Lock()
		cb.serviceToken = token
		cb.serviceTokenExp = time.Now().Add(23 * time.Hour) // 提前1小时刷新
		cb.tokenMu.Unlock()

		fmt.Printf("✅ Central Brain服务token已获取\n")
		return
	}

	fmt.Printf("❌ 获取服务token失败，已达到最大重试次数。将在首次请求时重试\n")
}

// checkAuthServiceHealth 检查Auth Service健康状态
func (cb *CentralBrain) checkAuthServiceHealth() bool {
	healthURL := fmt.Sprintf("%s/health", cb.authServiceURL)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(healthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// requestServiceToken 请求服务token
func (cb *CentralBrain) requestServiceToken() (string, error) {
	// 从配置读取服务凭证（不再硬编码）
	serviceID := cb.config.ServiceCredentials.ServiceID
	serviceSecret := cb.config.ServiceCredentials.ServiceSecret

	// 验证服务凭证是否配置
	if serviceID == "" {
		return "", fmt.Errorf("SERVICE_ID未配置")
	}
	if serviceSecret == "" {
		return "", fmt.Errorf("SERVICE_SECRET未配置（必须从环境变量设置）")
	}

	// 调用Auth Service获取服务token
	url := fmt.Sprintf("%s/api/v1/auth/service/login", cb.authServiceURL)
	payload := fmt.Sprintf(`{"service_id":"%s","service_secret":"%s"}`, serviceID, serviceSecret)

	fmt.Printf("🔐 请求服务token: %s (ServiceID: %s)\n", url, serviceID)

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cb.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求Auth Service失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体（用于错误日志）
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("服务认证失败: HTTP %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ServiceToken string `json:"service_token"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v, 响应: %s", err, string(respBody))
	}

	if result.Code != 0 {
		return "", fmt.Errorf("服务认证失败: %s (Code: %d)", result.Message, result.Code)
	}

	if result.Data.ServiceToken == "" {
		return "", fmt.Errorf("服务认证失败: token为空")
	}

	fmt.Printf("✅ 服务token获取成功\n")
	return result.Data.ServiceToken, nil
}

// getServiceToken 获取服务token（从缓存或重新获取，线程安全）
func (cb *CentralBrain) getServiceToken() string {
	// 先尝试读取（读锁）
	cb.tokenMu.RLock()
	token := cb.serviceToken
	exp := cb.serviceTokenExp
	refreshInProgress := cb.tokenRefreshInProgress
	cb.tokenMu.RUnlock()

	// 如果token有效且未过期，直接返回
	if token != "" && time.Now().Before(exp) {
		return token
	}

	// 如果正在刷新，等待一下再读取
	if refreshInProgress {
		time.Sleep(500 * time.Millisecond)
		cb.tokenMu.RLock()
		token = cb.serviceToken
		exp = cb.serviceTokenExp
		cb.tokenMu.RUnlock()

		// 如果刷新完成且token有效，返回
		if token != "" && time.Now().Before(exp) {
			return token
		}
	}

	// 需要刷新token（写锁）
	cb.tokenMu.Lock()
	defer cb.tokenMu.Unlock()

	// 双重检查（防止并发刷新）
	if cb.serviceToken != "" && time.Now().Before(cb.serviceTokenExp) {
		return cb.serviceToken
	}

	// 标记正在刷新
	cb.tokenRefreshInProgress = true
	defer func() {
		cb.tokenRefreshInProgress = false
	}()

	// 重新获取token
	fmt.Printf("🔄 刷新服务token...\n")
	newToken, err := cb.requestServiceToken()
	if err != nil {
		fmt.Printf("⚠️ 重新获取服务token失败: %v\n", err)
		// 如果旧token已过期，返回空字符串（不使用过期token）
		if time.Now().After(cb.serviceTokenExp) {
			fmt.Printf("❌ 服务token已过期且无法刷新，请求可能失败\n")
			return ""
		}
		// 返回旧的token（如果未过期）
		return cb.serviceToken
	}

	// 更新token
	cb.serviceToken = newToken
	cb.serviceTokenExp = time.Now().Add(23 * time.Hour)
	fmt.Printf("✅ 服务token已刷新\n")

	return cb.serviceToken
}

// extractUserToken 提取用户token
func (cb *CentralBrain) extractUserToken(req *http.Request) string {
	// 从Authorization头提取
	authHeader := req.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 从accessToken头提取（兼容前端）
	accessToken := req.Header.Get("accessToken")
	if accessToken != "" {
		return accessToken
	}

	return ""
}
