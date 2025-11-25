-- Go-Zervi框架动态路由管理数据库表结构
-- 基于govuecmf框架学习，实现多对多、多对一的路由管理

-- 路由配置表
CREATE TABLE IF NOT EXISTS route_config (
    id BIGSERIAL PRIMARY KEY,
    route_key VARCHAR(100) UNIQUE NOT NULL,        -- 路由唯一标识
    route_name VARCHAR(100) NOT NULL,             -- 路由名称
    route_path VARCHAR(200) NOT NULL,             -- 路由路径
    service_name VARCHAR(50) NOT NULL,            -- 目标服务名
    service_endpoint VARCHAR(200) NOT NULL,       -- 目标服务端点
    method VARCHAR(10) NOT NULL,                  -- HTTP方法
    route_type VARCHAR(20) DEFAULT 'api',         -- 路由类型：api, page, component
    description TEXT,                              -- 路由描述
    is_public BOOLEAN DEFAULT false,              -- 是否公开路由
    is_active BOOLEAN DEFAULT true,               -- 是否启用
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 权限配置表
CREATE TABLE IF NOT EXISTS route_permission (
    id BIGSERIAL PRIMARY KEY,
    route_key VARCHAR(100) NOT NULL,              -- 关联路由
    permission_code VARCHAR(100) NOT NULL,        -- 权限代码
    permission_name VARCHAR(100) NOT NULL,        -- 权限名称
    resource_type VARCHAR(50),                    -- 资源类型
    action VARCHAR(50),                           -- 操作类型
    service_name VARCHAR(50),                     -- 服务名称
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (route_key) REFERENCES route_config(route_key) ON DELETE CASCADE
);

-- 角色路由关联表
CREATE TABLE IF NOT EXISTS role_route_permission (
    id BIGSERIAL PRIMARY KEY,
    role_id INT NOT NULL,                         -- 角色ID
    route_key VARCHAR(100) NOT NULL,              -- 路由标识
    permission_code VARCHAR(100) NOT NULL,        -- 权限代码
    is_granted BOOLEAN DEFAULT true,              -- 是否授权
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (role_id) REFERENCES zervigo_auth_roles(id) ON DELETE CASCADE,
    FOREIGN KEY (route_key) REFERENCES route_config(route_key) ON DELETE CASCADE
);

-- 前端页面配置表
CREATE TABLE IF NOT EXISTS frontend_page_config (
    id BIGSERIAL PRIMARY KEY,
    page_key VARCHAR(100) UNIQUE NOT NULL,        -- 页面唯一标识
    page_name VARCHAR(100) NOT NULL,              -- 页面名称
    page_path VARCHAR(200) NOT NULL,             -- 页面路径
    component_name VARCHAR(100),                  -- 组件名称
    page_type VARCHAR(20) DEFAULT 'page',         -- 页面类型：page, component, modal
    required_routes TEXT,                         -- 所需路由列表(JSON)
    required_permissions TEXT,                    -- 所需权限列表(JSON)
    page_config JSON,                            -- 页面配置(JSON)
    is_active BOOLEAN DEFAULT true,               -- 是否启用
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_route_config_service_name ON route_config(service_name);
CREATE INDEX IF NOT EXISTS idx_route_config_route_type ON route_config(route_type);
CREATE INDEX IF NOT EXISTS idx_route_config_is_active ON route_config(is_active);
CREATE INDEX IF NOT EXISTS idx_route_permission_route_key ON route_permission(route_key);
CREATE INDEX IF NOT EXISTS idx_route_permission_permission_code ON route_permission(permission_code);
CREATE INDEX IF NOT EXISTS idx_role_route_permission_role_id ON role_route_permission(role_id);
CREATE INDEX IF NOT EXISTS idx_role_route_permission_route_key ON role_route_permission(route_key);
CREATE INDEX IF NOT EXISTS idx_frontend_page_config_page_type ON frontend_page_config(page_type);
CREATE INDEX IF NOT EXISTS idx_frontend_page_config_is_active ON frontend_page_config(is_active);

-- 插入示例路由配置数据
INSERT INTO route_config (route_key, route_name, route_path, service_name, service_endpoint, method, route_type, description, is_public, is_active) 
VALUES 
('resume.list', '简历列表', '/api/v1/resume/list', 'resume-service', '/api/v1/resume/list', 'GET', 'api', '获取简历列表', false, true),
('resume.detail', '简历详情', '/api/v1/resume/detail/*', 'resume-service', '/api/v1/resume/detail', 'GET', 'api', '获取简历详情', false, true),
('resume.create', '创建简历', '/api/v1/resume/create', 'resume-service', '/api/v1/resume/create', 'POST', 'api', '创建新简历', false, true),
('resume.update', '更新简历', '/api/v1/resume/update/*', 'resume-service', '/api/v1/resume/update', 'PUT', 'api', '更新简历信息', false, true),
('resume.delete', '删除简历', '/api/v1/resume/delete/*', 'resume-service', '/api/v1/resume/delete', 'DELETE', 'api', '删除简历', false, true),
('resume.permission', '简历权限', '/api/v1/resume/permission/*', 'resume-service', '/api/v1/resume/permission', 'GET', 'api', '获取简历权限配置', false, true),
('resume.blacklist', '简历黑名单', '/api/v1/resume/blacklist/*', 'resume-service', '/api/v1/resume/blacklist', 'GET', 'api', '获取简历黑名单', false, true),
('approve.list', '审批列表', '/api/v1/approve/list', 'resume-service', '/api/v1/approve/list', 'GET', 'api', '获取审批列表', false, true),
('approve.handle', '处理审批', '/api/v1/approve/handle/*', 'resume-service', '/api/v1/approve/handle', 'POST', 'api', '处理审批申请', false, true),
('points.user', '用户积分', '/api/v1/points/user/*', 'resume-service', '/api/v1/points/user', 'GET', 'api', '获取用户积分', false, true),
('roles.list', '角色列表', '/api/v1/roles', 'permission-service', '/api/v1/roles', 'GET', 'api', '获取角色列表', true, true),
('permissions.list', '权限列表', '/api/v1/permissions', 'permission-service', '/api/v1/permissions', 'GET', 'api', '获取权限列表', true, true),
('user.roles', '用户角色', '/api/v1/users/*/roles', 'permission-service', '/api/v1/users', 'GET', 'api', '获取用户角色', false, true),
('role.permissions', '角色权限', '/api/v1/roles/*/permissions', 'permission-service', '/api/v1/roles', 'GET', 'api', '获取角色权限', false, true)
ON CONFLICT (route_key) DO NOTHING;

-- 插入示例权限配置数据
INSERT INTO route_permission (route_key, permission_code, permission_name, resource_type, action, service_name) 
VALUES 
('resume.list', 'resume:view', '简历查看', 'resume', 'view', 'resume-service'),
('resume.detail', 'resume:view', '简历查看', 'resume', 'view', 'resume-service'),
('resume.create', 'resume:create', '简历创建', 'resume', 'create', 'resume-service'),
('resume.update', 'resume:update', '简历更新', 'resume', 'update', 'resume-service'),
('resume.delete', 'resume:delete', '简历删除', 'resume', 'delete', 'resume-service'),
('resume.permission', 'resume:permission', '简历权限管理', 'resume', 'permission', 'resume-service'),
('resume.blacklist', 'resume:blacklist', '简历黑名单管理', 'resume', 'blacklist', 'resume-service'),
('approve.list', 'approve:view', '审批查看', 'approve', 'view', 'resume-service'),
('approve.handle', 'approve:handle', '审批处理', 'approve', 'handle', 'resume-service'),
('points.user', 'points:view', '积分查看', 'points', 'view', 'resume-service'),
('user.roles', 'user:roles', '用户角色管理', 'user', 'roles', 'permission-service'),
('role.permissions', 'role:permissions', '角色权限管理', 'role', 'permissions', 'permission-service')
ON CONFLICT DO NOTHING;

-- 插入示例角色路由权限关联数据
INSERT INTO role_route_permission (role_id, route_key, permission_code, is_granted) 
SELECT r.id, rc.route_key, rp.permission_code, true
FROM zervigo_auth_roles r
CROSS JOIN route_config rc
CROSS JOIN route_permission rp
WHERE r.role_name = 'super_admin' 
AND rc.route_key = rp.route_key
ON CONFLICT DO NOTHING;

-- 为admin角色分配部分权限
INSERT INTO role_route_permission (role_id, route_key, permission_code, is_granted) 
SELECT r.id, rc.route_key, rp.permission_code, true
FROM zervigo_auth_roles r
CROSS JOIN route_config rc
CROSS JOIN route_permission rp
WHERE r.role_name = 'admin' 
AND rc.route_key = rp.route_key
AND rp.permission_code IN ('resume:view', 'resume:create', 'resume:update', 'approve:view', 'approve:handle', 'points:view')
ON CONFLICT DO NOTHING;

-- 为user角色分配基本权限
INSERT INTO role_route_permission (role_id, route_key, permission_code, is_granted) 
SELECT r.id, rc.route_key, rp.permission_code, true
FROM zervigo_auth_roles r
CROSS JOIN route_config rc
CROSS JOIN route_permission rp
WHERE r.role_name = 'user' 
AND rc.route_key = rp.route_key
AND rp.permission_code IN ('resume:view', 'resume:create', 'resume:update', 'points:view')
ON CONFLICT DO NOTHING;

-- 为enterprise角色分配企业权限
INSERT INTO role_route_permission (role_id, route_key, permission_code, is_granted) 
SELECT r.id, rc.route_key, rp.permission_code, true
FROM zervigo_auth_roles r
CROSS JOIN route_config rc
CROSS JOIN route_permission rp
WHERE r.role_name = 'enterprise' 
AND rc.route_key = rp.route_key
AND rp.permission_code IN ('resume:view', 'approve:view', 'approve:handle')
ON CONFLICT DO NOTHING;

-- 插入示例前端页面配置数据
INSERT INTO frontend_page_config (page_key, page_name, page_path, component_name, page_type, required_routes, required_permissions, page_config, is_active) 
VALUES 
('resume.list.page', '简历列表页', '/pages/resume/index', 'ResumeList', 'page', '["resume.list"]', '["resume:view"]', '{"title": "简历列表", "showCreate": true}', true),
('resume.detail.page', '简历详情页', '/pages/resume/detail', 'ResumeDetail', 'page', '["resume.detail", "resume.permission"]', '["resume:view"]', '{"title": "简历详情", "showEdit": true}', true),
('resume.create.page', '创建简历页', '/pages/resume/create', 'ResumeCreate', 'page', '["resume.create"]', '["resume:create"]', '{"title": "创建简历", "showPreview": true}', true),
('resume.edit.page', '编辑简历页', '/pages/resume/edit', 'ResumeEdit', 'page', '["resume.update"]', '["resume:update"]', '{"title": "编辑简历", "showPreview": true}', true),
('approve.list.page', '审批列表页', '/pages/approve/index', 'ApproveList', 'page', '["approve.list"]', '["approve:view"]', '{"title": "审批列表", "showHandle": true}', true),
('approve.detail.page', '审批详情页', '/pages/approve/detail', 'ApproveDetail', 'page', '["approve.handle"]', '["approve:handle"]', '{"title": "审批详情", "showActions": true}', true),
('points.user.page', '用户积分页', '/pages/points/index', 'PointsUser', 'page', '["points.user"]', '["points:view"]', '{"title": "我的积分", "showHistory": true}', true),
('permission.manage.page', '权限管理页', '/pages/permission/index', 'PermissionManage', 'page', '["roles.list", "permissions.list", "user.roles", "role.permissions"]', '["user:roles", "role:permissions"]', '{"title": "权限管理", "showRoles": true, "showPermissions": true}', true)
ON CONFLICT (page_key) DO NOTHING;
