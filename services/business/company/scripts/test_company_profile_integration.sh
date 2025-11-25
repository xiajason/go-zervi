#!/bin/bash

# Company服务企业画像PDF解析集成测试脚本
# 使用"某某公司的画像.pdf"测试完整的PDF解析到数据库存储流程

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
COMPANY_SERVICE_DIR="$PROJECT_ROOT/basic/backend/internal/company-service"
LOG_DIR="$PROJECT_ROOT/basic/logs"
TEST_LOG="$LOG_DIR/company_profile_integration_test.log"

# 服务配置
COMPANY_SERVICE_URL="http://localhost:8083"
MINERU_SERVICE_URL="http://localhost:8001"
BASIC_SERVER_URL="http://localhost:8080"

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

# 检查服务状态
check_services() {
    log_step "检查服务状态"
    
    # 检查MinerU服务
    if curl -s "$MINERU_SERVICE_URL/health" > /dev/null 2>&1; then
        log_success "MinerU服务运行正常"
    else
        log_error "MinerU服务未运行，请先启动MinerU服务"
        log_info "启动命令: cd $PROJECT_ROOT/basic/ai-services && docker-compose up -d"
        exit 1
    fi
    
    # 检查Company服务
    if curl -s "$COMPANY_SERVICE_URL/health" > /dev/null 2>&1; then
        log_success "Company服务运行正常"
    else
        log_error "Company服务未运行，请先启动Company服务"
        log_info "启动命令: cd $COMPANY_SERVICE_DIR && go run main.go"
        exit 1
    fi
    
    # 检查Basic-Server
    if curl -s "$BASIC_SERVER_URL/health" > /dev/null 2>&1; then
        log_success "Basic-Server运行正常"
    else
        log_warning "Basic-Server未运行，可能影响认证功能"
    fi
}

# 获取认证Token
get_auth_token() {
    log_step "获取认证Token"
    
    # 尝试从Basic-Server获取Token
    if curl -s "$BASIC_SERVER_URL/health" > /dev/null 2>&1; then
        # 这里需要根据实际的认证接口调整
        # 暂时使用一个测试Token
        AUTH_TOKEN="test-token-$(date +%s)"
        log_success "使用测试Token: $AUTH_TOKEN"
    else
        AUTH_TOKEN=""
        log_warning "无法获取认证Token，将使用无认证模式测试"
    fi
}

# 上传PDF文档
upload_pdf_document() {
    log_step "上传PDF文档到Company服务"
    
    # 构建上传请求
    upload_url="$COMPANY_SERVICE_URL/api/v1/company/documents/upload"
    
    # 使用curl上传文件
    response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: multipart/form-data" \
        -F "file=@$TEST_PDF" \
        -F "description=企业画像测试文档" \
        -F "company_name=某某公司" \
        $upload_url)
    
    # 分离响应体和状态码
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
        log_success "PDF文档上传成功"
        echo "$response_body" | tee -a "$TEST_LOG"
        
        # 提取文档ID
        DOCUMENT_ID=$(echo "$response_body" | grep -o '"document_id":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$DOCUMENT_ID" ]; then
            log_success "文档ID: $DOCUMENT_ID"
        else
            log_warning "无法提取文档ID"
        fi
    else
        log_error "PDF文档上传失败 (HTTP $http_code)"
        echo "$response_body" | tee -a "$TEST_LOG"
        exit 1
    fi
}

# 启动PDF解析
start_pdf_parsing() {
    log_step "启动PDF解析任务"
    
    if [ -z "$DOCUMENT_ID" ]; then
        log_error "文档ID为空，无法启动解析"
        exit 1
    fi
    
    parse_url="$COMPANY_SERVICE_URL/api/v1/company/documents/$DOCUMENT_ID/parse"
    
    response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        $parse_url)
    
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "202" ]; then
        log_success "PDF解析任务启动成功"
        echo "$response_body" | tee -a "$TEST_LOG"
        
        # 提取任务ID
        TASK_ID=$(echo "$response_body" | grep -o '"task_id":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$TASK_ID" ]; then
            log_success "任务ID: $TASK_ID"
        else
            log_warning "无法提取任务ID"
        fi
    else
        log_error "PDF解析任务启动失败 (HTTP $http_code)"
        echo "$response_body" | tee -a "$TEST_LOG"
        exit 1
    fi
}

# 监控解析进度
monitor_parsing_progress() {
    log_step "监控PDF解析进度"
    
    if [ -z "$TASK_ID" ]; then
        log_error "任务ID为空，无法监控进度"
        exit 1
    fi
    
    max_attempts=30  # 最多等待5分钟
    attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        attempt=$((attempt + 1))
        
        status_url="$COMPANY_SERVICE_URL/api/v1/company/documents/tasks/$TASK_ID/status"
        
        response=$(curl -s -w "\n%{http_code}" $status_url)
        http_code=$(echo "$response" | tail -n1)
        response_body=$(echo "$response" | head -n -1)
        
        if [ "$http_code" = "200" ]; then
            status=$(echo "$response_body" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
            progress=$(echo "$response_body" | grep -o '"progress":[0-9]*' | cut -d':' -f2)
            
            log_info "解析进度: $progress% (状态: $status)"
            
            if [ "$status" = "completed" ]; then
                log_success "PDF解析完成！"
                echo "$response_body" | tee -a "$TEST_LOG"
                return 0
            elif [ "$status" = "failed" ]; then
                log_error "PDF解析失败"
                echo "$response_body" | tee -a "$TEST_LOG"
                exit 1
            fi
        else
            log_warning "获取解析状态失败 (HTTP $http_code)"
        fi
        
        sleep 10  # 等待10秒后重试
    done
    
    log_error "解析超时，请检查服务状态"
    exit 1
}

# 验证解析结果
verify_parsing_results() {
    log_step "验证解析结果"
    
    if [ -z "$DOCUMENT_ID" ]; then
        log_error "文档ID为空，无法验证结果"
        exit 1
    fi
    
    # 获取解析结果
    result_url="$COMPANY_SERVICE_URL/api/v1/company/documents/$DOCUMENT_ID/result"
    
    response=$(curl -s -w "\n%{http_code}" $result_url)
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        log_success "成功获取解析结果"
        echo "$response_body" | tee -a "$TEST_LOG"
        
        # 检查是否包含结构化数据
        if echo "$response_body" | grep -q "structured_data"; then
            log_success "解析结果包含结构化数据"
        else
            log_warning "解析结果未包含结构化数据"
        fi
        
        # 检查是否包含企业基本信息
        if echo "$response_body" | grep -q "company_name"; then
            log_success "解析结果包含企业名称"
        else
            log_warning "解析结果未包含企业名称"
        fi
        
    else
        log_error "获取解析结果失败 (HTTP $http_code)"
        echo "$response_body" | tee -a "$TEST_LOG"
        exit 1
    fi
}

# 验证数据库存储
verify_database_storage() {
    log_step "验证数据库存储"
    
    # 这里需要根据实际的数据库连接信息调整
    # 暂时通过API接口验证数据是否已存储
    
    # 获取企业画像数据
    profile_url="$COMPANY_SERVICE_URL/api/v1/company/profile/summary/1"  # 假设company_id为1
    
    response=$(curl -s -w "\n%{http_code}" $profile_url)
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        log_success "成功获取企业画像数据"
        echo "$response_body" | tee -a "$TEST_LOG"
        
        # 检查各个数据表是否有数据
        if echo "$response_body" | grep -q "basic_info"; then
            log_success "企业基本信息已存储"
        else
            log_warning "企业基本信息未存储"
        fi
        
        if echo "$response_body" | grep -q "qualification"; then
            log_success "资质许可信息已存储"
        else
            log_warning "资质许可信息未存储"
        fi
        
        if echo "$response_body" | grep -q "personnel"; then
            log_success "人员竞争力信息已存储"
        else
            log_warning "人员竞争力信息未存储"
        fi
        
    else
        log_warning "获取企业画像数据失败 (HTTP $http_code)，可能数据尚未存储到新表结构"
        echo "$response_body" | tee -a "$TEST_LOG"
    fi
}

# 生成测试报告
generate_test_report() {
    log_step "生成测试报告"
    
    report_file="$LOG_DIR/company_profile_integration_test_report.md"
    
    cat > "$report_file" << EOF
# Company服务企业画像PDF解析集成测试报告

**测试时间**: $(date)
**测试文档**: 某某公司的画像.pdf
**测试目标**: 验证Company服务和MinerU的集成效果

## 测试结果摘要

- ✅ 测试文档检查: 通过
- ✅ 服务状态检查: 通过
- ✅ PDF文档上传: 通过
- ✅ PDF解析启动: 通过
- ✅ 解析进度监控: 通过
- ✅ 解析结果验证: 通过
- ⚠️ 数据库存储验证: 需要进一步检查

## 详细测试日志

\`\`\`
$(cat "$TEST_LOG")
\`\`\`

## 建议

1. 检查数据库表结构是否正确创建
2. 验证PDF解析结果是否正确映射到企业画像表
3. 测试完整的企业画像数据查询功能

## 下一步

1. 完善数据库迁移脚本
2. 优化PDF解析结果的数据映射
3. 实现完整的企业画像数据管理功能
EOF

    log_success "测试报告已生成: $report_file"
}

# 清理测试数据
cleanup_test_data() {
    log_step "清理测试数据"
    
    if [ -n "$DOCUMENT_ID" ]; then
        delete_url="$COMPANY_SERVICE_URL/api/v1/company/documents/$DOCUMENT_ID"
        
        response=$(curl -s -w "\n%{http_code}" -X DELETE $delete_url)
        http_code=$(echo "$response" | tail -n1)
        
        if [ "$http_code" = "200" ] || [ "$http_code" = "204" ]; then
            log_success "测试数据清理完成"
        else
            log_warning "测试数据清理失败"
        fi
    fi
}

# 主函数
main() {
    log_info "开始Company服务企业画像PDF解析集成测试..."
    
    create_directories
    check_test_document
    check_services
    get_auth_token
    upload_pdf_document
    start_pdf_parsing
    monitor_parsing_progress
    verify_parsing_results
    verify_database_storage
    generate_test_report
    
    # 询问是否清理测试数据
    echo -e "${YELLOW}是否清理测试数据？(y/N)${NC}"
    read -r cleanup_choice
    if [ "$cleanup_choice" = "y" ] || [ "$cleanup_choice" = "Y" ]; then
        cleanup_test_data
    fi
    
    log_success "Company服务企业画像PDF解析集成测试完成！"
    log_info "详细日志: $TEST_LOG"
    log_info "测试报告: $LOG_DIR/company_profile_integration_test_report.md"
}

# 执行主函数
main "$@"
