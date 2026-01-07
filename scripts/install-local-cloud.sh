#!/bin/bash

# ============================================
# GoZervi 本地云一键安装脚本
# 完全离线部署，所有资源本地化
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

# 检查命令是否存在
check_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "$1 未安装，请先安装 $1"
        exit 1
    fi
}

# 环境检查
check_environment() {
    log_info "检查环境..."
    
    # 检查Docker
    if ! check_command docker; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    # 检查Docker是否运行
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker 未运行，请先启动 Docker"
        exit 1
    fi
    
    # 检查Docker Compose
    if ! check_command docker-compose; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    # 检查系统资源
    log_info "检查系统资源..."
    MEM_TOTAL=$(free -m 2>/dev/null | grep Mem | awk '{print $2}' || echo "0")
    if [ "$MEM_TOTAL" -lt 2048 ] && [ "$MEM_TOTAL" -gt 0 ]; then
        log_warning "系统内存为 ${MEM_TOTAL}MB，建议至少 2GB 内存"
    fi
    
    log_success "环境检查通过"
}

# 导入镜像（如果存在本地镜像文件）
import_images() {
    log_info "检查本地镜像..."
    
    IMAGES_DIR="$PROJECT_ROOT/docker/images"
    if [ -d "$IMAGES_DIR" ]; then
        log_info "发现本地镜像目录，开始导入..."
        
        for image_file in "$IMAGES_DIR"/*.tar; do
            if [ -f "$image_file" ]; then
                log_info "导入镜像: $(basename "$image_file")"
                docker load -i "$image_file"
            fi
        done
        
        log_success "镜像导入完成"
    else
        log_info "未发现本地镜像目录，将使用 Docker Hub 或构建镜像"
    fi
}

# 生成配置文件
generate_config() {
    log_info "生成配置文件..."
    
    ENV_TEMPLATE="$DOCKER_DIR/.env.template"
    ENV_FILE="$DOCKER_DIR/.env"
    
    if [ ! -f "$ENV_FILE" ]; then
        if [ -f "$ENV_TEMPLATE" ]; then
            log_info "从模板生成 .env 文件..."
            cp "$ENV_TEMPLATE" "$ENV_FILE"
            log_warning "请编辑 $ENV_FILE 文件，修改配置后重新运行此脚本"
            log_info "或者运行: ./scripts/setup-env.sh"
            exit 0
        else
            log_error "未找到环境变量模板文件: $ENV_TEMPLATE"
            exit 1
        fi
    else
        log_info "使用现有 .env 文件"
    fi
    
    log_success "配置文件准备完成"
}

# 初始化数据库
init_database() {
    log_info "等待数据库启动..."
    
    # 等待PostgreSQL启动
    MAX_WAIT=60
    WAIT_COUNT=0
    
    while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
        if docker-compose -f "$DOCKER_DIR/docker-compose.local-cloud.yml" exec -T postgresql pg_isready -U "${POSTGRES_USER:-zervigo}" > /dev/null 2>&1; then
            log_success "数据库已就绪"
            return 0
        fi
        WAIT_COUNT=$((WAIT_COUNT + 5))
        sleep 5
        log_info "等待数据库启动... ($WAIT_COUNT/$MAX_WAIT 秒)"
    done
    
    log_error "数据库启动超时"
    return 1
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    cd "$DOCKER_DIR"
    
    # 停止现有服务
    log_info "停止现有服务..."
    docker-compose -f docker-compose.local-cloud.yml down 2>/dev/null || true
    
    # 构建镜像（如果需要）
    log_info "构建服务镜像..."
    docker-compose -f docker-compose.local-cloud.yml build --no-cache
    
    # 启动服务
    log_info "启动所有服务..."
    docker-compose -f docker-compose.local-cloud.yml up -d
    
    log_success "服务启动完成"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    cd "$DOCKER_DIR"
    
    # 等待服务启动
    sleep 10
    
    # 检查服务状态
    log_info "检查服务状态..."
    docker-compose -f docker-compose.local-cloud.yml ps
    
    # 检查各个服务的健康状态
    SERVICES=(
        "postgresql:5432"
        "redis:6379"
        "consul:8500"
        "auth-service:8207"
        "tenant-service:8088"
    )
    
    FAILED_SERVICES=()
    
    for service in "${SERVICES[@]}"; do
        SERVICE_NAME=$(echo "$service" | cut -d: -f1)
        SERVICE_PORT=$(echo "$service" | cut -d: -f2)
        
        log_info "检查 $SERVICE_NAME..."
        
        if docker-compose -f docker-compose.local-cloud.yml exec -T "$SERVICE_NAME" wget --no-verbose --tries=1 --spider "http://localhost:$SERVICE_PORT/health" > /dev/null 2>&1 || \
           docker-compose -f docker-compose.local-cloud.yml exec -T "$SERVICE_NAME" pg_isready -U "${POSTGRES_USER:-zervigo}" > /dev/null 2>&1 || \
           docker-compose -f docker-compose.local-cloud.yml exec -T "$SERVICE_NAME" redis-cli -a "${REDIS_PASSWORD:-zervigo2025}" ping > /dev/null 2>&1 || \
           docker-compose -f docker-compose.local-cloud.yml exec -T "$SERVICE_NAME" consul members > /dev/null 2>&1; then
            log_success "$SERVICE_NAME 健康检查通过"
        else
            log_warning "$SERVICE_NAME 健康检查失败（可能正在启动中）"
            FAILED_SERVICES+=("$SERVICE_NAME")
        fi
    done
    
    if [ ${#FAILED_SERVICES[@]} -eq 0 ]; then
        log_success "所有服务健康检查通过"
    else
        log_warning "以下服务健康检查失败: ${FAILED_SERVICES[*]}"
        log_info "请稍后运行: docker-compose -f docker/docker-compose.local-cloud.yml ps"
    fi
}

# 显示服务信息
show_service_info() {
    log_info "服务访问地址："
    echo ""
    echo "  📊 基础设施服务："
    echo "    PostgreSQL: localhost:${POSTGRES_PORT:-5432}"
    echo "    Redis:      localhost:${REDIS_PORT:-6379}"
    echo "    Consul:     http://localhost:${CONSUL_PORT:-8500}"
    echo ""
    echo "  🔐 核心服务："
    echo "    Auth Service:   http://localhost:${AUTH_SERVICE_PORT:-8207}"
    echo "    Tenant Service: http://localhost:${TENANT_SERVICE_PORT:-8088}"
    echo "    User Service:   http://localhost:${USER_SERVICE_PORT:-8082}"
    echo ""
    echo "  💼 业务服务："
    echo "    Job Service:    http://localhost:${JOB_SERVICE_PORT:-8084}"
    echo "    Company Service: http://localhost:${COMPANY_SERVICE_PORT:-8083}"
    echo ""
    echo "  🧪 快速验证命令："
    echo "    # 检查所有服务状态"
    echo "    docker-compose -f docker/docker-compose.local-cloud.yml ps"
    echo ""
    echo "    # 查看服务日志"
    echo "    docker-compose -f docker/docker-compose.local-cloud.yml logs -f [service-name]"
    echo ""
    echo "    # 停止所有服务"
    echo "    docker-compose -f docker/docker-compose.local-cloud.yml down"
    echo ""
}

# 主函数
main() {
    echo ""
    echo "=========================================="
    echo "  GoZervi 本地云一键安装脚本"
    echo "=========================================="
    echo ""
    
    # 加载环境变量
    if [ -f "$DOCKER_DIR/.env" ]; then
        set -a
        source "$DOCKER_DIR/.env"
        set +a
    fi
    
    # 执行安装步骤
    check_environment
    import_images
    generate_config
    
    # 如果配置文件是新生成的，退出让用户配置
    if [ ! -f "$DOCKER_DIR/.env" ]; then
        exit 0
    fi
    
    start_services
    init_database
    health_check
    show_service_info
    
    echo ""
    log_success "安装完成！"
    echo ""
}

# 运行主函数
main




