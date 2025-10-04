-- Migration 113: Create Content Collaborators System
-- Enables collaborative ownership with granular permissions

-- Create collaborator permission enum
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'collaborator_permission'
    ) THEN
        CREATE TYPE collaborator_permission AS ENUM (
            'READ',           -- Can view content
            'EDIT',           -- Can edit content
            'PUBLISH',        -- Can publish/unpublish content
            'DELETE',         -- Can delete content
            'MANAGE_CHAPTERS',-- Can add/remove chapters
            'MANAGE_PRICING', -- Can update pricing
            'VIEW_ANALYTICS', -- Can view analytics
            'MANAGE_COLLAB'   -- Can manage other collaborators
        );
    END IF;
END$$;

-- Create collaborator status enum
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'collaborator_status'
    ) THEN
        CREATE TYPE collaborator_status AS ENUM (
            'PENDING',    -- Invitation sent, not accepted
            'ACTIVE',     -- Actively collaborating
            'INACTIVE',   -- Temporarily inactive
            'REMOVED'     -- Removed from collaboration
        );
    END IF;
END$$;

-- ==========================
-- CONTENT COLLABORATORS TABLE
-- ==========================

CREATE TABLE content_collaborators (
    id UUID PRIMARY KEY DEFAULT uuidv7(),

    -- Content identification
    content_type content_type NOT NULL,
    content_id UUID NOT NULL,  -- ID of the novel/volume/chapter

    -- Collaborator identification
    collaborator_id UUID NOT NULL,      -- User ID of collaborator
    collaborator_type VARCHAR(20) NOT NULL DEFAULT 'user', -- Always 'user' for now, can extend to 'tenant'

    -- Collaboration details
    role VARCHAR(100),                  -- Optional role label (e.g., 'Co-Author', 'Editor', 'Translator')
    permissions collaborator_permission[] NOT NULL, -- Array of permissions
    status collaborator_status NOT NULL DEFAULT 'PENDING',

    -- Revenue sharing (if applicable)
    revenue_share_percent DECIMAL(5,2), -- Percentage of revenue (0.00 to 100.00)
    revenue_notes JSONB,                 -- Additional revenue sharing terms

    -- Workflow tracking
    invited_by_user_id UUID NOT NULL,    -- User who sent the invitation
    accepted_at TIMESTAMPTZ,             -- When invitation was accepted
    removed_by_user_id UUID,             -- User who removed the collaborator
    removed_at TIMESTAMPTZ,              -- When collaborator was removed

    -- Metadata
    collaboration_notes TEXT,            -- Notes about the collaboration
    custom_permissions JSONB,            -- Custom permission overrides

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT check_collaborator_type CHECK (
        collaborator_type IN ('user', 'tenant')
    ),
    CONSTRAINT check_revenue_share CHECK (
        revenue_share_percent IS NULL OR
        (revenue_share_percent >= 0 AND revenue_share_percent <= 100)
    ),
    CONSTRAINT check_status_timestamps CHECK (
        (status = 'ACTIVE' AND accepted_at IS NOT NULL) OR
        (status != 'ACTIVE') OR
        (status = 'REMOVED' AND removed_at IS NOT NULL) OR
        (status != 'REMOVED')
    ),
    -- Prevent duplicate collaborators
    UNIQUE (content_type, content_id, collaborator_id)
);

-- ====================
-- INDEXES
-- ====================

CREATE INDEX idx_content_collaborators_content ON content_collaborators(content_type, content_id);
CREATE INDEX idx_content_collaborators_user ON content_collaborators(collaborator_id);
CREATE INDEX idx_content_collaborators_status ON content_collaborators(status);
CREATE INDEX idx_content_collaborators_active ON content_collaborators(content_type, content_id, status)
    WHERE status = 'ACTIVE';

-- Composite index for permission checks
CREATE INDEX idx_content_collaborators_permissions ON content_collaborators(content_type, content_id, collaborator_id, status)
    WHERE status = 'ACTIVE';

-- ====================
-- TRIGGER FUNCTIONS
-- ====================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_collaborator_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_collaborator_updated_at
    BEFORE UPDATE ON content_collaborators
    FOR EACH ROW
    EXECUTE FUNCTION update_collaborator_updated_at();

-- Function to validate permissions array
CREATE OR REPLACE FUNCTION validate_collaborator_permissions()
RETURNS TRIGGER AS $$
BEGIN
    -- Ensure permissions array is not empty
    IF array_length(NEW.permissions, 1) IS NULL OR array_length(NEW.permissions, 1) = 0 THEN
        RAISE EXCEPTION 'Collaborator must have at least one permission';
    END IF;

    -- Ensure MANAGE_COLLAB requires other permissions
    IF 'MANAGE_COLLAB' = ANY(NEW.permissions) THEN
        IF NOT ('EDIT' = ANY(NEW.permissions)) THEN
            RAISE EXCEPTION 'MANAGE_COLLAB permission requires EDIT permission';
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_validate_collaborator_permissions
    BEFORE INSERT OR UPDATE ON content_collaborators
    FOR EACH ROW
    EXECUTE FUNCTION validate_collaborator_permissions();

-- Function to prevent owner from being added as collaborator
CREATE OR REPLACE FUNCTION prevent_owner_as_collaborator()
RETURNS TRIGGER AS $$
DECLARE
    v_primary_owner_id UUID;
    v_ownership_type ownership_type;
BEGIN
    -- Get ownership info based on content type
    CASE NEW.content_type
        WHEN 'NOVEL' THEN
            SELECT primary_owner_id, novel.ownership_type
            INTO v_primary_owner_id, v_ownership_type
            FROM novel WHERE id = NEW.content_id;
        WHEN 'VOLUME' THEN
            SELECT n.primary_owner_id, n.ownership_type
            INTO v_primary_owner_id, v_ownership_type
            FROM novel_volume nv
            JOIN novel n ON nv.novel_id = n.id
            WHERE nv.id = NEW.content_id;
        WHEN 'CHAPTER' THEN
            SELECT n.primary_owner_id, n.ownership_type
            INTO v_primary_owner_id, v_ownership_type
            FROM novel_chapter nc
            JOIN novel_volume nv ON nc.volume_id = nv.id
            JOIN novel n ON nv.novel_id = n.id
            WHERE nc.id = NEW.content_id;
    END CASE;

    -- For PERSONAL ownership, prevent owner from being collaborator
    IF v_ownership_type = 'PERSONAL' AND v_primary_owner_id = NEW.collaborator_id THEN
        RAISE EXCEPTION 'Content owner cannot be added as collaborator';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_prevent_owner_as_collaborator
    BEFORE INSERT OR UPDATE ON content_collaborators
    FOR EACH ROW
    EXECUTE FUNCTION prevent_owner_as_collaborator();

-- ====================
-- HELPER FUNCTIONS
-- ====================

-- Function to check if user has specific permission on content
CREATE OR REPLACE FUNCTION has_collaborator_permission(
    p_content_type content_type,
    p_content_id UUID,
    p_user_id UUID,
    p_permission collaborator_permission
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM content_collaborators
        WHERE content_type = p_content_type
          AND content_id = p_content_id
          AND collaborator_id = p_user_id
          AND status = 'ACTIVE'
          AND p_permission = ANY(permissions)
    );
END;
$$ LANGUAGE plpgsql;

-- Function to get all collaborators for content
CREATE OR REPLACE FUNCTION get_content_collaborators(
    p_content_type content_type,
    p_content_id UUID
) RETURNS TABLE (
    collaborator_id UUID,
    role VARCHAR(100),
    permissions collaborator_permission[],
    status collaborator_status,
    revenue_share_percent DECIMAL(5,2)
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        content_collaborators.collaborator_id,
        content_collaborators.role,
        content_collaborators.permissions,
        content_collaborators.status,
        content_collaborators.revenue_share_percent
    FROM content_collaborators
    WHERE content_collaborators.content_type = p_content_type
      AND content_collaborators.content_id = p_content_id
      AND content_collaborators.status = 'ACTIVE'
    ORDER BY content_collaborators.created_at;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate total revenue share for content
CREATE OR REPLACE FUNCTION calculate_total_revenue_share(
    p_content_type content_type,
    p_content_id UUID
) RETURNS DECIMAL(5,2) AS $$
DECLARE
    v_total DECIMAL(5,2);
BEGIN
    SELECT COALESCE(SUM(revenue_share_percent), 0)
    INTO v_total
    FROM content_collaborators
    WHERE content_type = p_content_type
      AND content_id = p_content_id
      AND status = 'ACTIVE'
      AND revenue_share_percent IS NOT NULL;

    RETURN v_total;
END;
$$ LANGUAGE plpgsql;

-- ====================
-- COMMENTS
-- ====================

COMMENT ON TABLE content_collaborators IS 'Manages collaborative content creation with granular permissions';
COMMENT ON COLUMN content_collaborators.permissions IS 'Array of permission enums granted to collaborator';
COMMENT ON COLUMN content_collaborators.revenue_share_percent IS 'Percentage of revenue allocated to this collaborator (0-100)';
COMMENT ON COLUMN content_collaborators.status IS 'Collaboration status: PENDING (invited) → ACTIVE (accepted) → REMOVED';
COMMENT ON COLUMN content_collaborators.role IS 'Human-readable role label like "Co-Author" or "Editor"';
COMMENT ON COLUMN content_collaborators.custom_permissions IS 'JSONB for future custom permission extensions';
