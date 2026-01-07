#!/bin/bash

echo "🚀 启动Central Brain服务..."

cd /Users/szjason72/gozervi/zervigo.demo/shared/central-brain

# 停止旧进程
pkill -f "central-brain" 2>/dev/null
sleep 2

# 编译
echo "📦 编译..."
go build .

# 启动
echo "🌟 启动服务..."
./central-brain &
CB_PID=$!
echo "进程PID: $CB_PID"

# 等待启动
echo "⏳ 等待服务启动..."
sleep 10

# 测试API
echo ""
echo "=== 测试用户管理API ==="
curl -s -X POST http://localhost:9000/api/v1/admin/index \
  -H "Content-Type: application/json" \
  -d '{"data":{"table_name":"admin","page":1,"page_size":10}}' \
  | jq '.'

echo ""
echo "=== 测试角色管理API ==="
curl -s -X POST http://localhost:9000/api/v1/roles/index \
  -H "Content-Type: application/json" \
  -d '{"data":{"table_name":"roles","page":1,"page_size":10}}' \
  | jq '.'

echo ""
echo "=== 测试权限管理API ==="
curl -s -X POST http://localhost:9000/api/v1/permissions/index \
  -H "Content-Type: application/json" \
  -d '{"data":{"table_name":"permissions","page":1,"page_size":10}}' \
  | jq '.'

echo ""
echo "✅ Central Brain服务正在运行，PID: $CB_PID"
echo "📊 访问 http://localhost:9000"
echo ""
echo "停止服务命令: kill $CB_PID"

