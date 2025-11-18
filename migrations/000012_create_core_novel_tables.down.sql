-- =====================================================
-- Migration 000012 Rollback: Core Novel System Tables
-- =====================================================

-- Drop triggers first
DROP TRIGGER IF EXISTS trg_chapters_version ON catalog.chapters;
DROP TRIGGER IF EXISTS trg_volumes_version ON catalog.volumes;
DROP TRIGGER IF EXISTS trg_novels_version ON catalog.novels;

-- Drop tables (cascade will handle foreign keys)
DROP TABLE IF EXISTS catalog.chapters CASCADE;
DROP TABLE IF EXISTS catalog.volumes CASCADE;
DROP TABLE IF EXISTS catalog.novels CASCADE;

-- Drop functions
DROP FUNCTION IF EXISTS catalog.update_timestamp();
DROP FUNCTION IF EXISTS catalog.increment_version();

-- Drop enum types
DROP TYPE IF EXISTS catalog.chapter_status;
DROP TYPE IF EXISTS catalog.novel_status;

-- Note: We don't drop the catalog schema as other migrations may use it
-- If you need to drop the schema completely, do it manually after all migrations are rolled back
