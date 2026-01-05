package shared

import (
	"os"
	"strconv"
)

// Config 共享配置
type Config struct {
	// 数据库配置
	PostgreSQLURL string
	RedisURL      string
	JWTSecret     string

	// 服务端口
	CentralBrainPort      int
	AuthServicePort       int
	AIServicePort         int
	BlockchainServicePort int
	UserServicePort       int
	JobServicePort        int
	ResumeServicePort     int
	CompanyServicePort    int

	// 服务发现配置
	ServiceDiscoveryEnabled bool
	ConsulAgentURL          string

	// 安全配置
	TokenValidationEnabled bool

	// 监控配置
	MetricsEnabled bool
	LoggingLevel   string

	// 环境
	Environment string
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *Config {
	return &Config{
		PostgreSQLURL:           getEnv("POSTGRESQL_URL", "postgres://postgres:dev_password@localhost:5432/zervigo_mvp?sslmode=disable"),
		RedisURL:                getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:               getEnv("JWT_SECRET", "zervigo-mvp-secret-key-2025"),
		CentralBrainPort:        getEnvInt("CENTRAL_BRAIN_PORT", 9000),
		AuthServicePort:         getEnvInt("AUTH_SERVICE_PORT", 8207),
		AIServicePort:           getEnvInt("AI_SERVICE_PORT", 8100),
		BlockchainServicePort:   getEnvInt("BLOCKCHAIN_SERVICE_PORT", 8208),
		UserServicePort:         getEnvInt("USER_SERVICE_PORT", 8082),
		JobServicePort:          getEnvInt("JOB_SERVICE_PORT", 8084),
		ResumeServicePort:       getEnvInt("RESUME_SERVICE_PORT", 8085),
		CompanyServicePort:      getEnvInt("COMPANY_SERVICE_PORT", 8083),
		ServiceDiscoveryEnabled: getEnvBool("SERVICE_DISCOVERY_ENABLED", false),
		ConsulAgentURL:          getEnv("CONSUL_AGENT_URL", "http://localhost:8500"),
		TokenValidationEnabled:  getEnvBool("TOKEN_VALIDATION_ENABLED", true),
		MetricsEnabled:          getEnvBool("METRICS_ENABLED", true),
		LoggingLevel:            getEnv("LOGGING_LEVEL", "INFO"),
		Environment:             getEnv("ENVIRONMENT", "development"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool 获取布尔环境变量
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
