-- Rollback: Remove name field from content tables

-- Drop indexes first
DROP INDEX IF EXISTS idx_novel_name_text_search;
DROP INDEX IF EXISTS idx_manga_name_text_search;
DROP INDEX IF EXISTS idx_anime_name_text_search;

DROP INDEX IF EXISTS idx_novel_name;
DROP INDEX IF EXISTS idx_manga_name;
DROP INDEX IF EXISTS idx_anime_name;

-- Remove name columns
ALTER TABLE novel DROP COLUMN IF EXISTS name;
ALTER TABLE manga DROP COLUMN IF EXISTS name;
ALTER TABLE anime DROP COLUMN IF EXISTS name;