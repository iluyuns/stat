#!/bin/bash

echo "🚀 启动统计服务..."
echo "📊 数据库重试机制已启用"
echo "⏳ 等待数据库服务就绪..."

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f stat-app

echo "✅ 应用启动完成" 