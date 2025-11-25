# User Service MySQL修复进度报告

**日期**: 2025-10-30  
**状态**: 🔧 进展中 - 已定位根本原因

---

## 🎯 目标

让user-service作为智能中央大脑的门户服务，连接到MySQL数据库。

---

## 📊 已完成的修复

### 1. 配置文件更新 ✅
- ✅ 已切换到MySQL配置
- ✅ PostgreSQL配置已注释

### 2. 代码修复 ✅
- ✅ TableName() 返回 "users" 
- ✅ 数据库选择逻辑优先MySQL
- ✅ 初始化失败不再终止服务
- ✅ 前端API路径已更新

### 3. 编译和重启 ✅
- ✅ 服务已重新编译
- ✅ 启动脚本已更新

---

## ⚠️ 根本问题

### 问题: 配置文件同时加载两个数据库

**根本原因**:
```bash
# configs/local.env 文件中
# MySQL配置 (启用)
MYSQL_HOST=localhost
MYSQL_DATABASE=jobfirst

# PostgreSQL配置 (禁用) - 但host等变量可能被其他服务设置
# 注释后，启动脚本仍然会导出旧的环境变量
```

**NewManager的逻辑**:
```go
// 只要Host不为空，就会初始化PostgreSQL
if config.PostgreSQL.Host != "" {
    postgresManager, err := NewPostgreSQLManager(config.PostgreSQL)
    manager.PostgreSQL = postgresManager  // ❌ 总是存在
}
```

**结果**:
- MySQL配置存在 → MySQLManager创建 ✅
- PostgreSQL配置存在 → PostgreSQLManager创建 ✅
- 代码优先检查PostgreSQL → 先返回PostgreSQLManager ❌

---

## 🔧 解决方案

### 方案1: 完全清除PostgreSQL环境变量（推荐）

修改配置文件，确保PostgreSQL变量完全不存在：

```env
# configs/local.env

# MySQL配置
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=
MYSQL_DATABASE=jobfirst

# PostgreSQL配置 - 完全删除或置为空
POSTGRESQL_HOST=
POSTGRESQL_PORT=
POSTGRESQL_USER=
POSTGRESQL_PASSWORD=
POSTGRESQL_DATABASE=
POSTGRESQL_SSL_MODE=
```

### 方案2: 修改数据库选择逻辑（更彻底）

在core.go中添加智能判断，真正优先MySQL：

```go
// 智能选择: 如果MySQL配置存在，优先MySQL
dbConfig := database.Config{
    MySQL: database.MySQLConfig{
        Host:        getEnvString("MYSQL_HOST", ""),
        // ...
    },
    PostgreSQL: database.PostgreSQLConfig{
        // 只有在MYSQL_HOST为空时才启用PostgreSQL
        Host:        getEnvString("POSTGRESQL_HOST", ""),
        // ...
    },
}

// 修改NewManager逻辑，智能选择数据库
// 如果MySQL存在，只初始化MySQL
// 如果MySQL不存在但PostgreSQL存在，初始化PostgreSQL
```

---

## 💡 立即执行的修复

### 步骤1: 完全清除PostgreSQL配置

```bash
# 修改配置文件
vim configs/local.env

# 将PostgreSQL行改为
POSTGRESQL_HOST=
POSTGRESQL_PORT=
POSTGRESQL_USER=
POSTGRESQL_PASSWORD=
POSTGRESQL_DATABASE=
POSTGRESQL_SSL_MODE=
```

### 步骤2: 重启服务

```bash
pkill -9 -f "user-service"
./scripts/start-local-services.sh
```

### 步骤3: 验证

```bash
# 检查数据库类型
curl -s http://localhost:8082/health | jq '.core_health.database'

# 测试API
TOKEN=$(curl -s -X POST http://localhost:8207/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' | jq -r '.data.accessToken')

curl -s -X GET "http://localhost:8082/api/v1/users/profile" \
  -H "Authorization: Bearer $TOKEN" | jq .
```

---

## 🎯 修复预期

修复后应该看到：
```json
{
  "core_health": {
    "database": {
      "mysql": {  // 不再是postgresql
        "database": "jobfirst",
        "host": "localhost",
        "port": 3306
      }
    }
  }
}
```

API响应应该返回用户信息：
```json
{
  "code": 0,
  "message": "User profile retrieved successfully",
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@jobfirst.com",
    "role": "super_admin"
  }
}
```

---

**作者**: Auto (AI Assistant)  
**日期**: 2025-10-30  
**状态**: 🔧 **已定位根本原因，等待执行修复**

