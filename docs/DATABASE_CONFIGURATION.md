# 数据库配置说明

## 配置概览

本项目使用 PostgreSQL 数据库，配置已更新为使用 `vuecmf` 用户。

### 数据库连接信息

- **主机**: localhost
- **端口**: 5432
- **数据库**: zervigo_mvp
- **用户名**: vuecmf
- **密码**: vuecmf

### 配置文件位置

主配置文件：`/Users/szjason72/gozervi/zervigo.demo/configs/local.env`

## 已完成的配置更改

### 1. 更新 local.env

```bash
# PostgreSQL 配置
POSTGRESQL_HOST=localhost
POSTGRESQL_PORT=5432
POSTGRESQL_USER=vuecmf
POSTGRESQL_PASSWORD=vuecmf
POSTGRESQL_DATABASE=zervigo_mvp
POSTGRESQL_SSL_MODE=disable
POSTGRES_DB=zervigo_mvp
POSTGRES_USER=vuecmf
```

### 2. 创建 vuecmf 数据库用户

```sql
-- 创建用户
CREATE USER vuecmf WITH PASSWORD 'vuecmf';

-- 授予权限
GRANT ALL PRIVILEGES ON DATABASE zervigo_mvp TO vuecmf;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO vuecmf;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO vuecmf;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO vuecmf;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO vuecmf;
```

### 3. 更新服务凭证

服务凭证表 `zervigo_service_credentials` 中的密钥已更新为正确的 bcrypt 哈希值：

- **服务ID**: central-brain
- **明文密钥** (配置在 local.env): central-brain-secret-2025
- **Bcrypt哈希** (存储在数据库): $2a$10$B0x/JcQ2Mza6O.HZWjR0TuvAdmPq1Fvdklwijdtz5O29vfGPMZ1h.

## 验证配置

### 测试数据库连接

```bash
PGPASSWORD=vuecmf psql -h localhost -U vuecmf -d zervigo_mvp -c "SELECT version();"
```

### 查看服务凭证

```bash
PGPASSWORD=vuecmf psql -h localhost -U vuecmf -d zervigo_mvp -c "SELECT service_id, service_name, status FROM zervigo_service_credentials;"
```

## 重启服务

配置更新后，需要重启 central-brain 服务以应用新配置：

```bash
cd /Users/szjason72/gozervi/zervigo.demo/shared/central-brain
go run .
```

## 故障排除

### 问题：服务认证失败（服务密钥错误）

**原因**: 数据库中存储的密钥哈希与配置文件中的明文密钥不匹配。

**解决方案**: 
1. 确保 local.env 中的 `SERVICE_SECRET` 为明文
2. 确保数据库中的 `service_secret` 为对应的 bcrypt 哈希值
3. 可使用工具生成新的哈希值：
   ```bash
   cd /Users/szjason72/gozervi/zervigo.demo/shared/core
   go run ./cmd/generate-hash --password "your-secret-here"
   ```

### 问题：连接数据库失败

**检查项**:
1. PostgreSQL 服务是否运行
2. 用户名和密码是否正确
3. 数据库是否存在
4. 用户是否有足够的权限

## 工具脚本

### 生成 Bcrypt 哈希

位置：`/Users/szjason72/gozervi/zervigo.demo/shared/core/cmd/generate-hash/main.go`

用法：
```bash
cd /Users/szjason72/gozervi/zervigo.demo/shared/core
go run ./cmd/generate-hash --password "your-plaintext-password"
```

### 检查密码哈希

位置：`/Users/szjason72/gozervi/zervigo.demo/shared/core/cmd/check-password/main.go`

用法：
```bash
cd /Users/szjason72/gozervi/zervigo.demo/shared/core
go run ./cmd/check-password --hash "bcrypt-hash" --passwords "pwd1,pwd2,pwd3"
```

## 安全注意事项

1. **生产环境密钥**: 在生产环境中，请更换所有默认密钥和密码
2. **密钥存储**: 数据库中始终存储 bcrypt 哈希值，配置文件中使用明文
3. **权限控制**: 确保数据库用户只有必要的权限
4. **密钥轮换**: 定期更换服务密钥以提高安全性

## 相关文件

- 配置文件: `configs/local.env`
- 数据库初始化: `databases/postgres/init/06-service-credentials-management.sql`
- 修复脚本: `databases/postgres/fix-service-credentials.sql`
- 认证代码: `shared/core/auth/service_auth.go`

