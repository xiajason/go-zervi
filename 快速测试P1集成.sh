#!/bin/bash

# 快速测试P1集成和自适应功能
# 使用方法：bash 快速测试P1集成.sh

echo "════════════════════════════════════════════"
echo "  🧪 快速测试P1集成和自适应功能"
echo "════════════════════════════════════════════"
echo ""

# 1. 检查前端是否运行
echo "【步骤1】检查前端运行状态..."
if lsof -ti:3000 > /dev/null; then
    echo "✅ 前端已在运行 (端口3000)"
else
    echo "❌ 前端未运行，正在启动..."
    cd zervigo-admin
    npm run dev &
    echo "⏳ 等待前端启动（10秒）..."
    sleep 10
    cd ..
fi
echo ""

# 2. 检查P0+P1服务
echo "【步骤2】检查P0+P1服务状态..."
check_service() {
    local service=$1
    local port=$2
    if lsof -ti:$port > /dev/null; then
        echo "✅ $service (端口 $port)"
        return 0
    else
        echo "❌ $service (端口 $port) - 未运行"
        return 1
    fi
}

check_service "Central Brain" 9000
check_service "Auth Service" 8081
check_service "Router Service" 8082
check_service "Permission Service" 8083
check_service "User Service" 8088
echo ""

# 3. 检查P2服务
echo "【步骤3】检查P2服务状态..."
job_running=0
resume_running=0
company_running=0

if check_service "Job Service" 8084; then job_running=1; fi
if check_service "Resume Service" 8085; then resume_running=1; fi
if check_service "Company Service" 8086; then company_running=1; fi
echo ""

# 4. 识别服务组合
echo "【步骤4】识别当前服务组合..."
total_p2=$((job_running + resume_running + company_running))

if [ $total_p2 -eq 0 ]; then
    echo "📊 当前组合: minimal（基础模式）"
    echo "   只有系统管理功能"
elif [ $total_p2 -eq 3 ]; then
    echo "📊 当前组合: all_services（完整模式）"
    echo "   所有业务功能可用"
elif [ $job_running -eq 1 ] && [ $resume_running -eq 0 ] && [ $company_running -eq 0 ]; then
    echo "📊 当前组合: job_only（职位模式）"
elif [ $resume_running -eq 1 ] && [ $job_running -eq 0 ] && [ $company_running -eq 0 ]; then
    echo "📊 当前组合: resume_only（简历模式）"
elif [ $company_running -eq 1 ] && [ $job_running -eq 0 ] && [ $resume_running -eq 0 ]; then
    echo "📊 当前组合: company_only（企业模式）"
elif [ $job_running -eq 1 ] && [ $resume_running -eq 1 ]; then
    echo "📊 当前组合: job_resume（职位+简历模式）"
elif [ $job_running -eq 1 ] && [ $company_running -eq 1 ]; then
    echo "📊 当前组合: job_company（职位+企业模式）"
elif [ $resume_running -eq 1 ] && [ $company_running -eq 1 ]; then
    echo "📊 当前组合: resume_company（简历+企业模式）"
fi
echo ""

# 5. 测试服务组合API
echo "【步骤5】测试服务组合API..."
response=$(curl -s http://localhost:9000/api/v1/router/service-combination)
if [ $? -eq 0 ]; then
    echo "✅ API响应成功"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
else
    echo "❌ API调用失败"
fi
echo ""

# 6. 显示测试说明
echo "════════════════════════════════════════════"
echo "  📖 接下来的测试步骤："
echo "════════════════════════════════════════════"
echo ""
echo "1. 打开浏览器访问："
echo "   http://localhost:3000"
echo ""
echo "2. 使用以下凭证登录："
echo "   用户名: admin"
echo "   密码: admin123"
echo ""
echo "3. 查看菜单（应该根据P2服务组合显示）："
if [ $job_running -eq 1 ]; then
    echo "   ✅ 应该看到：职位管理"
else
    echo "   ❌ 不应该看到：职位管理"
fi

if [ $resume_running -eq 1 ]; then
    echo "   ✅ 应该看到：简历管理"
else
    echo "   ❌ 不应该看到：简历管理"
fi

if [ $company_running -eq 1 ]; then
    echo "   ✅ 应该看到：企业管理"
else
    echo "   ❌ 不应该看到：企业管理"
fi
echo ""
echo "4. 点击'服务状态'菜单查看详细状态"
echo ""
echo "5. 测试自适应："
echo "   • 启动新的P2服务"
echo "   • 等待30秒或刷新页面"
echo "   • 观察菜单自动变化"
echo ""

# 7. 提供快捷命令
echo "════════════════════════════════════════════"
echo "  🚀 快捷操作命令："
echo "════════════════════════════════════════════"
echo ""
echo "启动Job Service:"
echo "  cd services/business/job && go run main.go &"
echo ""
echo "启动Resume Service:"
echo "  cd services/business/resume && go run main.go &"
echo ""
echo "启动Company Service:"
echo "  cd services/business/company && go run main.go &"
echo ""
echo "启动所有P2服务:"
echo "  START_P2=yes bash 完整服务协同启动.sh"
echo ""
echo "════════════════════════════════════════════"
echo "  🎊 测试愉快！"
echo "════════════════════════════════════════════"

