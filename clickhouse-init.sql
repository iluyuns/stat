-- ClickHouse 初始化脚本
-- 创建数据库
CREATE DATABASE IF NOT EXISTS stat;

-- 使用数据库
USE stat;

-- 创建统计表（如果不存在）
CREATE TABLE IF NOT EXISTS page_views (
    id UInt64,
    site_id String,
    user_id String,
    path String,
    ref String,
    ua String,
    ip String,
    city String,
    province String,
    isp String,
    device String,
    os String,
    browser String,
    created_at DateTime
) ENGINE = MergeTree()
ORDER BY (site_id, created_at);

CREATE TABLE IF NOT EXISTS stat_events (
    id UInt64,
    site_id String,
    event_name String,
    event_category String,
    user_id String,
    value String,
    extra String,
    path String,
    ip String,
    ua String,
    referer String,
    city String,
    province String,
    isp String,
    device String,
    os String,
    browser String,
    screen String,
    net String,
    created_at DateTime
) ENGINE = MergeTree()
ORDER BY (site_id, event_name, created_at);

-- 设置权限（允许默认用户访问）
GRANT ALL ON stat.* TO default; 