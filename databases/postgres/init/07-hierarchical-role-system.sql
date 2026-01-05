-- Go-Zervi 层级化角色体系创建脚本
-- 参考 govuecmf-master 的设计，创建层级化的角色体系
-- 创建时间: 2025-01-29

-- ==================== 更新角色表结构 ====================

-- 为角色表添加层级化字段
ALTER TABLE zervigo_auth_roles 
ADD COLUMN IF NOT EXISTS pid BIGINT DEFAULT 0, -- 父级角色ID
ADD COLUMN IF NOT EXISTS id_path VARCHAR(255) DEFAULT '', -- 角色ID层级路径，英文逗号分隔
ADD COLUMN IF NOT EXISTS path_name VARCHAR(255) DEFAULT '', -- 角色名称层级路径，英文逗号分隔
ADD COLUMN IF NOT EXISTS remark TEXT DEFAULT ''; -- 备注信息

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_roles_pid ON zervigo_auth_roles(pid);
CREATE INDEX IF NOT EXISTS idx_roles_id_path ON zervigo_auth_roles(id_path);

-- ==================== 创建PostgreSQL数据库角色（用于RLS测试） ====================

-- 删除已存在的角色（如果存在）
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'zervigo_super_admin') THEN
        DROP ROLE zervigo_super_admin;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'zervigo_admin') THEN
        DROP ROLE zervigo_admin;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'zervigo_manager') THEN
        DROP ROLE zervigo_manager;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'zervigo_enterprise') THEN
        DROP ROLE zervigo_enterprise;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'zervigo_user') THEN
        DROP ROLE zervigo_user;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'zervigo_guest') THEN
        DROP ROLE zervigo_guest;
    END IF;
END $$;

-- 创建PostgreSQL数据库角色（用于RLS权限控制）
CREATE ROLE zervigo_super_admin; -- 超级管理员
CREATE ROLE zervigo_admin; -- 系统管理员
CREATE ROLE zervigo_manager; -- 部门经理
CREATE ROLE zervigo_enterprise; -- 企业用户
CREATE ROLE zervigo_user; -- 普通用户
CREATE ROLE zervigo_guest; -- 访客用户

-- ==================== 初始化层级化角色数据 ====================

-- 清空现有角色数据（如果需要重新初始化）
-- TRUNCATE TABLE zervigo_auth_roles CASCADE;

-- 插入层级化角色数据
INSERT INTO zervigo_auth_roles (role_name, role_description, role_level, pid, id_path, path_name, remark, status) VALUES
-- 第一层：超级管理员（根角色）
('super_admin', '超级管理员，拥有所有权限', 10, 0, '1', '超级管理员', '系统最高权限角色，可管理所有模块和用户', 'active'),

-- 第二层：系统管理员（super_admin的子角色）
('system_admin', '系统管理员，负责系统配置和维护', 9, 1, '1,2', '超级管理员,系统管理员', '负责系统配置、用户管理、角色权限分配', 'active'),
('app_admin', '应用管理员，负责应用模块管理', 8, 1, '1,3', '超级管理员,应用管理员', '负责应用模块的配置和管理', 'active'),

-- 第三层：业务管理员（system_admin的子角色）
('user_admin', '用户管理员', 7, 2, '1,2,4', '超级管理员,系统管理员,用户管理员', '负责用户账户管理和审核', 'active'),
('role_admin', '角色管理员', 7, 2, '1,2,5', '超级管理员,系统管理员,角色管理员', '负责角色和权限的管理', 'active'),
('content_admin', '内容管理员', 6, 3, '1,3,6', '超级管理员,应用管理员,内容管理员', '负责内容审核和管理', 'active'),

-- 企业角色层级
('enterprise_admin', '企业管理员', 7, 0, '7', '企业管理员', '企业账号管理员，可管理企业信息和招聘', 'active'),
('enterprise_manager', '企业经理', 6, 7, '7,8', '企业管理员,企业经理', '企业部门经理，可管理招聘和简历', 'active'),
('enterprise_hr', '企业HR', 5, 8, '7,8,9', '企业管理员,企业经理,企业HR', '企业HR，可查看简历和发布招聘', 'active'),

-- 普通用户层级
('user', '普通用户', 3, 0, '10', '普通用户', '注册用户，可发布简历和投递职位', 'active'),
('user_premium', '高级用户', 4, 10, '10,11', '普通用户,高级用户', '付费用户，拥有更多功能和权限', 'active'),

-- 访客角色
('guest', '访客用户', 1, 0, '12', '访客用户', '未登录用户，只能浏览公开信息', 'active')

ON CONFLICT (role_name) DO UPDATE SET
    role_description = EXCLUDED.role_description,
    role_level = EXCLUDED.role_level,
    pid = EXCLUDED.pid,
    id_path = EXCLUDED.id_path,
    path_name = EXCLUDED.path_name,
    remark = EXCLUDED.remark,
    updated_at = CURRENT_TIMESTAMP;

-- ==================== 创建层级化角色管理函数 ====================

-- 获取角色的所有子角色（递归）
CREATE OR REPLACE FUNCTION get_role_children(role_id_param BIGINT)
RETURNS TABLE(
    id BIGINT,
    role_name VARCHAR,
    role_level INTEGER,
    pid BIGINT,
    id_path VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE role_tree AS (
        -- 起始角色
        SELECT r.id, r.role_name, r.role_level, r.pid, r.id_path
        FROM zervigo_auth_roles r
        WHERE r.id = role_id_param
        
        UNION ALL
        
        -- 递归查询子角色
        SELECT r.id, r.role_name, r.role_level, r.pid, r.id_path
        FROM zervigo_auth_roles r
        INNER JOIN role_tree rt ON r.pid = rt.id
    )
    SELECT rt.id, rt.role_name, rt.role_level, rt.pid, rt.id_path
    FROM role_tree rt
    WHERE rt.id != role_id_param; -- 排除自身
END;
$$ LANGUAGE plpgsql;

-- 获取角色的所有父角色（递归）
CREATE OR REPLACE FUNCTION get_role_parents(role_id_param BIGINT)
RETURNS TABLE(
    id BIGINT,
    role_name VARCHAR,
    role_level INTEGER,
    pid BIGINT,
    id_path VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE role_tree AS (
        -- 起始角色
        SELECT r.id, r.role_name, r.role_level, r.pid, r.id_path
        FROM zervigo_auth_roles r
        WHERE r.id = role_id_param
        
        UNION ALL
        
        -- 递归查询父角色
        SELECT r.id, r.role_name, r.role_level, r.pid, r.id_path
        FROM zervigo_auth_roles r
        INNER JOIN role_tree rt ON r.id = rt.pid
    )
    SELECT rt.id, rt.role_name, rt.role_level, rt.pid, rt.id_path
    FROM role_tree rt
    WHERE rt.id != role_id_param; -- 排除自身
END;
$$ LANGUAGE plpgsql;

-- 检查用户是否拥有某个角色或其子角色（层级检查）
CREATE OR REPLACE FUNCTION has_role_or_child(role_name_param VARCHAR)
RETURNS BOOLEAN AS $$
DECLARE
    user_id_var INTEGER := get_current_user_id();
    role_id_var BIGINT;
    has_role BOOLEAN := FALSE;
BEGIN
    IF user_id_var IS NULL THEN
        RETURN FALSE;
    END IF;
    
    -- 获取角色ID
    SELECT id INTO role_id_var FROM zervigo_auth_roles WHERE role_name = role_name_param;
    
    IF role_id_var IS NULL THEN
        RETURN FALSE;
    END IF;
    
    -- 检查用户是否拥有该角色或其任何子角色
    SELECT EXISTS(
        SELECT 1
        FROM zervigo_auth_user_roles ur
        JOIN zervigo_auth_roles r ON ur.role_id = r.id
        WHERE ur.user_id = user_id_var
        AND (
            r.id = role_id_var
            OR r.id_path LIKE role_id_var || ',%'
            OR r.id_path LIKE '%,' || role_id_var || ',%'
            OR r.id_path LIKE '%,' || role_id_var
        )
    ) INTO has_role;
    
    RETURN has_role;
END;
$$ LANGUAGE plpgsql STABLE;

-- ==================== 授权PostgreSQL角色权限 ====================

-- 超级管理员：拥有所有权限
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO zervigo_super_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO zervigo_super_admin;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO zervigo_super_admin;

-- 系统管理员：可管理用户和角色
GRANT SELECT, INSERT, UPDATE, DELETE ON zervigo_auth_users TO zervigo_admin;
GRANT SELECT, INSERT, UPDATE, DELETE ON zervigo_auth_roles TO zervigo_admin;
GRANT SELECT, INSERT, UPDATE, DELETE ON zervigo_auth_permissions TO zervigo_admin;
GRANT SELECT, INSERT, UPDATE, DELETE ON zervigo_auth_user_roles TO zervigo_admin;
GRANT SELECT, INSERT, UPDATE, DELETE ON zervigo_auth_role_permissions TO zervigo_admin;

-- 企业管理员：可管理企业相关数据
GRANT SELECT, INSERT, UPDATE ON zervigo_auth_users TO zervigo_enterprise;
GRANT SELECT ON zervigo_auth_roles TO zervigo_enterprise;

-- 普通用户：只能查看自己的数据
GRANT SELECT ON zervigo_auth_roles TO zervigo_user;
GRANT SELECT ON zervigo_auth_permissions TO zervigo_user;

-- 访客：只能查看公开信息
GRANT SELECT ON zervigo_auth_roles TO zervigo_guest;

-- ==================== 创建角色层级视图 ====================

CREATE OR REPLACE VIEW v_role_hierarchy AS
SELECT 
    r.id,
    r.role_name,
    r.role_description,
    r.role_level,
    r.pid,
    r.id_path,
    r.path_name,
    CASE 
        WHEN r.pid = 0 THEN 'ROOT'
        ELSE pr.role_name
    END AS parent_role_name,
    r.status,
    r.created_at,
    r.updated_at,
    (
        SELECT COUNT(*) 
        FROM zervigo_auth_roles cr 
        WHERE cr.pid = r.id
    ) AS child_count,
    (
        SELECT COUNT(*) 
        FROM zervigo_auth_user_roles ur 
        WHERE ur.role_id = r.id AND ur.status = 'active'
    ) AS user_count
FROM zervigo_auth_roles r
LEFT JOIN zervigo_auth_roles pr ON r.pid = pr.id
ORDER BY r.id_path;

-- 授予视图访问权限
GRANT SELECT ON v_role_hierarchy TO zervigo_admin;
GRANT SELECT ON v_role_hierarchy TO zervigo_user;
GRANT SELECT ON v_role_hierarchy TO zervigo_guest;

-- ==================== 输出初始化信息 ====================

DO $$
DECLARE
    role_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO role_count FROM zervigo_auth_roles;
    RAISE NOTICE '✅ 角色体系初始化完成！';
    RAISE NOTICE '📊 已创建 % 个角色', role_count;
    RAISE NOTICE '🔐 已创建 6 个PostgreSQL数据库角色用于RLS测试';
    RAISE NOTICE '📈 已创建角色层级管理函数和视图';
END $$;
