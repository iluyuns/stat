#!/bin/bash

echo "🔍 测试 ClickHouse 连接..."

# 等待 ClickHouse 容器启动
echo "⏳ 等待 ClickHouse 容器启动..."
sleep 10

# 测试连接
echo "📊 测试 ClickHouse 连接..."
docker exec stat-clickhouse clickhouse-client --query "SELECT 1" 2>/dev/null

if [ $? -eq 0 ]; then
    echo "✅ ClickHouse 连接成功"
    
    # 检查数据库
    echo "📋 检查数据库..."
    docker exec stat-clickhouse clickhouse-client --query "SHOW DATABASES"
    
    # 检查表
    echo "📋 检查 stat 数据库中的表..."
    docker exec stat-clickhouse clickhouse-client --query "SHOW TABLES FROM stat"
    
else
    echo "❌ ClickHouse 连接失败"
    echo "🔧 尝试重启 ClickHouse 容器..."
    docker-compose restart clickhouse
    sleep 15
    
    # 再次测试
    docker exec stat-clickhouse clickhouse-client --query "SELECT 1" 2>/dev/null
    if [ $? -eq 0 ]; then
        echo "✅ ClickHouse 重启后连接成功"
    else
        echo "❌ ClickHouse 仍然连接失败"
    fi
fi

echo "�� ClickHouse 测试完成" 