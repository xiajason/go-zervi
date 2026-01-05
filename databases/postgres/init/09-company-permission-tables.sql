-- Company服务权限管理表 - PostgreSQL版本
-- 创建企业用户关联表、权限审计日志表、数据同步状态表
-- 并在companies表中添加权限相关字段

-- 1. 在companies表中添加权限相关字段
DO $$ 
BEGIN
    -- 添加创建者字段
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'companies' AND column_name = 'created_by') THEN
        ALTER TABLE companies ADD COLUMN created_by BIGINT;
        COMMENT ON COLUMN companies.created_by IS '创建者用户ID';
    END IF;

    -- 添加法定代表人用户ID字段
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'companies' AND column_name = 'legal_rep_user_id') THEN
        ALTER TABLE companies ADD COLUMN legal_rep_user_id BIGINT;
        COMMENT ON COLUMN companies.legal_rep_user_id IS '法定代表人用户ID';
    END IF;

    -- 添加授权用户列表字段（JSON格式）
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'companies' AND column_name = 'authorized_users') THEN
        ALTER TABLE companies ADD COLUMN authorized_users JSONB;
        COMMENT ON COLUMN companies.authorized_users IS '授权用户列表（JSON数组）';
    END IF;

    -- 添加统一社会信用代码字段（如果不存在）
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'companies' AND column_name = 'unified_social_credit_code') THEN
        ALTER TABLE companies ADD COLUMN unified_social_credit_code VARCHAR(50);
        CREATE UNIQUE INDEX IF NOT EXISTS idx_companies_unified_social_credit_code ON companies(unified_social_credit_code) WHERE unified_social_credit_code IS NOT NULL;
        COMMENT ON COLUMN companies.unified_social_credit_code IS '统一社会信用代码';
    END IF;

    -- 添加法定代表人身份证号字段（如果legal_person_id不存在）
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'companies' AND column_name = 'legal_representative_id') THEN
        ALTER TABLE companies ADD COLUMN legal_representative_id VARCHAR(50);
        COMMENT ON COLUMN companies.legal_representative_id IS '法定代表人身份证号';
    END IF;
END $$;

-- 2. 创建企业用户关联表
CREATE TABLE IF NOT EXISTS company_users (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(50) NOT NULL, -- legal_rep, authorized_user, admin
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, pending
    permissions JSONB, -- 权限列表
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES zervigo_auth_users(id) ON DELETE CASCADE,
    UNIQUE (company_id, user_id)
);

COMMENT ON TABLE company_users IS '企业用户关联表 - 支持多用户管理和权限控制';
COMMENT ON COLUMN company_users.role IS '角色：legal_rep(法定代表人), authorized_user(授权用户), admin(管理员)';
COMMENT ON COLUMN company_users.status IS '状态：active(活跃), inactive(非活跃), pending(待审核)';
COMMENT ON COLUMN company_users.permissions IS '权限列表（JSON数组）';

-- 3. 创建企业权限审计日志表
CREATE TABLE IF NOT EXISTS company_permission_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    action VARCHAR(100) NOT NULL, -- 操作类型
    resource_type VARCHAR(50) NOT NULL, -- 资源类型
    resource_id BIGINT, -- 资源ID
    permission_result BOOLEAN NOT NULL DEFAULT FALSE, -- 权限检查结果
    ip_address VARCHAR(45), -- IP地址
    user_agent TEXT, -- 用户代理
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES zervigo_auth_users(id) ON DELETE CASCADE
);

COMMENT ON TABLE company_permission_audit_logs IS '企业权限审计日志表 - 记录所有权限检查操作';
COMMENT ON COLUMN company_permission_audit_logs.action IS '操作类型：view_company, edit_company, delete_company等';
COMMENT ON COLUMN company_permission_audit_logs.resource_type IS '资源类型：company, job, document等';
COMMENT ON COLUMN company_permission_audit_logs.permission_result IS '权限检查结果：true(允许), false(拒绝)';

-- 4. 创建企业数据同步状态表
CREATE TABLE IF NOT EXISTS company_data_sync_status (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL,
    sync_target VARCHAR(50) NOT NULL, -- postgresql, neo4j, redis
    sync_status VARCHAR(20) DEFAULT 'pending', -- pending, syncing, success, failed
    last_sync_time TIMESTAMP, -- 最后同步时间
    sync_error TEXT, -- 同步错误信息
    retry_count INT DEFAULT 0, -- 重试次数
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE,
    UNIQUE (company_id, sync_target)
);

COMMENT ON TABLE company_data_sync_status IS '企业数据同步状态表 - 跟踪多数据库同步状态';
COMMENT ON COLUMN company_data_sync_status.sync_target IS '同步目标：postgresql, neo4j, redis';
COMMENT ON COLUMN company_data_sync_status.sync_status IS '同步状态：pending(待同步), syncing(同步中), success(成功), failed(失败)';

-- 5. 创建索引优化查询性能
CREATE INDEX IF NOT EXISTS idx_company_users_company_id ON company_users(company_id);
CREATE INDEX IF NOT EXISTS idx_company_users_user_id ON company_users(user_id);
CREATE INDEX IF NOT EXISTS idx_company_users_role ON company_users(role);
CREATE INDEX IF NOT EXISTS idx_company_users_status ON company_users(status);
CREATE INDEX IF NOT EXISTS idx_company_users_company_user ON company_users(company_id, user_id, status);
CREATE INDEX IF NOT EXISTS idx_company_users_user_company ON company_users(user_id, company_id, role);

CREATE INDEX IF NOT EXISTS idx_company_audit_company_id ON company_permission_audit_logs(company_id);
CREATE INDEX IF NOT EXISTS idx_company_audit_user_id ON company_permission_audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_company_audit_action ON company_permission_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_company_audit_created_at ON company_permission_audit_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_company_sync_company_id ON company_data_sync_status(company_id);
CREATE INDEX IF NOT EXISTS idx_company_sync_target ON company_data_sync_status(sync_target);
CREATE INDEX IF NOT EXISTS idx_company_sync_status ON company_data_sync_status(sync_status);

-- 6. 创建更新时间触发器函数（如果不存在）
CREATE OR REPLACE FUNCTION update_company_users_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_company_sync_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 7. 创建触发器
DROP TRIGGER IF EXISTS trigger_update_company_users_updated_at ON company_users;
CREATE TRIGGER trigger_update_company_users_updated_at
    BEFORE UPDATE ON company_users
    FOR EACH ROW
    EXECUTE FUNCTION update_company_users_updated_at();

DROP TRIGGER IF EXISTS trigger_update_company_sync_updated_at ON company_data_sync_status;
CREATE TRIGGER trigger_update_company_sync_updated_at
    BEFORE UPDATE ON company_data_sync_status
    FOR EACH ROW
    EXECUTE FUNCTION update_company_sync_updated_at();

-- 8. 验证表创建
DO $$
BEGIN
    RAISE NOTICE 'company_users表创建完成';
    RAISE NOTICE 'company_permission_audit_logs表创建完成';
    RAISE NOTICE 'company_data_sync_status表创建完成';
    RAISE NOTICE 'companies表权限字段添加完成';
END $$;
