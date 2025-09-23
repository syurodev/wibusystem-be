-- Add summary field to main content tables (anime, manga, novel)
-- Add created_at, updated_at timestamps to all tables that don't have them

-- 1. Add summary field to main content tables (JSON format like user bio)
ALTER TABLE anime ADD COLUMN summary JSONB;
ALTER TABLE manga ADD COLUMN summary JSONB;
ALTER TABLE novel ADD COLUMN summary JSONB;

-- 1.1. Update novel_chapter content to JSON format
ALTER TABLE novel_chapter ALTER COLUMN content TYPE JSONB USING content::jsonb;

-- 2. Add timestamps to main content tables
ALTER TABLE anime ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE anime ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE manga ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE manga ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE novel ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE novel ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- 3. Add timestamps to anime sub-tables
ALTER TABLE anime_season ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE anime_season ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE anime_episode ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE anime_episode ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE episode_subtitle ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE episode_subtitle ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- 4. Add timestamps to manga sub-tables
ALTER TABLE manga_volume ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE manga_volume ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE manga_chapter ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE manga_chapter ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE manga_page ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE manga_page ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- 5. Add timestamps to novel sub-tables
ALTER TABLE novel_volume ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE novel_volume ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE novel_chapter ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE novel_chapter ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- 6. Add timestamps to relationship tables
ALTER TABLE anime_character ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE anime_character ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE manga_character ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE manga_character ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE novel_character ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE novel_character ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE anime_genre ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE anime_genre ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE manga_genre ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE manga_genre ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE novel_genre ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE novel_genre ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE anime_creator ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE anime_creator ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE manga_creator ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE manga_creator ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE novel_creator ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE novel_creator ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE content_relation ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE content_relation ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- 7. Add timestamps to translation tables
ALTER TABLE anime_translation ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE anime_translation ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE manga_translation ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE manga_translation ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE novel_translation ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE novel_translation ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- 8. Add timestamps to user interaction tables
ALTER TABLE user_content_purchases ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
-- Note: user_content_purchases already has purchase_date, but adding updated_at for consistency

ALTER TABLE user_content_rentals ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE user_content_rentals ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
-- Note: user_content_rentals already has rent_date, treating it as created_at equivalent

ALTER TABLE user_subscriptions ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE user_subscriptions ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE user_donations ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE user_donations ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
-- Note: user_donations already has donation_date, treating it as created_at equivalent

-- 9. Add comments for summary fields and content
COMMENT ON COLUMN anime.summary IS 'Brief summary/synopsis of the anime series stored as JSONB (similar to user bio)';
COMMENT ON COLUMN manga.summary IS 'Brief summary/synopsis of the manga series stored as JSONB (similar to user bio)';
COMMENT ON COLUMN novel.summary IS 'Brief summary/synopsis of the novel series stored as JSONB (similar to user bio)';
COMMENT ON COLUMN novel_chapter.content IS 'Novel chapter content stored as JSONB for rich text support';

-- 10. Add indexes for timestamp-based queries (optional but recommended for performance)
CREATE INDEX idx_anime_created_at ON anime(created_at);
CREATE INDEX idx_manga_created_at ON manga(created_at);
CREATE INDEX idx_novel_created_at ON novel(created_at);