-- OAuth2 Provider 数据库表迁移脚本
-- 创建时间: 2025-11-25
-- 说明: 为 gozervi OAuth2 Provider 功能创建必要的数据库表

-- OAuth2 客户端表
CREATE TABLE IF NOT EXISTS oauth2_clients (
    id VARCHAR(255) PRIMARY KEY,
    client_secret VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    redirect_uris JSONB NOT NULL DEFAULT '[]',
    scopes JSONB DEFAULT '[]',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_oauth2_clients_status ON oauth2_clients(status);
CREATE INDEX idx_oauth2_clients_name ON oauth2_clients(name);

COMMENT ON TABLE oauth2_clients IS 'OAuth2 客户端表，存储注册的 OAuth2 应用信息';
COMMENT ON COLUMN oauth2_clients.id IS '客户端 ID（唯一标识符）';
COMMENT ON COLUMN oauth2_clients.client_secret IS '客户端密钥（加密存储）';
COMMENT ON COLUMN oauth2_clients.name IS '客户端名称';
COMMENT ON COLUMN oauth2_clients.redirect_uris IS '允许的重定向 URI 列表（JSON 数组）';
COMMENT ON COLUMN oauth2_clients.scopes IS '允许的权限范围（JSON 数组）';
COMMENT ON COLUMN oauth2_clients.status IS '客户端状态：active, inactive, revoked';

-- 授权码表
CREATE TABLE IF NOT EXISTS oauth2_authorization_codes (
    code VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL,
    user_id INTEGER NOT NULL,
    redirect_uri VARCHAR(512) NOT NULL,
    scope VARCHAR(255),
    code_challenge VARCHAR(255),  -- PKCE support
    code_challenge_method VARCHAR(10),  -- 'plain' or 'S256'
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES zervigo_auth_users(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES oauth2_clients(id) ON DELETE CASCADE
);

CREATE INDEX idx_oauth2_codes_client ON oauth2_authorization_codes(client_id);
CREATE INDEX idx_oauth2_codes_user ON oauth2_authorization_codes(user_id);
CREATE INDEX idx_oauth2_codes_expires ON oauth2_authorization_codes(expires_at);

COMMENT ON TABLE oauth2_authorization_codes IS 'OAuth2 授权码表，存储临时授权码';
COMMENT ON COLUMN oauth2_authorization_codes.code IS '授权码（唯一标识符）';
COMMENT ON COLUMN oauth2_authorization_codes.client_id IS '客户端 ID';
COMMENT ON COLUMN oauth2_authorization_codes.user_id IS '用户 ID';
COMMENT ON COLUMN oauth2_authorization_codes.redirect_uri IS '重定向 URI';
COMMENT ON COLUMN oauth2_authorization_codes.scope IS '权限范围';
COMMENT ON COLUMN oauth2_authorization_codes.code_challenge IS 'PKCE code challenge';
COMMENT ON COLUMN oauth2_authorization_codes.code_challenge_method IS 'PKCE code challenge method';
COMMENT ON COLUMN oauth2_authorization_codes.expires_at IS '过期时间';

-- 刷新令牌表
CREATE TABLE IF NOT EXISTS oauth2_refresh_tokens (
    token VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL,
    user_id INTEGER NOT NULL,
    scope VARCHAR(255),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    revoked_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES zervigo_auth_users(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES oauth2_clients(id) ON DELETE CASCADE
);

CREATE INDEX idx_oauth2_refresh_client ON oauth2_refresh_tokens(client_id);
CREATE INDEX idx_oauth2_refresh_user ON oauth2_refresh_tokens(user_id);
CREATE INDEX idx_oauth2_refresh_expires ON oauth2_refresh_tokens(expires_at);
CREATE INDEX idx_oauth2_refresh_revoked ON oauth2_refresh_tokens(revoked_at);

COMMENT ON TABLE oauth2_refresh_tokens IS 'OAuth2 刷新令牌表，存储刷新令牌';
COMMENT ON COLUMN oauth2_refresh_tokens.token IS '刷新令牌（唯一标识符）';
COMMENT ON COLUMN oauth2_refresh_tokens.client_id IS '客户端 ID';
COMMENT ON COLUMN oauth2_refresh_tokens.user_id IS '用户 ID';
COMMENT ON COLUMN oauth2_refresh_tokens.scope IS '权限范围';
COMMENT ON COLUMN oauth2_refresh_tokens.expires_at IS '过期时间';
COMMENT ON COLUMN oauth2_refresh_tokens.revoked_at IS '撤销时间（NULL 表示未撤销）';

-- 访问令牌表（可选，如果使用数据库存储访问令牌）
-- 注意：如果使用 JWT，访问令牌可以存储在客户端，此表可选
CREATE TABLE IF NOT EXISTS oauth2_access_tokens (
    token VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL,
    user_id INTEGER NOT NULL,
    scope VARCHAR(255),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    revoked_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES zervigo_auth_users(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES oauth2_clients(id) ON DELETE CASCADE
);

CREATE INDEX idx_oauth2_access_client ON oauth2_access_tokens(client_id);
CREATE INDEX idx_oauth2_access_user ON oauth2_access_tokens(user_id);
CREATE INDEX idx_oauth2_access_expires ON oauth2_access_tokens(expires_at);
CREATE INDEX idx_oauth2_access_revoked ON oauth2_access_tokens(revoked_at);

COMMENT ON TABLE oauth2_access_tokens IS 'OAuth2 访问令牌表（可选，如果使用数据库存储）';
COMMENT ON COLUMN oauth2_access_tokens.token IS '访问令牌（JWT 或随机字符串）';
COMMENT ON COLUMN oauth2_access_tokens.client_id IS '客户端 ID';
COMMENT ON COLUMN oauth2_access_tokens.user_id IS '用户 ID';
COMMENT ON COLUMN oauth2_access_tokens.scope IS '权限范围';
COMMENT ON COLUMN oauth2_access_tokens.expires_at IS '过期时间';
COMMENT ON COLUMN oauth2_access_tokens.revoked_at IS '撤销时间（NULL 表示未撤销）';

-- 清理过期数据的函数（可选）
CREATE OR REPLACE FUNCTION cleanup_expired_oauth2_tokens()
RETURNS void AS $$
BEGIN
    -- 清理过期的授权码
    DELETE FROM oauth2_authorization_codes WHERE expires_at < NOW();
    
    -- 清理过期的刷新令牌
    DELETE FROM oauth2_refresh_tokens WHERE expires_at < NOW() AND revoked_at IS NULL;
    
    -- 清理过期的访问令牌（如果使用数据库存储）
    DELETE FROM oauth2_access_tokens WHERE expires_at < NOW() AND revoked_at IS NULL;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_expired_oauth2_tokens() IS '清理过期的 OAuth2 令牌和授权码';

