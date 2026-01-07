package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	jobfirst "github.com/szjason72/zervigo/shared/core"
)

func main() {
	if len(os.Args) > 0 {
		os.Args[0] = "job-service"
	}

	core, err := jobfirst.NewCore("")
	if err != nil {
		log.Fatalf("初始化JobFirst核心包失败: %v", err)
	}
	defer core.Close()

	db := core.GetDB()
	if db == nil {
		log.Fatal("未能获取数据库连接，无法启动 job-service")
	}

	jobService := NewJobService(db)
	if err := jobService.EnsureSeedData(context.Background()); err != nil {
		log.Printf("WARN: 初始化职位种子数据失败: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/health", healthCheck)
	r.GET("/info", serviceInfo)

	handler := newJobHandler(jobService)
	handler.registerRoutes(r)

	registerToConsul("job-service", "127.0.0.1", 8084)

	log.Println("Starting Job Service with jobfirst-core on 0.0.0.0:8084")
	if err := r.Run(":8084"); err != nil {
		log.Fatalf("Job Service启动失败: %v", err)
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
		Tags: []string{"job", "microservice"},
		Meta: map[string]string{
			"version":     "3.1.0",
			"environment": "production",
			"port":        "8084",
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
		"service":   "job-service",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "3.1.0",
	})
}

// 服务信息端点
func serviceInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "job-service",
		"version":   "3.1.0",
		"port":      8084,
		"status":    "running",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
