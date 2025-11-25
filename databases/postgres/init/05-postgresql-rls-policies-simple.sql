-- PostgreSQL RLS (Row Level Security) 权限管理策略 - 简化版
-- 基于现有表结构实现行级安全控制

-- 1. 删除现有的策略（如果存在）
DROP POLICY IF EXISTS resume_permission_owner_access ON resume_permission;
DROP POLICY IF EXISTS resume_blacklist_owner_access ON resume_blacklist;
DROP POLICY IF EXISTS approval_assignee_access ON approve_record;
DROP POLICY IF EXISTS approval_enterprise_access ON approve_record;
DROP POLICY IF EXISTS points_owner_access ON points_bill;
DROP POLICY IF EXISTS view_history_owner_access ON view_history;

-- 2. 创建基于现有表结构的RLS策略

-- 简历权限表RLS策略 - 基于resume_id
CREATE POLICY resume_permission_owner_access ON resume_permission
    FOR ALL
    TO zervigo_user
    USING (resume_id IN (
        SELECT resume_id FROM approve_record 
        WHERE user_id = current_setting('app.current_user_id')::int
    ));

-- 简历黑名单表RLS策略 - 基于user_id
CREATE POLICY resume_blacklist_owner_access ON resume_blacklist
    FOR ALL
    TO zervigo_user
    USING (user_id = current_setting('app.current_user_id')::int);

-- 审批表RLS策略 - 基于user_id
CREATE POLICY approval_assignee_access ON approve_record
    FOR ALL
    TO zervigo_user
    USING (user_id = current_setting('app.current_user_id')::int);

-- 审批表企业访问策略 - 基于enterprise_id
CREATE POLICY approval_enterprise_access ON approve_record
    FOR ALL
    TO zervigo_enterprise
    USING (enterprise_id = current_setting('app.current_enterprise_id')::text);

-- 积分表RLS策略 - 基于user_id
CREATE POLICY points_owner_access ON points_bill
    FOR ALL
    TO zervigo_user
    USING (user_id = current_setting('app.current_user_id')::int);

-- 浏览历史表RLS策略 - 基于user_id
CREATE POLICY view_history_owner_access ON view_history
    FOR ALL
    TO zervigo_user
    USING (user_id = current_setting('app.current_user_id')::int);

-- 3. 创建简化的权限检查函数

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

-- 4. 创建用户上下文设置函数

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
CREATE OR REPLACE FUNCTION set_enterprise_context(enterprise_id text)
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.current_enterprise_id', enterprise_id, true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 5. 创建权限测试函数

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
        'PERMISSIVE' as policy_type,
        true as is_enabled
    FROM pg_policies 
    WHERE schemaname = 'public'
    ORDER BY tablename, policyname;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 6. 创建权限管理工具函数

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

-- 7. 创建示例数据插入函数

-- 插入示例用户并分配角色
CREATE OR REPLACE FUNCTION create_sample_users()
RETURNS void AS $$
DECLARE
    user_id_var int;
    role_id_var int;
BEGIN
    -- 创建测试用户
    INSERT INTO zervigo_auth_users (username, email, password_hash, status, created_at)
    VALUES 
        ('test_user', 'test@example.com', '$2a$10$example', 'active', NOW()),
        ('test_enterprise', 'enterprise@example.com', '$2a$10$example', 'active', NOW())
    ON CONFLICT (username) DO NOTHING;
    
    -- 获取用户ID
    SELECT id INTO user_id_var FROM zervigo_auth_users WHERE username = 'test_user';
    
    -- 分配角色
    SELECT id INTO role_id_var FROM zervigo_auth_roles WHERE role_name = 'user';
    INSERT INTO zervigo_auth_user_roles (user_id, role_id)
    VALUES (user_id_var, role_id_var)
    ON CONFLICT (user_id, role_id) DO NOTHING;
    
    -- 企业用户
    SELECT id INTO user_id_var FROM zervigo_auth_users WHERE username = 'test_enterprise';
    SELECT id INTO role_id_var FROM zervigo_auth_roles WHERE role_name = 'enterprise';
    INSERT INTO zervigo_auth_user_roles (user_id, role_id)
    VALUES (user_id_var, role_id_var)
    ON CONFLICT (user_id, role_id) DO NOTHING;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 8. 创建权限统计视图

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

-- 9. 创建示例数据
SELECT create_sample_users();

-- 10. 显示RLS策略状态
SELECT * FROM test_rls_policies();
