-- ============================================
-- 手动执行SQL - 需要表所有者权限
-- 如果遇到权限错误，请使用postgres超级用户执行
-- ============================================

-- 为职位表添加tenant_id（如果不存在）
ALTER TABLE zervigo_jobs ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1;
CREATE INDEX IF NOT EXISTS idx_jobs_tenant_id ON zervigo_jobs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_jobs_tenant_created ON zervigo_jobs(tenant_id, created_at);
CREATE INDEX IF NOT EXISTS idx_jobs_tenant_user ON zervigo_jobs(tenant_id, created_by);

-- 为用户资料表添加tenant_id（如果不存在）
ALTER TABLE zervigo_user_profiles ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1;
CREATE INDEX IF NOT EXISTS idx_user_profiles_tenant_id ON zervigo_user_profiles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_user_profiles_tenant_user ON zervigo_user_profiles(tenant_id, user_id);

-- 为职位申请表添加tenant_id（如果不存在）
ALTER TABLE zervigo_job_applications ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1;
CREATE INDEX IF NOT EXISTS idx_job_applications_tenant_id ON zervigo_job_applications(tenant_id);
CREATE INDEX IF NOT EXISTS idx_job_applications_tenant_job ON zervigo_job_applications(tenant_id, job_id);

-- 为用户表添加last_tenant_id（如果不存在）
ALTER TABLE zervigo_auth_users ADD COLUMN IF NOT EXISTS last_tenant_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_users_last_tenant_id ON zervigo_auth_users(last_tenant_id);

-- 为角色表添加tenant_id（如果存在表）
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'zervigo_auth_roles') THEN
        ALTER TABLE zervigo_auth_roles ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1;
        CREATE INDEX IF NOT EXISTS idx_roles_tenant_id ON zervigo_auth_roles(tenant_id);
    END IF;
END $$;

-- 为权限表添加tenant_id（如果存在表）
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'zervigo_auth_permissions') THEN
        ALTER TABLE zervigo_auth_permissions ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1;
        CREATE INDEX IF NOT EXISTS idx_permissions_tenant_id ON zervigo_auth_permissions(tenant_id);
    END IF;
END $$;

-- 更新现有数据的tenant_id
UPDATE zervigo_jobs SET tenant_id = 1 WHERE tenant_id IS NULL OR tenant_id = 0;
UPDATE zervigo_user_profiles SET tenant_id = 1 WHERE tenant_id IS NULL OR tenant_id = 0;
UPDATE zervigo_job_applications SET tenant_id = 1 WHERE tenant_id IS NULL OR tenant_id = 0;

