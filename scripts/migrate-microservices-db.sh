#!/bin/bash

# Zervigo 微服务数据库迁移脚本
# 执行微服务数据库结构创建

set -e

echo "🚀 开始执行 Zervigo 微服务数据库迁移..."
echo "=================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 项目配置
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DB_NAME="zervigo_mvp"
DB_USER="$(whoami)"

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

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# 检查PostgreSQL连接
check_postgresql_connection() {
    log_step "检查PostgreSQL连接..."
    
    if ! psql -U "$DB_USER" -d postgres -c "SELECT 1;" > /dev/null 2>&1; then
        log_error "PostgreSQL连接失败！"
        exit 1
    fi
    
    log_success "PostgreSQL连接正常"
}

# 检查数据库是否存在
check_database_exists() {
    log_step "检查数据库是否存在..."
    
    if psql -U "$DB_USER" -d postgres -c "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME';" | grep -q "1 row"; then
        log_success "数据库 $DB_NAME 已存在"
        return 0
    else
        log_warning "数据库 $DB_NAME 不存在"
        return 1
    fi
}

# 创建数据库
create_database() {
    log_step "创建数据库 $DB_NAME..."
    
    psql -U "$DB_USER" -d postgres -c "CREATE DATABASE $DB_NAME;"
    log_success "数据库 $DB_NAME 创建成功"
}

# 执行迁移脚本
execute_migration() {
    log_step "执行微服务数据库迁移脚本..."
    
    local migration_file="$PROJECT_ROOT/databases/postgres/init/02-zervigo-microservices-schema.sql"
    
    if [ ! -f "$migration_file" ]; then
        log_error "迁移脚本不存在: $migration_file"
        exit 1
    fi
    
    log_info "执行迁移脚本: $migration_file"
    psql -U "$DB_USER" -d "$DB_NAME" -f "$migration_file"
    
    log_success "微服务数据库迁移完成"
}

# 验证迁移结果
verify_migration() {
    log_step "验证迁移结果..."
    
    # 检查表数量
    local table_count=$(psql -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE 'zervigo_%';" | tr -d ' ')
    
    if [ "$table_count" -ge 16 ]; then
        log_success "表创建成功 ($table_count 个表)"
    else
        log_error "表创建失败，期望至少16个表，实际 $table_count 个"
        exit 1
    fi
    
    # 检查角色数量
    local role_count=$(psql -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM zervigo_auth_roles;" | tr -d ' ')
    
    if [ "$role_count" -ge 7 ]; then
        log_success "角色创建成功 ($role_count 个角色)"
    else
        log_error "角色创建失败，期望至少7个角色，实际 $role_count 个"
        exit 1
    fi
    
    # 检查权限数量
    local permission_count=$(psql -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM zervigo_auth_permissions;" | tr -d ' ')
    
    if [ "$permission_count" -ge 20 ]; then
        log_success "权限创建成功 ($permission_count 个权限)"
    else
        log_error "权限创建失败，期望至少20个权限，实际 $permission_count 个"
        exit 1
    fi
    
    # 检查默认用户
    local user_count=$(psql -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM zervigo_auth_users WHERE username = 'admin';" | tr -d ' ')
    
    if [ "$user_count" -eq 1 ]; then
        log_success "默认管理员用户创建成功"
    else
        log_error "默认管理员用户创建失败"
        exit 1
    fi
}

# 显示数据库信息
show_database_info() {
    log_step "数据库信息总览"
    echo "=================================="
    echo "📊 数据库信息:"
    echo "  数据库名: $DB_NAME"
    echo "  用户名: $DB_USER"
    echo "  端口: 5432"
    echo "  连接字符串: postgres://$DB_USER@localhost:5432/$DB_NAME"
    echo ""
    echo "🏗️ 微服务表结构:"
    echo "  auth-service: 7个表"
    echo "    - zervigo_auth_users (用户认证表)"
    echo "    - zervigo_auth_roles (角色表)"
    echo "    - zervigo_auth_permissions (权限表)"
    echo "    - zervigo_auth_user_roles (用户角色关联表)"
    echo "    - zervigo_auth_role_permissions (角色权限关联表)"
    echo "    - zervigo_auth_tokens (JWT Token管理表)"
    echo "    - zervigo_auth_login_logs (登录审计表)"
    echo ""
    echo "  user-service: 5个表"
    echo "    - zervigo_user_profiles (用户档案表)"
    echo "    - zervigo_user_skills (用户技能表)"
    echo "    - zervigo_user_education (教育经历表)"
    echo "    - zervigo_user_experience (工作经历表)"
    echo "    - zervigo_user_statistics (用户行为统计表)"
    echo ""
    echo "  job-service: 4个表"
    echo "    - zervigo_jobs (职位表)"
    echo "    - zervigo_job_applications (职位申请表)"
    echo "    - zervigo_job_favorites (职位收藏表)"
    echo "    - zervigo_job_search_history (搜索历史表)"
    echo ""
    echo "👤 默认管理员账号:"
    echo "  用户名: admin"
    echo "  密码: admin123"
    echo "  邮箱: admin@zervigo.com"
    echo "  角色: super_admin"
    echo ""
    echo "🔧 管理命令:"
    echo "  连接数据库: psql -U $DB_USER -d $DB_NAME"
    echo "  查看所有表: psql -U $DB_USER -d $DB_NAME -c '\dt zervigo_*'"
    echo "  查看用户: psql -U $DB_USER -d $DB_NAME -c 'SELECT username, email FROM zervigo_auth_users;'"
    echo "  查看角色: psql -U $DB_USER -d $DB_NAME -c 'SELECT role_name FROM zervigo_auth_roles;'"
    echo "  查看权限: psql -U $DB_USER -d $DB_NAME -c 'SELECT permission_name FROM zervigo_auth_permissions;'"
}

# 主函数
main() {
    log_info "开始执行 Zervigo 微服务数据库迁移..."
    
    # 检查PostgreSQL连接
    check_postgresql_connection
    
    # 检查数据库是否存在
    if ! check_database_exists; then
        create_database
    fi
    
    # 执行迁移脚本
    execute_migration
    
    # 验证迁移结果
    verify_migration
    
    # 显示数据库信息
    show_database_info
    
    log_success "🎉 Zervigo 微服务数据库迁移完成！"
    log_info "现在可以开始启动微服务了"
}

# 执行主函数
main "$@"
