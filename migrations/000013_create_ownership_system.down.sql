-- =====================================================
-- Migration 000013 Rollback: Ownership System
-- =====================================================

-- Drop triggers
DROP TRIGGER IF EXISTS trg_exclusive_reports_version ON catalog.exclusive_translation_reports;
DROP TRIGGER IF EXISTS trg_ownership_transfers_version ON catalog.ownership_transfers;

-- Drop tables
DROP TABLE IF EXISTS catalog.exclusive_translation_reports CASCADE;
DROP TABLE IF EXISTS catalog.ownership_transfers CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS catalog.report_status;
DROP TYPE IF EXISTS catalog.transfer_status;
