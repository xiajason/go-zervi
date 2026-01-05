-- PostgreSQL RLS (Row Level Security) 权限管理策略 - 修复版
-- 基于用户角色和权限实现行级安全控制

-- 1. 创建数据库角色
DO $$
BEGIN
    -- 创建基础角色
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'zervigo_super_admin') THEN
        CREATE ROLE zervigo_super_admin;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'zervigo_admin') THEN
        CREATE ROLE zervigo_admin;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'zervigo_user') THEN
        CREATE ROLE zervigo_user;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'zervigo_enterprise') THEN
        CREATE ROLE zervigo_enterprise;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'zervigo_guest') THEN
        CREATE ROLE zervigo_guest;
    END IF;
END
$$;

-- 2. 为角色分配权限
-- 超级管理员权限
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO zervigo_super_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO zervigo_super_admin;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO zervigo_super_admin;

-- 管理员权限
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO zervigo_admin;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO zervigo_admin;

-- 普通用户权限
GRANT SELECT, INSERT, UPDATE, DELETE ON zervigo_auth_users TO zervigo_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON resume_permission TO zervigo_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON resume_blacklist TO zervigo_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON approve_record TO zervigo_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON points_bill TO zervigo_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON view_history TO zervigo_user;
GRANT SELECT ON zervigo_auth_roles TO zervigo_user;
GRANT SELECT ON zervigo_auth_permissions TO zervigo_user;
GRANT SELECT ON zervigo_auth_role_permissions TO zervigo_user;
GRANT SELECT ON zervigo_auth_user_roles TO zervigo_user;

-- 企业用户权限
GRANT SELECT ON zervigo_auth_users TO zervigo_enterprise;
GRANT SELECT ON resume_permission TO zervigo_enterprise;
GRANT SELECT, INSERT, UPDATE, DELETE ON approve_record TO zervigo_enterprise;
GRANT SELECT ON zervigo_auth_roles TO zervigo_enterprise;
GRANT SELECT ON zervigo_auth_permissions TO zervigo_enterprise;

-- 访客权限
GRANT SELECT ON zervigo_auth_roles TO zervigo_guest;
GRANT SELECT ON zervigo_auth_permissions TO zervigo_guest;

-- 3. 启用表的行级安全
ALTER TABLE zervigo_auth_users ENABLE ROW LEVEL SECURITY;
ALTER TABLE resume_permission ENABLE ROW LEVEL SECURITY;
ALTER TABLE resume_blacklist ENABLE ROW LEVEL SECURITY;
ALTER TABLE approve_record ENABLE ROW LEVEL SECURITY;
ALTER TABLE points_bill ENABLE ROW LEVEL SECURITY;
ALTER TABLE view_history ENABLE ROW LEVEL SECURITY;

-- 4. 创建RLS策略

-- 用户表RLS策略
CREATE POLICY user_self_access ON zervigo_auth_users
    FOR ALL
    TO zervigo_user
    USING (id = current_setting('app.current_user_id')::int);

CREATE POLICY user_admin_access ON zervigo_auth_users
    FOR ALL
    TO zervigo_admin
    USING (true);

CREATE POLICY user_super_admin_access ON zervigo_auth_users
    FOR ALL
    TO zervigo_super_admin
    USING (true);

-- 简历权限表RLS策略
CREATE POLICY resume_permission_owner_access ON resume_permission
    FOR ALL
    TO zervigo_user
    USING (user_id = current_setting('app.current_user_id')::int);

CREATE POLICY resume_permission_admin_access ON resume_permission
    FOR ALL
    TO zervigo_admin
    USING (true);

CREATE POLICY resume_permission_super_admin_access ON resume_permission
    FOR ALL
    TO zervigo_super_admin
    USING (true);

-- 简历黑名单表RLS策略
CREATE POLICY resume_blacklist_owner_access ON resume_blacklist
    FOR ALL
    TO zervigo_user
    USING (user_id = current_setting('app.current_user_id')::int);

CREATE POLICY resume_blacklist_admin_access ON resume_blacklist
    FOR ALL
    TO zervigo_admin
    USING (true);

CREATE POLICY resume_blacklist_super_admin_access ON resume_blacklist
    FOR ALL
    TO zervigo_super_admin
    USING (true);

-- 审批表RLS策略
CREATE POLICY approval_assignee_access ON approve_record
    FOR ALL
    TO zervigo_user
    USING (assigned_to = current_setting('app.current_user_id')::int);

CREATE POLICY approval_enterprise_access ON approve_record
    FOR ALL
    TO zervigo_enterprise
    USING (enterprise_id = current_setting('app.current_enterprise_id')::int);

CREATE POLICY approval_admin_access ON approve_record
    FOR ALL
    TO zervigo_admin
    USING (true);

CREATE POLICY approval_super_admin_access ON approve_record
    FOR ALL
    TO zervigo_super_admin
    USING (true);

-- 积分表RLS策略
CREATE POLICY points_owner_access ON points_bill
    FOR ALL
    TO zervigo_user
    USING (user_id = current_setting('app.current_user_id')::int);

CREATE POLICY points_admin_access ON points_bill
    FOR ALL
    TO zervigo_admin
    USING (true);

CREATE POLICY points_super_admin_access ON points_bill
    FOR ALL
    TO zervigo_super_admin
    USING (true);

-- 浏览历史表RLS策略
CREATE POLICY view_history_owner_access ON view_history
    FOR ALL
    TO zervigo_user
    USING (user_id = current_setting('app.current_user_id')::int);

CREATE POLICY view_history_admin_access ON view_history
    FOR ALL
    TO zervigo_admin
    USING (true);

CREATE POLICY view_history_super_admin_access ON view_history
    FOR ALL
    TO zervigo_super_admin
    USING (true);

-- 5. 创建基于角色的视图

-- 用户可访问的简历权限视图
CREATE OR REPLACE VIEW user_accessible_resume_permissions AS
SELECT rp.*
FROM resume_permission rp
WHERE rp.user_id = current_setting('app.current_user_id')::int
   OR current_user IN ('zervigo_admin', 'zervigo_super_admin');

-- 用户可访问的审批视图
CREATE OR REPLACE VIEW user_accessible_approvals AS
SELECT a.*
FROM approve_record a
WHERE a.assigned_to = current_setting('app.current_user_id')::int
   OR current_user IN ('zervigo_admin', 'zervigo_super_admin');

-- 企业可访问的审批视图
CREATE OR REPLACE VIEW enterprise_accessible_approvals AS
SELECT a.*
FROM approve_record a
WHERE a.enterprise_id = current_setting('app.current_enterprise_id')::int
   OR current_user IN ('zervigo_admin', 'zervigo_super_admin');

-- 6. 创建权限检查函数

-- 检查用户是否有特定权限
CREATE OR REPLACE FUNCTION has_permission(permission_name text)
RETURNS boolean AS $$
DECLARE
    user_id int;
    has_perm boolean := false;
BEGIN
    -- 获取当前用户ID
    user_id := current_setting('app.current_user_id')::int;
    
    -- 检查用户是否有该权限
    SELECT EXISTS(
        SELECT 1 
        FROM zervigo_auth_user_roles ur
        JOIN zervigo_auth_role_permissions rp ON ur.role_id = rp.role_id
        JOIN zervigo_auth_permissions p ON rp.permission_id = p.id
        WHERE ur.user_id = user_id 
        AND p.permission_code = permission_name
    ) INTO has_perm;
    
    RETURN has_perm;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 检查用户是否有特定角色
CREATE OR REPLACE FUNCTION has_role(role_name text)
RETURNS boolean AS $$
DECLARE
    user_id int;
    has_role boolean := false;
BEGIN
    -- 获取当前用户ID
    user_id := current_setting('app.current_user_id')::int;
    
    -- 检查用户是否有该角色
    SELECT EXISTS(
        SELECT 1 
        FROM zervigo_auth_user_roles ur
        JOIN zervigo_auth_roles r ON ur.role_id = r.id
        WHERE ur.user_id = user_id 
        AND r.role_name = role_name
    ) INTO has_role;
    
    RETURN has_role;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 7. 创建动态权限设置函数

-- 设置当前用户上下文
CREATE OR REPLACE FUNCTION set_user_context(user_id int, username text DEFAULT NULL)
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.current_user_id', user_id::text, true);
    IF username IS NOT NULL THEN
        PERFORM set_config('app.current_username', username, true);
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 设置当前企业上下文
CREATE OR REPLACE FUNCTION set_enterprise_context(enterprise_id int)
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.current_enterprise_id', enterprise_id::text, true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 8. 创建权限测试函数

-- 测试RLS策略
CREATE OR REPLACE FUNCTION test_rls_policies()
RETURNS TABLE(
    table_name text,
    policy_name text,
    policy_type text,
    is_enabled boolean
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        schemaname||'.'||tablename as table_name,
        policyname as policy_name,
        CASE WHEN permissive THEN 'PERMISSIVE' ELSE 'RESTRICTIVE' END as policy_type,
        true as is_enabled
    FROM pg_policies 
    WHERE schemaname = 'public'
    ORDER BY tablename, policyname;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 9. 创建权限审计函数

-- 记录权限检查日志
CREATE OR REPLACE FUNCTION log_permission_check(
    user_id int,
    permission_name text,
    resource_type text,
    resource_id text,
    granted boolean
)
RETURNS void AS $$
BEGIN
    INSERT INTO zervigo_auth_login_logs (
        user_id, action, resource, result, ip_address, user_agent, created_at
    ) VALUES (
        user_id, 
        'permission_check', 
        resource_type || ':' || resource_id,
        CASE WHEN granted THEN 'granted' ELSE 'denied' END,
        inet_client_addr(),
        current_setting('app.user_agent', true),
        NOW()
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 10. 创建示例数据插入函数

-- 插入示例用户并分配角色
CREATE OR REPLACE FUNCTION create_sample_users()
RETURNS void AS $$
DECLARE
    user_id int;
    role_id int;
BEGIN
    -- 创建测试用户
    INSERT INTO zervigo_auth_users (username, email, password_hash, status, created_at)
    VALUES 
        ('test_user', 'test@example.com', '$2a$10$example', 'active', NOW()),
        ('test_enterprise', 'enterprise@example.com', '$2a$10$example', 'active', NOW())
    ON CONFLICT (username) DO NOTHING;
    
    -- 获取用户ID
    SELECT id INTO user_id FROM zervigo_auth_users WHERE username = 'test_user';
    
    -- 分配角色
    SELECT id INTO role_id FROM zervigo_auth_roles WHERE role_name = 'user';
    INSERT INTO zervigo_auth_user_roles (user_id, role_id)
    VALUES (user_id, role_id)
    ON CONFLICT (user_id, role_id) DO NOTHING;
    
    -- 企业用户
    SELECT id INTO user_id FROM zervigo_auth_users WHERE username = 'test_enterprise';
    SELECT id INTO role_id FROM zervigo_auth_roles WHERE role_name = 'enterprise';
    INSERT INTO zervigo_auth_user_roles (user_id, role_id)
    VALUES (user_id, role_id)
    ON CONFLICT (user_id, role_id) DO NOTHING;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 11. 创建权限管理工具函数

-- 获取用户的所有权限
CREATE OR REPLACE FUNCTION get_user_permissions(user_id int)
RETURNS TABLE(permission_code text, permission_name text) AS $$
BEGIN
    RETURN QUERY
    SELECT p.permission_code, p.permission_name
    FROM zervigo_auth_user_roles ur
    JOIN zervigo_auth_role_permissions rp ON ur.role_id = rp.role_id
    JOIN zervigo_auth_permissions p ON rp.permission_id = p.id
    WHERE ur.user_id = user_id
    ORDER BY p.permission_code;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 获取用户的所有角色
CREATE OR REPLACE FUNCTION get_user_roles(user_id int)
RETURNS TABLE(role_name text, role_description text) AS $$
BEGIN
    RETURN QUERY
    SELECT r.role_name, r.role_description
    FROM zervigo_auth_user_roles ur
    JOIN zervigo_auth_roles r ON ur.role_id = r.id
    WHERE ur.user_id = user_id
    ORDER BY r.role_name;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 12. 创建权限统计视图

-- 权限使用统计
CREATE OR REPLACE VIEW permission_usage_stats AS
SELECT 
    p.permission_code,
    p.permission_name,
    COUNT(DISTINCT ur.user_id) as user_count,
    COUNT(DISTINCT r.id) as role_count
FROM zervigo_auth_permissions p
LEFT JOIN zervigo_auth_role_permissions rp ON p.id = rp.permission_id
LEFT JOIN zervigo_auth_roles r ON rp.role_id = r.id
LEFT JOIN zervigo_auth_user_roles ur ON r.id = ur.role_id
GROUP BY p.id, p.permission_code, p.permission_name
ORDER BY user_count DESC;

-- 用户权限统计
CREATE OR REPLACE VIEW user_permission_stats AS
SELECT 
    u.username,
    u.email,
    COUNT(DISTINCT p.id) as permission_count,
    COUNT(DISTINCT r.id) as role_count,
    u.status
FROM zervigo_auth_users u
LEFT JOIN zervigo_auth_user_roles ur ON u.id = ur.user_id
LEFT JOIN zervigo_auth_roles r ON ur.role_id = r.id
LEFT JOIN zervigo_auth_role_permissions rp ON r.id = rp.role_id
LEFT JOIN zervigo_auth_permissions p ON rp.permission_id = p.id
GROUP BY u.id, u.username, u.email, u.status
ORDER BY permission_count DESC;

-- 13. 创建示例数据
SELECT create_sample_users();

-- 14. 显示RLS策略状态
SELECT * FROM test_rls_policies();
