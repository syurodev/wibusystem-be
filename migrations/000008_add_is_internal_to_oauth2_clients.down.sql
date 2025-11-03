-- Migration Rollback: Remove is_internal field from oauth2_clients
-- Description: Rollback migration 000008 - xóa trường is_internal
-- Author: System
-- Created: 2025-11-02

-- Drop index
DROP INDEX IF EXISTS identify.idx_oauth2_clients_is_internal;

-- Remove column
ALTER TABLE identify.oauth2_clients
DROP COLUMN IF EXISTS is_internal;
