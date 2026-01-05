# JobFirst-Core 与 Go-Zervi 框架集成指南

## 📋 文档概述

本文档提供JobFirst-Core与Go-Zervi框架集成的具体实现指南，包括代码示例、配置说明和最佳实践。

**文档版本**: v1.0  
**创建日期**: 2025-10-29  
**作者**: Go-Zervi Framework Team  
**状态**: 已完成

## 🔧 核心集成组件

### 1. 认证适配器 (ZerviAuthAdapter)

**文件位置**: `shared/core/auth/zervi_auth_adapter.go`

```go
// ZerviAuthAdapter Go-Zervi认证适配器
// 将jobfirst-core的认证中间件适配到Go-Zervi统一认证系统
type ZerviAuthAdapter struct {
    unifiedAuth *UnifiedAuthSystem
}

// NewZerviAuthAdapter 创建Go-Zervi认证适配器
func NewZerviAuthAdapter(db *sql.DB, jwtSecret string) *ZerviAuthAdapter {
    unifiedAuth := NewUnifiedAuthSystem(db, jwtSecret)
    return &ZerviAuthAdapter{
        unifiedAuth: unifiedAuth,
    }
}
```

**关键特性**:
- 统一认证系统与jobfirst-core接口的桥接
- 支持PostgreSQL原生SQL连接
- 实现Go-Zervi标准响应格式

### 2. 中间件包装器

**文件位置**: `shared/core/core.go`

```go
// AuthMiddlewareInterface 认证中间件接口
type AuthMiddlewareInterface interface {
    RequireAuth() gin.HandlerFunc
    RequireDevTeam() gin.HandlerFunc
}

// ZerviAuthMiddlewareInterface 接口，使Go-Zervi认证适配器兼容jobfirst-core接口
type ZerviAuthMiddlewareInterface struct {
    adapter *auth.ZerviAuthAdapter
}

// RequireAuth 需要登录的中间件
func (w *ZerviAuthMiddlewareInterface) RequireAuth() gin.HandlerFunc {
    return w.adapter.RequireAuth()
}
```

### 3. 数据库连接适配

**PostgreSQL管理器扩展**:

```go
// GetSQLDB 获取原生SQL数据库连接
func (pm *PostgreSQLManager) GetSQLDB() (*sql.DB, error) {
    return pm.db.DB()
}
```

**核心集成代码**:

```go
// 获取PostgreSQL数据库连接
pgManager := dbManager.GetPostgreSQL()
if pgManager == nil {
    return nil, fmt.Errorf("PostgreSQL数据库未初始化")
}

// 获取原生SQL连接
sqlDB, err := pgManager.GetSQLDB()
if err != nil {
    return nil, fmt.Errorf("获取PostgreSQL SQL连接失败: %w", err)
}

// 创建Go-Zervi认证适配器
zerviAuthAdapter := auth.NewZerviAuthAdapter(sqlDB, appConfig.Auth.JWTSecret)
```

## ⚙️ 配置集成

### 1. 数据库配置

**YAML配置** (`configs/jobfirst-core-config.yaml`):

```yaml
database:
  # PostgreSQL配置 (主要数据库)
  postgres:
    host: localhost
    port: 5432
    username: szjason72
    password: ""
    database: zervigo_mvp
    sslmode: disable
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600
  
  # MySQL配置 (备用)
  mysql:
    host: ""  # 空字符串禁用MySQL
    port: 3306
    username: root
    password: ""
    database: zervigo_mvp
```

**环境变量配置** (`.env`):

```bash
# PostgreSQL数据库配置 (主要数据库)
POSTGRESQL_URL=postgres://postgres:dev_password@localhost:5432/zervigo_mvp?sslmode=disable
POSTGRES_DB=zervigo_mvp
POSTGRES_USER=postgres
POSTGRES_PASSWORD=dev_password

# 禁用MySQL
MYSQL_HOST=
MYSQL_PORT=3306
```

### 2. 认证配置

```yaml
# JWT配置
jwt:
  secret: "zervigo-mvp-secret-key-2025"
  access_expire: 7200  # 2小时
  refresh_secret: "zervigo-mvp-refresh-secret-key-2025"
  refresh_expire: 604800  # 7天
```

## 🚀 服务集成示例

### 1. 用户服务集成

**文件位置**: `services/core/user/main.go`

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/szjason72/zervigo/shared/core"
    "github.com/szjason72/zervigo/shared/core/response"
)

func main() {
    // 初始化jobfirst-core
    coreInstance, err := jobfirst.NewCore("./configs/jobfirst-core-config.yaml")
    if err != nil {
        log.Fatalf("初始化核心失败: %v", err)
    }

    // 设置Gin模式
    gin.SetMode(gin.ReleaseMode)
    r := gin.Default()

    // 使用认证中间件
    authMiddleware := coreInstance.AuthMiddleware

    // 用户API路由
    users := r.Group("/api/v1/users")
    users.Use(authMiddleware.RequireAuth())
    {
        users.GET("/", getUserList)
        users.GET("/:id", getUserByID)
        users.POST("/", createUser)
        users.PUT("/:id", updateUser)
        users.DELETE("/:id", deleteUser)
    }

    // 健康检查
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "healthy",
            "service": "user-service",
        })
    })

    // 启动服务
    log.Printf("Starting User Service on 0.0.0.0:8082")
    r.Run(":8082")
}
```

### 2. API响应格式统一

**标准成功响应**:

```go
func getUserList(c *gin.Context) {
    // 业务逻辑...
    
    // 使用Go-Zervi标准响应格式
    resp := response.Success("获取用户列表成功", response.NewPageResponse(
        users,      // 用户列表
        total,      // 总数
        pageNum,    // 页码
        pageSize,   // 页大小
    ))
    
    c.JSON(http.StatusOK, resp)
}
```

**标准错误响应**:

```go
func getUserByID(c *gin.Context) {
    userID := c.Param("id")
    
    if userID == "" {
        resp := response.Error(response.CodeInvalidParams, "用户ID不能为空")
        c.JSON(http.StatusOK, resp)
        return
    }
    
    // 业务逻辑...
}
```

## 🔍 认证流程集成

### 1. 登录流程

```go
func (adapter *ZerviAuthAdapter) handleLogin(w http.ResponseWriter, r *http.Request) {
    // 解析登录请求
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        adapter.writeErrorResponse(w, response.Error(response.CodeInvalidParams, "Invalid JSON"))
        return
    }
    
    // 调用统一认证系统
    result, err := adapter.unifiedAuth.Authenticate(req.Username, req.Password)
    if err != nil {
        adapter.writeErrorResponse(w, response.Error(response.CodeInternalError, err.Error()))
        return
    }
    
    // 构建标准响应
    if result.Success && result.User != nil {
        loginData := map[string]interface{}{
            "userId":       result.User.ID,
            "userName":     result.User.Username,
            "userPhone":    result.User.Phone,
            "userStatus":   result.User.Status,
            "loginStatus":  adapter.calculateLoginStatus(adapter.getUserStatusInt(result.User.Status)),
            "accessToken":  result.Token,
            "refreshToken": "",
        }
        adapter.writeSuccessResponse(w, response.Success("登录成功", loginData))
    } else {
        adapter.writeErrorResponse(w, response.Error(response.CodeUserNotFound, result.Error))
    }
}
```

### 2. JWT验证流程

```go
func (adapter *ZerviAuthAdapter) RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := adapter.extractToken(c)
        
        if token == "" {
            adapter.writeErrorResponse(c, response.Error(response.CodeUnauthorized, "未登录"))
            c.Abort()
            return
        }
        
        result, err := adapter.unifiedAuth.ValidateJWT(token)
        if err != nil || !result.Success {
            adapter.writeErrorResponse(c, response.Error(response.CodeUnauthorized, "无效的token"))
            c.Abort()
            return
        }
        
        // 设置用户信息到上下文
        c.Set("user_id", result.User.ID)
        c.Set("username", result.User.Username)
        c.Set("role", result.User.Role)
        c.Set("email", result.User.Email)
        
        c.Next()
    }
}
```

## 📊 数据库集成

### 1. PostgreSQL连接配置

```go
// 在core.go中的数据库配置
dbConfig := database.Config{
    PostgreSQL: database.PostgreSQLConfig{
        Host:        "localhost",
        Port:        5432,
        Username:    "szjason72",
        Password:    "",
        Database:    "zervigo_mvp",
        SSLMode:     "disable",
        MaxIdle:     10,
        MaxOpen:     100,
        MaxLifetime: parseDuration("1h"),
        LogLevel:    parseGORMLogLevel("warn"),
    },
    // MySQL配置为空，禁用MySQL
    MySQL: database.MySQLConfig{
        Host: "", // 禁用MySQL
    },
}
```

### 2. 数据库迁移

```go
// 执行数据库迁移（迁移失败时继续启动服务）
if err := dbManager.Migrate(&auth.User{}, &auth.DevTeamUser{}, &auth.DevOperationLog{}); err != nil {
    // 记录迁移错误但不中断服务启动
    logManager.Warn("数据库迁移失败，但服务将继续启动: %v", err)
}
```

## 🧪 测试和验证

### 1. 服务健康检查

```bash
# 检查用户服务健康状态
curl -s http://localhost:8082/health | jq .status

# 预期输出: "healthy"
```

### 2. 认证测试

```bash
# 测试登录API
curl -X POST http://localhost:8207/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq .

# 预期输出: Go-Zervi标准响应格式
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "userId": 1,
    "userName": "admin",
    "userPhone": null,
    "userStatus": "active",
    "loginStatus": 0,
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": ""
  },
  "timestamp": 1732861234567
}
```

### 3. 用户API测试

```bash
# 测试用户列表API（需要认证）
curl -s http://localhost:8082/api/v1/users/ \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" | jq .

# 预期输出: Go-Zervi标准分页响应格式
{
  "code": 0,
  "message": "获取用户列表成功",
  "data": {
    "list": [...],
    "total": 100,
    "pageNum": 1,
    "pageSize": 10,
    "pages": 10
  },
  "timestamp": 1732861234567
}
```

## 🚨 常见问题和解决方案

### 1. 数据库连接问题

**问题**: `PostgreSQL数据库未初始化`

**解决方案**:
```go
// 确保PostgreSQL配置正确
PostgreSQL: database.PostgreSQLConfig{
    Host: "localhost", // 不能为空
    Port: 5432,
    Username: "szjason72",
    Password: "",
    Database: "zervigo_mvp",
    SSLMode: "disable",
}
```

### 2. 认证中间件问题

**问题**: 返回旧格式响应

**解决方案**:
```go
// 确保使用ZerviAuthAdapter
zerviAuthAdapter := auth.NewZerviAuthAdapter(sqlDB, appConfig.Auth.JWTSecret)
authMiddleware := &ZerviAuthMiddlewareInterface{adapter: zerviAuthAdapter}
```

### 3. 模块路径问题

**问题**: `no required module provides package`

**解决方案**:
```go
// 使用正确的模块路径
import "github.com/szjason72/zervigo/shared/core/database"
```

## 📈 性能优化建议

### 1. 连接池配置

```yaml
database:
  postgres:
    max_open_conns: 100    # 根据并发量调整
    max_idle_conns: 10     # 根据空闲连接需求调整
    conn_max_lifetime: 3600 # 1小时，避免长时间连接
```

### 2. 缓存策略

```go
// 使用Redis缓存用户信息
func (adapter *ZerviAuthAdapter) getUserFromCache(userID int) (*UserInfo, error) {
    // 实现Redis缓存逻辑
}
```

### 3. 数据库索引

```sql
-- 为认证相关表添加索引
CREATE INDEX idx_auth_users_username ON zervigo_auth_users(username);
CREATE INDEX idx_auth_users_email ON zervigo_auth_users(email);
CREATE INDEX idx_auth_login_logs_user_id ON zervigo_auth_login_logs(user_id);
```

## 🔄 升级和维护

### 1. 版本兼容性

- **Go-Zervi Framework**: v0.1.0-alpha
- **JobFirst-Core**: 兼容版本
- **PostgreSQL**: 12+
- **Redis**: 6+

### 2. 配置迁移

```bash
# 从MySQL迁移到PostgreSQL
# 1. 更新配置文件
# 2. 执行数据迁移脚本
# 3. 验证数据完整性
# 4. 切换服务配置
```

### 3. 监控和维护

```yaml
# 添加监控配置
monitoring:
  prometheus:
    enabled: true
    port: 9090
    path: "/metrics"
  
  jaeger:
    enabled: true
    endpoint: "http://localhost:14268/api/traces"
    service_name: "zervigo-mvp"
```

## 📚 相关文档

- [JobFirst-Core 多数据库架构分析](./JOBFIRST_CORE_DATABASE_ARCHITECTURE_ANALYSIS.md)
- [Go-Zervi Framework 架构设计](./GO_ZERVI_FRAMEWORK_RELEASE_ANNOUNCEMENT.md)
- [API响应格式标准](./FRONTEND_API_FIELD_EXPECTATIONS.md)
- [数据库字段映射](./DATABASE_FRONTEND_FIELD_MAPPING.md)

---

**文档维护**: 本文档将随着集成功能的完善持续更新，确保实现指南的准确性和实用性。
