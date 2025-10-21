-- Migration: 000002_create_permissions_system (DOWN)
-- Description: Drops permissions and roles system
-- Schema: identity

BEGIN;

SET search_path TO identity, public;

-- Drop views
DROP VIEW IF EXISTS identity.tenant_member_effective_permissions;
DROP VIEW IF EXISTS identity.user_effective_global_permissions;

-- Drop tables (in reverse dependency order)
DROP TABLE IF EXISTS identity.tenant_member_permissions;
DROP TABLE IF EXISTS identity.user_global_roles;
DROP TABLE IF EXISTS identity.global_role_permissions;
DROP TABLE IF EXISTS identity.permissions;
DROP TABLE IF EXISTS identity.global_roles;
DROP TABLE IF EXISTS identity.global_permissions;

COMMIT;
