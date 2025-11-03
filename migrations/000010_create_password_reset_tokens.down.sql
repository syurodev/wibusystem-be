-- Migration Rollback: Remove password reset tokens table
-- Description: Rollback migration 000010
-- Author: System
-- Created: 2025-11-02

-- Drop function
DROP FUNCTION IF EXISTS identify.cleanup_expired_password_reset_tokens();

-- Drop table
DROP TABLE IF EXISTS identify.password_reset_tokens;
