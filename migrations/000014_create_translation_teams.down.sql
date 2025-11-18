-- =====================================================
-- Migration 000014 Rollback: Translation Teams
-- =====================================================

-- Drop triggers
DROP TRIGGER IF EXISTS trg_novel_team_assignments_version ON catalog.novel_team_assignments;
DROP TRIGGER IF EXISTS trg_team_members_version ON catalog.team_members;
DROP TRIGGER IF EXISTS trg_translation_teams_version ON catalog.translation_teams;

-- Drop tables
DROP TABLE IF EXISTS catalog.novel_team_assignments CASCADE;
DROP TABLE IF EXISTS catalog.team_members CASCADE;
DROP TABLE IF EXISTS catalog.translation_teams CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS catalog.assignment_status;
DROP TYPE IF EXISTS catalog.team_member_role;
