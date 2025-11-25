package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/szjason72/zervigo/shared/core/response"
)

func main() {
	// 设置进程名称
	if len(os.Args) > 0 {
		os.Args[0] = "resume-service"
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由器
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "resume-service",
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "3.1.0",
		})
	})

	// 版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "resume-service",
			"version": "3.1.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 服务信息
	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":    "resume-service",
			"version":    "3.1.0",
			"port":       8085,
			"status":     "running",
			"started_at": time.Now().Format(time.RFC3339),
		})
	})

	// 公开API路由（不需要认证）
	public := r.Group("/api/v1")
	{
		// 简历模板列表
		public.GET("/resume/templates", func(c *gin.Context) {
			templates := []gin.H{
				{"templateId": "template_001", "templateName": "经典模板", "previewUrl": "https://example.com/preview1.jpg"},
				{"templateId": "template_002", "templateName": "现代模板", "previewUrl": "https://example.com/preview2.jpg"},
				{"templateId": "template_003", "templateName": "创意模板", "previewUrl": "https://example.com/preview3.jpg"},
			}
			standardSuccessResponse(c, templates, "简历模板列表获取成功")
		})

		// 简历权限配置API（测试用）
		public.GET("/resume/permission/:resumeId", func(c *gin.Context) {
			resumeID := c.Param("resumeId")

			// 模拟数据
			config := gin.H{
				"resumeId":           resumeID,
				"privacyLevel":       "PRIVATE",
				"allowDownload":      false,
				"requireApproval":    true,
				"allowedEnterprises": []string{"ent_001", "ent_002"},
				"deniedEnterprises":  []string{"ent_003"},
				"updateTime":         time.Now().UnixMilli(),
			}

			standardSuccessResponse(c, config, "简历权限配置获取成功")
		})

		// 简历黑名单API（测试用）
		public.GET("/resume/blacklist/:resumeId", func(c *gin.Context) {
			_ = c.Param("resumeId") // 使用下划线忽略未使用的变量

			// 模拟数据
			blacklist := []gin.H{
				{
					"enterpriseId":   "ent_003",
					"enterpriseName": "不良企业A",
					"reason":         "恶意下载简历",
					"addTime":        time.Now().UnixMilli(),
				},
				{
					"enterpriseId":   "ent_004",
					"enterpriseName": "不良企业B",
					"reason":         "骚扰用户",
					"addTime":        time.Now().UnixMilli(),
				},
			}

			standardSuccessResponse(c, blacklist, "简历黑名单获取成功")
		})

		// 审批列表API（测试用）
		public.GET("/approve/list", func(c *gin.Context) {
			// 模拟数据
			approves := []gin.H{
				{
					"approveId":      "approve_001",
					"type":           "简历查看",
					"enterpriseName": "优质企业A",
					"resumeName":     "张三的简历",
					"status":         "待审批",
					"cost":           10,
					"createTime":     time.Now().UnixMilli(),
				},
				{
					"approveId":      "approve_002",
					"type":           "简历下载",
					"enterpriseName": "优质企业B",
					"resumeName":     "李四的简历",
					"status":         "已通过",
					"cost":           20,
					"createTime":     time.Now().UnixMilli(),
				},
			}

			result := gin.H{
				"list":         approves,
				"total":        2,
				"pageNum":      1,
				"pageSize":     10,
				"pendingCount": 1,
			}

			standardSuccessResponse(c, result, "审批列表获取成功")
		})

		// 积分查询API（测试用）
		public.GET("/points/user/:userId", func(c *gin.Context) {
			_ = c.Param("userId") // 使用下划线忽略未使用的变量

			// 模拟数据
			points := []gin.H{
				{
					"type":       "收入",
					"amount":     100,
					"reason":     "注册奖励",
					"balance":    100,
					"createTime": time.Now().UnixMilli(),
				},
				{
					"type":       "支出",
					"amount":     10,
					"reason":     "查看简历",
					"balance":    90,
					"createTime": time.Now().UnixMilli(),
				},
			}

			standardSuccessResponse(c, points, "用户积分获取成功")
		})
	}

	// 注册到Consul
	registerToConsul("resume-service", "127.0.0.1", 8085)

	// 启动服务
	log.Println("Starting Resume Service with Go-Zervi Framework on 0.0.0.0:8085")
	if err := r.Run(":8085"); err != nil {
		log.Fatalf("Resume Service启动失败: %v", err)
	}
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
		Tags:    []string{"resume", "privacy", "permission", "jobfirst", "microservice", "version:3.1.0"},
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
