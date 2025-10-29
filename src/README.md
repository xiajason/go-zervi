# Zervigo 二次开发环境

## 项目结构

```
src/
├── auth-service-go/          # Go认证服务源码
│   ├── main.go              # 主程序入口
│   ├── go.mod               # Go模块配置
│   ├── jobfirst-core/       # 核心认证库
│   └── unified-auth         # 编译后的二进制文件
│
├── ai-service-python/        # Python AI服务源码
│   ├── ai_service_with_zervigo.py  # 主AI服务
│   ├── zervigo_auth_middleware.py  # 认证中间件
│   ├── unified_auth_client.py      # 认证客户端
│   ├── requirements.txt             # Python依赖
│   └── Dockerfile                  # Docker配置
│
└── shared/                  # 共享配置和工具
    └── config.go            # 共享配置
```

## 开发环境配置

### Go服务开发

```bash
# 进入Go服务目录
cd src/auth-service-go

# 安装依赖
go mod download

# 运行开发服务器
go run main.go

# 构建二进制文件
go build -o unified-auth main.go
```

### Python服务开发

```bash
# 进入Python服务目录
cd src/ai-service-python

# 创建虚拟环境
python3 -m venv venv
source venv/bin/activate

# 安装依赖
pip install -r requirements.txt

# 运行AI服务
python ai_service_with_zervigo.py
```

## 环境变量配置

### 认证服务 (Go)
```bash
export DATABASE_URL="root:@tcp(localhost:3306)/jobfirst?charset=utf8mb4&parseTime=True&loc=Local"
export JWT_SECRET="jobfirst-unified-auth-secret-key-2024"
export AUTH_SERVICE_PORT=8207
```

### AI服务 (Python)
```bash
export ZERVIGO_AUTH_URL="http://localhost:8207"
export AI_SERVICE_PORT=8100
export POSTGRES_HOST="localhost"
export POSTGRES_USER="szjason72"
export POSTGRES_DB="jobfirst_vector"
```

## Docker开发环境

```bash
# 启动数据库集群
docker-compose -f docker-compose.local.yml up -d

# 运行Go认证服务
cd src/auth-service-go
go run main.go &

# 运行Python AI服务
cd src/ai-service-python
source venv/bin/activate
python ai_service_with_zervigo.py &
```

## API端点

### 认证服务 (8207端口)
- POST /api/v1/auth/login - 用户登录
- POST /api/v1/auth/validate - JWT验证
- GET  /api/v1/auth/permission - 权限检查
- GET  /api/v1/auth/user - 获取用户信息
- GET  /health - 健康检查

### AI服务 (8100端口)
- GET  /health - 健康检查
- GET  /api/v1/ai/user-info - 获取用户信息
- GET  /api/v1/ai/permissions - 获取权限列表
- POST /api/v1/ai/job-matching - 职位匹配
- POST /api/v1/ai/resume-analysis - 简历分析
- POST /api/v1/ai/chat - AI聊天

## 开发建议

1. **先启动认证服务**：确保Zervigo服务在8207端口运行
2. **再启动AI服务**：AI服务依赖认证服务
3. **使用热重载**：Go使用air，Python使用watchdog
4. **测试集成**：确保两个服务能正常通信
5. **日志监控**：查看服务日志进行调试

## 下一步开发

- [ ] 添加新的认证方式
- [ ] 扩展AI功能
- [ ] 优化性能
- [ ] 添加单元测试
- [ ] 配置CI/CD流水线
