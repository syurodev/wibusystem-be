-- =====================================================
-- Migration 000016 Rollback: Chapter Translations
-- =====================================================

-- Drop triggers
DROP TRIGGER IF EXISTS trg_translation_contributions_version ON catalog.translation_contributions;
DROP TRIGGER IF EXISTS trg_chapter_translations_version ON catalog.chapter_translations;

-- Drop tables
DROP TABLE IF EXISTS catalog.translation_history CASCADE;
DROP TABLE IF EXISTS catalog.translation_contributions CASCADE;
DROP TABLE IF EXISTS catalog.chapter_translations CASCADE;
