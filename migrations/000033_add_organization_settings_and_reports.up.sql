-- Migration: Add organization settings and reports feature
-- Version: 000033

-- =====================================================
-- 1. Add bypass_invite_approval to organizations.settings
-- =====================================================

-- Update existing organizations with default settings if null
UPDATE identify.organizations 
SET settings = '{"bypass_invite_approval": false}'::jsonb 
WHERE settings IS NULL OR settings = '{}'::jsonb;

-- Add report_count column to track reports
ALTER TABLE identify.organizations
ADD COLUMN IF NOT EXISTS report_count INTEGER DEFAULT 0 NOT NULL;

-- Add constraint for report_count
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'organizations_report_count_check'
    ) THEN
        ALTER TABLE identify.organizations
        ADD CONSTRAINT organizations_report_count_check CHECK (report_count >= 0);
    END IF;
END $$;

-- Update status constraint to include 'flagged'
ALTER TABLE identify.organizations
DROP CONSTRAINT IF EXISTS organizations_status_check;

ALTER TABLE identify.organizations
ADD CONSTRAINT organizations_status_check CHECK (status IN (
    'active', 'flagged', 'suspended', 'archived'
));

-- =====================================================
-- 2. Create organization_reports table
-- =====================================================

CREATE TABLE IF NOT EXISTS identify.organization_reports (
    id              uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    organization_id uuid NOT NULL REFERENCES identify.organizations ON DELETE CASCADE,
    reporter_id     uuid NOT NULL REFERENCES identify.users ON DELETE CASCADE,
    
    -- Report info
    reason          VARCHAR(50) NOT NULL 
        CONSTRAINT organization_reports_reason_check CHECK (reason IN (
            'spam', 'harassment', 'inappropriate_content', 
            'copyright_violation', 'fake_translations', 'other'
        )),
    description     TEXT,
    
    -- Org response (owner/admin phản hồi)
    org_response        TEXT,
    org_responded_by    uuid REFERENCES identify.users ON DELETE SET NULL,
    org_responded_at    TIMESTAMP WITH TIME ZONE,
    
    -- Admin resolution
    status          VARCHAR(50) DEFAULT 'pending' NOT NULL
        CONSTRAINT organization_reports_status_check CHECK (status IN (
            'pending', 'org_responded', 'reviewing', 'resolved', 'dismissed'
        )),
    resolved_by     uuid REFERENCES identify.users ON DELETE SET NULL,
    resolved_at     TIMESTAMP WITH TIME ZONE,
    resolution_note TEXT,
    
    -- Audit
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    -- Một user chỉ có thể report 1 org 1 lần (pending/org_responded)
    CONSTRAINT organization_reports_unique_report UNIQUE (organization_id, reporter_id)
);

COMMENT ON TABLE identify.organization_reports IS 'Lưu trữ các report về organization từ users';
COMMENT ON COLUMN identify.organization_reports.org_response IS 'Phản hồi từ owner/admin của org';
COMMENT ON COLUMN identify.organization_reports.org_responded_by IS 'User (owner/admin) đã phản hồi';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_organization_reports_org_id ON identify.organization_reports(organization_id);
CREATE INDEX IF NOT EXISTS idx_organization_reports_reporter_id ON identify.organization_reports(reporter_id);
CREATE INDEX IF NOT EXISTS idx_organization_reports_status ON identify.organization_reports(status);
CREATE INDEX IF NOT EXISTS idx_organization_reports_created_at ON identify.organization_reports(created_at);

-- Trigger for updated_at
DROP TRIGGER IF EXISTS update_organization_reports_updated_at ON identify.organization_reports;
CREATE TRIGGER update_organization_reports_updated_at
    BEFORE UPDATE ON identify.organization_reports
    FOR EACH ROW
EXECUTE PROCEDURE public.update_updated_at_column();

-- =====================================================
-- 3. Create pending_member_invites table (for approval flow)
-- =====================================================

CREATE TABLE IF NOT EXISTS identify.organization_pending_invites (
    id              uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    organization_id uuid NOT NULL REFERENCES identify.organizations ON DELETE CASCADE,
    user_id         uuid NOT NULL REFERENCES identify.users ON DELETE CASCADE,
    invited_by      uuid NOT NULL REFERENCES identify.users ON DELETE CASCADE,
    status          VARCHAR(50) DEFAULT 'pending' NOT NULL
        CONSTRAINT organization_pending_invites_status_check CHECK (status IN (
            'pending', 'approved', 'rejected', 'expired'
        )),
    approved_by     uuid REFERENCES identify.users ON DELETE SET NULL,
    processed_at    TIMESTAMP WITH TIME ZONE,
    expires_at      TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '7 days') NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    CONSTRAINT organization_pending_invites_unique UNIQUE (organization_id, user_id)
);

COMMENT ON TABLE identify.organization_pending_invites IS 'Pending invites chờ owner/admin duyệt';

CREATE INDEX IF NOT EXISTS idx_org_pending_invites_org_id ON identify.organization_pending_invites(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_pending_invites_user_id ON identify.organization_pending_invites(user_id);
CREATE INDEX IF NOT EXISTS idx_org_pending_invites_status ON identify.organization_pending_invites(status);
CREATE INDEX IF NOT EXISTS idx_org_pending_invites_expires_at ON identify.organization_pending_invites(expires_at);

-- =====================================================
-- 4. Function to update organization status based on reports
-- =====================================================

CREATE OR REPLACE FUNCTION identify.update_organization_report_status()
RETURNS TRIGGER AS $$
BEGIN
    -- Update report_count (chỉ đếm pending + org_responded)
    UPDATE identify.organizations
    SET report_count = (
        SELECT COUNT(*) FROM identify.organization_reports
        WHERE organization_id = NEW.organization_id 
        AND status IN ('pending', 'org_responded')
    )
    WHERE id = NEW.organization_id;
    
    -- Auto-flag if report_count >= 5
    UPDATE identify.organizations
    SET status = 'flagged'
    WHERE id = NEW.organization_id 
      AND report_count >= 5
      AND status = 'active';
    
    -- Auto-suspend if report_count >= 10
    UPDATE identify.organizations
    SET status = 'suspended'
    WHERE id = NEW.organization_id 
      AND report_count >= 10
      AND status IN ('active', 'flagged');
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_org_report_status ON identify.organization_reports;
CREATE TRIGGER trigger_update_org_report_status
    AFTER INSERT OR UPDATE ON identify.organization_reports
    FOR EACH ROW
EXECUTE FUNCTION identify.update_organization_report_status();
