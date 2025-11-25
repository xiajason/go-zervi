#!/bin/bash

# Company服务PDF文档解析测试脚本
# 用于测试MinerU服务与Company服务的集成

set -e

# 配置
COMPANY_SERVICE_URL="http://localhost:8083"
MINERU_SERVICE_URL="http://localhost:8001"
TEST_USER_ID=1
TEST_COMPANY_ID=1

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查服务状态
check_service_health() {
    local service_name=$1
    local service_url=$2
    
    log_info "检查 $service_name 服务状态..."
    
    if curl -s "$service_url/health" > /dev/null; then
        log_success "$service_name 服务运行正常"
        return 0
    else
        log_error "$service_name 服务不可用"
        return 1
    fi
}

# 创建测试PDF文件
create_test_pdf() {
    local test_file="test_company_document.pdf"
    
    log_info "创建测试PDF文件..."
    
    # 创建简单的PDF内容（这里使用文本文件模拟）
    cat > "$test_file" << 'EOF'
公司名称：测试科技有限公司
简称：测试科技
成立年份：2020
公司规模：51-100人
行业：互联网/电子商务
地址：北京市海淀区中关村大街1号
网站：https://www.testtech.com

主营业务：软件开发、技术咨询
主要产品：企业管理系统、移动应用开发
目标客户：中小企业、政府机构
竞争优势：技术领先、服务优质

组织架构：技术部、市场部、人事部、财务部
部门设置：研发中心、销售中心、运营中心
人员规模：80人
管理团队：CEO、CTO、CFO

注册资本：1000万元
年营业额：5000万元
融资情况：A轮融资2000万元
上市状态：未上市
EOF
    
    log_success "测试PDF文件创建完成: $test_file"
    echo "$test_file"
}

# 测试MinerU服务健康检查
test_mineru_health() {
    log_info "测试MinerU服务健康检查..."
    
    response=$(curl -s "$MINERU_SERVICE_URL/health")
    if echo "$response" | grep -q "healthy"; then
        log_success "MinerU服务健康检查通过"
        echo "$response" | jq .
    else
        log_error "MinerU服务健康检查失败"
        return 1
    fi
}

# 测试MinerU解析状态
test_mineru_status() {
    log_info "测试MinerU解析状态..."
    
    response=$(curl -s "$MINERU_SERVICE_URL/api/v1/parse/status")
    if echo "$response" | grep -q "success"; then
        log_success "MinerU解析状态获取成功"
        echo "$response" | jq .
    else
        log_error "MinerU解析状态获取失败"
        return 1
    fi
}

# 测试Company服务健康检查
test_company_health() {
    log_info "测试Company服务健康检查..."
    
    response=$(curl -s "$COMPANY_SERVICE_URL/health")
    if echo "$response" | grep -q "healthy"; then
        log_success "Company服务健康检查通过"
        echo "$response" | jq .
    else
        log_error "Company服务健康检查失败"
        return 1
    fi
}

# 测试文档上传
test_document_upload() {
    local test_file=$1
    
    log_info "测试文档上传..."
    
    # 注意：这里需要有效的JWT token，实际使用时需要先登录获取token
    # 这里使用模拟的token进行测试
    local auth_token="Bearer your_jwt_token_here"
    
    response=$(curl -s -X POST \
        -H "Authorization: $auth_token" \
        -F "file=@$test_file" \
        -F "company_id=$TEST_COMPANY_ID" \
        -F "title=测试企业文档" \
        "$COMPANY_SERVICE_URL/api/v1/company/documents/upload")
    
    if echo "$response" | grep -q "success"; then
        log_success "文档上传成功"
        echo "$response" | jq .
        
        # 提取文档ID
        local document_id=$(echo "$response" | jq -r '.document_id')
        echo "$document_id"
    else
        log_error "文档上传失败"
        echo "$response"
        return 1
    fi
}

# 测试文档解析
test_document_parsing() {
    local document_id=$1
    
    log_info "测试文档解析..."
    
    local auth_token="Bearer your_jwt_token_here"
    
    response=$(curl -s -X POST \
        -H "Authorization: $auth_token" \
        "$COMPANY_SERVICE_URL/api/v1/company/documents/$document_id/parse")
    
    if echo "$response" | grep -q "success"; then
        log_success "文档解析任务创建成功"
        echo "$response" | jq .
        
        # 提取任务ID
        local task_id=$(echo "$response" | jq -r '.task_id')
        echo "$task_id"
    else
        log_error "文档解析任务创建失败"
        echo "$response"
        return 1
    fi
}

# 测试解析状态查询
test_parse_status() {
    local document_id=$1
    local max_attempts=10
    local attempt=1
    
    log_info "测试解析状态查询..."
    
    local auth_token="Bearer your_jwt_token_here"
    
    while [ $attempt -le $max_attempts ]; do
        log_info "第 $attempt 次查询解析状态..."
        
        response=$(curl -s -H "Authorization: $auth_token" \
            "$COMPANY_SERVICE_URL/api/v1/company/documents/$document_id/parse/status")
        
        if echo "$response" | grep -q "success"; then
            local status=$(echo "$response" | jq -r '.status')
            local progress=$(echo "$response" | jq -r '.progress')
            
            log_info "解析状态: $status, 进度: $progress%"
            
            if [ "$status" = "completed" ]; then
                log_success "文档解析完成"
                echo "$response" | jq .
                return 0
            elif [ "$status" = "failed" ]; then
                log_error "文档解析失败"
                echo "$response" | jq .
                return 1
            fi
        else
            log_error "解析状态查询失败"
            echo "$response"
            return 1
        fi
        
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_warning "解析超时，请手动检查"
    return 1
}

# 测试文档列表查询
test_document_list() {
    log_info "测试文档列表查询..."
    
    local auth_token="Bearer your_jwt_token_here"
    
    response=$(curl -s -H "Authorization: $auth_token" \
        "$COMPANY_SERVICE_URL/api/v1/company/documents/?company_id=$TEST_COMPANY_ID")
    
    if echo "$response" | grep -q "success"; then
        log_success "文档列表查询成功"
        echo "$response" | jq .
    else
        log_error "文档列表查询失败"
        echo "$response"
        return 1
    fi
}

# 清理测试文件
cleanup() {
    log_info "清理测试文件..."
    
    if [ -f "test_company_document.pdf" ]; then
        rm -f "test_company_document.pdf"
        log_success "测试文件已清理"
    fi
}

# 主测试流程
main() {
    log_info "开始Company服务PDF文档解析测试..."
    
    # 检查服务状态
    if ! check_service_health "MinerU" "$MINERU_SERVICE_URL"; then
        log_error "MinerU服务不可用，请先启动服务"
        exit 1
    fi
    
    if ! check_service_health "Company" "$COMPANY_SERVICE_URL"; then
        log_error "Company服务不可用，请先启动服务"
        exit 1
    fi
    
    # 创建测试文件
    test_file=$(create_test_pdf)
    
    # 测试MinerU服务
    test_mineru_health
    test_mineru_status
    
    # 测试Company服务
    test_company_health
    
    # 测试文档上传（需要有效的认证token）
    log_warning "注意：文档上传和解析测试需要有效的JWT token"
    log_warning "请先通过登录接口获取token，然后修改脚本中的auth_token变量"
    
    # 以下测试需要有效的认证token
    # document_id=$(test_document_upload "$test_file")
    # task_id=$(test_document_parsing "$document_id")
    # test_parse_status "$document_id"
    # test_document_list
    
    # 清理
    cleanup
    
    log_success "Company服务PDF文档解析测试完成"
}

# 显示帮助信息
show_help() {
    echo "Company服务PDF文档解析测试脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示帮助信息"
    echo "  -c, --check    仅检查服务状态"
    echo "  -t, --test     运行完整测试"
    echo ""
    echo "示例:"
    echo "  $0 --check     # 检查服务状态"
    echo "  $0 --test      # 运行完整测试"
}

# 解析命令行参数
case "${1:-}" in
    -h|--help)
        show_help
        exit 0
        ;;
    -c|--check)
        check_service_health "MinerU" "$MINERU_SERVICE_URL"
        check_service_health "Company" "$COMPANY_SERVICE_URL"
        exit 0
        ;;
    -t|--test)
        main
        ;;
    *)
        show_help
        exit 1
        ;;
esac
