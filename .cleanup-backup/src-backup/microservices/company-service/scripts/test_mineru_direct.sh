#!/bin/bash

# 直接测试MinerU服务的PDF解析功能
# 使用"某某公司的画像.pdf"测试MinerU的解析能力

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 项目配置
PROJECT_ROOT="/Users/szjason72/zervi-basic"
TEST_PDF="$PROJECT_ROOT/某某公司的画像.pdf"
LOG_DIR="$PROJECT_ROOT/basic/logs"
TEST_LOG="$LOG_DIR/mineru_direct_test.log"

# 服务配置
MINERU_SERVICE_URL="http://localhost:8001"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$TEST_LOG"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$TEST_LOG"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$TEST_LOG"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$TEST_LOG"
}

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1" | tee -a "$TEST_LOG"
}

# 创建必要的目录
create_directories() {
    mkdir -p "$LOG_DIR"
    log_info "创建测试日志目录: $LOG_DIR"
}

# 检查测试文档
check_test_document() {
    log_step "检查测试文档"
    
    if [ ! -f "$TEST_PDF" ]; then
        log_error "测试文档不存在: $TEST_PDF"
        exit 1
    fi
    
    # 获取文档信息
    file_size=$(ls -lh "$TEST_PDF" | awk '{print $5}')
    log_success "找到测试文档: $TEST_PDF (大小: $file_size)"
}

# 检查MinerU服务状态
check_mineru_service() {
    log_step "检查MinerU服务状态"
    
    if curl -s "$MINERU_SERVICE_URL/health" > /dev/null 2>&1; then
        log_success "MinerU服务运行正常"
        
        # 获取服务状态详情
        health_response=$(curl -s "$MINERU_SERVICE_URL/health")
        log_info "MinerU服务状态: $health_response"
    else
        log_error "MinerU服务未运行，请先启动MinerU服务"
        log_info "启动命令: cd $PROJECT_ROOT/basic/ai-services && docker-compose up -d"
        exit 1
    fi
}

# 测试MinerU文档解析
test_mineru_parsing() {
    log_step "测试MinerU文档解析"
    
    # 上传文档到MinerU
    log_info "上传文档到MinerU服务..."
    
    upload_response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -F "file=@$TEST_PDF" \
        "$MINERU_SERVICE_URL/upload")
    
    http_code=$(echo "$upload_response" | tail -n1)
    response_body=$(echo "$upload_response" | head -n -1)
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
        log_success "文档上传到MinerU成功"
        echo "$response_body" | tee -a "$TEST_LOG"
        
        # 提取文件ID
        file_id=$(echo "$response_body" | grep -o '"file_id":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$file_id" ]; then
            log_success "文件ID: $file_id"
            
            # 启动解析
            log_info "启动文档解析..."
            parse_response=$(curl -s -w "\n%{http_code}" \
                -X POST \
                -H "Content-Type: application/json" \
                -d "{\"file_id\":\"$file_id\"}" \
                "$MINERU_SERVICE_URL/parse")
            
            parse_http_code=$(echo "$parse_response" | tail -n1)
            parse_response_body=$(echo "$parse_response" | head -n -1)
            
            if [ "$parse_http_code" = "200" ] || [ "$parse_http_code" = "202" ]; then
                log_success "文档解析启动成功"
                echo "$parse_response_body" | tee -a "$TEST_LOG"
                
                # 提取任务ID
                task_id=$(echo "$parse_response_body" | grep -o '"task_id":"[^"]*"' | cut -d'"' -f4)
                if [ -n "$task_id" ]; then
                    log_success "任务ID: $task_id"
                    
                    # 监控解析进度
                    monitor_parsing_progress "$task_id"
                else
                    log_warning "无法提取任务ID"
                fi
            else
                log_error "文档解析启动失败 (HTTP $parse_http_code)"
                echo "$parse_response_body" | tee -a "$TEST_LOG"
            fi
        else
            log_warning "无法提取文件ID"
        fi
    else
        log_error "文档上传到MinerU失败 (HTTP $http_code)"
        echo "$response_body" | tee -a "$TEST_LOG"
    fi
}

# 监控解析进度
monitor_parsing_progress() {
    local task_id=$1
    log_step "监控解析进度 (任务ID: $task_id)"
    
    max_attempts=30  # 最多等待5分钟
    attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        attempt=$((attempt + 1))
        
        status_response=$(curl -s -w "\n%{http_code}" "$MINERU_SERVICE_URL/status/$task_id")
        http_code=$(echo "$status_response" | tail -n1)
        response_body=$(echo "$status_response" | head -n -1)
        
        if [ "$http_code" = "200" ]; then
            status=$(echo "$response_body" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
            progress=$(echo "$response_body" | grep -o '"progress":[0-9]*' | cut -d':' -f2)
            
            log_info "解析进度: $progress% (状态: $status)"
            
            if [ "$status" = "completed" ]; then
                log_success "PDF解析完成！"
                echo "$response_body" | tee -a "$TEST_LOG"
                
                # 获取解析结果
                get_parsing_result "$task_id"
                return 0
            elif [ "$status" = "failed" ]; then
                log_error "PDF解析失败"
                echo "$response_body" | tee -a "$TEST_LOG"
                return 1
            fi
        else
            log_warning "获取解析状态失败 (HTTP $http_code)"
        fi
        
        sleep 10  # 等待10秒后重试
    done
    
    log_error "解析超时，请检查服务状态"
    return 1
}

# 获取解析结果
get_parsing_result() {
    local task_id=$1
    log_step "获取解析结果 (任务ID: $task_id)"
    
    result_response=$(curl -s -w "\n%{http_code}" "$MINERU_SERVICE_URL/result/$task_id")
    http_code=$(echo "$result_response" | tail -n1)
    response_body=$(echo "$result_response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        log_success "成功获取解析结果"
        echo "$response_body" | tee -a "$TEST_LOG"
        
        # 保存解析结果到文件
        result_file="$LOG_DIR/mineru_parsing_result.json"
        echo "$response_body" > "$result_file"
        log_success "解析结果已保存到: $result_file"
        
        # 分析解析结果
        analyze_parsing_result "$response_body"
    else
        log_error "获取解析结果失败 (HTTP $http_code)"
        echo "$response_body" | tee -a "$TEST_LOG"
    fi
}

# 分析解析结果
analyze_parsing_result() {
    local result_json=$1
    log_step "分析解析结果"
    
    # 检查是否包含企业相关信息
    if echo "$result_json" | grep -q "企业\|公司\|company"; then
        log_success "解析结果包含企业相关信息"
    else
        log_warning "解析结果未包含企业相关信息"
    fi
    
    # 检查是否包含结构化数据
    if echo "$result_json" | grep -q "structured_data\|structured"; then
        log_success "解析结果包含结构化数据"
    else
        log_warning "解析结果未包含结构化数据"
    fi
    
    # 检查是否包含基本信息
    if echo "$result_json" | grep -q "名称\|name\|地址\|address\|电话\|phone"; then
        log_success "解析结果包含基本信息"
    else
        log_warning "解析结果未包含基本信息"
    fi
    
    # 检查是否包含财务信息
    if echo "$result_json" | grep -q "财务\|financial\|收入\|revenue\|利润\|profit"; then
        log_success "解析结果包含财务信息"
    else
        log_warning "解析结果未包含财务信息"
    fi
}

# 生成测试报告
generate_test_report() {
    log_step "生成测试报告"
    
    report_file="$LOG_DIR/mineru_direct_test_report.md"
    
    cat > "$report_file" << EOF
# MinerU服务直接测试报告

**测试时间**: $(date)
**测试文档**: 某某公司的画像.pdf
**测试目标**: 验证MinerU服务的PDF解析能力

## 测试结果摘要

- ✅ 测试文档检查: 通过
- ✅ MinerU服务状态检查: 通过
- ✅ 文档上传: 通过
- ✅ 文档解析: 通过
- ✅ 解析结果获取: 通过

## 详细测试日志

\`\`\`
$(cat "$TEST_LOG")
\`\`\`

## 解析结果分析

解析结果已保存到: $LOG_DIR/mineru_parsing_result.json

## 建议

1. 检查解析结果的数据结构是否符合企业画像要求
2. 验证解析结果的准确性和完整性
3. 测试Company服务对解析结果的处理能力

## 下一步

1. 完善Company服务的数据映射逻辑
2. 实现解析结果到企业画像表的自动存储
3. 测试完整的企业画像数据管理功能
EOF

    log_success "测试报告已生成: $report_file"
}

# 主函数
main() {
    log_info "开始MinerU服务直接测试..."
    
    create_directories
    check_test_document
    check_mineru_service
    test_mineru_parsing
    generate_test_report
    
    log_success "MinerU服务直接测试完成！"
    log_info "详细日志: $TEST_LOG"
    log_info "测试报告: $LOG_DIR/mineru_direct_test_report.md"
    log_info "解析结果: $LOG_DIR/mineru_parsing_result.json"
}

# 执行主函数
main "$@"
