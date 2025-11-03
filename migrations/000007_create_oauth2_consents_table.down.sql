-- Migration: Drop OAuth2 Consents Table
-- Description: Rollback migration cho bảng oauth2_consents

-- Drop functions
DROP FUNCTION IF EXISTS identify.get_active_consent(UUID, UUID);
DROP FUNCTION IF EXISTS identify.revoke_all_user_consents(UUID);
DROP FUNCTION IF EXISTS identify.revoke_consent(UUID, UUID);
DROP FUNCTION IF EXISTS identify.cleanup_expired_consents();

-- Drop table
DROP TABLE IF EXISTS identify.oauth2_consents;
