#!/bin/bash
# 设置MySQL从库为只读模式
# 用途: 防止在从库误写入，保证数据一致性
# 部署位置: 腾讯云服务器 101.33.251.158
# 作者: AI Assistant
# 日期: 2025-10-10

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

echo ""
echo "=========================================="
echo "设置MySQL从库只读模式"
echo "=========================================="
echo ""

log_info "设置只读模式..."

docker exec test-mysql mysql -uroot -ptest_mysql_password << 'EOSQL'

-- 设置全局只读
SET GLOBAL read_only = ON;
SET GLOBAL super_read_only = ON;

-- 验证设置
SELECT @@read_only AS read_only_status, 
       @@super_read_only AS super_read_only_status;

-- 创建只读用户（供应用使用）
CREATE USER IF NOT EXISTS 'app_readonly'@'%' IDENTIFIED BY 'readonly_password_2025';
GRANT SELECT ON *.* TO 'app_readonly'@'%';

-- 显示只读用户
SELECT User, Host FROM mysql.user WHERE User = 'app_readonly';

FLUSH PRIVILEGES;

EOSQL

log_success "只读模式已设置"

echo ""
echo "=========================================="
echo "配置完成"
echo "=========================================="
echo ""
echo "当前状态:"
echo "  ✅ read_only = ON"
echo "  ✅ super_read_only = ON"
echo ""
echo "只读用户已创建:"
echo "  用户名: app_readonly"
echo "  密码: readonly_password_2025"
echo "  权限: SELECT (全部数据库)"
echo ""
echo "应用配置建议:"
echo "  # 连接腾讯云从库（只读）"
echo "  DB_SLAVE_HOST=101.33.251.158"
echo "  DB_SLAVE_USER=app_readonly"
echo "  DB_SLAVE_PASSWORD=readonly_password_2025"
echo ""
echo "注意:"
echo "  ⚠️  root用户仍可写入（super权限可绕过只读）"
echo "  ⚠️  应用服务应使用只读用户"
echo "  ⚠️  所有写入操作应连接阿里云主库"
echo ""

