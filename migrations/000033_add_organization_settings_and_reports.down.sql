-- Rollback migration 000033

-- Drop triggers and functions
DROP TRIGGER IF EXISTS trigger_update_org_report_status ON identify.organization_reports;
DROP FUNCTION IF EXISTS identify.update_organization_report_status();

-- Drop tables
DROP TABLE IF EXISTS identify.organization_pending_invites;
DROP TABLE IF EXISTS identify.organization_reports;

-- Drop report_count column and constraint
ALTER TABLE identify.organizations DROP CONSTRAINT IF EXISTS organizations_report_count_check;
ALTER TABLE identify.organizations DROP COLUMN IF EXISTS report_count;

-- Restore original status constraint
ALTER TABLE identify.organizations DROP CONSTRAINT IF EXISTS organizations_status_check;
ALTER TABLE identify.organizations ADD CONSTRAINT organizations_status_check 
    CHECK (status IN ('active', 'suspended', 'archived'));
