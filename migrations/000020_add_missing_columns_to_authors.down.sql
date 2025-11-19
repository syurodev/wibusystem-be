-- Drop indexes
DROP INDEX IF EXISTS catalog.idx_authors_is_verified;
DROP INDEX IF EXISTS catalog.idx_authors_follower_count;

-- Remove columns from authors table
ALTER TABLE catalog.authors
DROP COLUMN IF EXISTS total_chapters,
DROP COLUMN IF EXISTS follower_count,
DROP COLUMN IF EXISTS is_verified;
