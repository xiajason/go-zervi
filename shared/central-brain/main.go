package main

import (
	"log"

	"github.com/szjason72/zervigo/shared/core/shared"
)

func main() {
	// 加载配置（从环境变量）
	config, err := shared.LoadConfig()
	if err != nil {
		log.Fatalf("❌ 配置加载失败: %v", err)
	}

	// 创建中央大脑服务
	centralBrain := NewCentralBrain(config)

	// 启动服务
	port := config.CentralBrainPort
	log.Printf("🧠 Zervigo中央大脑启动在端口 %d", port)
	log.Printf("📊 配置信息:")
	log.Printf("  服务主机: %s", config.ServiceDiscovery.ServiceHost)
	log.Printf("  服务发现: %v", config.ServiceDiscovery.Enabled)
	log.Printf("  Consul URL: %s", config.ServiceDiscovery.ConsulURL)
	log.Printf("📊 服务路由:")
	log.Printf("  /api/v1/auth/**      → Auth Service (%s:%d)",
		config.ServiceDiscovery.ServiceHost, config.AuthServicePort)
	log.Printf("  /api/v1/ai/**        → AI Service (%s:%d)",
		config.ServiceDiscovery.ServiceHost, config.AIServicePort)
	log.Printf("  /api/v1/blockchain/** → Blockchain Service (%s:%d)",
		config.ServiceDiscovery.ServiceHost, config.BlockchainServicePort)
	log.Printf("  /api/v1/users/**     → User Service (%s:%d)",
		config.ServiceDiscovery.ServiceHost, config.UserServicePort)
	log.Printf("  /api/v1/job/**       → Job Service (%s:%d)",
		config.ServiceDiscovery.ServiceHost, config.JobServicePort)
	log.Printf("  /api/v1/resume/**    → Resume Service (%s:%d)",
		config.ServiceDiscovery.ServiceHost, config.ResumeServicePort)
	log.Printf("  /api/v1/company/**   → Company Service (%s:%d)",
		config.ServiceDiscovery.ServiceHost, config.CompanyServicePort)
	log.Printf("  /api/v1/router/**    → Router Service (%s:%d)",
		config.ServiceDiscovery.ServiceHost, config.RouterServicePort)
	log.Printf("  /health              → 健康检查")
	log.Printf("  /api/v1/metrics      → 性能指标")
	log.Printf("  /api/v1/circuit-breakers → 熔断器状态")
	log.Printf("  /api/v1/router/routes → 路由配置（公开）")
	log.Printf("  /api/v1/router/pages  → 页面配置（公开）")
	log.Printf("  /api/v1/router/user-routes → 用户路由（需认证）")
	log.Printf("  /api/v1/router/user-pages  → 用户页面（需认证）")
	log.Printf("  /api/v1/permission/roles         → 角色列表（公开）")
	log.Printf("  /api/v1/permission/permissions    → 权限列表（公开）")
	log.Printf("  /api/v1/permission/user/:userId/roles → 用户角色（需认证）")
	log.Printf("  /api/v1/permission/user/:userId/permissions → 用户权限（需认证）")
	log.Printf("  /api/v1/permission/role/:roleId/permissions → 角色权限（需认证）")

	if err := centralBrain.Start(); err != nil {
		log.Fatalf("中央大脑启动失败: %v", err)
	}
}
