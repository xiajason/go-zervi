#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ -f "$PROJECT_ROOT/configs/local.env" ]; then
  set -a
  # shellcheck disable=SC1091
  source "$PROJECT_ROOT/configs/local.env"
  set +a
fi

PGHOST="${POSTGRESQL_HOST:-localhost}"
PGPORT="${POSTGRESQL_PORT:-5432}"
PGUSER="${POSTGRESQL_USER:-$(whoami)}"
PGDATABASE="${POSTGRESQL_DATABASE:-zervigo_mvp}"
PGPASSWORD_VALUE="${POSTGRESQL_PASSWORD:-}"
PGSSLMODE="${POSTGRESQL_SSL_MODE:-disable}"

echo "📥 导入测试账号到 PostgreSQL ($PGHOST:$PGPORT/$PGDATABASE)"

if [ -n "$PGPASSWORD_VALUE" ]; then
  export PGPASSWORD="$PGPASSWORD_VALUE"
fi

psql "host=$PGHOST port=$PGPORT user=$PGUSER dbname=$PGDATABASE sslmode=$PGSSLMODE" <<'SQL'
CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO zervigo_auth_roles (role_name, description, created_at, updated_at)
VALUES
  ('super_admin', 'Super administrator with full privileges', NOW(), NOW()),
  ('candidate', 'Default candidate role', NOW(), NOW())
ON CONFLICT (role_name) DO UPDATE
  SET description = EXCLUDED.description,
      updated_at = NOW();

INSERT INTO zervigo_auth_users (username, email, phone, password_hash, status, email_verified, phone_verified, subscription_status, created_at, updated_at)
VALUES
  ('admin', 'admin@zervigo.com', '13800000000', crypt('admin123', gen_salt('bf')), 'ACTIVE', TRUE, TRUE, 'standard', NOW(), NOW()),
  ('demo_candidate', 'candidate@zervigo.com', '13900000000', crypt('candidate123', gen_salt('bf')), 'ACTIVE', TRUE, TRUE, 'standard', NOW(), NOW())
ON CONFLICT (username) DO UPDATE
  SET email = EXCLUDED.email,
      phone = EXCLUDED.phone,
      updated_at = NOW();

INSERT INTO zervigo_auth_user_roles (user_id, role_id)
SELECT u.id, r.id
FROM zervigo_auth_users u
JOIN zervigo_auth_roles r ON r.role_name = 'super_admin'
WHERE u.username = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO zervigo_auth_user_roles (user_id, role_id)
SELECT u.id, r.id
FROM zervigo_auth_users u
JOIN zervigo_auth_roles r ON r.role_name = 'candidate'
WHERE u.username = 'demo_candidate'
ON CONFLICT DO NOTHING;
SQL

if [ -n "$PGPASSWORD_VALUE" ]; then
  unset PGPASSWORD
fi

echo "✅ 测试账号导入完成：\n  • 管理员：admin / admin123\n  • 候选人：demo_candidate / candidate123"

