-- Migration Down: Rollback RBAC Relations
-- Description: Xóa các bảng quan hệ RBAC và helper functions

-- Drop helper functions
DROP FUNCTION IF EXISTS get_user_global_permissions(UUID);
DROP FUNCTION IF EXISTS get_user_tenant_permissions(UUID, UUID);
DROP FUNCTION IF EXISTS user_has_global_permission(UUID, VARCHAR);
DROP FUNCTION IF EXISTS user_has_tenant_permission(UUID, UUID, VARCHAR);

-- Drop triggers
DROP TRIGGER IF EXISTS update_user_tenant_memberships_updated_at ON user_tenant_memberships;

-- Drop tables (theo thứ tự dependency)
DROP TABLE IF EXISTS user_global_roles CASCADE;
DROP TABLE IF EXISTS user_tenant_roles CASCADE;
DROP TABLE IF EXISTS user_tenant_memberships CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
