-- =====================================================
-- Migration 000015 Rollback: Synopsis Translations
-- =====================================================

-- Drop triggers
DROP TRIGGER IF EXISTS trg_synopsis_contributions_version ON catalog.synopsis_translation_contributions;
DROP TRIGGER IF EXISTS trg_synopsis_translations_version ON catalog.novel_synopsis_translations;

-- Drop tables
DROP TABLE IF EXISTS catalog.synopsis_translation_contributions CASCADE;
DROP TABLE IF EXISTS catalog.novel_synopsis_translations CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS catalog.contribution_status;
DROP TYPE IF EXISTS catalog.translation_status;
