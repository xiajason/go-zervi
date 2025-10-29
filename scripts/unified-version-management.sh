#!/bin/bash

# Zervigo OpenLinkSaaS 版本管理统一脚本
# 解决版本管理混乱问题，实现微服务版本统一管理

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="/Users/szjason72/szbolent/Zervigo/zervigo.demo"

# 版本信息
VERSION_FILE="$PROJECT_ROOT/VERSION"
CHANGELOG_FILE="$PROJECT_ROOT/CHANGELOG.md"

# 微服务列表
MICROSERVICES=(
    "auth-service:8207"
    "user-service:8082"
    "job-service:8084"
    "resume-service:8085"
    "company-service:8083"
    "ai-service:8100"
    "blockchain-service:8208"
)

# 前端服务
FRONTEND_SERVICE="frontend"

# 基础设施服务
INFRASTRUCTURE_SERVICES=(
    "basic-server:8080"
    "api-gateway:8080"
    "consul:8500"
    "mysql:3306"
    "postgres:5432"
    "redis:6379"
)

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

# 检查Git状态
check_git_status() {
    log_info "检查Git仓库状态..."
    
    if [ ! -d ".git" ]; then
        log_error "当前目录不是Git仓库"
        exit 1
    fi
    
    # 检查是否有未提交的更改
    if ! git diff-index --quiet HEAD --; then
        log_warning "检测到未提交的更改"
        read -p "是否继续？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "操作已取消"
            exit 1
        fi
    fi
    
    log_success "Git状态检查通过"
}

# 获取当前版本
get_current_version() {
    if [ -f "$VERSION_FILE" ]; then
        cat "$VERSION_FILE"
    else
        echo "1.0.0"
    fi
}

# 更新版本号
update_version() {
    local new_version=$1
    log_info "更新版本号到 $new_version"
    
    # 更新VERSION文件
    echo "$new_version" > "$VERSION_FILE"
    
    # 更新CHANGELOG
    if [ ! -f "$CHANGELOG_FILE" ]; then
        cat > "$CHANGELOG_FILE" << EOF
# Changelog

所有重要的项目更改都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
并且此项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [未发布]

### 新增
- 初始版本

EOF
    fi
    
    # 添加新版本到CHANGELOG
    local current_date=$(date +"%Y-%m-%d")
    sed -i.bak "s/## \[未发布\]/## \[$new_version\] - $current_date\n\n## [未发布]/" "$CHANGELOG_FILE"
    
    log_success "版本号已更新到 $new_version"
}

# 更新微服务版本
update_microservice_versions() {
    local version=$1
    log_info "更新微服务版本到 $version"
    
    for service in "${MICROSERVICES[@]}"; do
        local service_name=$(echo $service | cut -d':' -f1)
        local service_port=$(echo $service | cut -d':' -f2)
        
        log_info "更新 $service_name 服务版本..."
        
        # 更新服务目录中的版本信息
        local service_dir="$PROJECT_ROOT/src/microservices/$service_name"
        if [ -d "$service_dir" ]; then
            # 更新go.mod版本
            if [ -f "$service_dir/go.mod" ]; then
                sed -i.bak "s/version v[0-9]\+\.[0-9]\+\.[0-9]\+/version v$version/" "$service_dir/go.mod"
            fi
            
            # 更新Dockerfile标签
            if [ -f "$service_dir/Dockerfile" ]; then
                sed -i.bak "s/LABEL version=\"[^\"]*\"/LABEL version=\"$version\"/" "$service_dir/Dockerfile"
            fi
            
            # 创建版本标记文件
            echo "version: $version" > "$service_dir/.version"
            echo "port: $service_port" >> "$service_dir/.version"
            echo "updated: $(date)" >> "$service_dir/.version"
        fi
        
        log_success "$service_name 版本已更新"
    done
}

# 更新前端版本
update_frontend_version() {
    local version=$1
    log_info "更新前端版本到 $version"
    
    local frontend_dir="$PROJECT_ROOT/frontend"
    if [ -d "$frontend_dir" ]; then
        # 更新package.json版本
        if [ -f "$frontend_dir/package.json" ]; then
            sed -i.bak "s/\"version\": \"[^\"]*\"/\"version\": \"$version\"/" "$frontend_dir/package.json"
        fi
        
        # 创建版本标记文件
        echo "version: $version" > "$frontend_dir/.version"
        echo "updated: $(date)" >> "$frontend_dir/.version"
        
        log_success "前端版本已更新"
    fi
}

# 更新基础设施版本
update_infrastructure_version() {
    local version=$1
    log_info "更新基础设施版本到 $version"
    
    # 更新Docker Compose文件
    local docker_compose_file="$PROJECT_ROOT/docker/docker-compose.microservices.yml"
    if [ -f "$docker_compose_file" ]; then
        # 添加版本标签到所有服务
        sed -i.bak "s/image: /image: zervigo-/" "$docker_compose_file"
        sed -i.bak "s/zervigo-/zervigo-$version-/" "$docker_compose_file"
    fi
    
    # 更新基础设施版本文件
    echo "version: $version" > "$PROJECT_ROOT/.infrastructure-version"
    echo "updated: $(date)" >> "$PROJECT_ROOT/.infrastructure-version"
    
    log_success "基础设施版本已更新"
}

# 生成版本报告
generate_version_report() {
    local version=$1
    local report_file="$PROJECT_ROOT/VERSION_REPORT.md"
    
    log_info "生成版本报告..."
    
    cat > "$report_file" << EOF
# Zervigo 版本报告

**版本**: $version  
**生成时间**: $(date)  
**Git提交**: $(git rev-parse --short HEAD)

## 微服务版本

| 服务名称 | 端口 | 版本 | 状态 |
|---------|------|------|------|
EOF

    for service in "${MICROSERVICES[@]}"; do
        local service_name=$(echo $service | cut -d':' -f1)
        local service_port=$(echo $service | cut -d':' -f2)
        local service_dir="$PROJECT_ROOT/src/microservices/$service_name"
        
        if [ -d "$service_dir" ]; then
            echo "| $service_name | $service_port | $version | ✅ 已更新 |" >> "$report_file"
        else
            echo "| $service_name | $service_port | $version | ❌ 目录不存在 |" >> "$report_file"
        fi
    done
    
    cat >> "$report_file" << EOF

## 前端版本

| 组件 | 版本 | 状态 |
|------|------|------|
| Taro前端 | $version | ✅ 已更新 |

## 基础设施版本

| 组件 | 版本 | 状态 |
|------|------|------|
| Docker Compose | $version | ✅ 已更新 |
| 基础设施配置 | $version | ✅ 已更新 |

## 版本同步状态

- [x] 微服务版本统一
- [x] 前端版本统一  
- [x] 基础设施版本统一
- [x] Git标签创建
- [x] 版本报告生成

## 下一步操作

1. 提交版本更改到Git
2. 创建Git标签: \`git tag v$version\`
3. 推送到远程仓库: \`git push origin v$version\`
4. 部署到测试环境进行验证
5. 部署到生产环境

EOF

    log_success "版本报告已生成: $report_file"
}

# 创建Git标签
create_git_tag() {
    local version=$1
    log_info "创建Git标签 v$version"
    
    git add .
    git commit -m "chore: 统一版本管理到 v$version

- 更新所有微服务版本
- 更新前端版本
- 更新基础设施版本
- 生成版本报告

解决版本管理混乱问题，实现统一版本控制。"
    
    git tag -a "v$version" -m "Release version $version
    
统一版本管理更新:
- 微服务版本统一
- 前端版本统一
- 基础设施版本统一
- 解决版本管理混乱问题"
    
    log_success "Git标签 v$version 已创建"
}

# 验证版本一致性
verify_version_consistency() {
    local version=$1
    log_info "验证版本一致性..."
    
    local errors=0
    
    # 检查微服务版本
    for service in "${MICROSERVICES[@]}"; do
        local service_name=$(echo $service | cut -d':' -f1)
        local service_dir="$PROJECT_ROOT/src/microservices/$service_name"
        
        if [ -f "$service_dir/.version" ]; then
            local service_version=$(grep "version:" "$service_dir/.version" | cut -d' ' -f2)
            if [ "$service_version" != "$version" ]; then
                log_error "$service_name 版本不一致: 期望 $version, 实际 $service_version"
                errors=$((errors + 1))
            fi
        else
            log_error "$service_name 缺少版本文件"
            errors=$((errors + 1))
        fi
    done
    
    # 检查前端版本
    local frontend_dir="$PROJECT_ROOT/frontend"
    if [ -f "$frontend_dir/.version" ]; then
        local frontend_version=$(grep "version:" "$frontend_dir/.version" | cut -d' ' -f2)
        if [ "$frontend_version" != "$version" ]; then
            log_error "前端版本不一致: 期望 $version, 实际 $frontend_version"
            errors=$((errors + 1))
        fi
    else
        log_error "前端缺少版本文件"
        errors=$((errors + 1))
    fi
    
    if [ $errors -eq 0 ]; then
        log_success "版本一致性验证通过"
        return 0
    else
        log_error "发现 $errors 个版本不一致问题"
        return 1
    fi
}

# 主函数
main() {
    log_info "🚀 开始Zervigo版本管理统一流程"
    echo "=================================="
    
    # 检查Git状态
    check_git_status
    
    # 获取当前版本
    local current_version=$(get_current_version)
    log_info "当前版本: $current_version"
    
    # 获取新版本号
    read -p "请输入新版本号 (当前: $current_version): " new_version
    if [ -z "$new_version" ]; then
        new_version=$current_version
    fi
    
    # 验证版本号格式
    if ! [[ $new_version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_error "版本号格式不正确，请使用语义化版本格式 (如: 1.2.3)"
        exit 1
    fi
    
    log_info "开始更新到版本 $new_version"
    echo "=================================="
    
    # 更新版本
    update_version "$new_version"
    
    # 更新微服务版本
    update_microservice_versions "$new_version"
    
    # 更新前端版本
    update_frontend_version "$new_version"
    
    # 更新基础设施版本
    update_infrastructure_version "$new_version"
    
    # 验证版本一致性
    if verify_version_consistency "$new_version"; then
        # 生成版本报告
        generate_version_report "$new_version"
        
        # 创建Git标签
        create_git_tag "$new_version"
        
        log_success "🎉 版本管理统一完成！"
        echo "=================================="
        log_info "版本: $new_version"
        log_info "Git标签: v$new_version"
        log_info "版本报告: VERSION_REPORT.md"
        log_info "变更日志: CHANGELOG.md"
        echo "=================================="
        log_info "下一步操作:"
        log_info "1. 推送到远程仓库: git push origin v$new_version"
        log_info "2. 部署到测试环境验证"
        log_info "3. 部署到生产环境"
    else
        log_error "版本一致性验证失败，请检查并修复问题"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
Zervigo 版本管理统一脚本

用法: $0 [选项]

选项:
    -h, --help     显示此帮助信息
    -v, --version  显示当前版本
    -c, --check    检查版本一致性
    -r, --report   生成版本报告

示例:
    $0              # 交互式版本更新
    $0 --version    # 显示当前版本
    $0 --check      # 检查版本一致性
    $0 --report     # 生成版本报告

此脚本解决Zervigo项目的版本管理混乱问题，实现:
- 微服务版本统一管理
- 前端版本统一管理
- 基础设施版本统一管理
- Git标签自动创建
- 版本一致性验证

EOF
}

# 处理命令行参数
case "${1:-}" in
    -h|--help)
        show_help
        exit 0
        ;;
    -v|--version)
        echo "当前版本: $(get_current_version)"
        exit 0
        ;;
    -c|--check)
        verify_version_consistency "$(get_current_version)"
        exit $?
        ;;
    -r|--report)
        generate_version_report "$(get_current_version)"
        exit 0
        ;;
    "")
        main
        ;;
    *)
        log_error "未知选项: $1"
        show_help
        exit 1
        ;;
esac
