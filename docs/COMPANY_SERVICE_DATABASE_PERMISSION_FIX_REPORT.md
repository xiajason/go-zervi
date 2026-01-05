# Company服务PostgreSQL数据库权限字段配置检查报告

## 📋 检查概述

**检查日期**: 2025-01-29  
**检查范围**: Company服务在PostgreSQL数据库中的表结构和权限字段配置  
**数据库**: zervigo_mvp (PostgreSQL)

## 🔍 发现的问题

### 1. 权限管理表缺失 ❌

**问题**: 代码中定义的权限管理表在PostgreSQL数据库中不存在

**缺失的表**:
- ❌ `company_users` - 企业用户关联表
- ❌ `company_permission_audit_logs` - 权限审计日志表
- ❌ `company_data_sync_status` - 数据同步状态表

**影响**: Company服务无法正常工作，因为`CompanyPermissionManager`依赖这些表进行权限检查。

### 2. companies表缺少权限相关字段 ❌

**问题**: `companies`表中缺少权限管理必需的字段

**缺失的字段**:
- ❌ `created_by` - 创建者用户ID
- ❌ `legal_rep_user_id` - 法定代表人用户ID
- ❌ `authorized_users` - 授权用户列表（JSON格式）
- ❌ `unified_social_credit_code` - 统一社会信用代码
- ❌ `legal_representative_id` - 法定代表人身份证号

**影响**: `EnhancedCompany`结构体中的权限相关字段无法映射到数据库。

### 3. 迁移脚本不兼容 ❌

**问题**: 现有的迁移脚本（`006_enhance_company_auth.sql`, `007_create_missing_tables.sql`）是为MySQL设计的

**不兼容的语法**:
- MySQL的`AUTO_INCREMENT` vs PostgreSQL的`BIGSERIAL`
- MySQL的`JSON` vs PostgreSQL的`JSONB`
- MySQL的`UNIQUE KEY` vs PostgreSQL的`UNIQUE`
- MySQL的`COMMENT`语法 vs PostgreSQL的`COMMENT ON`

## ✅ 修复方案

### 1. 创建PostgreSQL版本的迁移脚本

已创建: `databases/postgres/init/09-company-permission-tables.sql`

**包含内容**:
1. ✅ 在`companies`表中添加权限相关字段
2. ✅ 创建`company_users`表（企业用户关联表）
3. ✅ 创建`company_permission_audit_logs`表（权限审计日志表）
4. ✅ 创建`company_data_sync_status`表（数据同步状态表）
5. ✅ 创建必要的索引
6. ✅ 创建触发器（自动更新时间戳）

### 2. 修复后的表结构

#### companies表新增字段

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `created_by` | BIGINT | 创建者用户ID |
| `legal_rep_user_id` | BIGINT | 法定代表人用户ID |
| `authorized_users` | JSONB | 授权用户列表（JSON数组） |
| `unified_social_credit_code` | VARCHAR(50) | 统一社会信用代码（唯一索引） |
| `legal_representative_id` | VARCHAR(50) | 法定代表人身份证号 |

#### company_users表

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `id` | BIGSERIAL | 主键 |
| `company_id` | BIGINT | 企业ID（外键） |
| `user_id` | BIGINT | 用户ID（外键） |
| `role` | VARCHAR(50) | 角色：legal_rep, authorized_user, admin |
| `status` | VARCHAR(20) | 状态：active, inactive, pending |
| `permissions` | JSONB | 权限列表（JSON数组） |
| `created_at` | TIMESTAMP | 创建时间 |
| `updated_at` | TIMESTAMP | 更新时间 |

**唯一约束**: `(company_id, user_id)` - 一个用户在一个企业中只能有一个角色

#### company_permission_audit_logs表

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `id` | BIGSERIAL | 主键 |
| `company_id` | BIGINT | 企业ID（外键） |
| `user_id` | BIGINT | 用户ID（外键） |
| `action` | VARCHAR(100) | 操作类型 |
| `resource_type` | VARCHAR(50) | 资源类型 |
| `resource_id` | BIGINT | 资源ID |
| `permission_result` | BOOLEAN | 权限检查结果 |
| `ip_address` | VARCHAR(45) | IP地址 |
| `user_agent` | TEXT | 用户代理 |
| `created_at` | TIMESTAMP | 创建时间 |

#### company_data_sync_status表

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `id` | BIGSERIAL | 主键 |
| `company_id` | BIGINT | 企业ID（外键） |
| `sync_target` | VARCHAR(50) | 同步目标：postgresql, neo4j, redis |
| `sync_status` | VARCHAR(20) | 同步状态：pending, syncing, success, failed |
| `last_sync_time` | TIMESTAMP | 最后同步时间 |
| `sync_error` | TEXT | 同步错误信息 |
| `retry_count` | INT | 重试次数 |
| `created_at` | TIMESTAMP | 创建时间 |
| `updated_at` | TIMESTAMP | 更新时间 |

**唯一约束**: `(company_id, sync_target)` - 一个企业对每个同步目标只能有一条记录

### 3. 索引优化

**company_users表索引**:
- `idx_company_users_company_id` - 按企业ID查询
- `idx_company_users_user_id` - 按用户ID查询
- `idx_company_users_role` - 按角色查询
- `idx_company_users_status` - 按状态查询
- `idx_company_users_company_user` - 复合索引（company_id, user_id, status）
- `idx_company_users_user_company` - 复合索引（user_id, company_id, role）

**company_permission_audit_logs表索引**:
- `idx_company_audit_company_id` - 按企业ID查询
- `idx_company_audit_user_id` - 按用户ID查询
- `idx_company_audit_action` - 按操作类型查询
- `idx_company_audit_created_at` - 按创建时间查询

**company_data_sync_status表索引**:
- `idx_company_sync_company_id` - 按企业ID查询
- `idx_company_sync_target` - 按同步目标查询
- `idx_company_sync_status` - 按同步状态查询

## 🔗 外键关联

### company_users表
- `company_id` → `companies(company_id)` ON DELETE CASCADE
- `user_id` → `zervigo_auth_users(id)` ON DELETE CASCADE

### company_permission_audit_logs表
- `company_id` → `companies(company_id)` ON DELETE CASCADE
- `user_id` → `zervigo_auth_users(id)` ON DELETE CASCADE

### company_data_sync_status表
- `company_id` → `companies(company_id)` ON DELETE CASCADE

## 📊 权限管理机制

### 权限检查流程

1. **系统管理员权限**: 检查用户是否为`admin`或`super_admin`
2. **企业创建者权限**: 检查`companies.created_by`字段
3. **法定代表人权限**: 检查`companies.legal_rep_user_id`字段
4. **企业用户关联权限**: 检查`company_users`表中的`(company_id, user_id)`记录
5. **授权用户列表权限**: 检查`companies.authorized_users` JSONB字段

### 权限审计

所有权限检查操作都会记录到`company_permission_audit_logs`表中，包括：
- 用户ID
- 企业ID
- 操作类型
- 资源类型和ID
- 权限检查结果（允许/拒绝）
- IP地址和用户代理
- 操作时间

## ✅ 验证步骤

### 1. 执行迁移脚本

```bash
psql -h localhost -U szjason72 -d zervigo_mvp -f databases/postgres/init/09-company-permission-tables.sql
```

### 2. 验证表创建

```sql
-- 检查表是否存在
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name IN ('company_users', 'company_permission_audit_logs', 'company_data_sync_status')
ORDER BY table_name;

-- 检查companies表字段
SELECT column_name, data_type FROM information_schema.columns 
WHERE table_name = 'companies' 
AND column_name IN ('created_by', 'legal_rep_user_id', 'authorized_users', 'unified_social_credit_code')
ORDER BY column_name;
```

### 3. 验证索引

```sql
-- 检查索引
SELECT indexname, tablename FROM pg_indexes 
WHERE tablename IN ('company_users', 'company_permission_audit_logs', 'company_data_sync_status')
ORDER BY tablename, indexname;
```

### 4. 验证外键

```sql
-- 检查外键约束
SELECT conname, conrelid::regclass, confrelid::regclass 
FROM pg_constraint 
WHERE contype = 'f' 
AND conrelid::regclass::text IN ('company_users', 'company_permission_audit_logs', 'company_data_sync_status');
```

## 🎯 修复后的影响

### 正面影响

1. ✅ **Company服务可以正常启动**: 所有必需的表和字段都已存在
2. ✅ **权限管理功能完整**: 支持企业用户关联、权限审计、数据同步跟踪
3. ✅ **数据一致性**: 外键约束确保数据完整性
4. ✅ **查询性能**: 索引优化提升查询效率
5. ✅ **审计能力**: 完整的权限检查日志记录

### 需要注意的事项

1. ⚠️ **数据迁移**: 如果已有企业数据，需要更新`created_by`字段
2. ⚠️ **代码兼容性**: 确保`CompanyPermissionManager`使用正确的表名（`company_users`而不是`company_user`）
3. ⚠️ **外键依赖**: `user_id`外键指向`zervigo_auth_users(id)`，确保该表存在

## 📝 后续建议

1. **创建测试数据**: 添加一些测试企业用户关联记录
2. **验证权限检查**: 测试`CompanyPermissionManager`的所有权限检查逻辑
3. **性能测试**: 验证索引是否有效提升查询性能
4. **文档更新**: 更新API文档，说明权限管理机制

---

**报告生成时间**: 2025-01-29  
**修复状态**: ✅ 迁移脚本已创建，待执行  
**下一步**: 执行迁移脚本并验证表创建
