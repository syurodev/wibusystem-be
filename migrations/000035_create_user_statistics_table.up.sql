-- Migration: Create user_statistics table
-- Description: Tách statistics từ users table sang bảng riêng
-- Author: System
-- Created: 2025-12-15

-- =====================================================
-- Table: user_statistics
-- Description: Lưu trữ các chỉ số thống kê của user
-- =====================================================
CREATE TABLE identify.user_statistics (
    user_id uuid PRIMARY KEY REFERENCES identify.users(id) ON DELETE CASCADE,
    
    -- Social stats
    follower_count INTEGER DEFAULT 0 NOT NULL,
    following_count INTEGER DEFAULT 0 NOT NULL,
    
    -- Content counts (by type)
    novel_count INTEGER DEFAULT 0 NOT NULL,
    manga_count INTEGER DEFAULT 0 NOT NULL,
    anime_count INTEGER DEFAULT 0 NOT NULL,
    
    -- Chapter/Episode counts
    novel_chapter_count INTEGER DEFAULT 0 NOT NULL,
    manga_chapter_count INTEGER DEFAULT 0 NOT NULL,
    anime_episode_count INTEGER DEFAULT 0 NOT NULL,
    
    -- Engagement stats
    total_views BIGINT DEFAULT 0 NOT NULL,
    
    -- Activity timestamps
    last_content_updated_at TIMESTAMPTZ,
    last_activity_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    
    -- Constraints
    CONSTRAINT user_statistics_counts_check CHECK (
        novel_count >= 0 AND manga_count >= 0 AND anime_count >= 0 AND
        novel_chapter_count >= 0 AND manga_chapter_count >= 0 AND anime_episode_count >= 0 AND
        follower_count >= 0 AND following_count >= 0 AND total_views >= 0
    )
);

-- Indexes
CREATE INDEX idx_user_statistics_follower_count ON identify.user_statistics(follower_count DESC);
CREATE INDEX idx_user_statistics_novel_count ON identify.user_statistics(novel_count DESC);
CREATE INDEX idx_user_statistics_total_views ON identify.user_statistics(total_views DESC);
CREATE INDEX idx_user_statistics_last_content_updated_at ON identify.user_statistics(last_content_updated_at DESC NULLS LAST);

-- Comments
COMMENT ON TABLE identify.user_statistics IS 'Bảng lưu trữ các chỉ số thống kê của user, tách biệt khỏi thông tin identity';
COMMENT ON COLUMN identify.user_statistics.novel_count IS 'Số lượng novel đã đăng';
COMMENT ON COLUMN identify.user_statistics.manga_count IS 'Số lượng manga đã đăng';
COMMENT ON COLUMN identify.user_statistics.anime_count IS 'Số lượng anime đã đăng';
COMMENT ON COLUMN identify.user_statistics.novel_chapter_count IS 'Tổng số chapter novel đã đăng';
COMMENT ON COLUMN identify.user_statistics.manga_chapter_count IS 'Tổng số chapter manga đã đăng';
COMMENT ON COLUMN identify.user_statistics.anime_episode_count IS 'Tổng số episode anime đã đăng';
COMMENT ON COLUMN identify.user_statistics.total_views IS 'Tổng lượt xem tất cả nội dung';
COMMENT ON COLUMN identify.user_statistics.last_content_updated_at IS 'Thời điểm cập nhật nội dung gần nhất';
COMMENT ON COLUMN identify.user_statistics.last_activity_at IS 'Thời điểm hoạt động gần nhất';

-- Trigger for updated_at
CREATE TRIGGER update_user_statistics_updated_at
    BEFORE UPDATE ON identify.user_statistics
    FOR EACH ROW
EXECUTE FUNCTION public.update_updated_at_column();

-- =====================================================
-- Migrate existing data from users table
-- =====================================================
INSERT INTO identify.user_statistics (
    user_id, 
    follower_count, 
    novel_count,  -- Migrate works_count to novel_count as default
    last_content_updated_at,
    created_at,
    updated_at
)
SELECT 
    id, 
    COALESCE(follower_count, 0), 
    COALESCE(works_count, 0),
    last_content_updated_at,
    created_at,
    NOW()
FROM identify.users
WHERE id IS NOT NULL;

-- =====================================================
-- Drop old columns from users table
-- =====================================================
DROP INDEX IF EXISTS identify.idx_users_follower_count;
DROP INDEX IF EXISTS identify.idx_users_last_content_updated_at;

ALTER TABLE identify.users 
    DROP COLUMN IF EXISTS follower_count,
    DROP COLUMN IF EXISTS works_count,
    DROP COLUMN IF EXISTS last_content_updated_at;
