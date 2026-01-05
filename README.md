# 🚀 Go-Zervi Framework - Zervigo MVP 项目

> **"Innovation Beyond Go-Zero"** - 一个创新的Go微服务框架

## 🎯 Go-Zervi 框架介绍

**Go-Zervi** 是一个基于Go语言的全新微服务框架，它借鉴了Go-Zero的优秀设计理念，同时创新了分层架构和模块管理方式。

### ✨ 核心特性
- 🏗️ **分层微服务架构**: core/business/infrastructure 三层设计
- 🏷️ **统一模块命名**: `github.com/xiajason/go-zervi/{type}/{service}` 格式
- 📦 **共享库管理**: 统一的共享库和工具
- 🔧 **复杂技术栈**: PostgreSQL + Redis + Neo4j + Consul
- ⚙️ **自定义业务逻辑**: 完全控制服务实现
- 📋 **API/RPC定义**: 借鉴Go-Zero的优秀设计格式

### 🎨 框架理念
```
Go-Zervi = Go-Zero的优秀设计 + 我们的创新架构
Go-Zervi = 借鉴 + 创新 + 超越
Go-Zervi = 未来微服务框架的新标准
```

## 📁 Go-Zervi 项目结构

```
zervigo.demo/
├── services/                    # Go-Zervi 微服务目录
│   ├── core/                   # 核心服务层
│   │   ├── auth/               # 认证服务 (端口: 8207)
│   │   │   ├── main.go
│   │   │   ├── go.mod
│   │   │   └── README.md
│   │   └── user/               # 用户服务 (端口: 8082)
│   │       ├── main.go
│   │       ├── go.mod
│   │       └── README.md
│   ├── business/               # 业务服务层
│   │   ├── job/                # 职位服务 (端口: 8084)
│   │   ├── resume/             # 简历服务 (端口: 8085)
│   │   └── company/            # 企业服务 (端口: 8083)
│   └── infrastructure/         # 基础设施层
│       ├── blockchain/         # 区块链服务 (端口: 8208)
│       ├── ai/                 # AI服务 (端口: 8100)
│       ├── notification/       # 通知服务
│       ├── statistics/         # 统计服务
│       ├── banner/             # 横幅服务
│       └── template/            # 模板服务
├── shared/                     # Go-Zervi 共享库
│   ├── core/                  # 核心共享库
│   │   ├── auth/              # 认证模块
│   │   ├── database/          # 数据库模块
│   │   ├── middleware/        # 中间件模块
│   │   └── go.mod
│   └── central-brain/        # 中央大脑 (端口: 9000)
├── cleanup-backup/             # 历史方案归档（含原 Taro 前端文档与脚本）
│   ├── docs/
│   └── scripts/
├── api/                        # API定义 (借鉴Go-Zero格式)
│   ├── auth.api
│   ├── user.api
│   ├── job.api
│   ├── resume.api
│   ├── company.api
│   ├── blockchain.api
│   └── ai.api
├── rpc/                        # RPC定义 (借鉴Go-Zero格式)
│   ├── auth/
│   ├── user/
│   ├── job/
│   ├── resume/
│   ├── company/
│   ├── blockchain/
│   └── ai/
├── databases/                  # 数据库配置
│   └── postgres/
│       └── init/
│           ├── 01-init-schema.sql
│           └── 02-zervigo-microservices-schema.sql
├── configs/                    # 配置文件
│   ├── dev.env
│   ├── local.env
│   └── jobfirst-core-config.yaml
├── scripts/                    # 脚本文件
│   ├── start-local-services.sh
│   ├── stop-local-services.sh
│   ├── migrate-microservices-db.sh
│   └── unified-version-management.sh
├── docs/                       # 文档
│   ├── GO_ZERVI_FRAMEWORK_NAMING_PLAN.md
│   ├── FRAMEWORK_INNOVATION_CHECK_REPORT.md
│   ├── MICROSERVICE_DATABASE_DESIGN.md
│   └── PROJECT_COMPLETENESS_REPORT.md
├── go.work                     # Go工作空间
└── README.md                   # 项目说明
```

> ⚠️ 前端客户端正在迁移至 **NativeScript-Vue**；原先的 Taro 方案文档与脚本已整体归档于 `cleanup-backup/` 目录，方便回溯。

## 🎯 Go-Zervi 核心服务

### 🏗️ 分层架构设计

#### 1. **Core Layer (核心服务层)**
- **认证服务 (Auth Service)**: 端口 8207 - 用户认证、权限管理、JWT Token
- **用户服务 (User Service)**: 端口 8082 - 用户管理、个人资料

#### 2. **Business Layer (业务服务层)**  
- **职位服务 (Job Service)**: 端口 8084 - 职位管理、职位搜索
- **简历服务 (Resume Service)**: 端口 8085 - 简历管理、简历分析
- **企业服务 (Company Service)**: 端口 8083 - 企业管理、企业认证

#### 3. **Infrastructure Layer (基础设施层)**
- **区块链服务 (Blockchain Service)**: 端口 8208 - 数据审计、不可篡改记录
- **AI服务 (AI Service)**: 端口 8100 - 智能匹配、简历分析、AI聊天
- **通知服务 (Notification Service)**: 端口 8086 - 消息通知、邮件推送
- **统计服务 (Statistics Service)**: 端口 8087 - 数据统计、报表分析
- **横幅服务 (Banner Service)**: 端口 8088 - 广告横幅、内容管理
- **模板服务 (Template Service)**: 端口 8089 - 模板管理、内容模板

#### 4. **Shared Layer (共享层)**
- **中央大脑 (Central Brain)**: 端口 9000 - 统一入口、智能路由、请求代理
- **核心共享库 (Core Library)**: 统一认证、数据库、中间件、工具函数

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

## 🚀 Go-Zervi 快速启动

### 📋 环境要求
- **Go版本**: 1.25+
- **Python版本**: 3.9+
- **PostgreSQL版本**: 14+
- **Redis版本**: 7.0+
- **Node.js版本**: 22+

### 🛠️ 本地开发环境启动

```bash
# 1. 启动本地PostgreSQL和Redis
brew services start postgresql@14
brew services start redis

# 2. 初始化数据库
./scripts/migrate-microservices-db.sh

# 3. 启动所有Go-Zervi服务
./scripts/start-local-services.sh

# 4. 检查服务状态
curl http://localhost:9000/health
curl http://localhost:8207/health
curl http://localhost:8082/health

# 5. 停止所有服务
./scripts/stop-local-services.sh
```

### 🔧 服务管理

```bash
# 启动单个服务
cd services/core/auth && go run main.go
cd services/core/user && go run main.go
cd services/business/job && go run main.go

# 启动前端开发
cd frontend && npm run dev:h5
cd frontend && npm run dev:weapp

# 版本管理
./scripts/unified-version-management.sh
```

## 📊 Go-Zervi 服务端口分配

### 🏗️ 分层服务架构

| 服务类型 | 服务名称 | 端口 | 状态 | 说明 |
|---------|---------|------|------|------|
| **Core Layer** | Central Brain | 9000 | ✅ | 中央大脑，统一入口 |
| | Auth Service | 8207 | ✅ | 认证服务 |
| | User Service | 8082 | ✅ | 用户服务 |
| **Business Layer** | Job Service | 8084 | ✅ | 职位服务 |
| | Resume Service | 8085 | ✅ | 简历服务 |
| | Company Service | 8083 | ✅ | 企业服务 |
| **Infrastructure Layer** | Blockchain Service | 8208 | ✅ | 区块链服务 |
| | AI Service | 8100 | ✅ | AI服务 |
| | Notification Service | 8086 | ✅ | 通知服务 |
| | Statistics Service | 8087 | ✅ | 统计服务 |
| | Banner Service | 8088 | ✅ | 横幅服务 |
| | Template Service | 8089 | ✅ | 模板服务 |
| **Infrastructure** | PostgreSQL | 5432 | ✅ | 主数据库 |
| | Redis | 6379 | ✅ | 缓存和会话 |
| | Consul | 8500 | ✅ | 服务发现 |

## 🔧 Go-Zervi 开发环境

### 📋 技术栈
- **后端框架**: Go-Zervi (基于Go语言)
- **前端框架**: NativeScript-Vue（移动端客户端，迁移进行中）
- **数据库**: PostgreSQL 14+ (主数据库)
- **缓存**: Redis 7.0+ (缓存和会话)
- **服务发现**: Consul
- **AI服务**: Python + Sanic
- **版本管理**: Git + 统一版本管理脚本

### 🛠️ 开发工具
- **Go版本**: 1.25+
- **Python版本**: 3.9+
- **Node.js版本**: 22+
- **PostgreSQL版本**: 14+
- **Redis版本**: 7.0+
- **Docker**: 可选 (本地开发环境)

## 📚 Go-Zervi 相关文档

### 🎯 框架文档
- [Go-Zervi框架命名方案](docs/GO_ZERVI_FRAMEWORK_NAMING_PLAN.md)
- [框架创新检查报告](docs/FRAMEWORK_INNOVATION_CHECK_REPORT.md)
- [项目结构重构计划](docs/PROJECT_STRUCTURE_REFACTORING_PLAN.md)

### 🗄️ 数据库文档
- [微服务数据库设计](docs/MICROSERVICE_DATABASE_DESIGN.md)
- [数据库迁移完成报告](docs/DATABASE_MIGRATION_COMPLETION_REPORT.md)
- [本地开发环境报告](docs/LOCAL_DEVELOPMENT_ENVIRONMENT_REPORT.md)

### 🚀 部署和运维
- [OpenLinkSaaS集成计划](docs/OPENLINKSASS_INTEGRATION_PLAN.md)
- [三阶段实施计划](docs/THREE_PHASE_IMPLEMENTATION_PLAN.md)
- [项目完整性报告](docs/PROJECT_COMPLETENESS_REPORT.md)

### 🔧 技术分析
- [JobFirst-Core 多数据库架构分析](docs/JOBFIRST_CORE_DATABASE_ARCHITECTURE_ANALYSIS.md) ⭐ **新增**
- [JobFirst-Core 与 Go-Zervi 集成指南](docs/JOBFIRST_CORE_INTEGRATION_GUIDE.md) ⭐ **新增**
- [Go-Zero冲突分析和优化](docs/GO_ZERO_CONFLICT_ANALYSIS_AND_OPTIMIZATION.md)
- [项目清理完成报告](docs/PROJECT_CLEANUP_COMPLETION_REPORT.md)
- [Docker依赖清理报告](docs/DOCKER_DEPENDENCY_CLEANUP_REPORT.md)

### 🎉 前端交付文档 ⭐ **新增**
- [后端前端交付检查清单](docs/BACKEND_FRONTEND_HANDOVER_CHECKLIST.md) - 完整交付清单和检查标准
- [前端快速入门指南](docs/FRONTEND_GETTING_STARTED.md) - 前端开发快速开始
- [Postman Collection使用指南](docs/POSTMAN_COLLECTION_GUIDE.md) - API测试和调试
- [后端交付总结](docs/BACKEND_DELIVERY_SUMMARY.md) - 交付状态和验收标准
