-- Rollback Migration 112: Remove Content Transfer System

-- Drop helper functions
DROP FUNCTION IF EXISTS get_active_transfer(content_type, UUID);
DROP FUNCTION IF EXISTS expire_old_transfers();
DROP FUNCTION IF EXISTS check_duplicate_transfer();

-- Drop indexes
DROP INDEX IF EXISTS idx_content_transfers_active;
DROP INDEX IF EXISTS idx_content_transfers_pending;
DROP INDEX IF EXISTS idx_content_transfers_initiated_by;
DROP INDEX IF EXISTS idx_content_transfers_status;
DROP INDEX IF EXISTS idx_content_transfers_to_owner;
DROP INDEX IF EXISTS idx_content_transfers_from_owner;
DROP INDEX IF EXISTS idx_content_transfers_content;

-- Drop table
DROP TABLE IF EXISTS content_transfers;

-- Drop enums
DROP TYPE IF EXISTS content_type;
DROP TYPE IF EXISTS transfer_status;
