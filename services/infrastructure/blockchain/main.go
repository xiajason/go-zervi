package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) > 0 {
		os.Args[0] = "blockchain-service"
	}

	dbURL := resolveDatabaseURL()
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(16)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	jwtSecret := getEnvString("JWT_SECRET", "zervigo-local-dev-secret-key-2025")
	service := NewBlockchainService(db, jwtSecret)

	log.Println("正在初始化区块链数据库...")
	if err := service.InitializeDatabase(); err != nil {
		log.Fatalf("区块链数据库初始化失败: %v", err)
	}
	log.Println("区块链数据库初始化完成")

	port := getEnvInt("BLOCKCHAIN_SERVICE_PORT", 8208)
	api := NewBlockchainAPI(service, port)

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

func resolveDatabaseURL() string {
	if urlValue := os.Getenv("DATABASE_URL"); urlValue != "" {
		return urlValue
	}

	host := getEnvString("POSTGRESQL_HOST", getEnvString("POSTGRES_HOST", "localhost"))
	port := getEnvString("POSTGRESQL_PORT", getEnvString("POSTGRES_PORT", "5432"))
	user := getEnvString("POSTGRESQL_USER", getEnvString("POSTGRES_USER", os.Getenv("USER")))
	if user == "" {
		user = "postgres"
	}
	password := getEnvString("POSTGRESQL_PASSWORD", getEnvString("POSTGRES_PASSWORD", ""))
	database := getEnvString("POSTGRESQL_DATABASE", getEnvString("POSTGRES_DB", "zervigo_mvp"))
	sslMode := getEnvString("POSTGRESQL_SSL_MODE", "disable")

	if password != "" {
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user,
			url.QueryEscape(password),
			host,
			port,
			database,
			sslMode,
		)
	}

	return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s",
		user,
		host,
		port,
		database,
		sslMode,
	)
}

func getEnvString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}
