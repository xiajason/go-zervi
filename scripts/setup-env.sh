#!/bin/bash

# ============================================
# GoZervi 环境配置生成工具
# 交互式配置生成，支持完全离线部署
# ============================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DOCKER_DIR="$PROJECT_ROOT/docker"

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

# 读取用户输入（带默认值）
read_input() {
    local prompt="$1"
    local default="$2"
    local var_name="$3"
    
    if [ -n "$default" ]; then
        read -p "$prompt [$default]: " input
        eval "$var_name=\"\${input:-$default}\""
    else
        read -p "$prompt: " input
        eval "$var_name=\"$input\""
    fi
}

# 读取密码（隐藏输入）
read_password() {
    local prompt="$1"
    local var_name="$2"
    
    read -sp "$prompt: " password
    echo ""
    eval "$var_name=\"$password\""
}

# 生成随机密码
generate_password() {
    openssl rand -base64 32 | tr -d "=+/" | cut -c1-25
}

# 主函数
main() {
    echo ""
    echo "=========================================="
    echo "  GoZervi 环境配置生成工具"
    echo "=========================================="
    echo ""
    
    ENV_TEMPLATE="$DOCKER_DIR/.env.template"
    ENV_FILE="$DOCKER_DIR/.env"
    
    # 检查模板文件
    if [ ! -f "$ENV_TEMPLATE" ]; then
        log_error "未找到环境变量模板文件: $ENV_TEMPLATE"
        exit 1
    fi
    
    # 如果.env已存在，询问是否覆盖
    if [ -f "$ENV_FILE" ]; then
        log_warning "配置文件已存在: $ENV_FILE"
        read -p "是否覆盖现有配置? (y/N): " overwrite
        if [ "$overwrite" != "y" ] && [ "$overwrite" != "Y" ]; then
            log_info "取消配置生成"
            exit 0
        fi
    fi
    
    echo ""
    log_info "开始配置环境变量..."
    echo ""
    
    # ==================== 数据库配置 ====================
    echo "📊 数据库配置"
    echo "----------------------------------------"
    read_input "数据库名称" "zervigo_mvp" POSTGRES_DB
    read_input "数据库用户" "zervigo" POSTGRES_USER
    
    echo ""
    read -p "数据库密码（留空自动生成）: " POSTGRES_PASSWORD
    if [ -z "$POSTGRES_PASSWORD" ]; then
        POSTGRES_PASSWORD=$(generate_password)
        log_info "自动生成数据库密码: $POSTGRES_PASSWORD"
    fi
    
    read_input "数据库端口" "5432" POSTGRES_PORT
    echo ""
    
    # ==================== Redis配置 ====================
    echo "🔴 Redis配置"
    echo "----------------------------------------"
    read -p "Redis密码（留空自动生成）: " REDIS_PASSWORD
    if [ -z "$REDIS_PASSWORD" ]; then
        REDIS_PASSWORD=$(generate_password)
        log_info "自动生成Redis密码: $REDIS_PASSWORD"
    fi
    
    read_input "Redis端口" "6379" REDIS_PORT
    echo ""
    
    # ==================== Consul配置 ====================
    echo "🔍 Consul配置"
    echo "----------------------------------------"
    read_input "Consul端口" "8500" CONSUL_PORT
    echo ""
    
    # ==================== 服务端口配置 ====================
    echo "🚀 服务端口配置"
    echo "----------------------------------------"
    read_input "认证服务端口" "8207" AUTH_SERVICE_PORT
    read_input "租户服务端口" "8088" TENANT_SERVICE_PORT
    read_input "用户服务端口" "8082" USER_SERVICE_PORT
    read_input "职位服务端口" "8084" JOB_SERVICE_PORT
    read_input "企业服务端口" "8083" COMPANY_SERVICE_PORT
    read_input "网关端口" "9000" GATEWAY_PORT
    echo ""
    
    # ==================== 安全配置 ====================
    echo "🔐 安全配置"
    echo "----------------------------------------"
    read -p "JWT密钥（留空自动生成）: " JWT_SECRET
    if [ -z "$JWT_SECRET" ]; then
        JWT_SECRET=$(generate_password)
        log_info "自动生成JWT密钥: $JWT_SECRET"
    fi
    
    read_input "运行环境 (development/production)" "production" ENVIRONMENT
    read_input "Cookie过期时间（秒）" "604800" COOKIE_MAX_AGE
    echo ""
    
    # ==================== 其他配置 ====================
    echo "⚙️  其他配置"
    echo "----------------------------------------"
    read_input "时区" "Asia/Shanghai" TZ
    read_input "域名（可选）" "localhost" DOMAIN
    echo ""
    
    # ==================== 生成配置文件 ====================
    log_info "生成配置文件..."
    
    cat > "$ENV_FILE" <<EOF
# ============================================
# GoZervi 本地云部署环境变量
# 生成时间: $(date '+%Y-%m-%d %H:%M:%S')
# ============================================

# ==================== 数据库配置 ====================
POSTGRES_DB=$POSTGRES_DB
POSTGRES_USER=$POSTGRES_USER
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
POSTGRES_PORT=$POSTGRES_PORT

# ==================== Redis配置 ====================
REDIS_PASSWORD=$REDIS_PASSWORD
REDIS_PORT=$REDIS_PORT

# ==================== Consul配置 ====================
CONSUL_PORT=$CONSUL_PORT

# ==================== 服务端口配置 ====================
AUTH_SERVICE_PORT=$AUTH_SERVICE_PORT
TENANT_SERVICE_PORT=$TENANT_SERVICE_PORT
USER_SERVICE_PORT=$USER_SERVICE_PORT
JOB_SERVICE_PORT=$JOB_SERVICE_PORT
COMPANY_SERVICE_PORT=$COMPANY_SERVICE_PORT
GATEWAY_PORT=$GATEWAY_PORT

# ==================== 安全配置 ====================
JWT_SECRET=$JWT_SECRET
ENVIRONMENT=$ENVIRONMENT
COOKIE_MAX_AGE=$COOKIE_MAX_AGE

# ==================== 时区配置 ====================
TZ=$TZ

# ==================== 域名配置（可选） ====================
DOMAIN=$DOMAIN
EOF
    
    log_success "配置文件已生成: $ENV_FILE"
    echo ""
    
    # 显示配置摘要
    echo "📋 配置摘要："
    echo "----------------------------------------"
    echo "  数据库: $POSTGRES_DB@localhost:$POSTGRES_PORT"
    echo "  Redis: localhost:$REDIS_PORT"
    echo "  Consul: localhost:$CONSUL_PORT"
    echo "  认证服务: localhost:$AUTH_SERVICE_PORT"
    echo "  租户服务: localhost:$TENANT_SERVICE_PORT"
    echo "  环境: $ENVIRONMENT"
    echo ""
    
    log_success "配置完成！"
    echo ""
    log_info "下一步：运行 ./scripts/install-local-cloud.sh 开始安装"
    echo ""
}

# 运行主函数
main




