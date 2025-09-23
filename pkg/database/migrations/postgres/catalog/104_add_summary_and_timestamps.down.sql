-- Reverse the summary and timestamps additions

-- 1. Drop indexes
DROP INDEX IF EXISTS idx_novel_created_at;
DROP INDEX IF EXISTS idx_manga_created_at;
DROP INDEX IF EXISTS idx_anime_created_at;

-- 2. Remove timestamps from user interaction tables
ALTER TABLE user_donations DROP COLUMN IF EXISTS updated_at;
ALTER TABLE user_donations DROP COLUMN IF EXISTS created_at;

ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS updated_at;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS created_at;

ALTER TABLE user_content_rentals DROP COLUMN IF EXISTS updated_at;
ALTER TABLE user_content_rentals DROP COLUMN IF EXISTS created_at;

ALTER TABLE user_content_purchases DROP COLUMN IF EXISTS updated_at;

-- 3. Remove timestamps from translation tables
ALTER TABLE novel_translation DROP COLUMN IF EXISTS updated_at;
ALTER TABLE novel_translation DROP COLUMN IF EXISTS created_at;

ALTER TABLE manga_translation DROP COLUMN IF EXISTS updated_at;
ALTER TABLE manga_translation DROP COLUMN IF EXISTS created_at;

ALTER TABLE anime_translation DROP COLUMN IF EXISTS updated_at;
ALTER TABLE anime_translation DROP COLUMN IF EXISTS created_at;

-- 4. Remove timestamps from relationship tables
ALTER TABLE content_relation DROP COLUMN IF EXISTS updated_at;
ALTER TABLE content_relation DROP COLUMN IF EXISTS created_at;

ALTER TABLE novel_creator DROP COLUMN IF EXISTS updated_at;
ALTER TABLE novel_creator DROP COLUMN IF EXISTS created_at;

ALTER TABLE manga_creator DROP COLUMN IF EXISTS updated_at;
ALTER TABLE manga_creator DROP COLUMN IF EXISTS created_at;

ALTER TABLE anime_creator DROP COLUMN IF EXISTS updated_at;
ALTER TABLE anime_creator DROP COLUMN IF EXISTS created_at;

ALTER TABLE novel_genre DROP COLUMN IF EXISTS updated_at;
ALTER TABLE novel_genre DROP COLUMN IF EXISTS created_at;

ALTER TABLE manga_genre DROP COLUMN IF EXISTS updated_at;
ALTER TABLE manga_genre DROP COLUMN IF EXISTS created_at;

ALTER TABLE anime_genre DROP COLUMN IF EXISTS updated_at;
ALTER TABLE anime_genre DROP COLUMN IF EXISTS created_at;

ALTER TABLE novel_character DROP COLUMN IF EXISTS updated_at;
ALTER TABLE novel_character DROP COLUMN IF EXISTS created_at;

ALTER TABLE manga_character DROP COLUMN IF EXISTS updated_at;
ALTER TABLE manga_character DROP COLUMN IF EXISTS created_at;

ALTER TABLE anime_character DROP COLUMN IF EXISTS updated_at;
ALTER TABLE anime_character DROP COLUMN IF EXISTS created_at;

-- 5. Remove timestamps from novel sub-tables
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS updated_at;
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS created_at;

ALTER TABLE novel_volume DROP COLUMN IF EXISTS updated_at;
ALTER TABLE novel_volume DROP COLUMN IF EXISTS created_at;

-- 6. Remove timestamps from manga sub-tables
ALTER TABLE manga_page DROP COLUMN IF EXISTS updated_at;
ALTER TABLE manga_page DROP COLUMN IF EXISTS created_at;

ALTER TABLE manga_chapter DROP COLUMN IF EXISTS updated_at;
ALTER TABLE manga_chapter DROP COLUMN IF EXISTS created_at;

ALTER TABLE manga_volume DROP COLUMN IF EXISTS updated_at;
ALTER TABLE manga_volume DROP COLUMN IF EXISTS created_at;

-- 7. Remove timestamps from anime sub-tables
ALTER TABLE episode_subtitle DROP COLUMN IF EXISTS updated_at;
ALTER TABLE episode_subtitle DROP COLUMN IF EXISTS created_at;

ALTER TABLE anime_episode DROP COLUMN IF EXISTS updated_at;
ALTER TABLE anime_episode DROP COLUMN IF EXISTS created_at;

ALTER TABLE anime_season DROP COLUMN IF EXISTS updated_at;
ALTER TABLE anime_season DROP COLUMN IF EXISTS created_at;

-- 8. Remove timestamps from main content tables
ALTER TABLE novel DROP COLUMN IF EXISTS updated_at;
ALTER TABLE novel DROP COLUMN IF EXISTS created_at;

ALTER TABLE manga DROP COLUMN IF EXISTS updated_at;
ALTER TABLE manga DROP COLUMN IF EXISTS created_at;

ALTER TABLE anime DROP COLUMN IF EXISTS updated_at;
ALTER TABLE anime DROP COLUMN IF EXISTS created_at;

-- 9. Remove summary fields from main content tables and revert novel_chapter content
ALTER TABLE novel DROP COLUMN IF EXISTS summary;
ALTER TABLE manga DROP COLUMN IF EXISTS summary;
ALTER TABLE anime DROP COLUMN IF EXISTS summary;

-- 9.1. Revert novel_chapter content back to TEXT (if needed)
ALTER TABLE novel_chapter ALTER COLUMN content TYPE TEXT USING content::text;