# MongoDB 使用指南

本文档说明如何在 gozervi 微服务架构中使用 MongoDB。

## 配置

### 环境变量

```bash
# MongoDB 连接 URL（必需）
export MONGODB_URL="mongodb://localhost:27017"
# 或使用 MongoDB Atlas
# export MONGODB_URL="mongodb+srv://username:password@cluster.mongodb.net/"

# MongoDB 数据库名（必需）
export MONGODB_DATABASE="your_database_name"
```

### 配置优先级

1. **统一 URL** (`DATABASE_URL` 或 `MONGODB_URL`) - 优先级最高
2. **MongoDB 配置项** (`MONGODB_URL` + `MONGODB_DATABASE`)

## 在微服务中使用

### 1. 初始化数据库管理器

```go
package main

import (
    "github.com/szjason72/zervigo/shared/core/database"
    "github.com/szjason72/zervigo/shared/core/shared"
)

func main() {
    // 加载配置
    config, err := shared.LoadConfig()
    if err != nil {
        log.Fatalf("配置加载失败: %v", err)
    }

    // 创建数据库配置
    dbConfig := database.Config{
        MongoDB: database.MongoDBConfig{
            URL:            config.Database.MongoDB.URL,
            Database:       config.Database.MongoDB.Database,
            ConnectTimeout: 10 * time.Second,
            MaxPoolSize:    100,
            MinPoolSize:    10,
        },
        // 可以同时配置其他数据库
        Redis: database.RedisConfig{
            Host:     config.Database.Redis.Host,
            Port:     config.Database.Redis.Port,
            Password: config.Database.Redis.Password,
            Database: config.Database.Redis.DB,
        },
    }

    // 创建数据库管理器
    dbManager, err := database.NewManager(dbConfig)
    if err != nil {
        log.Fatalf("数据库管理器初始化失败: %v", err)
    }
    defer dbManager.Close()

    // 使用 MongoDB
    if dbManager.MongoDB != nil {
        mongoDB := dbManager.MongoDB.GetDB()
        // 初始化 DAO
        dao.InitDao(mongoDB, dbManager.Redis.GetClient())
    }
}
```

### 2. 在 DAO 层使用

```go
package dao

import (
    "context"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson"
)

type UserDao struct {
    coll *mongo.Collection
}

func NewUserDao(db *mongo.Database) *UserDao {
    return &UserDao{
        coll: db.Collection("users"),
    }
}

func (dao *UserDao) GetUser(ctx context.Context, userID string) (*User, error) {
    var user User
    err := dao.coll.FindOne(ctx, bson.M{"userId": userID}).Decode(&user)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (dao *UserDao) CreateUser(ctx context.Context, user *User) error {
    _, err := dao.coll.InsertOne(ctx, user)
    return err
}
```

### 3. 使用事务

```go
// MongoDB 支持事务（需要副本集）
err := dbManager.MongoDB.WithTransaction(ctx, func(sc mongo.SessionContext) error {
    // 在事务中执行操作
    if err := userDao.CreateUser(sc, user); err != nil {
        return err
    }
    if err := profileDao.CreateProfile(sc, profile); err != nil {
        return err
    }
    return nil
})
```

### 4. 健康检查

```go
// 检查 MongoDB 连接
health := dbManager.MongoDB.Health()
fmt.Printf("MongoDB 状态: %v\n", health)

// 或使用统一的健康检查
allHealth := dbManager.Health()
fmt.Printf("所有数据库状态: %v\n", allHealth)
```

## 与现有项目集成

### 适配现有 DAO 层

现有项目的 DAO 层使用 `*mongo.Database`，可以直接复用：

```go
// 现有代码
func InitDao(mgoDb *mongo.Database, rdb *redis.Client) error {
    UserDao = NewUserDao(mgoDb)
    ProjectDao = NewProjectDao(mgoDb)
    // ...
    return nil
}

// 在微服务中使用
dbManager, _ := database.NewManager(dbConfig)
mongoDB := dbManager.MongoDB.GetDB()
redisClient := dbManager.Redis.GetClient()
dao.InitDao(mongoDB, redisClient) // 完全兼容！
```

## 迁移路径

### 阶段 1: 保持 MongoDB（当前阶段）

- ✅ 使用 MongoDB Manager
- ✅ 复用现有 DAO 代码
- ✅ 无需数据迁移

### 阶段 2: 混合使用（可选）

- 部分服务使用 MongoDB
- 部分服务使用 PostgreSQL
- 通过 Central Brain 统一路由

### 阶段 3: 逐步迁移（可选）

- 新功能使用 PostgreSQL
- 旧功能保持 MongoDB
- 通过数据同步保持一致性

## 示例：完整的微服务

```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/szjason72/zervigo/shared/core/database"
    "github.com/szjason72/zervigo/shared/core/shared"
    "go.mongodb.org/mongo-driver/mongo"
)

type UserService struct {
    dbManager *database.Manager
    userDao   *UserDao
}

func NewUserService() (*UserService, error) {
    // 加载配置
    config, err := shared.LoadConfig()
    if err != nil {
        return nil, err
    }

    // 创建数据库管理器
    dbConfig := database.Config{
        MongoDB: database.MongoDBConfig{
            URL:            config.Database.MongoDB.URL,
            Database:       config.Database.MongoDB.Database,
            ConnectTimeout: 10 * time.Second,
        },
        Redis: database.RedisConfig{
            Host:     config.Database.Redis.Host,
            Port:     config.Database.Redis.Port,
            Password: config.Database.Redis.Password,
        },
    }

    dbManager, err := database.NewManager(dbConfig)
    if err != nil {
        return nil, err
    }

    // 初始化 DAO
    mongoDB := dbManager.MongoDB.GetDB()
    userDao := NewUserDao(mongoDB)

    return &UserService{
        dbManager: dbManager,
        userDao:   userDao,
    }, nil
}

func (s *UserService) GetUser(c *gin.Context) {
    userID := c.Param("id")
    user, err := s.userDao.GetUser(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
        return
    }
    c.JSON(http.StatusOK, user)
}

func main() {
    service, err := NewUserService()
    if err != nil {
        log.Fatalf("服务初始化失败: %v", err)
    }
    defer service.dbManager.Close()

    router := gin.Default()
    router.GET("/api/v1/users/:id", service.GetUser)

    log.Println("用户服务启动在端口 8082")
    router.Run(":8082")
}
```

## 注意事项

1. **连接池**: MongoDB Manager 自动管理连接池，无需手动管理
2. **超时设置**: 默认连接超时为 10 秒，可通过配置调整
3. **事务支持**: MongoDB 事务需要副本集，单节点不支持
4. **健康检查**: 启动时会自动检查 MongoDB 连接（如果启用）
5. **错误处理**: 连接失败时会自动重试（根据配置）

## 与 Central Brain 集成

Central Brain 会自动识别 MongoDB 配置并检查连接：

```bash
# 启动时会看到：
🔍 检查数据库连接...
✅ 数据库连接检查成功: MongoDB连接成功: 4.4.0 (耗时: 123ms)
```

如果连接失败，会根据配置决定是否阻止启动（`DATABASE_CHECK_REQUIRED=true`）。




