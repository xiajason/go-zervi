package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/joho/godotenv"

	"github.com/xiajason/zervi-basic/basic/backend/internal/api-gateway/middleware"
	"github.com/xiajason/zervi-basic/basic/backend/internal/api-gateway/router"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/logger"
)

const (
	ServiceName = "api-gateway"
	Version     = "1.0.0"
	Port        = 8601
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// 初始化日志
	log := logger.NewLogger("info")

	log.Infof("Starting JobFirst Professional API Gateway - Version: %s", Version)

	// 初始化Consul客户端
	consulClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to create Consul client - Error: %v", err)
	}

	// 创建API Gateway实例
	gateway := &APIGateway{
		consulClient: consulClient,
		logger:       log,
		port:         Port,
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.New()

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.Default())

	// 添加自定义中间件
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))

	// 注册路由
	router.RegisterRoutes(r, gateway)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", Port),
		Handler: r,
	}

	// 注册服务到Consul
	if err := gateway.registerService(); err != nil {
		log.Fatalf("Failed to register service with Consul - Error: %v", err)
	}

	// 启动服务器
	go func() {
		log.Infof("API Gateway starting - Port: %d", Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start API Gateway - Error: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down API Gateway...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 从Consul注销服务
	if err := gateway.deregisterService(); err != nil {
		log.Errorf("Failed to deregister service from Consul - Error: %v", err)
	}

	// 关闭HTTP服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("API Gateway forced to shutdown - Error: %v", err)
	}

	log.Info("API Gateway exited")
}

// APIGateway 表示API Gateway服务
type APIGateway struct {
	consulClient *api.Client
	logger       logger.Logger
	port         int
	serviceID    string
}

// registerService 注册服务到Consul
func (g *APIGateway) registerService() error {
	hostname, _ := os.Hostname()
	g.serviceID = fmt.Sprintf("%s-%s-%d", ServiceName, hostname, g.port)

	serviceDef := &api.AgentServiceRegistration{
		ID:      g.serviceID,
		Name:    ServiceName,
		Port:    g.port,
		Address: "127.0.0.1",
		Tags:    []string{"api-gateway", "professional", "v1"},
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://127.0.0.1:%d/health", g.port),
			Interval:                       "10s",
			Timeout:                        "3s",
			DeregisterCriticalServiceAfter: "30s",
		},
		Meta: map[string]string{
			"version": Version,
			"type":    "api-gateway",
		},
	}

	return g.consulClient.Agent().ServiceRegister(serviceDef)
}

// deregisterService 从Consul注销服务
func (g *APIGateway) deregisterService() error {
	return g.consulClient.Agent().ServiceDeregister(g.serviceID)
}

// GetConsulClient 获取Consul客户端
func (g *APIGateway) GetConsulClient() *api.Client {
	return g.consulClient
}

// GetLogger 获取日志记录器
func (g *APIGateway) GetLogger() logger.Logger {
	return g.logger
}

// GetPort 获取端口
func (g *APIGateway) GetPort() int {
	return g.port
}

// GetServiceID 获取服务ID
func (g *APIGateway) GetServiceID() string {
	return g.serviceID
}
