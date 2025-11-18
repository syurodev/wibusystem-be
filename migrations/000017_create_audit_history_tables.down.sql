-- =====================================================
-- Migration 000017 Rollback: Audit History Tables
-- =====================================================

-- Drop tables
DROP TABLE IF EXISTS catalog.chapter_history CASCADE;
DROP TABLE IF EXISTS catalog.volume_history CASCADE;
DROP TABLE IF EXISTS catalog.novel_history CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS catalog.audit_action;
