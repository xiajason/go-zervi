#!/bin/bash

# Zervigo功能完整性对比测试脚本
# 测试基础版和专业版的功能完整性

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试结果目录
RESULT_DIR="functional_test_results"
mkdir -p "$RESULT_DIR"

# 测试时间戳
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
RESULT_FILE="$RESULT_DIR/functional_comparison_${TIMESTAMP}.json"

echo -e "${BLUE}=== Zervigo功能完整性对比测试 ===${NC}"
echo "测试时间: $(date)"
echo "结果文件: $RESULT_FILE"
echo ""

# 测试基础版服务
echo -e "${YELLOW}=== 测试基础版服务 ===${NC}"

# 基础版服务列表
BASIC_SERVICES=(
    "Auth Service:8207"
    "User Service:8081"
    "Resume Service:8082"
    "Company Service:8083"
    "Notification Service:8084"
    "Template Service:8085"
    "Statistics Service:8086"
    "Banner Service:8087"
    "Dev Team Service:8088"
    "Job Service:8089"
    "AI Service:8206"
)

BASIC_RESULTS=()

for service_info in "${BASIC_SERVICES[@]}"; do
    IFS=':' read -r service_name port <<< "$service_info"
    echo -e "${BLUE}测试 $service_name (端口 $port)...${NC}"
    
    # 健康检查
    health_status="unhealthy"
    response_time=0
    
    if curl -s --max-time 5 "http://localhost:$port/health" > /dev/null 2>&1; then
        health_status="healthy"
        start_time=$(date +%s%3N)
        curl -s --max-time 5 "http://localhost:$port/health" > /dev/null 2>&1
        end_time=$(date +%s%3N)
        response_time=$((end_time - start_time))
        echo -e "  ${GREEN}✓ 健康检查通过 (响应时间: ${response_time}ms)${NC}"
    else
        echo -e "  ${RED}✗ 健康检查失败${NC}"
    fi
    
    BASIC_RESULTS+=("$service_name:$health_status:$response_time")
done

# 测试专业版服务
echo -e "${YELLOW}=== 测试专业版服务 ===${NC}"

# 专业版服务列表
PRO_SERVICES=(
    "API Gateway:8601"
    "User Service:8602"
    "Resume Service:8603"
    "Company Service:8604"
    "Notification Service:8605"
    "Statistics Service:8606"
    "Multi Database Service:8607"
    "Job Service:8609"
    "Template Service:8611"
    "Banner Service:8612"
    "Dev Team Service:8613"
    "AI Service:8620"
)

PRO_RESULTS=()

for service_info in "${PRO_SERVICES[@]}"; do
    IFS=':' read -r service_name port <<< "$service_info"
    echo -e "${BLUE}测试 $service_name (端口 $port)...${NC}"
    
    # 健康检查
    health_status="unhealthy"
    response_time=0
    
    if curl -s --max-time 5 "http://localhost:$port/health" > /dev/null 2>&1; then
        health_status="healthy"
        start_time=$(date +%s%3N)
        curl -s --max-time 5 "http://localhost:$port/health" > /dev/null 2>&1
        end_time=$(date +%s%3N)
        response_time=$((end_time - start_time))
        echo -e "  ${GREEN}✓ 健康检查通过 (响应时间: ${response_time}ms)${NC}"
    else
        echo -e "  ${RED}✗ 健康检查失败${NC}"
    fi
    
    PRO_RESULTS+=("$service_name:$health_status:$response_time")
done

# 计算统计结果
basic_healthy=0
basic_total=${#BASIC_SERVICES[@]}
basic_avg_response=0

for result in "${BASIC_RESULTS[@]}"; do
    IFS=':' read -r name status response_time <<< "$result"
    if [ "$status" = "healthy" ]; then
        ((basic_healthy++))
        basic_avg_response=$((basic_avg_response + response_time))
    fi
done

if [ $basic_healthy -gt 0 ]; then
    basic_avg_response=$((basic_avg_response / basic_healthy))
fi

pro_healthy=0
pro_total=${#PRO_SERVICES[@]}
pro_avg_response=0

for result in "${PRO_RESULTS[@]}"; do
    IFS=':' read -r name status response_time <<< "$result"
    if [ "$status" = "healthy" ]; then
        ((pro_healthy++))
        pro_avg_response=$((pro_avg_response + response_time))
    fi
done

if [ $pro_healthy -gt 0 ]; then
    pro_avg_response=$((pro_avg_response / pro_healthy))
fi

# 显示测试结果摘要
echo -e "${GREEN}=== 测试结果摘要 ===${NC}"
echo -e "${BLUE}基础版服务:${NC}"
echo -e "  总服务数: $basic_total"
echo -e "  健康服务数: $basic_healthy"
echo -e "  健康率: $(echo "scale=2; $basic_healthy * 100 / $basic_total" | bc -l)%"
echo -e "  平均响应时间: ${basic_avg_response}ms"
echo ""

echo -e "${BLUE}专业版服务:${NC}"
echo -e "  总服务数: $pro_total"
echo -e "  健康服务数: $pro_healthy"
echo -e "  健康率: $(echo "scale=2; $pro_healthy * 100 / $pro_total" | bc -l)%"
echo -e "  平均响应时间: ${pro_avg_response}ms"
echo ""

echo -e "${BLUE}对比结果:${NC}"
echo -e "  服务覆盖率: $(echo "scale=2; $pro_total / $basic_total" | bc -l)"
echo -e "  健康率提升: $(echo "scale=2; ($pro_healthy * 100 / $pro_total) - ($basic_healthy * 100 / $basic_total)" | bc -l)%"
echo -e "  响应时间改善: $(echo "scale=2; $basic_avg_response - $pro_avg_response" | bc -l)ms"
echo ""

echo -e "${GREEN}功能完整性对比测试完成！${NC}"
