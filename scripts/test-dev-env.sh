# Zervigo 二次开发环境快速测试脚本

echo "🧪 Zervigo 二次开发环境测试"
echo "================================"

# 测试认证服务
echo "🔐 测试认证服务..."
echo "健康检查:"
curl -s http://localhost:8207/health | jq . || echo "❌ 认证服务健康检查失败"

echo ""
echo "API端点测试:"
echo "获取角色列表:"
curl -s http://localhost:8207/api/v1/auth/roles | jq . || echo "❌ 角色列表获取失败"

echo ""
echo "获取权限列表:"
curl -s http://localhost:8207/api/v1/auth/permissions | jq . || echo "❌ 权限列表获取失败"

echo ""
echo "================================"

# 测试AI服务
echo "🤖 测试AI服务..."
echo "健康检查:"
curl -s http://localhost:8100/health | jq . || echo "❌ AI服务健康检查失败"

echo ""
echo "API端点测试:"
echo "获取用户信息 (需要认证):"
curl -s -H "Authorization: Bearer test-token" http://localhost:8100/api/v1/ai/user-info | jq . || echo "❌ 用户信息获取失败"

echo ""
echo "获取权限列表 (需要认证):"
curl -s -H "Authorization: Bearer test-token" http://localhost:8100/api/v1/ai/permissions | jq . || echo "❌ 权限列表获取失败"

echo ""
echo "================================"

# 测试数据库连接
echo "🗄️  测试数据库连接..."
echo "MySQL连接测试:"
docker exec zervigo-mysql mysql -uroot -pdev_password -e "SELECT 1;" || echo "❌ MySQL连接失败"

echo ""
echo "PostgreSQL连接测试:"
docker exec zervigo-postgres psql -U postgres -d jobfirst_vector -c "SELECT 1;" || echo "❌ PostgreSQL连接失败"

echo ""
echo "Redis连接测试:"
docker exec zervigo-redis redis-cli ping || echo "❌ Redis连接失败"

echo ""
echo "🎉 测试完成！"
echo "================================"
