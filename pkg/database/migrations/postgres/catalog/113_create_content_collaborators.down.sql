-- Rollback Migration 113: Remove Content Collaborators System

-- Drop helper functions
DROP FUNCTION IF EXISTS calculate_total_revenue_share(content_type, UUID);
DROP FUNCTION IF EXISTS get_content_collaborators(content_type, UUID);
DROP FUNCTION IF EXISTS has_collaborator_permission(content_type, UUID, UUID, collaborator_permission);
DROP FUNCTION IF EXISTS prevent_owner_as_collaborator();
DROP FUNCTION IF EXISTS validate_collaborator_permissions();
DROP FUNCTION IF EXISTS update_collaborator_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_content_collaborators_permissions;
DROP INDEX IF EXISTS idx_content_collaborators_active;
DROP INDEX IF EXISTS idx_content_collaborators_status;
DROP INDEX IF EXISTS idx_content_collaborators_user;
DROP INDEX IF EXISTS idx_content_collaborators_content;

-- Drop table
DROP TABLE IF EXISTS content_collaborators;

-- Drop enums
DROP TYPE IF EXISTS collaborator_status;
DROP TYPE IF EXISTS collaborator_permission;
