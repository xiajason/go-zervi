package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	jobfirst "github.com/szjason72/zervigo/shared/core"
	"github.com/szjason72/zervigo/shared/core/auth"
	"github.com/szjason72/zervigo/shared/core/response"
)

// 全局服务发现实例
var serviceDiscovery *ServiceDiscovery

func main() {
	// 设置进程名称
	if len(os.Args) > 0 {
		os.Args[0] = "router-service"
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

	// 初始化服务发现（Consul集成）
	serviceDiscovery = NewServiceDiscovery()
	// 启动自动刷新（每30秒检查一次服务状态）
	serviceDiscovery.StartAutoRefresh(30 * time.Second)

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由器
	r := gin.Default()

	// 设置标准路由
	setupStandardRoutes(r, core)

	// 设置业务路由
	setupBusinessRoutes(r, core, sqlDB, jwtSecret)

	// 注册到Consul
	registerToConsul("router-service", "127.0.0.1", 8087)

	// 启动服务
	log.Println("Starting Router Service with Go-Zervi Framework on 0.0.0.0:8087")
	log.Printf("🎯 Router Service已启用智能服务发现和自适应路由过滤")
	if err := r.Run(":8087"); err != nil {
		log.Fatalf("Router Service启动失败: %v", err)
	}
}

// setupStandardRoutes 设置标准路由
func setupStandardRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		health := core.Health()
		c.JSON(http.StatusOK, gin.H{
			"service":     "router-service",
			"status":      "healthy",
			"timestamp":   time.Now().Format(time.RFC3339),
			"version":     "3.1.0",
			"core_health": health,
		})
	})

	// 版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "router-service",
			"version": "3.1.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 服务信息
	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":    "router-service",
			"version":    "3.1.0",
			"port":       8087,
			"status":     "running",
			"started_at": time.Now().Format(time.RFC3339),
		})
	})
}

// setupBusinessRoutes 设置业务路由
func setupBusinessRoutes(r *gin.Engine, core *jobfirst.Core, sqlDB *sql.DB, jwtSecret string) {
	// 公开API路由（不需要认证）
	public := r.Group("/api/v1/router")
	{
		// 获取所有路由配置（公开）- 智能过滤版本
		public.GET("/routes", func(c *gin.Context) {
			routes := getAllRouteConfigs(sqlDB)
			standardSuccessResponse(c, routes, "路由配置获取成功")
		})

		// 获取所有页面配置（公开）- 智能过滤版本
		public.GET("/pages", func(c *gin.Context) {
			pages := getAllPageConfigs(sqlDB)
			standardSuccessResponse(c, pages, "页面配置获取成功")
		})
		
		// 获取当前服务组合信息
		public.GET("/service-combination", func(c *gin.Context) {
			if serviceDiscovery == nil {
				standardErrorResponse(c, http.StatusServiceUnavailable, "服务发现未初始化", "")
				return
			}
			
			combination := serviceDiscovery.GetServiceCombination()
			availableServices := serviceDiscovery.GetAvailableServices()
			
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"message": "服务组合信息获取成功",
				"data": gin.H{
					"combination": combination,
					"available_services": availableServices,
					"total_services": len(availableServices),
				},
			})
		})
	}

	// 需要认证的API路由
	zerviAuthAdapter := auth.NewZerviAuthAdapter(sqlDB, jwtSecret)
	authMiddleware := zerviAuthAdapter.RequireAuth()
	api := r.Group("/api/v1/router")
	api.Use(authMiddleware)
	{
		// 获取用户可访问的路由
		api.GET("/user-routes", func(c *gin.Context) {
			userID := c.GetUint("user_id")
			if userID == 0 {
				standardErrorResponse(c, http.StatusUnauthorized, "未登录", "")
				return
			}

			// 获取用户角色
			roles := getUserRoles(sqlDB, userID)
			if len(roles) == 0 {
				standardErrorResponse(c, http.StatusForbidden, "无权限", "")
				return
			}

			// 获取用户可访问的路由
			accessibleRoutes := getAccessibleRoutes(sqlDB, roles)
			standardSuccessResponse(c, accessibleRoutes, "获取用户路由成功")
		})

		// 获取用户可访问的页面
		api.GET("/user-pages", func(c *gin.Context) {
			userID := c.GetUint("user_id")
			if userID == 0 {
				standardErrorResponse(c, http.StatusUnauthorized, "未登录", "")
				return
			}

			// 获取用户角色
			roles := getUserRoles(sqlDB, userID)
			if len(roles) == 0 {
				standardErrorResponse(c, http.StatusForbidden, "无权限", "")
				return
			}

			// 获取用户可访问的页面
			accessiblePages := getAccessiblePages(sqlDB, roles)
			standardSuccessResponse(c, accessiblePages, "获取用户页面成功")
		})

		// 刷新路由缓存
		api.POST("/refresh", func(c *gin.Context) {
			// 这里可以实现缓存刷新逻辑
			standardSuccessResponse(c, "缓存刷新成功", "缓存刷新成功")
		})

		// 动态代理路由
		api.Any("/proxy/*path", func(c *gin.Context) {
			path := c.Param("path")
			method := c.Request.Method

			// 查找匹配的路由配置
			routeConfig := findRouteConfig(sqlDB, path, method)
			if routeConfig == nil {
				standardErrorResponse(c, http.StatusNotFound, "路由不存在", "")
				return
			}

			// 检查权限
			if !routeConfig.IsPublic {
				userID := c.GetUint("user_id")
				if userID == 0 {
					standardErrorResponse(c, http.StatusUnauthorized, "未登录", "")
					return
				}

				roles := getUserRoles(sqlDB, userID)
				if !hasRoutePermission(sqlDB, roles, routeConfig.RouteKey, routeConfig.Permissions) {
					standardErrorResponse(c, http.StatusForbidden, "无权限访问", "")
					return
				}
			}

			// 代理请求到目标服务
			proxyRequest(c, routeConfig)
		})
	}
}

// RouteConfig 路由配置
type RouteConfig struct {
	RouteKey        string   `json:"routeKey"`
	RouteName       string   `json:"routeName"`
	RoutePath       string   `json:"routePath"`
	ServiceName     string   `json:"serviceName"`
	ServiceEndpoint string   `json:"serviceEndpoint"`
	Method          string   `json:"method"`
	RouteType       string   `json:"routeType"`
	Description     string   `json:"description"`
	IsPublic        bool     `json:"isPublic"`
	IsActive        bool     `json:"isActive"`
	Permissions     []string `json:"permissions"`
}

// PageConfig 页面配置
type PageConfig struct {
	PageKey             string                 `json:"pageKey"`
	PageName            string                 `json:"pageName"`
	PagePath            string                 `json:"pagePath"`
	ComponentName       string                 `json:"componentName"`
	PageType            string                 `json:"pageType"`
	RequiredRoutes      []string               `json:"requiredRoutes"`
	RequiredPermissions []string               `json:"requiredPermissions"`
	PageConfig          map[string]interface{} `json:"pageConfig"`
	IsActive            bool                   `json:"isActive"`
}

// 业务逻辑函数
func getAllRouteConfigs(sqlDB *sql.DB) []RouteConfig {
	// 🎯 智能过滤：根据Consul发现的可用服务过滤路由
	var availableServices []string
	if serviceDiscovery != nil {
		availableServices = serviceDiscovery.GetAvailableServices()
		log.Printf("🔍 Router Service智能过滤：当前可用服务 %v", availableServices)
	}

	var query string
	var rows *sql.Rows
	var err error

	if len(availableServices) > 0 {
		// 只返回可用服务的路由（智能过滤）
		query = `
			SELECT rc.route_key, rc.route_name, rc.route_path, rc.service_name, 
			       rc.service_endpoint, rc.method, rc.route_type, rc.description, 
			       rc.is_public, rc.is_active,
			       COALESCE(
			           (SELECT array_agg(rp.permission_code) 
			            FROM route_permission rp 
			            WHERE rp.route_key = rc.route_key), 
			           ARRAY[]::text[]
			       ) as permissions
			FROM route_config rc
			WHERE rc.is_active = true
			AND rc.service_name = ANY($1)
			ORDER BY rc.route_type, rc.route_name
		`
		rows, err = sqlDB.Query(query, availableServices)
		log.Printf("✅ 智能过滤模式：只返回 %d 个服务的路由", len(availableServices))
	} else {
		// 如果服务发现不可用，返回所有路由（降级模式）
		query = `
			SELECT rc.route_key, rc.route_name, rc.route_path, rc.service_name, 
			       rc.service_endpoint, rc.method, rc.route_type, rc.description, 
			       rc.is_public, rc.is_active,
			       COALESCE(
			           (SELECT array_agg(rp.permission_code) 
			            FROM route_permission rp 
			            WHERE rp.route_key = rc.route_key), 
			           ARRAY[]::text[]
			       ) as permissions
			FROM route_config rc
			WHERE rc.is_active = true
			ORDER BY rc.route_type, rc.route_name
		`
		rows, err = sqlDB.Query(query)
		log.Printf("⚠️  降级模式：返回所有路由（服务发现不可用）")
	}

	if err != nil {
		log.Printf("查询路由配置失败: %v", err)
		return []RouteConfig{}
	}
	defer rows.Close()

	var routes []RouteConfig
	for rows.Next() {
		var route RouteConfig
		var permissions string

		err := rows.Scan(
			&route.RouteKey, &route.RouteName, &route.RoutePath,
			&route.ServiceName, &route.ServiceEndpoint, &route.Method,
			&route.RouteType, &route.Description, &route.IsPublic,
			&route.IsActive, &permissions,
		)
		if err != nil {
			log.Printf("扫描路由配置失败: %v", err)
			continue
		}

		// 解析权限数组
		if permissions != "" {
			json.Unmarshal([]byte(permissions), &route.Permissions)
		}

		routes = append(routes, route)
	}

	log.Printf("📊 返回 %d 条路由配置", len(routes))
	return routes
}

func getAllPageConfigs(sqlDB *sql.DB) []PageConfig {
	query := `
		SELECT page_key, page_name, page_path, component_name, page_type,
		       required_routes, required_permissions, page_config, is_active
		FROM frontend_page_config
		WHERE is_active = true
		ORDER BY page_type, page_name
	`

	rows, err := sqlDB.Query(query)
	if err != nil {
		log.Printf("查询页面配置失败: %v", err)
		return []PageConfig{}
	}
	defer rows.Close()

	var pages []PageConfig
	for rows.Next() {
		var page PageConfig
		var requiredRoutes, requiredPermissions, pageConfigJSON string

		err := rows.Scan(
			&page.PageKey, &page.PageName, &page.PagePath,
			&page.ComponentName, &page.PageType, &requiredRoutes,
			&requiredPermissions, &pageConfigJSON, &page.IsActive,
		)
		if err != nil {
			log.Printf("扫描页面配置失败: %v", err)
			continue
		}

		// 解析JSON字段
		json.Unmarshal([]byte(requiredRoutes), &page.RequiredRoutes)
		json.Unmarshal([]byte(requiredPermissions), &page.RequiredPermissions)
		json.Unmarshal([]byte(pageConfigJSON), &page.PageConfig)

		pages = append(pages, page)
	}

	return pages
}

func getUserRoles(sqlDB *sql.DB, userID uint) []string {
	query := `
		SELECT r.role_name
		FROM zervigo_auth_user_roles ur
		JOIN zervigo_auth_roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`

	rows, err := sqlDB.Query(query, userID)
	if err != nil {
		log.Printf("查询用户角色失败: %v", err)
		return []string{}
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err == nil {
			roles = append(roles, role)
		}
	}

	return roles
}

func getAccessibleRoutes(sqlDB *sql.DB, roles []string) []RouteConfig {
	query := `
		SELECT DISTINCT rc.route_key, rc.route_name, rc.route_path, rc.service_name, 
		       rc.service_endpoint, rc.method, rc.route_type, rc.description, 
		       rc.is_public, rc.is_active,
		       COALESCE(
		           (SELECT array_agg(rp.permission_code) 
		            FROM route_permission rp 
		            WHERE rp.route_key = rc.route_key), 
		           ARRAY[]::text[]
		       ) as permissions
		FROM route_config rc
		LEFT JOIN role_route_permission rrp ON rc.route_key = rrp.route_key
		LEFT JOIN zervigo_auth_roles r ON rrp.role_id = r.id
		WHERE rc.is_active = true
		AND (rc.is_public = true OR r.role_name = ANY($1))
		ORDER BY rc.route_type, rc.route_name
	`

	rows, err := sqlDB.Query(query, roles)
	if err != nil {
		log.Printf("查询可访问路由失败: %v", err)
		return []RouteConfig{}
	}
	defer rows.Close()

	var routes []RouteConfig
	for rows.Next() {
		var route RouteConfig
		var permissions string

		err := rows.Scan(
			&route.RouteKey, &route.RouteName, &route.RoutePath,
			&route.ServiceName, &route.ServiceEndpoint, &route.Method,
			&route.RouteType, &route.Description, &route.IsPublic,
			&route.IsActive, &permissions,
		)
		if err != nil {
			log.Printf("扫描可访问路由失败: %v", err)
			continue
		}

		// 解析权限数组
		if permissions != "" {
			json.Unmarshal([]byte(permissions), &route.Permissions)
		}

		routes = append(routes, route)
	}

	return routes
}

func getAccessiblePages(sqlDB *sql.DB, roles []string) []PageConfig {
	query := `
		SELECT page_key, page_name, page_path, component_name, page_type,
		       required_routes, required_permissions, page_config, is_active
		FROM frontend_page_config
		WHERE is_active = true
		ORDER BY page_type, page_name
	`

	rows, err := sqlDB.Query(query)
	if err != nil {
		log.Printf("查询可访问页面失败: %v", err)
		return []PageConfig{}
	}
	defer rows.Close()

	var pages []PageConfig
	for rows.Next() {
		var page PageConfig
		var requiredRoutes, requiredPermissions, pageConfigJSON string

		err := rows.Scan(
			&page.PageKey, &page.PageName, &page.PagePath,
			&page.ComponentName, &page.PageType, &requiredRoutes,
			&requiredPermissions, &pageConfigJSON, &page.IsActive,
		)
		if err != nil {
			log.Printf("扫描可访问页面失败: %v", err)
			continue
		}

		// 解析JSON字段
		json.Unmarshal([]byte(requiredRoutes), &page.RequiredRoutes)
		json.Unmarshal([]byte(requiredPermissions), &page.RequiredPermissions)
		json.Unmarshal([]byte(pageConfigJSON), &page.PageConfig)

		// 检查用户是否有访问该页面所需的所有权限
		if hasPagePermission(sqlDB, roles, page.RequiredPermissions) {
			pages = append(pages, page)
		}
	}

	return pages
}

func findRouteConfig(sqlDB *sql.DB, path, method string) *RouteConfig {
	query := `
		SELECT rc.route_key, rc.route_name, rc.route_path, rc.service_name, 
		       rc.service_endpoint, rc.method, rc.route_type, rc.description, 
		       rc.is_public, rc.is_active,
		       COALESCE(
		           (SELECT array_agg(rp.permission_code) 
		            FROM route_permission rp 
		            WHERE rp.route_key = rc.route_key), 
		           ARRAY[]::text[]
		       ) as permissions
		FROM route_config rc
		WHERE rc.is_active = true
		AND rc.method = $1
		AND ($2 LIKE rc.route_path || '%' OR rc.route_path = $2)
		ORDER BY LENGTH(rc.route_path) DESC
		LIMIT 1
	`

	row := sqlDB.QueryRow(query, method, path)

	var route RouteConfig
	var permissions string

	err := row.Scan(
		&route.RouteKey, &route.RouteName, &route.RoutePath,
		&route.ServiceName, &route.ServiceEndpoint, &route.Method,
		&route.RouteType, &route.Description, &route.IsPublic,
		&route.IsActive, &permissions,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Printf("查找路由配置失败: %v", err)
		return nil
	}

	// 解析权限数组
	if permissions != "" {
		json.Unmarshal([]byte(permissions), &route.Permissions)
	}

	return &route
}

func hasRoutePermission(sqlDB *sql.DB, roles []string, routeKey string, requiredPermissions []string) bool {
	// 检查角色是否有访问该路由的权限
	query := `
		SELECT COUNT(*) FROM role_route_permission rrp
		JOIN zervigo_auth_roles r ON rrp.role_id = r.id
		WHERE r.role_name = ANY($1) 
		AND rrp.route_key = $2 
		AND rrp.is_granted = true
	`

	var count int
	err := sqlDB.QueryRow(query, roles, routeKey).Scan(&count)
	if err != nil {
		log.Printf("检查路由权限失败: %v", err)
		return false
	}

	return count > 0
}

func hasPagePermission(sqlDB *sql.DB, roles []string, requiredPermissions []string) bool {
	if len(requiredPermissions) == 0 {
		return true
	}

	// 检查用户是否有访问该页面所需的所有权限
	for _, permission := range requiredPermissions {
		if !hasPermission(sqlDB, roles, permission) {
			return false
		}
	}

	return true
}

func hasPermission(sqlDB *sql.DB, roles []string, permission string) bool {
	query := `
		SELECT COUNT(*) FROM zervigo_auth_role_permissions rp
		JOIN zervigo_auth_roles r ON rp.role_id = r.id
		JOIN zervigo_auth_permissions p ON rp.permission_id = p.id
		WHERE r.role_name = ANY($1) AND p.permission_code = $2
	`

	var count int
	err := sqlDB.QueryRow(query, roles, permission).Scan(&count)
	if err != nil {
		log.Printf("检查权限失败: %v", err)
		return false
	}

	return count > 0
}

func proxyRequest(c *gin.Context, routeConfig *RouteConfig) {
	// 构建目标URL
	targetURL := fmt.Sprintf("http://%s%s", routeConfig.ServiceName, routeConfig.ServiceEndpoint)

	// 简化实现，直接返回配置信息
	// 在实际实现中，这里应该使用HTTP客户端库进行代理
	c.JSON(http.StatusOK, gin.H{
		"message":    "代理请求",
		"target":     targetURL,
		"routeKey":   routeConfig.RouteKey,
		"routeName":  routeConfig.RouteName,
		"method":     routeConfig.Method,
		"service":    routeConfig.ServiceName,
		"endpoint":   routeConfig.ServiceEndpoint,
		"timestamp":  time.Now().UnixMilli(),
	})
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
		Tags:    []string{"router", "dynamic", "proxy", "rbac", "jobfirst", "microservice", "version:3.1.0"},
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
