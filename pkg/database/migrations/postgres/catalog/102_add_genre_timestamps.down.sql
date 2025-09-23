-- Reverse the genre table timestamp additions
DROP INDEX IF EXISTS idx_genre_name;
ALTER TABLE genre DROP CONSTRAINT IF EXISTS genre_name_unique;
ALTER TABLE genre DROP COLUMN IF EXISTS updated_at;
ALTER TABLE genre DROP COLUMN IF EXISTS created_at;