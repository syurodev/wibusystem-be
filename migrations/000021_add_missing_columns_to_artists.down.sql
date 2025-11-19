-- Drop indexes
DROP INDEX IF EXISTS catalog.idx_artists_specialization;
DROP INDEX IF EXISTS catalog.idx_artists_is_verified;
DROP INDEX IF EXISTS catalog.idx_artists_follower_count;

-- Remove columns from artists table
ALTER TABLE catalog.artists
DROP COLUMN IF EXISTS specialization,
DROP COLUMN IF EXISTS artwork_count,
DROP COLUMN IF EXISTS follower_count,
DROP COLUMN IF EXISTS is_verified;
