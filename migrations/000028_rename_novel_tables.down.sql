-- =====================================================
-- Migration 000028 Rollback: Rename Novel Tables Back
-- Description: Rollback - rename tables back to original names
-- =====================================================

-- Rename tables back to original names
ALTER TABLE catalog.novel_chapter_histories RENAME TO chapter_history;
ALTER TABLE catalog.novel_chapter_translations RENAME TO chapter_translations;
ALTER TABLE catalog.novel_chapters RENAME TO chapters;
ALTER TABLE catalog.novel_volume_histories RENAME TO volume_history;
ALTER TABLE catalog.novel_volumes RENAME TO volumes;
