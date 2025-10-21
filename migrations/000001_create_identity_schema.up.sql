-- Migration: 000001_create_identity_schema
-- Description: Creates identity schema with users, tenants, OAuth2, and session tables
-- Schema: identity

BEGIN;

-- Create identity schema
CREATE SCHEMA IF NOT EXISTS identity;

-- Set search path for this migration
SET search_path TO identity, public;

-- ============================================================================
-- ENUMS
-- ============================================================================

-- User status enum
CREATE TYPE identity.user_status AS ENUM (
    'active',
    'inactive',
    'suspended',
    'pending_verification'
);

-- Tenant status enum
CREATE TYPE identity.tenant_status AS ENUM (
    'active',
    'inactive',
    'suspended',
    'trial'
);

-- OAuth2 grant types
CREATE TYPE identity.grant_type AS ENUM (
    'authorization_code',
    'client_credentials',
    'refresh_token',
    'password',
    'implicit'
);

-- ============================================================================
-- CORE TABLES
-- ============================================================================

-- Users table
CREATE TABLE identity.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    password_hash TEXT NOT NULL,
    display_name VARCHAR(255),
    avatar_url TEXT,
    status identity.user_status DEFAULT 'pending_verification',
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for users
CREATE INDEX idx_users_email ON identity.users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON identity.users(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON identity.users(created_at);

-- Tenants table (for multi-tenancy)
CREATE TABLE identity.tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    logo_url TEXT,
    status identity.tenant_status DEFAULT 'trial',
    owner_id UUID NOT NULL,
    settings JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for tenants
CREATE INDEX idx_tenants_slug ON identity.tenants(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_owner_id ON identity.tenants(owner_id);
CREATE INDEX idx_tenants_status ON identity.tenants(status);

-- Tenant members (users belonging to tenants)
CREATE TABLE identity.tenant_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    permissions JSONB DEFAULT '[]',
    invited_by UUID,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, user_id)
);

-- Create indexes for tenant_members
CREATE INDEX idx_tenant_members_tenant_id ON identity.tenant_members(tenant_id);
CREATE INDEX idx_tenant_members_user_id ON identity.tenant_members(user_id);
CREATE INDEX idx_tenant_members_role ON identity.tenant_members(role);

-- ============================================================================
-- OAUTH2 TABLES
-- ============================================================================

-- OAuth2 Clients
CREATE TABLE identity.oauth2_clients (
    id VARCHAR(255) PRIMARY KEY,
    client_secret_hash TEXT,
    redirect_uris TEXT[] NOT NULL DEFAULT '{}',
    grant_types TEXT[] NOT NULL DEFAULT '{}',
    response_types TEXT[] NOT NULL DEFAULT '{}',
    scopes TEXT[] NOT NULL DEFAULT '{}',
    audience TEXT[] NOT NULL DEFAULT '{}',
    public BOOLEAN DEFAULT FALSE,
    client_name VARCHAR(255),
    client_uri TEXT,
    logo_uri TEXT,
    contacts TEXT[] DEFAULT '{}',
    tos_uri TEXT,
    policy_uri TEXT,
    jwks_uri TEXT,
    jwks TEXT,
    token_endpoint_auth_method VARCHAR(50) DEFAULT 'client_secret_basic',
    registration_access_token TEXT,
    registration_client_uri TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for oauth2_clients
CREATE INDEX idx_oauth2_clients_public ON identity.oauth2_clients(public);
CREATE INDEX idx_oauth2_clients_created_at ON identity.oauth2_clients(created_at);

-- OAuth2 Authorization Codes
CREATE TABLE identity.oauth2_authorization_codes (
    code VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL,
    redirect_uri TEXT NOT NULL,
    scope TEXT NOT NULL,
    nonce TEXT,
    state TEXT,
    code_challenge TEXT,
    code_challenge_method VARCHAR(10),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked BOOLEAN DEFAULT FALSE
);

-- Create indexes for authorization codes
CREATE INDEX idx_oauth2_auth_codes_client_id ON identity.oauth2_authorization_codes(client_id);
CREATE INDEX idx_oauth2_auth_codes_user_id ON identity.oauth2_authorization_codes(user_id);
CREATE INDEX idx_oauth2_auth_codes_expires_at ON identity.oauth2_authorization_codes(expires_at);

-- OAuth2 Access Tokens
CREATE TABLE identity.oauth2_access_tokens (
    signature VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL,
    user_id UUID,
    scope TEXT NOT NULL,
    token_type VARCHAR(50) DEFAULT 'Bearer',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for access tokens
CREATE INDEX idx_oauth2_access_tokens_client_id ON identity.oauth2_access_tokens(client_id);
CREATE INDEX idx_oauth2_access_tokens_user_id ON identity.oauth2_access_tokens(user_id);
CREATE INDEX idx_oauth2_access_tokens_expires_at ON identity.oauth2_access_tokens(expires_at);
CREATE INDEX idx_oauth2_access_tokens_revoked ON identity.oauth2_access_tokens(revoked);

-- OAuth2 Refresh Tokens
CREATE TABLE identity.oauth2_refresh_tokens (
    signature VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL,
    user_id UUID,
    scope TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for refresh tokens
CREATE INDEX idx_oauth2_refresh_tokens_client_id ON identity.oauth2_refresh_tokens(client_id);
CREATE INDEX idx_oauth2_refresh_tokens_user_id ON identity.oauth2_refresh_tokens(user_id);
CREATE INDEX idx_oauth2_refresh_tokens_expires_at ON identity.oauth2_refresh_tokens(expires_at);
CREATE INDEX idx_oauth2_refresh_tokens_revoked ON identity.oauth2_refresh_tokens(revoked);

-- ============================================================================
-- SESSION TABLES
-- ============================================================================

-- User Sessions
CREATE TABLE identity.sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked BOOLEAN DEFAULT FALSE
);

-- Create indexes for sessions
CREATE INDEX idx_sessions_user_id ON identity.sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON identity.sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON identity.sessions(expires_at);
CREATE INDEX idx_sessions_revoked ON identity.sessions(revoked);

-- ============================================================================
-- AUDIT TABLES
-- ============================================================================

-- User activity log
CREATE TABLE identity.user_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    activity_type VARCHAR(100) NOT NULL,
    description TEXT,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for user_activities
CREATE INDEX idx_user_activities_user_id ON identity.user_activities(user_id);
CREATE INDEX idx_user_activities_type ON identity.user_activities(activity_type);
CREATE INDEX idx_user_activities_created_at ON identity.user_activities(created_at);

-- ============================================================================
-- FOREIGN KEY CONSTRAINTS
-- ============================================================================

-- Tenants foreign keys
ALTER TABLE identity.tenants
    ADD CONSTRAINT fk_tenants_owner
    FOREIGN KEY (owner_id)
    REFERENCES identity.users(id)
    ON DELETE RESTRICT;

-- Tenant members foreign keys
ALTER TABLE identity.tenant_members
    ADD CONSTRAINT fk_tenant_members_tenant
    FOREIGN KEY (tenant_id)
    REFERENCES identity.tenants(id)
    ON DELETE CASCADE;

ALTER TABLE identity.tenant_members
    ADD CONSTRAINT fk_tenant_members_user
    FOREIGN KEY (user_id)
    REFERENCES identity.users(id)
    ON DELETE CASCADE;

ALTER TABLE identity.tenant_members
    ADD CONSTRAINT fk_tenant_members_invited_by
    FOREIGN KEY (invited_by)
    REFERENCES identity.users(id)
    ON DELETE SET NULL;

-- OAuth2 foreign keys
ALTER TABLE identity.oauth2_authorization_codes
    ADD CONSTRAINT fk_oauth2_auth_codes_client
    FOREIGN KEY (client_id)
    REFERENCES identity.oauth2_clients(id)
    ON DELETE CASCADE;

ALTER TABLE identity.oauth2_authorization_codes
    ADD CONSTRAINT fk_oauth2_auth_codes_user
    FOREIGN KEY (user_id)
    REFERENCES identity.users(id)
    ON DELETE CASCADE;

ALTER TABLE identity.oauth2_access_tokens
    ADD CONSTRAINT fk_oauth2_access_tokens_client
    FOREIGN KEY (client_id)
    REFERENCES identity.oauth2_clients(id)
    ON DELETE CASCADE;

ALTER TABLE identity.oauth2_access_tokens
    ADD CONSTRAINT fk_oauth2_access_tokens_user
    FOREIGN KEY (user_id)
    REFERENCES identity.users(id)
    ON DELETE CASCADE;

ALTER TABLE identity.oauth2_refresh_tokens
    ADD CONSTRAINT fk_oauth2_refresh_tokens_client
    FOREIGN KEY (client_id)
    REFERENCES identity.oauth2_clients(id)
    ON DELETE CASCADE;

ALTER TABLE identity.oauth2_refresh_tokens
    ADD CONSTRAINT fk_oauth2_refresh_tokens_user
    FOREIGN KEY (user_id)
    REFERENCES identity.users(id)
    ON DELETE CASCADE;

-- Sessions foreign keys
ALTER TABLE identity.sessions
    ADD CONSTRAINT fk_sessions_user
    FOREIGN KEY (user_id)
    REFERENCES identity.users(id)
    ON DELETE CASCADE;

-- User activities foreign keys
ALTER TABLE identity.user_activities
    ADD CONSTRAINT fk_user_activities_user
    FOREIGN KEY (user_id)
    REFERENCES identity.users(id)
    ON DELETE CASCADE;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION identity.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply update trigger to tables
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON identity.users
    FOR EACH ROW
    EXECUTE FUNCTION identity.update_updated_at_column();

CREATE TRIGGER update_tenants_updated_at
    BEFORE UPDATE ON identity.tenants
    FOR EACH ROW
    EXECUTE FUNCTION identity.update_updated_at_column();

CREATE TRIGGER update_tenant_members_updated_at
    BEFORE UPDATE ON identity.tenant_members
    FOR EACH ROW
    EXECUTE FUNCTION identity.update_updated_at_column();

CREATE TRIGGER update_oauth2_clients_updated_at
    BEFORE UPDATE ON identity.oauth2_clients
    FOR EACH ROW
    EXECUTE FUNCTION identity.update_updated_at_column();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON SCHEMA identity IS 'Identity and authentication module schema';
COMMENT ON TABLE identity.users IS 'Core users table for authentication';
COMMENT ON TABLE identity.tenants IS 'Multi-tenant organizations';
COMMENT ON TABLE identity.tenant_members IS 'Users membership in tenants';
COMMENT ON TABLE identity.oauth2_clients IS 'OAuth2 registered clients';
COMMENT ON TABLE identity.oauth2_authorization_codes IS 'OAuth2 authorization codes for code flow';
COMMENT ON TABLE identity.oauth2_access_tokens IS 'OAuth2 access tokens';
COMMENT ON TABLE identity.oauth2_refresh_tokens IS 'OAuth2 refresh tokens';
COMMENT ON TABLE identity.sessions IS 'User session management';
COMMENT ON TABLE identity.user_activities IS 'Audit log for user activities';

COMMIT;
