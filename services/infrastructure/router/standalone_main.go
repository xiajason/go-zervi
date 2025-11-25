package main

import (
	"encoding/json"
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

	// 设置路由
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/version", versionHandler)
	http.HandleFunc("/info", infoHandler)
	http.HandleFunc("/api/v1/router/routes", routesHandler)
	http.HandleFunc("/api/v1/router/pages", pagesHandler)
	http.HandleFunc("/api/v1/router/user-routes", userRoutesHandler)
	http.HandleFunc("/api/v1/router/user-pages", userPagesHandler)

	// 启动服务
	log.Println("Starting Router Service on 0.0.0.0:8087")
	if err := http.ListenAndServe(":8087", nil); err != nil {
		log.Fatalf("Router Service启动失败: %v", err)
	}
}

// 健康检查
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service":   "router-service",
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "3.1.0",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 版本信息
func versionHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service": "router-service",
		"version": "3.1.0",
		"build":   time.Now().Format("2006-01-02 15:04:05"),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 服务信息
func infoHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service":    "router-service",
		"version":    "3.1.0",
		"port":       8087,
		"status":     "running",
		"started_at": time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 获取所有路由配置
func routesHandler(w http.ResponseWriter, r *http.Request) {
	routes := getAllRouteConfigs()
	response := map[string]interface{}{
		"code":    0,
		"message": "路由配置获取成功",
		"data":    routes,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 获取所有页面配置
func pagesHandler(w http.ResponseWriter, r *http.Request) {
	pages := getAllPageConfigs()
	response := map[string]interface{}{
		"code":    0,
		"message": "页面配置获取成功",
		"data":    pages,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 获取用户可访问的路由
func userRoutesHandler(w http.ResponseWriter, r *http.Request) {
	// 模拟用户ID
	userID := uint(1)
	routes := getUserAccessibleRoutes(userID)
	response := map[string]interface{}{
		"code":    0,
		"message": "获取用户路由成功",
		"data":    routes,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 获取用户可访问的页面
func userPagesHandler(w http.ResponseWriter, r *http.Request) {
	// 模拟用户ID
	userID := uint(1)
	pages := getUserAccessiblePages(userID)
	response := map[string]interface{}{
		"code":    0,
		"message": "获取用户页面成功",
		"data":    pages,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
			RequiredRoutes:      []string{"resume.detail"},
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
			if route.IsPublic || contains(route.Permissions, "resume:view") || contains(route.Permissions, "resume:create") {
				userRoutes = append(userRoutes, route)
			}
		}
		return userRoutes
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
			if hasPagePermission(page.RequiredPermissions, []string{"resume:view", "resume:create"}) {
				userPages = append(userPages, page)
			}
		}
		return userPages
	}
	
	// 默认返回空页面
	return []PageConfig{}
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
