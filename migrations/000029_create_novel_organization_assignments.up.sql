-- =====================================================
-- Migration 000029: Novel Organization Assignments
-- Description: Assigns organizations to translate specific novels
-- Merged from migration 000014 novel_team_assignments
-- =====================================================

-- Enum type for assignment_status (merged from 000014)
-- Use IF NOT EXISTS in case it was created by old migration 000014
DO $$ BEGIN
    CREATE TYPE catalog.assignment_status AS ENUM (
        'active',
        'inactive',
        'suspended'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- =====================================================
-- NOVEL ORGANIZATION ASSIGNMENTS TABLE
-- =====================================================
CREATE TABLE catalog.novel_organization_assignments (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES identify.organizations(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- ISO 639-1: en, vi, zh, ja, ko, etc.

    -- Assignment details
    status catalog.assignment_status NOT NULL DEFAULT 'active',
    has_exclusive_rights BOOLEAN NOT NULL DEFAULT FALSE,

    -- Statistics (auto-updated by application)
    chapters_translated INTEGER NOT NULL DEFAULT 0 CHECK (chapters_translated >= 0),
    chapters_proofread INTEGER NOT NULL DEFAULT 0 CHECK (chapters_proofread >= 0),

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    -- One organization per novel per language (can have multiple if no exclusive rights)
    UNIQUE(novel_id, language, organization_id)
);

-- Indexes for novel_organization_assignments
CREATE INDEX idx_novel_organization_assignments_novel ON catalog.novel_organization_assignments(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_novel_organization_assignments_organization ON catalog.novel_organization_assignments(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_novel_organization_assignments_language ON catalog.novel_organization_assignments(novel_id, language) WHERE deleted_at IS NULL;
CREATE INDEX idx_novel_organization_assignments_exclusive ON catalog.novel_organization_assignments(novel_id, language, has_exclusive_rights) WHERE has_exclusive_rights = TRUE AND status = 'active' AND deleted_at IS NULL;
CREATE INDEX idx_novel_organization_assignments_status ON catalog.novel_organization_assignments(status) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.novel_organization_assignments IS 'Assigns organizations to translate specific novels in specific languages';
COMMENT ON COLUMN catalog.novel_organization_assignments.has_exclusive_rights IS 'If TRUE, this organization claims exclusive translation rights (can be challenged via reports)';
COMMENT ON COLUMN catalog.novel_organization_assignments.status IS 'Assignment status: active, inactive, or suspended';

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Novel organization assignments triggers
CREATE TRIGGER trg_novel_organization_assignments_version
    BEFORE UPDATE ON catalog.novel_organization_assignments
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();
