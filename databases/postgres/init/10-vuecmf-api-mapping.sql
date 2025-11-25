-- VueCMF API 映射表
-- 用于 VueCMF 前端的 API 路径映射

-- 创建 API 映射表
CREATE TABLE IF NOT EXISTS vuecmf_api_map (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(100) NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    api_path VARCHAR(255) NOT NULL,
    request_method VARCHAR(10) NOT NULL DEFAULT 'GET',
    note TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(table_name, action_type)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_vuecmf_api_map_table_name ON vuecmf_api_map(table_name);
CREATE INDEX IF NOT EXISTS idx_vuecmf_api_map_action_type ON vuecmf_api_map(action_type);

-- 插入核心 API 映射
INSERT INTO vuecmf_api_map (table_name, action_type, api_path, request_method, note) VALUES
-- 认证相关
('admin', 'login', '/api/v1/auth/login', 'POST', '管理员登录'),
('user', 'login', '/api/v1/auth/login', 'POST', '用户登录'),
('admin', 'logout', '/api/v1/auth/logout', 'POST', '退出登录'),
('user', 'info', '/api/v1/auth/user/info', 'GET', '获取当前用户信息'),

-- 菜单和导航
('menu', 'nav', '/api/v1/router/menu/nav', 'GET', '获取导航菜单'),
('menu', 'list', '/api/v1/router/menu/list', 'GET', '菜单列表'),

-- 用户管理
('admin', 'index', '/api/v1/users/list', 'GET', '用户列表'),
('admin', 'save', '/api/v1/users', 'POST', '创建用户'),
('admin', 'update', '/api/v1/users', 'PUT', '更新用户'),
('admin', 'delete', '/api/v1/users', 'DELETE', '删除用户'),
('admin', 'detail', '/api/v1/users/detail', 'GET', '用户详情'),

-- 角色管理
('roles', 'index', '/api/v1/permission/roles', 'GET', '角色列表'),
('roles', 'save', '/api/v1/permission/roles', 'POST', '创建角色'),
('roles', 'update', '/api/v1/permission/roles', 'PUT', '更新角色'),
('roles', 'delete', '/api/v1/permission/roles', 'DELETE', '删除角色'),

-- 权限管理
('permissions', 'index', '/api/v1/permission/permissions', 'GET', '权限列表'),
('permissions', 'save', '/api/v1/permission/permissions', 'POST', '创建权限'),
('permissions', 'update', '/api/v1/permission/permissions', 'PUT', '更新权限'),
('permissions', 'delete', '/api/v1/permission/permissions', 'DELETE', '删除权限')

ON CONFLICT (table_name, action_type) DO NOTHING;

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_vuecmf_api_map_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_vuecmf_api_map_updated_at ON vuecmf_api_map;
CREATE TRIGGER trigger_update_vuecmf_api_map_updated_at
    BEFORE UPDATE ON vuecmf_api_map
    FOR EACH ROW
    EXECUTE FUNCTION update_vuecmf_api_map_updated_at();

-- 输出初始化完成信息
DO $$
BEGIN
    RAISE NOTICE 'VueCMF API 映射表初始化完成！';
    RAISE NOTICE '已添加 % 条 API 映射', (SELECT COUNT(*) FROM vuecmf_api_map);
END $$;

