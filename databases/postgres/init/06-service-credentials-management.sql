-- 服务凭证管理表
-- 用于管理微服务集群内部的认证凭证

CREATE TABLE IF NOT EXISTS zervigo_service_credentials (
    id SERIAL PRIMARY KEY,
    service_id VARCHAR(100) NOT NULL UNIQUE,
    service_name VARCHAR(200) NOT NULL,
    service_type VARCHAR(50) NOT NULL, -- core/infrastructure/business
    service_secret VARCHAR(255) NOT NULL,
    description TEXT,
    allowed_apis TEXT[], -- 允许调用的API列表，空数组表示全部允许
    status VARCHAR(20) DEFAULT 'active', -- active/inactive/revoked
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    last_used_at TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_service_credentials_service_id ON zervigo_service_credentials(service_id);
CREATE INDEX IF NOT EXISTS idx_service_credentials_status ON zervigo_service_credentials(status);

-- 服务token记录表（可选，用于审计）
CREATE TABLE IF NOT EXISTS zervigo_service_tokens (
    id SERIAL PRIMARY KEY,
    service_id VARCHAR(100) NOT NULL,
    token_hash VARCHAR(255) NOT NULL, -- token的hash值，不存储明文
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    last_used_at TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES zervigo_service_credentials(service_id)
);

CREATE INDEX IF NOT EXISTS idx_service_tokens_service_id ON zervigo_service_tokens(service_id);
CREATE INDEX IF NOT EXISTS idx_service_tokens_token_hash ON zervigo_service_tokens(token_hash);

-- 插入默认的核心基础设施服务凭证
INSERT INTO zervigo_service_credentials (service_id, service_name, service_type, service_secret, description, allowed_apis, created_by)
VALUES 
    ('central-brain', 'Central Brain (API Gateway)', 'infrastructure', 
     '$2a$10$central.brain.secret.key.generated.at.runtime', 
     'API网关服务，负责路由转发和请求代理', 
     ARRAY['*'], 
     'system'),
    ('auth-service', 'Auth Service', 'core', 
     '$2a$10$auth.service.secret.key.generated.at.runtime', 
     '认证服务，负责用户认证和服务认证', 
     ARRAY['*'], 
     'system'),
    ('permission-service', 'Permission Service', 'infrastructure', 
     '$2a$10$permission.service.secret.key.generated.at.runtime', 
     '权限服务，负责权限管理', 
     ARRAY['*'], 
     'system'),
    ('router-service', 'Router Service', 'infrastructure', 
     '$2a$10$router.service.secret.key.generated.at.runtime', 
     '路由服务，负责动态路由管理', 
     ARRAY['*'], 
     'system')
ON CONFLICT (service_id) DO NOTHING;

-- 创建服务凭证更新时间的触发器
CREATE OR REPLACE FUNCTION update_service_credentials_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_service_credentials_updated_at ON zervigo_service_credentials;
CREATE TRIGGER trigger_update_service_credentials_updated_at
    BEFORE UPDATE ON zervigo_service_credentials
    FOR EACH ROW
    EXECUTE FUNCTION update_service_credentials_updated_at();
