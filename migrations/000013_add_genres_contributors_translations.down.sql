-- Drop triggers
DROP TRIGGER IF EXISTS trg_chapter_translations_history ON catalog.chapter_translations;
DROP TRIGGER IF EXISTS trg_chapter_translations_update_stats ON catalog.chapter_translations;
DROP TRIGGER IF EXISTS trg_novel_authors_update_stats ON catalog.novel_authors;

-- Drop trigger functions
DROP FUNCTION IF EXISTS catalog.create_translation_history();
DROP FUNCTION IF EXISTS catalog.update_translator_stats();
DROP FUNCTION IF EXISTS catalog.update_author_novel_count();

-- Drop indexes
DROP INDEX IF EXISTS catalog.idx_translation_history_version;
DROP INDEX IF EXISTS catalog.idx_translation_history_translation_id;
DROP INDEX IF EXISTS catalog.idx_translation_contributions_reviewed_by;
DROP INDEX IF EXISTS catalog.idx_translation_contributions_language;
DROP INDEX IF EXISTS catalog.idx_translation_contributions_status;
DROP INDEX IF EXISTS catalog.idx_translation_contributions_contributor_id;
DROP INDEX IF EXISTS catalog.idx_translation_contributions_chapter_id;
DROP INDEX IF EXISTS catalog.idx_chapter_translations_content;
DROP INDEX IF EXISTS catalog.idx_chapter_translations_status;
DROP INDEX IF EXISTS catalog.idx_chapter_translations_translator_id;
DROP INDEX IF EXISTS catalog.idx_chapter_translations_language;
DROP INDEX IF EXISTS catalog.idx_chapter_translations_chapter_id;
DROP INDEX IF EXISTS catalog.idx_novel_translators_language;
DROP INDEX IF EXISTS catalog.idx_novel_translators_translator_id;
DROP INDEX IF EXISTS catalog.idx_novel_artists_artist_id;
DROP INDEX IF EXISTS catalog.idx_novel_authors_author_id;
DROP INDEX IF EXISTS catalog.idx_translators_languages;
DROP INDEX IF EXISTS catalog.idx_translators_slug;
DROP INDEX IF EXISTS catalog.idx_translators_user_id;
DROP INDEX IF EXISTS catalog.idx_artists_slug;
DROP INDEX IF EXISTS catalog.idx_artists_user_id;
DROP INDEX IF EXISTS catalog.idx_authors_verified;
DROP INDEX IF EXISTS catalog.idx_authors_slug;
DROP INDEX IF EXISTS catalog.idx_authors_user_id;
DROP INDEX IF EXISTS catalog.idx_novel_genres_novel_id;
DROP INDEX IF EXISTS catalog.idx_novel_genres_genre_id;
DROP INDEX IF EXISTS catalog.idx_genres_active;
DROP INDEX IF EXISTS catalog.idx_genres_slug;
DROP INDEX IF EXISTS catalog.idx_genres_parent_id;

-- Drop tables (in reverse order of creation due to foreign keys)
DROP TABLE IF EXISTS catalog.translation_history;
DROP TABLE IF EXISTS catalog.translation_contributions;
DROP TABLE IF EXISTS catalog.chapter_translations;
DROP TABLE IF EXISTS catalog.novel_translators;
DROP TABLE IF EXISTS catalog.novel_artists;
DROP TABLE IF EXISTS catalog.novel_authors;
DROP TABLE IF EXISTS catalog.translators;
DROP TABLE IF EXISTS catalog.artists;
DROP TABLE IF EXISTS catalog.authors;
DROP TABLE IF EXISTS catalog.novel_genres;
DROP TABLE IF EXISTS catalog.genres;

-- Drop enum types
DROP TYPE IF EXISTS catalog.contribution_type;
DROP TYPE IF EXISTS catalog.translation_status;
