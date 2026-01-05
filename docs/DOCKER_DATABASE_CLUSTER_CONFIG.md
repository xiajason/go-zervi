# Zervigo Docker 多数据库集群配置文档

## 概览

本文档详细记录了 Zervigo 项目中 Docker 多数据库集群的配置信息，包括数据库类型、端口映射、连接凭证、容器名称等。

**最后更新**: 2025-10-30  
**Docker 环境**: macOS (Darwin 24.6.0)

---

## 当前运行的容器

### 1. PostgreSQL 16 数据库 ✅

| 配置项 | 值 |
|--------|-----|
| **容器名称** | `zervigo-postgres` |
| **镜像版本** | `postgres:16-alpine` |
| **状态** | 🟢 Up 9 hours (healthy) |
| **主机端口** | `15432` |
| **容器端口** | `5432` |
| **数据库名称** | `zervigo_unified` |
| **用户名** | `postgres` |
| **密码** | `postgres` |
| **连接字符串** | `postgresql://postgres:postgres@localhost:15432/zervigo_unified` |

**访问方式**:
```bash
# 命令行连接
psql -h localhost -p 15432 -U postgres -d zervigo_unified

# 应用连接
POSTGRESQL_HOST=localhost
POSTGRESQL_PORT=15432
POSTGRESQL_USER=postgres
POSTGRESQL_PASSWORD=postgres
POSTGRESQL_DATABASE=zervigo_unified
```

---

## Docker Compose 配置文件

项目包含多个 Docker Compose 配置文件，针对不同的使用场景：

### 1. `docker-compose.yml` (MVP 生产配置)

**用途**: Zervigo MVP 微服务架构的主配置

#### PostgreSQL 配置
| 配置项 | 值 |
|--------|-----|
| **容器名称** | `zervigo-postgres-mvp` |
| **镜像** | `postgres:15-alpine` |
| **端口映射** | `5432:5432` |
| **数据库** | `zervigo_mvp` |
| **用户名** | `postgres` |
| **密码** | `dev_password` |
| **时区** | `Asia/Shanghai` |
| **健康检查** | `pg_isready -U postgres -d zervigo_mvp` |

#### Redis 配置
| 配置项 | 值 |
|--------|-----|
| **容器名称** | `zervigo-redis-mvp` |
| **镜像** | `redis:7-alpine` |
| **端口映射** | `6379:6379` |
| **持久化** | `--appendonly yes` |
| **健康检查** | `redis-cli ping` |

#### Consul 配置
| 配置项 | 值 |
|--------|-----|
| **容器名称** | `zervigo-consul-mvp` |
| **镜像** | `consul:1.16` |
| **端口映射** | `8500:8500` |
| **UI** | 启用 |

---

### 2. `docker-compose.local.yml` (本地开发配置)

**用途**: 本地 Mac 完整开发环境（包含所有数据库和管理工具）

#### 核心数据库

##### MySQL 8.0
| 配置项 | 值 |
|--------|-----|
| **容器名称** | `local-mysql` |
| **镜像** | `mysql:8.0` |
| **端口映射** | `3306:3306` |
| **数据库** | `jobfirst_basic` |
| **用户名** | `root` |
| **密码** | `local_dev_password` |
| **字符集** | `utf8mb4` |
| **时区** | `Asia/Shanghai` |

##### PostgreSQL 14
| 配置项 | 值 |
|--------|-----|
| **容器名称** | `local-postgres` |
| **镜像** | `postgres:14` |
| **端口映射** | `5432:5432` |
| **数据库** | `jobfirst_vector` |
| **用户名** | `postgres` |
| **密码** | `local_dev_password` |
| **时区** | `Asia/Shanghai` |

##### Redis 6
| 配置项 | 值 |
|--------|-----|
| **容器名称** | `local-redis` |
| **镜像** | `redis:6-alpine` |
| **端口映射** | `6379:6379` |
| **密码** | `local_dev_password` |
| **持久化** | `--appendonly yes` |

##### MongoDB 6.0
| 配置项 | 值 |
|--------|-----|
| **容器名称** | `local-mongodb` |
| **镜像** | `mongo:6.0` |
| **端口映射** | `27017:27017` |
| **用户名** | `admin` |
| **密码** | `local_dev_password` |
| **时区** | `Asia/Shanghai` |

##### Neo4j
| 配置项 | 值 |
|--------|-----|
| **容器名称** | `local-neo4j` |
| **镜像** | `neo4j:latest` |
| **HTTP端口** | `7474:7474` |
| **Bolt端口** | `7687:7687` |
| **用户名** | `neo4j` |
| **密码** | `local_dev_password` |

#### 扩展数据库（需使用 `--profile full` 启动）

##### Elasticsearch 8.8.0
| 配置项 | 值 |
|--------|-----|
| **容器名称** | `local-elasticsearch` |
| **镜像** | `elasticsearch:8.8.0` |
| **端口映射** | `9200:9200`, `9300:9300` |
| **内存限制** | `768m` |
| **安全** | 禁用 |

##### Weaviate
| 配置项 | 值 |
|--------|-----|
| **容器名称** | `local-weaviate` |
| **镜像** | `semitechnologies/weaviate:latest` |
| **端口映射** | `8080:8080` |
| **内存限制** | `256m` |
| **匿名访问** | 启用 |

#### 管理工具

##### Adminer
| 配置项 | 值 |
|--------|-----|
| **容器名称** | `local-adminer` |
| **镜像** | `adminer:latest` |
| **端口映射** | `8888:8080` |
| **访问** | http://localhost:8888 |

##### Redis Commander（需使用 `--profile tools`）
| 配置项 | 值 |
|--------|-----|
| **容器名称** | `local-redis-commander` |
| **镜像** | `rediscommander/redis-commander:latest` |
| **端口映射** | `8081:8081` |
| **访问** | http://localhost:8081 |

##### Mongo Express（需使用 `--profile tools`）
| 配置项 | 值 |
|--------|-----|
| **容器名称** | `local-mongo-express` |
| **镜像** | `mongo-express:latest` |
| **端口映射** | `8082:8081` |
| **访问** | http://localhost:8082 |
| **用户名** | `admin` |
| **密码** | `local_dev_password` |

**启动命令**:
```bash
# 基础环境
docker-compose -f docker/docker-compose.local.yml up -d

# 完整环境（包含扩展数据库）
docker-compose -f docker/docker-compose.local.yml --profile full up -d

# 包含管理工具
docker-compose -f docker/docker-compose.local.yml --profile full --profile tools up -d
```

---

### 3. `docker-compose.dev.yml` (二次开发配置)

**用途**: Zervigo 二次开发环境（Go 认证 + Python AI）

| 服务 | 容器名称 | 数据库 | 密码 |
|------|---------|--------|------|
| MySQL | `zervigo-mysql` | `jobfirst` | `dev_password` |
| PostgreSQL | `zervigo-postgres` | `jobfirst_vector` | `dev_password` |
| Redis | `zervigo-redis` | - | - |

**特殊配置**:
- Redis 无需密码
- 用于 Go 认证服务和 Python AI 服务开发

---

### 4. `docker-compose-postgres.yml` (PostgreSQL 专用)

**用途**: 统一使用 PostgreSQL 作为主要数据库

| 配置项 | 值 |
|--------|-----|
| **数据库版本** | PostgreSQL 15-alpine |
| **容器名称** | `zervigo-postgres-mvp` |
| **端口映射** | `5432:5432` |
| **数据库** | `zervigo_mvp` |
| **用户名** | `postgres` |
| **密码** | `dev_password` |
| **Redis密码** | `dev_password` |
| **Consul版本** | 1.15 |

**环境变量配置**:
```bash
# 所有微服务使用相同的环境变量格式
POSTGRESQL_URL=postgres://postgres:dev_password@postgres:5432/zervigo_mvp?sslmode=disable
REDIS_URL=redis://:dev_password@redis:6379
JWT_SECRET=zervigo-mvp-secret-key-2025
```

---

## Docker 卷（数据持久化）

当前项目中的 Docker 卷：

| 卷名称 | 类型 | 用途 |
|--------|------|------|
| `zervigo_pgadmin_data` | local | PgAdmin 数据 |
| `zervigo_postgres_data` | local | PostgreSQL 数据 |
| `zervigo_redis_data` | local | Redis 数据 |
| `jobfirst_local_mysql` | local | 本地 MySQL 数据 |
| `jobfirst_local_postgres` | local | 本地 PostgreSQL 数据 |
| `jobfirst_local_redis` | local | 本地 Redis 数据 |
| `jobfirst_local_mongodb` | local | 本地 MongoDB 数据 |
| `jobfirst_local_neo4j_data` | local | Neo4j 数据目录 |
| `jobfirst_local_neo4j_logs` | local | Neo4j 日志目录 |
| `jobfirst_local_elasticsearch` | local | Elasticsearch 数据 |
| `jobfirst_local_weaviate` | local | Weaviate 数据 |
| `postgres_mvp_data` | local | MVP PostgreSQL 数据 |
| `redis_mvp_data` | local | MVP Redis 数据 |
| `consul_mvp_data` | local | Consul 数据 |

**查看命令**:
```bash
docker volume ls | grep zervigo
docker volume inspect <volume-name>
```

---

## Docker 网络

### 当前网络

| 网络名称 | 驱动 | 用途 |
|----------|------|------|
| `bridge` | bridge | Docker 默认网桥 |
| `docker_zervigo-mvp` | bridge | Zervigo MVP 网络 |
| `host` | host | 主机网络 |
| `none` | null | 无网络 |
| `zervigo_zervigo-network` | bridge | Zervigo 主要网络 |
| `zervitest_zervi-network` | bridge | Zervigo 测试网络 |
| `jobfirst_local_network` | bridge | JobFirst 本地网络 |
| `172.20.0.0/16` | bridge | 自定义子网 |

**查看命令**:
```bash
docker network ls
docker network inspect <network-name>
```

---

## 端口汇总表

### 数据库端口

| 服务 | 本地端口 | 容器端口 | 说明 |
|------|---------|---------|------|
| **zervigo-postgres** | `15432` | `5432` | 当前运行 |
| **local-postgres** | `5432` | `5432` | 开发环境 |
| **mvp-postgres** | `5432` | `5432` | MVP 环境 |
| **local-mysql** | `3306` | `3306` | 开发环境 |
| **mvp-mysql** | `3306` | `3306` | MVP 环境 |
| **local-redis** | `6379` | `6379` | 所有环境 |
| **local-mongodb** | `27017` | `27017` | 开发环境 |
| **local-neo4j** | `7474`/`7687` | `7474`/`7687` | 开发环境 |
| **elasticsearch** | `9200`/`9300` | `9200`/`9300` | 扩展环境 |
| **weaviate** | `8080` | `8080` | 扩展环境 |

### 服务端口

| 服务 | 本地端口 | 说明 |
|------|---------|------|
| **central-brain** | `9000` | API Gateway |
| **auth-service** | `8207` | 认证服务 |
| **user-service** | `8082` | 用户服务 |
| **job-service** | `8084` | 职位服务 |
| **resume-service** | `8085` | 简历服务 |
| **company-service** | `8083` | 企业服务 |
| **ai-service** | `8100` | AI 服务 |
| **blockchain-service** | `8208` | 区块链服务 |
| **consul** | `8500` | 服务发现 |
| **adminer** | `8888` | 数据库管理 |
| **redis-commander** | `8081` | Redis 管理 |
| **mongo-express** | `8082` | MongoDB 管理 |

---

## 连接示例

### 1. PostgreSQL 连接

```bash
# 当前运行容器
psql -h localhost -p 15432 -U postgres -d zervigo_unified
# 密码: postgres

# 开发环境
psql -h localhost -p 5432 -U postgres -d jobfirst_vector
# 密码: local_dev_password

# MVP 环境
psql -h localhost -p 5432 -U postgres -d zervigo_mvp
# 密码: dev_password
```

### 2. Redis 连接

```bash
# 无密码（MVP）
redis-cli -h localhost -p 6379

# 有密码（开发环境）
redis-cli -h localhost -p 6379 -a local_dev_password
```

### 3. MySQL 连接

```bash
# 开发环境
mysql -h localhost -P 3306 -u root -p local_dev_password jobfirst_basic

# MVP 环境
mysql -h localhost -P 3306 -u root -p dev_password zervigo_mvp
```

### 4. MongoDB 连接

```bash
# 开发环境
mongosh mongodb://admin:local_dev_password@localhost:27017/
```

### 5. Neo4j 连接

```bash
# HTTP 界面
open http://localhost:7474
# 用户名: neo4j
# 密码: local_dev_password

# Bolt 连接
cypher-shell -a bolt://localhost:7687 -u neo4j -p local_dev_password
```

---

## 常用 Docker 命令

### 启动和停止

```bash
# 启动所有服务
docker-compose -f docker/docker-compose-postgres.yml up -d

# 停止所有服务
docker-compose -f docker/docker-compose-postgres.yml down

# 查看日志
docker-compose -f docker/docker-compose-postgres.yml logs -f

# 重启特定服务
docker-compose -f docker/docker-compose-postgres.yml restart postgres
```

### 数据备份和恢复

```bash
# PostgreSQL 备份
docker exec zervigo-postgres pg_dump -U postgres zervigo_unified > backup.sql

# PostgreSQL 恢复
docker exec -i zervigo-postgres psql -U postgres zervigo_unified < backup.sql

# Redis 备份
docker exec zervigo-redis redis-cli SAVE
docker cp zervigo-redis:/data/dump.rdb ./backup.rdb

# 清理所有容器和卷
docker-compose down -v
```

---

## 故障排查

### 端口冲突

如果遇到端口被占用：

```bash
# 查看端口占用
lsof -i :5432

# 修改端口映射（在 docker-compose.yml 中）
ports:
  - "15432:5432"  # 主机端口:容器端口
```

### 连接失败

```bash
# 检查容器状态
docker ps -a

# 查看容器日志
docker logs zervigo-postgres

# 进入容器
docker exec -it zervigo-postgres sh

# 测试连接
docker exec zervigo-postgres pg_isready -U postgres
```

### 数据持久化问题

```bash
# 查看卷信息
docker volume inspect zervigo_postgres_data

# 备份卷
docker run --rm -v zervigo_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres-backup.tar.gz /data

# 恢复卷
docker run --rm -v zervigo_postgres_data:/data -v $(pwd):/backup alpine tar xzf /backup/postgres-backup.tar.gz -C /
```

---

## 安全建议

1. **生产环境密码**: 所有配置中的密码都是开发环境密码，生产环境必须修改
2. **网络隔离**: 使用 Docker 网络隔离服务，不要暴露不必要的端口
3. **数据加密**: 敏感数据应加密存储
4. **定期备份**: 设置自动备份策略
5. **访问控制**: 限制数据库的访问权限

---

## 相关文档

- [Docker Compose 官方文档](https://docs.docker.com/compose/)
- [PostgreSQL 官方文档](https://www.postgresql.org/docs/)
- [Redis 官方文档](https://redis.io/documentation)
- 项目内部文档: `docs/START_STOP_SCRIPTS_VERIFICATION_REPORT.md`

---

**文档维护者**: AI Assistant  
**最后更新**: 2025-10-30  
**版本**: 1.0
