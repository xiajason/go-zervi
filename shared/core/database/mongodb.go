package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBManager MongoDB数据库管理器
type MongoDBManager struct {
	client   *mongo.Client
	database *mongo.Database
	config   MongoDBConfig
}

// MongoDBConfig MongoDB配置
type MongoDBConfig struct {
	URL            string        `json:"url"`
	Database       string        `json:"database"`
	ConnectTimeout time.Duration `json:"connect_timeout"`
	MaxPoolSize    uint64        `json:"max_pool_size"`
	MinPoolSize    uint64        `json:"min_pool_size"`
}

// NewMongoDBManager 创建MongoDB管理器
func NewMongoDBManager(config MongoDBConfig) (*MongoDBManager, error) {
	// 设置默认值
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 10 * time.Second
	}
	if config.MaxPoolSize == 0 {
		config.MaxPoolSize = 100
	}
	if config.MinPoolSize == 0 {
		config.MinPoolSize = 10
	}

	// 创建客户端选项
	clientOptions := options.Client().ApplyURI(config.URL)
	clientOptions.ConnectTimeout = &config.ConnectTimeout
	clientOptions.MaxPoolSize = &config.MaxPoolSize
	clientOptions.MinPoolSize = &config.MinPoolSize

	// 连接MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("MongoDB连接失败: %w", err)
	}

	// 测试连接
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		client.Disconnect(context.Background())
		return nil, fmt.Errorf("MongoDB Ping失败: %w", err)
	}

	// 获取数据库实例
	database := client.Database(config.Database)

	return &MongoDBManager{
		client:   client,
		database: database,
		config:   config,
	}, nil
}

// GetDB 获取数据库实例（兼容现有代码）
func (mm *MongoDBManager) GetDB() *mongo.Database {
	return mm.database
}

// GetClient 获取MongoDB客户端
func (mm *MongoDBManager) GetClient() *mongo.Client {
	return mm.client
}

// GetCollection 获取集合
func (mm *MongoDBManager) GetCollection(name string) *mongo.Collection {
	return mm.database.Collection(name)
}

// Close 关闭数据库连接
func (mm *MongoDBManager) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return mm.client.Disconnect(ctx)
}

// Ping 测试数据库连接
func (mm *MongoDBManager) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return mm.client.Ping(ctx, nil)
}

// Health 健康检查
func (mm *MongoDBManager) Health() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 测试连接
	pingErr := mm.Ping()
	status := "healthy"
	if pingErr != nil {
		status = "unhealthy"
	}

	// 获取服务器信息
	var serverStatus map[string]interface{}
	var version string
	if err := mm.client.Database("admin").RunCommand(ctx, map[string]interface{}{"buildInfo": 1}).Decode(&serverStatus); err == nil {
		if v, ok := serverStatus["version"].(string); ok {
			version = v
		}
	}

	// 获取数据库统计信息
	var dbStats map[string]interface{}
	mm.database.RunCommand(ctx, map[string]interface{}{"dbStats": 1}).Decode(&dbStats)

	return map[string]interface{}{
		"status":   status,
		"url":      mm.config.URL,
		"database": mm.config.Database,
		"version":  version,
		"db_stats": dbStats,
		"error":    pingErr,
	}
}

// StartSession 启动会话（用于事务）
func (mm *MongoDBManager) StartSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	return mm.client.StartSession(opts...)
}

// WithTransaction 执行事务
func (mm *MongoDBManager) WithTransaction(ctx context.Context, fn func(mongo.SessionContext) error, opts ...*options.TransactionOptions) error {
	session, err := mm.client.StartSession()
	if err != nil {
		return fmt.Errorf("启动MongoDB会话失败: %w", err)
	}
	defer session.EndSession(ctx)

	return mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := session.StartTransaction(opts...); err != nil {
			return fmt.Errorf("启动MongoDB事务失败: %w", err)
		}

		if err := fn(sc); err != nil {
			if abortErr := session.AbortTransaction(sc); abortErr != nil {
				return fmt.Errorf("回滚MongoDB事务失败: %w (原始错误: %v)", abortErr, err)
			}
			return err
		}

		if err := session.CommitTransaction(sc); err != nil {
			return fmt.Errorf("提交MongoDB事务失败: %w", err)
		}

		return nil
	})
}


