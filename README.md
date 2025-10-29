# Zervigo MVP 项目结构

## 📁 目录结构

```
MVPDEMO/
├── src/                          # 源代码目录
│   ├── central-brain/            # 中央大脑 (Go)
│   │   ├── main.go
│   │   ├── central_brain.go
│   │   ├── proxy_controller.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── auth-service/             # 认证服务 (Go)
│   │   ├── main.go
│   │   ├── auth_service.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── ai-service/               # AI服务 (Python)
│   │   ├── main.py
│   │   ├── ai_service.py
│   │   ├── requirements.txt
│   │   └── Dockerfile
│   ├── blockchain-service/       # 区块链服务 (Go)
│   │   ├── main.go
│   │   ├── blockchain_service.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── user-service/             # 用户服务 (Go)
│   │   ├── main.go
│   │   ├── user_service.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── job-service/              # 职位服务 (Go)
│   │   ├── main.go
│   │   ├── job_service.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── resume-service/           # 简历服务 (Go)
│   │   ├── main.go
│   │   ├── resume_service.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── company-service/          # 企业服务 (Go)
│   │   ├── main.go
│   │   ├── company_service.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   └── shared/                   # 共享库
│       ├── config.go
│       ├── models.go
│       ├── utils.go
│       └── go.mod
├── frontend/                     # 前端代码 (Taro多端开发) ✅
│   ├── config/                  # 构建配置
│   │   ├── index.js             # 默认配置
│   │   ├── dev.js               # 开发环境配置
│   │   └── prod.js               # 生产环境配置
│   ├── src/                     # 源码目录
│   │   ├── pages/               # 页面文件
│   │   │   ├── index/           # 首页
│   │   │   ├── login/           # 登录页
│   │   │   ├── register/        # 注册页
│   │   │   ├── profile/         # 个人中心
│   │   │   ├── resume/          # 简历管理
│   │   │   ├── job/             # 职位列表
│   │   │   ├── company/         # 企业信息
│   │   │   ├── chat/            # AI聊天
│   │   │   └── search/          # 搜索页面
│   │   ├── components/          # 公共组件
│   │   ├── assets/              # 静态资源
│   │   │   └── prototypes/      # 原型图文件 (93张) ✅
│   │   ├── services/            # API服务
│   │   ├── utils/               # 工具函数
│   │   ├── store/               # 状态管理
│   │   ├── types/               # 类型定义
│   │   ├── app.tsx              # 应用入口
│   │   ├── app.config.ts        # 应用配置
│   │   └── app.scss             # 全局样式
│   ├── dist/                    # 构建输出目录
│   ├── package.json             # 项目依赖
│   ├── project.config.json      # 小程序项目配置
│   ├── babel.config.js          # Babel配置
│   ├── tsconfig.json            # TypeScript配置
│   └── README.md                # 前端项目说明
├── docker/                       # Docker配置
│   ├── docker-compose.yml
│   ├── docker-compose.dev.yml
│   └── docker-compose.prod.yml
├── databases/                    # 数据库配置
│   ├── mysql/
│   ├── redis/
│   └── init-scripts/
├── configs/                      # 配置文件
│   ├── dev.env
│   ├── prod.env
│   └── consul/
├── scripts/                      # 脚本文件
│   ├── start-mvp.sh
│   ├── stop-mvp.sh
│   ├── build-all.sh
│   └── test-mvp.sh
├── monitoring/                   # 监控配置
│   ├── prometheus/
│   ├── grafana/
│   └── logs/
├── docs/                         # 文档
│   ├── MVP_STRATEGY_DISCUSSION_RECORD.md
│   ├── API_DOCUMENTATION.md
│   ├── DEPLOYMENT_GUIDE.md
│   └── DEVELOPMENT_GUIDE.md
├── tests/                        # 测试文件
│   ├── unit/
│   ├── integration/
│   └── e2e/
├── README.md                     # 项目说明
├── go.work                       # Go工作空间
└── .gitignore                    # Git忽略文件
```

## 🎯 MVP核心服务

### 1. 中央大脑 (Central Brain)
- **端口**: 9000
- **功能**: 统一入口、智能路由、请求代理
- **技术**: Go + Gin

### 2. 认证服务 (Auth Service)
- **端口**: 8207
- **功能**: 用户认证、权限管理、JWT Token
- **技术**: Go + MySQL

### 3. AI服务 (AI Service)
- **端口**: 8100
- **功能**: 智能匹配、简历分析、AI聊天
- **技术**: Python + Sanic

### 4. 区块链服务 (Blockchain Service)
- **端口**: 8208
- **功能**: 数据审计、不可篡改记录
- **技术**: Go + MySQL

### 5. 用户服务 (User Service)
- **端口**: 8082
- **功能**: 用户管理、个人资料
- **技术**: Go + MySQL

### 6. 职位服务 (Job Service)
- **端口**: 8084
- **功能**: 职位管理、职位搜索
- **技术**: Go + MySQL

### 7. 简历服务 (Resume Service)
- **端口**: 8085
- **功能**: 简历管理、简历分析
- **技术**: Go + MySQL

### 8. 企业服务 (Company Service)
- **端口**: 8083
- **功能**: 企业管理、企业认证
- **技术**: Go + MySQL

## 🎨 **原型图说明**

### **原型图文件**
- ✅ **总览模式.html** - 整体产品架构和功能模块展示
- ✅ **标注模式.html** - 详细功能标注和交互说明
- ✅ **演示模式.html** - 产品演示和用户体验展示
- ✅ **93张原型图图片** - 详细的UI设计参考

### **原型图用途**
- 📱 **UI设计参考** - 前端开发的重要设计依据
- 🔄 **交互逻辑** - 用户操作流程和界面交互
- 🎯 **功能模块** - 产品功能划分和模块设计
- 📊 **用户体验** - 界面布局和用户友好性

### **前端开发指导**
```bash
# 查看原型图
open MVPDEMO/frontend/prototypes/总览模式.html
open MVPDEMO/frontend/prototypes/标注模式.html
open MVPDEMO/frontend/prototypes/演示模式.html

# 原型图分析
cat MVPDEMO/frontend/prototypes/README.md
cat MVPDEMO/frontend/prototypes/PROTOTYPE_ANALYSIS.md
```

## 🚀 快速启动

```bash
# 1. 进入MVP目录
cd MVPDEMO

# 2. 启动所有服务
./scripts/start-mvp.sh

# 3. 检查服务状态
./scripts/test-mvp.sh

# 4. 停止所有服务
./scripts/stop-mvp.sh
```

## 📊 服务端口分配

| 服务 | 端口 | 状态 | 说明 |
|------|------|------|------|
| Central Brain | 9000 | ✅ | 中央大脑，统一入口 |
| Auth Service | 8207 | ✅ | 认证服务 |
| AI Service | 8100 | ✅ | AI服务 |
| Blockchain Service | 8208 | ✅ | 区块链服务 |
| User Service | 8082 | ✅ | 用户服务 |
| Job Service | 8084 | ✅ | 职位服务 |
| Resume Service | 8085 | ✅ | 简历服务 |
| Company Service | 8083 | ✅ | 企业服务 |
| MySQL | 3306 | ✅ | 数据库 |
| Redis | 6379 | ✅ | 缓存 |
| Consul | 8500 | ✅ | 服务发现 |

## 🔧 开发环境

- **Go版本**: 1.21+
- **Python版本**: 3.9+
- **Docker版本**: 20.10+
- **MySQL版本**: 8.0+
- **Redis版本**: 7.0+

## 📚 相关文档

- [MVP战略讨论记录](docs/MVP_STRATEGY_DISCUSSION_RECORD.md)
- [API接口文档](docs/API_DOCUMENTATION.md)
- [部署运维指南](docs/DEPLOYMENT_GUIDE.md)
- [开发指南](docs/DEVELOPMENT_GUIDE.md)
