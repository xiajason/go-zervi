-- ============================================
-- 租户表和多租户支持
-- 创建时间: 2025-01-XX
-- 说明: 为GoZervi SaaS系统添加多租户支持
-- 参考: CordysCRM的organization_id实现模式
-- ============================================

-- 租户表
CREATE TABLE IF NOT EXISTS zervigo_tenants (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    
    -- 状态信息
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
    
    -- 配置信息（JSON格式存储租户特定配置）
    settings JSONB DEFAULT '{}',
    
    -- 配额信息（JSON格式存储资源配额）
    quota JSONB DEFAULT '{}',
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- 索引
    CONSTRAINT idx_tenants_code UNIQUE (code),
    CONSTRAINT idx_tenants_status CHECK (status IN ('active', 'suspended', 'deleted'))
);

CREATE INDEX IF NOT EXISTS idx_tenants_code ON zervigo_tenants(code);
CREATE INDEX IF NOT EXISTS idx_tenants_status ON zervigo_tenants(status);
CREATE INDEX IF NOT EXISTS idx_tenants_created_at ON zervigo_tenants(created_at);

COMMENT ON TABLE zervigo_tenants IS '租户表';
COMMENT ON COLUMN zervigo_tenants.id IS '租户ID';
COMMENT ON COLUMN zervigo_tenants.name IS '租户名称';
COMMENT ON COLUMN zervigo_tenants.code IS '租户代码（唯一标识）';
COMMENT ON COLUMN zervigo_tenants.description IS '租户描述';
COMMENT ON COLUMN zervigo_tenants.status IS '状态: active-活跃, suspended-暂停, deleted-已删除';
COMMENT ON COLUMN zervigo_tenants.settings IS '租户配置（JSON格式）';
COMMENT ON COLUMN zervigo_tenants.quota IS '资源配额（JSON格式）';

-- 用户-租户关联表
CREATE TABLE IF NOT EXISTS zervigo_user_tenants (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES zervigo_auth_users(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL REFERENCES zervigo_tenants(id) ON DELETE CASCADE,
    
    -- 角色信息（在该租户中的角色）
    role VARCHAR(50) DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member', 'guest')),
    
    -- 状态信息
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    
    -- 时间戳
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 唯一约束：一个用户在一个租户中只能有一条记录
    CONSTRAINT uk_user_tenant UNIQUE(user_id, tenant_id)
);

CREATE INDEX IF NOT EXISTS idx_user_tenants_user_id ON zervigo_user_tenants(user_id);
CREATE INDEX IF NOT EXISTS idx_user_tenants_tenant_id ON zervigo_user_tenants(tenant_id);
CREATE INDEX IF NOT EXISTS idx_user_tenants_role ON zervigo_user_tenants(role);
CREATE INDEX IF NOT EXISTS idx_user_tenants_status ON zervigo_user_tenants(status);

COMMENT ON TABLE zervigo_user_tenants IS '用户-租户关联表';
COMMENT ON COLUMN zervigo_user_tenants.id IS '关联ID';
COMMENT ON COLUMN zervigo_user_tenants.user_id IS '用户ID';
COMMENT ON COLUMN zervigo_user_tenants.tenant_id IS '租户ID';
COMMENT ON COLUMN zervigo_user_tenants.role IS '角色: owner-所有者, admin-管理员, member-成员, guest-访客';
COMMENT ON COLUMN zervigo_user_tenants.status IS '状态: active-活跃, inactive-非活跃';
COMMENT ON COLUMN zervigo_user_tenants.joined_at IS '加入租户时间';

-- 为用户表添加last_tenant_id字段（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'zervigo_auth_users' 
        AND column_name = 'last_tenant_id'
    ) THEN
        ALTER TABLE zervigo_auth_users ADD COLUMN last_tenant_id BIGINT;
        CREATE INDEX IF NOT EXISTS idx_users_last_tenant_id ON zervigo_auth_users(last_tenant_id);
        COMMENT ON COLUMN zervigo_auth_users.last_tenant_id IS '最后使用的租户ID';
    END IF;
END $$;

-- 创建默认租户（ID=1）
INSERT INTO zervigo_tenants (id, name, code, description, status)
VALUES (1, 'Default Tenant', 'default', '默认租户', 'active')
ON CONFLICT (id) DO NOTHING;

-- 为现有用户分配默认租户
INSERT INTO zervigo_user_tenants (user_id, tenant_id, role, status)
SELECT id, 1, 'member', 'active'
FROM zervigo_auth_users
WHERE id NOT IN (SELECT user_id FROM zervigo_user_tenants)
ON CONFLICT (user_id, tenant_id) DO NOTHING;

