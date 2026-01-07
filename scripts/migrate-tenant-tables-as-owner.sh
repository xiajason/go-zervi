#!/bin/bash

# 数据库迁移脚本 - 使用表所有者或postgres超级用户
# 注意：表的所有者是szjason72，需要使用该用户或postgres超级用户执行

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "=========================================="
echo "GoZervi 数据库迁移 - 租户表和多租户支持"
echo "=========================================="
echo ""
echo "⚠️  注意: 表的所有者是szjason72，需要使用该用户或postgres超级用户执行"
echo ""

# 数据库连接参数（使用postgres超级用户或表所有者）
DB_HOST="${POSTGRESQL_HOST:-localhost}"
DB_PORT="${POSTGRESQL_PORT:-5432}"
DB_NAME="${POSTGRESQL_DATABASE:-zervigo_mvp}"
DB_USER="${1:-postgres}"  # 第一个参数指定用户，默认postgres

echo "数据库配置:"
echo "  主机: $DB_HOST"
echo "  端口: $DB_PORT"
echo "  数据库: $DB_NAME"
echo "  用户: $DB_USER"
echo ""

# 提示输入密码
read -sp "请输入数据库密码（$DB_USER）: " DB_PASSWORD
echo ""

if [ -n "$DB_PASSWORD" ]; then
    export PGPASSWORD="$DB_PASSWORD"
    CONN_STRING="postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME"
else
    CONN_STRING="postgresql://$DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
fi

# 测试数据库连接
echo "测试数据库连接..."
if ! psql "$CONN_STRING" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "❌ 错误: 无法连接到数据库"
    echo "请检查数据库配置和连接信息"
    exit 1
fi
echo "✓ 数据库连接成功"
echo ""

# 执行手动SQL
echo "=========================================="
echo "为现有表添加tenant_id字段"
echo "=========================================="
echo ""

MANUAL_SQL="scripts/migrate-tenant-tables-manual.sql"

if [ ! -f "$MANUAL_SQL" ]; then
    echo "❌ 错误: 找不到SQL文件: $MANUAL_SQL"
    exit 1
fi

echo "执行SQL文件: $MANUAL_SQL"
if psql "$CONN_STRING" -f "$MANUAL_SQL"; then
    echo "✓ tenant_id字段添加成功"
else
    echo "⚠️  警告: 部分操作可能失败，请检查错误信息"
fi
echo ""

# 验证结果
echo "=========================================="
echo "验证迁移结果"
echo "=========================================="
echo ""

echo "检查tenant_id字段..."
psql "$CONN_STRING" -c "SELECT table_name, column_name FROM information_schema.columns WHERE table_name LIKE 'zervigo_%' AND column_name = 'tenant_id' ORDER BY table_name;"

echo ""
echo "=========================================="
echo "迁移完成！"
echo "=========================================="

