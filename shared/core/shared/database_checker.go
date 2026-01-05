package shared

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// DatabaseChecker 数据库检查器
type DatabaseChecker struct {
	config *Config
	logger *log.Logger
}

// DatabaseCheckResult 数据库检查结果
type DatabaseCheckResult struct {
	Type     string                 `json:"type"`
	Status   string                 `json:"status"` // "connected", "failed", "not_configured", "invalid_config"
	Message  string                 `json:"message"`
	Config   map[string]interface{} `json:"config"`
	Errors   []string               `json:"errors"`
	Warnings []string               `json:"warnings"`
	Duration time.Duration          `json:"duration"`
}

// NewDatabaseChecker 创建数据库检查器
func NewDatabaseChecker(config *Config) *DatabaseChecker {
	return &DatabaseChecker{
		config: config,
		logger: log.New(os.Stdout, "[DB-CHECK] ", log.LstdFlags),
	}
}

// CheckDatabase 检查数据库连接
func (dc *DatabaseChecker) CheckDatabase() (*DatabaseCheckResult, error) {
	startTime := time.Now()

	result := &DatabaseCheckResult{
		Type:     dc.identifyDatabaseType(),
		Status:   "unknown",
		Message:  "",
		Config:   map[string]interface{}{},
		Errors:   []string{},
		Warnings: []string{},
	}

	if result.Type == "none" {
		result.Status = "not_configured"
		result.Message = "未配置数据库"
		result.Warnings = append(result.Warnings, "数据库未配置，某些功能可能不可用")
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// 根据类型检查数据库
	var err error
	switch result.Type {
	case "postgresql":
		err = dc.checkPostgreSQL(result)
	case "mysql":
		err = dc.checkMySQL(result)
	case "redis":
		err = dc.checkRedis(result)
	default:
		result.Status = "unsupported"
		result.Message = fmt.Sprintf("不支持的数据库类型: %s", result.Type)
		result.Errors = append(result.Errors, result.Message)
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf(result.Message)
	}

	result.Duration = time.Since(startTime)
	return result, err
}

// identifyDatabaseType 识别数据库类型
func (dc *DatabaseChecker) identifyDatabaseType() string {
	config := dc.config.Database

	// 1. 检查统一URL
	if config.URL != "" {
		url := strings.ToLower(config.URL)
		if strings.HasPrefix(url, "postgres://") || strings.HasPrefix(url, "postgresql://") {
			return "postgresql"
		}
		if strings.HasPrefix(url, "mysql://") {
			return "mysql"
		}
		if strings.HasPrefix(url, "redis://") {
			return "redis"
		}
	}

	// 2. 检查MySQL配置（优先于PostgreSQL）
	if config.MySQL.Enabled && config.MySQL.Host != "" {
		return "mysql"
	}

	// 3. 检查PostgreSQL配置
	if config.PostgreSQL.Enabled && config.PostgreSQL.Host != "" {
		return "postgresql"
	}

	// 4. 检查Redis配置
	if config.Redis.Enabled && config.Redis.Host != "" {
		return "redis"
	}

	return "none" // 未配置
}

// checkPostgreSQL 检查PostgreSQL连接
func (dc *DatabaseChecker) checkPostgreSQL(result *DatabaseCheckResult) error {
	pgConfig := dc.config.Database.PostgreSQL

	// 记录配置信息
	result.Config = map[string]interface{}{
		"host":     pgConfig.Host,
		"port":     pgConfig.Port,
		"user":     pgConfig.User,
		"database": pgConfig.Database,
		"ssl_mode": pgConfig.SSLMode,
	}

	// 验证配置完整性
	if pgConfig.Host == "" {
		result.Status = "invalid_config"
		result.Message = "PostgreSQL配置不完整: HOST未设置"
		result.Errors = append(result.Errors, result.Message)
		return fmt.Errorf(result.Message)
	}

	if pgConfig.Port == 0 {
		pgConfig.Port = 5432 // 默认端口
		result.Warnings = append(result.Warnings, "PostgreSQL端口未设置，使用默认值: 5432")
	}

	if pgConfig.Database == "" {
		result.Status = "invalid_config"
		result.Message = "PostgreSQL配置不完整: DATABASE未设置"
		result.Errors = append(result.Errors, result.Message)
		return fmt.Errorf(result.Message)
	}

	if pgConfig.User == "" {
		result.Status = "invalid_config"
		result.Message = "PostgreSQL配置不完整: USER未设置"
		result.Errors = append(result.Errors, result.Message)
		return fmt.Errorf(result.Message)
	}

	// 构建DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
		pgConfig.Host, pgConfig.Port, pgConfig.User, pgConfig.Database, pgConfig.SSLMode)
	if pgConfig.Password != "" {
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			pgConfig.Host, pgConfig.Port, pgConfig.User, pgConfig.Password, pgConfig.Database, pgConfig.SSLMode)
	}

	// 尝试连接（带重试）
	var db *sql.DB
	var err error
	maxRetries := dc.config.DatabaseCheck.RetryCount
	if maxRetries <= 0 {
		maxRetries = 1
	}

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(dc.config.DatabaseCheck.RetryDelay) * time.Second)
			dc.logger.Printf("重试连接PostgreSQL (%d/%d)...", i+1, maxRetries)
		}

		db, err = sql.Open("postgres", dsn)
		if err != nil {
			result.Status = "connection_failed"
			result.Message = fmt.Sprintf("PostgreSQL连接失败: %v", err)
			result.Errors = append(result.Errors, result.Message)
			if i < maxRetries-1 {
				continue // 继续重试
			}
			return fmt.Errorf(result.Message)
		}
		defer db.Close()

		// 设置超时
		timeout := time.Duration(dc.config.DatabaseCheck.Timeout) * time.Second
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		// Ping测试
		err = db.PingContext(ctx)
		cancel()

		if err != nil {
			result.Status = "ping_failed"
			result.Message = fmt.Sprintf("PostgreSQL Ping失败: %v", err)
			result.Errors = append(result.Errors, result.Message)
			if i < maxRetries-1 {
				db.Close()
				continue // 继续重试
			}
			return fmt.Errorf(result.Message)
		}

		// 成功
		break
	}

	// 获取数据库版本
	timeout := time.Duration(dc.config.DatabaseCheck.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	var version string
	if err := db.QueryRowContext(ctx, "SELECT version()").Scan(&version); err == nil {
		result.Message = fmt.Sprintf("PostgreSQL连接成功: %s", strings.Split(version, "\n")[0])
	} else {
		result.Message = "PostgreSQL连接成功"
	}
	cancel()

	result.Status = "connected"
	return nil
}

// checkRedis 检查Redis连接
func (dc *DatabaseChecker) checkRedis(result *DatabaseCheckResult) error {
	redisConfig := dc.config.Database.Redis

	// 记录配置信息
	result.Config = map[string]interface{}{
		"host": redisConfig.Host,
		"port": redisConfig.Port,
		"db":   redisConfig.DB,
	}

	// 验证配置完整性
	if redisConfig.Host == "" {
		result.Status = "invalid_config"
		result.Message = "Redis配置不完整: HOST未设置"
		result.Errors = append(result.Errors, result.Message)
		return fmt.Errorf(result.Message)
	}

	if redisConfig.Port == 0 {
		redisConfig.Port = 6379 // 默认端口
		result.Warnings = append(result.Warnings, "Redis端口未设置，使用默认值: 6379")
	}

	// 尝试连接（带重试）
	maxRetries := dc.config.DatabaseCheck.RetryCount
	if maxRetries <= 0 {
		maxRetries = 1
	}

	var client *redis.Client
	var err error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(dc.config.DatabaseCheck.RetryDelay) * time.Second)
			dc.logger.Printf("重试连接Redis (%d/%d)...", i+1, maxRetries)
		}

		addr := fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)
		client = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: redisConfig.Password,
			DB:       redisConfig.DB,
		})

		// 设置超时
		timeout := time.Duration(dc.config.DatabaseCheck.Timeout) * time.Second
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		// Ping测试
		err = client.Ping(ctx).Err()
		cancel()

		if err != nil {
			client.Close()
			result.Status = "ping_failed"
			result.Message = fmt.Sprintf("Redis Ping失败: %v", err)
			result.Errors = append(result.Errors, result.Message)
			if i < maxRetries-1 {
				continue // 继续重试
			}
			return fmt.Errorf(result.Message)
		}

		// 成功
		defer client.Close()
		break
	}

	result.Status = "connected"
	result.Message = "Redis连接成功"
	return nil
}

// checkMySQL 检查MySQL连接
func (dc *DatabaseChecker) checkMySQL(result *DatabaseCheckResult) error {
	mysqlConfig := dc.config.Database.MySQL

	// 记录配置信息
	result.Config = map[string]interface{}{
		"host":     mysqlConfig.Host,
		"port":     mysqlConfig.Port,
		"user":     mysqlConfig.User,
		"database": mysqlConfig.Database,
	}

	// 验证配置完整性
	if mysqlConfig.Host == "" {
		result.Status = "invalid_config"
		result.Message = "MySQL配置不完整: HOST未设置"
		result.Errors = append(result.Errors, result.Message)
		return fmt.Errorf(result.Message)
	}

	if mysqlConfig.Port == 0 {
		mysqlConfig.Port = 3306 // 默认端口
		result.Warnings = append(result.Warnings, "MySQL端口未设置，使用默认值: 3306")
	}

	if mysqlConfig.Database == "" {
		result.Status = "invalid_config"
		result.Message = "MySQL配置不完整: DATABASE未设置"
		result.Errors = append(result.Errors, result.Message)
		return fmt.Errorf(result.Message)
	}

	if mysqlConfig.User == "" {
		result.Status = "invalid_config"
		result.Message = "MySQL配置不完整: USER未设置"
		result.Errors = append(result.Errors, result.Message)
		return fmt.Errorf(result.Message)
	}

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlConfig.User, mysqlConfig.Password, mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.Database)

	// 尝试连接（带重试）
	var db *sql.DB
	var err error
	maxRetries := dc.config.DatabaseCheck.RetryCount
	if maxRetries <= 0 {
		maxRetries = 1
	}

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(dc.config.DatabaseCheck.RetryDelay) * time.Second)
			dc.logger.Printf("重试连接MySQL (%d/%d)...", i+1, maxRetries)
		}

		db, err = sql.Open("mysql", dsn)
		if err != nil {
			result.Status = "connection_failed"
			result.Message = fmt.Sprintf("MySQL连接失败: %v", err)
			result.Errors = append(result.Errors, result.Message)
			if i < maxRetries-1 {
				continue // 继续重试
			}
			return fmt.Errorf(result.Message)
		}
		defer db.Close()

		// 设置超时
		timeout := time.Duration(dc.config.DatabaseCheck.Timeout) * time.Second
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		// Ping测试
		err = db.PingContext(ctx)
		cancel()

		if err != nil {
			result.Status = "ping_failed"
			result.Message = fmt.Sprintf("MySQL Ping失败: %v", err)
			result.Errors = append(result.Errors, result.Message)
			if i < maxRetries-1 {
				db.Close()
				continue // 继续重试
			}
			return fmt.Errorf(result.Message)
		}

		// 成功
		break
	}

	// 获取数据库版本
	timeout := time.Duration(dc.config.DatabaseCheck.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	var version string
	if err := db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err == nil {
		result.Message = fmt.Sprintf("MySQL连接成功: %s", version)
	} else {
		result.Message = "MySQL连接成功"
	}
	cancel()

	result.Status = "connected"
	return nil
}

// FormatDatabaseError 格式化数据库错误信息
func FormatDatabaseError(result *DatabaseCheckResult) string {
	var buf strings.Builder
	buf.WriteString("\n")
	buf.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	buf.WriteString("数据库连接检查失败\n")
	buf.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	buf.WriteString(fmt.Sprintf("数据库类型: %s\n", result.Type))
	buf.WriteString(fmt.Sprintf("状态: %s\n", result.Status))
	buf.WriteString(fmt.Sprintf("错误信息: %s\n", result.Message))
	buf.WriteString(fmt.Sprintf("耗时: %v\n", result.Duration))

	if len(result.Config) > 0 {
		buf.WriteString("\n配置信息:\n")
		for key, value := range result.Config {
			// 隐藏敏感信息
			if key == "password" {
				buf.WriteString(fmt.Sprintf("  %s: ***\n", key))
			} else {
				buf.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
			}
		}
	}

	if len(result.Errors) > 0 {
		buf.WriteString("\n错误详情:\n")
		for i, err := range result.Errors {
			buf.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err))
		}
	}

	if len(result.Warnings) > 0 {
		buf.WriteString("\n警告:\n")
		for i, warning := range result.Warnings {
			buf.WriteString(fmt.Sprintf("  %d. %s\n", i+1, warning))
		}
	}

	buf.WriteString("\n解决方案:\n")
	switch result.Type {
	case "postgresql":
		buf.WriteString("  1. 检查PostgreSQL服务是否运行\n")
		buf.WriteString("  2. 检查环境变量: POSTGRESQL_HOST, POSTGRESQL_PORT, POSTGRESQL_USER, POSTGRESQL_DATABASE\n")
		buf.WriteString("  3. 检查网络连接和防火墙设置\n")
		buf.WriteString("  4. 验证用户名和密码是否正确\n")
	case "redis":
		buf.WriteString("  1. 检查Redis服务是否运行\n")
		buf.WriteString("  2. 检查环境变量: REDIS_HOST, REDIS_PORT\n")
		buf.WriteString("  3. 检查网络连接和防火墙设置\n")
		buf.WriteString("  4. 验证Redis密码是否正确\n")
	}

	buf.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	return buf.String()
}
