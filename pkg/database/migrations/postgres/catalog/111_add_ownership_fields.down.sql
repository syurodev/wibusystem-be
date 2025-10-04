-- Rollback Migration 111: Remove Ownership Model

-- ====================
-- NOVEL TABLE ROLLBACK
-- ====================

-- Drop indexes
DROP INDEX IF EXISTS idx_novel_owner_type_composite;
DROP INDEX IF EXISTS idx_novel_access_level;
DROP INDEX IF EXISTS idx_novel_original_creator_id;
DROP INDEX IF EXISTS idx_novel_primary_owner_id;
DROP INDEX IF EXISTS idx_novel_ownership_type;

-- Remove ownership columns
ALTER TABLE novel DROP COLUMN IF EXISTS ownership_transferred_at;
ALTER TABLE novel DROP COLUMN IF EXISTS last_modified_by_user_id;
ALTER TABLE novel DROP COLUMN IF EXISTS access_level;
ALTER TABLE novel DROP COLUMN IF EXISTS original_creator_id;
ALTER TABLE novel DROP COLUMN IF EXISTS primary_owner_id;
ALTER TABLE novel DROP COLUMN IF EXISTS ownership_type;

-- Restore old user tracking columns
ALTER TABLE novel
    ADD COLUMN created_by_user_id UUID,
    ADD COLUMN updated_by_user_id UUID,
    ADD COLUMN tenant_id UUID;

-- ==========================
-- NOVEL VOLUME TABLE ROLLBACK
-- ==========================

DROP INDEX IF EXISTS idx_novel_volume_modified_by;
ALTER TABLE novel_volume DROP COLUMN IF EXISTS last_modified_by_user_id;

ALTER TABLE novel_volume
    ADD COLUMN created_by_user_id UUID,
    ADD COLUMN updated_by_user_id UUID;

-- ===========================
-- NOVEL CHAPTER TABLE ROLLBACK
-- ===========================

DROP INDEX IF EXISTS idx_novel_chapter_modified_by;
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS last_modified_by_user_id;

ALTER TABLE novel_chapter
    ADD COLUMN created_by_user_id UUID,
    ADD COLUMN updated_by_user_id UUID;

-- ====================
-- DROP HELPER FUNCTION
-- ====================

DROP FUNCTION IF EXISTS get_content_primary_owner(ownership_type, UUID);

-- ====================
-- DROP ENUMS
-- ====================

DROP TYPE IF EXISTS access_level;
DROP TYPE IF EXISTS ownership_type;
