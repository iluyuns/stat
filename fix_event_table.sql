-- 修复事件表结构迁移脚本
-- 为现有的 stat_event 表添加缺失的字段

-- MySQL 版本
USE stat;

-- 检查并添加缺失的字段
ALTER TABLE stat_event 
ADD COLUMN IF NOT EXISTS ip VARCHAR(64) AFTER path,
ADD COLUMN IF NOT EXISTS ua VARCHAR(255) AFTER ip,
ADD COLUMN IF NOT EXISTS referer VARCHAR(255) AFTER ua,
ADD COLUMN IF NOT EXISTS event_category VARCHAR(32) AFTER event_name,
ADD COLUMN IF NOT EXISTS device VARCHAR(64) AFTER isp,
ADD COLUMN IF NOT EXISTS os VARCHAR(64) AFTER device,
ADD COLUMN IF NOT EXISTS browser VARCHAR(64) AFTER os,
ADD COLUMN IF NOT EXISTS screen VARCHAR(32) AFTER browser,
ADD COLUMN IF NOT EXISTS net VARCHAR(32) AFTER screen;

-- 添加索引
ALTER TABLE stat_event 
ADD INDEX IF NOT EXISTS idx_ip (ip),
ADD INDEX IF NOT EXISTS idx_event_category (event_category),
ADD INDEX IF NOT EXISTS idx_device (device),
ADD INDEX IF NOT EXISTS idx_os (os),
ADD INDEX IF NOT EXISTS idx_browser (browser);

-- 更新现有数据的事件类型分类
UPDATE stat_event SET event_category = 
  CASE 
    WHEN LOWER(event_name) LIKE '%click%' OR LOWER(event_name) LIKE '%tap%' OR LOWER(event_name) LIKE '%button%' THEN '点击事件'
    WHEN LOWER(event_name) LIKE '%scroll%' THEN '滚动事件'
    WHEN LOWER(event_name) LIKE '%form%' OR LOWER(event_name) LIKE '%submit%' OR LOWER(event_name) LIKE '%input%' THEN '表单事件'
    WHEN LOWER(event_name) LIKE '%page%' OR LOWER(event_name) LIKE '%view%' OR LOWER(event_name) LIKE '%load%' THEN '页面事件'
    WHEN LOWER(event_name) LIKE '%user%' OR LOWER(event_name) LIKE '%interaction%' OR LOWER(event_name) LIKE '%action%' THEN '用户交互'
    WHEN LOWER(event_name) LIKE '%custom%' OR LOWER(event_name) LIKE '%custom_%' THEN '自定义事件'
    ELSE '其他事件'
  END
WHERE event_category IS NULL OR event_category = '';

-- ClickHouse 版本（如果需要重新创建表）
-- 注意：ClickHouse 不支持 ALTER TABLE ADD COLUMN，需要重新创建表
-- 以下是重新创建表的 SQL（请谨慎使用，会丢失数据）

/*
-- 备份现有数据（如果表存在）
CREATE TABLE IF NOT EXISTS stat_events_backup AS SELECT * FROM stat_events;

-- 删除旧表
DROP TABLE IF EXISTS stat_events;

-- 创建新表
CREATE TABLE stat_events (
    id UInt64,
    site_id String,
    event_name String,
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

-- 恢复数据（如果需要）
INSERT INTO stat_events SELECT * FROM stat_events_backup;
*/ 