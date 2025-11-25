#!/bin/bash

echo "========================================"
echo "🧪 测试前端数据格式适配修复"
echo "========================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 测试1: 字段配置接口
echo -e "${YELLOW}测试1: 字段配置接口${NC}"
echo "---------------------------------------"
RESULT=$(curl -s -X POST http://localhost:9000/api/v1/model_field/index \
  -H "Content-Type: application/json" \
  -d '{"data":{"table_name":"admin","page":1,"page_size":100}}' | jq -r '.code')

if [ "$RESULT" == "0" ]; then
    echo -e "${GREEN}✅ 字段配置接口正常 (code: 0)${NC}"
    FIELD_COUNT=$(curl -s -X POST http://localhost:9000/api/v1/model_field/index \
      -H "Content-Type: application/json" \
      -d '{"data":{"table_name":"admin","page":1,"page_size":100}}' | jq -r '.data.list | length')
    echo -e "   字段数量: $FIELD_COUNT 个"
else
    echo -e "${RED}❌ 字段配置接口异常 (code: $RESULT)${NC}"
fi
echo ""

# 测试2: 用户列表接口
echo -e "${YELLOW}测试2: 用户列表接口${NC}"
echo "---------------------------------------"
RESULT=$(curl -s -X POST http://localhost:9000/api/v1/admin/index \
  -H "Content-Type: application/json" \
  -d '{"data":{"table_name":"admin","page":1,"page_size":20}}' | jq -r '.code')

if [ "$RESULT" == "0" ]; then
    echo -e "${GREEN}✅ 用户列表接口正常 (code: 0)${NC}"
    USER_COUNT=$(curl -s -X POST http://localhost:9000/api/v1/admin/index \
      -H "Content-Type: application/json" \
      -d '{"data":{"table_name":"admin","page":1,"page_size":20}}' | jq -r '.data.list | length')
    echo -e "   用户数量: $USER_COUNT 个"
else
    echo -e "${RED}❌ 用户列表接口异常 (code: $RESULT)${NC}"
fi
echo ""

# 测试3: 角色列表接口
echo -e "${YELLOW}测试3: 角色列表接口${NC}"
echo "---------------------------------------"
RESULT=$(curl -s -X POST http://localhost:9000/api/v1/roles/index \
  -H "Content-Type: application/json" \
  -d '{"data":{"table_name":"roles","page":1,"page_size":20}}' | jq -r '.code')

if [ "$RESULT" == "0" ]; then
    echo -e "${GREEN}✅ 角色列表接口正常 (code: 0)${NC}"
    ROLE_COUNT=$(curl -s -X POST http://localhost:9000/api/v1/roles/index \
      -H "Content-Type: application/json" \
      -d '{"data":{"table_name":"roles","page":1,"page_size":20}}' | jq -r '.data.list | length')
    echo -e "   角色数量: $ROLE_COUNT 个"
else
    echo -e "${RED}❌ 角色列表接口异常 (code: $RESULT)${NC}"
fi
echo ""

# 测试4: 权限列表接口
echo -e "${YELLOW}测试4: 权限列表接口${NC}"
echo "---------------------------------------"
RESULT=$(curl -s -X POST http://localhost:9000/api/v1/permissions/index \
  -H "Content-Type: application/json" \
  -d '{"data":{"table_name":"permissions","page":1,"page_size":20}}' | jq -r '.code')

if [ "$RESULT" == "0" ]; then
    echo -e "${GREEN}✅ 权限列表接口正常 (code: 0)${NC}"
    PERM_COUNT=$(curl -s -X POST http://localhost:9000/api/v1/permissions/index \
      -H "Content-Type: application/json" \
      -d '{"data":{"table_name":"permissions","page":1,"page_size":20}}' | jq -r '.data.list | length')
    echo -e "   权限数量: $PERM_COUNT 个"
else
    echo -e "${RED}❌ 权限列表接口异常 (code: $RESULT)${NC}"
fi
echo ""

# 测试5: 检查响应格式
echo -e "${YELLOW}测试5: 响应格式验证${NC}"
echo "---------------------------------------"
echo "检查响应格式是否符合标准..."
SAMPLE=$(curl -s -X POST http://localhost:9000/api/v1/admin/index \
  -H "Content-Type: application/json" \
  -d '{"data":{"table_name":"admin","page":1,"page_size":20}}')

HAS_CODE=$(echo $SAMPLE | jq 'has("code")')
HAS_DATA=$(echo $SAMPLE | jq 'has("data")')
HAS_LIST=$(echo $SAMPLE | jq '.data | has("list")')
HAS_TOTAL=$(echo $SAMPLE | jq '.data | has("total")')

if [ "$HAS_CODE" == "true" ] && [ "$HAS_DATA" == "true" ] && [ "$HAS_LIST" == "true" ] && [ "$HAS_TOTAL" == "true" ]; then
    echo -e "${GREEN}✅ 响应格式符合标准${NC}"
    echo "   结构: {code, data: {list, total}}"
else
    echo -e "${RED}❌ 响应格式不符合标准${NC}"
    echo "   当前结构:"
    echo $SAMPLE | jq '.| keys'
fi
echo ""

echo "========================================"
echo -e "${GREEN}🎉 测试完成！${NC}"
echo "========================================"
echo ""
echo "接下来请在浏览器中测试："
echo "1. 访问 http://localhost:8081"
echo "2. 使用 admin/admin123 登录"
echo "3. 查看用户管理、角色管理、权限管理页面"
echo "4. 检查浏览器控制台是否还有 TypeError"
echo ""

