#!/bin/bash

echo "🧪 Zervigo MVP微服务集群测试..."

# 进入MVP目录
cd "$(dirname "$0")/.."

# 测试服务健康状态
echo "🏥 测试服务健康状态..."
echo ""

# 中央大脑测试
echo "1. 测试中央大脑..."
if curl -s http://localhost:9000/health | grep -q "UP"; then
    echo "   ✅ 中央大脑服务正常"
else
    echo "   ❌ 中央大脑服务异常"
fi

# 认证服务测试
echo "2. 测试认证服务..."
if curl -s http://localhost:8207/health | grep -q "UP"; then
    echo "   ✅ 认证服务正常"
else
    echo "   ❌ 认证服务异常"
fi

# AI服务测试
echo "3. 测试AI服务..."
if curl -s http://localhost:8100/health | grep -q "UP"; then
    echo "   ✅ AI服务正常"
else
    echo "   ❌ AI服务异常"
fi

# 区块链服务测试
echo "4. 测试区块链服务..."
if curl -s http://localhost:8208/health | grep -q "UP"; then
    echo "   ✅ 区块链服务正常"
else
    echo "   ❌ 区块链服务异常"
fi

# 用户服务测试
echo "5. 测试用户服务..."
if curl -s http://localhost:8082/health | grep -q "UP"; then
    echo "   ✅ 用户服务正常"
else
    echo "   ❌ 用户服务异常"
fi

# 职位服务测试
echo "6. 测试职位服务..."
if curl -s http://localhost:8084/health | grep -q "UP"; then
    echo "   ✅ 职位服务正常"
else
    echo "   ❌ 职位服务异常"
fi

# 简历服务测试
echo "7. 测试简历服务..."
if curl -s http://localhost:8085/health | grep -q "UP"; then
    echo "   ✅ 简历服务正常"
else
    echo "   ❌ 简历服务异常"
fi

# 企业服务测试
echo "8. 测试企业服务..."
if curl -s http://localhost:8083/health | grep -q "UP"; then
    echo "   ✅ 企业服务正常"
else
    echo "   ❌ 企业服务异常"
fi

echo ""
echo "📊 服务状态总结："
docker-compose -f docker/docker-compose.yml ps

echo ""
echo "🔍 详细测试命令："
echo "   # 测试中央大脑路由"
echo "   curl http://localhost:9000/api/v1/auth/health"
echo ""
echo "   # 测试认证接口"
echo "   curl -X POST http://localhost:9000/api/v1/auth/login \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"username\":\"test\",\"password\":\"test\"}'"
echo ""
echo "   # 测试AI接口"
echo "   curl http://localhost:9000/api/v1/ai/health"
echo ""
echo "   # 测试区块链接口"
echo "   curl http://localhost:9000/api/v1/blockchain/health"
echo ""

echo "✅ MVP测试完成！"
