# Go-Zervi 框架代码整理方案

## 🎯 目标

将 **Go-Zervi 框架核心代码** 与 **Zervigo Demo/原型文件** 分离，只将框架核心代码推送到 GitHub。

## 📋 核心框架代码（应推送到 GitHub）

### ✅ 1. 框架核心库
```
shared/core/              # Go-Zervi 框架核心库
├── auth/                # 认证模块
├── config/              # 配置管理
├── database/            # 数据库管理
├── logging/             # 日志系统
├── middleware/          # 中间件
├── service/             # 服务管理
├── response/            # 响应格式化
└── utils/               # 工具函数
```

### ✅ 2. 中央大脑（可选，如果属于框架）
```
shared/central-brain/    # 微服务中央大脑（如果属于框架核心）
```

### ✅ 3. API/RPC 定义（框架标准格式）
```
api/                     # API 定义文件（.api 格式）
rpc/                     # RPC 定义文件（.proto 格式）
```

### ✅ 4. 数据库初始化脚本
```
databases/postgres/init/ # PostgreSQL 初始化脚本（框架需要）
```

### ✅ 5. 配置文件模板
```
configs/                 # 配置文件模板
├── dev.env
├── local.env
└── *.yaml
```

### ✅ 6. 框架文档
```
docs/
├── GO_ZERVI_*.md        # 框架相关文档
├── FRAMEWORK_*.md
├── JOBFIRST_CORE_*.md
└── AUTHENTICATION_*.md
```

### ✅ 7. 框架脚本
```
scripts/
├── test-go-zervi-framework.sh
└── 其他框架相关的工具脚本
```

### ✅ 8. Docker 配置（框架示例）
```
docker/
└── docker-compose*.yml  # Docker 配置示例
```

### ✅ 9. 项目基础文件
```
README.md
CHANGELOG.md
VERSION
.gitignore
```

---

## ❌ Demo/原型文件（不应推送到框架仓库）

### 🚫 1. 原型和参考项目
```
prototypes/              # 所有原型文件（包含大量大文件）
cleanup-backup/          # 备份文件
```

### 🚫 2. 示例服务（这些是使用框架的 Demo，不是框架本身）
```
services/                # 所有服务（auth, user, job, resume 等）
├── core/
├── business/
└── infrastructure/
```

### 🚫 3. Python 服务
```
src/ai-service-python/   # Python 服务（不是 Go 框架的一部分）
```

### 🚫 4. 编译产物和日志
```
bin/                     # 编译后的二进制文件
logs/                    # 日志文件
*.pid                    # 进程 ID 文件
```

### 🚫 5. 运行时文件
```
.gocache/                # Go 构建缓存
*.log                    # 日志文件
```

### 🚫 6. Zervigo 特定文档（非框架文档）
```
docs/ 中与 zervigo demo 业务相关的文档（保留框架文档）
```

---

## 🔧 实施步骤

### 步骤 1: 更新 .gitignore

确保以下内容被忽略：
```gitignore
# Demo 和原型文件
prototypes/
cleanup-backup/
services/
src/

# 编译产物
bin/
*.exe
*.so
*.dylib
**/*-service
**/main

# 运行时文件
logs/
*.log
*.pid
.gocache/
.cache/

# 大文件
*.jar
*.zip
*.tar.gz
*.db
*.sqlite
```

### 步骤 2: 清理 Git 历史（如果需要）

如果这些大文件已经在 Git 历史中，需要从历史中移除。

### 步骤 3: 创建新的干净提交

只提交核心框架代码。

---

## 📊 预期结果

### 框架仓库应包含：
- ✅ `shared/core/` - 框架核心库
- ✅ `api/` 和 `rpc/` - 定义文件
- ✅ `docs/` - 框架文档
- ✅ `README.md` - 框架说明
- ✅ `.gitignore` - 忽略规则
- ✅ `configs/` - 配置模板

### 框架仓库不应包含：
- ❌ 任何服务实现（services/）
- ❌ 原型文件（prototypes/）
- ❌ 备份文件（cleanup-backup/）
- ❌ 编译产物
- ❌ 大文件（模型、JAR 包等）

---

## 🎯 建议的仓库结构

```
go-zervi/                    # GitHub 仓库
├── shared/
│   └── core/                # 框架核心库
├── api/                     # API 定义
├── rpc/                     # RPC 定义
├── databases/               # 数据库脚本
├── configs/                 # 配置模板
├── docs/                    # 框架文档
├── scripts/                 # 框架工具脚本
├── docker/                  # Docker 示例
├── examples/                # 简单的使用示例（可选）
├── README.md
├── CHANGELOG.md
└── .gitignore
```

---

## ⚠️ 注意事项

1. **服务实现是 Demo**：`services/` 目录下的所有服务都是使用框架的示例，不应包含在框架仓库中
2. **文档筛选**：只保留框架相关文档，移除 Zervigo 业务特定文档
3. **大小限制**：确保仓库大小在合理范围内（< 100MB 理想）
4. **后续维护**：可以创建单独的 `go-zervi-examples` 仓库来存放示例服务

