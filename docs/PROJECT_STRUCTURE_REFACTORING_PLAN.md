# Zervigo 项目结构重构方案

## 🎯 重构目标
统一所有微服务的模块命名规范，解决Go模块冲突问题，建立清晰的项目架构。

## 📋 当前问题分析

### 1. 模块命名混乱
- **Go-Zero生成的模块**：使用简单名称（如 `auth`, `user`, `job`）
- **手动创建的模块**：使用完整路径（如 `github.com/szjason72/zervigo/auth-service-go`）
- **第三方模块**：使用不同的命名空间（如 `github.com/jobfirst/jobfirst-core`）

### 2. 目录结构混乱
- `service/` - Go-Zero生成的API服务
- `src/microservices/` - 手动创建的微服务
- `src/auth-service-go/` - 独立的认证服务
- `tools/`, `rpc/` - Go-Zero生成的工具和RPC定义

## 🏗️ 统一命名规范

### 模块命名规范
```
github.com/szjason72/zervigo/{service-type}/{service-name}
```

### 服务类型分类
- `core` - 核心服务（认证、用户管理）
- `business` - 业务服务（职位、简历、公司）
- `infrastructure` - 基础设施服务（区块链、AI、通知）
- `shared` - 共享库和工具

### 具体模块映射
```
# 核心服务
auth-service -> github.com/szjason72/zervigo/core/auth
user-service -> github.com/szjason72/zervigo/core/user

# 业务服务  
job-service -> github.com/szjason72/zervigo/business/job
resume-service -> github.com/szjason72/zervigo/business/resume
company-service -> github.com/szjason72/zervigo/business/company

# 基础设施服务
blockchain-service -> github.com/szjason72/zervigo/infrastructure/blockchain
ai-service -> github.com/szjason72/zervigo/infrastructure/ai
notification-service -> github.com/szjason72/zervigo/infrastructure/notification
statistics-service -> github.com/szjason72/zervigo/infrastructure/statistics
banner-service -> github.com/szjason72/zervigo/infrastructure/banner
template-service -> github.com/szjason72/zervigo/infrastructure/template

# 共享库
jobfirst-core -> github.com/szjason72/zervigo/shared/core
central-brain -> github.com/szjason72/zervigo/shared/central-brain
```

## 📁 目录结构重构

### 新的目录结构
```
zervigo.demo/
├── services/                    # 所有微服务
│   ├── core/                   # 核心服务
│   │   ├── auth/               # 认证服务
│   │   └── user/               # 用户服务
│   ├── business/               # 业务服务
│   │   ├── job/                # 职位服务
│   │   ├── resume/             # 简历服务
│   │   └── company/            # 公司服务
│   └── infrastructure/         # 基础设施服务
│       ├── blockchain/         # 区块链服务
│       ├── ai/                 # AI服务
│       ├── notification/       # 通知服务
│       ├── statistics/         # 统计服务
│       ├── banner/             # 横幅服务
│       └── template/           # 模板服务
├── shared/                     # 共享库
│   ├── core/                  # 核心共享库
│   └── central-brain/          # 中央大脑
├── api/                        # API定义（保留Go-Zero结构）
├── rpc/                        # RPC定义（保留Go-Zero结构）
└── tools/                      # 工具（保留Go-Zero结构）
```

## 🔄 重构步骤

### 第一阶段：创建新的目录结构
1. 创建 `services/` 目录
2. 按服务类型创建子目录
3. 移动现有服务到新位置

### 第二阶段：更新模块定义
1. 更新所有 `go.mod` 文件的模块名
2. 更新导入路径
3. 统一依赖管理

### 第三阶段：更新工作区配置
1. 重新配置 `go.work` 文件
2. 测试模块解析
3. 验证服务启动

### 第四阶段：清理旧结构
1. 删除冲突的模块定义
2. 清理重复的依赖
3. 更新文档和脚本

## ✅ 预期效果

1. **模块命名统一**：所有模块使用一致的命名规范
2. **依赖关系清晰**：明确的模块依赖关系
3. **开发体验提升**：IDE支持更好，代码导航更清晰
4. **维护成本降低**：结构清晰，易于理解和维护
5. **扩展性增强**：新服务可以轻松集成到现有架构中
