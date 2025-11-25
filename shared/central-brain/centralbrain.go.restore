package centralbrain

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core/shared"
)

// CentralBrain 中央大脑服务
type CentralBrain struct {
	config     *shared.Config
	httpClient *http.Client
	router     *gin.Engine
}

// ServiceProxy 服务代理配置
type ServiceProxy struct {
	ServiceName string
	BaseURL     string
	PathPrefix  string
}

// NewCentralBrain 创建中央大脑服务
func NewCentralBrain(config *shared.Config) *CentralBrain {
	return &CentralBrain{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		router: gin.Default(),
	}
}

// Start 启动中央大脑服务
func (cb *CentralBrain) Start() error {
	// 配置CORS
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

	// 注册服务代理
	cb.registerServiceProxies()

	// 健康检查
	cb.router.GET("/health", cb.healthCheck)

	return cb.router.Run(fmt.Sprintf(":%d", cb.config.CentralBrainPort))
}

// registerServiceProxies 注册服务代理
func (cb *CentralBrain) registerServiceProxies() {
	// 定义服务映射
	services := map[string]ServiceProxy{
		"auth": {
			ServiceName: "auth-service",
			BaseURL:     fmt.Sprintf("http://localhost:%d", cb.config.AuthServicePort),
			PathPrefix:  "/api/v1/auth",
		},
		"ai": {
			ServiceName: "ai-service",
			BaseURL:     fmt.Sprintf("http://localhost:%d", cb.config.AIServicePort),
			PathPrefix:  "/api/v1/ai",
		},
		"blockchain": {
			ServiceName: "blockchain-service",
			BaseURL:     fmt.Sprintf("http://localhost:%d", cb.config.BlockchainServicePort),
			PathPrefix:  "/api/v1/blockchain",
		},
		"user": {
			ServiceName: "user-service",
			BaseURL:     fmt.Sprintf("http://localhost:%d", cb.config.UserServicePort),
			PathPrefix:  "/api/v1/user",
		},
		"job": {
			ServiceName: "job-service",
			BaseURL:     fmt.Sprintf("http://localhost:%d", cb.config.JobServicePort),
			PathPrefix:  "/api/v1/job",
		},
		"resume": {
			ServiceName: "resume-service",
			BaseURL:     fmt.Sprintf("http://localhost:%d", cb.config.ResumeServicePort),
			PathPrefix:  "/api/v1/resume",
		},
		"company": {
			ServiceName: "company-service",
			BaseURL:     fmt.Sprintf("http://localhost:%d", cb.config.CompanyServicePort),
			PathPrefix:  "/api/v1/company",
		},
	}

	// 为每个服务注册代理路由
	for serviceKey, service := range services {
		cb.registerServiceProxy(serviceKey, service)
	}
}

// registerServiceProxy 注册单个服务代理
func (cb *CentralBrain) registerServiceProxy(serviceKey string, service ServiceProxy) {
	// 创建服务代理组
	proxyGroup := cb.router.Group(service.PathPrefix)

	// 注册通配符路由
	proxyGroup.Any("/*path", func(c *gin.Context) {
		cb.proxyRequest(c, service)
	})

	fmt.Printf("✅ 注册服务代理: %s -> %s\n", service.PathPrefix, service.BaseURL)
}

// proxyRequest 代理请求
func (cb *CentralBrain) proxyRequest(c *gin.Context, service ServiceProxy) {
	// 1. 提取路径
	originalPath := c.Request.URL.Path
	targetPath := strings.TrimPrefix(originalPath, service.PathPrefix)
	if targetPath == "" {
		targetPath = "/"
	}

	// 2. 构建目标URL
	targetURL := service.BaseURL + targetPath
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

	// 5. 复制请求头
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 6. 发送请求
	resp, err := cb.httpClient.Do(req)
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

	errorResponse := map[string]interface{}{
		"code":      500,
		"message":   fmt.Sprintf("中央大脑代理失败: %v", err),
		"data":      nil,
		"timestamp": time.Now().Unix(),
	}

	c.JSON(http.StatusInternalServerError, errorResponse)
}

// healthCheck 健康检查
func (cb *CentralBrain) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "中央大脑服务健康",
		"data": gin.H{
			"service":   "central-brain",
			"status":    "UP",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
		},
	})
}
