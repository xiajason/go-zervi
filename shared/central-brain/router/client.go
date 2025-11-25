package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RouterClient Router Service客户端
type RouterClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewRouterClient 创建Router Service客户端
func NewRouterClient(baseURL string) *RouterClient {
	return &RouterClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
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

// RouterResponse Router Service响应格式
type RouterResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// GetAllRoutes 获取所有路由配置（公开）
func (c *RouterClient) GetAllRoutes() ([]RouteConfig, error) {
	url := fmt.Sprintf("%s/api/v1/router/routes", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求Router Service失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Router Service返回错误状态码: %d", resp.StatusCode)
	}

	var routerResp RouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&routerResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if routerResp.Code != 0 {
		return nil, fmt.Errorf("Router Service返回错误: %s", routerResp.Message)
	}

	// 转换数据
	routesJSON, _ := json.Marshal(routerResp.Data)
	var routes []RouteConfig
	if err := json.Unmarshal(routesJSON, &routes); err != nil {
		return nil, fmt.Errorf("解析路由配置失败: %v", err)
	}

	return routes, nil
}

// GetAllPages 获取所有页面配置（公开）
func (c *RouterClient) GetAllPages() ([]PageConfig, error) {
	url := fmt.Sprintf("%s/api/v1/router/pages", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求Router Service失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Router Service返回错误状态码: %d", resp.StatusCode)
	}

	var routerResp RouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&routerResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if routerResp.Code != 0 {
		return nil, fmt.Errorf("Router Service返回错误: %s", routerResp.Message)
	}

	// 转换数据
	pagesJSON, _ := json.Marshal(routerResp.Data)
	var pages []PageConfig
	if err := json.Unmarshal(pagesJSON, &pages); err != nil {
		return nil, fmt.Errorf("解析页面配置失败: %v", err)
	}

	return pages, nil
}

// GetUserRoutes 获取用户可访问的路由（需要认证）
func (c *RouterClient) GetUserRoutes(userToken string) ([]RouteConfig, error) {
	url := fmt.Sprintf("%s/api/v1/router/user-routes", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 添加认证头
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求Router Service失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Router Service返回错误状态码: %d", resp.StatusCode)
	}

	var routerResp RouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&routerResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if routerResp.Code != 0 {
		return nil, fmt.Errorf("Router Service返回错误: %s", routerResp.Message)
	}

	// 转换数据
	routesJSON, _ := json.Marshal(routerResp.Data)
	var routes []RouteConfig
	if err := json.Unmarshal(routesJSON, &routes); err != nil {
		return nil, fmt.Errorf("解析路由配置失败: %v", err)
	}

	return routes, nil
}

// GetUserPages 获取用户可访问的页面（需要认证）
func (c *RouterClient) GetUserPages(userToken string) ([]PageConfig, error) {
	url := fmt.Sprintf("%s/api/v1/router/user-pages", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 添加认证头
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求Router Service失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Router Service返回错误状态码: %d", resp.StatusCode)
	}

	var routerResp RouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&routerResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if routerResp.Code != 0 {
		return nil, fmt.Errorf("Router Service返回错误: %s", routerResp.Message)
	}

	// 转换数据
	pagesJSON, _ := json.Marshal(routerResp.Data)
	var pages []PageConfig
	if err := json.Unmarshal(pagesJSON, &pages); err != nil {
		return nil, fmt.Errorf("解析页面配置失败: %v", err)
	}

	return pages, nil
}
