# Central Brain 数据库连接检查机制设计

## 📋 设计目标

**目标**: Central Brain启动后，自动检查并验证数据库连接，确保能够为后续服务提供核心能力。

**核心需求**:
1. ✅ 自动识别数据库类型（PostgreSQL、MySQL、Redis等）
2. ✅ 从环境变量读取数据库配置
3. ✅ 安全高效地握手验证
4. ✅ 提供清晰的错误信息（如果握手失败）

---

## 🎯 架构设计

### **数据库连接检查流程**

```
Central Brain启动
  ↓
1. 加载环境变量配置
  ↓
2. 识别数据库类型和配置
  ├── PostgreSQL (通过POSTGRESQL_*环境变量)
  ├── MySQL (通过MYSQL_*环境变量)
  ├── Redis (通过REDIS_*环境变量)
  └── Neo4j (通过NEO4J_*环境变量)
  ↓
3. 数据库连接检查
  ├── 检查连接配置完整性
  ├── 尝试建立连接
  ├── 执行Ping测试
  └── 记录连接信息
  ↓
4. 检查结果处理
  ├── 成功 → 继续启动
  ├── 失败 → 提供详细错误信息
  └── 警告 → 记录但继续启动（可选）
```

---

## 📊 数据库配置识别策略

### **策略1: 环境变量模式识别**

**PostgreSQL识别**:
```bash
# 识别标志
POSTGRESQL_HOST 存在
POSTGRESQL_PORT 存在
POSTGRESQL_USER 存在
POSTGRESQL_DATABASE 存在

# 或通过DATABASE_URL
DATABASE_URL=postgres://user:pass@host:port/db
```

**MySQL识别**:
```bash
# 识别标志
MYSQL_HOST 存在
MYSQL_PORT 存在
MYSQL_USER 存在
MYSQL_DATABASE 存在

# 或通过DATABASE_URL
DATABASE_URL=mysql://user:pass@host:port/db
```

**Redis识别**:
```bash
# 识别标志
REDIS_HOST 存在
REDIS_PORT 存在

# 或通过REDIS_URL
REDIS_URL=redis://:password@host:port/db
```

**Neo4j识别**:
```bash
# 识别标志
NEO4J_URI 存在
NEO4J_USER 存在
NEO4J_PASSWORD 存在
```

---

### **策略2: 数据库URL解析**

**统一格式**:
```
postgres://user:password@host:port/database?sslmode=disable
mysql://user:password@host:port/database?charset=utf8mb4
redis://:password@host:port/db
neo4j://user:password@host:port
```

**解析逻辑**:
1. 提取协议（postgres/mysql/redis/neo4j）
2. 提取主机、端口、用户名、密码、数据库名
3. 根据协议识别数据库类型

---

## 🏗️ 实现方案

### **步骤1: 增强Config结构**

在`shared/core/shared/config.go`中添加数据库配置：

```go
type Config struct {
    // ... 现有配置
    
    // 数据库配置
    Database struct {
        // 数据库类型（自动识别）
        Type string // "postgresql", "mysql", "redis", "neo4j", "auto"
        
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
        
        // Neo4j配置
        Neo4j struct {
            URI      string
            User     string
            Password string
            Enabled  bool
        }
        
        // 统一URL（可选）
        URL string // DATABASE_URL或REDIS_URL
    }
    
    // 数据库检查配置
    DatabaseCheck struct {
        Enabled     bool   // 是否启用数据库检查
        Required    bool   // 是否必需（失败时阻止启动）
        Timeout     int    // 连接超时（秒）
        RetryCount  int    // 重试次数
        RetryDelay  int    // 重试延迟（秒）
    }
}
```

---

### **步骤2: 实现数据库类型识别**

```go
// identifyDatabaseType 识别数据库类型
func identifyDatabaseType(config *Config) string {
    // 1. 检查统一URL
    if config.Database.URL != "" {
        if strings.HasPrefix(config.Database.URL, "postgres://") ||
           strings.HasPrefix(config.Database.URL, "postgresql://") {
            return "postgresql"
        }
        if strings.HasPrefix(config.Database.URL, "mysql://") {
            return "mysql"
        }
        if strings.HasPrefix(config.Database.URL, "redis://") {
            return "redis"
        }
        if strings.HasPrefix(config.Database.URL, "neo4j://") {
            return "neo4j"
        }
    }
    
    // 2. 检查PostgreSQL配置
    if config.Database.PostgreSQL.Host != "" {
        return "postgresql"
    }
    
    // 3. 检查MySQL配置
    if config.Database.MySQL.Host != "" {
        return "mysql"
    }
    
    // 4. 检查Redis配置
    if config.Database.Redis.Host != "" {
        return "redis"
    }
    
    // 5. 检查Neo4j配置
    if config.Database.Neo4j.URI != "" {
        return "neo4j"
    }
    
    return "none" // 未配置
}
```

---

### **步骤3: 实现数据库连接检查**

```go
// DatabaseChecker 数据库检查器
type DatabaseChecker struct {
    config *Config
    logger *log.Logger
}

// CheckDatabase 检查数据库连接
func (dc *DatabaseChecker) CheckDatabase() (*DatabaseCheckResult, error) {
    result := &DatabaseCheckResult{
        Type:     identifyDatabaseType(dc.config),
        Status:   "unknown",
        Message:  "",
        Config:  map[string]interface{}{},
        Errors:   []string{},
        Warnings: []string{},
    }
    
    if result.Type == "none" {
        result.Status = "not_configured"
        result.Message = "未配置数据库"
        result.Warnings = append(result.Warnings, "数据库未配置，某些功能可能不可用")
        return result, nil
    }
    
    // 根据类型检查数据库
    switch result.Type {
    case "postgresql":
        return dc.checkPostgreSQL(result)
    case "mysql":
        return dc.checkMySQL(result)
    case "redis":
        return dc.checkRedis(result)
    case "neo4j":
        return dc.checkNeo4j(result)
    default:
        result.Status = "unsupported"
        result.Message = fmt.Sprintf("不支持的数据库类型: %s", result.Type)
        result.Errors = append(result.Errors, result.Message)
        return result, fmt.Errorf(result.Message)
    }
}

// checkPostgreSQL 检查PostgreSQL连接
func (dc *DatabaseChecker) checkPostgreSQL(result *DatabaseCheckResult) (*DatabaseCheckResult, error) {
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
        return result, fmt.Errorf(result.Message)
    }
    
    if pgConfig.Port == 0 {
        pgConfig.Port = 5432 // 默认端口
        result.Warnings = append(result.Warnings, "PostgreSQL端口未设置，使用默认值: 5432")
    }
    
    if pgConfig.Database == "" {
        result.Status = "invalid_config"
        result.Message = "PostgreSQL配置不完整: DATABASE未设置"
        result.Errors = append(result.Errors, result.Message)
        return result, fmt.Errorf(result.Message)
    }
    
    // 尝试连接
    dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
        pgConfig.Host, pgConfig.Port, pgConfig.User, pgConfig.Database, pgConfig.SSLMode)
    if pgConfig.Password != "" {
        dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
            pgConfig.Host, pgConfig.Port, pgConfig.User, pgConfig.Password, pgConfig.Database, pgConfig.SSLMode)
    }
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        result.Status = "connection_failed"
        result.Message = fmt.Sprintf("PostgreSQL连接失败: %v", err)
        result.Errors = append(result.Errors, result.Message)
        return result, fmt.Errorf(result.Message)
    }
    defer db.Close()
    
    // 设置超时
    ctx, cancel := context.WithTimeout(context.Background(), 
        time.Duration(dc.config.DatabaseCheck.Timeout)*time.Second)
    defer cancel()
    
    // Ping测试
    if err := db.PingContext(ctx); err != nil {
        result.Status = "ping_failed"
        result.Message = fmt.Sprintf("PostgreSQL Ping失败: %v", err)
        result.Errors = append(result.Errors, result.Message)
        return result, fmt.Errorf(result.Message)
    }
    
    // 获取数据库版本
    var version string
    if err := db.QueryRowContext(ctx, "SELECT version()").Scan(&version); err == nil {
        result.Message = fmt.Sprintf("PostgreSQL连接成功: %s", version)
    } else {
        result.Message = "PostgreSQL连接成功"
    }
    
    result.Status = "connected"
    return result, nil
}

// checkRedis 检查Redis连接
func (dc *DatabaseChecker) checkRedis(result *DatabaseCheckResult) (*DatabaseCheckResult, error) {
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
        return result, fmt.Errorf(result.Message)
    }
    
    if redisConfig.Port == 0 {
        redisConfig.Port = 6379 // 默认端口
        result.Warnings = append(result.Warnings, "Redis端口未设置，使用默认值: 6379")
    }
    
    // 尝试连接
    addr := fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: redisConfig.Password,
        DB:       redisConfig.DB,
    })
    defer client.Close()
    
    // 设置超时
    ctx, cancel := context.WithTimeout(context.Background(),
        time.Duration(dc.config.DatabaseCheck.Timeout)*time.Second)
    defer cancel()
    
    // Ping测试
    if err := client.Ping(ctx).Err(); err != nil {
        result.Status = "ping_failed"
        result.Message = fmt.Sprintf("Redis Ping失败: %v", err)
        result.Errors = append(result.Errors, result.Message)
        return result, fmt.Errorf(result.Message)
    }
    
    result.Status = "connected"
    result.Message = "Redis连接成功"
    return result, nil
}

// DatabaseCheckResult 数据库检查结果
type DatabaseCheckResult struct {
    Type     string                 `json:"type"`
    Status  string                 `json:"status"` // "connected", "failed", "not_configured", "invalid_config"
    Message  string                 `json:"message"`
    Config   map[string]interface{} `json:"config"`
    Errors   []string               `json:"errors"`
    Warnings []string               `json:"warnings"`
    Duration time.Duration          `json:"duration"`
}
```

---

### **步骤4: 集成到Central Brain启动流程**

```go
// NewCentralBrain 创建中央大脑服务
func NewCentralBrain(config *shared.Config) *CentralBrain {
    // ... 现有代码
    
    // 数据库连接检查
    if config.DatabaseCheck.Enabled {
        checker := &DatabaseChecker{
            config: config,
            logger: log.New(os.Stdout, "[DB-CHECK] ", log.LstdFlags),
        }
        
        result, err := checker.CheckDatabase()
        if err != nil {
            if config.DatabaseCheck.Required {
                log.Fatalf("❌ 数据库连接检查失败（必需）: %v\n%s", err, formatDatabaseError(result))
            } else {
                log.Printf("⚠️ 数据库连接检查失败（可选）: %v\n%s", err, formatDatabaseWarning(result))
            }
        } else {
            log.Printf("✅ 数据库连接检查成功: %s", result.Message)
            if len(result.Warnings) > 0 {
                for _, warning := range result.Warnings {
                    log.Printf("   ⚠️ %s", warning)
                }
            }
        }
    }
    
    return cb
}

// formatDatabaseError 格式化数据库错误信息
func formatDatabaseError(result *DatabaseCheckResult) string {
    var buf strings.Builder
    buf.WriteString("\n")
    buf.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
    buf.WriteString("数据库连接检查失败\n")
    buf.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
    buf.WriteString(fmt.Sprintf("数据库类型: %s\n", result.Type))
    buf.WriteString(fmt.Sprintf("状态: %s\n", result.Status))
    buf.WriteString(fmt.Sprintf("错误信息: %s\n", result.Message))
    
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
```

---

## 📋 环境变量配置示例

### **configs/local.env**

```bash
# 数据库检查配置
DATABASE_CHECK_ENABLED=true
DATABASE_CHECK_REQUIRED=false  # false=可选，true=必需
DATABASE_CHECK_TIMEOUT=5
DATABASE_CHECK_RETRY_COUNT=3
DATABASE_CHECK_RETRY_DELAY=2

# PostgreSQL配置
POSTGRESQL_HOST=localhost
POSTGRESQL_PORT=5432
POSTGRESQL_USER=szjason72
POSTGRESQL_PASSWORD=
POSTGRESQL_DATABASE=zervigo_mvp
POSTGRESQL_SSL_MODE=disable

# 或使用统一URL
# DATABASE_URL=postgres://szjason72@localhost:5432/zervigo_mvp?sslmode=disable

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# 或使用统一URL
# REDIS_URL=redis://localhost:6379/0
```

---

## 🎯 错误信息设计

### **错误信息格式**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
数据库连接检查失败
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
数据库类型: postgresql
状态: connection_failed
错误信息: PostgreSQL连接失败: dial tcp 127.0.0.1:5432: connect: connection refused

配置信息:
  host: localhost
  port: 5432
  user: szjason72
  database: zervigo_mvp
  ssl_mode: disable

错误详情:
  1. PostgreSQL连接失败: dial tcp 127.0.0.1:5432: connect: connection refused

解决方案:
  1. 检查PostgreSQL服务是否运行
  2. 检查环境变量: POSTGRESQL_HOST, POSTGRESQL_PORT, POSTGRESQL_USER, POSTGRESQL_DATABASE
  3. 检查网络连接和防火墙设置
  4. 验证用户名和密码是否正确

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## ✅ 实施建议

### **阶段1: 基础实现** (优先级: 🔥 高)

1. ✅ 增强Config结构（添加数据库配置）
2. ✅ 实现数据库类型识别
3. ✅ 实现PostgreSQL和Redis连接检查
4. ✅ 集成到Central Brain启动流程
5. ✅ 提供清晰的错误信息

### **阶段2: 增强功能** (优先级: 🟡 中)

1. ⚠️ 支持MySQL和Neo4j
2. ⚠️ 实现重试机制
3. ⚠️ 实现连接池预热
4. ⚠️ 健康检查端点（暴露数据库状态）

### **阶段3: 高级功能** (优先级: 🟢 低)

1. ⚠️ 数据库连接监控
2. ⚠️ 自动故障恢复
3. ⚠️ 连接性能指标收集

---

## 🎯 总结

**核心设计原则**:
1. ✅ **自动识别**: 根据环境变量自动识别数据库类型
2. ✅ **安全高效**: 使用超时和连接池，避免阻塞
3. ✅ **清晰错误**: 提供详细的错误信息和解决方案
4. ✅ **灵活配置**: 支持必需/可选模式

**下一步**: 开始实施这个方案！

