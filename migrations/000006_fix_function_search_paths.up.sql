-- Migration: Fix Function Search Paths
-- Description: Set search_path cho functions trong identify schema để tìm được tables
-- Author: System
-- Created: 2025-10-26

-- =====================================================
-- Set Search Path for RBAC Functions
-- Description: Functions cần tìm tables trong identify schema
-- =====================================================

ALTER FUNCTION identify.user_has_tenant_permission(UUID, UUID, VARCHAR) SET search_path = identify, public;
ALTER FUNCTION identify.user_has_global_permission(UUID, VARCHAR) SET search_path = identify, public;
ALTER FUNCTION identify.get_user_tenant_permissions(UUID, UUID) SET search_path = identify, public;
ALTER FUNCTION identify.get_user_global_permissions(UUID) SET search_path = identify, public;

-- =====================================================
-- Set Search Path for OAuth2 Functions
-- Description: Functions cần tìm tables trong identify schema
-- =====================================================

ALTER FUNCTION identify.cleanup_expired_oauth2_data() SET search_path = identify, public;
ALTER FUNCTION identify.is_token_revoked(VARCHAR) SET search_path = identify, public;
ALTER FUNCTION identify.revoke_token(VARCHAR, TIMESTAMPTZ) SET search_path = identify, public;
ALTER FUNCTION identify.revoke_user_client_tokens(UUID, UUID) SET search_path = identify, public;
ALTER FUNCTION identify.revoke_all_user_tokens(UUID) SET search_path = identify, public;

-- =====================================================
-- Verify Function Configuration
-- Description: Kiểm tra search_path đã được set
-- =====================================================

DO $$
DECLARE
    func_count INTEGER;
BEGIN
    -- Count functions with correct search_path
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
        )
        AND p.proconfig IS NOT NULL;  -- proconfig chứa SET options

    IF func_count < 9 THEN
        RAISE WARNING 'Some functions may not have search_path set. Found % functions with configuration', func_count;
    END IF;

    RAISE NOTICE 'Function search_path configuration complete. % functions configured', func_count;
END $$;

COMMENT ON SCHEMA identify IS 'Migration 000006: Function search paths configured for identify schema';
