#!/bin/bash
# 本地环境数据库恢复脚本
# 用途: 在本地Docker环境中恢复从腾讯云下载的数据库备份
# 日期: 2025-10-10
# 前置条件: docker-compose.local.yml已启动所有数据库容器

BACKUP_DIR="./backups/databases"

echo "=========================================="
echo "本地环境数据库恢复"
echo "=========================================="

# 检查备份目录
if [ ! -d "$BACKUP_DIR" ]; then
    echo "❌ 备份目录不存在: $BACKUP_DIR"
    echo "请确保已下载并解压备份文件"
    exit 1
fi

# 检查Docker是否运行
if ! docker ps >/dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 等待数据库启动
echo ""
echo "等待数据库容器启动..."
sleep 10

# 1. 恢复MySQL数据库
echo ""
echo "=== 1. 恢复MySQL数据库 ==="

for db in jobfirst_basic jobfirst_professional jobfirst_future; do
    backup_file="$BACKUP_DIR/mysql/${db}.sql.gz"
    
    if [ -f "$backup_file" ]; then
        echo "恢复数据库: $db"
        
        # 创建数据库
        docker exec local-mysql mysql -uroot -plocal_dev_password \
            -e "CREATE DATABASE IF NOT EXISTS $db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null
        
        # 恢复数据
        gunzip -c "$backup_file" | docker exec -i local-mysql mysql -uroot -plocal_dev_password "$db" 2>/dev/null
        
        if [ $? -eq 0 ]; then
            # 检查表数量
            table_count=$(docker exec local-mysql mysql -uroot -plocal_dev_password "$db" -e "SHOW TABLES;" 2>/dev/null | wc -l)
            table_count=$((table_count - 1))
            echo "  ✅ $db 恢复完成 ($table_count 个表)"
        else
            echo "  ❌ $db 恢复失败"
        fi
    else
        echo "  ⚠️  备份文件不存在: $backup_file"
    fi
done

# 创建本地用户（与远程隔离）
echo ""
echo "创建本地数据库用户..."
docker exec local-mysql mysql -uroot -plocal_dev_password << 'EOSQL' 2>/dev/null
-- 基础版用户
CREATE USER IF NOT EXISTS 'jobfirst_basic_user'@'%' IDENTIFIED BY 'local_dev_password';
GRANT SELECT, INSERT, UPDATE, DELETE ON jobfirst_basic.* TO 'jobfirst_basic_user'@'%';

-- 专业版用户
CREATE USER IF NOT EXISTS 'jobfirst_pro_user'@'%' IDENTIFIED BY 'local_dev_password';
GRANT SELECT, INSERT, UPDATE, DELETE ON jobfirst_professional.* TO 'jobfirst_pro_user'@'%';

-- Future版用户
CREATE USER IF NOT EXISTS 'jobfirst_future_user'@'%' IDENTIFIED BY 'local_dev_password';
GRANT SELECT, INSERT, UPDATE, DELETE ON jobfirst_future.* TO 'jobfirst_future_user'@'%';

FLUSH PRIVILEGES;
EOSQL

echo "  ✅ 本地用户创建完成"

# 2. 恢复PostgreSQL
echo ""
echo "=== 2. 恢复PostgreSQL ==="

backup_file="$BACKUP_DIR/postgres/all_databases.sql.gz"

if [ -f "$backup_file" ]; then
    echo "恢复PostgreSQL数据库..."
    gunzip -c "$backup_file" | docker exec -i local-postgres psql -U postgres 2>/dev/null
    
    if [ $? -eq 0 ]; then
        # 检查数据库列表
        db_count=$(docker exec local-postgres psql -U postgres -t -c "SELECT count(*) FROM pg_database WHERE datistemplate = false;" 2>/dev/null | tr -d ' ')
        echo "  ✅ PostgreSQL恢复完成 ($db_count 个数据库)"
    else
        echo "  ❌ PostgreSQL恢复失败"
    fi
else
    echo "  ⚠️  备份文件不存在: $backup_file"
fi

# 3. 恢复MongoDB
echo ""
echo "=== 3. 恢复MongoDB ==="

backup_file="$BACKUP_DIR/mongodb/dump.archive.gz"

if [ -f "$backup_file" ]; then
    echo "恢复MongoDB数据库..."
    gunzip -c "$backup_file" | docker exec -i local-mongodb mongorestore --archive --username admin --password local_dev_password --authenticationDatabase admin 2>/dev/null
    
    if [ $? -eq 0 ]; then
        # 检查数据库列表
        db_count=$(docker exec local-mongodb mongosh --username admin --password local_dev_password --authenticationDatabase admin --quiet --eval "db.adminCommand('listDatabases').databases.length" 2>/dev/null)
        echo "  ✅ MongoDB恢复完成 ($db_count 个数据库)"
    else
        echo "  ❌ MongoDB恢复失败"
    fi
else
    echo "  ⚠️  备份文件不存在: $backup_file"
fi

# 4. 恢复Redis
echo ""
echo "=== 4. 恢复Redis ==="

backup_file="$BACKUP_DIR/redis/dump.rdb.gz"

if [ -f "$backup_file" ]; then
    echo "恢复Redis数据..."
    
    # 停止Redis以便替换RDB文件
    docker stop local-redis 2>/dev/null
    
    # 解压并复制RDB文件
    gunzip -c "$backup_file" > /tmp/dump.rdb
    docker cp /tmp/dump.rdb local-redis:/data/dump.rdb
    rm /tmp/dump.rdb
    
    # 重启Redis
    docker start local-redis
    sleep 3
    
    # 验证
    if docker exec local-redis redis-cli -a local_dev_password PING 2>/dev/null | grep -q "PONG"; then
        # 检查键数量
        key_count=$(docker exec local-redis redis-cli -a local_dev_password DBSIZE 2>/dev/null | awk '{print $2}')
        echo "  ✅ Redis恢复完成 ($key_count 个键)"
    else
        echo "  ❌ Redis恢复失败"
    fi
else
    echo "  ⚠️  备份文件不存在: $backup_file"
fi

# 5. 恢复Neo4j (DAO/区块链数据)
echo ""
echo "=== 5. 恢复Neo4j (DAO/区块链) ==="

dump_file="$BACKUP_DIR/neo4j/neo4j.dump.gz"
data_file="$BACKUP_DIR/neo4j/data.tar.gz"

if [ -f "$dump_file" ]; then
    echo "恢复Neo4j数据（dump方式）..."
    
    # 停止Neo4j
    docker stop local-neo4j 2>/dev/null
    
    # 解压dump文件
    gunzip -c "$dump_file" > /tmp/neo4j.dump
    docker cp /tmp/neo4j.dump local-neo4j:/tmp/
    rm /tmp/neo4j.dump
    
    # 启动并加载
    docker start local-neo4j
    sleep 5
    
    # 使用neo4j-admin加载dump
    docker exec local-neo4j neo4j-admin database load neo4j --from-path=/tmp --overwrite-destination=true 2>/dev/null
    
    if [ $? -eq 0 ]; then
        echo "  ✅ Neo4j恢复完成"
        docker restart local-neo4j
        sleep 5
    else
        echo "  ❌ Neo4j dump恢复失败"
    fi
    
elif [ -f "$data_file" ]; then
    echo "恢复Neo4j数据（data目录方式）..."
    
    # 停止Neo4j
    docker stop local-neo4j 2>/dev/null
    
    # 清空现有数据
    docker exec local-neo4j rm -rf /data/* 2>/dev/null || true
    
    # 解压并复制数据目录
    mkdir -p /tmp/neo4j_restore
    tar -xzf "$data_file" -C /tmp/neo4j_restore/
    docker cp /tmp/neo4j_restore/data/. local-neo4j:/data/
    rm -rf /tmp/neo4j_restore
    
    # 重启Neo4j
    docker start local-neo4j
    sleep 10
    
    echo "  ✅ Neo4j恢复完成"
else
    echo "  ⚠️  Neo4j备份文件不存在"
fi

# 6. 验证恢复结果
echo ""
echo "=========================================="
echo "验证恢复结果"
echo "=========================================="

echo ""
echo "MySQL:"
docker exec local-mysql mysql -uroot -plocal_dev_password -e "
SELECT 
    SCHEMA_NAME as '数据库',
    ROUND(SUM(DATA_LENGTH + INDEX_LENGTH) / 1024 / 1024, 2) as '大小(MB)',
    COUNT(*) as '表数量'
FROM information_schema.TABLES 
WHERE SCHEMA_NAME IN ('jobfirst_basic', 'jobfirst_professional', 'jobfirst_future')
GROUP BY SCHEMA_NAME;
" 2>/dev/null

echo ""
echo "PostgreSQL:"
docker exec local-postgres psql -U postgres -c "\l" 2>/dev/null | grep -E "Name|jobfirst|postgres" | head -10

echo ""
echo "MongoDB:"
docker exec local-mongodb mongosh --username admin --password local_dev_password --authenticationDatabase admin --quiet --eval "db.adminCommand('listDatabases')" 2>/dev/null

echo ""
echo "Redis:"
docker exec local-redis redis-cli -a local_dev_password INFO keyspace 2>/dev/null | grep "^db"

echo ""
echo "Neo4j:"
if curl -s http://localhost:7474 >/dev/null 2>&1; then
    echo "  ✅ Neo4j HTTP接口可访问 (http://localhost:7474)"
    echo "  账号: neo4j / local_dev_password"
else
    echo "  ⚠️  Neo4j可能还在启动中，请稍后访问 http://localhost:7474"
fi

echo ""
echo "=========================================="
echo "✅ 数据库恢复完成！"
echo "=========================================="
echo ""
echo "访问信息:"
echo "  MySQL:      localhost:3306 (root / local_dev_password)"
echo "  PostgreSQL: localhost:5432 (postgres / local_dev_password)"
echo "  MongoDB:    localhost:27017 (admin / local_dev_password)"
echo "  Redis:      localhost:6379 (密码: local_dev_password)"
echo "  Neo4j:      http://localhost:7474 (neo4j / local_dev_password)"
echo "  Adminer:    http://localhost:8888 (Web数据库管理)"
echo ""
echo "测试连接:"
echo "  docker exec local-mysql mysql -uroot -plocal_dev_password -e 'SHOW DATABASES;'"
echo "  docker exec local-postgres psql -U postgres -l"
echo "  docker exec local-mongodb mongosh --username admin --password local_dev_password --eval 'show dbs'"
echo "  docker exec local-redis redis-cli -a local_dev_password PING"
echo ""

