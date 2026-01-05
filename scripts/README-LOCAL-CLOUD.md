# GoZervi 本地云部署指南

## 📋 概述

GoZervi本地云部署方案实现了完全离线、一键部署的能力，所有资源本地化存储，不依赖外部服务。

---

## 🚀 快速开始

### 1. 生成配置文件

```bash
# 交互式配置生成
./scripts/setup-env.sh
```

或者手动复制模板：

```bash
cp docker/.env.template docker/.env
# 编辑 docker/.env 文件
```

### 2. 运行安装脚本

```bash
# 一键安装
./scripts/install-local-cloud.sh
```

### 3. 检查服务状态

```bash
# 查看所有服务状态
docker-compose -f docker/docker-compose.local-cloud.yml ps

# 查看服务日志
docker-compose -f docker/docker-compose.local-cloud.yml logs -f [service-name]
```

---

## 📊 服务访问地址

### 基础设施服务
- **PostgreSQL**: `localhost:5432`
- **Redis**: `localhost:6379`
- **Consul**: `http://localhost:8500`

### 核心服务
- **Auth Service**: `http://localhost:8207`
- **Tenant Service**: `http://localhost:8088`
- **User Service**: `http://localhost:8082`

### 业务服务
- **Job Service**: `http://localhost:8084`
- **Company Service**: `http://localhost:8083`

---

## 🔧 常用命令

### 服务管理

```bash
# 启动所有服务
docker-compose -f docker/docker-compose.local-cloud.yml up -d

# 停止所有服务
docker-compose -f docker/docker-compose.local-cloud.yml down

# 重启服务
docker-compose -f docker/docker-compose.local-cloud.yml restart [service-name]

# 查看服务状态
docker-compose -f docker/docker-compose.local-cloud.yml ps

# 查看服务日志
docker-compose -f docker/docker-compose.local-cloud.yml logs -f [service-name]

# 进入容器
docker-compose -f docker/docker-compose.local-cloud.yml exec [service-name] sh
```

### 健康检查

```bash
# 检查PostgreSQL
docker-compose -f docker/docker-compose.local-cloud.yml exec postgresql pg_isready -U zervigo

# 检查Redis
docker-compose -f docker/docker-compose.local-cloud.yml exec redis redis-cli -a zervigo2025 ping

# 检查Consul
docker-compose -f docker/docker-compose.local-cloud.yml exec consul consul members

# 检查Auth Service
curl http://localhost:8207/health

# 检查Tenant Service
curl http://localhost:8088/health
```

---

## 📝 配置文件说明

### 环境变量文件

配置文件位置：`docker/.env`

主要配置项：
- `POSTGRES_*`: 数据库配置
- `REDIS_*`: Redis配置
- `CONSUL_PORT`: Consul端口
- `*_SERVICE_PORT`: 各服务端口
- `JWT_SECRET`: JWT密钥
- `ENVIRONMENT`: 运行环境（development/production）

---

## 🎯 核心特性

### 1. 完全离线部署
- ✅ 所有Docker镜像本地存储
- ✅ 数据库初始化脚本本地化
- ✅ 配置文件模板化
- ✅ 不依赖外部服务

### 2. 一键安装
- ✅ 自动化环境检查
- ✅ 自动化镜像导入
- ✅ 自动化配置生成
- ✅ 自动化服务启动

### 3. 可扩展性
- ✅ 支持单机部署
- ✅ 支持多环境配置
- ✅ 支持版本升级

---

## 🔍 故障排查

### 服务无法启动

1. **检查Docker状态**
   ```bash
   docker info
   ```

2. **检查端口占用**
   ```bash
   lsof -i :5432  # PostgreSQL
   lsof -i :6379  # Redis
   lsof -i :8500  # Consul
   ```

3. **查看服务日志**
   ```bash
   docker-compose -f docker/docker-compose.local-cloud.yml logs [service-name]
   ```

### 数据库连接失败

1. **检查数据库是否启动**
   ```bash
   docker-compose -f docker/docker-compose.local-cloud.yml ps postgresql
   ```

2. **检查数据库日志**
   ```bash
   docker-compose -f docker/docker-compose.local-cloud.yml logs postgresql
   ```

3. **验证数据库连接**
   ```bash
   docker-compose -f docker/docker-compose.local-cloud.yml exec postgresql psql -U zervigo -d zervigo_mvp
   ```

---

## 📚 更多信息

- [本地云部署分析](./LOCAL_CLOUD_DEPLOYMENT_ANALYSIS.md)
- [本地云部署计划](./LOCAL_CLOUD_DEPLOYMENT_PLAN.md)
- [实施完成报告](./LOCAL_CLOUD_DEPLOYMENT_COMPLETE.md)

---

**最后更新**: 2025-01-XX




