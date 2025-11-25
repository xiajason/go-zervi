# Zervigo 数据库架构优化方案

## 🎯 优化目标

基于项目现状，将数据库架构简化为：
- **PostgreSQL**: 主力数据库，承担所有业务数据存储
- **Redis**: 缓存和会话存储
- **SQLite3**: 用户个人数据隔离和本地开发

## 📊 当前问题分析

### 数据库使用现状
| 数据库 | 当前用途 | 问题 |
|--------|----------|------|
| MySQL | 主数据库 | 功能重复，增加复杂度 |
| PostgreSQL | 向量存储 | 未充分利用其强大功能 |
| Redis | 缓存 | ✅ 合理使用 |
| SQLite3 | 简历服务 | ✅ 合理使用 |

### 优化理由
1. **减少复杂度**: 避免维护多种数据库类型
2. **功能统一**: PostgreSQL功能更强大，支持所有需求
3. **成本降低**: 减少运维复杂度
4. **性能提升**: PostgreSQL查询优化器更先进

## 🚀 优化后的架构

### 主力数据库：PostgreSQL
```yaml
PostgreSQL (端口: 5432):
  业务数据:
    - 用户管理 (users, roles, permissions)
    - 企业管理 (companies, verifications)
    - 职位管理 (jobs, applications)
    - 简历管理 (resumes, templates)
    - 区块链记录 (transactions, blocks)
  
  AI数据:
    - 向量存储 (embeddings)
    - 文档解析结果
    - 智能匹配数据
  
  技术特性:
    - JSON字段支持
    - 全文搜索
    - 数组类型
    - 地理信息
    - 并行查询
```

### 辅助数据库
```yaml
Redis (端口: 6379):
  用途:
    - 用户会话存储
    - API响应缓存
    - 临时数据存储
    - 分布式锁
  
SQLite3 (本地):
  用途:
    - 用户个人数据隔离
    - 本地开发和测试
    - 敏感数据存储
    - 离线数据支持
```

## 📋 迁移实施计划

### 第一阶段：PostgreSQL主数据库建设 (1周)

#### 1.1 数据库初始化
```sql
-- 创建主数据库
CREATE DATABASE zervigo_main;

-- 创建业务数据表
-- 用户管理
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 企业管理
CREATE TABLE companies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    industry VARCHAR(50),
    size VARCHAR(20),
    address TEXT,
    website VARCHAR(100),
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 职位管理
CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    company_id INTEGER REFERENCES companies(id),
    description TEXT,
    requirements TEXT,
    salary_min INTEGER,
    salary_max INTEGER,
    location VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 简历管理
CREATE TABLE resumes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    title VARCHAR(100) NOT NULL,
    content JSONB,
    template_id INTEGER,
    status VARCHAR(20) DEFAULT 'draft',
    view_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 区块链记录
CREATE TABLE blockchain_transactions (
    id SERIAL PRIMARY KEY,
    transaction_hash VARCHAR(64) UNIQUE NOT NULL,
    block_height INTEGER NOT NULL,
    transaction_type VARCHAR(50) NOT NULL,
    data JSONB,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- AI向量数据
CREATE TABLE ai_embeddings (
    id SERIAL PRIMARY KEY,
    content_type VARCHAR(50) NOT NULL,
    content_id INTEGER NOT NULL,
    embedding VECTOR(1536), -- 假设使用1536维向量
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 1.2 服务配置更新
```yaml
# 更新所有服务的数据库配置
database:
  postgresql:
    host: localhost
    port: 5432
    user: postgres
    password: dev_password
    database: zervigo_main
    ssl_mode: disable
    max_connections: 100
    max_idle_connections: 10
    connection_max_lifetime: 30m
```

### 第二阶段：服务迁移 (2周)

#### 2.1 认证服务迁移
```go
// 更新认证服务数据库连接
func initDatabase() (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        config.PostgreSQL.Host,
        config.PostgreSQL.Port,
        config.PostgreSQL.User,
        config.PostgreSQL.Password,
        config.PostgreSQL.Database,
        config.PostgreSQL.SSLMode,
    )
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    
    // 自动迁移表结构
    db.AutoMigrate(&User{}, &Role{}, &Permission{}, &UserRole{})
    
    return db, nil
}
```

#### 2.2 业务服务迁移
```go
// 更新所有业务服务的数据库连接
// 用户服务、公司服务、职位服务、简历服务
// 统一使用PostgreSQL作为主数据库
```

#### 2.3 AI服务集成
```python
# AI服务直接使用PostgreSQL存储向量数据
import psycopg2
from pgvector import Vector

def store_embedding(content_type, content_id, embedding):
    conn = psycopg2.connect(
        host="localhost",
        port=5432,
        user="postgres",
        password="dev_password",
        database="zervigo_main"
    )
    
    cursor = conn.cursor()
    cursor.execute(
        "INSERT INTO ai_embeddings (content_type, content_id, embedding) VALUES (%s, %s, %s)",
        (content_type, content_id, Vector(embedding))
    )
    conn.commit()
    conn.close()
```

### 第三阶段：测试和优化 (1周)

#### 3.1 数据迁移测试
```bash
# 测试数据迁移脚本
./scripts/migrate-to-postgresql.sh

# 验证数据完整性
./scripts/verify-data-integrity.sh

# 性能测试
./scripts/postgresql-performance-test.sh
```

#### 3.2 服务集成测试
```bash
# 启动所有服务
./scripts/start-all-services.sh

# 运行集成测试
./scripts/test-mvp.sh

# 健康检查
./scripts/comprehensive_health_check.sh
```

## 🔧 实施脚本

### PostgreSQL初始化脚本
```bash
#!/bin/bash
# scripts/init-postgresql.sh

echo "🚀 初始化PostgreSQL主数据库..."

# 创建数据库
psql -U postgres -c "CREATE DATABASE zervigo_main;"

# 创建pgvector扩展
psql -U postgres -d zervigo_main -c "CREATE EXTENSION IF NOT EXISTS vector;"

# 执行初始化脚本
psql -U postgres -d zervigo_main -f databases/postgres/init/01-init-schema.sql

# 创建索引
psql -U postgres -d zervigo_main -c "
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_companies_name ON companies(name);
CREATE INDEX idx_jobs_company_id ON jobs(company_id);
CREATE INDEX idx_resumes_user_id ON resumes(user_id);
CREATE INDEX idx_ai_embeddings_content ON ai_embeddings(content_type, content_id);
"

echo "✅ PostgreSQL初始化完成！"
```

### 数据迁移脚本
```bash
#!/bin/bash
# scripts/migrate-to-postgresql.sh

echo "🔄 开始数据迁移到PostgreSQL..."

# 备份现有数据
./scripts/server_full_backup.sh

# 迁移用户数据
python scripts/migrate_users.py

# 迁移公司数据
python scripts/migrate_companies.py

# 迁移职位数据
python scripts/migrate_jobs.py

# 迁移简历数据
python scripts/migrate_resumes.py

# 验证数据完整性
./scripts/verify-data-integrity.sh

echo "✅ 数据迁移完成！"
```

## 📊 优化效果预期

### 性能提升
- **查询性能**: PostgreSQL查询优化器更先进
- **并发处理**: 支持更好的并发访问
- **索引优化**: 更灵活的索引策略

### 运维简化
- **数据库类型**: 从4种减少到3种
- **配置管理**: 统一的PostgreSQL配置
- **备份策略**: 简化的备份和恢复流程

### 功能增强
- **JSON支持**: 原生JSON字段支持
- **全文搜索**: 内置全文搜索功能
- **向量搜索**: pgvector扩展支持AI向量搜索
- **地理信息**: 支持地理位置数据

## 🎯 验收标准

### 第一阶段验收
- [ ] PostgreSQL数据库初始化完成
- [ ] 所有表结构创建成功
- [ ] 索引创建完成
- [ ] 连接池配置正确

### 第二阶段验收
- [ ] 所有服务成功连接到PostgreSQL
- [ ] 数据迁移完成
- [ ] 服务功能正常
- [ ] API接口正常响应

### 第三阶段验收
- [ ] 性能测试通过
- [ ] 数据完整性验证通过
- [ ] 集成测试通过
- [ ] 生产环境准备就绪

## 🚀 立即行动计划

### 第1天：PostgreSQL环境准备
```bash
# 1. 启动PostgreSQL
docker-compose -f docker/docker-compose-postgres.yml up -d postgres

# 2. 初始化数据库
./scripts/init-postgresql.sh

# 3. 验证连接
psql -U postgres -d zervigo_main -c "SELECT version();"
```

### 第2-3天：服务配置更新
```bash
# 1. 更新认证服务配置
# 2. 更新业务服务配置
# 3. 更新AI服务配置
# 4. 测试服务连接
```

### 第4-5天：数据迁移和测试
```bash
# 1. 执行数据迁移
# 2. 运行集成测试
# 3. 性能验证
# 4. 问题修复
```

这个优化方案将：
- **简化架构**: 减少数据库类型，降低复杂度
- **提升性能**: 充分利用PostgreSQL的强大功能
- **降低成本**: 减少运维工作量
- **增强功能**: 支持更多高级特性

您觉得这个优化方案如何？需要我立即开始实施吗？
