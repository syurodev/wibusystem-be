-- Migration: Create OAuth2 Tables
-- Description: Tạo các bảng cho OAuth 2.0 Authorization Server (Fosite-compatible)
-- Author: System
-- Created: 2025-10-26

-- =====================================================
-- Table: oauth2_clients
-- Description: Lưu trữ thông tin về OAuth 2.0 clients đã đăng ký
-- =====================================================
CREATE TABLE oauth2_clients (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    client_name VARCHAR(255) NOT NULL,

    -- Secret (hashed)
    secret_hash VARCHAR(255) NOT NULL,

    -- URIs và configurations (stored as arrays)
    redirect_uris TEXT[] NOT NULL DEFAULT '{}',
    grant_types TEXT[] NOT NULL DEFAULT '{}',
    response_types TEXT[] NOT NULL DEFAULT '{}',
    scopes TEXT[] NOT NULL DEFAULT '{}',

    -- Client type
    is_public BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE for public clients (mobile, SPA)

    -- Multi-organization support
    -- NULL = global/first-party client (e.g., admin dashboard)
    -- NOT NULL = organization-specific client
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,

    -- Metadata
    owner_user_id UUID REFERENCES users(id) ON DELETE SET NULL, -- User who created this client
    logo_url TEXT,
    terms_of_service_url TEXT,
    policy_url TEXT,
    client_uri TEXT, -- Homepage của client application

    -- Security settings
    token_endpoint_auth_method VARCHAR(50) NOT NULL DEFAULT 'client_secret_basic',
    -- Options: client_secret_basic, client_secret_post, none (for public clients)

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT oauth2_clients_auth_method_check CHECK (
        token_endpoint_auth_method IN ('client_secret_basic', 'client_secret_post', 'client_secret_jwt', 'private_key_jwt', 'none')
    ),
    CONSTRAINT oauth2_clients_public_check CHECK (
        (is_public = TRUE AND token_endpoint_auth_method = 'none') OR
        (is_public = FALSE)
    )
);

-- Indexes
CREATE INDEX idx_oauth2_clients_organization_id ON oauth2_clients(organization_id);
CREATE INDEX idx_oauth2_clients_owner_user_id ON oauth2_clients(owner_user_id);
CREATE INDEX idx_oauth2_clients_is_public ON oauth2_clients(is_public);

-- Trigger for updated_at
CREATE TRIGGER update_oauth2_clients_updated_at BEFORE UPDATE ON oauth2_clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comment
COMMENT ON TABLE oauth2_clients IS 'Lưu trữ OAuth 2.0 clients đã đăng ký - hỗ trợ cả global và organization-specific clients';
COMMENT ON COLUMN oauth2_clients.organization_id IS 'NULL = global client (first-party), NOT NULL = organization-specific client (third-party)';
COMMENT ON COLUMN oauth2_clients.is_public IS 'TRUE cho public clients (mobile apps, SPAs) không có secret';
COMMENT ON COLUMN oauth2_clients.secret_hash IS 'Hashed client secret (bcrypt hoặc argon2)';
COMMENT ON COLUMN oauth2_clients.redirect_uris IS 'Danh sách redirect URIs được phép (OAuth 2.0 callback)';
COMMENT ON COLUMN oauth2_clients.grant_types IS 'Grant types được phép: authorization_code, refresh_token, client_credentials, etc.';

-- =====================================================
-- Type: session_type
-- Description: Các loại session/token được lưu trữ
-- =====================================================
CREATE TYPE session_type AS ENUM (
    'authorize_code',    -- Authorization codes
    'access_token',      -- Access tokens
    'refresh_token',     -- Refresh tokens
    'pkce',             -- PKCE sessions
    'openid'            -- OpenID Connect sessions
);

COMMENT ON TYPE session_type IS 'Loại session/token trong OAuth 2.0 flow';

-- =====================================================
-- Table: oauth2_sessions
-- Description: Bảng chung lưu trữ tất cả OAuth 2.0 sessions và tokens
-- =====================================================
CREATE TABLE oauth2_sessions (
    -- Primary identifier
    signature VARCHAR(255) PRIMARY KEY,

    -- Session metadata
    request_id VARCHAR(255) NOT NULL,
    session_type session_type NOT NULL,

    -- Status
    active BOOLEAN NOT NULL DEFAULT TRUE,

    -- Session data (Fosite Requester serialized to JSON)
    session_data JSONB NOT NULL,

    -- Expiration
    expires_at TIMESTAMPTZ NOT NULL,

    -- References
    client_id UUID NOT NULL REFERENCES oauth2_clients(id) ON DELETE CASCADE,
    subject_id UUID REFERENCES users(id) ON DELETE CASCADE, -- NULL for client_credentials grant

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_oauth2_sessions_request_id ON oauth2_sessions(request_id);
CREATE INDEX idx_oauth2_sessions_session_type ON oauth2_sessions(session_type);
CREATE INDEX idx_oauth2_sessions_expires_at ON oauth2_sessions(expires_at);
CREATE INDEX idx_oauth2_sessions_client_id ON oauth2_sessions(client_id);
CREATE INDEX idx_oauth2_sessions_subject_id ON oauth2_sessions(subject_id);
CREATE INDEX idx_oauth2_sessions_active ON oauth2_sessions(active);

-- Index cho cleanup expired sessions
CREATE INDEX idx_oauth2_sessions_cleanup ON oauth2_sessions(expires_at, active);

-- GIN index cho JSONB queries (nếu cần query vào session_data)
CREATE INDEX idx_oauth2_sessions_session_data ON oauth2_sessions USING GIN (session_data);

-- Comment
COMMENT ON TABLE oauth2_sessions IS 'Bảng chung lưu trữ tất cả OAuth 2.0 sessions, codes, và tokens';
COMMENT ON COLUMN oauth2_sessions.signature IS 'Unique signature của token/code (thường là hash)';
COMMENT ON COLUMN oauth2_sessions.request_id IS 'ID của authorization request';
COMMENT ON COLUMN oauth2_sessions.session_data IS 'Fosite Requester object serialized thành JSON';
COMMENT ON COLUMN oauth2_sessions.subject_id IS 'User ID (NULL cho client_credentials grant)';

-- =====================================================
-- Table: oauth2_jti_blacklist
-- Description: Blacklist cho revoked tokens (JWT Token Identifier)
-- =====================================================
CREATE TABLE oauth2_jti_blacklist (
    signature VARCHAR(255) PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index
CREATE INDEX idx_oauth2_jti_blacklist_expires_at ON oauth2_jti_blacklist(expires_at);

-- Comment
COMMENT ON TABLE oauth2_jti_blacklist IS 'Blacklist cho revoked tokens - ngăn chặn replay attacks';
COMMENT ON COLUMN oauth2_jti_blacklist.signature IS 'Token signature đã bị revoke';
COMMENT ON COLUMN oauth2_jti_blacklist.expires_at IS 'Thời điểm token hết hạn (có thể xóa khỏi blacklist sau đó)';

-- =====================================================
-- Cleanup Function: Xóa expired sessions
-- Description: Function để cleanup sessions và blacklist entries đã hết hạn
-- =====================================================
CREATE OR REPLACE FUNCTION cleanup_expired_oauth2_data()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete expired sessions
    WITH deleted AS (
        DELETE FROM oauth2_sessions
        WHERE expires_at < NOW()
        RETURNING 1
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;

    -- Delete expired blacklist entries
    DELETE FROM oauth2_jti_blacklist
    WHERE expires_at < NOW();

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_expired_oauth2_data IS 'Cleanup expired OAuth2 sessions và blacklist entries. Nên chạy định kỳ (cron job).';

-- =====================================================
-- Helper Functions: OAuth2 Operations
-- =====================================================

-- Function: Kiểm tra token có bị revoked không
CREATE OR REPLACE FUNCTION is_token_revoked(
    p_signature VARCHAR
)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM oauth2_jti_blacklist
        WHERE signature = p_signature
            AND expires_at > NOW()
    );
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION is_token_revoked IS 'Kiểm tra xem token có trong blacklist không';

-- Function: Revoke token
CREATE OR REPLACE FUNCTION revoke_token(
    p_signature VARCHAR,
    p_expires_at TIMESTAMPTZ
)
RETURNS VOID AS $$
BEGIN
    -- Add to blacklist
    INSERT INTO oauth2_jti_blacklist (signature, expires_at)
    VALUES (p_signature, p_expires_at)
    ON CONFLICT (signature) DO NOTHING;

    -- Mark session as inactive
    UPDATE oauth2_sessions
    SET active = FALSE
    WHERE signature = p_signature;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION revoke_token IS 'Revoke một token bằng cách thêm vào blacklist và mark session inactive';

-- Function: Revoke all tokens của một user trong một client
CREATE OR REPLACE FUNCTION revoke_user_client_tokens(
    p_user_id UUID,
    p_client_id UUID
)
RETURNS INTEGER AS $$
DECLARE
    revoked_count INTEGER;
BEGIN
    -- Mark sessions as inactive
    WITH updated AS (
        UPDATE oauth2_sessions
        SET active = FALSE
        WHERE subject_id = p_user_id
            AND client_id = p_client_id
            AND active = TRUE
        RETURNING signature, expires_at
    )
    -- Add to blacklist
    INSERT INTO oauth2_jti_blacklist (signature, expires_at)
    SELECT signature, expires_at FROM updated
    ON CONFLICT (signature) DO NOTHING;

    GET DIAGNOSTICS revoked_count = ROW_COUNT;
    RETURN revoked_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION revoke_user_client_tokens IS 'Revoke tất cả tokens của một user cho một client cụ thể';

-- Function: Revoke all tokens của một user (logout từ tất cả clients)
CREATE OR REPLACE FUNCTION revoke_all_user_tokens(
    p_user_id UUID
)
RETURNS INTEGER AS $$
DECLARE
    revoked_count INTEGER;
BEGIN
    WITH updated AS (
        UPDATE oauth2_sessions
        SET active = FALSE
        WHERE subject_id = p_user_id
            AND active = TRUE
        RETURNING signature, expires_at
    )
    INSERT INTO oauth2_jti_blacklist (signature, expires_at)
    SELECT signature, expires_at FROM updated
    ON CONFLICT (signature) DO NOTHING;

    GET DIAGNOSTICS revoked_count = ROW_COUNT;
    RETURN revoked_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION revoke_all_user_tokens IS 'Revoke tất cả tokens của một user (global logout)';
