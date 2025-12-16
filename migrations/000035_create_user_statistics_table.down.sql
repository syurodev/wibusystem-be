-- Migration Down: Restore columns to users table and drop user_statistics
-- Description: Rollback user_statistics separation

-- =====================================================
-- Add columns back to users table
-- =====================================================
ALTER TABLE identify.users 
    ADD COLUMN IF NOT EXISTS follower_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS works_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_content_updated_at TIMESTAMPTZ;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_users_follower_count ON identify.users(follower_count);
CREATE INDEX IF NOT EXISTS idx_users_last_content_updated_at ON identify.users(last_content_updated_at);

-- =====================================================
-- Migrate data back from user_statistics
-- =====================================================
UPDATE identify.users u
SET 
    follower_count = COALESCE(us.follower_count, 0),
    works_count = COALESCE(us.novel_count, 0) + COALESCE(us.manga_count, 0) + COALESCE(us.anime_count, 0),
    last_content_updated_at = us.last_content_updated_at
FROM identify.user_statistics us
WHERE u.id = us.user_id;

-- =====================================================
-- Drop user_statistics table
-- =====================================================
DROP TRIGGER IF EXISTS update_user_statistics_updated_at ON identify.user_statistics;
DROP TABLE IF EXISTS identify.user_statistics;
