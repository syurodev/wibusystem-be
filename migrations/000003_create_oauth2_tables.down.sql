-- Migration Down: Rollback OAuth2 Tables
-- Description: Xóa các bảng và functions liên quan đến OAuth 2.0

-- Drop helper functions
DROP FUNCTION IF EXISTS revoke_all_user_tokens(UUID);
DROP FUNCTION IF EXISTS revoke_user_client_tokens(UUID, UUID);
DROP FUNCTION IF EXISTS revoke_token(VARCHAR, TIMESTAMPTZ);
DROP FUNCTION IF EXISTS is_token_revoked(VARCHAR);
DROP FUNCTION IF EXISTS cleanup_expired_oauth2_data();

-- Drop triggers
DROP TRIGGER IF EXISTS update_oauth2_clients_updated_at ON oauth2_clients;

-- Drop tables (theo thứ tự dependency)
DROP TABLE IF EXISTS oauth2_jti_blacklist CASCADE;
DROP TABLE IF EXISTS oauth2_sessions CASCADE;
DROP TABLE IF EXISTS oauth2_clients CASCADE;

-- Drop types
DROP TYPE IF EXISTS session_type;
