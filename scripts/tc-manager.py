#!/usr/bin/env python3
"""
腾讯云服务器管理CLI工具
用途: 本地远程管理腾讯云服务器
作者: AI Assistant
日期: 2025-10-10

使用方法:
  python tc-manager.py status          # 查看系统状态
  python tc-manager.py services        # 查看服务列表
  python tc-manager.py containers      # 查看容器列表
  python tc-manager.py restart zervigo # 重启服务
  python tc-manager.py logs ai-service # 查看日志
"""

import click
import subprocess
import json
from tabulate import tabulate
from datetime import datetime

# 配置
SERVER_IP = "101.33.251.158"
SSH_KEY = "~/Downloads/basic.pem"
SSH_USER = "ubuntu"

def ssh_execute(command):
    """执行SSH命令"""
    ssh_cmd = f"ssh -i {SSH_KEY} {SSH_USER}@{SERVER_IP} '{command}'"
    result = subprocess.run(ssh_cmd, shell=True, capture_output=True, text=True)
    return result.stdout

@click.group()
def cli():
    """腾讯云服务器管理工具"""
    pass

@cli.command()
def status():
    """查看系统状态"""
    click.echo("\n" + "="*50)
    click.echo("系统状态")
    click.echo("="*50 + "\n")
    
    # 获取系统信息
    cmd = """
    echo "CPU使用率: $(top -bn1 | grep 'Cpu(s)' | awk '{print $2}')%"
    echo "内存: $(free -h | awk 'NR==2{printf "%s / %s (%.1f%%)", $3, $2, $3/$2*100}')"
    echo "磁盘: $(df -h / | awk 'NR==2{printf "%s / %s (%s)", $3, $2, $5}')"
    echo "运行时间: $(uptime -p)"
    echo "系统负载: $(uptime | awk -F'load average:' '{print $2}')"
    """
    
    output = ssh_execute(cmd)
    click.echo(output)

@cli.command()
def services():
    """查看服务状态"""
    click.echo("\n" + "="*50)
    click.echo("服务状态")
    click.echo("="*50 + "\n")
    
    cmd = """
    echo "=== Zervigo统一认证 (8207) ==="
    if pgrep -f 'unified-auth' > /dev/null; then
        echo "状态: ✅ 运行中"
        ps -p $(pgrep -f 'unified-auth') -o pid,etime,pmem,pcpu,cmd | tail -1
        echo ""
        curl -s http://localhost:8207/health | python3 -m json.tool 2>/dev/null || echo "健康检查失败"
    else
        echo "状态: ❌ 未运行"
    fi
    
    echo ""
    echo "=== AI Service 1 (8100) ==="
    if pgrep -f 'ai_service_with_zervigo' > /dev/null; then
        echo "状态: ✅ 运行中"
        ps -p $(pgrep -f 'ai_service_with_zervigo') -o pid,etime,pmem,pcpu,cmd | tail -1
        echo ""
        curl -s http://localhost:8100/health | python3 -m json.tool 2>/dev/null || echo "健康检查失败"
    else
        echo "状态: ❌ 未运行"
    fi
    """
    
    output = ssh_execute(cmd)
    click.echo(output)

@cli.command()
def containers():
    """查看容器状态"""
    click.echo("\n" + "="*50)
    click.echo("容器状态")
    click.echo("="*50 + "\n")
    
    cmd = """
    echo "=== 运行中的容器 ==="
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    echo ""
    echo "=== 容器资源使用 ==="
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"
    """
    
    output = ssh_execute(cmd)
    click.echo(output)

@cli.command()
def databases():
    """查看数据库状态"""
    click.echo("\n" + "="*50)
    click.echo("数据库状态")
    click.echo("="*50 + "\n")
    
    cmd = """
    echo "=== MySQL数据库 ==="
    docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW DATABASES LIKE 'jobfirst%';" 2>/dev/null
    echo ""
    echo "=== MySQL连接数 ==="
    docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW STATUS LIKE 'Threads_connected';" 2>/dev/null
    echo ""
    echo "=== 主从复制状态 ==="
    docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW SLAVE STATUS\\G" 2>/dev/null | grep -E "Slave_IO_Running|Slave_SQL_Running|Seconds_Behind_Master"
    echo ""
    echo "=== Redis状态 ==="
    echo -n "PING: "
    docker exec test-redis redis-cli -a test_redis_password PING 2>/dev/null
    """
    
    output = ssh_execute(cmd)
    click.echo(output)

@cli.command()
@click.argument('service_name')
def restart(service_name):
    """重启服务 (zervigo/ai-service/container名)"""
    click.echo(f"\n重启 {service_name}...")
    
    if service_name == 'zervigo':
        cmd = """
        cd /opt/services/zervigo
        pkill unified-auth || true
        sleep 2
        nohup ./unified-auth > logs/unified-auth.log 2>&1 &
        sleep 2
        pgrep -f unified-auth && echo "✅ Zervigo已重启" || echo "❌ 重启失败"
        """
    elif service_name == 'ai-service':
        cmd = """
        cd /opt/services/ai-service-1/current
        pkill -f ai_service_with_zervigo || true
        sleep 2
        source venv/bin/activate
        nohup python ai_service_with_zervigo.py > logs/ai_service_1.log 2>&1 &
        sleep 3
        pgrep -f ai_service_with_zervigo && echo "✅ AI Service已重启" || echo "❌ 重启失败"
        """
    else:
        # 假设是容器名
        cmd = f"docker restart {service_name} && echo '✅ {service_name}已重启' || echo '❌ 重启失败'"
    
    output = ssh_execute(cmd)
    click.echo(output)

@cli.command()
@click.argument('service_name')
@click.option('--lines', default=50, help='显示行数')
def logs(service_name, lines):
    """查看服务日志"""
    click.echo(f"\n{service_name} 日志 (最近{lines}行):\n")
    
    if service_name == 'zervigo':
        cmd = f"tail -n{lines} /opt/services/zervigo/logs/*.log 2>/dev/null | tail -50"
    elif service_name == 'ai-service':
        cmd = f"tail -n{lines} /opt/services/ai-service-1/current/logs/ai_service_1.log"
    else:
        # 容器日志
        cmd = f"docker logs --tail {lines} {service_name}"
    
    output = ssh_execute(cmd)
    click.echo(output)

@cli.command()
def health():
    """全面健康检查"""
    click.echo("\n" + "="*50)
    click.echo("全面健康检查")
    click.echo("="*50 + "\n")
    
    cmd = """
    # 健康检查脚本
    score=100
    
    echo "=== 1. 内存检查 ==="
    mem_percent=$(free | awk 'NR==2{printf "%.0f", $3/$2 * 100}')
    echo "内存使用率: ${mem_percent}%"
    if [ "$mem_percent" -gt 85 ]; then
        echo "❌ 内存过高 (-20分)"
        score=$((score - 20))
    elif [ "$mem_percent" -gt 70 ]; then
        echo "⚠️  内存偏高 (-10分)"
        score=$((score - 10))
    else
        echo "✅ 正常"
    fi
    
    echo ""
    echo "=== 2. 容器检查 ==="
    container_count=$(docker ps -q | wc -l)
    echo "运行中容器: ${container_count}/7"
    if [ "$container_count" -lt 7 ]; then
        echo "❌ 部分容器未运行 (-30分)"
        score=$((score - 30))
    else
        echo "✅ 全部运行"
    fi
    
    echo ""
    echo "=== 3. 服务检查 ==="
    if pgrep -f "unified-auth" > /dev/null; then
        echo "✅ Zervigo运行中"
    else
        echo "❌ Zervigo未运行 (-15分)"
        score=$((score - 15))
    fi
    
    if pgrep -f "ai_service" > /dev/null; then
        echo "✅ AI Service运行中"
    else
        echo "❌ AI Service未运行 (-15分)"
        score=$((score - 15))
    fi
    
    echo ""
    echo "=== 4. 数据库检查 ==="
    if docker exec test-mysql mysqladmin -uroot -ptest_mysql_password ping 2>/dev/null | grep -q "alive"; then
        echo "✅ MySQL正常"
    else
        echo "❌ MySQL异常 (-20分)"
        score=$((score - 20))
    fi
    
    if docker exec test-redis redis-cli -a test_redis_password PING 2>/dev/null | grep -q "PONG"; then
        echo "✅ Redis正常"
    else
        echo "❌ Redis异常 (-10分)"
        score=$((score - 10))
    fi
    
    echo ""
    echo "=========================================="
    echo "健康分数: ${score}/100"
    
    if [ "$score" -ge 90 ]; then
        echo "状态: ✅ 优秀"
    elif [ "$score" -ge 70 ]; then
        echo "状态: ⚠️  良好（建议优化）"
    elif [ "$score" -ge 50 ]; then
        echo "状态: ⚠️  警告（需要处理）"
    else
        echo "状态: 🔴 危险（立即处理）"
    fi
    echo "=========================================="
    """
    
    output = ssh_execute(cmd)
    click.echo(output)

@cli.command()
def quick():
    """快速概览（一屏显示所有关键信息）"""
    click.echo("\n" + "="*70)
    click.echo(f"腾讯云服务器快速概览 - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    click.echo("="*70 + "\n")
    
    cmd = """
    echo "【系统资源】"
    echo "CPU: $(top -bn1 | grep 'Cpu(s)' | awk '{print $2}')% | 内存: $(free | awk 'NR==2{printf "%.1f%%", $3/$2*100}') | 磁盘: $(df / | awk 'NR==2{print $5}')"
    
    echo ""
    echo "【服务状态】"
    pgrep -f unified-auth > /dev/null && echo "✅ Zervigo (8207)" || echo "❌ Zervigo"
    pgrep -f ai_service > /dev/null && echo "✅ AI Service (8100)" || echo "❌ AI Service"
    
    echo ""
    echo "【容器状态】$(docker ps -q | wc -l)/7 运行中"
    docker ps --format "{{.Names}}: {{.Status}}" | grep test- | sed 's/^/  /'
    
    echo ""
    echo "【数据库】"
    docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW STATUS LIKE 'Threads_connected';" 2>/dev/null | tail -1 | awk '{print "  MySQL连接: "$2"/300"}'
    docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW SLAVE STATUS\\G" 2>/dev/null | grep "Seconds_Behind_Master:" | awk '{print "  主从延迟: "$2"秒"}'
    """
    
    output = ssh_execute(cmd)
    click.echo(output)
    
    click.echo("\n提示: 使用 'tc-manager.py --help' 查看所有命令")

@cli.command()
def monitor():
    """打开监控面板（浏览器）"""
    import webbrowser
    
    click.echo("\n打开监控面板...")
    
    urls = {
        'Grafana': f'http://{SERVER_IP}:3000',
        'Portainer': f'https://{SERVER_IP}:9443',
        'Prometheus': f'http://{SERVER_IP}:9090',
    }
    
    for name, url in urls.items():
        click.echo(f"  {name}: {url}")
    
    # 打开Grafana
    webbrowser.open(urls['Grafana'])
    click.echo("\n✅ 已在浏览器中打开Grafana")

@cli.command()
@click.argument('container_name')
def exec_shell(container_name):
    """进入容器Shell"""
    click.echo(f"\n进入容器 {container_name} ...")
    click.echo("(退出请输入: exit)\n")
    
    ssh_cmd = f"ssh -i {SSH_KEY} -t {SSH_USER}@{SERVER_IP} 'docker exec -it {container_name} bash'"
    subprocess.run(ssh_cmd, shell=True)

@cli.command()
def backup():
    """备份数据库"""
    click.echo("\n开始备份数据库...")
    
    cmd = """
    BACKUP_DIR="/opt/backups/databases"
    DATE=$(date +%Y%m%d_%H%M%S)
    mkdir -p "$BACKUP_DIR"
    
    echo "备份MySQL数据库..."
    for db in jobfirst_basic jobfirst_professional jobfirst_future; do
        echo "  备份 $db..."
        docker exec test-mysql mysqldump -uroot -ptest_mysql_password \
            --single-transaction "$db" 2>/dev/null | gzip > "$BACKUP_DIR/${db}_${DATE}.sql.gz"
        echo "  ✅ $db 备份完成"
    done
    
    echo ""
    echo "备份完成！"
    ls -lh "$BACKUP_DIR" | tail -5
    """
    
    output = ssh_execute(cmd)
    click.echo(output)

@cli.command()
def optimize():
    """执行优化建议"""
    click.echo("\n" + "="*50)
    click.echo("系统优化建议")
    click.echo("="*50 + "\n")
    
    cmd = """
    echo "=== 内存优化建议 ==="
    mem_available=$(free -m | awk 'NR==2{print $7}')
    if [ "$mem_available" -lt 500 ]; then
        echo "⚠️  可用内存不足500MB，建议:"
        echo "  1. 设置容器内存限制"
        echo "  2. 添加Swap空间"
        echo "  3. 优化Elasticsearch配置"
    else
        echo "✅ 内存充足"
    fi
    
    echo ""
    echo "=== 容器内存限制检查 ==="
    for container in test-mysql test-elasticsearch test-neo4j; do
        limit=$(docker inspect $container --format '{{.HostConfig.Memory}}')
        if [ "$limit" = "0" ]; then
            echo "⚠️  $container 无内存限制"
        else
            echo "✅ $container 已限制"
        fi
    done
    
    echo ""
    echo "=== 磁盘清理建议 ==="
    docker_images=$(docker images -q | wc -l)
    echo "Docker镜像数: $docker_images"
    if [ "$docker_images" -gt 20 ]; then
        echo "⚠️  镜像过多，可执行: docker image prune -a"
    fi
    
    echo ""
    echo "未使用的卷:"
    docker volume ls -qf dangling=true | wc -l
    echo "可执行清理: docker volume prune"
    """
    
    output = ssh_execute(cmd)
    click.echo(output)

@cli.command()
def deploy():
    """部署监控平台"""
    click.echo("\n" + "="*50)
    click.echo("部署监控平台")
    click.echo("="*50 + "\n")
    
    if not click.confirm('确认部署Grafana + Prometheus + Portainer?'):
        click.echo("已取消")
        return
    
    click.echo("\n上传部署脚本...")
    
    # 上传脚本
    scp_cmd = f"scp -i {SSH_KEY} scripts/quick_deploy_monitoring.sh {SSH_USER}@{SERVER_IP}:/tmp/"
    subprocess.run(scp_cmd, shell=True)
    
    click.echo("执行部署...")
    
    # 执行部署
    cmd = "chmod +x /tmp/quick_deploy_monitoring.sh && /tmp/quick_deploy_monitoring.sh"
    output = ssh_execute(cmd)
    
    click.echo(output)
    
    click.echo("\n✅ 部署完成！")
    click.echo(f"访问Grafana: http://{SERVER_IP}:3000 (admin/Admin@2025)")
    click.echo(f"访问Portainer: https://{SERVER_IP}:9443")

if __name__ == '__main__':
    cli()

