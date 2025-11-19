-- Drop index
DROP INDEX IF EXISTS catalog.idx_genres_is_active;

-- Remove is_active column from genres table
ALTER TABLE catalog.genres
DROP COLUMN IF EXISTS is_active;
