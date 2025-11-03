-- Migration Rollback: Remove active column from oauth2_clients
-- Description: Rollback migration 000011
-- Author: System
-- Created: 2025-11-02

-- Drop index
DROP INDEX IF EXISTS identify.idx_oauth2_clients_active;

-- Remove column
ALTER TABLE identify.oauth2_clients
DROP COLUMN IF EXISTS active;
