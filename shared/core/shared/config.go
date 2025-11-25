package shared

import (
	"fmt"
	"os"
	"strconv"
)

// Config Central Brain配置结构
type Config struct {
	CentralBrainPort      int
	AuthServicePort       int
	AIServicePort         int
	BlockchainServicePort int
	UserServicePort       int
	JobServicePort        int
	ResumeServicePort     int
	CompanyServicePort    int
	RouterServicePort     int // Router Service端口
	PermissionServicePort int // Permission Service端口

	// 服务发现配置
	ServiceDiscovery struct {
		Enabled     bool
		ConsulURL   string
		ServiceHost string // 服务主机地址（localhost或Docker服务名）
	}

	// 服务凭证配置
	ServiceCredentials struct {
		ServiceID     string
		ServiceSecret string
	}

	// 数据库配置
	Database struct {
		// 统一URL（可选，优先级最高）
		URL string // DATABASE_URL或REDIS_URL

		// PostgreSQL配置
		PostgreSQL struct {
			Host     string
			Port     int
			User     string
			Password string
			Database string
			SSLMode  string
			Enabled  bool
		}

		// MySQL配置
		MySQL struct {
			Host     string
			Port     int
			User     string
			Password string
			Database string
			Enabled  bool
		}

		// Redis配置
		Redis struct {
			Host     string
			Port     int
			Password string
			DB       int
			Enabled  bool
		}

		// MongoDB配置
		MongoDB struct {
			URL      string // mongodb://host:port 或 mongodb+srv://host
			Database string
			Enabled  bool
		}
	}

	// 数据库检查配置
	DatabaseCheck struct {
		Enabled    bool // 是否启用数据库<｜place▁holder▁no▁196｜>
		Required   bool // 是否必需（失败时阻止启动）
		Timeout    int  // 连接超时（秒）
		RetryCount int  // 重试次数
		RetryDelay int  // 重试延迟（秒）
	}
}

// GetDefaultConfig 获取默认配置（从环境变量读取，如果没有则使用默认值）
func GetDefaultConfig() *Config {
	config := &Config{
		CentralBrainPort:      getEnvInt("CENTRAL_BRAIN_PORT", 9000),
		AuthServicePort:       getEnvInt("AUTH_SERVICE_PORT", 8207),
		AIServicePort:         getEnvInt("AI_SERVICE_PORT", 8100),
		BlockchainServicePort: getEnvInt("BLOCKCHAIN_SERVICE_PORT", 8208),
		UserServicePort:       getEnvInt("USER_SERVICE_PORT", 8082),
		JobServicePort:        getEnvInt("JOB_SERVICE_PORT", 8084),
		ResumeServicePort:     getEnvInt("RESUME_SERVICE_PORT", 8085),
		CompanyServicePort:    getEnvInt("COMPANY_SERVICE_PORT", 8083),
		RouterServicePort:     getEnvInt("ROUTER_SERVICE_PORT", 8087),
		PermissionServicePort: getEnvInt("PERMISSION_SERVICE_PORT", 8086),
	}

	// 服务发现配置
	config.ServiceDiscovery.Enabled = getEnvBool("SERVICE_DISCOVERY_ENABLED", false)
	config.ServiceDiscovery.ConsulURL = getEnvString("CONSUL_AGENT_URL", "http://localhost:8500")
	config.ServiceDiscovery.ServiceHost = getEnvString("SERVICE_HOST", "localhost")

	// 服务凭证配置
	config.ServiceCredentials.ServiceID = getEnvString("SERVICE_ID", "central-brain")
	config.ServiceCredentials.ServiceSecret = getEnvString("SERVICE_SECRET", "central-brain-secret-2025")

	// 数据库配置
	config.Database.URL = getEnvString("DATABASE_URL", "")
	if config.Database.URL == "" {
		config.Database.URL = getEnvString("REDIS_URL", "")
	}
	if config.Database.URL == "" {
		config.Database.URL = getEnvString("MONGODB_URL", "")
	}

	// PostgreSQL配置
	config.Database.PostgreSQL.Host = getEnvString("POSTGRESQL_HOST", "")
	config.Database.PostgreSQL.Port = getEnvInt("POSTGRESQL_PORT", 5432)
	config.Database.PostgreSQL.User = getEnvString("POSTGRESQL_USER", "")
	config.Database.PostgreSQL.Password = getEnvString("POSTGRESQL_PASSWORD", "")
	config.Database.PostgreSQL.Database = getEnvString("POSTGRESQL_DATABASE", "")
	config.Database.PostgreSQL.SSLMode = getEnvString("POSTGRESQL_SSL_MODE", "disable")
	config.Database.PostgreSQL.Enabled = config.Database.PostgreSQL.Host != ""

	// MySQL配置
	config.Database.MySQL.Host = getEnvString("MYSQL_HOST", "")
	config.Database.MySQL.Port = getEnvInt("MYSQL_PORT", 3306)
	config.Database.MySQL.User = getEnvString("MYSQL_USER", "")
	config.Database.MySQL.Password = getEnvString("MYSQL_PASSWORD", "")
	config.Database.MySQL.Database = getEnvString("MYSQL_DATABASE", "")
	config.Database.MySQL.Enabled = config.Database.MySQL.Host != ""

	// Redis配置
	config.Database.Redis.Host = getEnvString("REDIS_HOST", "")
	config.Database.Redis.Port = getEnvInt("REDIS_PORT", 6379)
	config.Database.Redis.Password = getEnvString("REDIS_PASSWORD", "")
	config.Database.Redis.DB = getEnvInt("REDIS_DB", 0)
	config.Database.Redis.Enabled = config.Database.Redis.Host != ""

	// MongoDB配置
	config.Database.MongoDB.URL = getEnvString("MONGODB_URL", "")
	config.Database.MongoDB.Database = getEnvString("MONGODB_DATABASE", "")
	config.Database.MongoDB.Enabled = config.Database.MongoDB.URL != ""

	// 数据库检查配置
	config.DatabaseCheck.Enabled = getEnvBool("DATABASE_CHECK_ENABLED", true)
	config.DatabaseCheck.Required = getEnvBool("DATABASE_CHECK_REQUIRED", false)
	config.DatabaseCheck.Timeout = getEnvInt("DATABASE_CHECK_TIMEOUT", 5)
	config.DatabaseCheck.RetryCount = getEnvInt("DATABASE_CHECK_RETRY_COUNT", 3)
	config.DatabaseCheck.RetryDelay = getEnvInt("DATABASE_CHECK_RETRY_DELAY", 2)

	return config
}

// LoadConfig 从环境变量加载配置（支持.env文件）
func LoadConfig() (*Config, error) {
	config := GetDefaultConfig()

	// 验证必需的配置
	if config.ServiceCredentials.ServiceSecret == "" {
		return nil, fmt.Errorf("SERVICE_SECRET未配置（必须从环境变量设置）")
	}

	return config, nil
}

// 辅助函数：从环境变量读取字符串，如果没有则返回默认值
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 辅助函数：从环境变量读取整数，如果没有则返回默认值
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// 辅助函数：从环境变量读取布尔值，如果没有则返回默认值
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
