#!/bin/bash

# 数据库迁移脚本 - 租户表和多租户支持
# 使用方法: ./scripts/migrate-tenant-tables.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "=========================================="
echo "GoZervi 数据库迁移 - 租户表和多租户支持"
echo "=========================================="
echo ""

# 加载环境变量
if [ -f "configs/local.env" ]; then
    set -a
    source <(cat configs/local.env | grep -v "^#" | grep -v "^$" | sed 's/^/export /')
    set +a
fi

# 数据库连接参数
DB_HOST="${POSTGRESQL_HOST:-localhost}"
DB_PORT="${POSTGRESQL_PORT:-5432}"
DB_NAME="${POSTGRESQL_DATABASE:-zervigo_mvp}"
DB_USER="${POSTGRESQL_USER:-postgres}"
DB_PASSWORD="${POSTGRESQL_PASSWORD:-}"

echo "数据库配置:"
echo "  主机: $DB_HOST"
echo "  端口: $DB_PORT"
echo "  数据库: $DB_NAME"
echo "  用户: $DB_USER"
echo ""

# 检查psql是否可用
if ! command -v psql &> /dev/null; then
    echo "❌ 错误: psql 命令未找到"
    echo "请安装 PostgreSQL 客户端工具"
    exit 1
fi

# 构建连接字符串
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

# Step 1: 执行租户表创建
echo "=========================================="
echo "Step 1: 创建租户表"
echo "=========================================="
echo ""

TENANT_TABLES_SQL="databases/postgres/init/03-tenant-tables.sql"

if [ ! -f "$TENANT_TABLES_SQL" ]; then
    echo "❌ 错误: 找不到SQL文件: $TENANT_TABLES_SQL"
    exit 1
fi

echo "执行SQL文件: $TENANT_TABLES_SQL"
if psql "$CONN_STRING" -f "$TENANT_TABLES_SQL"; then
    echo "✓ 租户表创建成功"
else
    echo "❌ 错误: 租户表创建失败"
    exit 1
fi
echo ""

# Step 2: 为现有表添加tenant_id
echo "=========================================="
echo "Step 2: 为现有表添加tenant_id字段"
echo "=========================================="
echo ""

ADD_TENANT_ID_SQL="databases/postgres/migrations/add_tenant_id_to_tables.sql"

if [ ! -f "$ADD_TENANT_ID_SQL" ]; then
    echo "❌ 错误: 找不到SQL文件: $ADD_TENANT_ID_SQL"
    exit 1
fi

echo "执行SQL文件: $ADD_TENANT_ID_SQL"
if psql "$CONN_STRING" -f "$ADD_TENANT_ID_SQL"; then
    echo "✓ tenant_id字段添加成功"
else
    echo "❌ 错误: tenant_id字段添加失败"
    exit 1
fi
echo ""

# Step 3: 验证迁移结果
echo "=========================================="
echo "Step 3: 验证迁移结果"
echo "=========================================="
echo ""

echo "检查租户表..."
TENANT_COUNT=$(psql "$CONN_STRING" -t -c "SELECT COUNT(*) FROM zervigo_tenants;" 2>/dev/null | xargs)
if [ "$TENANT_COUNT" -ge 1 ]; then
    echo "✓ 租户表存在，当前租户数: $TENANT_COUNT"
else
    echo "⚠️  警告: 租户表可能为空"
fi

echo ""
echo "检查用户-租户关联表..."
USER_TENANT_COUNT=$(psql "$CONN_STRING" -t -c "SELECT COUNT(*) FROM zervigo_user_tenants;" 2>/dev/null | xargs)
if [ -n "$USER_TENANT_COUNT" ]; then
    echo "✓ 用户-租户关联表存在，当前关联数: $USER_TENANT_COUNT"
else
    echo "⚠️  警告: 用户-租户关联表可能为空"
fi

echo ""
echo "检查业务表的tenant_id字段..."
TABLES=("zervigo_jobs" "zervigo_user_profiles" "zervigo_companies" "zervigo_resumes")
for table in "${TABLES[@]}"; do
    if psql "$CONN_STRING" -t -c "SELECT column_name FROM information_schema.columns WHERE table_name = '$table' AND column_name = 'tenant_id';" 2>/dev/null | grep -q tenant_id; then
        echo "✓ $table 表已包含 tenant_id 字段"
    else
        echo "⚠️  警告: $table 表未找到 tenant_id 字段"
    fi
done

echo ""
echo "=========================================="
echo "迁移完成！"
echo "=========================================="
echo ""
echo "下一步操作:"
echo "1. 验证数据: SELECT * FROM zervigo_tenants;"
echo "2. 检查索引: \\d zervigo_jobs"
echo "3. 更新BaseModel以支持tenant_id"
echo ""

