-- 修复管理员密码
-- 密码: admin123
-- MD5哈希: 0192023a7bbd73250516f069df18b500

USE stat;

-- 更新管理员密码
UPDATE stat_user 
SET password = '0192023a7bbd73250516f069df18b500' 
WHERE username = 'admin';

-- 如果管理员用户不存在，则创建
INSERT INTO stat_user (username, email, password, nickname, status) 
SELECT 'admin', 'admin@example.com', '0192023a7bbd73250516f069df18b500', '管理员', 1
WHERE NOT EXISTS (SELECT 1 FROM stat_user WHERE username = 'admin');

SELECT 'Admin password updated successfully!' as message; 