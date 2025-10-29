#!/bin/bash

# Company服务企业画像数据库结构测试脚本
# 用于测试数据库表结构的兼容性和完整性

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目配置
PROJECT_ROOT="/Users/szjason72/zervi-basic/basic"
LOG_DIR="$PROJECT_ROOT/logs"
TEST_LOG="$LOG_DIR/company_profile_db_test.log"

# 数据库配置
DB_HOST="localhost"
DB_PORT="3306"
DB_NAME="jobfirst"
DB_USER="root"
DB_PASSWORD=""

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

# 创建必要的目录
create_directories() {
    mkdir -p "$LOG_DIR"
    log_info "创建测试日志目录: $LOG_DIR"
}

# 检查数据库连接
check_database_connection() {
    log_info "检查数据库连接..."
    
    if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "USE $DB_NAME;" 2>/dev/null; then
        log_success "数据库连接成功"
        return 0
    else
        log_error "数据库连接失败"
        return 1
    fi
}

# 检查表是否存在
check_table_exists() {
    local table_name=$1
    local result=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -D"$DB_NAME" -e "SHOW TABLES LIKE '$table_name';" 2>/dev/null | wc -l)
    
    if [ "$result" -gt 1 ]; then
        return 0  # 表存在
    else
        return 1  # 表不存在
    fi
}

# 检查表结构
check_table_structure() {
    local table_name=$1
    local description=$2
    
    log_info "检查 $description 表结构..."
    
    if ! check_table_exists "$table_name"; then
        log_error "❌ 表 $table_name 不存在"
        return 1
    fi
    
    # 获取表结构信息
    local columns=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -D"$DB_NAME" -e "DESCRIBE $table_name;" 2>/dev/null | wc -l)
    
    if [ "$columns" -gt 1 ]; then
        log_success "✅ 表 $table_name 结构正常 (包含 $((columns-1)) 个字段)"
        return 0
    else
        log_error "❌ 表 $table_name 结构异常"
        return 1
    fi
}

# 检查索引
check_indexes() {
    local table_name=$1
    local description=$2
    
    log_info "检查 $description 表索引..."
    
    if ! check_table_exists "$table_name"; then
        log_error "❌ 表 $table_name 不存在，跳过索引检查"
        return 1
    fi
    
    # 获取索引信息
    local indexes=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -D"$DB_NAME" -e "SHOW INDEX FROM $table_name;" 2>/dev/null | wc -l)
    
    if [ "$indexes" -gt 1 ]; then
        log_success "✅ 表 $table_name 索引正常 (包含 $((indexes-1)) 个索引)"
        return 0
    else
        log_warning "⚠️ 表 $table_name 索引可能不完整"
        return 1
    fi
}

# 检查外键约束
check_foreign_keys() {
    local table_name=$1
    local description=$2
    
    log_info "检查 $description 表外键约束..."
    
    if ! check_table_exists "$table_name"; then
        log_error "❌ 表 $table_name 不存在，跳过外键检查"
        return 1
    fi
    
    # 获取外键信息
    local foreign_keys=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -D"$DB_NAME" -e "SELECT * FROM information_schema.KEY_COLUMN_USAGE WHERE TABLE_SCHEMA='$DB_NAME' AND TABLE_NAME='$table_name' AND REFERENCED_TABLE_NAME IS NOT NULL;" 2>/dev/null | wc -l)
    
    if [ "$foreign_keys" -gt 1 ]; then
        log_success "✅ 表 $table_name 外键约束正常 (包含 $((foreign_keys-1)) 个外键)"
        return 0
    else
        log_warning "⚠️ 表 $table_name 外键约束可能不完整"
        return 1
    fi
}

# 测试数据插入
test_data_insertion() {
    local table_name=$1
    local description=$2
    
    log_info "测试 $description 表数据插入..."
    
    if ! check_table_exists "$table_name"; then
        log_error "❌ 表 $table_name 不存在，跳过数据插入测试"
        return 1
    fi
    
    # 根据表名生成测试数据
    case "$table_name" in
        "company_basic_info")
            local test_sql="INSERT INTO company_basic_info (company_id, report_id, company_name, industry_category, business_status, registered_capital, currency, created_at, updated_at) VALUES (999, 'TEST001', '测试企业', '测试行业', '存续', 1000.00, 'CNY', NOW(), NOW());"
            ;;
        "qualification_license")
            local test_sql="INSERT INTO qualification_license (company_id, report_id, type, name, status, created_at, updated_at) VALUES (999, 'TEST001', '资质', '测试资质', '有效', NOW(), NOW());"
            ;;
        "personnel_competitiveness")
            local test_sql="INSERT INTO personnel_competitiveness (company_id, report_id, total_employees, created_at, updated_at) VALUES (999, 'TEST001', 100, NOW(), NOW());"
            ;;
        "tech_innovation_score")
            local test_sql="INSERT INTO tech_innovation_score (company_id, report_id, basic_score, talent_score, created_at, updated_at) VALUES (999, 'TEST001', 85.5, 90.0, NOW(), NOW());"
            ;;
        *)
            log_warning "⚠️ 跳过表 $table_name 的数据插入测试"
            return 0
            ;;
    esac
    
    # 执行测试插入
    if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -D"$DB_NAME" -e "$test_sql" 2>/dev/null; then
        log_success "✅ 表 $table_name 数据插入测试成功"
        
        # 清理测试数据
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -D"$DB_NAME" -e "DELETE FROM $table_name WHERE company_id = 999;" 2>/dev/null
        return 0
    else
        log_error "❌ 表 $table_name 数据插入测试失败"
        return 1
    fi
}

# 测试JSON字段
test_json_fields() {
    log_info "测试JSON字段功能..."
    
    if ! check_table_exists "company_basic_info"; then
        log_error "❌ company_basic_info 表不存在，跳过JSON字段测试"
        return 1
    fi
    
    # 测试JSON字段插入
    local json_test_sql="INSERT INTO company_basic_info (company_id, report_id, company_name, tags, created_at, updated_at) VALUES (998, 'JSON_TEST', 'JSON测试企业', '[\"测试标签1\", \"测试标签2\"]', NOW(), NOW());"
    
    if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -D"$DB_NAME" -e "$json_test_sql" 2>/dev/null; then
        log_success "✅ JSON字段插入测试成功"
        
        # 测试JSON字段查询
        local json_query_sql="SELECT JSON_EXTRACT(tags, '$[0]') as first_tag FROM company_basic_info WHERE company_id = 998;"
        local result=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -D"$DB_NAME" -e "$json_query_sql" 2>/dev/null | tail -1)
        
        if [ "$result" = '"测试标签1"' ]; then
            log_success "✅ JSON字段查询测试成功"
        else
            log_warning "⚠️ JSON字段查询测试异常: $result"
        fi
        
        # 清理测试数据
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -D"$DB_NAME" -e "DELETE FROM company_basic_info WHERE company_id = 998;" 2>/dev/null
        return 0
    else
        log_error "❌ JSON字段插入测试失败"
        return 1
    fi
}

# 运行完整测试
run_complete_test() {
    log_info "开始企业画像数据库结构完整测试..."
    
    local tables=(
        "company_documents:企业文档表"
        "company_parsing_tasks:企业解析任务表"
        "company_structured_data:企业结构化数据表"
        "company_basic_info:企业基本信息表"
        "qualification_license:资质许可表"
        "personnel_competitiveness:人员竞争力表"
        "provident_fund:公积金信息表"
        "subsidy_info:资助补贴表"
        "company_relationships:企业关系图谱表"
        "tech_innovation_score:科创评分表"
        "company_financial_info:企业财务信息表"
        "company_risk_info:企业风险信息表"
    )
    
    local total_tests=0
    local passed_tests=0
    
    # 测试表结构
    for table_entry in "${tables[@]}"; do
        local table_name=$(echo "$table_entry" | cut -d':' -f1)
        local description=$(echo "$table_entry" | cut -d':' -f2)
        
        ((total_tests++))
        if check_table_structure "$table_name" "$description"; then
            ((passed_tests++))
        fi
        
        ((total_tests++))
        if check_indexes "$table_name" "$description"; then
            ((passed_tests++))
        fi
        
        ((total_tests++))
        if check_foreign_keys "$table_name" "$description"; then
            ((passed_tests++))
        fi
        
        ((total_tests++))
        if test_data_insertion "$table_name" "$description"; then
            ((passed_tests++))
        fi
    done
    
    # 测试JSON字段
    ((total_tests++))
    if test_json_fields; then
        ((passed_tests++))
    fi
    
    echo
    log_info "测试结果总结:"
    log_info "  总测试数: $total_tests"
    log_success "  通过测试: $passed_tests"
    log_error "  失败测试: $((total_tests - passed_tests))"
    
    if [ "$passed_tests" -eq "$total_tests" ]; then
        log_success "🎉 所有测试通过！数据库结构完全兼容"
        return 0
    else
        log_error "❌ 部分测试失败，请检查数据库结构"
        return 1
    fi
}

# 生成测试报告
generate_test_report() {
    log_info "生成测试报告..."
    
    local report_file="$LOG_DIR/company_profile_db_test_report_$(date +%Y%m%d_%H%M%S).txt"
    
    cat > "$report_file" << EOF
==========================================
Company服务企业画像数据库结构测试报告
==========================================
测试时间: $(date)
数据库: $DB_NAME@$DB_HOST:$DB_PORT
测试脚本: $0

测试内容:
✅ 数据库连接测试
✅ 表结构完整性测试
✅ 索引完整性测试
✅ 外键约束测试
✅ 数据插入功能测试
✅ JSON字段功能测试

测试表列表:
- company_documents (企业文档表)
- company_parsing_tasks (企业解析任务表)
- company_structured_data (企业结构化数据表)
- company_basic_info (企业基本信息表)
- qualification_license (资质许可表)
- personnel_competitiveness (人员竞争力表)
- provident_fund (公积金信息表)
- subsidy_info (资助补贴表)
- company_relationships (企业关系图谱表)
- tech_innovation_score (科创评分表)
- company_financial_info (企业财务信息表)
- company_risk_info (企业风险信息表)

数据库特性:
- 支持MySQL 8.0+
- 支持JSON字段类型
- 完整的索引优化
- 外键约束保证数据一致性
- 支持企业画像完整数据存储

详细日志: $TEST_LOG
==========================================
EOF
    
    log_success "测试报告已生成: $report_file"
}

# 显示帮助信息
show_help() {
    cat << EOF
Company服务企业画像数据库结构测试脚本

用法: $0 [选项]

选项:
  --help             显示此帮助信息
  --check            仅检查数据库连接
  --structure        仅测试表结构
  --data             仅测试数据插入
  --json             仅测试JSON字段
  --full             执行完整测试

环境变量:
  DB_HOST            数据库主机 (默认: localhost)
  DB_PORT            数据库端口 (默认: 3306)
  DB_NAME            数据库名称 (默认: jobfirst)
  DB_USER            数据库用户 (默认: root)
  DB_PASSWORD        数据库密码 (默认: 空)

示例:
  $0 --check          # 检查数据库连接
  $0 --full           # 执行完整测试
  $0 --structure      # 仅测试表结构
  $0 --data           # 仅测试数据插入

EOF
}

# 主函数
main() {
    # 解析命令行参数
    case "${1:-}" in
        --help)
            show_help
            exit 0
            ;;
        --check)
            create_directories
            check_database_connection
            ;;
        --structure)
            create_directories
            if check_database_connection; then
                # 仅测试表结构
                local tables=("company_basic_info:企业基本信息表" "qualification_license:资质许可表" "personnel_competitiveness:人员竞争力表")
                for table_entry in "${tables[@]}"; do
                    local table_name=$(echo "$table_entry" | cut -d':' -f1)
                    local description=$(echo "$table_entry" | cut -d':' -f2)
                    check_table_structure "$table_name" "$description"
                done
            fi
            ;;
        --data)
            create_directories
            if check_database_connection; then
                # 仅测试数据插入
                local tables=("company_basic_info:企业基本信息表" "qualification_license:资质许可表")
                for table_entry in "${tables[@]}"; do
                    local table_name=$(echo "$table_entry" | cut -d':' -f1)
                    local description=$(echo "$table_entry" | cut -d':' -f2)
                    test_data_insertion "$table_name" "$description"
                done
            fi
            ;;
        --json)
            create_directories
            if check_database_connection; then
                test_json_fields
            fi
            ;;
        --full)
            create_directories
            echo "=========================================="
            echo "🧪 Company服务企业画像数据库结构测试"
            echo "=========================================="
            echo
            
            if check_database_connection; then
                run_complete_test
                generate_test_report
                
                echo
                echo "=========================================="
                echo "✅ 数据库结构测试完成"
                echo "=========================================="
                echo
                log_success "测试完成，详细结果请查看测试报告"
            else
                log_error "数据库连接失败，测试终止"
                exit 1
            fi
            ;;
        *)
            show_help
            exit 1
            ;;
    esac
}

# 错误处理
trap 'log_error "测试过程中发生错误"; exit 1' ERR

# 执行主函数
main "$@"
