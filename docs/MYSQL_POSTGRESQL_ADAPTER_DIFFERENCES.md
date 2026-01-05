# MySQL 和 PostgreSQL 适配器主要区别分析

## 📊 核心区别总结

### 1. 🔐 **数据库迁移策略** (最大区别)

#### MySQL: 安全迁移模式（Safe Migration）
```go
// MySQL适配器 - 75-99行
func (mm *MySQLManager) Migrate(models ...interface{}) error {
    // 检查表是否存在，如果存在则跳过迁移
    for _, model := range models {
        stmt := &gorm.Statement{DB: mm.db}
        if err := stmt.Parse(model); err != nil {
            return fmt.Errorf("解析模型失败: %w", err)
        }

        // 检查表是否存在
        if mm.db.Migrator().HasTable(stmt.Schema.Table) {
            // 表已存在，只添加缺失的列，不修改现有约束
            if err := mm.db.Migrator().AutoMigrate(model); err != nil {
                // 如果迁移失败，记录错误但不中断
                fmt.Printf("警告: 表 %s 迁移失败: %v\n", stmt.Schema.Table, err)
            }
        } else {
            // 表不存在，正常创建
            if err := mm.db.Migrator().CreateTable(model); err != nil {
                return fmt.Errorf("创建表失败: %w", err)
            }
        }
    }
    return nil
}
```

**特点**:
- ✅ **安全模式**: 先检查表是否存在
- ✅ **容错处理**: 迁移失败只记录警告，不中断
- ✅ **增量更新**: 只添加缺失的列，不修改现有约束
- ✅ **外键约束**: 禁用外键约束以避免迁移冲突

#### PostgreSQL: 标准迁移模式（Standard Migration）
```go
// PostgreSQL适配器 - 105-108行
func (pm *PostgreSQLManager) Migrate(models ...interface{}) error {
    return pm.db.AutoMigrate(models...)
}
```

**特点**:
- ⚠️ **直接迁移**: 直接调用GORM的AutoMigrate
- ⚠️ **无安全检查**: 不检查表是否存在
- ⚠️ **可能破坏数据**: 如果表结构冲突可能导致错误

**这是最大的区别！MySQL采用了更安全的迁移策略。**

---

### 2. 🔌 **连接字符串构建方式**

#### MySQL: TCP连接格式
```go
// MySQL适配器 - 19-26行
dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
    config.Username,
    config.Password,
    config.Host,
    config.Port,
    config.Database,
    config.Charset,
)
```

**格式**: `username:password@tcp(host:port)/database?charset=xxx&parseTime=True&loc=Local`
- 使用TCP协议
- 需要指定charset（字符集）
- 需要设置时区参数

#### PostgreSQL: DSN格式
```go
// PostgreSQL适配器 - 43-49行
// 构建DSN，确保dbname参数正确传递（即使密码为空）
dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
    config.Host, config.Port, config.Username, config.Database, config.SSLMode)
if config.Password != "" {
    dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)
}
```

**格式**: `host=xxx port=xxx user=xxx password=xxx dbname=xxx sslmode=xxx`
- 使用参数格式（key=value）
- 处理空密码的特殊情况（不包含password参数）
- 需要设置SSL模式

**区别**: PostgreSQL对空密码有特殊处理，避免DSN解析错误

---

### 3. ⚙️ **GORM配置差异**

#### MySQL配置
```go
// MySQL适配器 - 28-31行
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger:                                   logger.Default.LogMode(config.LogLevel),
    DisableForeignKeyConstraintWhenMigrating: true,  // ⚠️ 禁用外键约束
})
```

**特殊配置**:
- ✅ **禁用外键约束**: `DisableForeignKeyConstraintWhenMigrating: true`
- ⚠️ **原因**: MySQL的外键约束在迁移时容易导致冲突

#### PostgreSQL配置
```go
// PostgreSQL适配器 - 52-54行
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(config.LogLevel),
})
```

**配置**:
- ⚠️ **标准配置**: 没有特殊配置
- ✅ **原因**: PostgreSQL对迁移更友好，不需要禁用外键

---

### 4. 🗄️ **配置结构差异**

#### MySQL配置
```go
type MySQLConfig struct {
    Host        string
    Port        int
    Username    string
    Password    string
    Database    string
    Charset     string          // ⚠️ MySQL特有：字符集
    MaxIdle     int
    MaxOpen     int
    MaxLifetime time.Duration
    LogLevel    logger.LogLevel
}
```

**特有字段**:
- `Charset`: MySQL需要指定字符集（如utf8mb4）

#### PostgreSQL配置
```go
type PostgreSQLConfig struct {
    Host        string
    Port        int
    Username    string
    Password    string
    Database    string
    SSLMode     string          // ⚠️ PostgreSQL特有：SSL模式
    MaxIdle     int
    MaxOpen     int
    MaxLifetime time.Duration
    LogLevel    logger.LogLevel
}
```

**特有字段**:
- `SSLMode`: PostgreSQL需要指定SSL模式（如disable, require, verify-full）
- 默认数据库名处理（空时设为`zervigo_mvp`）

---

### 5. 🔧 **API方法差异**

#### MySQL适配器方法
```go
GetDB() *gorm.DB           // 获取GORM实例
Close() error              // 关闭连接
Ping() error               // 连接测试
Migrate(...) error         // 安全迁移
Transaction(...) error     // 事务
Create(...) error          // CRUD操作
First(...) error
Find(...) error
Update(...) error
Delete(...) error
Raw(...) *gorm.DB          // 原生SQL
Exec(...) error
Health() map[string]interface{}  // 健康检查
```

#### PostgreSQL适配器方法
```go
GetDB() *gorm.DB           // 获取GORM实例
GetSQLDB() (*sql.DB, error) // ⚠️ 新增：获取原生SQL连接
Close() error              // 关闭连接
Ping() error               // 连接测试
Migrate(...) error         // 标准迁移
Transaction(...) error     // 事务
Create(...) error          // CRUD操作
First(...) error
Find(...) error
Update(...) error
Delete(...) error
Raw(...) *gorm.DB          // 原生SQL
Exec(...) error
Health() map[string]interface{}  // 健康检查
CreateVectorExtension() error    // ⚠️ 新增：向量扩展
CreateVectorIndex(...) error     // ⚠️ 新增：向量索引
VectorSearch(...) (*gorm.DB, error) // ⚠️ 新增：向量搜索
```

**PostgreSQL特有功能**:
- ✅ **GetSQLDB()**: 获取原生SQL连接（用于需要原生SQL的场景）
- ✅ **向量扩展支持**: CreateVectorExtension, CreateVectorIndex, VectorSearch
- ✅ **AI功能**: 支持向量搜索（适合AI服务）

---

### 6. 📝 **初始化时的特殊处理**

#### MySQL: 无特殊处理
- 直接使用配置参数连接

#### PostgreSQL: 有默认值处理
```go
// PostgreSQL适配器 - 35-39行
// 设置默认数据库名（如果为空）
if config.Database == "" {
    fmt.Printf("DEBUG: PostgreSQL Database为空，设置为默认值: zervigo_mvp\n")
    config.Database = "zervigo_mvp"
}
```

**特殊处理**:
- ✅ 数据库名为空时自动设置为`zervigo_mvp`
- ✅ 有DEBUG日志输出（便于调试）

---

### 7. 🚨 **错误处理差异**

#### MySQL: 详细错误处理
```go
// Migrate方法中的错误处理
if err := mm.db.Migrator().AutoMigrate(model); err != nil {
    // 如果迁移失败，记录错误但不中断
    fmt.Printf("警告: 表 %s 迁移失败: %v\n", stmt.Schema.Table, err)
}
```

#### PostgreSQL: 标准错误处理
```go
// Migrate方法直接返回错误
return pm.db.AutoMigrate(models...)
```

**差异**: MySQL迁移失败时只记录警告，PostgreSQL会直接返回错误

---

## 📊 对比总结表

| 特性 | MySQL适配器 | PostgreSQL适配器 |
|------|------------|------------------|
| **迁移策略** | ✅ 安全模式（检查表存在） | ⚠️ 标准模式（直接迁移） |
| **外键约束** | ✅ 禁用外键约束 | ⚠️ 启用外键约束 |
| **连接格式** | TCP格式 (`@tcp()`) | DSN格式 (`host=`) |
| **字符集** | ✅ 需要指定Charset | ❌ 不需要 |
| **SSL模式** | ❌ 不需要 | ✅ 需要指定SSLMode |
| **空密码处理** | ❌ 不支持空密码 | ✅ 特殊处理空密码 |
| **默认值处理** | ❌ 无 | ✅ 数据库名默认值 |
| **向量搜索** | ❌ 不支持 | ✅ 支持（CreateVectorExtension等） |
| **原生SQL** | ✅ 仅GORM | ✅ GORM + 原生SQL |
| **错误处理** | ✅ 容错（警告不中断） | ⚠️ 直接返回错误 |
| **DEBUG日志** | ❌ 无 | ✅ 有 |

---

## 🎯 最大区别总结

### 🔴 **第一最大区别：迁移策略**

**MySQL**: 
- ✅ **安全迁移模式**
- ✅ 检查表是否存在
- ✅ 容错处理（失败只记录警告）
- ✅ 只添加缺失列，不修改约束

**PostgreSQL**: 
- ⚠️ **标准迁移模式**
- ⚠️ 直接执行AutoMigrate
- ⚠️ 无安全检查
- ⚠️ 可能破坏现有数据

### 🟡 **第二最大区别：PostgreSQL特有功能**

**PostgreSQL独有**:
- ✅ 向量扩展支持（AI功能）
- ✅ 原生SQL连接（GetSQLDB）
- ✅ 默认数据库名处理
- ✅ DEBUG日志输出

### 🟢 **第三最大区别：配置差异**

**MySQL配置**:
- 需要Charset（字符集）
- 禁用外键约束
- TCP连接格式

**PostgreSQL配置**:
- 需要SSLMode（SSL模式）
- 启用外键约束
- DSN连接格式
- 空密码特殊处理

---

## 💡 建议

### 1. 统一迁移策略
建议PostgreSQL适配器也采用MySQL的安全迁移模式，避免破坏现有数据：

```go
func (pm *PostgreSQLManager) Migrate(models ...interface{}) error {
    // 建议添加安全检查
    for _, model := range models {
        stmt := &gorm.Statement{DB: pm.db}
        if err := stmt.Parse(model); err != nil {
            return fmt.Errorf("解析模型失败: %w", err)
        }
        
        if pm.db.Migrator().HasTable(stmt.Schema.Table) {
            // 表存在，安全迁移
            if err := pm.db.Migrator().AutoMigrate(model); err != nil {
                fmt.Printf("警告: 表 %s 迁移失败: %v\n", stmt.Schema.Table, err)
            }
        } else {
            // 表不存在，创建
            if err := pm.db.Migrator().CreateTable(model); err != nil {
                return fmt.Errorf("创建表失败: %w", err)
            }
        }
    }
    return nil
}
```

### 2. 统一错误处理
建议PostgreSQL也采用容错处理，避免迁移失败导致服务无法启动。

### 3. 统一调试日志
建议MySQL也添加DEBUG日志，便于调试和问题排查。

---

**报告生成时间**: 2025-01-29  
**检查范围**: `shared/core/database/mysql.go` vs `shared/core/database/postgresql.go`

