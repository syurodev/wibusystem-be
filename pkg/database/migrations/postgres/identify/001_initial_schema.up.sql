-- Identity and Tenant Service Schema for PostgreSQL
-- Based on the provided PDF documentation

-- Create enum type if not exists (idempotent for dev resets)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'auth_type'
    ) THEN
        CREATE TYPE auth_type AS ENUM (
            'password',
            'oauth',
            'oidc',
            'saml',
            'webauthn',
            'totp',
            'passkey'
        );
    END IF;
END$$;

-- Users table - Global user identities and profile info
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) UNIQUE,
    display_name VARCHAR(100),
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ
);

-- Tenants table - Organizations or teams in multi-tenant system
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Memberships table - Links users to tenants (many-to-many)
CREATE TABLE memberships (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role_id UUID NULL, -- optional default role assignment
    status VARCHAR(50) DEFAULT 'active',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, tenant_id)
);

-- Permissions table - Global list of fine-grained permissions
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    key VARCHAR(100) NOT NULL UNIQUE,
    description TEXT
);

-- Roles table - Tenant-scoped roles (collections of permissions)
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (tenant_id, name)
);

-- Role-permissions mapping (many-to-many)
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- Role assignments - Links user memberships to roles
CREATE TABLE role_assignments (
    membership_id UUID NOT NULL REFERENCES memberships(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (membership_id, role_id)
);

-- Credentials table - Stores all authentication data and identify provider links
CREATE TABLE credentials (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type auth_type NOT NULL,
    provider VARCHAR(50), -- e.g. 'google', 'github', 'okta'; NULL for password, totp, etc.
    identifier TEXT, -- external user ID or credential ID (unique per provider or credential)
    secret_hash TEXT, -- password hash OR TOTP secret, etc. (sensitive data hashed/encoded)
    public_key TEXT, -- for WebAuthn: the credential public key (PEM/Base64)
    sign_count INTEGER, -- for WebAuthn: signature counter to prevent replay
    attestation_aaguid CHAR(36), -- for WebAuthn: device AAGUID (UUID format) if needed
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    UNIQUE (provider, identifier)
);

-- Devices table - Records trusted or known devices used by users
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_info TEXT,
    last_seen_at TIMESTAMPTZ,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Sessions table - Tracks active login sessions or tokens
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID REFERENCES devices(id) ON DELETE CASCADE,
    token_hash TEXT, -- hash of session token or refresh token (if using token-based auth)
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
);

-- Create indexes for performance
CREATE INDEX idx_memberships_user_id ON memberships(user_id);
CREATE INDEX idx_memberships_tenant_id ON memberships(tenant_id);
CREATE INDEX idx_role_assignments_membership_id ON role_assignments(membership_id);
CREATE INDEX idx_credentials_user_id ON credentials(user_id);
CREATE INDEX idx_credentials_type_provider ON credentials(type, provider);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_device_id ON sessions(device_id);
CREATE INDEX idx_devices_user_id ON devices(user_id);

-- Insert default permissions
INSERT INTO permissions (key, description) VALUES
    ('CREATE_PROJECT', 'Create new projects'),
    ('EDIT_PROJECT', 'Edit existing projects'),
    ('DELETE_PROJECT', 'Delete projects'),
    ('VIEW_PROJECT', 'View projects'),
    ('MANAGE_USERS', 'Manage users in tenant'),
    ('MANAGE_ROLES', 'Manage roles and permissions'),
    ('VIEW_ANALYTICS', 'View analytics and reports'),
    ('ADMIN_ACCESS', 'Full administrative access'),
    ('CREATE_TENANT', 'Create new tenants'),
    ('MANAGE_TENANT', 'Manage tenant settings'),
    ('INVITE_USERS', 'Invite users to tenant'),
    ('REMOVE_USERS', 'Remove users from tenant');

-- Create a trigger to update last_login_at when credentials are used
CREATE OR REPLACE FUNCTION update_last_login()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE users SET last_login_at = NOW() WHERE id = NEW.user_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_last_login
    AFTER UPDATE OF last_used_at ON credentials
    FOR EACH ROW
    WHEN (OLD.last_used_at IS DISTINCT FROM NEW.last_used_at)
    EXECUTE FUNCTION update_last_login();

-- Create a trigger to update device last_seen_at when session is used
CREATE OR REPLACE FUNCTION update_device_last_seen()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.device_id IS NOT NULL THEN
        UPDATE devices SET last_seen_at = NEW.last_active_at
        WHERE id = NEW.device_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_device_last_seen
    AFTER UPDATE OF last_active_at ON sessions
    FOR EACH ROW
    WHEN (OLD.last_active_at IS DISTINCT FROM NEW.last_active_at)
    EXECUTE FUNCTION update_device_last_seen();
