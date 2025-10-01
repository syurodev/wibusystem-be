-- Rollback Enhanced Novel Tables Migration
-- Removes all fields added in 105_enhance_novel_tables.up.sql

-- =============================================================================
-- DROP INDEXES
-- =============================================================================

-- Novel indexes
DROP INDEX IF EXISTS idx_novel_tenant_id;
DROP INDEX IF EXISTS idx_novel_created_by_user_id;
DROP INDEX IF EXISTS idx_novel_published_at;
DROP INDEX IF EXISTS idx_novel_is_public;
DROP INDEX IF EXISTS idx_novel_is_featured;
DROP INDEX IF EXISTS idx_novel_is_deleted;
DROP INDEX IF EXISTS idx_novel_age_rating;
DROP INDEX IF EXISTS idx_novel_mature_content;
DROP INDEX IF EXISTS idx_novel_tags;
DROP INDEX IF EXISTS idx_novel_rating_average;

-- Volume indexes
DROP INDEX IF EXISTS idx_novel_volume_created_by_user_id;
DROP INDEX IF EXISTS idx_novel_volume_published_at;
DROP INDEX IF EXISTS idx_novel_volume_is_deleted;

-- Chapter indexes
DROP INDEX IF EXISTS idx_novel_chapter_created_by_user_id;
DROP INDEX IF EXISTS idx_novel_chapter_scheduled_publish_at;
DROP INDEX IF EXISTS idx_novel_chapter_is_draft;
DROP INDEX IF EXISTS idx_novel_chapter_is_deleted;
DROP INDEX IF EXISTS idx_novel_chapter_has_mature_content;

-- =============================================================================
-- REMOVE NOVEL TABLE COLUMNS
-- =============================================================================

-- User and Tenant tracking
ALTER TABLE novel DROP COLUMN IF EXISTS created_by_user_id;
ALTER TABLE novel DROP COLUMN IF EXISTS updated_by_user_id;
ALTER TABLE novel DROP COLUMN IF EXISTS tenant_id;

-- Publishing information
ALTER TABLE novel DROP COLUMN IF EXISTS published_at;
ALTER TABLE novel DROP COLUMN IF EXISTS original_language;
ALTER TABLE novel DROP COLUMN IF EXISTS source_url;
ALTER TABLE novel DROP COLUMN IF EXISTS isbn;

-- Content Rating và Warnings
ALTER TABLE novel DROP COLUMN IF EXISTS age_rating;
ALTER TABLE novel DROP COLUMN IF EXISTS content_warnings;
ALTER TABLE novel DROP COLUMN IF EXISTS mature_content;

-- Visibility and Status
ALTER TABLE novel DROP COLUMN IF EXISTS is_public;
ALTER TABLE novel DROP COLUMN IF EXISTS is_featured;
ALTER TABLE novel DROP COLUMN IF EXISTS is_completed;
ALTER TABLE novel DROP COLUMN IF EXISTS is_deleted;
ALTER TABLE novel DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE novel DROP COLUMN IF EXISTS deleted_by_user_id;

-- SEO and Discovery
ALTER TABLE novel DROP COLUMN IF EXISTS slug;
ALTER TABLE novel DROP COLUMN IF EXISTS tags;
ALTER TABLE novel DROP COLUMN IF EXISTS keywords;
ALTER TABLE novel DROP COLUMN IF EXISTS meta_description;

-- Analytics và Engagement
ALTER TABLE novel DROP COLUMN IF EXISTS view_count;
ALTER TABLE novel DROP COLUMN IF EXISTS like_count;
ALTER TABLE novel DROP COLUMN IF EXISTS bookmark_count;
ALTER TABLE novel DROP COLUMN IF EXISTS comment_count;
ALTER TABLE novel DROP COLUMN IF EXISTS rating_average;
ALTER TABLE novel DROP COLUMN IF EXISTS rating_count;

-- Pricing và Monetization
ALTER TABLE novel DROP COLUMN IF EXISTS price_coins;
ALTER TABLE novel DROP COLUMN IF EXISTS rental_price_coins;
ALTER TABLE novel DROP COLUMN IF EXISTS rental_duration_days;
ALTER TABLE novel DROP COLUMN IF EXISTS is_premium;

-- Metadata bổ sung
ALTER TABLE novel DROP COLUMN IF EXISTS total_chapters;
ALTER TABLE novel DROP COLUMN IF EXISTS total_volumes;
ALTER TABLE novel DROP COLUMN IF EXISTS estimated_reading_time;
ALTER TABLE novel DROP COLUMN IF EXISTS word_count;

-- =============================================================================
-- REMOVE NOVEL_VOLUME TABLE COLUMNS
-- =============================================================================

-- User tracking
ALTER TABLE novel_volume DROP COLUMN IF EXISTS created_by_user_id;
ALTER TABLE novel_volume DROP COLUMN IF EXISTS updated_by_user_id;

-- Publishing và Status
ALTER TABLE novel_volume DROP COLUMN IF EXISTS published_at;
ALTER TABLE novel_volume DROP COLUMN IF EXISTS is_deleted;
ALTER TABLE novel_volume DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE novel_volume DROP COLUMN IF EXISTS is_available;

-- Content metadata
ALTER TABLE novel_volume DROP COLUMN IF EXISTS page_count;
ALTER TABLE novel_volume DROP COLUMN IF EXISTS word_count;
ALTER TABLE novel_volume DROP COLUMN IF EXISTS chapter_count;
ALTER TABLE novel_volume DROP COLUMN IF EXISTS estimated_reading_time;

-- =============================================================================
-- REMOVE NOVEL_CHAPTER TABLE COLUMNS
-- =============================================================================

-- User tracking
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS created_by_user_id;
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS updated_by_user_id;

-- Publishing workflow
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS scheduled_publish_at;
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS is_draft;
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS is_deleted;
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS version;

-- Content metadata
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS word_count;
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS reading_time_minutes;
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS character_count;

-- Analytics
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS view_count;
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS like_count;
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS comment_count;

-- Content warnings
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS content_warnings;
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS has_mature_content;