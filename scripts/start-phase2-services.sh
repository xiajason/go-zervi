#!/bin/bash
# 第二阶段业务服务启动脚本
# 用途: 启动所有业务层微服务（用户、简历、职位、公司）

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🚀 启动第二阶段业务服务...${NC}"

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 检查Consul是否运行
if ! curl -s http://localhost:8500/v1/status/leader > /dev/null 2>&1; then
    echo -e "${RED}❌ Consul未运行，请先启动Consul:${NC}"
    echo "   ./scripts/start-consul.sh"
    exit 1
fi

echo -e "${GREEN}✅ Consul运行正常${NC}"

# 检查Central Brain是否运行
if ! curl -s http://localhost:9000/health > /dev/null 2>&1; then
    echo -e "${RED}❌ Central Brain未运行，请先启动Central Brain:${NC}"
    echo "   ./scripts/start-central-brain.sh"
    exit 1
fi

echo -e "${GREEN}✅ Central Brain运行正常${NC}"

# 启动函数
start_service() {
    local service_name=$1
    local service_dir=$2
    local port=$3
    
    echo ""
    echo -e "${YELLOW}📦 启动 $service_name...${NC}"
    
    # 检查是否已经运行
    if pgrep -f "$service_name" > /dev/null; then
        echo "⚠️  $service_name 已经在运行中"
        return 0
    fi
    
    cd "$PROJECT_ROOT/$service_dir"
    
    # 构建服务（排除simple_main.go等测试文件）
    if go build -o "$service_name" -tags exclude_simple_main $(ls *.go 2>/dev/null | grep -v "^simple_main.go$" | grep -v "^.*_test.go$" | tr '\n' ' ') 2>&1 | tee "/tmp/${service_name}_build.log"; then
        echo "✅ 构建成功"
    else
        # 如果上面的命令失败，尝试只编译main.go
        if go build -o "$service_name" main.go 2>&1 | tee "/tmp/${service_name}_build.log"; then
            echo "✅ 构建成功（使用main.go）"
        else
            echo -e "${RED}❌ 构建失败，查看错误:${NC}"
            cat "/tmp/${service_name}_build.log"
            return 1
        fi
    fi
    
    # 启动服务
    nohup "./$service_name" > "/tmp/${service_name}.log" 2>&1 &
    SERVICE_PID=$!
    
    # 等待启动
    sleep 3
    
    # 检查是否启动成功
    if ps -p $SERVICE_PID > /dev/null; then
        echo "✅ $service_name 已启动 (PID: $SERVICE_PID)"
        echo "   📊 日志文件: /tmp/${service_name}.log"
        echo "   🌐 访问地址: http://localhost:$port"
        echo "   🏥 健康检查: http://localhost:$port/health"
        
        # 等待一下然后测试健康检查
        sleep 2
        if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
            echo "✅ 健康检查通过"
        else
            echo "⚠️  健康检查失败，请查看日志: tail -f /tmp/${service_name}.log"
        fi
    else
        echo -e "${RED}❌ $service_name 启动失败${NC}"
        echo "   查看日志: tail -f /tmp/${service_name}.log"
        return 1
    fi
}

# 启动各个业务服务
start_service "user-service" "services/core/user" 8082
start_service "resume-service" "services/business/resume" 8085
start_service "job-service" "services/business/job" 8084
start_service "company-service" "services/business/company" 8083

echo ""
echo "=========================================="
echo "📋 业务服务启动完成！"
echo "=========================================="
echo ""
echo "🌐 服务访问地址:"
echo "   用户服务: http://localhost:8082"
echo "   简历服务: http://localhost:8085"
echo "   职位服务: http://localhost:8084"
echo "   公司服务: http://localhost:8083"
echo ""
echo "🔍 通过Central Brain访问:"
echo "   http://localhost:9000/api/v1/user/**"
echo "   http://localhost:9000/api/v1/resume/**"
echo "   http://localhost:9000/api/v1/job/**"
echo "   http://localhost:9000/api/v1/company/**"
echo ""
echo "📊 查看服务注册:"
echo "   curl http://localhost:8500/v1/agent/services | jq ."
echo ""
echo "✅ 第二阶段业务服务启动完成！"
