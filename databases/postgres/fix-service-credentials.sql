-- 修复服务凭证
-- 更新服务密钥为正确的 bcrypt 哈希值
-- 
-- 说明：数据库中的 service_secret 字段应存储 bcrypt 哈希值
-- 以下哈希值对应的明文密钥在 local.env 中配置

-- 更新 central-brain 服务凭证
-- bcrypt 哈希对应明文: central-brain-secret-2025
INSERT INTO zervigo_service_credentials (service_id, service_name, service_type, service_secret, description, allowed_apis, created_by, status)
VALUES 
    ('central-brain', 'Central Brain (API Gateway)', 'infrastructure', 
     '$2a$10$B0x/JcQ2Mza6O.HZWjR0TuvAdmPq1Fvdklwijdtz5O29vfGPMZ1h.',
     'API网关服务，负责路由转发和请求代理', 
     ARRAY['*'], 
     'system',
     'active')
ON CONFLICT (service_id) 
DO UPDATE SET 
    service_secret = '$2a$10$B0x/JcQ2Mza6O.HZWjR0TuvAdmPq1Fvdklwijdtz5O29vfGPMZ1h.',
    service_name = 'Central Brain (API Gateway)',
    status = 'active',
    updated_at = CURRENT_TIMESTAMP;

-- 显示更新结果
SELECT service_id, service_name, status, 
       LEFT(service_secret, 20) || '...' as secret_preview,
       updated_at
FROM zervigo_service_credentials
WHERE service_id = 'central-brain';
