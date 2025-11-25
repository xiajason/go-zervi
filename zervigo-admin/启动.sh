#!/bin/bash

echo "╔════════════════════════════════════════════╗"
echo "║  🧠 Zervigo Admin - 自主可控管理平台      ║"
echo "╚════════════════════════════════════════════╝"
echo ""

# 检查后端服务
echo "📊 检查后端服务..."
curl -s http://localhost:9000/health > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "  ✅ Central Brain (9000) 运行中"
else
    echo "  ❌ Central Brain未运行，请先启动后端"
    echo "     cd ../shared/central-brain && go run . &"
    exit 1
fi

curl -s http://localhost:8207/health > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "  ✅ Auth Service (8207) 运行中"
else
    echo "  ⚠️  Auth Service未运行（可选）"
fi

echo ""
echo "🚀 启动前端服务..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

npm run dev

echo ""
echo "✅ 服务已启动！"
echo ""
echo "📖 访问地址: http://localhost:3000"
echo "🔑 登录信息: admin / admin123"
echo ""





