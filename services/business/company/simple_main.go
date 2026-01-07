package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
)

func main() {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.Default()

	// 添加健康检查端点
	r.GET("/health", healthCheck)
	r.GET("/info", serviceInfo)

	// 注册到Consul
	registerToConsul("company-service", "127.0.0.1", 8083)

	// 启动服务
	log.Println("Starting Company Service on 0.0.0.0:8083")
	if err := r.Run(":8083"); err != nil {
		log.Fatalf("Company Service启动失败: %v", err)
	}
}

func registerToConsul(serviceName, serviceHost string, servicePort int) {
	// 创建Consul客户端
	config := api.DefaultConfig()
	config.Address = "localhost:8500"
	client, err := api.NewClient(config)
	if err != nil {
		log.Printf("创建Consul客户端失败: %v", err)
		return
	}

	// 服务注册信息
	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%d", serviceName, servicePort),
		Name:    serviceName,
		Address: serviceHost,
		Port:    servicePort,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", serviceHost, servicePort),
			Interval:                       "10s",
			Timeout:                        "3s",
			DeregisterCriticalServiceAfter: "30s",
		},
		Tags: []string{"company", "microservice"},
		Meta: map[string]string{
			"version":     "3.1.0",
			"environment": "production",
			"port":        "8083",
		},
	}

	// 注册服务
	if err := client.Agent().ServiceRegister(registration); err != nil {
		log.Printf("注册服务到Consul失败: %v", err)
	} else {
		log.Printf("服务 %s 已注册到Consul: %s:%d", serviceName, serviceHost, servicePort)
	}
}

// 健康检查端点
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "company-service",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "3.1.0",
	})
}

// 服务信息端点
func serviceInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "company-service",
		"version":   "3.1.0",
		"port":      8083,
		"status":    "running",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
