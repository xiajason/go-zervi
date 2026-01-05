# GoZervi本地云部署 - 快速开始指南

## 🚀 5分钟快速部署

### 步骤1: 生成配置文件

```bash
cd /Users/szjason72/gozervi/zervigo.demo
./scripts/setup-env.sh
```

**说明**: 
- 交互式配置生成
- 可以留空密码，系统会自动生成
- 建议使用默认端口配置

---

### 步骤2: 运行安装脚本

```bash
./scripts/install-local-cloud.sh
```

**说明**:
- 自动检查环境（Docker, Docker Compose）
- 自动导入镜像（如果存在本地镜像）
- 自动启动所有服务
- 自动执行健康检查

---

### 步骤3: 验证部署

```bash
# 检查所有服务状态
docker-compose -f docker/docker-compose.local-cloud.yml ps

# 检查服务健康
curl http://localhost:8207/health  # Auth Service
curl http://localhost:8088/health  # Tenant Service
```

---

## 📊 服务访问地址

### 基础设施
- **PostgreSQL**: `localhost:5432`
- **Redis**: `localhost:6379`
- **Consul UI**: `http://localhost:8500`

### 核心服务
- **Auth Service**: `http://localhost:8207`
- **Tenant Service**: `http://localhost:8088`
- **User Service**: `http://localhost:8082`

### 业务服务
- **Job Service**: `http://localhost:8084`
- **Company Service**: `http://localhost:8083`

---

## 🔧 常用操作

### 启动服务
```bash
docker-compose -f docker/docker-compose.local-cloud.yml up -d
```

### 停止服务
```bash
docker-compose -f docker/docker-compose.local-cloud.yml down
```

### 查看日志
```bash
# 查看所有服务日志
docker-compose -f docker/docker-compose.local-cloud.yml logs -f

# 查看特定服务日志
docker-compose -f docker/docker-compose.local-cloud.yml logs -f auth-service
```

### 重启服务
```bash
docker-compose -f docker/docker-compose.local-cloud.yml restart [service-name]
```

---

## 🎯 下一步

1. **测试API**: 使用Postman或curl测试各个服务的API
2. **查看文档**: 阅读API文档了解接口详情
3. **配置前端**: 配置前端连接到本地服务

---

**快速开始完成！** 🎉




