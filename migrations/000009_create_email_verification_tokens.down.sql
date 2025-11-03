-- Migration Rollback: Remove email verification tokens table
-- Description: Rollback migration 000009
-- Author: System
-- Created: 2025-11-02

-- Drop function
DROP FUNCTION IF EXISTS identify.cleanup_expired_verification_tokens();

-- Drop table
DROP TABLE IF EXISTS identify.email_verification_tokens;
