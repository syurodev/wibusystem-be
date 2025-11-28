-- Migration: Rollback WebAuthn Credentials Table
-- Description: Xóa bảng WebAuthn credentials và revert changes to users table
-- Author: System
-- Created: 2025-11-22

-- Drop tables (in reverse order of creation)
DROP TABLE IF EXISTS identify.webauthn_sessions;
DROP TABLE IF EXISTS identify.webauthn_credentials;

-- Revert users table changes
ALTER TABLE identify.users ALTER COLUMN password_hash SET NOT NULL;
COMMENT ON COLUMN identify.users.password_hash IS 'Password hash';
