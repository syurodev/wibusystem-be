-- =====================================================
-- Migration 000028: Rename Novel Tables
-- Description: Rename tables to add 'novel_' prefix for consistency
-- =====================================================

-- Rename tables
-- PostgreSQL automatically renames dependent objects (indexes, constraints, triggers)
ALTER TABLE catalog.volumes RENAME TO novel_volumes;
ALTER TABLE catalog.volume_history RENAME TO novel_volume_histories;
ALTER TABLE catalog.chapters RENAME TO novel_chapters;
ALTER TABLE catalog.chapter_translations RENAME TO novel_chapter_translations;
ALTER TABLE catalog.chapter_history RENAME TO novel_chapter_histories;
