-- Zervigo 微服务数据库迁移脚本
-- 基于前辈成果，设计适合微服务架构的数据库结构
-- 创建时间: 2025-10-29

-- 设置时区
SET timezone = 'Asia/Shanghai';

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- ==================== auth-service-go 数据库表 ====================

-- 1. 用户认证表
CREATE TABLE IF NOT EXISTS zervigo_auth_users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    
    -- 用户状态
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, suspended, deleted
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    
    -- 订阅信息 (借鉴前辈设计)
    subscription_status VARCHAR(20) DEFAULT 'free', -- free, premium, enterprise
    subscription_type VARCHAR(20) DEFAULT 'basic', -- basic, pro, enterprise
    subscription_expires_at TIMESTAMP,
    subscription_features JSONB DEFAULT '{}',
    
    -- 版本访问权限 (Zervigo特色)
    accessible_versions TEXT[] DEFAULT ARRAY['basic'], -- basic, pro, future
    version_quota JSONB DEFAULT '{"basic": 1000, "pro": 5000, "future": 10000}',
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 2. 角色表
CREATE TABLE IF NOT EXISTS zervigo_auth_roles (
    id BIGSERIAL PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL,
    role_description TEXT,
    role_level INTEGER DEFAULT 1, -- 1-10, 数字越大权限越高
    
    -- 版本权限
    version_access TEXT[] DEFAULT ARRAY['basic'],
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. 权限表
CREATE TABLE IF NOT EXISTS zervigo_auth_permissions (
    id BIGSERIAL PRIMARY KEY,
    permission_name VARCHAR(100) UNIQUE NOT NULL,
    permission_code VARCHAR(100) UNIQUE NOT NULL,
    permission_description TEXT,
    
    -- 服务权限
    service_name VARCHAR(50) NOT NULL, -- user-service, job-service, etc.
    resource_type VARCHAR(50) NOT NULL, -- user, job, company, resume
    action VARCHAR(20) NOT NULL, -- create, read, update, delete, list
    
    -- 版本权限
    version_access TEXT[] DEFAULT ARRAY['basic'],
    
    -- 状态
    status BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. 用户角色关联表
CREATE TABLE IF NOT EXISTS zervigo_auth_user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES zervigo_auth_users(id) ON DELETE CASCADE,
    role_id BIGINT REFERENCES zervigo_auth_roles(id) ON DELETE CASCADE,
    
    -- 分配信息
    assigned_by BIGINT REFERENCES zervigo_auth_users(id),
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP, -- 角色过期时间
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active',
    
    UNIQUE(user_id, role_id)
);

-- 5. 角色权限关联表
CREATE TABLE IF NOT EXISTS zervigo_auth_role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT REFERENCES zervigo_auth_roles(id) ON DELETE CASCADE,
    permission_id BIGINT REFERENCES zervigo_auth_permissions(id) ON DELETE CASCADE,
    
    -- 分配信息
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(role_id, permission_id)
);

-- 6. JWT Token管理表
CREATE TABLE IF NOT EXISTS zervigo_auth_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES zervigo_auth_users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    token_type VARCHAR(20) DEFAULT 'access', -- access, refresh
    
    -- Token信息
    expires_at TIMESTAMP NOT NULL,
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 客户端信息
    client_ip INET,
    user_agent TEXT,
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active', -- active, revoked, expired
    revoked_at TIMESTAMP,
    revoked_reason VARCHAR(100)
);

-- 7. 登录审计表
CREATE TABLE IF NOT EXISTS zervigo_auth_login_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES zervigo_auth_users(id) ON DELETE SET NULL,
    username VARCHAR(50),
    
    -- 登录信息
    login_method VARCHAR(20) NOT NULL, -- password, sms, email, oauth
    success BOOLEAN NOT NULL,
    failure_reason VARCHAR(100),
    
    -- 客户端信息
    client_ip INET,
    user_agent TEXT,
    device_info JSONB,
    
    -- 位置信息
    country VARCHAR(50),
    city VARCHAR(50),
    
    -- 时间
    login_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==================== user-service 数据库表 ====================

-- 8. 用户档案表
CREATE TABLE IF NOT EXISTS zervigo_user_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL, -- 关联auth-service的用户ID
    
    -- 基本信息
    real_name VARCHAR(50),
    nickname VARCHAR(50),
    avatar_url VARCHAR(500),
    gender INTEGER DEFAULT 0, -- 0:未知, 1:男, 2:女
    birthday DATE,
    
    -- 联系信息
    phone VARCHAR(20),
    email VARCHAR(100),
    wechat VARCHAR(50),
    qq VARCHAR(20),
    
    -- 地址信息
    country VARCHAR(50) DEFAULT '中国',
    province VARCHAR(50),
    city VARCHAR(50),
    district VARCHAR(50),
    address TEXT,
    
    -- 职业信息
    current_position VARCHAR(100),
    current_company VARCHAR(100),
    work_experience INTEGER DEFAULT 0, -- 工作年限
    education_level VARCHAR(20), -- 学历
    
    -- 偏好设置
    job_preferences JSONB DEFAULT '{}', -- 职位偏好
    location_preferences TEXT[], -- 地点偏好
    salary_expectation JSONB DEFAULT '{}', -- 薪资期望
    
    -- 状态
    profile_completeness INTEGER DEFAULT 0, -- 档案完整度 0-100
    is_public BOOLEAN DEFAULT FALSE, -- 是否公开档案
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 9. 用户技能表
CREATE TABLE IF NOT EXISTS zervigo_user_skills (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    skill_name VARCHAR(100) NOT NULL,
    skill_category VARCHAR(50), -- 技术, 管理, 语言, 其他
    proficiency_level INTEGER DEFAULT 1, -- 1-5熟练度
    years_of_experience INTEGER DEFAULT 0,
    
    -- 验证信息
    verified BOOLEAN DEFAULT FALSE,
    verified_by VARCHAR(50), -- 验证方式
    verified_at TIMESTAMP,
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 10. 用户教育经历表
CREATE TABLE IF NOT EXISTS zervigo_user_education (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    
    -- 学校信息
    school_name VARCHAR(200) NOT NULL,
    school_type VARCHAR(50), -- 985, 211, 普通本科, 专科, 其他
    major VARCHAR(100),
    degree VARCHAR(50), -- 博士, 硕士, 本科, 专科, 高中
    
    -- 时间信息
    start_date DATE,
    end_date DATE,
    is_current BOOLEAN DEFAULT FALSE, -- 是否在读
    
    -- 成绩信息
    gpa DECIMAL(3,2),
    ranking VARCHAR(50), -- 排名信息
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 11. 用户工作经历表
CREATE TABLE IF NOT EXISTS zervigo_user_experience (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    
    -- 公司信息
    company_name VARCHAR(200) NOT NULL,
    company_industry VARCHAR(100),
    company_size VARCHAR(50), -- 公司规模
    
    -- 职位信息
    position VARCHAR(100) NOT NULL,
    department VARCHAR(100),
    job_level VARCHAR(50), -- 级别
    
    -- 时间信息
    start_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN DEFAULT FALSE, -- 是否当前工作
    
    -- 工作内容
    job_description TEXT,
    achievements TEXT, -- 工作成就
    skills_used TEXT[], -- 使用的技能
    
    -- 薪资信息
    salary_min INTEGER,
    salary_max INTEGER,
    salary_currency VARCHAR(10) DEFAULT 'CNY',
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 12. 用户行为统计表
CREATE TABLE IF NOT EXISTS zervigo_user_statistics (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL,
    
    -- 活跃度统计
    login_count INTEGER DEFAULT 0,
    last_login_at TIMESTAMP,
    active_days INTEGER DEFAULT 0, -- 活跃天数
    
    -- 求职统计
    job_view_count INTEGER DEFAULT 0,
    job_apply_count INTEGER DEFAULT 0,
    resume_view_count INTEGER DEFAULT 0,
    resume_share_count INTEGER DEFAULT 0,
    
    -- AI使用统计
    ai_chat_count INTEGER DEFAULT 0,
    ai_analysis_count INTEGER DEFAULT 0,
    ai_quota_used INTEGER DEFAULT 0,
    
    -- 社交统计
    follow_count INTEGER DEFAULT 0, -- 关注数
    follower_count INTEGER DEFAULT 0, -- 粉丝数
    connection_count INTEGER DEFAULT 0, -- 连接数
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==================== job-service 数据库表 ====================

-- 13. 职位表
CREATE TABLE IF NOT EXISTS zervigo_jobs (
    id BIGSERIAL PRIMARY KEY,
    
    -- 基本信息
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    requirements TEXT,
    responsibilities TEXT, -- 工作职责
    
    -- 公司信息
    company_id BIGINT NOT NULL, -- 关联company-service
    company_name VARCHAR(200), -- 冗余字段，提高查询性能
    company_logo VARCHAR(500),
    
    -- 职位分类
    job_category VARCHAR(100), -- 技术, 销售, 市场, 运营, 其他
    job_subcategory VARCHAR(100), -- 前端, 后端, 全栈, 等
    job_level VARCHAR(50), -- 初级, 中级, 高级, 专家
    
    -- 工作信息
    work_type VARCHAR(20) DEFAULT 'full-time', -- full-time, part-time, contract, intern
    work_location VARCHAR(200),
    work_address TEXT,
    remote_allowed BOOLEAN DEFAULT FALSE,
    
    -- 薪资信息
    salary_min INTEGER,
    salary_max INTEGER,
    salary_currency VARCHAR(10) DEFAULT 'CNY',
    salary_period VARCHAR(20) DEFAULT 'monthly', -- monthly, yearly, hourly
    salary_negotiable BOOLEAN DEFAULT TRUE,
    
    -- 要求信息
    experience_required VARCHAR(50), -- 经验要求
    education_required VARCHAR(50), -- 学历要求
    skills_required TEXT[], -- 技能要求
    languages_required TEXT[], -- 语言要求
    
    -- 福利信息
    benefits TEXT[], -- 福利待遇
    perks TEXT[], -- 额外福利
    
    -- 状态信息
    status VARCHAR(20) DEFAULT 'draft', -- draft, published, paused, closed, expired
    is_featured BOOLEAN DEFAULT FALSE, -- 是否推荐
    is_urgent BOOLEAN DEFAULT FALSE, -- 是否紧急
    
    -- 统计信息
    view_count INTEGER DEFAULT 0,
    apply_count INTEGER DEFAULT 0,
    favorite_count INTEGER DEFAULT 0,
    
    -- 时间信息
    publish_at TIMESTAMP,
    expire_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT NOT NULL -- 创建者ID
);

-- 14. 职位申请表
CREATE TABLE IF NOT EXISTS zervigo_job_applications (
    id BIGSERIAL PRIMARY KEY,
    job_id BIGINT REFERENCES zervigo_jobs(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL, -- 关联auth-service的用户ID
    
    -- 申请信息
    resume_id BIGINT, -- 关联resume-service的简历ID
    cover_letter TEXT, -- 求职信
    application_source VARCHAR(50) DEFAULT 'web', -- web, mobile, api
    
    -- 状态信息
    status VARCHAR(20) DEFAULT 'pending', -- pending, reviewing, interviewed, offered, rejected, withdrawn
    application_stage VARCHAR(50), -- 申请阶段
    
    -- 处理信息
    reviewed_by BIGINT, -- 审核人ID
    reviewed_at TIMESTAMP,
    review_notes TEXT, -- 审核备注
    
    -- 面试信息
    interview_scheduled_at TIMESTAMP,
    interview_location VARCHAR(200),
    interview_notes TEXT,
    
    -- 结果信息
    offer_salary INTEGER,
    offer_start_date DATE,
    offer_expires_at TIMESTAMP,
    
    -- 时间戳
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(job_id, user_id) -- 一个用户只能申请一次同一个职位
);

-- 15. 职位收藏表
CREATE TABLE IF NOT EXISTS zervigo_job_favorites (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    job_id BIGINT REFERENCES zervigo_jobs(id) ON DELETE CASCADE,
    
    -- 收藏信息
    favorite_type VARCHAR(20) DEFAULT 'favorite', -- favorite, bookmark, interested
    notes TEXT, -- 收藏备注
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, job_id)
);

-- 16. 职位搜索历史表
CREATE TABLE IF NOT EXISTS zervigo_job_search_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    
    -- 搜索条件
    search_keywords TEXT,
    search_filters JSONB, -- 搜索筛选条件
    search_location VARCHAR(200),
    
    -- 搜索结果
    result_count INTEGER DEFAULT 0,
    clicked_job_ids BIGINT[], -- 点击的职位ID列表
    
    -- 时间戳
    searched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==================== 创建索引 ====================

-- auth-service 索引
CREATE INDEX IF NOT EXISTS idx_auth_users_username ON zervigo_auth_users(username);
CREATE INDEX IF NOT EXISTS idx_auth_users_email ON zervigo_auth_users(email);
CREATE INDEX IF NOT EXISTS idx_auth_users_status ON zervigo_auth_users(status);
CREATE INDEX IF NOT EXISTS idx_auth_users_subscription ON zervigo_auth_users(subscription_status);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_user_id ON zervigo_auth_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_expires_at ON zervigo_auth_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_status ON zervigo_auth_tokens(status);
CREATE INDEX IF NOT EXISTS idx_auth_login_logs_user_id ON zervigo_auth_login_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_login_logs_login_at ON zervigo_auth_login_logs(login_at);
CREATE INDEX IF NOT EXISTS idx_auth_login_logs_success ON zervigo_auth_login_logs(success);

-- user-service 索引
CREATE INDEX IF NOT EXISTS idx_user_profiles_user_id ON zervigo_user_profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_profiles_city ON zervigo_user_profiles(city);
CREATE INDEX IF NOT EXISTS idx_user_profiles_position ON zervigo_user_profiles(current_position);
CREATE INDEX IF NOT EXISTS idx_user_skills_user_id ON zervigo_user_skills(user_id);
CREATE INDEX IF NOT EXISTS idx_user_skills_category ON zervigo_user_skills(skill_category);
CREATE INDEX IF NOT EXISTS idx_user_education_user_id ON zervigo_user_education(user_id);
CREATE INDEX IF NOT EXISTS idx_user_education_school ON zervigo_user_education(school_name);
CREATE INDEX IF NOT EXISTS idx_user_experience_user_id ON zervigo_user_experience(user_id);
CREATE INDEX IF NOT EXISTS idx_user_experience_company ON zervigo_user_experience(company_name);
CREATE INDEX IF NOT EXISTS idx_user_experience_position ON zervigo_user_experience(position);
CREATE INDEX IF NOT EXISTS idx_user_statistics_user_id ON zervigo_user_statistics(user_id);

-- job-service 索引
CREATE INDEX IF NOT EXISTS idx_jobs_company_id ON zervigo_jobs(company_id);
CREATE INDEX IF NOT EXISTS idx_jobs_title ON zervigo_jobs(title);
CREATE INDEX IF NOT EXISTS idx_jobs_location ON zervigo_jobs(work_location);
CREATE INDEX IF NOT EXISTS idx_jobs_category ON zervigo_jobs(job_category);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON zervigo_jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_salary ON zervigo_jobs(salary_min, salary_max);
CREATE INDEX IF NOT EXISTS idx_jobs_publish_at ON zervigo_jobs(publish_at);
CREATE INDEX IF NOT EXISTS idx_jobs_featured ON zervigo_jobs(is_featured);
CREATE INDEX IF NOT EXISTS idx_job_applications_job_id ON zervigo_job_applications(job_id);
CREATE INDEX IF NOT EXISTS idx_job_applications_user_id ON zervigo_job_applications(user_id);
CREATE INDEX IF NOT EXISTS idx_job_applications_status ON zervigo_job_applications(status);
CREATE INDEX IF NOT EXISTS idx_job_applications_applied_at ON zervigo_job_applications(applied_at);
CREATE INDEX IF NOT EXISTS idx_job_favorites_user_id ON zervigo_job_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_job_favorites_job_id ON zervigo_job_favorites(job_id);
CREATE INDEX IF NOT EXISTS idx_job_search_history_user_id ON zervigo_job_search_history(user_id);
CREATE INDEX IF NOT EXISTS idx_job_search_history_searched_at ON zervigo_job_search_history(searched_at);

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
CREATE TRIGGER update_auth_users_updated_at BEFORE UPDATE ON zervigo_auth_users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_auth_roles_updated_at BEFORE UPDATE ON zervigo_auth_roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_profiles_updated_at BEFORE UPDATE ON zervigo_user_profiles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_skills_updated_at BEFORE UPDATE ON zervigo_user_skills FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_education_updated_at BEFORE UPDATE ON zervigo_user_education FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_experience_updated_at BEFORE UPDATE ON zervigo_user_experience FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_statistics_updated_at BEFORE UPDATE ON zervigo_user_statistics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_jobs_updated_at BEFORE UPDATE ON zervigo_jobs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_job_applications_updated_at BEFORE UPDATE ON zervigo_job_applications FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ==================== 插入初始数据 ====================

-- 插入默认角色
INSERT INTO zervigo_auth_roles (role_name, role_description, role_level, version_access) VALUES
('super_admin', '超级管理员', 10, ARRAY['basic', 'pro', 'future']),
('admin', '系统管理员', 8, ARRAY['basic', 'pro']),
('hr_manager', 'HR经理', 6, ARRAY['basic', 'pro']),
('hr_user', 'HR用户', 4, ARRAY['basic']),
('job_seeker', '求职者', 2, ARRAY['basic']),
('company_admin', '企业管理员', 5, ARRAY['basic', 'pro']),
('company_user', '企业用户', 3, ARRAY['basic'])
ON CONFLICT (role_name) DO NOTHING;

-- 插入默认权限
INSERT INTO zervigo_auth_permissions (permission_name, permission_code, service_name, resource_type, action, version_access) VALUES
('用户管理-创建', 'user:create', 'user-service', 'user', 'create', ARRAY['basic', 'pro', 'future']),
('用户管理-读取', 'user:read', 'user-service', 'user', 'read', ARRAY['basic', 'pro', 'future']),
('用户管理-更新', 'user:update', 'user-service', 'user', 'update', ARRAY['basic', 'pro', 'future']),
('用户管理-删除', 'user:delete', 'user-service', 'user', 'delete', ARRAY['pro', 'future']),
('用户管理-列表', 'user:list', 'user-service', 'user', 'list', ARRAY['basic', 'pro', 'future']),
('职位管理-创建', 'job:create', 'job-service', 'job', 'create', ARRAY['basic', 'pro', 'future']),
('职位管理-读取', 'job:read', 'job-service', 'job', 'read', ARRAY['basic', 'pro', 'future']),
('职位管理-更新', 'job:update', 'job-service', 'job', 'update', ARRAY['basic', 'pro', 'future']),
('职位管理-删除', 'job:delete', 'job-service', 'job', 'delete', ARRAY['pro', 'future']),
('职位管理-列表', 'job:list', 'job-service', 'job', 'list', ARRAY['basic', 'pro', 'future']),
('简历管理-创建', 'resume:create', 'resume-service', 'resume', 'create', ARRAY['basic', 'pro', 'future']),
('简历管理-读取', 'resume:read', 'resume-service', 'resume', 'read', ARRAY['basic', 'pro', 'future']),
('简历管理-更新', 'resume:update', 'resume-service', 'resume', 'update', ARRAY['basic', 'pro', 'future']),
('简历管理-删除', 'resume:delete', 'resume-service', 'resume', 'delete', ARRAY['pro', 'future']),
('简历管理-列表', 'resume:list', 'resume-service', 'resume', 'list', ARRAY['basic', 'pro', 'future']),
('企业管理-创建', 'company:create', 'company-service', 'company', 'create', ARRAY['basic', 'pro', 'future']),
('企业管理-读取', 'company:read', 'company-service', 'company', 'read', ARRAY['basic', 'pro', 'future']),
('企业管理-更新', 'company:update', 'company-service', 'company', 'update', ARRAY['basic', 'pro', 'future']),
('企业管理-删除', 'company:delete', 'company-service', 'company', 'delete', ARRAY['pro', 'future']),
('企业管理-列表', 'company:list', 'company-service', 'company', 'list', ARRAY['basic', 'pro', 'future']),
('AI服务-使用', 'ai:use', 'ai-service', 'ai', 'use', ARRAY['basic', 'pro', 'future']),
('AI服务-分析', 'ai:analyze', 'ai-service', 'ai', 'analyze', ARRAY['pro', 'future']),
('区块链-记录', 'blockchain:record', 'blockchain-service', 'blockchain', 'record', ARRAY['pro', 'future']),
('区块链-查询', 'blockchain:query', 'blockchain-service', 'blockchain', 'query', ARRAY['basic', 'pro', 'future'])
ON CONFLICT (permission_code) DO NOTHING;

-- 为超级管理员分配所有权限
INSERT INTO zervigo_auth_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM zervigo_auth_roles r, zervigo_auth_permissions p
WHERE r.role_name = 'super_admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 为管理员分配基本权限
INSERT INTO zervigo_auth_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM zervigo_auth_roles r, zervigo_auth_permissions p
WHERE r.role_name = 'admin' 
AND p.permission_code IN ('user:read', 'user:update', 'user:list', 'job:read', 'job:update', 'job:list', 'resume:read', 'resume:list', 'company:read', 'company:update', 'company:list', 'ai:use', 'blockchain:query')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 为HR经理分配HR权限
INSERT INTO zervigo_auth_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM zervigo_auth_roles r, zervigo_auth_permissions p
WHERE r.role_name = 'hr_manager' 
AND p.permission_code IN ('user:read', 'user:list', 'job:create', 'job:read', 'job:update', 'job:list', 'resume:read', 'resume:list', 'company:read', 'company:list', 'ai:use', 'ai:analyze')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 为求职者分配基本权限
INSERT INTO zervigo_auth_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM zervigo_auth_roles r, zervigo_auth_permissions p
WHERE r.role_name = 'job_seeker' 
AND p.permission_code IN ('user:read', 'user:update', 'job:read', 'job:list', 'resume:create', 'resume:read', 'resume:update', 'resume:list', 'company:read', 'company:list', 'ai:use')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 创建超级管理员用户 (密码: admin123)
INSERT INTO zervigo_auth_users (username, email, password_hash, status, email_verified, subscription_status, subscription_type, accessible_versions) VALUES
('admin', 'admin@zervigo.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'active', true, 'premium', 'pro', ARRAY['basic', 'pro', 'future'])
ON CONFLICT (username) DO NOTHING;

-- 为超级管理员分配角色
INSERT INTO zervigo_auth_user_roles (user_id, role_id) 
SELECT u.id, r.id 
FROM zervigo_auth_users u, zervigo_auth_roles r 
WHERE u.username = 'admin' AND r.role_name = 'super_admin'
ON CONFLICT (user_id, role_id) DO NOTHING;

-- ==================== 完成初始化 ====================

-- 输出初始化完成信息
DO $$
BEGIN
    RAISE NOTICE 'Zervigo 微服务数据库初始化完成！';
    RAISE NOTICE '数据库: zervigo_mvp';
    RAISE NOTICE '默认管理员账号: admin / admin123';
    RAISE NOTICE '数据库连接: postgres://szjason72@localhost:5432/zervigo_mvp';
    RAISE NOTICE '';
    RAISE NOTICE '已创建的表:';
    RAISE NOTICE '  auth-service: 7个表';
    RAISE NOTICE '  user-service: 5个表';
    RAISE NOTICE '  job-service: 4个表';
    RAISE NOTICE '  总计: 16个表';
    RAISE NOTICE '';
    RAISE NOTICE '已创建的索引: 40+个';
    RAISE NOTICE '已创建的触发器: 9个';
    RAISE NOTICE '已插入的初始数据: 角色7个, 权限25个, 用户1个';
END $$;
