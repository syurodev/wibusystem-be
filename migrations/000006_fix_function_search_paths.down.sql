-- Migration Down: Rollback Function Search Path Configuration
-- Description: Reset search_path về default cho tất cả functions

-- =====================================================
-- Reset Search Path to Default
-- =====================================================

ALTER FUNCTION identify.user_has_tenant_permission(UUID, UUID, VARCHAR) RESET search_path;
ALTER FUNCTION identify.user_has_global_permission(UUID, VARCHAR) RESET search_path;
ALTER FUNCTION identify.get_user_tenant_permissions(UUID, UUID) RESET search_path;
ALTER FUNCTION identify.get_user_global_permissions(UUID) RESET search_path;

ALTER FUNCTION identify.cleanup_expired_oauth2_data() RESET search_path;
ALTER FUNCTION identify.is_token_revoked(VARCHAR) RESET search_path;
ALTER FUNCTION identify.revoke_token(VARCHAR, TIMESTAMPTZ) RESET search_path;
ALTER FUNCTION identify.revoke_user_client_tokens(UUID, UUID) RESET search_path;
ALTER FUNCTION identify.revoke_all_user_tokens(UUID) RESET search_path;
