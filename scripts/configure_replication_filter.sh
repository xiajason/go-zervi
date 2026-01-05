#!/bin/bash
# 配置MySQL主从复制过滤
# 用途: 设置仅复制production数据库，其他数据库独立使用
# 部署位置: 腾讯云服务器 101.33.251.158
# 作者: AI Assistant
# 日期: 2025-10-10

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo ""
echo "=========================================="
echo "配置MySQL主从复制过滤"
echo "=========================================="
echo ""

log_warning "此操作将修改主从复制配置"
log_warning "建议在非业务高峰期执行"
echo ""

read -p "确认继续？(y/N): " confirm
if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    log_info "操作已取消"
    exit 0
fi

echo ""
log_info "开始配置..."

# 检查当前复制状态
log_info "1. 检查当前复制状态"
SLAVE_STATUS=$(docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW SLAVE STATUS\G" 2>/dev/null)

SLAVE_IO=$(echo "$SLAVE_STATUS" | grep "Slave_IO_Running:" | awk '{print $2}')
SLAVE_SQL=$(echo "$SLAVE_STATUS" | grep "Slave_SQL_Running:" | awk '{print $2}')

if [ "$SLAVE_IO" != "Yes" ] || [ "$SLAVE_SQL" != "Yes" ]; then
    log_error "主从复制当前未正常运行"
    log_error "IO: $SLAVE_IO, SQL: $SLAVE_SQL"
    log_error "请先修复主从复制"
    exit 1
fi

log_success "主从复制运行正常"

# 配置复制过滤
log_info "2. 配置复制过滤"

docker exec test-mysql mysql -uroot -ptest_mysql_password << 'EOSQL'

-- 停止复制
STOP SLAVE;

-- 配置复制过滤
-- 仅复制production数据库（如果存在）
-- 忽略basic, professional, future（独立使用）
CHANGE REPLICATION FILTER
  REPLICATE_IGNORE_DB = (jobfirst_basic, jobfirst_professional, jobfirst_future);

-- 启动复制
START SLAVE;

-- 等待复制恢复
SELECT SLEEP(2);

EOSQL

log_success "复制过滤已配置"

# 验证配置
log_info "3. 验证配置"

SLAVE_STATUS=$(docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW SLAVE STATUS\G" 2>/dev/null)

echo "$SLAVE_STATUS" | grep -E "Slave_IO_Running|Slave_SQL_Running|Replicate_Ignore_DB|Seconds_Behind_Master"

SLAVE_IO=$(echo "$SLAVE_STATUS" | grep "Slave_IO_Running:" | awk '{print $2}')
SLAVE_SQL=$(echo "$SLAVE_STATUS" | grep "Slave_SQL_Running:" | awk '{print $2}')

if [ "$SLAVE_IO" = "Yes" ] && [ "$SLAVE_SQL" = "Yes" ]; then
    log_success "主从复制已恢复正常"
else
    log_error "主从复制异常，请检查"
    exit 1
fi

echo ""
echo "=========================================="
log_success "配置完成！"
echo "=========================================="
echo ""
echo "当前配置:"
echo "  复制的数据库: jobfirst_production (如果存在于主库)"
echo "  忽略的数据库: jobfirst_basic, jobfirst_professional, jobfirst_future"
echo ""
echo "使用规则:"
echo "  ✅ jobfirst_basic/professional/future - 可以在腾讯云独立使用"
echo "  ✅ 这些数据库的操作不会影响阿里云"
echo "  ✅ 也不会从阿里云同步数据"
echo ""
echo "  ⚠️  jobfirst_production - 从阿里云同步（只读）"
echo "  ⚠️  请使用只读用户访问"
echo ""

