# Docker数据库类型清单与local.env配置指南

## 📋 本地Docker数据库清单

根据 `docker/docker-compose.local.yml` 和 `docker/docker-compose.yml` 配置，以下是可用的数据库类型：

### ✅ 核心数据库（已配置）

| 数据库类型 | Docker容器名 | 镜像 | 端口 | 用途 | 状态 |
|-----------|------------|------|------|------|------|
| **MySQL** | `local-mysql` | `mysql:8.0` | 3306 | 核心关系数据库 | ✅ |
| **PostgreSQL** | `local-postgres` | `postgres:14` | 5432 | 向量存储 | ✅ |
| **PostgreSQL 16** | `zervigo-postgres` | `postgres:16-alpine` | 5432 | MVP数据库 | ✅（已安装） |
| **Redis** | `local-redis` | `redis:6-alpine` | 6379 | 缓存 | ✅ |
| **MongoDB** | `local-mongodb` | `mongo:6.0` | 27017 | 文档存储 | ✅ |
| **Neo4j** | `local-neo4j` | `neo4j:latest` | 7474(HTTP)<br>7687(Bolt) | 图数据库 | ✅ |

### 🔧 扩展数据库（可选，需profile）

| 数据库类型 | Docker容器名 | 镜像 | 端口 | 用途 | 启动方式 |
|-----------|------------|------|------|------|----------|
| **Elasticsearch** | `local-elasticsearch` | `elasticsearch:8.8.0` | 9200 | 全文搜索 | `--profile full` |
| **Weaviate** | `local-weaviate` | `semitechnologies/weaviate:latest` | 8080 | 向量搜索 | `--profile full` |

### 📦 MVP数据库（Zervigo项目）

| 数据库类型 | Docker容器名 | 镜像 | 端口 | 用途 | 状态 |
|-----------|------------|------|------|------|------|
| **PostgreSQL 15** | `zervigo-postgres-mvp` | `postgres:15-alpine` | 5432 | MVP主数据库 | ✅ |
| **PostgreSQL 16** | `zervigo-postgres` | `postgres:16-alpine` | 5432 | MVP数据库（新版本） | ✅（已安装） |
| **Redis** | `zervigo-redis-mvp` | `redis:7-alpine` | 6379 | MVP缓存 | ✅ |

**注意**: 您的Docker环境中已经安装了PostgreSQL 16版本！可以使用 `postgres:16-alpine` 镜像。

---

## 🔧 local.env 配置参数指南

### 1️⃣ MySQL 配置

```bash
# MySQL数据库配置
MYSQL_HOST=localhost              # 数据库主机（Docker: localhost 或 local-mysql）
MYSQL_PORT=3306                  # 数据库端口
MYSQL_USER=root                  # 数据库用户名（Docker默认: root）
MYSQL_PASSWORD=local_dev_password # 数据库密码（Docker默认: local_dev_password）
MYSQL_DATABASE=looma             # 数据库名称（自定义）
```

**Docker连接方式**:
- **主机名**: `localhost` (宿主机访问) 或 `local-mysql` (Docker网络内)
- **默认密码**: `local_dev_password`
- **默认数据库**: `jobfirst_basic`

---

### 2️⃣ PostgreSQL 配置

```bash
# PostgreSQL数据库配置
POSTGRESQL_HOST=localhost         # 数据库主机（Docker: localhost 或 local-postgres）
POSTGRESQL_PORT=5432             # 数据库端口
POSTGRESQL_USER=postgres         # 数据库用户名（Docker默认: postgres）
POSTGRESQL_PASSWORD=local_dev_password # 数据库密码（Docker默认: local_dev_password）
POSTGRESQL_DATABASE=jobfirst_vector # 数据库名称（Docker默认: jobfirst_vector）
POSTGRESQL_SSL_MODE=disable      # SSL模式（开发环境: disable）
```

**Docker连接方式**:
- **主机名**: `localhost` (宿主机访问) 或 `local-postgres` (Docker网络内)
- **默认密码**: `local_dev_password`
- **默认数据库**: `jobfirst_vector` (本地) 或 `zervigo_mvp` (MVP)

**可用版本**:
- ✅ `postgres:14` (docker-compose.local.yml中使用)
- ✅ `postgres:15-alpine` (docker-compose.yml中使用)
- ✅ `postgres:16-alpine` (**已安装，可使用**)

---

### 3️⃣ Redis 配置

```bash
# Redis缓存配置
REDIS_HOST=localhost             # Redis主机（Docker: localhost 或 local-redis）
REDIS_PORT=6379                  # Redis端口
REDIS_PASSWORD=local_dev_password # Redis密码（Docker默认: local_dev_password）
REDIS_DB=0                       # Redis数据库编号（默认: 0）
```

**Docker连接方式**:
- **主机名**: `localhost` (宿主机访问) 或 `local-redis` (Docker网络内)
- **默认密码**: `local_dev_password`
- **无密码模式**: 如果未设置密码，留空 `REDIS_PASSWORD=`

---

### 4️⃣ MongoDB 配置（可选）

```bash
# MongoDB文档数据库配置
MONGODB_HOST=localhost           # MongoDB主机（Docker: localhost 或 local-mongodb）
MONGODB_PORT=27017               # MongoDB端口
MONGODB_USER=admin               # MongoDB用户名（Docker默认: admin）
MONGODB_PASSWORD=local_dev_password # MongoDB密码（Docker默认: local_dev_password）
MONGODB_DATABASE=jobfirst        # 数据库名称（自定义）
```

**Docker连接方式**:
- **主机名**: `localhost` (宿主机访问) 或 `local-mongodb` (Docker网络内)
- **默认用户名**: `admin`
- **默认密码**: `local_dev_password`

---

### 5️⃣ Neo4j 配置（可选）

```bash
# Neo4j图数据库配置
NEO4J_HOST=localhost             # Neo4j主机（Docker: localhost 或 local-neo4j）
NEO4J_PORT=7687                  # Neo4j Bolt端口（推荐）
NEO4J_HTTP_PORT=7474              # Neo4j HTTP端口
NEO4J_USER=neo4j                  # Neo4j用户名（默认: neo4j）
NEO4J_PASSWORD=local_dev_password # Neo4j密码（Docker默认: local_dev_password）
```

**Docker连接方式**:
- **主机名**: `localhost` (宿主机访问) 或 `local-neo4j` (Docker网络内)
- **Bolt端口**: `7687` (推荐用于应用连接)
- **HTTP端口**: `7474` (用于Web界面)
- **默认密码**: `local_dev_password`

---

### 6️⃣ Elasticsearch 配置（可选，需profile）

```bash
# Elasticsearch全文搜索引擎配置
ELASTICSEARCH_HOST=localhost      # Elasticsearch主机
ELASTICSEARCH_PORT=9200          # Elasticsearch端口
ELASTICSEARCH_USER=              # 用户名（默认无认证）
ELASTICSEARCH_PASSWORD=          # 密码（默认无认证）
```

**启动方式**:
```bash
docker-compose -f docker/docker-compose.local.yml --profile full up -d elasticsearch
```

---

### 7️⃣ Weaviate 配置（可选，需profile）

```bash
# Weaviate向量搜索引擎配置
WEAVIATE_HOST=localhost           # Weaviate主机
WEAVIATE_PORT=8080               # Weaviate端口
```

**启动方式**:
```bash
docker-compose -f docker/docker-compose.local.yml --profile full up -d weaviate
```

---

## 📝 完整配置示例

### 方式1: 使用MySQL（当前配置）

```bash
# configs/local.env

# MySQL数据库配置（当前使用）
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=local_dev_password
MYSQL_DATABASE=looma

# PostgreSQL配置（已注释）
# POSTGRESQL_HOST=localhost
# POSTGRESQL_PORT=5432
# POSTGRESQL_USER=postgres
# POSTGRESQL_PASSWORD=local_dev_password
# POSTGRESQL_DATABASE=jobfirst_vector
# POSTGRESQL_SSL_MODE=disable

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=local_dev_password
REDIS_DB=0
```

---

### 方式2: 使用PostgreSQL

```bash
# configs/local.env

# PostgreSQL数据库配置（当前使用）
POSTGRESQL_HOST=localhost
POSTGRESQL_PORT=5432
POSTGRESQL_USER=postgres
POSTGRESQL_PASSWORD=local_dev_password
POSTGRESQL_DATABASE=zervigo_mvp
POSTGRESQL_SSL_MODE=disable

# MySQL配置（已注释）
# MYSQL_HOST=localhost
# MYSQL_PORT=3306
# MYSQL_USER=root
# MYSQL_PASSWORD=local_dev_password
# MYSQL_DATABASE=looma

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=local_dev_password
REDIS_DB=0
```

---

### 方式3: 使用Docker网络内服务名（Docker Compose环境）

```bash
# configs/local.env (Docker Compose环境)

# MySQL数据库配置（Docker网络内）
MYSQL_HOST=local-mysql           # Docker服务名
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=local_dev_password
MYSQL_DATABASE=looma

# PostgreSQL配置（Docker网络内）
# POSTGRESQL_HOST=local-postgres  # Docker服务名
# POSTGRESQL_PORT=5432
# POSTGRESQL_USER=postgres
# POSTGRESQL_PASSWORD=local_dev_password
# POSTGRESQL_DATABASE=jobfirst_vector
# POSTGRESQL_SSL_MODE=disable

# Redis配置（Docker网络内）
REDIS_HOST=local-redis           # Docker服务名
REDIS_PORT=6379
REDIS_PASSWORD=local_dev_password
REDIS_DB=0
```

---

## 🚀 启动Docker数据库

### 启动所有数据库（本地开发环境）

```bash
cd docker
docker-compose -f docker-compose.local.yml up -d
```

### 启动特定数据库

```bash
# 只启动MySQL
docker-compose -f docker-compose.local.yml up -d mysql

# 只启动PostgreSQL
docker-compose -f docker-compose.local.yml up -d postgres

# 只启动Redis
docker-compose -f docker-compose.local.yml up -d redis

# 启动MySQL + PostgreSQL + Redis
docker-compose -f docker-compose.local.yml up -d mysql postgres redis
```

### 启动MVP数据库

```bash
cd docker
docker-compose -f docker-compose-postgres.yml up -d
```

### 使用PostgreSQL 16版本

**您的Docker环境已安装PostgreSQL 16**！可以修改配置文件使用16版本：

**方式1: 修改docker-compose.local.yml**
```yaml
postgres:
  image: postgres:16-alpine  # 改为16版本
  container_name: local-postgres
  # ... 其他配置
```

**方式2: 直接启动现有的PostgreSQL 16容器**
```bash
# 启动现有的zervigo-postgres容器（PostgreSQL 16）
docker start zervigo-postgres

# 或创建新的PostgreSQL 16容器
docker run -d \
  --name local-postgres-16 \
  -e POSTGRES_PASSWORD=local_dev_password \
  -e POSTGRES_DB=jobfirst_vector \
  -p 5432:5432 \
  postgres:16-alpine
```

**验证PostgreSQL 16版本**:
```bash
docker exec zervigo-postgres psql -U postgres -c "SELECT version();"
```

### 启动扩展数据库（带profile）

```bash
# 启动Elasticsearch
docker-compose -f docker-compose.local.yml --profile full up -d elasticsearch

# 启动Weaviate
docker-compose -f docker-compose.local.yml --profile full up -d weaviate
```

---

## 🔍 数据库连接检查配置

```bash
# 数据库检查配置
DATABASE_CHECK_ENABLED=true      # 是否启用数据库检查（true/false）
DATABASE_CHECK_REQUIRED=false    # 是否必需（false=可选，true=必需，失败时阻止启动）
DATABASE_CHECK_TIMEOUT=5         # 连接超时（秒）
DATABASE_CHECK_RETRY_COUNT=3     # 重试次数
DATABASE_CHECK_RETRY_DELAY=2     # 重试延迟（秒）
```

---

## 📊 数据库类型识别优先级

Central Brain在启动时会按以下优先级识别数据库类型：

```
1. DATABASE_URL (统一URL，最高优先级)
   ↓
2. MySQL配置 (MYSQL_HOST) ← 优先于PostgreSQL
   ↓
3. PostgreSQL配置 (POSTGRESQL_HOST)
   ↓
4. Redis配置 (REDIS_HOST)
```

**说明**: 如果同时配置了MySQL和PostgreSQL，会优先使用MySQL。

---

## 🔐 Docker默认密码总结

| 数据库 | 默认用户名 | 默认密码 | 默认数据库 |
|--------|----------|---------|-----------|
| MySQL | `root` | `local_dev_password` | `jobfirst_basic` |
| PostgreSQL | `postgres` | `local_dev_password` | `jobfirst_vector` (本地)<br>`zervigo_mvp` (MVP) |
| Redis | - | `local_dev_password` | `0` |
| MongoDB | `admin` | `local_dev_password` | - |
| Neo4j | `neo4j` | `local_dev_password` | - |

---

## ✅ 当前配置推荐

基于您的 `configs/local.env` 文件，当前使用：

- ✅ **MySQL** - 本地数据库 `looma`
- ✅ **Redis** - 本地缓存
- ⚠️ **PostgreSQL** - 已注释（可切换）

**建议**: 如需切换到PostgreSQL，请：
1. 注释MySQL配置
2. 取消注释PostgreSQL配置
3. 确保Docker容器运行：`docker-compose -f docker/docker-compose.local.yml up -d postgres`

---

## 📚 相关文档

- Docker Compose配置: `docker/docker-compose.local.yml`
- MVP配置: `docker/docker-compose-postgres.yml`
- 数据库检查实现: `shared/core/shared/database_checker.go`

---

**文档更新时间**: 2025-01-29  
**适用版本**: Central Brain v1.0.0

