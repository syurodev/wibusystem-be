-- Migration 112: Create Content Transfer System
-- Handles ownership transfer workflow with two-step approval

-- Create transfer status enum
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'transfer_status'
    ) THEN
        CREATE TYPE transfer_status AS ENUM (
            'PENDING',      -- Waiting for approval
            'APPROVED',     -- Approved, ready to execute
            'COMPLETED',    -- Transfer completed successfully
            'REJECTED',     -- Transfer rejected
            'CANCELLED',    -- Transfer cancelled by initiator
            'EXPIRED'       -- Transfer request expired
        );
    END IF;
END$$;

-- Create content type enum for transfers
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'content_type'
    ) THEN
        CREATE TYPE content_type AS ENUM (
            'NOVEL',
            'VOLUME',
            'CHAPTER'
        );
    END IF;
END$$;

-- ==========================
-- CONTENT TRANSFERS TABLE
-- ==========================

CREATE TABLE content_transfers (
    id UUID PRIMARY KEY DEFAULT uuidv7(),

    -- Content identification
    content_type content_type NOT NULL,
    content_id UUID NOT NULL,  -- ID of the novel/volume/chapter being transferred

    -- Transfer parties
    from_owner_id UUID NOT NULL,        -- Current owner (user_id or tenant_id)
    from_owner_type VARCHAR(20) NOT NULL, -- 'user' or 'tenant'
    to_owner_id UUID NOT NULL,          -- New owner (user_id or tenant_id)
    to_owner_type VARCHAR(20) NOT NULL, -- 'user' or 'tenant'

    -- Transfer details
    new_ownership_type ownership_type NOT NULL, -- Target ownership type after transfer
    new_access_level access_level NOT NULL,     -- Target access level after transfer
    status transfer_status NOT NULL DEFAULT 'PENDING',

    -- Workflow tracking
    initiated_by_user_id UUID NOT NULL,  -- User who initiated the transfer
    approved_by_user_id UUID,            -- User who approved (for tenant transfers)
    completed_by_user_id UUID,           -- User who executed the final transfer

    -- Additional metadata
    transfer_reason TEXT,                 -- Optional: reason for transfer
    transfer_notes JSONB,                -- Optional: additional metadata
    conditions JSONB,                    -- Optional: conditions for transfer (e.g., royalty split)

    -- Timestamps
    initiated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,              -- Transfer request expiration

    -- Constraints
    CONSTRAINT check_different_owners CHECK (
        NOT (from_owner_id = to_owner_id AND from_owner_type = to_owner_type)
    ),
    CONSTRAINT check_owner_types CHECK (
        from_owner_type IN ('user', 'tenant') AND
        to_owner_type IN ('user', 'tenant')
    ),
    CONSTRAINT check_status_timestamps CHECK (
        (status = 'APPROVED' AND approved_at IS NOT NULL) OR
        (status != 'APPROVED') OR
        (status = 'COMPLETED' AND completed_at IS NOT NULL) OR
        (status != 'COMPLETED')
    )
);

-- ====================
-- INDEXES
-- ====================

CREATE INDEX idx_content_transfers_content ON content_transfers(content_type, content_id);
CREATE INDEX idx_content_transfers_from_owner ON content_transfers(from_owner_type, from_owner_id);
CREATE INDEX idx_content_transfers_to_owner ON content_transfers(to_owner_type, to_owner_id);
CREATE INDEX idx_content_transfers_status ON content_transfers(status);
CREATE INDEX idx_content_transfers_initiated_by ON content_transfers(initiated_by_user_id);
CREATE INDEX idx_content_transfers_pending ON content_transfers(status, expires_at)
    WHERE status = 'PENDING';

-- Composite index for active transfer lookup
CREATE INDEX idx_content_transfers_active ON content_transfers(content_type, content_id, status)
    WHERE status IN ('PENDING', 'APPROVED');

-- ====================
-- TRIGGER FUNCTIONS
-- ====================

-- Function to prevent duplicate active transfers
CREATE OR REPLACE FUNCTION check_duplicate_transfer()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status IN ('PENDING', 'APPROVED') THEN
        IF EXISTS (
            SELECT 1 FROM content_transfers
            WHERE content_type = NEW.content_type
              AND content_id = NEW.content_id
              AND status IN ('PENDING', 'APPROVED')
              AND id != NEW.id
        ) THEN
            RAISE EXCEPTION 'An active transfer already exists for this content';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_check_duplicate_transfer
    BEFORE INSERT OR UPDATE ON content_transfers
    FOR EACH ROW
    EXECUTE FUNCTION check_duplicate_transfer();

-- Function to auto-expire old pending transfers
CREATE OR REPLACE FUNCTION expire_old_transfers()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE content_transfers
    SET status = 'EXPIRED'
    WHERE status = 'PENDING'
      AND expires_at IS NOT NULL
      AND expires_at < NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to run expiration check periodically (on insert/update)
CREATE TRIGGER trigger_expire_old_transfers
    AFTER INSERT OR UPDATE ON content_transfers
    FOR EACH STATEMENT
    EXECUTE FUNCTION expire_old_transfers();

-- ====================
-- HELPER FUNCTIONS
-- ====================

-- Function to get active transfer for content
CREATE OR REPLACE FUNCTION get_active_transfer(
    p_content_type content_type,
    p_content_id UUID
) RETURNS TABLE (
    transfer_id UUID,
    status transfer_status,
    from_owner_id UUID,
    to_owner_id UUID,
    initiated_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        id,
        content_transfers.status,
        content_transfers.from_owner_id,
        content_transfers.to_owner_id,
        content_transfers.initiated_at
    FROM content_transfers
    WHERE content_transfers.content_type = p_content_type
      AND content_transfers.content_id = p_content_id
      AND content_transfers.status IN ('PENDING', 'APPROVED')
    ORDER BY initiated_at DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- ====================
-- COMMENTS
-- ====================

COMMENT ON TABLE content_transfers IS 'Tracks ownership transfer requests and workflow';
COMMENT ON COLUMN content_transfers.content_type IS 'Type of content: NOVEL, VOLUME, or CHAPTER';
COMMENT ON COLUMN content_transfers.from_owner_id IS 'Current owner UUID (user_id or tenant_id)';
COMMENT ON COLUMN content_transfers.to_owner_id IS 'Target owner UUID (user_id or tenant_id)';
COMMENT ON COLUMN content_transfers.status IS 'Transfer workflow status: PENDING → APPROVED → COMPLETED';
COMMENT ON COLUMN content_transfers.conditions IS 'JSONB for transfer conditions like royalty splits';
COMMENT ON COLUMN content_transfers.expires_at IS 'When the transfer request expires if not acted upon';
