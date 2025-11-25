package jobfirst

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/szjason72/zervigo/shared/core/auth"
	"github.com/szjason72/zervigo/shared/core/config"
	"github.com/szjason72/zervigo/shared/core/database"
	"github.com/szjason72/zervigo/shared/core/logger"
	"github.com/szjason72/zervigo/shared/core/service/errors"
	"github.com/szjason72/zervigo/shared/core/service/health"
	"github.com/szjason72/zervigo/shared/core/service/registry"

	// 注释掉superadmin导入，因为它是独立模块
	// "github.com/szjason72/zervigo/shared/core/superadmin"
	"github.com/szjason72/zervigo/shared/core/team"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// AuthMiddlewareInterface 认证中间件接口
type AuthMiddlewareInterface interface {
	RequireAuth() gin.HandlerFunc
	RequireDevTeam() gin.HandlerFunc
}

// Core JobFirst核心包
type Core struct {
	Config         *config.Manager
	Database       *database.Manager
	Logger         *logger.Manager
	AuthManager    *auth.AuthManager
	TeamManager    *team.Manager
	AuthMiddleware AuthMiddlewareInterface
	ErrorHandler   *errors.ErrorHandler
	ServiceHealth  *health.ServiceHealth
	ConsulRegistry *registry.ConsulRegistry
	// SuperAdmin     *superadmin.Manager  // 注释掉，因为superadmin是独立模块
}

// NewCore 创建JobFirst核心实例
func NewCore(configPath string) (*Core, error) {
	// 1. 初始化配置管理器
	configManager, err := config.NewManager(configPath)
	if err != nil {
		return nil, fmt.Errorf("初始化配置管理器失败: %w", err)
	}

	// 2. 加载应用配置
	appConfig, err := configManager.LoadAppConfig()
	if err != nil {
		return nil, fmt.Errorf("加载应用配置失败: %w", err)
	}

	// 设置默认值（如果配置文件不存在）
	if appConfig.Database.Database == "" {
		appConfig.Database.Database = "zervigo_mvp"
	}
	if appConfig.Database.Host == "" {
		appConfig.Database.Host = "localhost"
	}
	if appConfig.Database.Port == 0 {
		appConfig.Database.Port = 3306
	}
	if appConfig.Database.Username == "" {
		appConfig.Database.Username = "root"
	}
	if appConfig.Auth.JWTSecret == "" {
		appConfig.Auth.JWTSecret = "zervigo-mvp-secret-key-2025"
	}
	if appConfig.Log.Level == "" {
		appConfig.Log.Level = "info"
	}
	if appConfig.Log.Format == "" {
		appConfig.Log.Format = "json"
	}
	if appConfig.Log.Output == "" {
		appConfig.Log.Output = "stdout"
	}

	// 3. 初始化日志管理器
	logConfig := logger.Config{
		Level:  logger.Level(appConfig.Log.Level),
		Format: logger.Format(appConfig.Log.Format),
		Output: appConfig.Log.Output,
		File:   appConfig.Log.File,
	}

	logManager, err := logger.NewManager(logConfig)
	if err != nil {
		return nil, fmt.Errorf("初始化日志管理器失败: %w", err)
	}

	// 设置全局日志
	if err := logger.InitGlobal(logConfig); err != nil {
		return nil, fmt.Errorf("初始化全局日志失败: %w", err)
	}

	// 4. 初始化数据库管理器
	mysqlHost := getEnvString("MYSQL_HOST", "")
	mysqlUser := getEnvString("MYSQL_USER", "")

	pgHost := getEnvString("POSTGRESQL_HOST", "")
	if pgHost == "" {
		pgHost = getEnvString("POSTGRES_HOST", "")
	}
	if pgHost == "" {
		pgHost = "localhost"
	}
	pgPort := getEnvInt("POSTGRESQL_PORT", getEnvInt("POSTGRES_PORT", 5432))
	pgUser := getEnvString("POSTGRESQL_USER", "")
	if pgUser == "" {
		pgUser = getEnvString("POSTGRES_USER", "")
	}
	if pgUser == "" {
		pgUser = os.Getenv("USER")
	}
	if pgUser == "" {
		pgUser = "postgres"
	}
	pgPassword := getEnvString("POSTGRESQL_PASSWORD", getEnvString("POSTGRES_PASSWORD", ""))
	pgDatabase := getEnvString("POSTGRESQL_DATABASE", "")
	if pgDatabase == "" {
		pgDatabase = getEnvString("POSTGRES_DB", "zervigo_mvp")
	}
	pgSSLMode := getEnvString("POSTGRESQL_SSL_MODE", "disable")

	dbConfig := database.Config{
		MySQL: database.MySQLConfig{
			// 仅当显式配置时才启用 MySQL
			Host:        mysqlHost,
			Port:        getEnvInt("MYSQL_PORT", 3306),
			Username:    mysqlUser,
			Password:    getEnvString("MYSQL_PASSWORD", ""),
			Database:    getEnvString("MYSQL_DATABASE", ""),
			Charset:     "utf8mb4",
			MaxIdle:     10,
			MaxOpen:     100,
			MaxLifetime: parseDuration("1h"),
			LogLevel:    parseGORMLogLevel("warn"),
		},
		Redis: database.RedisConfig{
			Host:     getEnvString("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnvString("REDIS_PASSWORD", ""),
			Database: getEnvInt("REDIS_DB", 0),
			PoolSize: 10,
			MinIdle:  5,
		},
		PostgreSQL: database.PostgreSQLConfig{
			Host:        pgHost,
			Port:        pgPort,
			Username:    pgUser,
			Password:    pgPassword,
			Database:    pgDatabase,
			SSLMode:     pgSSLMode,
			MaxIdle:     10,
			MaxOpen:     100,
			MaxLifetime: parseDuration("1h"),
			LogLevel:    parseGORMLogLevel("warn"),
		},
		Neo4j: database.Neo4jConfig{
			URI:      "", // 设置为空以禁用Neo4j
			Username: "neo4j",
			Password: "password",
			Database: "neo4j",
		},
		MongoDB: database.MongoDBConfig{
			URL:            getEnvString("MONGODB_URL", ""),
			Database:       getEnvString("MONGODB_DATABASE", ""),
			ConnectTimeout: parseDuration("10s"),
			MaxPoolSize:    100,
			MinPoolSize:    10,
		},
	}

	dbManager, err := database.NewManager(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库管理器失败: %w", err)
	}

	// 5. 执行数据库迁移（迁移失败时继续启动服务）
	if dbManager.GetMySQL() != nil {
		if err := dbManager.GetMySQL().Migrate(&auth.User{}, &auth.DevTeamUser{}, &auth.DevOperationLog{}); err != nil {
			// 记录迁移错误但不中断服务启动
			fmt.Printf("警告: MySQL 数据库迁移失败，但服务将继续启动: %v\n", err)
		}
	}

	// 6. 初始化认证管理器
	authConfig := auth.AuthConfig{
		JWTSecret:        appConfig.Auth.JWTSecret,
		TokenExpiry:      parseDuration(appConfig.Auth.TokenExpiry),
		RefreshExpiry:    parseDuration(appConfig.Auth.RefreshExpiry),
		PasswordMin:      appConfig.Auth.PasswordMin,
		MaxLoginAttempts: appConfig.Auth.MaxLoginAttempts,
		LockoutDuration:  parseDuration(appConfig.Auth.LockoutDuration),
	}

	authManager := auth.NewAuthManager(dbManager.GetDB(), authConfig)

	// 7. 初始化团队管理器
	teamManager := team.NewManager(dbManager.GetDB())

	// 8. 初始化认证中间件（使用Go-Zervi适配器）
	// 智能识别：根据环境变量选择数据库（优先MySQL）
	var dbType string
	var gormDB *gorm.DB

	// 优先使用MySQL（如果配置了）
	if dbManager.GetPostgreSQL() != nil {
		gormDB = dbManager.GetPostgreSQL().GetDB()
		dbType = "PostgreSQL"
	} else if dbManager.GetMySQL() != nil {
		gormDB = dbManager.GetMySQL().GetDB()
		dbType = "MySQL"
	} else {
		return nil, fmt.Errorf("未检测到任何数据库配置（MySQL或PostgreSQL）")
	}

	fmt.Printf("INFO: 使用 %s 数据库进行认证\n", dbType)

	// 获取原生SQL连接（使用database/sql包）
	type SQLDB interface {
		DB() (interface{}, error)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("获取%s SQL连接失败: %w", dbType, err)
	}

	// 创建Go-Zervi认证适配器
	zerviAuthAdapter := auth.NewZerviAuthAdapter(sqlDB, appConfig.Auth.JWTSecret)

	// 为了兼容性，创建一个包装器
	// authMiddleware := &ZerviAuthMiddlewareWrapper{adapter: zerviAuthAdapter}

	// 9. 初始化错误处理器
	errorHandler := errors.NewErrorHandler()

	// 10. 初始化服务健康检查器
	serviceHealth := health.NewServiceHealth("jobfirst-core", "1.0.0")

	// 添加数据库健康检查
	if dbManager.GetDB() != nil {
		dbChecker := health.NewComponentHealth("database", func() error {
			sqlDB, err := dbManager.GetDB().DB()
			if err != nil {
				return err
			}
			return sqlDB.Ping()
		})
		serviceHealth.AddChecker(dbChecker)
	}

	// 11. 初始化Consul注册器
	consulRegistry, err := registry.NewConsulRegistry("localhost:8500")
	if err != nil {
		logManager.Warn("创建Consul注册器失败: %v", err)
	}

	// 12. 注释掉超级管理员管理器初始化，因为superadmin是独立模块
	// superAdminConfig := &superadmin.Config{
	// 	System: system.MonitorConfig{
	// 		// 系统监控配置
	// 	},
	// 	User: user.UserConfig{
	// 		// 用户管理配置
	// 	},
	// 	Database: superadmindatabase.DatabaseConfig{
	// 		// 数据库管理配置
	// 	},
	// 	AI: ai.AIConfig{
	// 		// AI管理配置
	// 	},
	// 	Config: superadminconfig.ConfigManagerConfig{
	// 		// 配置管理配置
	// 	},
	// 	CICD: cicd.CICDConfig{
	// 		// CI/CD管理配置
	// 	},
	// }
	// superAdminManager, err := superadmin.NewManager(superAdminConfig)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create super admin manager: %v", err)
	// }

	core := &Core{
		Config:         configManager,
		Database:       dbManager,
		Logger:         logManager,
		AuthManager:    authManager,
		TeamManager:    teamManager,
		AuthMiddleware: &ZerviAuthMiddlewareInterface{adapter: zerviAuthAdapter},
		ErrorHandler:   errorHandler,
		ServiceHealth:  serviceHealth,
		ConsulRegistry: consulRegistry,
		// SuperAdmin:     superAdminManager,  // 注释掉
	}

	logManager.Info("JobFirst核心包初始化成功")
	return core, nil
}

// GetDB 获取数据库实例
func (c *Core) GetDB() *gorm.DB {
	return c.Database.GetDB()
}

// Close 关闭核心包
func (c *Core) Close() error {
	if err := c.Database.Close(); err != nil {
		return fmt.Errorf("关闭数据库连接失败: %w", err)
	}
	c.Logger.Info("JobFirst核心包已关闭")
	return nil
}

// Health 健康检查
func (c *Core) Health() map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// 检查数据库健康状态
	dbHealth := c.Database.Health()
	health["database"] = dbHealth

	// 检查配置
	health["config"] = map[string]interface{}{
		"loaded": c.Config != nil,
	}

	// 检查日志
	health["logger"] = map[string]interface{}{
		"initialized": c.Logger != nil,
	}

	return health
}

// CreateServiceTemplate 创建服务模板
// func (c *Core) CreateServiceTemplate(config *template.ServiceConfig) (*template.ServiceTemplate, error) {
// 	return template.NewServiceTemplate(config)
// }

// GetErrorHandler 获取错误处理器
func (c *Core) GetErrorHandler() *errors.ErrorHandler {
	return c.ErrorHandler
}

// GetServiceHealth 获取服务健康检查器
func (c *Core) GetServiceHealth() *health.ServiceHealth {
	return c.ServiceHealth
}

// GetConsulRegistry 获取Consul注册器
func (c *Core) GetConsulRegistry() *registry.ConsulRegistry {
	return c.ConsulRegistry
}

// 辅助函数

// parseDuration 解析时间字符串
func parseDuration(s string) time.Duration {
	if s == "" {
		return time.Hour * 24 // 默认24小时
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return time.Hour * 24 // 默认24小时
	}

	return duration
}

// parseGORMLogLevel 解析GORM日志级别
func parseGORMLogLevel(level string) gormlogger.LogLevel {
	switch level {
	case "trace":
		return gormlogger.Silent
	case "debug":
		return gormlogger.Info
	case "info":
		return gormlogger.Info
	case "warn":
		return gormlogger.Warn
	case "error":
		return gormlogger.Error
	case "fatal":
		return gormlogger.Error
	case "panic":
		return gormlogger.Error
	default:
		return gormlogger.Info
	}
}

// ZerviAuthMiddlewareWrapper 包装器，使Go-Zervi认证适配器兼容jobfirst-core接口
type ZerviAuthMiddlewareWrapper struct {
	adapter *auth.ZerviAuthAdapter
}

// RequireAuth 需要登录的中间件
func (w *ZerviAuthMiddlewareWrapper) RequireAuth() gin.HandlerFunc {
	return w.adapter.RequireAuth()
}

// RequireDevTeam 需要开发团队权限的中间件
func (w *ZerviAuthMiddlewareWrapper) RequireDevTeam() gin.HandlerFunc {
	return w.adapter.RequireDevTeam()
}

// ZerviAuthMiddlewareInterface 接口，使Go-Zervi认证适配器兼容jobfirst-core接口
type ZerviAuthMiddlewareInterface struct {
	adapter *auth.ZerviAuthAdapter
}

// RequireAuth 需要登录的中间件
func (w *ZerviAuthMiddlewareInterface) RequireAuth() gin.HandlerFunc {
	return w.adapter.RequireAuth()
}

// RequireDevTeam 需要开发团队权限的中间件
func (w *ZerviAuthMiddlewareInterface) RequireDevTeam() gin.HandlerFunc {
	return w.adapter.RequireDevTeam()
}

// getEnvString 从环境变量读取字符串，如果不存在则返回默认值
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 从环境变量读取整数，如果不存在则返回默认值
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
