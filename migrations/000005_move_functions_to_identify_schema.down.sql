-- Migration Down: Rollback Function Schema Migration
-- Description: Di chuyển functions về public schema

-- =====================================================
-- Move Functions Back to Public Schema
-- =====================================================

ALTER FUNCTION identify.user_has_tenant_permission(UUID, UUID, VARCHAR) SET SCHEMA public;
ALTER FUNCTION identify.user_has_global_permission(UUID, VARCHAR) SET SCHEMA public;
ALTER FUNCTION identify.get_user_tenant_permissions(UUID, UUID) SET SCHEMA public;
ALTER FUNCTION identify.get_user_global_permissions(UUID) SET SCHEMA public;

ALTER FUNCTION identify.cleanup_expired_oauth2_data() SET SCHEMA public;
ALTER FUNCTION identify.is_token_revoked(VARCHAR) SET SCHEMA public;
ALTER FUNCTION identify.revoke_token(VARCHAR, TIMESTAMPTZ) SET SCHEMA public;
ALTER FUNCTION identify.revoke_user_client_tokens(UUID, UUID) SET SCHEMA public;
ALTER FUNCTION identify.revoke_all_user_tokens(UUID) SET SCHEMA public;
