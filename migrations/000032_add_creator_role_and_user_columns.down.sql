-- Rollback: Remove CREATOR role and user columns
-- Version: 000032

-- Remove CREATOR role permissions and role
DELETE FROM identify.role_permissions 
WHERE role_id = (SELECT id FROM identify.roles WHERE name = 'CREATOR');

DELETE FROM identify.roles WHERE name = 'CREATOR';

-- Remove indexes
DROP INDEX IF EXISTS identify.idx_users_username;
DROP INDEX IF EXISTS identify.idx_users_is_verified;
DROP INDEX IF EXISTS identify.idx_users_last_content_updated_at;
DROP INDEX IF EXISTS identify.idx_users_follower_count;

-- Remove columns
ALTER TABLE identify.users 
DROP COLUMN IF EXISTS display_name,
DROP COLUMN IF EXISTS username,
DROP COLUMN IF EXISTS bio,
DROP COLUMN IF EXISTS is_verified,
DROP COLUMN IF EXISTS follower_count,
DROP COLUMN IF EXISTS works_count,
DROP COLUMN IF EXISTS last_content_updated_at;
