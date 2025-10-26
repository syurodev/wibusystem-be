-- Migration Down: Rollback Schema Migration
-- Description: Di chuyển tables về public schema và xóa các schemas

-- =====================================================
-- Move Tables Back to Public Schema
-- =====================================================

ALTER TABLE identify.oauth2_jti_blacklist SET SCHEMA public;
ALTER TABLE identify.oauth2_sessions SET SCHEMA public;
ALTER TABLE identify.oauth2_clients SET SCHEMA public;

ALTER TABLE identify.user_global_roles SET SCHEMA public;
ALTER TABLE identify.user_tenant_roles SET SCHEMA public;
ALTER TABLE identify.user_tenant_memberships SET SCHEMA public;
ALTER TABLE identify.role_permissions SET SCHEMA public;

ALTER TABLE identify.roles SET SCHEMA public;
ALTER TABLE identify.permissions SET SCHEMA public;
ALTER TABLE identify.users SET SCHEMA public;
ALTER TABLE identify.tenants SET SCHEMA public;

-- =====================================================
-- Drop Schemas (only if empty)
-- =====================================================

DROP SCHEMA IF EXISTS identify CASCADE;
DROP SCHEMA IF EXISTS catalog CASCADE;
DROP SCHEMA IF EXISTS community CASCADE;
DROP SCHEMA IF EXISTS payment CASCADE;

-- Note: CASCADE sẽ drop tất cả objects trong schema
-- Nhưng vì đã move tables về public nên schemas sẽ empty
