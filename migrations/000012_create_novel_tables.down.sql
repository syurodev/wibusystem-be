-- Drop triggers first
DROP TRIGGER IF EXISTS trg_chapters_updated_at ON catalog.chapters;
DROP TRIGGER IF EXISTS trg_volumes_updated_at ON catalog.volumes;
DROP TRIGGER IF EXISTS trg_novels_updated_at ON catalog.novels;
DROP TRIGGER IF EXISTS trg_chapters_update_stats ON catalog.chapters;
DROP TRIGGER IF EXISTS trg_volumes_update_novel_stats ON catalog.volumes;

-- Drop trigger functions
DROP FUNCTION IF EXISTS catalog.update_updated_at_column();
DROP FUNCTION IF EXISTS catalog.update_chapter_stats();
DROP FUNCTION IF EXISTS catalog.update_novel_volume_stats();

-- Drop indexes
DROP INDEX IF EXISTS catalog.idx_chapters_content;
DROP INDEX IF EXISTS catalog.idx_chapters_status;
DROP INDEX IF EXISTS catalog.idx_chapters_published;
DROP INDEX IF EXISTS catalog.idx_chapters_display_order;
DROP INDEX IF EXISTS catalog.idx_chapters_volume_id;
DROP INDEX IF EXISTS catalog.idx_chapters_novel_id;

DROP INDEX IF EXISTS catalog.idx_volumes_published;
DROP INDEX IF EXISTS catalog.idx_volumes_display_order;
DROP INDEX IF EXISTS catalog.idx_volumes_novel_id;

DROP INDEX IF EXISTS catalog.idx_novels_synopsis;
DROP INDEX IF EXISTS catalog.idx_novels_metadata;
DROP INDEX IF EXISTS catalog.idx_novels_views;
DROP INDEX IF EXISTS catalog.idx_novels_rating;
DROP INDEX IF EXISTS catalog.idx_novels_created_at;
DROP INDEX IF EXISTS catalog.idx_novels_status;
DROP INDEX IF EXISTS catalog.idx_novels_author_id;

-- Drop tables (in reverse order of creation due to foreign keys)
DROP TABLE IF EXISTS catalog.chapters;
DROP TABLE IF EXISTS catalog.volumes;
DROP TABLE IF EXISTS catalog.novels;

-- Drop enum types
DROP TYPE IF EXISTS catalog.chapter_status;
DROP TYPE IF EXISTS catalog.novel_status;

-- Note: We don't drop the catalog schema as it might be used by other tables
-- If you want to drop the schema entirely, uncomment the following:
-- DROP SCHEMA IF EXISTS catalog CASCADE;
