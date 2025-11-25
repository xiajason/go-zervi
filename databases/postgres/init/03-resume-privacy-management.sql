-- 简历权限管理相关表结构
-- 基于学习文档的业务需求设计

-- 简历权限表
CREATE TABLE IF NOT EXISTS resume_permission (
    id BIGSERIAL PRIMARY KEY,
    resume_id VARCHAR(50) UNIQUE NOT NULL,
    privacy_level VARCHAR(20) DEFAULT 'PRIVATE' CHECK (privacy_level IN ('PUBLIC', 'PRIVATE', 'FRIENDS')),
    allow_download BOOLEAN DEFAULT false,
    require_approval BOOLEAN DEFAULT true,
    allowed_enterprises TEXT, -- JSON数组存储允许的企业ID
    denied_enterprises TEXT, -- JSON数组存储禁止的企业ID
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 简历黑名单表
CREATE TABLE IF NOT EXISTS resume_blacklist (
    id BIGSERIAL PRIMARY KEY,
    resume_id VARCHAR(50) NOT NULL,
    enterprise_id VARCHAR(50) NOT NULL,
    enterprise_name VARCHAR(100) NOT NULL,
    reason VARCHAR(255),
    add_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(resume_id, enterprise_id)
);

-- 审批记录表
CREATE TABLE IF NOT EXISTS approve_record (
    approve_id VARCHAR(50) PRIMARY KEY,
    type VARCHAR(50) NOT NULL CHECK (type IN ('简历查看', '简历下载', '简历收藏')),
    user_id BIGINT NOT NULL,
    enterprise_id VARCHAR(50) NOT NULL,
    enterprise_name VARCHAR(100) NOT NULL,
    resume_id VARCHAR(50) NOT NULL,
    resume_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT '待审批' CHECK (status IN ('待审批', '已通过', '已拒绝')),
    cost INT DEFAULT 0, -- 积分消耗
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    handle_time TIMESTAMP
);

-- 积分账单表
CREATE TABLE IF NOT EXISTS points_bill (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('收入', '支出')),
    amount INT NOT NULL,
    reason VARCHAR(255) NOT NULL,
    balance INT NOT NULL, -- 余额
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 查看历史表
CREATE TABLE IF NOT EXISTS view_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL, -- 查看者ID
    resume_id VARCHAR(50) NOT NULL,
    resume_name VARCHAR(100) NOT NULL,
    resume_owner VARCHAR(50) NOT NULL, -- 简历所有者
    view_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    cost INT DEFAULT 0 -- 消耗积分
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_resume_permission_resume_id ON resume_permission(resume_id);
CREATE INDEX IF NOT EXISTS idx_resume_blacklist_resume_id ON resume_blacklist(resume_id);
CREATE INDEX IF NOT EXISTS idx_resume_blacklist_enterprise_id ON resume_blacklist(enterprise_id);
CREATE INDEX IF NOT EXISTS idx_approve_record_user_id ON approve_record(user_id);
CREATE INDEX IF NOT EXISTS idx_approve_record_status ON approve_record(status);
CREATE INDEX IF NOT EXISTS idx_approve_record_create_time ON approve_record(create_time);
CREATE INDEX IF NOT EXISTS idx_points_bill_user_id ON points_bill(user_id);
CREATE INDEX IF NOT EXISTS idx_points_bill_create_time ON points_bill(create_time);
CREATE INDEX IF NOT EXISTS idx_view_history_user_id ON view_history(user_id);
CREATE INDEX IF NOT EXISTS idx_view_history_resume_id ON view_history(resume_id);

-- 插入示例数据
INSERT INTO resume_permission (resume_id, privacy_level, allow_download, require_approval, allowed_enterprises, denied_enterprises) 
VALUES 
('resume_001', 'PUBLIC', true, false, '["ent_001", "ent_002"]', '["ent_003"]'),
('resume_002', 'PRIVATE', false, true, '["ent_001"]', '[]'),
('resume_003', 'FRIENDS', true, true, '[]', '["ent_004"]')
ON CONFLICT (resume_id) DO NOTHING;

INSERT INTO resume_blacklist (resume_id, enterprise_id, enterprise_name, reason) 
VALUES 
('resume_001', 'ent_003', '不良企业A', '恶意下载简历'),
('resume_002', 'ent_004', '不良企业B', '骚扰用户')
ON CONFLICT (resume_id, enterprise_id) DO NOTHING;

INSERT INTO approve_record (approve_id, type, user_id, enterprise_id, enterprise_name, resume_id, resume_name, status, cost) 
VALUES 
('approve_001', '简历查看', 1, 'ent_001', '优质企业A', 'resume_001', '张三的简历', '待审批', 10),
('approve_002', '简历下载', 1, 'ent_002', '优质企业B', 'resume_002', '李四的简历', '已通过', 20)
ON CONFLICT (approve_id) DO NOTHING;

INSERT INTO points_bill (user_id, type, amount, reason, balance) 
VALUES 
(1, '收入', 100, '注册奖励', 100),
(1, '支出', 10, '查看简历', 90),
(1, '支出', 20, '下载简历', 70)
ON CONFLICT DO NOTHING;
