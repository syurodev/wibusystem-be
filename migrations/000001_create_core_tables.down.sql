-- Migration Down: Rollback Core Tables
-- Description: Xóa các bảng và types đã tạo trong migration 000001

-- Drop triggers first
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (theo thứ tự ngược lại với creation để tránh foreign key issues)
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;

-- Drop types
DROP TYPE IF EXISTS role_scope;
DROP TYPE IF EXISTS permission_scope;

-- Note: Không drop UUID extension vì có thể được sử dụng bởi các tables khác
-- DROP EXTENSION IF EXISTS "uuid-ossp";
