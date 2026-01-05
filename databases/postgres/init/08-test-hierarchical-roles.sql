-- Go-Zervi 层级化角色体系测试脚本
-- 创建时间: 2025-01-29

-- ==================== 测试1: 测试角色层级函数 ====================

-- 测试获取super_admin的所有子角色
SELECT '=== 测试1: super_admin的子角色 ===' AS test_name;
SELECT * FROM get_role_children(1);

-- 测试获取system_admin的所有父角色
SELECT '=== 测试2: system_admin的父角色 ===' AS test_name;
SELECT * FROM get_role_parents(2);

-- 测试获取enterprise_hr的所有父角色
SELECT '=== 测试3: enterprise_hr的父角色 ===' AS test_name;
SELECT * FROM get_role_parents(16);

-- ==================== 测试2: 创建测试用户并分配角色 ====================

-- 创建测试用户（如果不存在）
INSERT INTO zervigo_auth_users (username, email, password_hash, status)
VALUES 
    ('test_super_admin', 'superadmin@test.com', '$2a$10$test', 'active'),
    ('test_admin', 'admin@test.com', '$2a$10$test', 'active'),
    ('test_manager', 'manager@test.com', '$2a$10$test', 'active'),
    ('test_user', 'user@test.com', '$2a$10$test', 'active')
ON CONFLICT (username) DO NOTHING;

-- 获取用户ID
DO $$
DECLARE
    super_admin_id BIGINT;
    admin_id BIGINT;
    manager_id BIGINT;
    user_id BIGINT;
    super_role_id BIGINT;
    admin_role_id BIGINT;
    manager_role_id BIGINT;
    user_role_id BIGINT;
BEGIN
    -- 获取用户ID
    SELECT id INTO super_admin_id FROM zervigo_auth_users WHERE username = 'test_super_admin';
    SELECT id INTO admin_id FROM zervigo_auth_users WHERE username = 'test_admin';
    SELECT id INTO manager_id FROM zervigo_auth_users WHERE username = 'test_manager';
    SELECT id INTO user_id FROM zervigo_auth_users WHERE username = 'test_user';
    
    -- 获取角色ID
    SELECT id INTO super_role_id FROM zervigo_auth_roles WHERE role_name = 'super_admin';
    SELECT id INTO admin_role_id FROM zervigo_auth_roles WHERE role_name = 'system_admin';
    SELECT id INTO manager_role_id FROM zervigo_auth_roles WHERE role_name = 'enterprise_manager';
    SELECT id INTO user_role_id FROM zervigo_auth_roles WHERE role_name = 'user';
    
    -- 分配角色
    INSERT INTO zervigo_auth_user_roles (user_id, role_id, status)
    VALUES 
        (super_admin_id, super_role_id, 'active'),
        (admin_id, admin_role_id, 'active'),
        (manager_id, manager_role_id, 'active'),
        (user_id, user_role_id, 'active')
    ON CONFLICT ON CONSTRAINT zervigo_auth_user_roles_user_id_role_id_key DO UPDATE SET status = 'active';
    
    RAISE NOTICE '✅ 测试用户角色分配完成';
END $$;

-- ==================== 测试3: 测试层级权限检查 ====================

-- 设置当前用户ID（模拟用户登录）
SELECT '=== 测试3: 层级权限检查 ===' AS test_name;

-- 测试：检查test_admin用户是否拥有system_admin权限（应该返回true，因为它直接拥有该角色）
DO $$
DECLARE
    test_user_id BIGINT;
    has_permission BOOLEAN;
BEGIN
    SELECT id INTO test_user_id FROM zervigo_auth_users WHERE username = 'test_admin';
    
    -- 设置会话变量
    PERFORM set_config('app.current_user_id', test_user_id::text, true);
    
    -- 检查是否拥有system_admin角色
    SELECT EXISTS(
        SELECT 1
        FROM zervigo_auth_user_roles ur
        JOIN zervigo_auth_roles r ON ur.role_id = r.id
        WHERE ur.user_id = test_user_id
        AND r.role_name = 'system_admin'
    ) INTO has_permission;
    
    RAISE NOTICE 'test_admin 拥有 system_admin 角色: %', has_permission;
END $$;

-- ==================== 测试4: 查看角色层级关系 ====================

SELECT '=== 测试4: 角色层级关系可视化 ===' AS test_name;

-- 使用递归CTE可视化角色层级
WITH RECURSIVE role_tree AS (
    -- 根角色
    SELECT 
        id,
        role_name,
        pid,
        id_path,
        path_name,
        0 AS level,
        role_name::text AS hierarchy_path
    FROM zervigo_auth_roles
    WHERE pid = 0
    
    UNION ALL
    
    -- 子角色
    SELECT 
        r.id,
        r.role_name,
        r.pid,
        r.id_path,
        r.path_name,
        rt.level + 1,
        rt.hierarchy_path || ' -> ' || r.role_name
    FROM zervigo_auth_roles r
    JOIN role_tree rt ON r.pid = rt.id
)
SELECT 
    level,
    REPEAT('  ', level) || role_name AS role_tree,
    path_name,
    hierarchy_path
FROM role_tree
ORDER BY id_path, level;

-- ==================== 测试5: 统计角色使用情况 ====================

SELECT '=== 测试5: 角色使用统计 ===' AS test_name;

SELECT 
    r.role_name,
    r.role_level,
    COUNT(DISTINCT ur.user_id) AS user_count,
    COUNT(DISTINCT rp.permission_id) AS permission_count,
    CASE 
        WHEN r.pid = 0 THEN 'ROOT'
        ELSE pr.role_name
    END AS parent_role
FROM zervigo_auth_roles r
LEFT JOIN zervigo_auth_user_roles ur ON r.id = ur.role_id AND ur.status = 'active'
LEFT JOIN zervigo_auth_role_permissions rp ON r.id = rp.role_id
LEFT JOIN zervigo_auth_roles pr ON r.pid = pr.id
GROUP BY r.id, r.role_name, r.role_level, r.pid, pr.role_name
ORDER BY r.role_level DESC, r.role_name;

-- ==================== 测试6: 验证PostgreSQL数据库角色 ====================

SELECT '=== 测试6: PostgreSQL数据库角色列表 ===' AS test_name;

SELECT 
    rolname AS role_name,
    rolsuper AS is_superuser,
    rolcanlogin AS can_login
FROM pg_roles
WHERE rolname LIKE 'zervigo_%'
ORDER BY rolname;

-- ==================== 完成提示 ====================

SELECT '✅ 层级化角色体系测试完成！' AS result;
