-- ============================================
-- 为现有业务表添加tenant_id字段
-- 创建时间: 2025-01-XX
-- 说明: 为所有业务表添加tenant_id字段以实现数据隔离
-- ============================================

-- 为职位表添加tenant_id
-- 注意：需要表所有者权限，如果遇到权限错误，请使用postgres超级用户执行
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'zervigo_jobs') THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'zervigo_jobs' 
            AND column_name = 'tenant_id'
        ) THEN
            ALTER TABLE zervigo_jobs ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
            CREATE INDEX IF NOT EXISTS idx_jobs_tenant_id ON zervigo_jobs(tenant_id);
            CREATE INDEX IF NOT EXISTS idx_jobs_tenant_created ON zervigo_jobs(tenant_id, created_at);
            CREATE INDEX IF NOT EXISTS idx_jobs_tenant_user ON zervigo_jobs(tenant_id, created_by);
            COMMENT ON COLUMN zervigo_jobs.tenant_id IS '租户ID';
        END IF;
    END IF;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE '无法为zervigo_jobs添加tenant_id字段: %', SQLERRM;
END $$;

-- 为用户资料表添加tenant_id
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'zervigo_user_profiles') THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'zervigo_user_profiles' 
            AND column_name = 'tenant_id'
        ) THEN
            ALTER TABLE zervigo_user_profiles ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
            CREATE INDEX IF NOT EXISTS idx_user_profiles_tenant_id ON zervigo_user_profiles(tenant_id);
            CREATE INDEX IF NOT EXISTS idx_user_profiles_tenant_user ON zervigo_user_profiles(tenant_id, user_id);
            COMMENT ON COLUMN zervigo_user_profiles.tenant_id IS '租户ID';
        END IF;
    END IF;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE '无法为zervigo_user_profiles添加tenant_id字段: %', SQLERRM;
END $$;

-- 为公司表添加tenant_id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'zervigo_companies' 
        AND column_name = 'tenant_id'
    ) THEN
        ALTER TABLE zervigo_companies ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
        CREATE INDEX IF NOT EXISTS idx_companies_tenant_id ON zervigo_companies(tenant_id);
        CREATE INDEX IF NOT EXISTS idx_companies_tenant_created ON zervigo_companies(tenant_id, created_at);
        COMMENT ON COLUMN zervigo_companies.tenant_id IS '租户ID';
    END IF;
END $$;

-- 为简历表添加tenant_id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'zervigo_resumes' 
        AND column_name = 'tenant_id'
    ) THEN
        ALTER TABLE zervigo_resumes ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
        CREATE INDEX IF NOT EXISTS idx_resumes_tenant_id ON zervigo_resumes(tenant_id);
        CREATE INDEX IF NOT EXISTS idx_resumes_tenant_user ON zervigo_resumes(tenant_id, user_id);
        COMMENT ON COLUMN zervigo_resumes.tenant_id IS '租户ID';
    END IF;
END $$;

-- 为职位申请表添加tenant_id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'zervigo_job_applications' 
        AND column_name = 'tenant_id'
    ) THEN
        ALTER TABLE zervigo_job_applications ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
        CREATE INDEX IF NOT EXISTS idx_job_applications_tenant_id ON zervigo_job_applications(tenant_id);
        CREATE INDEX IF NOT EXISTS idx_job_applications_tenant_job ON zervigo_job_applications(tenant_id, job_id);
        COMMENT ON COLUMN zervigo_job_applications.tenant_id IS '租户ID';
    END IF;
END $$;

-- 为角色表添加tenant_id（如果存在）
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'zervigo_auth_roles') THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'zervigo_auth_roles' 
            AND column_name = 'tenant_id'
        ) THEN
            ALTER TABLE zervigo_auth_roles ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
            CREATE INDEX IF NOT EXISTS idx_roles_tenant_id ON zervigo_auth_roles(tenant_id);
            COMMENT ON COLUMN zervigo_auth_roles.tenant_id IS '租户ID';
        END IF;
    END IF;
END $$;

-- 为权限表添加tenant_id（如果存在）
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'zervigo_auth_permissions') THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'zervigo_auth_permissions' 
            AND column_name = 'tenant_id'
        ) THEN
            ALTER TABLE zervigo_auth_permissions ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
            CREATE INDEX IF NOT EXISTS idx_permissions_tenant_id ON zervigo_auth_permissions(tenant_id);
            COMMENT ON COLUMN zervigo_auth_permissions.tenant_id IS '租户ID';
        END IF;
    END IF;
END $$;

-- 更新现有数据的tenant_id（确保所有数据都有租户ID）
-- 只更新存在的表和字段
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'zervigo_jobs' AND column_name = 'tenant_id') THEN
        UPDATE zervigo_jobs SET tenant_id = 1 WHERE tenant_id IS NULL OR tenant_id = 0;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'zervigo_user_profiles' AND column_name = 'tenant_id') THEN
        UPDATE zervigo_user_profiles SET tenant_id = 1 WHERE tenant_id IS NULL OR tenant_id = 0;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'zervigo_companies') 
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'zervigo_companies' AND column_name = 'tenant_id') THEN
        UPDATE zervigo_companies SET tenant_id = 1 WHERE tenant_id IS NULL OR tenant_id = 0;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'zervigo_resumes')
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'zervigo_resumes' AND column_name = 'tenant_id') THEN
        UPDATE zervigo_resumes SET tenant_id = 1 WHERE tenant_id IS NULL OR tenant_id = 0;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'zervigo_job_applications' AND column_name = 'tenant_id') THEN
        UPDATE zervigo_job_applications SET tenant_id = 1 WHERE tenant_id IS NULL OR tenant_id = 0;
    END IF;
END $$;

