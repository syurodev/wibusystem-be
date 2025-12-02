-- Migration Down: Rollback Novel Organization Assignments
-- Description: Drop novel_organization_assignments table và assignment_status type

-- Drop triggers
DROP TRIGGER IF EXISTS trg_novel_organization_assignments_version ON catalog.novel_organization_assignments;

-- Drop table
DROP TABLE IF EXISTS catalog.novel_organization_assignments CASCADE;

-- Drop type
DROP TYPE IF EXISTS catalog.assignment_status;
