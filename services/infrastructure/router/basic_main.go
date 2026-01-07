package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// 设置进程名称
	if len(os.Args) > 0 {
		os.Args[0] = "router-service"
	}

	// 创建HTTP路由器
	mux := http.NewServeMux()

	// 设置路由
	setupRoutes(mux)

	// 启动服务
	log.Println("Starting Router Service with Go-Zervi Framework on 0.0.0.0:8087")
	if err := http.ListenAndServe(":8087", mux); err != nil {
		log.Fatalf("Router Service启动失败: %v", err)
	}
}

// setupRoutes 设置路由
func setupRoutes(mux *http.ServeMux) {
	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"service":   "router-service",
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "3.1.0",
		}
		json.NewEncoder(w).Encode(response)
	})

	// 版本信息
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"service": "router-service",
			"version": "3.1.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		}
		json.NewEncoder(w).Encode(response)
	})

	// 服务信息
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"service":    "router-service",
			"version":    "3.1.0",
			"port":       8087,
			"status":     "running",
			"started_at": time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	})

	// 路由配置API
	mux.HandleFunc("/api/v1/router/routes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		routes := getAllRouteConfigs()
		response := map[string]interface{}{
			"code":      0,
			"message":   "路由配置获取成功",
			"data":      routes,
			"timestamp": time.Now().UnixMilli(),
		}
		json.NewEncoder(w).Encode(response)
	})

	// 页面配置API
	mux.HandleFunc("/api/v1/router/pages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		pages := getAllPageConfigs()
		response := map[string]interface{}{
			"code":      0,
			"message":   "页面配置获取成功",
			"data":      pages,
			"timestamp": time.Now().UnixMilli(),
		}
		json.NewEncoder(w).Encode(response)
	})

	// 用户路由API
	mux.HandleFunc("/api/v1/router/user-routes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// 模拟用户ID
		userID := uint(1)
		routes := getUserAccessibleRoutes(userID)
		response := map[string]interface{}{
			"code":      0,
			"message":   "获取用户路由成功",
			"data":      routes,
			"timestamp": time.Now().UnixMilli(),
		}
		json.NewEncoder(w).Encode(response)
	})

	// 用户页面API
	mux.HandleFunc("/api/v1/router/user-pages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// 模拟用户ID
		userID := uint(1)
		pages := getUserAccessiblePages(userID)
		response := map[string]interface{}{
			"code":      0,
			"message":   "获取用户页面成功",
			"data":      pages,
			"timestamp": time.Now().UnixMilli(),
		}
		json.NewEncoder(w).Encode(response)
	})

	// 动态代理路由
	mux.HandleFunc("/api/v1/router/proxy/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		path := r.URL.Path[len("/api/v1/router/proxy"):]
		method := r.Method
		
		// 查找匹配的路由配置
		routeConfig := findRouteConfig(path, method)
		if routeConfig == nil {
			response := map[string]interface{}{
				"code":      10004,
				"message":   "路由不存在",
				"timestamp": time.Now().UnixMilli(),
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// 代理请求到目标服务
		targetURL := fmt.Sprintf("http://%s%s", routeConfig.ServiceName, routeConfig.ServiceEndpoint)
		response := map[string]interface{}{
			"code":      0,
			"message":   "代理请求",
			"data": map[string]interface{}{
				"target":     targetURL,
				"routeKey":   routeConfig.RouteKey,
				"routeName":  routeConfig.RouteName,
				"method":     routeConfig.Method,
				"service":    routeConfig.ServiceName,
				"endpoint":   routeConfig.ServiceEndpoint,
			},
			"timestamp": time.Now().UnixMilli(),
		}
		json.NewEncoder(w).Encode(response)
	})
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
func getAllRouteConfigs() []RouteConfig {
	// 模拟数据
	routes := []RouteConfig{
		{
			RouteKey:        "resume.list",
			RouteName:       "简历列表",
			RoutePath:       "/api/v1/resume/list",
			ServiceName:     "resume-service",
			ServiceEndpoint: "/api/v1/resume/list",
			Method:          "GET",
			RouteType:       "api",
			Description:     "获取简历列表",
			IsPublic:        false,
			IsActive:        true,
			Permissions:     []string{"resume:view"},
		},
		{
			RouteKey:        "resume.detail",
			RouteName:       "简历详情",
			RoutePath:       "/api/v1/resume/detail/*",
			ServiceName:     "resume-service",
			ServiceEndpoint: "/api/v1/resume/detail",
			Method:          "GET",
			RouteType:       "api",
			Description:     "获取简历详情",
			IsPublic:        false,
			IsActive:        true,
			Permissions:     []string{"resume:view"},
		},
		{
			RouteKey:        "resume.create",
			RouteName:       "创建简历",
			RoutePath:       "/api/v1/resume/create",
			ServiceName:     "resume-service",
			ServiceEndpoint: "/api/v1/resume/create",
			Method:          "POST",
			RouteType:       "api",
			Description:     "创建新简历",
			IsPublic:        false,
			IsActive:        true,
			Permissions:     []string{"resume:create"},
		},
		{
			RouteKey:        "resume.permission",
			RouteName:       "简历权限",
			RoutePath:       "/api/v1/resume/permission/*",
			ServiceName:     "resume-service",
			ServiceEndpoint: "/api/v1/resume/permission",
			Method:          "GET",
			RouteType:       "api",
			Description:     "获取简历权限配置",
			IsPublic:        false,
			IsActive:        true,
			Permissions:     []string{"resume:permission"},
		},
		{
			RouteKey:        "approve.list",
			RouteName:       "审批列表",
			RoutePath:       "/api/v1/approve/list",
			ServiceName:     "resume-service",
			ServiceEndpoint: "/api/v1/approve/list",
			Method:          "GET",
			RouteType:       "api",
			Description:     "获取审批列表",
			IsPublic:        false,
			IsActive:        true,
			Permissions:     []string{"approve:view"},
		},
		{
			RouteKey:        "points.user",
			RouteName:       "用户积分",
			RoutePath:       "/api/v1/points/user/*",
			ServiceName:     "resume-service",
			ServiceEndpoint: "/api/v1/points/user",
			Method:          "GET",
			RouteType:       "api",
			Description:     "获取用户积分",
			IsPublic:        false,
			IsActive:        true,
			Permissions:     []string{"points:view"},
		},
		{
			RouteKey:        "roles.list",
			RouteName:       "角色列表",
			RoutePath:       "/api/v1/roles",
			ServiceName:     "permission-service",
			ServiceEndpoint: "/api/v1/roles",
			Method:          "GET",
			RouteType:       "api",
			Description:     "获取角色列表",
			IsPublic:        true,
			IsActive:        true,
			Permissions:     []string{},
		},
		{
			RouteKey:        "permissions.list",
			RouteName:       "权限列表",
			RoutePath:       "/api/v1/permissions",
			ServiceName:     "permission-service",
			ServiceEndpoint: "/api/v1/permissions",
			Method:          "GET",
			RouteType:       "api",
			Description:     "获取权限列表",
			IsPublic:        true,
			IsActive:        true,
			Permissions:     []string{},
		},
	}

	return routes
}

func getAllPageConfigs() []PageConfig {
	// 模拟数据
	pages := []PageConfig{
		{
			PageKey:             "resume.list.page",
			PageName:            "简历列表页",
			PagePath:            "/pages/resume/index",
			ComponentName:       "ResumeList",
			PageType:            "page",
			RequiredRoutes:      []string{"resume.list"},
			RequiredPermissions: []string{"resume:view"},
			PageConfig: map[string]interface{}{
				"title":      "简历列表",
				"showCreate": true,
			},
			IsActive: true,
		},
		{
			PageKey:             "resume.detail.page",
			PageName:            "简历详情页",
			PagePath:            "/pages/resume/detail",
			ComponentName:       "ResumeDetail",
			PageType:            "page",
			RequiredRoutes:      []string{"resume.detail", "resume.permission"},
			RequiredPermissions: []string{"resume:view"},
			PageConfig: map[string]interface{}{
				"title":    "简历详情",
				"showEdit": true,
			},
			IsActive: true,
		},
		{
			PageKey:             "resume.create.page",
			PageName:            "创建简历页",
			PagePath:            "/pages/resume/create",
			ComponentName:       "ResumeCreate",
			PageType:            "page",
			RequiredRoutes:      []string{"resume.create"},
			RequiredPermissions: []string{"resume:create"},
			PageConfig: map[string]interface{}{
				"title":       "创建简历",
				"showPreview": true,
			},
			IsActive: true,
		},
		{
			PageKey:             "approve.list.page",
			PageName:            "审批列表页",
			PagePath:            "/pages/approve/index",
			ComponentName:       "ApproveList",
			PageType:            "page",
			RequiredRoutes:      []string{"approve.list"},
			RequiredPermissions: []string{"approve:view"},
			PageConfig: map[string]interface{}{
				"title":      "审批列表",
				"showHandle": true,
			},
			IsActive: true,
		},
		{
			PageKey:             "points.user.page",
			PageName:            "用户积分页",
			PagePath:            "/pages/points/index",
			ComponentName:       "PointsUser",
			PageType:            "page",
			RequiredRoutes:      []string{"points.user"},
			RequiredPermissions: []string{"points:view"},
			PageConfig: map[string]interface{}{
				"title":       "我的积分",
				"showHistory": true,
			},
			IsActive: true,
		},
		{
			PageKey:             "permission.manage.page",
			PageName:            "权限管理页",
			PagePath:            "/pages/permission/index",
			ComponentName:       "PermissionManage",
			PageType:            "page",
			RequiredRoutes:      []string{"roles.list", "permissions.list"},
			RequiredPermissions: []string{"user:roles", "role:permissions"},
			PageConfig: map[string]interface{}{
				"title":           "权限管理",
				"showRoles":       true,
				"showPermissions": true,
			},
			IsActive: true,
		},
	}

	return pages
}

func getUserAccessibleRoutes(userID uint) []RouteConfig {
	// 模拟数据 - 根据用户ID返回不同路由
	allRoutes := getAllRouteConfigs()
	
	// 模拟不同用户的权限
	if userID == 1 { // admin用户
		return allRoutes
	} else if userID == 2 { // 普通用户
		// 只返回基本权限的路由
		var userRoutes []RouteConfig
		for _, route := range allRoutes {
			if route.IsPublic || contains(route.Permissions, "resume:view") || contains(route.Permissions, "resume:create") || contains(route.Permissions, "points:view") {
				userRoutes = append(userRoutes, route)
			}
		}
		return userRoutes
	} else if userID == 3 { // 企业用户
		// 只返回企业相关权限的路由
		var enterpriseRoutes []RouteConfig
		for _, route := range allRoutes {
			if route.IsPublic || contains(route.Permissions, "resume:view") || contains(route.Permissions, "approve:view") {
				enterpriseRoutes = append(enterpriseRoutes, route)
			}
		}
		return enterpriseRoutes
	}
	
	// 默认返回公开路由
	var publicRoutes []RouteConfig
	for _, route := range allRoutes {
		if route.IsPublic {
			publicRoutes = append(publicRoutes, route)
		}
	}
	return publicRoutes
}

func getUserAccessiblePages(userID uint) []PageConfig {
	// 模拟数据 - 根据用户ID返回不同页面
	allPages := getAllPageConfigs()
	
	// 模拟不同用户的权限
	if userID == 1 { // admin用户
		return allPages
	} else if userID == 2 { // 普通用户
		// 只返回基本权限的页面
		var userPages []PageConfig
		for _, page := range allPages {
			if hasPagePermission(page.RequiredPermissions, []string{"resume:view", "resume:create", "points:view"}) {
				userPages = append(userPages, page)
			}
		}
		return userPages
	} else if userID == 3 { // 企业用户
		// 只返回企业相关权限的页面
		var enterprisePages []PageConfig
		for _, page := range allPages {
			if hasPagePermission(page.RequiredPermissions, []string{"resume:view", "approve:view"}) {
				enterprisePages = append(enterprisePages, page)
			}
		}
		return enterprisePages
	}
	
	// 默认返回空页面
	return []PageConfig{}
}

func findRouteConfig(path, method string) *RouteConfig {
	routes := getAllRouteConfigs()
	
	for _, route := range routes {
		if route.Method == method && matchPath(route.RoutePath, path) {
			return &route
		}
	}
	
	return nil
}

func matchPath(pattern, path string) bool {
	// 简单的路径匹配
	return len(path) >= len(pattern) && path[:len(pattern)] == pattern
}

func hasPagePermission(requiredPermissions []string, userPermissions []string) bool {
	if len(requiredPermissions) == 0 {
		return true
	}
	
	for _, required := range requiredPermissions {
		if !contains(userPermissions, required) {
			return false
		}
	}
	
	return true
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
