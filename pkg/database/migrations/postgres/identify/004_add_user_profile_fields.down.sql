-- Remove trigger and function
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Remove index
DROP INDEX IF EXISTS idx_users_is_blocked;

-- Remove columns
ALTER TABLE users
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS is_blocked,
DROP COLUMN IF EXISTS bio,
DROP COLUMN IF EXISTS cover_image_url,
DROP COLUMN IF EXISTS avatar_url;
