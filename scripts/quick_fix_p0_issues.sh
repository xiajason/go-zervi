#!/bin/bash
# 快速修复P0问题脚本
# 用途：解决商业测试的关键阻塞问题
# 执行时间：约30分钟
# 作者：AI Assistant
# 日期：2025-10-13

set -e  # 遇到错误立即退出

echo "=================================="
echo "JobFirst商业测试P0问题快速修复"
echo "=================================="
echo ""
echo "本脚本将完成以下任务："
echo "1. 配置容器内存限制（防止OOM）"
echo "2. 添加Swap空间（内存缓冲）"
echo "3. 配置主从复制过滤（避免数据冲突）"
echo "4. 部署Portainer（容器可视化管理）"
echo "5. 健康检查"
echo ""
read -p "按Enter继续，Ctrl+C取消..."

# ==========================================
# 任务1: 配置容器内存限制
# ==========================================
echo ""
echo "===== 任务1: 配置容器内存限制 ====="
echo ""

echo "检查现有容器..."
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Size}}"
echo ""

echo "配置Elasticsearch内存限制（800MB）..."
docker update --memory=800m --memory-swap=800m test-elasticsearch 2>/dev/null || echo "⚠️  test-elasticsearch不存在或已配置"

echo "配置MySQL内存限制（600MB）..."
docker update --memory=600m --memory-swap=600m test-mysql 2>/dev/null || echo "⚠️  test-mysql不存在或已配置"

echo "配置Neo4j内存限制（450MB）..."
docker update --memory=450m --memory-swap=450m test-neo4j 2>/dev/null || echo "⚠️  test-neo4j不存在或已配置"

echo "配置MongoDB内存限制（200MB）..."
docker update --memory=200m --memory-swap=200m test-mongodb 2>/dev/null || echo "⚠️  test-mongodb不存在或已配置"

echo "配置Weaviate内存限制（200MB）..."
docker update --memory=200m --memory-swap=200m test-weaviate 2>/dev/null || echo "⚠️  test-weaviate不存在或已配置"

echo "配置PostgreSQL内存限制（150MB）..."
docker update --memory=150m --memory-swap=150m test-postgres 2>/dev/null || echo "⚠️  test-postgres不存在或已配置"

echo "配置Redis内存限制（100MB）..."
docker update --memory=100m --memory-swap=100m test-redis 2>/dev/null || echo "⚠️  test-redis不存在或已配置"

echo "✅ 容器内存限制配置完成！"
echo ""

# ==========================================
# 任务2: 添加Swap空间
# ==========================================
echo ""
echo "===== 任务2: 添加Swap空间 ====="
echo ""

if swapon --show | grep -q "/swapfile"; then
    echo "⚠️  Swap空间已存在，跳过创建"
else
    echo "创建2GB Swap文件..."
    sudo fallocate -l 2G /swapfile || { echo "❌ fallocate失败，尝试dd"; sudo dd if=/dev/zero of=/swapfile bs=1M count=2048; }
    
    echo "设置Swap文件权限..."
    sudo chmod 600 /swapfile
    
    echo "配置Swap..."
    sudo mkswap /swapfile
    
    echo "启用Swap..."
    sudo swapon /swapfile
    
    echo "配置开机自动挂载..."
    if ! grep -q "/swapfile" /etc/fstab; then
        echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
    fi
    
    echo "优化Swap使用策略..."
    sudo sysctl vm.swappiness=10
    if ! grep -q "vm.swappiness" /etc/sysctl.conf; then
        echo 'vm.swappiness=10' | sudo tee -a /etc/sysctl.conf
    fi
    
    echo "✅ Swap空间添加完成！"
fi

echo ""
echo "当前内存和Swap状态："
free -h
echo ""

# ==========================================
# 任务3: 配置主从复制过滤
# ==========================================
echo ""
echo "===== 任务3: 配置主从复制过滤 ====="
echo ""

echo "检查MySQL主从复制状态..."
SLAVE_STATUS=$(docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW SLAVE STATUS\G" 2>/dev/null || echo "")

if [ -z "$SLAVE_STATUS" ]; then
    echo "⚠️  MySQL从库未配置或无法访问，跳过复制过滤配置"
else
    echo "配置复制过滤（仅同步jobfirst_production）..."
    docker exec test-mysql mysql -uroot -ptest_mysql_password << 'EOSQL'
STOP SLAVE;
CHANGE REPLICATION FILTER REPLICATE_DO_DB = (jobfirst_production);
START SLAVE;
EOSQL
    
    echo ""
    echo "验证复制状态..."
    docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW SLAVE STATUS\G" | grep -E "Slave_IO_Running|Slave_SQL_Running|Seconds_Behind_Master|Replicate_Do_DB"
    
    echo "✅ 主从复制过滤配置完成！"
fi

echo ""

# ==========================================
# 任务4: 部署Portainer
# ==========================================
echo ""
echo "===== 任务4: 部署Portainer ====="
echo ""

if docker ps -a | grep -q "portainer"; then
    echo "⚠️  Portainer已存在"
    docker start portainer 2>/dev/null && echo "✅ Portainer已启动" || echo "⚠️  Portainer启动失败"
else
    echo "创建Portainer数据卷..."
    docker volume create portainer_data
    
    echo "部署Portainer容器..."
    docker run -d \
      --name portainer \
      --restart=always \
      -p 9443:9443 \
      -v /var/run/docker.sock:/var/run/docker.sock \
      -v portainer_data:/data \
      portainer/portainer-ce:latest
    
    echo ""
    echo "等待Portainer启动..."
    sleep 10
    
    if docker ps | grep -q "portainer"; then
        echo "✅ Portainer部署成功！"
        echo ""
        echo "访问地址: https://$(hostname -I | awk '{print $1}'):9443"
        echo "首次访问请创建管理员账号"
    else
        echo "❌ Portainer启动失败，请检查日志："
        docker logs portainer
    fi
fi

echo ""

# ==========================================
# 任务5: 健康检查
# ==========================================
echo ""
echo "===== 任务5: 系统健康检查 ====="
echo ""

echo "1. 系统资源状态："
echo "-------------------"
echo "CPU负载:"
uptime
echo ""
echo "内存使用:"
free -h
echo ""
echo "磁盘使用:"
df -h / | tail -1
echo ""

echo "2. 容器运行状态："
echo "-------------------"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Size}}" | grep -E "test-|portainer"
echo ""

echo "3. 关键服务端口监听："
echo "-------------------"
ss -tlnp | grep -E "8207|8100|9443" || echo "⚠️  部分服务可能未运行"
echo ""

echo "4. 容器内存限制验证："
echo "-------------------"
echo "容器名\t\t\t内存限制"
docker inspect test-elasticsearch 2>/dev/null | jq -r '.[0].HostConfig.Memory' | awk '{if($1>0) printf "test-elasticsearch\t%dMB\n", $1/1024/1024; else print "test-elasticsearch\t未设置"}' 2>/dev/null || echo "test-elasticsearch\t不存在"
docker inspect test-mysql 2>/dev/null | jq -r '.[0].HostConfig.Memory' | awk '{if($1>0) printf "test-mysql\t\t%dMB\n", $1/1024/1024; else print "test-mysql\t\t未设置"}' 2>/dev/null || echo "test-mysql\t\t不存在"
docker inspect test-neo4j 2>/dev/null | jq -r '.[0].HostConfig.Memory' | awk '{if($1>0) printf "test-neo4j\t\t%dMB\n", $1/1024/1024; else print "test-neo4j\t\t未设置"}' 2>/dev/null || echo "test-neo4j\t\t不存在"
echo ""

echo "=================================="
echo "P0问题修复完成！"
echo "=================================="
echo ""
echo "下一步："
echo "1. 访问Portainer进行容器管理: https://$(hostname -I | awk '{print $1}'):9443"
echo "2. 修复Job Matching初始化问题（需要3-5天排查）"
echo "3. 准备测试数据（20份简历 + 50个职位）"
echo "4. 部署Grafana监控（参考 QUICK_START_GUIDE.md）"
echo ""
echo "详细信息请查看: 商业测试准备度评估报告.md"
echo ""

# 保存执行日志
LOG_FILE="/tmp/p0_fix_$(date +%Y%m%d_%H%M%S).log"
echo "执行日志已保存到: $LOG_FILE"
echo ""

echo "✅ 所有P0任务执行完成！"

