-- Zervigo MVP PostgreSQL 数据库初始化脚本
-- 创建所有微服务需要的数据库表

-- 设置时区
SET timezone = 'Asia/Shanghai';

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- ==================== 用户相关表 ====================

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    user_id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    real_name VARCHAR(50),
    avatar VARCHAR(500),
    gender INTEGER DEFAULT 0,
    birthday DATE,
    location VARCHAR(100),
    bio TEXT,
    status INTEGER DEFAULT 1,
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 用户角色表
CREATE TABLE IF NOT EXISTS user_roles (
    role_id BIGSERIAL PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL,
    role_description TEXT,
    permissions JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS user_role_assignments (
    assignment_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    role_id BIGINT REFERENCES user_roles(role_id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by BIGINT REFERENCES users(user_id),
    UNIQUE(user_id, role_id)
);

-- ==================== 认证相关表 ====================

-- JWT Token黑名单表
CREATE TABLE IF NOT EXISTS token_blacklist (
    token_id BIGSERIAL PRIMARY KEY,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 登录日志表
CREATE TABLE IF NOT EXISTS login_logs (
    log_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    login_ip INET,
    user_agent TEXT,
    login_method VARCHAR(20),
    success BOOLEAN NOT NULL,
    failure_reason VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==================== 企业相关表 ====================

-- 企业表
CREATE TABLE IF NOT EXISTS companies (
    company_id BIGSERIAL PRIMARY KEY,
    company_name VARCHAR(100) NOT NULL,
    company_logo VARCHAR(500),
    company_description TEXT,
    industry VARCHAR(50),
    company_size VARCHAR(20),
    website VARCHAR(200),
    address TEXT,
    city VARCHAR(50),
    province VARCHAR(50),
    country VARCHAR(50) DEFAULT '中国',
    contact_person VARCHAR(50),
    contact_phone VARCHAR(20),
    contact_email VARCHAR(100),
    status INTEGER DEFAULT 1,
    verification_status INTEGER DEFAULT 0,
    business_license VARCHAR(500),
    tax_number VARCHAR(50),
    legal_person VARCHAR(50),
    legal_person_id VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 企业认证记录表
CREATE TABLE IF NOT EXISTS company_verifications (
    verification_id BIGSERIAL PRIMARY KEY,
    company_id BIGINT REFERENCES companies(company_id) ON DELETE CASCADE,
    verification_type VARCHAR(20) NOT NULL,
    verification_data JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    verified_by BIGINT REFERENCES users(user_id),
    verified_at TIMESTAMP,
    rejection_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==================== 职位相关表 ====================

-- 职位表
CREATE TABLE IF NOT EXISTS jobs (
    job_id BIGSERIAL PRIMARY KEY,
    company_id BIGINT REFERENCES companies(company_id) ON DELETE CASCADE,
    job_title VARCHAR(100) NOT NULL,
    job_description TEXT NOT NULL,
    job_requirements TEXT,
    job_type VARCHAR(20) DEFAULT 'full-time',
    work_location VARCHAR(100),
    salary_min INTEGER,
    salary_max INTEGER,
    salary_currency VARCHAR(10) DEFAULT 'CNY',
    experience VARCHAR(20),
    education VARCHAR(20),
    skills TEXT[],
    benefits TEXT[],
    status INTEGER DEFAULT 1,
    is_featured BOOLEAN DEFAULT FALSE,
    view_count INTEGER DEFAULT 0,
    apply_count INTEGER DEFAULT 0,
    created_by BIGINT REFERENCES users(user_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 职位申请表
CREATE TABLE IF NOT EXISTS job_applications (
    application_id BIGSERIAL PRIMARY KEY,
    job_id BIGINT REFERENCES jobs(job_id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    resume_id BIGINT,
    cover_letter TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reviewed_at TIMESTAMP,
    reviewed_by BIGINT REFERENCES users(user_id),
    review_notes TEXT,
    UNIQUE(job_id, user_id)
);

-- ==================== 简历相关表 ====================

-- 简历表
CREATE TABLE IF NOT EXISTS resumes (
    resume_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    resume_name VARCHAR(100) NOT NULL,
    personal_info JSONB NOT NULL,
    work_experience JSONB DEFAULT '[]',
    education JSONB DEFAULT '[]',
    skills JSONB DEFAULT '[]',
    projects JSONB DEFAULT '[]',
    certificates JSONB DEFAULT '[]',
    status INTEGER DEFAULT 1,
    is_public BOOLEAN DEFAULT FALSE,
    view_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 简历文件表
CREATE TABLE IF NOT EXISTS resume_files (
    file_id BIGSERIAL PRIMARY KEY,
    resume_id BIGINT REFERENCES resumes(resume_id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT,
    file_type VARCHAR(50),
    upload_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==================== AI相关表 ====================

-- AI匹配记录表
CREATE TABLE IF NOT EXISTS ai_matches (
    match_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    resume_id BIGINT REFERENCES resumes(resume_id) ON DELETE CASCADE,
    job_id BIGINT REFERENCES jobs(job_id) ON DELETE CASCADE,
    match_type VARCHAR(20) NOT NULL,
    match_score DECIMAL(5,2) NOT NULL,
    match_details JSONB NOT NULL,
    recommendations JSONB DEFAULT '[]',
    analysis_result JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- AI聊天记录表
CREATE TABLE IF NOT EXISTS ai_chats (
    chat_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    session_id VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    response TEXT NOT NULL,
    chat_type VARCHAR(20) DEFAULT 'general',
    context JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- AI分析记录表
CREATE TABLE IF NOT EXISTS ai_analyses (
    analysis_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    resume_id BIGINT REFERENCES resumes(resume_id) ON DELETE CASCADE,
    job_id BIGINT REFERENCES jobs(job_id) ON DELETE SET NULL,
    analysis_type VARCHAR(20) NOT NULL,
    analysis_result JSONB NOT NULL,
    skills_analysis JSONB,
    experience_analysis JSONB,
    education_analysis JSONB,
    overall_score DECIMAL(5,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==================== 区块链相关表 ====================

-- 区块链交易表
CREATE TABLE IF NOT EXISTS blockchain_transactions (
    transaction_id VARCHAR(100) PRIMARY KEY,
    transaction_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(100) NOT NULL,
    version_source VARCHAR(20) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    change_reason TEXT,
    operator_id VARCHAR(100),
    transaction_hash VARCHAR(255) UNIQUE NOT NULL,
    transaction_data JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    block_height BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    confirmed_at TIMESTAMP
);

-- 版本状态记录表
CREATE TABLE IF NOT EXISTS version_status_records (
    record_id VARCHAR(100) PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    version_source VARCHAR(20) NOT NULL,
    old_status VARCHAR(20),
    new_status VARCHAR(20) NOT NULL,
    change_reason TEXT,
    operator_id VARCHAR(100),
    transaction_hash VARCHAR(255) REFERENCES blockchain_transactions(transaction_hash),
    block_height BIGINT,
    record_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 权限变更记录表
CREATE TABLE IF NOT EXISTS permission_change_records (
    record_id VARCHAR(100) PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    version_source VARCHAR(20) NOT NULL,
    old_permission VARCHAR(50),
    new_permission VARCHAR(50) NOT NULL,
    change_reason TEXT,
    operator_id VARCHAR(100),
    transaction_hash VARCHAR(255) REFERENCES blockchain_transactions(transaction_hash),
    block_height BIGINT,
    record_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==================== 系统相关表 ====================

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    config_id BIGSERIAL PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    config_type VARCHAR(20) DEFAULT 'string',
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 操作日志表
CREATE TABLE IF NOT EXISTS operation_logs (
    log_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id) ON DELETE SET NULL,
    operation_type VARCHAR(50) NOT NULL,
    operation_target VARCHAR(100),
    operation_data JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==================== 创建索引 ====================

-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- 企业表索引
CREATE INDEX IF NOT EXISTS idx_companies_name ON companies(company_name);
CREATE INDEX IF NOT EXISTS idx_companies_industry ON companies(industry);
CREATE INDEX IF NOT EXISTS idx_companies_city ON companies(city);
CREATE INDEX IF NOT EXISTS idx_companies_status ON companies(status);
CREATE INDEX IF NOT EXISTS idx_companies_verification_status ON companies(verification_status);

-- 职位表索引
CREATE INDEX IF NOT EXISTS idx_jobs_company_id ON jobs(company_id);
CREATE INDEX IF NOT EXISTS idx_jobs_title ON jobs(job_title);
CREATE INDEX IF NOT EXISTS idx_jobs_location ON jobs(work_location);
CREATE INDEX IF NOT EXISTS idx_jobs_type ON jobs(job_type);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_featured ON jobs(is_featured);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at);

-- 简历表索引
CREATE INDEX IF NOT EXISTS idx_resumes_user_id ON resumes(user_id);
CREATE INDEX IF NOT EXISTS idx_resumes_status ON resumes(status);
CREATE INDEX IF NOT EXISTS idx_resumes_public ON resumes(is_public);
CREATE INDEX IF NOT EXISTS idx_resumes_created_at ON resumes(created_at);

-- AI相关表索引
CREATE INDEX IF NOT EXISTS idx_ai_matches_user_id ON ai_matches(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_matches_resume_id ON ai_matches(resume_id);
CREATE INDEX IF NOT EXISTS idx_ai_matches_job_id ON ai_matches(job_id);
CREATE INDEX IF NOT EXISTS idx_ai_matches_type ON ai_matches(match_type);
CREATE INDEX IF NOT EXISTS idx_ai_matches_score ON ai_matches(match_score);
CREATE INDEX IF NOT EXISTS idx_ai_matches_created_at ON ai_matches(created_at);

CREATE INDEX IF NOT EXISTS idx_ai_chats_user_id ON ai_chats(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_chats_session_id ON ai_chats(session_id);
CREATE INDEX IF NOT EXISTS idx_ai_chats_type ON ai_chats(chat_type);
CREATE INDEX IF NOT EXISTS idx_ai_chats_created_at ON ai_chats(created_at);

-- 区块链表索引
CREATE INDEX IF NOT EXISTS idx_blockchain_transactions_type ON blockchain_transactions(transaction_type);
CREATE INDEX IF NOT EXISTS idx_blockchain_transactions_entity_id ON blockchain_transactions(entity_id);
CREATE INDEX IF NOT EXISTS idx_blockchain_transactions_status ON blockchain_transactions(status);
CREATE INDEX IF NOT EXISTS idx_blockchain_transactions_block_height ON blockchain_transactions(block_height);
CREATE INDEX IF NOT EXISTS idx_blockchain_transactions_created_at ON blockchain_transactions(created_at);

-- ==================== 创建触发器 ====================

-- 更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要的表添加更新时间触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_companies_updated_at BEFORE UPDATE ON companies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_jobs_updated_at BEFORE UPDATE ON jobs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_resumes_updated_at BEFORE UPDATE ON resumes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_system_configs_updated_at BEFORE UPDATE ON system_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ==================== 插入初始数据 ====================

-- 插入默认用户角色
INSERT INTO user_roles (role_name, role_description, permissions) VALUES
('admin', '系统管理员', '["user:read", "user:write", "user:delete", "company:read", "company:write", "company:delete", "job:read", "job:write", "job:delete", "resume:read", "resume:write", "resume:delete", "ai:read", "ai:write", "blockchain:read", "blockchain:write"]'),
('hr', 'HR用户', '["user:read", "company:read", "company:write", "job:read", "job:write", "resume:read", "ai:read"]'),
('job_seeker', '求职者', '["user:read", "user:write", "company:read", "job:read", "resume:read", "resume:write", "ai:read"]'),
('company_admin', '企业管理员', '["user:read", "company:read", "company:write", "job:read", "job:write", "resume:read", "ai:read"]')
ON CONFLICT (role_name) DO NOTHING;

-- 插入系统配置
INSERT INTO system_configs (config_key, config_value, config_type, description) VALUES
('system_name', 'Zervigo MVP', 'string', '系统名称'),
('system_version', '1.0.0', 'string', '系统版本'),
('ai_enabled', 'true', 'boolean', 'AI功能是否启用'),
('blockchain_enabled', 'true', 'boolean', '区块链功能是否启用'),
('max_resume_size', '10485760', 'number', '简历文件最大大小(字节)'),
('max_job_applications', '100', 'number', '单个职位最大申请数'),
('default_page_size', '20', 'number', '默认分页大小')
ON CONFLICT (config_key) DO NOTHING;

-- 创建超级管理员用户 (密码: admin123)
INSERT INTO users (username, email, password_hash, real_name, status, email_verified) VALUES
('admin', 'admin@zervigo.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '系统管理员', 1, true)
ON CONFLICT (username) DO NOTHING;

-- 为超级管理员分配角色
INSERT INTO user_role_assignments (user_id, role_id) 
SELECT u.user_id, r.role_id 
FROM users u, user_roles r 
WHERE u.username = 'admin' AND r.role_name = 'admin'
ON CONFLICT (user_id, role_id) DO NOTHING;

-- ==================== 完成初始化 ====================

-- 输出初始化完成信息
DO $$
BEGIN
    RAISE NOTICE 'Zervigo MVP PostgreSQL 数据库初始化完成！';
    RAISE NOTICE '数据库: zervigo_mvp';
    RAISE NOTICE '默认管理员账号: admin / admin123';
    RAISE NOTICE '数据库连接: postgres://postgres:dev_password@localhost:5432/zervigo_mvp';
END $$;
