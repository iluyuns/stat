-- MySQL 数据库初始化脚本

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS stat CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE stat;

-- 创建站点管理表
CREATE TABLE IF NOT EXISTS stat_site (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    site_id VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    remark VARCHAR(255),
    created_by VARCHAR(64),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_site_id (site_id),
    INDEX idx_name (name),
    INDEX idx_created_by (created_by),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建用户表
CREATE TABLE IF NOT EXISTS stat_user (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    email VARCHAR(128) NOT NULL UNIQUE,
    password VARCHAR(128) NOT NULL,
    nickname VARCHAR(64),
    avatar VARCHAR(255),
    status TINYINT DEFAULT 1 COMMENT '1:正常 0:禁用',
    last_login TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_status (status),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建验证码表
CREATE TABLE IF NOT EXISTS stat_verify_code (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(128) NOT NULL,
    code VARCHAR(8) NOT NULL,
    type VARCHAR(16) NOT NULL COMMENT 'register, reset, login',
    used BOOLEAN DEFAULT FALSE,
    expired_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_code (code),
    INDEX idx_type (type),
    INDEX idx_used (used),
    INDEX idx_expired_at (expired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建用户会话表
CREATE TABLE IF NOT EXISTS stat_user_session (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token VARCHAR(128) NOT NULL UNIQUE,
    ip VARCHAR(64),
    ua VARCHAR(255),
    expired_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_token (token),
    INDEX idx_expired_at (expired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入示例数据
INSERT INTO stat_site (site_id, name, remark, created_by) VALUES 
('demo123456', '示例站点', '这是一个示例统计站点', 'admin')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- 插入默认管理员用户（密码：admin123）
INSERT INTO stat_user (username, email, password, nickname, status) VALUES 
('admin', 'admin@example.com', '0192023a7bbd73250516f069df18b500', '管理员', 1)
ON DUPLICATE KEY UPDATE nickname = VALUES(nickname);

-- 创建页面访问统计表
CREATE TABLE IF NOT EXISTS stat_page_view (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    site_id VARCHAR(64) NOT NULL,
    path VARCHAR(255),
    ip VARCHAR(64),
    ua VARCHAR(255),
    referer VARCHAR(255),
    user_id VARCHAR(64),
    city VARCHAR(64),
    province VARCHAR(64),
    isp VARCHAR(64),
    device VARCHAR(64),
    os VARCHAR(64),
    browser VARCHAR(64),
    screen VARCHAR(32),
    net VARCHAR(32),
    duration INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_site_time (site_id, created_at),
    INDEX idx_site_path (site_id, path),
    INDEX idx_user_id (user_id),
    INDEX idx_ip (ip),
    INDEX idx_city (city),
    INDEX idx_province (province),
    INDEX idx_device (device),
    INDEX idx_os (os),
    INDEX idx_browser (browser),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建事件统计表
CREATE TABLE IF NOT EXISTS stat_event (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    site_id VARCHAR(64) NOT NULL,
    event_name VARCHAR(64) NOT NULL,
    event_category VARCHAR(32),
    user_id VARCHAR(64),
    value VARCHAR(255),
    extra TEXT,
    path VARCHAR(255),
    ip VARCHAR(64),
    ua VARCHAR(255),
    referer VARCHAR(255),
    city VARCHAR(64),
    province VARCHAR(64),
    isp VARCHAR(64),
    device VARCHAR(64),
    os VARCHAR(64),
    browser VARCHAR(64),
    screen VARCHAR(32),
    net VARCHAR(32),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_event_site_time (site_id, event_name, created_at),
    INDEX idx_event_category (event_category),
    INDEX idx_site_path (site_id, path),
    INDEX idx_user_id (user_id),
    INDEX idx_event_name (event_name),
    INDEX idx_ip (ip),
    INDEX idx_city (city),
    INDEX idx_province (province),
    INDEX idx_device (device),
    INDEX idx_os (os),
    INDEX idx_browser (browser),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci; 