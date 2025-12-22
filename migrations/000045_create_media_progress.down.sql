-- =====================================================
-- Rollback Migration 000045: Media Progress Tracking System
-- =====================================================

-- Drop trigger first
DROP TRIGGER IF EXISTS trg_media_progress_updated_at ON catalog.media_progress;
DROP FUNCTION IF EXISTS catalog.update_media_progress_timestamp();

-- Drop tables (order matters due to potential future FKs)
DROP TABLE IF EXISTS catalog.unit_progress;
DROP TABLE IF EXISTS catalog.media_progress;
