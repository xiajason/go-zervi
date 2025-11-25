# Zervigo 区块链微服务

## 📋 概述

Zervigo区块链微服务是基于Go语言开发的专门用于三版本架构（基础版、专业版、Future版）数据不可篡改记录和跨版本状态同步的微服务。

## 🎯 核心功能

### 1. 版本状态变化记录
- 记录用户在三个版本间的状态变化
- 提供不可篡改的审计轨迹
- 支持状态变化历史查询

### 2. 权限变更追踪
- 记录用户权限的变更历史
- 跨版本权限同步
- 权限变更审计

### 3. 数据一致性校验
- 定期校验跨版本数据一致性
- 自动检测数据不一致
- 提供一致性报告

### 4. 区块链交易管理
- 生成唯一的交易哈希
- 记录区块高度和时间戳
- 提供完整的交易历史

## 🏗️ 技术架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Zervigo 三版本架构                        │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   基础版     │  │   专业版     │  │   Future版   │          │
│  │  (8080-8089)│  │ (8601-8627) │  │ (7530-7540) │          │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘          │
│         │                │                │                │
│         └────────────────┼────────────────┘                │
│                          │                                 │
│  ┌──────────────────────▼──────────────────────┐           │
│  │        区块链微服务 (Blockchain Service)        │           │
│  │              Port: 8208                      │           │
│  │  ┌─────────────────────────────────────────┐ │           │
│  │  │  核心功能:                               │ │           │
│  │  │  • 跨版本状态记录                        │ │           │
│  │  │  • 权限变更追踪                          │ │           │
│  │  │  • 数据一致性校验                        │ │           │
│  │  │  • 不可篡改审计                          │ │           │
│  │  └─────────────────────────────────────────┘ │           │
│  └─────────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 快速开始

### 1. 环境要求
- Go 1.21+
- MySQL 8.0+
- Docker (可选)

### 2. 本地开发

```bash
# 1. 克隆项目
cd src/microservices/blockchain-service

# 2. 安装依赖
go mod tidy

# 3. 创建数据库
mysql -u root -p
CREATE DATABASE zervigo_blockchain DEFAULT CHARACTER SET utf8mb4;

# 4. 设置环境变量
export DATABASE_URL="root:@tcp(localhost:3306)/zervigo_blockchain?charset=utf8mb4&parseTime=True&loc=Local"
export JWT_SECRET="jobfirst-unified-auth-secret-key-2024"
export BLOCKCHAIN_SERVICE_PORT=8208

# 5. 启动服务
go run main.go
```

### 3. Docker部署

```bash
# 使用Docker Compose启动所有服务
docker-compose -f docker-compose.microservices.yml up -d blockchain-service
```

## 📡 API接口

### 健康检查
```http
GET /health
```

### 版本状态记录
```http
POST /api/v1/blockchain/version/status/record
Content-Type: application/json

{
  "user_id": "user_001",
  "version_source": "BASIC",
  "old_status": "ACTIVE",
  "new_status": "INACTIVE",
  "change_reason": "用户注销账户",
  "operator_id": "admin_001",
  "remark": "系统自动处理"
}
```

### 权限变更记录
```http
POST /api/v1/blockchain/permission/change/record
Content-Type: application/json

{
  "user_id": "user_001",
  "version_source": "PROFESSIONAL",
  "old_permission": "READ",
  "new_permission": "WRITE",
  "change_reason": "用户升级到专业版",
  "operator_id": "admin_001",
  "remark": "权限升级"
}
```

### 查询用户状态历史
```http
GET /api/v1/blockchain/version/status/history/{user_id}
```

### 查询用户权限历史
```http
GET /api/v1/blockchain/permission/change/history/{user_id}
```

### 数据一致性校验
```http
POST /api/v1/blockchain/consistency/validate
```

### 查询交易列表
```http
GET /api/v1/blockchain/transaction/list?page=1&size=10&transaction_type=VERSION_STATUS
```

## 🔧 在其他微服务中集成

### 1. 添加依赖

```go
import "github.com/jobfirst/jobfirst-core/blockchain"
```

### 2. 创建客户端

```go
// 创建区块链服务客户端
blockchainClient := blockchain.NewClient("http://localhost:8208")
```

### 3. 记录状态变化

```go
// 在用户状态变化时记录到区块链
func (s *UserService) UpdateUserStatus(ctx context.Context, userID, versionSource, oldStatus, newStatus, reason, operatorID string) error {
    // 1. 更新数据库
    if err := s.updateDatabaseUserStatus(userID, newStatus); err != nil {
        return err
    }
    
    // 2. 异步记录到区块链
    go func() {
        req := &blockchain.VersionStatusChangeRequest{
            UserID:        userID,
            VersionSource: versionSource,
            OldStatus:     oldStatus,
            NewStatus:     newStatus,
            ChangeReason:  reason,
            OperatorID:    operatorID,
        }
        
        if _, err := blockchainClient.RecordVersionStatusChange(context.Background(), req); err != nil {
            log.Printf("记录到区块链失败: %v", err)
        }
    }()
    
    return nil
}
```

### 4. 查询历史记录

```go
// 查询用户状态历史
resp, err := blockchainClient.GetUserStatusHistory(ctx, userID)
if err != nil {
    return err
}

// 处理响应数据
if records, ok := resp.Data.([]map[string]interface{}); ok {
    for _, record := range records {
        fmt.Printf("状态变化: %s -> %s (%s)\n", 
            record["old_status"], 
            record["new_status"], 
            record["record_time"])
    }
}
```

## 🗄️ 数据库设计

### 核心表结构

#### 1. blockchain_transaction (区块链交易总表)
```sql
CREATE TABLE blockchain_transaction (
    transaction_id VARCHAR(64) PRIMARY KEY COMMENT '交易ID',
    transaction_hash VARCHAR(128) NOT NULL COMMENT '区块链交易哈希',
    transaction_type VARCHAR(32) NOT NULL COMMENT '交易类型',
    version_source VARCHAR(32) COMMENT '版本来源',
    user_id VARCHAR(64) COMMENT '用户ID',
    old_status VARCHAR(32) COMMENT '旧状态',
    new_status VARCHAR(32) COMMENT '新状态',
    change_reason VARCHAR(500) COMMENT '变更原因',
    operator_id VARCHAR(64) COMMENT '操作人ID',
    transaction_data TEXT COMMENT '交易数据JSON',
    status VARCHAR(32) NOT NULL COMMENT '交易状态',
    block_height BIGINT COMMENT '区块高度',
    create_time DATETIME NOT NULL COMMENT '创建时间',
    confirm_time DATETIME COMMENT '确认时间',
    remark VARCHAR(500) COMMENT '备注'
);
```

#### 2. version_status_record (版本状态记录表)
```sql
CREATE TABLE version_status_record (
    record_id VARCHAR(64) PRIMARY KEY COMMENT '记录ID',
    user_id VARCHAR(64) NOT NULL COMMENT '用户ID',
    version_source VARCHAR(32) NOT NULL COMMENT '版本来源',
    old_status VARCHAR(32) COMMENT '旧状态',
    new_status VARCHAR(32) NOT NULL COMMENT '新状态',
    change_reason VARCHAR(500) COMMENT '变更原因',
    operator_id VARCHAR(64) COMMENT '操作人ID',
    transaction_hash VARCHAR(128) NOT NULL COMMENT '区块链交易哈希',
    block_height BIGINT COMMENT '区块高度',
    record_time DATETIME NOT NULL COMMENT '记录时间'
);
```

#### 3. permission_change_record (权限变更记录表)
```sql
CREATE TABLE permission_change_record (
    record_id VARCHAR(64) PRIMARY KEY COMMENT '记录ID',
    user_id VARCHAR(64) NOT NULL COMMENT '用户ID',
    version_source VARCHAR(32) NOT NULL COMMENT '版本来源',
    old_permission VARCHAR(32) COMMENT '旧权限',
    new_permission VARCHAR(32) NOT NULL COMMENT '新权限',
    change_reason VARCHAR(500) COMMENT '变更原因',
    operator_id VARCHAR(64) COMMENT '操作人ID',
    transaction_hash VARCHAR(128) NOT NULL COMMENT '区块链交易哈希',
    block_height BIGINT COMMENT '区块高度',
    record_time DATETIME NOT NULL COMMENT '记录时间'
);
```

## 🧪 测试

### 1. 单元测试
```bash
go test ./...
```

### 2. 集成测试
```bash
# 启动服务
go run main.go &

# 运行测试脚本
go run example.go
```

### 3. API测试
```bash
# 健康检查
curl http://localhost:8208/health

# 记录状态变化
curl -X POST http://localhost:8208/api/v1/blockchain/version/status/record \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_001",
    "version_source": "BASIC",
    "old_status": "ACTIVE",
    "new_status": "INACTIVE",
    "change_reason": "测试",
    "operator_id": "admin_001"
  }'

# 查询历史
curl http://localhost:8208/api/v1/blockchain/version/status/history/user_001
```

## 🔒 安全考虑

### 1. 数据验证
- 所有输入数据都经过严格验证
- 防止SQL注入攻击
- 限制请求频率

### 2. 权限控制
- API访问需要适当的权限
- 操作记录包含操作人信息
- 敏感操作需要额外验证

### 3. 数据加密
- 敏感数据在传输和存储时加密
- 使用HTTPS进行API通信
- 定期更新加密密钥

## 📊 监控和运维

### 1. 健康检查
```bash
curl http://localhost:8208/health
```

### 2. 日志查看
```bash
# 查看服务日志
docker logs zervigo-blockchain-service

# 查看错误日志
docker logs zervigo-blockchain-service 2>&1 | grep ERROR
```

### 3. 性能监控
- 监控API响应时间
- 监控数据库连接数
- 监控内存和CPU使用率

## 🚀 部署建议

### 1. 生产环境配置
```yaml
# docker-compose.prod.yml
blockchain-service:
  image: zervigo-blockchain-service:latest
  environment:
    - DATABASE_URL=mysql://user:password@mysql:3306/zervigo_blockchain
    - JWT_SECRET=${JWT_SECRET}
    - BLOCKCHAIN_SERVICE_PORT=8208
  ports:
    - "8208:8208"
  restart: always
  deploy:
    resources:
      limits:
        memory: 512M
        cpus: '1.0'
```

### 2. 高可用部署
- 使用负载均衡器
- 部署多个实例
- 配置数据库主从复制
- 设置自动故障转移

## 📚 相关文档

- [Zervigo三版本架构设计](../README.md)
- [微服务集成指南](../INTEGRATION_GUIDE.md)
- [API接口文档](./API_DOCS.md)
- [部署运维指南](./DEPLOYMENT_GUIDE.md)

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](../LICENSE) 文件了解详情。

---

**维护者**: Zervigo团队  
**最后更新**: 2025-01-10  
**版本**: 1.0.0
