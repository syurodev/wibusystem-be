-- Migration: Move Helper Functions to Identify Schema
-- Description: Di chuyển tất cả helper functions vào identify schema để đồng nhất với tables
-- Author: System
-- Created: 2025-10-26

-- =====================================================
-- Move RBAC Helper Functions
-- Description: Di chuyển functions kiểm tra permissions
-- =====================================================

ALTER FUNCTION user_has_tenant_permission(UUID, UUID, VARCHAR) SET SCHEMA identify;
ALTER FUNCTION user_has_global_permission(UUID, VARCHAR) SET SCHEMA identify;
ALTER FUNCTION get_user_tenant_permissions(UUID, UUID) SET SCHEMA identify;
ALTER FUNCTION get_user_global_permissions(UUID) SET SCHEMA identify;

-- =====================================================
-- Move OAuth2 Helper Functions
-- Description: Di chuyển functions quản lý OAuth2 tokens
-- =====================================================

ALTER FUNCTION cleanup_expired_oauth2_data() SET SCHEMA identify;
ALTER FUNCTION is_token_revoked(VARCHAR) SET SCHEMA identify;
ALTER FUNCTION revoke_token(VARCHAR, TIMESTAMPTZ) SET SCHEMA identify;
ALTER FUNCTION revoke_user_client_tokens(UUID, UUID) SET SCHEMA identify;
ALTER FUNCTION revoke_all_user_tokens(UUID) SET SCHEMA identify;

-- =====================================================
-- Verify Function Migration
-- Description: Kiểm tra tất cả functions đã được move
-- =====================================================

DO $$
DECLARE
    func_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO func_count
    FROM pg_proc p
    JOIN pg_namespace n ON p.pronamespace = n.oid
    WHERE n.nspname = 'identify'
        AND p.proname IN (
            'user_has_tenant_permission',
            'user_has_global_permission',
            'get_user_tenant_permissions',
            'get_user_global_permissions',
            'cleanup_expired_oauth2_data',
            'is_token_revoked',
            'revoke_token',
            'revoke_user_client_tokens',
            'revoke_all_user_tokens'
        );

    IF func_count < 9 THEN
        RAISE EXCEPTION 'Function migration incomplete. Expected 9 functions in identify schema, found %', func_count;
    END IF;

    RAISE NOTICE 'Function migration successful. % functions in identify schema', func_count;
END $$;

COMMENT ON SCHEMA identify IS 'Migration 000005: All helper functions moved to this schema';
