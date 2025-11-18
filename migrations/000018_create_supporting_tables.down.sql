-- =====================================================
-- Migration 000018 Rollback: Supporting Tables
-- =====================================================

-- Drop triggers
DROP TRIGGER IF EXISTS trg_translators_version ON catalog.translators;
DROP TRIGGER IF EXISTS trg_artists_version ON catalog.artists;
DROP TRIGGER IF EXISTS trg_authors_version ON catalog.authors;
DROP TRIGGER IF EXISTS trg_genres_version ON catalog.genres;

-- Drop junction tables
DROP TABLE IF EXISTS catalog.novel_translators CASCADE;
DROP TABLE IF EXISTS catalog.novel_artists CASCADE;
DROP TABLE IF EXISTS catalog.novel_authors CASCADE;
DROP TABLE IF EXISTS catalog.novel_genres CASCADE;

-- Drop main tables
DROP TABLE IF EXISTS catalog.translators CASCADE;
DROP TABLE IF EXISTS catalog.artists CASCADE;
DROP TABLE IF EXISTS catalog.authors CASCADE;
DROP TABLE IF EXISTS catalog.genres CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS catalog.translator_role;
DROP TYPE IF EXISTS catalog.artist_role;
DROP TYPE IF EXISTS catalog.author_role;
