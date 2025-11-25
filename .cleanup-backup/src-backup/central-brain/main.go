package main

import (
	"log"

	"github.com/jobfirst/jobfirst-core/centralbrain"
	"github.com/jobfirst/jobfirst-core/shared"
)

func main() {
	// 加载配置
	config := shared.GetDefaultConfig()

	// 创建中央大脑服务
	centralBrain := centralbrain.NewCentralBrain(config)

	// 启动服务
	port := config.CentralBrainPort
	log.Printf("🧠 Zervigo中央大脑启动在端口 %d", port)
	log.Printf("📊 服务路由:")
	log.Printf("  /api/v1/auth/**      → Auth Service (8207)")
	log.Printf("  /api/v1/ai/**        → AI Service (8100)")
	log.Printf("  /api/v1/blockchain/** → Blockchain Service (8208)")
	log.Printf("  /api/v1/user/**      → User Service (8082)")
	log.Printf("  /api/v1/job/**       → Job Service (8084)")
	log.Printf("  /api/v1/resume/**    → Resume Service (8085)")
	log.Printf("  /api/v1/company/**   → Company Service (8083)")
	log.Printf("  /health              → 健康检查")

	if err := centralBrain.Start(); err != nil {
		log.Fatalf("中央大脑启动失败: %v", err)
	}
}
