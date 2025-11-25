#!/bin/bash

# PostgreSQL 数据库初始化脚本
# 用于Zervigo项目的PostgreSQL主数据库初始化

set -e

echo "🚀 开始PostgreSQL数据库初始化..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 检查PostgreSQL容器是否已存在
if docker ps -a --format "table {{.Names}}" | grep -q "zervigo-postgres-mvp"; then
    echo "📦 发现已存在的PostgreSQL容器，正在停止并删除..."
    docker stop zervigo-postgres-mvp > /dev/null 2>&1 || true
    docker rm zervigo-postgres-mvp > /dev/null 2>&1 || true
fi

# 检查PostgreSQL数据卷是否已存在
if docker volume ls --format "table {{.Name}}" | grep -q "zervigo-postgres-mvp"; then
    echo "🗑️ 发现已存在的数据卷，正在删除..."
    docker volume rm zervigo-postgres-mvp > /dev/null 2>&1 || true
fi

echo "📦 启动PostgreSQL容器..."
docker-compose -f docker/docker-compose-postgres.yml up -d postgres

echo "⏳ 等待PostgreSQL启动..."
sleep 10

# 检查PostgreSQL是否健康
echo "🔍 检查PostgreSQL健康状态..."
for i in {1..30}; do
    if docker exec zervigo-postgres-mvp pg_isready -U postgres -d zervigo_mvp > /dev/null 2>&1; then
        echo "✅ PostgreSQL已就绪！"
        break
    fi
    echo "⏳ 等待PostgreSQL启动... ($i/30)"
    sleep 2
done

# 验证数据库连接
echo "🔍 验证数据库连接..."
if docker exec zervigo-postgres-mvp psql -U postgres -d zervigo_mvp -c "SELECT version();" > /dev/null 2>&1; then
    echo "✅ 数据库连接成功！"
else
    echo "❌ 数据库连接失败！"
    exit 1
fi

# 检查表是否已创建
echo "🔍 检查数据库表..."
TABLE_COUNT=$(docker exec zervigo-postgres-mvp psql -U postgres -d zervigo_mvp -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" | tr -d ' ')

if [ "$TABLE_COUNT" -gt 0 ]; then
    echo "✅ 数据库表已创建 ($TABLE_COUNT 个表)"
else
    echo "❌ 数据库表未创建！"
    exit 1
fi

# 检查默认用户
echo "🔍 检查默认用户..."
USER_COUNT=$(docker exec zervigo-postgres-mvp psql -U postgres -d zervigo_mvp -t -c "SELECT COUNT(*) FROM users WHERE username = 'admin';" | tr -d ' ')

if [ "$USER_COUNT" -eq 1 ]; then
    echo "✅ 默认管理员用户已创建"
else
    echo "❌ 默认管理员用户未创建！"
    exit 1
fi

# 检查角色
echo "🔍 检查用户角色..."
ROLE_COUNT=$(docker exec zervigo-postgres-mvp psql -U postgres -d zervigo_mvp -t -c "SELECT COUNT(*) FROM user_roles;" | tr -d ' ')

if [ "$ROLE_COUNT" -gt 0 ]; then
    echo "✅ 用户角色已创建 ($ROLE_COUNT 个角色)"
else
    echo "❌ 用户角色未创建！"
    exit 1
fi

# 显示数据库信息
echo ""
echo "🎉 PostgreSQL数据库初始化完成！"
echo "=================================="
echo "📊 数据库信息:"
echo "  数据库名: zervigo_mvp"
echo "  用户名: postgres"
echo "  密码: dev_password"
echo "  端口: 5432"
echo "  连接字符串: postgres://postgres:dev_password@localhost:5432/zervigo_mvp"
echo ""
echo "👤 默认管理员账号:"
echo "  用户名: admin"
echo "  密码: admin123"
echo "  邮箱: admin@zervigo.com"
echo ""
echo "🔧 管理命令:"
echo "  连接数据库: docker exec -it zervigo-postgres-mvp psql -U postgres -d zervigo_mvp"
echo "  查看表: docker exec zervigo-postgres-mvp psql -U postgres -d zervigo_mvp -c '\dt'"
echo "  查看用户: docker exec zervigo-postgres-mvp psql -U postgres -d zervigo_mvp -c 'SELECT username, email FROM users;'"
echo ""
echo "✅ 初始化完成！可以开始启动微服务了。"
