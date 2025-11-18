-- =====================================================
-- Migration 000013: Ownership System
-- Description: Ownership transfers and exclusive rights reporting
-- =====================================================

-- =====================================================
-- ENUM TYPES
-- =====================================================

-- Transfer status
CREATE TYPE catalog.transfer_status AS ENUM (
    'pending',
    'approved',
    'rejected',
    'cancelled'
);

-- Report status
CREATE TYPE catalog.report_status AS ENUM (
    'pending',
    'under_review',
    'resolved',
    'rejected'
);

-- =====================================================
-- OWNERSHIP TRANSFERS TABLE
-- =====================================================
CREATE TABLE catalog.ownership_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- Source ownership
    from_owner_type VARCHAR(20) NOT NULL CHECK (from_owner_type IN ('user', 'tenant')),
    from_owner_id UUID NOT NULL,

    -- Target ownership
    to_owner_type VARCHAR(20) NOT NULL CHECK (to_owner_type IN ('user', 'tenant')),
    to_owner_id UUID NOT NULL,

    -- Transfer info
    status catalog.transfer_status NOT NULL DEFAULT 'pending',
    reason TEXT,

    -- Admin review (required for 2-way transfers)
    requires_approval BOOLEAN NOT NULL DEFAULT FALSE,
    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    review_notes TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    -- Ensure from and to are different
    CHECK (
        from_owner_type != to_owner_type OR
        from_owner_id != to_owner_id
    )
);

-- Indexes for ownership_transfers
CREATE INDEX idx_ownership_transfers_novel_id ON catalog.ownership_transfers(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_ownership_transfers_from ON catalog.ownership_transfers(from_owner_type, from_owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_ownership_transfers_to ON catalog.ownership_transfers(to_owner_type, to_owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_ownership_transfers_status ON catalog.ownership_transfers(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_ownership_transfers_pending_approval ON catalog.ownership_transfers(status, requires_approval) WHERE requires_approval = TRUE AND status = 'pending' AND deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.ownership_transfers IS 'Tracks ownership transfer requests between users and tenants';
COMMENT ON COLUMN catalog.ownership_transfers.from_owner_type IS 'Source owner type: user or tenant';
COMMENT ON COLUMN catalog.ownership_transfers.from_owner_id IS 'Source owner UUID (validated in application)';
COMMENT ON COLUMN catalog.ownership_transfers.to_owner_type IS 'Target owner type: user or tenant';
COMMENT ON COLUMN catalog.ownership_transfers.to_owner_id IS 'Target owner UUID (validated in application)';
COMMENT ON COLUMN catalog.ownership_transfers.requires_approval IS 'TRUE for 2-way transfers (user<->tenant or tenant<->tenant), FALSE for 1-way (user->tenant)';

-- =====================================================
-- EXCLUSIVE TRANSLATION REPORTS TABLE
-- =====================================================
CREATE TABLE catalog.exclusive_translation_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- ISO 639-1

    -- Reporting team
    reporting_team_id UUID NOT NULL, -- References catalog.translation_teams (will be created in next migration)

    -- Reported team (claiming exclusive rights)
    reported_team_id UUID NOT NULL, -- References catalog.translation_teams

    -- Report details
    reason TEXT NOT NULL,
    evidence_urls TEXT[], -- Array of URLs to evidence
    status catalog.report_status NOT NULL DEFAULT 'pending',

    -- Admin review
    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    review_notes TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,

    -- Resolution
    resolution TEXT,
    resolved_at TIMESTAMP WITH TIME ZONE,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    -- Ensure reporting and reported teams are different
    CHECK (reporting_team_id != reported_team_id)
);

-- Indexes for exclusive_translation_reports
CREATE INDEX idx_exclusive_reports_novel_id ON catalog.exclusive_translation_reports(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_exclusive_reports_language ON catalog.exclusive_translation_reports(language) WHERE deleted_at IS NULL;
CREATE INDEX idx_exclusive_reports_reporting_team ON catalog.exclusive_translation_reports(reporting_team_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_exclusive_reports_reported_team ON catalog.exclusive_translation_reports(reported_team_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_exclusive_reports_status ON catalog.exclusive_translation_reports(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_exclusive_reports_pending ON catalog.exclusive_translation_reports(status) WHERE status = 'pending' AND deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.exclusive_translation_reports IS 'Reports for teams claiming exclusive translation rights';
COMMENT ON COLUMN catalog.exclusive_translation_reports.reporting_team_id IS 'Team filing the report';
COMMENT ON COLUMN catalog.exclusive_translation_reports.reported_team_id IS 'Team being reported for claiming exclusive rights';
COMMENT ON COLUMN catalog.exclusive_translation_reports.evidence_urls IS 'URLs to evidence supporting the report';

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Ownership transfers triggers
CREATE TRIGGER trg_ownership_transfers_version
    BEFORE UPDATE ON catalog.ownership_transfers
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Exclusive reports triggers
CREATE TRIGGER trg_exclusive_reports_version
    BEFORE UPDATE ON catalog.exclusive_translation_reports
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();
