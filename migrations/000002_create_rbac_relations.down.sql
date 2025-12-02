-- Migration Down: Rollback RBAC Relation Tables
-- Description: Xóa các bảng và functions đã tạo trong migration 000002

-- Drop functions
DROP FUNCTION IF EXISTS get_user_global_permissions(UUID);
DROP FUNCTION IF EXISTS get_user_organization_permissions(UUID, UUID);
DROP FUNCTION IF EXISTS user_has_global_permission(UUID, VARCHAR);
DROP FUNCTION IF EXISTS user_has_organization_permission(UUID, UUID, VARCHAR);

-- Drop triggers
DROP TRIGGER IF EXISTS update_user_organization_memberships_updated_at ON user_organization_memberships;

-- Drop tables (theo thứ tự ngược lại)
DROP TABLE IF EXISTS user_global_roles CASCADE;
DROP TABLE IF EXISTS user_organization_roles CASCADE;
DROP TABLE IF EXISTS user_organization_memberships CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;

-- Drop types
DROP TYPE IF EXISTS organization_member_role;
