-- OAuth2 and OpenID Connect specific tables for Fosite

-- OAuth2 Clients table
CREATE TABLE oauth2_clients (
    id VARCHAR(255) PRIMARY KEY,
    client_secret_hash VARCHAR(255) NOT NULL,
    redirect_uris TEXT[] NOT NULL DEFAULT '{}',
    grant_types TEXT[] NOT NULL DEFAULT '{}',
    response_types TEXT[] NOT NULL DEFAULT '{}',
    scopes TEXT[] NOT NULL DEFAULT '{}',
    audience TEXT[] NOT NULL DEFAULT '{}',
    public BOOLEAN NOT NULL DEFAULT FALSE,
    client_name VARCHAR(255),
    client_uri VARCHAR(255),
    logo_uri VARCHAR(255),
    contacts TEXT[] NOT NULL DEFAULT '{}',
    tos_uri VARCHAR(255),
    policy_uri VARCHAR(255),
    jwks_uri VARCHAR(255),
    jwks TEXT,
    sector_identifier_uri VARCHAR(255),
    subject_type VARCHAR(255) DEFAULT 'public',
    token_endpoint_auth_method VARCHAR(255) DEFAULT 'client_secret_basic',
    userinfo_signed_response_alg VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- OAuth2 Authorization Codes
CREATE TABLE oauth2_authorization_codes (
    signature VARCHAR(255) PRIMARY KEY,
    request_id VARCHAR(255) NOT NULL,
    requested_at TIMESTAMPTZ NOT NULL,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth2_clients(id) ON DELETE CASCADE,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    granted_scopes TEXT[] NOT NULL DEFAULT '{}',
    form_data TEXT NOT NULL,
    session_data TEXT NOT NULL,
    subject VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    requested_audience TEXT[] NOT NULL DEFAULT '{}',
    granted_audience TEXT[] NOT NULL DEFAULT '{}',
    challenge_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- OAuth2 Access Tokens
CREATE TABLE oauth2_access_tokens (
    signature VARCHAR(255) PRIMARY KEY,
    request_id VARCHAR(255) NOT NULL,
    requested_at TIMESTAMPTZ NOT NULL,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth2_clients(id) ON DELETE CASCADE,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    granted_scopes TEXT[] NOT NULL DEFAULT '{}',
    form_data TEXT NOT NULL,
    session_data TEXT NOT NULL,
    subject VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    requested_audience TEXT[] NOT NULL DEFAULT '{}',
    granted_audience TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

-- OAuth2 Refresh Tokens
CREATE TABLE oauth2_refresh_tokens (
    signature VARCHAR(255) PRIMARY KEY,
    request_id VARCHAR(255) NOT NULL,
    requested_at TIMESTAMPTZ NOT NULL,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth2_clients(id) ON DELETE CASCADE,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    granted_scopes TEXT[] NOT NULL DEFAULT '{}',
    form_data TEXT NOT NULL,
    session_data TEXT NOT NULL,
    subject VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    requested_audience TEXT[] NOT NULL DEFAULT '{}',
    granted_audience TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- PKCE table for Proof Key for Code Exchange
CREATE TABLE oauth2_pkce (
    signature VARCHAR(255) PRIMARY KEY,
    request_id VARCHAR(255) NOT NULL,
    requested_at TIMESTAMPTZ NOT NULL,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth2_clients(id) ON DELETE CASCADE,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    granted_scopes TEXT[] NOT NULL DEFAULT '{}',
    form_data TEXT NOT NULL,
    session_data TEXT NOT NULL,
    subject VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    requested_audience TEXT[] NOT NULL DEFAULT '{}',
    granted_audience TEXT[] NOT NULL DEFAULT '{}',
    challenge_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- OpenID Connect ID Tokens
CREATE TABLE oauth2_oidc_sessions (
    signature VARCHAR(255) PRIMARY KEY,
    request_id VARCHAR(255) NOT NULL,
    requested_at TIMESTAMPTZ NOT NULL,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth2_clients(id) ON DELETE CASCADE,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    granted_scopes TEXT[] NOT NULL DEFAULT '{}',
    form_data TEXT NOT NULL,
    session_data TEXT NOT NULL,
    subject VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    requested_audience TEXT[] NOT NULL DEFAULT '{}',
    granted_audience TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Blacklisted JTI tokens (for logout)
CREATE TABLE oauth2_blacklisted_jtis (
    signature VARCHAR(255) PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for OAuth2 tables
CREATE INDEX idx_oauth2_authorization_codes_client_id ON oauth2_authorization_codes(client_id);
CREATE INDEX idx_oauth2_authorization_codes_created_at ON oauth2_authorization_codes(created_at);
CREATE INDEX idx_oauth2_access_tokens_client_id ON oauth2_access_tokens(client_id);
CREATE INDEX idx_oauth2_access_tokens_subject ON oauth2_access_tokens(subject);
CREATE INDEX idx_oauth2_access_tokens_expires_at ON oauth2_access_tokens(expires_at);
CREATE INDEX idx_oauth2_refresh_tokens_client_id ON oauth2_refresh_tokens(client_id);
CREATE INDEX idx_oauth2_refresh_tokens_subject ON oauth2_refresh_tokens(subject);
CREATE INDEX idx_oauth2_blacklisted_jtis_expires_at ON oauth2_blacklisted_jtis(expires_at);

-- Create cleanup function for expired tokens
CREATE OR REPLACE FUNCTION cleanup_expired_oauth2_tokens()
RETURNS void AS $$
BEGIN
    -- Clean up expired access tokens
    DELETE FROM oauth2_access_tokens WHERE expires_at < NOW();

    -- Clean up expired blacklisted JTIs
    DELETE FROM oauth2_blacklisted_jtis WHERE expires_at < NOW();

    -- Clean up old authorization codes (older than 10 minutes)
    DELETE FROM oauth2_authorization_codes WHERE created_at < NOW() - INTERVAL '10 minutes';

    -- Clean up old PKCE entries (older than 10 minutes)
    DELETE FROM oauth2_pkce WHERE created_at < NOW() - INTERVAL '10 minutes';
END;
$$ LANGUAGE plpgsql;

-- Create a scheduled cleanup (this would typically be done via cron or a scheduler)
-- This is just a placeholder - actual scheduling should be done externally
COMMENT ON FUNCTION cleanup_expired_oauth2_tokens() IS 'Run this function periodically to clean up expired OAuth2 tokens';

-- ------------------------------------------------------------
-- OAuth2 User Consents and DCR (IAT/RAT)
-- Consolidated in this file to simplify migrations
-- ------------------------------------------------------------

-- Create table for storing user consent per client
CREATE TABLE IF NOT EXISTS oauth2_consents (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    client_id TEXT NOT NULL,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'approved',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth2_consents_user_client
    ON oauth2_consents (user_id, client_id);

-- Initial Access Tokens for Dynamic Client Registration (RFC 7591)
CREATE TABLE IF NOT EXISTS oauth2_initial_access_tokens (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    token_hash TEXT NOT NULL,
    description TEXT,
    issued_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth2_iat_hash
    ON oauth2_initial_access_tokens (token_hash);

-- Registration Access Tokens (RAT) per client (RFC 7592)
CREATE TABLE IF NOT EXISTS oauth2_registration_access_tokens (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    client_id TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth2_rat_hash
    ON oauth2_registration_access_tokens (token_hash);
CREATE INDEX IF NOT EXISTS idx_oauth2_rat_client
    ON oauth2_registration_access_tokens (client_id);
