-- Migration 111: Add Clean Ownership Model
-- Replaces old user tracking fields with ownership system

-- Create ownership type enum
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'ownership_type'
    ) THEN
        CREATE TYPE ownership_type AS ENUM (
            'PERSONAL',      -- Owned by individual user
            'TENANT',        -- Owned by organization/tenant
            'COLLABORATIVE'  -- Shared ownership between user and tenant
        );
    END IF;
END$$;

-- Create access level enum
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'access_level'
    ) THEN
        CREATE TYPE access_level AS ENUM (
            'PRIVATE',      -- Only owner can access
            'TENANT_ONLY',  -- Only tenant members can access
            'PUBLIC'        -- Public access
        );
    END IF;
END$$;

-- ====================
-- NOVEL TABLE CHANGES
-- ====================

-- Drop old user tracking columns
ALTER TABLE novel DROP COLUMN IF EXISTS created_by_user_id;
ALTER TABLE novel DROP COLUMN IF EXISTS updated_by_user_id;
ALTER TABLE novel DROP COLUMN IF EXISTS tenant_id;

-- Add clean ownership fields
ALTER TABLE novel
    ADD COLUMN ownership_type ownership_type NOT NULL DEFAULT 'PERSONAL',
    ADD COLUMN primary_owner_id UUID NOT NULL,
    ADD COLUMN original_creator_id UUID NOT NULL,
    ADD COLUMN access_level access_level NOT NULL DEFAULT 'PRIVATE';

-- Add metadata tracking
ALTER TABLE novel
    ADD COLUMN last_modified_by_user_id UUID,
    ADD COLUMN ownership_transferred_at TIMESTAMPTZ;

-- Create indexes for ownership queries
CREATE INDEX idx_novel_ownership_type ON novel(ownership_type);
CREATE INDEX idx_novel_primary_owner_id ON novel(primary_owner_id);
CREATE INDEX idx_novel_original_creator_id ON novel(original_creator_id);
CREATE INDEX idx_novel_access_level ON novel(access_level);
CREATE INDEX idx_novel_owner_type_composite ON novel(ownership_type, primary_owner_id);

-- ==========================
-- NOVEL VOLUME TABLE CHANGES
-- ==========================

-- Drop old user tracking columns
ALTER TABLE novel_volume DROP COLUMN IF EXISTS created_by_user_id;
ALTER TABLE novel_volume DROP COLUMN IF EXISTS updated_by_user_id;

-- Add metadata tracking (volumes inherit ownership from novel)
ALTER TABLE novel_volume
    ADD COLUMN last_modified_by_user_id UUID;

-- Create index for tracking modifications
CREATE INDEX idx_novel_volume_modified_by ON novel_volume(last_modified_by_user_id);

-- ===========================
-- NOVEL CHAPTER TABLE CHANGES
-- ===========================

-- Drop old user tracking columns
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS created_by_user_id;
ALTER TABLE novel_chapter DROP COLUMN IF EXISTS updated_by_user_id;

-- Add metadata tracking (chapters inherit ownership from novel)
ALTER TABLE novel_chapter
    ADD COLUMN last_modified_by_user_id UUID;

-- Create index for tracking modifications
CREATE INDEX idx_novel_chapter_modified_by ON novel_chapter(last_modified_by_user_id);

-- ====================
-- HELPER FUNCTION
-- ====================

-- Function to get effective owner for content (returns user_id or tenant_id)
CREATE OR REPLACE FUNCTION get_content_primary_owner(
    p_ownership_type ownership_type,
    p_primary_owner_id UUID
) RETURNS TABLE (
    owner_id UUID,
    owner_type TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        p_primary_owner_id as owner_id,
        CASE
            WHEN p_ownership_type = 'PERSONAL' THEN 'user'
            WHEN p_ownership_type = 'TENANT' THEN 'tenant'
            WHEN p_ownership_type = 'COLLABORATIVE' THEN 'tenant'
        END as owner_type;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ====================
-- COMMENTS
-- ====================

COMMENT ON COLUMN novel.ownership_type IS 'Type of ownership: PERSONAL (individual), TENANT (organization), COLLABORATIVE (shared)';
COMMENT ON COLUMN novel.primary_owner_id IS 'UUID of primary owner - user_id for PERSONAL, tenant_id for TENANT/COLLABORATIVE';
COMMENT ON COLUMN novel.original_creator_id IS 'UUID of the user who originally created the content - never changes';
COMMENT ON COLUMN novel.access_level IS 'Access level: PRIVATE (owner only), TENANT_ONLY (tenant members), PUBLIC (everyone)';
COMMENT ON COLUMN novel.last_modified_by_user_id IS 'UUID of user who last modified this content';
COMMENT ON COLUMN novel.ownership_transferred_at IS 'Timestamp when ownership was last transferred';
