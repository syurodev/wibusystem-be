-- Migration: 000001_create_identity_schema (DOWN)
-- Description: Drops identity schema and all its objects
-- Schema: identity

BEGIN;

-- Set search path
SET search_path TO identity, public;

-- ============================================================================
-- DROP TRIGGERS
-- ============================================================================

DROP TRIGGER IF EXISTS update_users_updated_at ON identity.users;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON identity.tenants;
DROP TRIGGER IF EXISTS update_tenant_members_updated_at ON identity.tenant_members;
DROP TRIGGER IF EXISTS update_oauth2_clients_updated_at ON identity.oauth2_clients;

-- ============================================================================
-- DROP FUNCTIONS
-- ============================================================================

DROP FUNCTION IF EXISTS identity.update_updated_at_column();

-- ============================================================================
-- DROP TABLES (in reverse order of dependencies)
-- ============================================================================

-- Drop audit tables
DROP TABLE IF EXISTS identity.user_activities CASCADE;

-- Drop session tables
DROP TABLE IF EXISTS identity.sessions CASCADE;

-- Drop OAuth2 tables
DROP TABLE IF EXISTS identity.oauth2_refresh_tokens CASCADE;
DROP TABLE IF EXISTS identity.oauth2_access_tokens CASCADE;
DROP TABLE IF EXISTS identity.oauth2_authorization_codes CASCADE;
DROP TABLE IF EXISTS identity.oauth2_clients CASCADE;

-- Drop tenant tables
DROP TABLE IF EXISTS identity.tenant_members CASCADE;
DROP TABLE IF EXISTS identity.tenants CASCADE;

-- Drop core tables
DROP TABLE IF EXISTS identity.users CASCADE;

-- ============================================================================
-- DROP ENUMS
-- ============================================================================

DROP TYPE IF EXISTS identity.grant_type CASCADE;
DROP TYPE IF EXISTS identity.tenant_status CASCADE;
DROP TYPE IF EXISTS identity.user_status CASCADE;

-- ============================================================================
-- DROP SCHEMA
-- ============================================================================

DROP SCHEMA IF EXISTS identity CASCADE;

COMMIT;
