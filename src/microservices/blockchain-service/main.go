package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jobfirst/jobfirst-core/blockchain"
	"github.com/jobfirst/jobfirst-core/shared"
)

func main() {
	// 加载配置
	config := shared.NewConfig()

	// 配置数据库连接
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "root:@tcp(localhost:3306)/zervigo_blockchain?charset=utf8mb4&parseTime=True&loc=Local"
	}

	// 连接数据库
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	// 创建区块链服务
	blockchainService := blockchain.NewBlockchainService(db, config.JWTSecret)

	// 初始化数据库
	log.Println("正在初始化区块链数据库...")
	if err := blockchainService.InitializeDatabase(); err != nil {
		log.Fatalf("区块链数据库初始化失败: %v", err)
	}
	log.Println("区块链数据库初始化完成")

	// 创建API服务器
	port := config.BlockchainServicePort
	if portEnv := os.Getenv("BLOCKCHAIN_SERVICE_PORT"); portEnv != "" {
		fmt.Sscanf(portEnv, "%d", &port)
	}

	api := blockchain.NewBlockchainAPI(blockchainService, port)

	// 启动服务器
	log.Printf("区块链微服务启动在端口 %d", port)
	log.Println("支持的API端点:")
	log.Println("  POST /api/v1/blockchain/version/status/record - 记录版本状态变化")
	log.Println("  GET  /api/v1/blockchain/version/status/history/{userId} - 查询用户状态历史")
	log.Println("  POST /api/v1/blockchain/permission/change/record - 记录权限变更")
	log.Println("  GET  /api/v1/blockchain/permission/change/history/{userId} - 查询权限变更历史")
	log.Println("  POST /api/v1/blockchain/consistency/validate - 数据一致性校验")
	log.Println("  GET  /api/v1/blockchain/transaction/list - 查询区块链交易列表")
	log.Println("  GET  /health - 健康检查")

	if err := api.Start(); err != nil {
		log.Fatalf("区块链服务启动失败: %v", err)
	}
}
