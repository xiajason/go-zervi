-- VueCMF字段定义
-- 为用户管理、角色管理、权限管理添加字段配置

-- 确保model_config有记录
INSERT INTO model_config (id, table_name, label, app_id, status, created_at, updated_at)
VALUES 
  (1, 'admin', '管理员', 1, 10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
  (3, 'roles', '角色', 1, 10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
  (6, 'permissions', '权限', 1, 10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (table_name) DO UPDATE SET
  label = EXCLUDED.label,
  updated_at = CURRENT_TIMESTAMP;

-- 清空旧的字段定义（如果有）
DELETE FROM model_field WHERE model_id IN (1, 3, 6);

-- ============================================
-- 1. 管理员（admin）字段定义
-- ============================================
INSERT INTO model_field (model_id, field_name, label, field_type, is_show, sort_num, status) VALUES
(1, 'id', 'ID', 'number', 10, 1, 10),
(1, 'username', '用户名', 'text', 10, 2, 10),
(1, 'email', '邮箱', 'text', 10, 3, 10),
(1, 'phone', '手机号', 'text', 10, 4, 10),
(1, 'role', '角色', 'text', 10, 5, 10),
(1, 'status', '状态', 'switch', 10, 6, 10),
(1, 'last_login_ip', '最后登录IP', 'text', 10, 7, 10),
(1, 'last_login_at', '最后登录时间', 'datetime', 10, 8, 10),
(1, 'created_at', '创建时间', 'datetime', 10, 9, 10);

-- ============================================
-- 2. 角色（roles）字段定义
-- ============================================
INSERT INTO model_field (model_id, field_name, label, field_type, is_show, sort_num, status) VALUES
(3, 'id', 'ID', 'number', 10, 1, 10),
(3, 'role_name', '角色名称', 'text', 10, 2, 10),
(3, 'description', '描述', 'text', 10, 3, 10),
(3, 'status', '状态', 'switch', 10, 4, 10),
(3, 'created_at', '创建时间', 'datetime', 10, 5, 10);

-- ============================================
-- 3. 权限（permissions）字段定义
-- ============================================
INSERT INTO model_field (model_id, field_name, label, field_type, is_show, sort_num, status) VALUES
(6, 'id', 'ID', 'number', 10, 1, 10),
(6, 'permission_name', '权限名称', 'text', 10, 2, 10),
(6, 'resource', '资源', 'text', 10, 3, 10),
(6, 'action', '操作', 'text', 10, 4, 10),
(6, 'description', '描述', 'text', 10, 5, 10),
(6, 'status', '状态', 'switch', 10, 6, 10),
(6, 'created_at', '创建时间', 'datetime', 10, 7, 10);

-- 验证
SELECT 
  mc.table_name,
  mc.label as model_label,
  COUNT(mf.id) as field_count
FROM model_config mc
LEFT JOIN model_field mf ON mc.id = mf.model_id
WHERE mc.table_name IN ('admin', 'roles', 'permissions')
GROUP BY mc.id, mc.table_name, mc.label
ORDER BY mc.id;

