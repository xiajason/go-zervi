#!/bin/bash

# 本地PostgreSQL数据库初始化脚本
# 用于Zervigo项目的本地PostgreSQL主数据库初始化

set -e

echo "🚀 开始本地PostgreSQL数据库初始化..."

# 检查PostgreSQL是否运行
if ! brew services list | grep postgresql@14 | grep -q "started"; then
    echo "📦 启动PostgreSQL服务..."
    brew services start postgresql@14
    sleep 3
fi

# 检查PostgreSQL连接
echo "🔍 检查PostgreSQL连接..."
if ! psql -U $(whoami) -d postgres -c "SELECT 1;" > /dev/null 2>&1; then
    echo "❌ PostgreSQL连接失败！"
    exit 1
fi

echo "✅ PostgreSQL连接成功！"

# 检查数据库是否存在
echo "🔍 检查数据库是否存在..."
if psql -U $(whoami) -d postgres -c "SELECT 1 FROM pg_database WHERE datname = 'zervigo_mvp';" | grep -q "1 row"; then
    echo "📦 数据库已存在，正在删除..."
    psql -U $(whoami) -d postgres -c "DROP DATABASE IF EXISTS zervigo_mvp;"
fi

# 创建数据库
echo "📦 创建数据库..."
psql -U $(whoami) -d postgres -c "CREATE DATABASE zervigo_mvp;"

# 执行初始化脚本
echo "📦 执行初始化脚本..."
psql -U $(whoami) -d zervigo_mvp < databases/postgres/init/01-init-schema.sql

# 验证数据库
echo "🔍 验证数据库..."
TABLE_COUNT=$(psql -U $(whoami) -d zervigo_mvp -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" | tr -d ' ')

if [ "$TABLE_COUNT" -gt 0 ]; then
    echo "✅ 数据库表已创建 ($TABLE_COUNT 个表)"
else
    echo "❌ 数据库表未创建！"
    exit 1
fi

# 检查默认用户
echo "🔍 检查默认用户..."
USER_COUNT=$(psql -U $(whoami) -d zervigo_mvp -t -c "SELECT COUNT(*) FROM users WHERE username = 'admin';" | tr -d ' ')

if [ "$USER_COUNT" -eq 1 ]; then
    echo "✅ 默认管理员用户已创建"
else
    echo "❌ 默认管理员用户未创建！"
    exit 1
fi

# 检查角色
echo "🔍 检查用户角色..."
ROLE_COUNT=$(psql -U $(whoami) -d zervigo_mvp -t -c "SELECT COUNT(*) FROM user_roles;" | tr -d ' ')

if [ "$ROLE_COUNT" -gt 0 ]; then
    echo "✅ 用户角色已创建 ($ROLE_COUNT 个角色)"
else
    echo "❌ 用户角色未创建！"
    exit 1
fi

# 显示数据库信息
echo ""
echo "🎉 本地PostgreSQL数据库初始化完成！"
echo "=================================="
echo "📊 数据库信息:"
echo "  数据库名: zervigo_mvp"
echo "  用户名: $(whoami)"
echo "  端口: 5432"
echo "  连接字符串: postgres://$(whoami)@localhost:5432/zervigo_mvp"
echo ""
echo "👤 默认管理员账号:"
echo "  用户名: admin"
echo "  密码: admin123"
echo "  邮箱: admin@zervigo.com"
echo ""
echo "🔧 管理命令:"
echo "  连接数据库: psql -U $(whoami) -d zervigo_mvp"
echo "  查看表: psql -U $(whoami) -d zervigo_mvp -c '\dt'"
echo "  查看用户: psql -U $(whoami) -d zervigo_mvp -c 'SELECT username, email FROM users;'"
echo ""
echo "✅ 本地PostgreSQL初始化完成！可以开始开发了。"
