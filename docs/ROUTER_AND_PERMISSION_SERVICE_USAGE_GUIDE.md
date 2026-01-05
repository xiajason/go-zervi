# Router Service 和 Permission Service 使用指南

## 📋 目录

- [概述](#概述)
- [服务版本说明](#服务版本说明)
- [环境配置](#环境配置)
- [开发环境](#开发环境)
- [测试环境](#测试环境)
- [生产环境](#生产环境)
- [常见问题](#常见问题)

---

## 概述

Router Service 和 Permission Service 是 Zervigo 微服务架构中的核心基础设施服务：

- **Router Service** (端口 8087): 提供动态路由配置和用户路由权限管理
- **Permission Service** (端口 8086): 提供角色权限管理和用户权限验证

这两个服务都支持两种运行模式：
1. **Standalone 模式** - 简化版本，适用于快速测试和开发
2. **数据库模式** - 完整版本，适用于生产环境

---

## 服务版本说明

### Router Service

#### 1. standalone_main.go (简化版)

**特点**：
- ✅ 不依赖数据库
- ✅ 返回模拟数据
- ✅ 快速启动
- ✅ 适合开发测试

**使用场景**：
- 本地快速开发
- API 接口测试
- 前端集成测试
- 无需真实数据验证

#### 2. main.go (数据库版)

**特点**：
- ✅ 从数据库读取真实路由配置
- ✅ 支持动态路由管理
- ✅ 集成权限验证
- ✅ 适合生产环境

**使用场景**：
- 生产环境部署
- 需要真实路由数据
- 需要权限验证
- 需要数据持久化

### Permission Service

#### 只有一个版本 (main.go)

**特点**：
- ✅ 必须使用数据库
- ✅ 从数据库读取角色和权限数据
- ✅ 提供完整的权限管理功能
- ✅ 必须配置 PostgreSQL 环境变量

**注意**：Permission Service 没有 standalone 版本，必须依赖数据库。

---

## 环境配置

### 配置文件说明

所有环境使用 `configs/local.env` 配置文件：

```bash
# PostgreSQL配置（PostgreSQL 16 Docker容器 - 端口15432）
POSTGRESQL_HOST=localhost
POSTGRESQL_PORT=15432
POSTGRESQL_USER=postgres
POSTGRESQL_PASSWORD=postgres
POSTGRESQL_DATABASE=zervigo_unified
POSTGRESQL_SSL_MODE=disable

# 服务端口配置
ROUTER_SERVICE_PORT=8087
PERMISSION_SERVICE_PORT=8086
CENTRAL_BRAIN_PORT=9000
```

### 关键配置项

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `POSTGRESQL_DATABASE` | 数据库名称 | `zervigo_unified` |
| `POSTGRESQL_PORT` | 数据库端口 | `15432` |
| `POSTGRESQL_USER` | 数据库用户名 | `postgres` |
| `POSTGRESQL_PASSWORD` | 数据库密码 | `postgres` |
| `ROUTER_SERVICE_PORT` | Router服务端口 | `8087` |
| `PERMISSION_SERVICE_PORT` | Permission服务端口 | `8086` |

---

## 开发环境

### 快速启动（Standalone 模式）

适用于前端开发和 API 测试：

```bash
# 1. 启动 Router Service (Standalone 模式)
cd /Users/szjason72/szbolent/Zervigo/zervigo.demo/services/infrastructure/router
nohup go run standalone_main.go > /Users/szjason72/szbolent/Zervigo/zervigo.demo/logs/router-service.log 2>&1 &

# 2. 验证服务
curl http://localhost:8087/health
```

**优点**：
- 无需数据库配置
- 快速启动
- 返回模拟数据

### 完整启动（数据库模式）

适用于需要真实数据的开发：

```bash
# 1. 加载环境变量
cd /Users/szjason72/szbolent/Zervigo/zervigo.demo
source <(cat configs/local.env | grep "^[^#]" | grep -v "^$")

# 2. 启动 Router Service (数据库模式)
cd services/infrastructure/router
nohup go run main.go > /Users/szjason72/szbolent/Zervigo/zervigo.demo/logs/router-service.log 2>&1 &

# 3. 启动 Permission Service (必须加载环境变量)
cd /Users/szjason72/szbolent/Zervigo/zervigo.demo
source <(cat configs/local.env | grep "^[^#]" | grep -v "^$")
cd services/infrastructure/permission
nohup go run main.go > /Users/szjason72/szbolent/Zervigo/zervigo.demo/logs/permission-service.log 2>&1 &

# 4. 验证服务
curl http://localhost:8087/health
curl http://localhost:8086/health
```

### 验证配置

```bash
# 检查 Router Service
curl http://localhost:8087/health | jq .

# 检查 Permission Service 数据库配置
curl http://localhost:8086/health | jq '.core_health.database.postgresql'
```

**期望输出**：
```json
{
  "database": "zervigo_unified",
  "host": "localhost",
  "port": 15432
}
```

---

## 测试环境

### 使用 Central Brain 代理

测试环境通常通过 Central Brain 访问服务：

```bash
# 1. 启动所有服务（按照依赖顺序）

# a. 启动数据库
docker-compose -f docker/docker-compose.local.yml up -d postgres

# b. 加载环境变量并启动基础设施服务
cd /Users/szjason72/szbolent/Zervigo/zervigo.demo

# 启动 Permission Service
source <(cat configs/local.env | grep "^[^#]" | grep -v "^$")
cd services/infrastructure/permission
nohup go run main.go > /Users/szjason72/szbolent/Zervigo/zervigo.demo/logs/permission-service.log 2>&1 &

# 启动 Router Service (数据库模式)
cd ../router
nohup go run main.go > /Users/szjason72/szbolent/Zervigo/zervigo.demo/logs/router-service.log 2>&1 &

# 启动 Central Brain
cd ../../shared/central-brain
nohup go run main.go > /Users/szjason72/szbolent/Zervigo/zervigo.demo/logs/central-brain.log 2>&1 &

# 2. 验证服务
curl http://localhost:9000/health  # Central Brain
curl http://localhost:8087/health  # Router Service
curl http://localhost:8086/health  # Permission Service
```

### 通过 Central Brain 访问

```bash
# 获取路由配置
curl http://localhost:9000/api/v1/router/routes | jq .

# 获取角色列表
curl http://localhost:9000/api/v1/permission/roles | jq .

# 获取权限列表
curl http://localhost:9000/api/v1/permission/permissions | jq .
```

### 测试脚本

创建测试脚本 `scripts/test-router-permission.sh`：

```bash
#!/bin/bash

echo "=== Router & Permission Service 测试 ==="

# 测试 Router Service
echo ""
echo "1. 测试 Router Service..."
curl -s http://localhost:8087/health | jq '{"service": .service, "status": .status}'

# 测试 Permission Service
echo ""
echo "2. 测试 Permission Service..."
curl -s http://localhost:8086/health | jq '{
  "service": .service,
  "status": .status,
  "database": .core_health.database.postgresql | {database, host, port}
}'

# 测试 Central Brain 代理
echo ""
echo "3. 测试 Central Brain 代理..."
curl -s http://localhost:9000/api/v1/router/routes | jq '{"code": .code, "message": .message}'

echo ""
echo "=== 测试完成 ==="
```

---

## 生产环境

### Docker 部署

#### 1. Dockerfile 示例

**Router Service Dockerfile**：

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -o router-service main.go

# 运行镜像
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/router-service .

EXPOSE 8087

CMD ["./router-service"]
```

**Permission Service Dockerfile**：

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -o permission-service main.go

# 运行镜像
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/permission-service .

EXPOSE 8086

CMD ["./permission-service"]
```

#### 2. docker-compose 配置

```yaml
version: '3.8'

services:
  router-service:
    build: 
      context: ./services/infrastructure/router
    container_name: router-service
    ports:
      - "8087:8087"
    environment:
      - POSTGRESQL_HOST=postgres
      - POSTGRESQL_PORT=5432
      - POSTGRESQL_USER=postgres
      - POSTGRESQL_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRESQL_DATABASE=zervigo_unified
      - POSTGRESQL_SSL_MODE=disable
    depends_on:
      - postgres
    networks:
      - zervigo-network

  permission-service:
    build:
      context: ./services/infrastructure/permission
    container_name: permission-service
    ports:
      - "8086:8086"
    environment:
      - POSTGRESQL_HOST=postgres
      - POSTGRESQL_PORT=5432
      - POSTGRESQL_USER=postgres
      - POSTGRESQL_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRESQL_DATABASE=zervigo_unified
      - POSTGRESQL_SSL_MODE=disable
    depends_on:
      - postgres
    networks:
      - zervigo-network

  postgres:
    image: postgres:16-alpine
    container_name: zervigo-postgres
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_DB=zervigo_unified
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - zervigo-network

volumes:
  postgres-data:

networks:
  zervigo-network:
    driver: bridge
```

#### 3. 启动命令

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f router-service permission-service

# 停止服务
docker-compose down
```

### Kubernetes 部署

#### Deployment 示例

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: router-service
  namespace: zervigo
spec:
  replicas: 2
  selector:
    matchLabels:
      app: router-service
  template:
    metadata:
      labels:
        app: router-service
    spec:
      containers:
      - name: router-service
        image: zervigo/router-service:latest
        ports:
        - containerPort: 8087
        env:
        - name: POSTGRESQL_HOST
          value: "postgres-service"
        - name: POSTGRESQL_DATABASE
          value: "zervigo_unified"
        - name: POSTGRESQL_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
---
apiVersion: v1
kind: Service
metadata:
  name: router-service
  namespace: zervigo
spec:
  selector:
    app: router-service
  ports:
  - port: 8087
    targetPort: 8087
```

---

## 常见问题

### Q1: Router Service 是否必须使用数据库？

**A**: 不一定。有两个版本：
- `standalone_main.go` - 不需要数据库，返回模拟数据
- `main.go` - 需要数据库，读取真实数据

### Q2: Permission Service 为什么必须加载环境变量？

**A**: Permission Service 只有一个版本（`main.go`），它必须连接 PostgreSQL 数据库。如果不加载环境变量，会使用默认配置（`zervigo_mvp` 数据库，端口 `5432`），导致连接失败。

### Q3: 如何确认服务使用的是正确的数据库配置？

**A**: 检查健康检查接口：

```bash
# Permission Service
curl http://localhost:8086/health | jq '.core_health.database.postgresql'

# 期望输出
{
  "database": "zervigo_unified",  # 正确的数据库名
  "host": "localhost",
  "port": 15432  # 正确的端口
}
```

### Q4: 为什么要修改 `shared/core/core.go`？

**A**: 原来的代码中 PostgreSQL 配置是硬编码的，不会从环境变量读取。修改后：
- 支持从环境变量读取配置
- 可以在不同环境中使用不同的数据库配置
- 避免了硬编码配置的局限性

### Q5: Central Brain 如何代理 Router 和 Permission 服务？

**A**: Central Brain 作为 API Gateway，会：
1. 注册 Router 和 Permission 服务的代理路由
2. 请求 `/api/v1/router/**` 会转发到 Router Service
3. 请求 `/api/v1/permission/**` 会转发到 Permission Service
4. 添加认证和授权中间件

### Q6: 如何切换 Router Service 的版本？

**A**: 

```bash
# 停止当前版本
pkill -f router

# 启动 standalone 版本
cd services/infrastructure/router
go run standalone_main.go

# 或启动数据库版本
cd /Users/szjason72/szbolent/Zervigo/zervigo.demo
source <(cat configs/local.env | grep "^[^#]" | grep -v "^$")
cd services/infrastructure/router
go run main.go
```

### Q7: 服务启动顺序是什么？

**A**: 
1. 数据库服务（PostgreSQL）
2. Permission Service（依赖数据库）
3. Router Service（可选，数据库模式需要）
4. Central Brain（API Gateway，依赖其他服务）

### Q8: 如何查看服务日志？

**A**: 

```bash
# 实时查看日志
tail -f logs/router-service.log
tail -f logs/permission-service.log
tail -f logs/central-brain.log

# 查看最近50行
tail -50 logs/router-service.log
```

---

## 快速参考命令

```bash
# 加载环境变量
source <(cat configs/local.env | grep "^[^#]" | grep -v "^$")

# 启动 Router Service (standalone)
cd services/infrastructure/router
go run standalone_main.go

# 启动 Router Service (数据库模式)
source <(cat configs/local.env | grep "^[^#]" | grep -v "^$")
cd services/infrastructure/router
go run main.go

# 启动 Permission Service
source <(cat configs/local.env | grep "^[^#]" | grep -v "^$")
cd services/infrastructure/permission
go run main.go

# 检查服务状态
curl http://localhost:8087/health  # Router
curl http://localhost:8086/health  # Permission
curl http://localhost:9000/health  # Central Brain

# 停止服务
pkill -f router
pkill -f permission
lsof -ti:8087 | xargs kill -9
lsof -ti:8086 | xargs kill -9
```

---

**文档版本**: 1.0  
**最后更新**: 2025-10-30  
**维护者**: Zervigo 开发团队

