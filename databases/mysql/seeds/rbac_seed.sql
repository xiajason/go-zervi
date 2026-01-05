-- RBAC 种子数据（MySQL）
-- 数据库: jobfirst

START TRANSACTION;

-- 角色
CREATE TABLE IF NOT EXISTS roles (
  id INT PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(64) UNIQUE NOT NULL,
  name VARCHAR(128) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO roles (code, name) VALUES
  ('super_admin', '超级管理员'),
  ('recruiter', '招聘方'),
  ('candidate', '候选人')
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 权限
CREATE TABLE IF NOT EXISTS permissions (
  id INT PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(128) UNIQUE NOT NULL,
  name VARCHAR(128) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO permissions (code, name) VALUES
  ('resume:view','查看简历'),
  ('resume:edit','编辑简历'),
  ('resume:publish','发布简历'),
  ('job:view','查看职位'),
  ('job:publish','发布职位')
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 角色-权限
CREATE TABLE IF NOT EXISTS role_permissions (
  role_code VARCHAR(64) NOT NULL,
  perm_code VARCHAR(128) NOT NULL,
  PRIMARY KEY(role_code, perm_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- super_admin → 全部权限
INSERT IGNORE INTO role_permissions(role_code, perm_code)
SELECT 'super_admin', code FROM permissions;

-- recruiter → 职位查看/发布
INSERT IGNORE INTO role_permissions(role_code, perm_code) VALUES
  ('recruiter','job:view'),
  ('recruiter','job:publish');

-- candidate → 简历查看/编辑
INSERT IGNORE INTO role_permissions(role_code, perm_code) VALUES
  ('candidate','resume:view'),
  ('candidate','resume:edit');

-- 用户-角色（示例：admin=super_admin）
CREATE TABLE IF NOT EXISTS user_roles (
  user_id INT NOT NULL,
  role_code VARCHAR(64) NOT NULL,
  PRIMARY KEY(user_id, role_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO user_roles(user_id, role_code) VALUES (1,'super_admin');

COMMIT;


