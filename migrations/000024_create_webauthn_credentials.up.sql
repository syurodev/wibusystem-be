-- Migration: Create WebAuthn Credentials Table
-- Description: Tạo bảng lưu trữ passkey/WebAuthn credentials cho passwordless authentication
-- Author: System
-- Created: 2025-11-22

-- =====================================================
-- Table: webauthn_credentials
-- Description: Lưu trữ WebAuthn/FIDO2 credentials cho passwordless authentication
-- =====================================================
CREATE TABLE identify.webauthn_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,

    -- WebAuthn credential data
    credential_id TEXT UNIQUE NOT NULL, -- Base64URL encoded credential ID từ authenticator
    public_key BYTEA NOT NULL, -- Public key của credential (binary format)

    -- Attestation information
    attestation_type VARCHAR(50) NOT NULL DEFAULT 'none', -- none, indirect, direct
    aaguid BYTEA, -- Authenticator AAGUID (16 bytes)

    -- Security and tracking
    sign_count INTEGER NOT NULL DEFAULT 0, -- Counter để phát hiện cloned authenticators

    -- Credential metadata
    transports TEXT[], -- Supported transports: usb, nfc, ble, internal, hybrid
    backup_eligible BOOLEAN NOT NULL DEFAULT FALSE, -- Có thể backup credential không
    backup_state BOOLEAN NOT NULL DEFAULT FALSE, -- Credential đã được backup chưa

    -- User-friendly information
    credential_name VARCHAR(255), -- User-defined name (e.g., "iPhone 15 Pro", "YubiKey 5")

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ, -- Lần cuối credential được sử dụng để authentication

    -- Foreign key constraint
    CONSTRAINT fk_webauthn_credentials_user
        FOREIGN KEY (user_id)
        REFERENCES identify.users(id)
        ON DELETE CASCADE,

    -- Check constraints
    CONSTRAINT webauthn_credentials_attestation_type_check
        CHECK (attestation_type IN ('none', 'indirect', 'direct'))
);

-- Indexes
CREATE INDEX idx_webauthn_credentials_user_id ON identify.webauthn_credentials(user_id);
CREATE INDEX idx_webauthn_credentials_credential_id ON identify.webauthn_credentials(credential_id);
CREATE INDEX idx_webauthn_credentials_created_at ON identify.webauthn_credentials(created_at);
CREATE INDEX idx_webauthn_credentials_last_used_at ON identify.webauthn_credentials(last_used_at);

-- Comments
COMMENT ON TABLE identify.webauthn_credentials IS 'Bảng lưu trữ WebAuthn/FIDO2 credentials cho passwordless authentication';
COMMENT ON COLUMN identify.webauthn_credentials.credential_id IS 'Unique identifier của credential từ authenticator (Base64URL encoded)';
COMMENT ON COLUMN identify.webauthn_credentials.public_key IS 'Public key của credential trong binary format';
COMMENT ON COLUMN identify.webauthn_credentials.attestation_type IS 'Loại attestation: none (no attestation), indirect (anonymized), direct (full)';
COMMENT ON COLUMN identify.webauthn_credentials.aaguid IS 'Authenticator AAGUID (16 bytes) - identifies authenticator model';
COMMENT ON COLUMN identify.webauthn_credentials.sign_count IS 'Counter tăng dần mỗi lần authentication - dùng để phát hiện cloned authenticators';
COMMENT ON COLUMN identify.webauthn_credentials.transports IS 'Supported transports (usb, nfc, ble, internal, hybrid)';
COMMENT ON COLUMN identify.webauthn_credentials.backup_eligible IS 'Credential có thể được backup (multi-device credentials)';
COMMENT ON COLUMN identify.webauthn_credentials.backup_state IS 'Credential hiện đang được backup hay không';
COMMENT ON COLUMN identify.webauthn_credentials.credential_name IS 'User-defined name để nhận diện credential';
COMMENT ON COLUMN identify.webauthn_credentials.last_used_at IS 'Timestamp lần cuối credential được sử dụng để authentication';

-- =====================================================
-- Table: webauthn_sessions
-- Description: Lưu trữ temporary sessions cho WebAuthn registration/authentication flow
-- =====================================================
CREATE TABLE identify.webauthn_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID, -- NULL for registration of new users, set for authentication

    -- Session data
    challenge TEXT NOT NULL, -- Base64URL encoded challenge
    session_type VARCHAR(50) NOT NULL, -- registration, authentication

    -- Session metadata
    user_agent TEXT, -- Browser/device information
    ip_address VARCHAR(45), -- IPv4 or IPv6

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '5 minutes', -- Sessions expire after 5 minutes

    -- Foreign key constraint (optional, as user_id can be NULL for registration)
    CONSTRAINT fk_webauthn_sessions_user
        FOREIGN KEY (user_id)
        REFERENCES identify.users(id)
        ON DELETE CASCADE,

    -- Check constraints
    CONSTRAINT webauthn_sessions_session_type_check
        CHECK (session_type IN ('registration', 'authentication'))
);

-- Indexes
CREATE INDEX idx_webauthn_sessions_user_id ON identify.webauthn_sessions(user_id);
CREATE INDEX idx_webauthn_sessions_challenge ON identify.webauthn_sessions(challenge);
CREATE INDEX idx_webauthn_sessions_expires_at ON identify.webauthn_sessions(expires_at);
CREATE INDEX idx_webauthn_sessions_created_at ON identify.webauthn_sessions(created_at);

-- Comments
COMMENT ON TABLE identify.webauthn_sessions IS 'Bảng lưu trữ temporary sessions cho WebAuthn registration/authentication flow';
COMMENT ON COLUMN identify.webauthn_sessions.challenge IS 'Random challenge string (Base64URL encoded) dùng cho ceremony';
COMMENT ON COLUMN identify.webauthn_sessions.session_type IS 'Loại session: registration (đăng ký credential mới) hoặc authentication (xác thực)';
COMMENT ON COLUMN identify.webauthn_sessions.expires_at IS 'Session timeout (default 5 minutes)';

-- =====================================================
-- Update users table to support passwordless accounts
-- =====================================================
-- Make password_hash optional to support passwordless accounts
ALTER TABLE identify.users ALTER COLUMN password_hash DROP NOT NULL;

-- Add comment
COMMENT ON COLUMN identify.users.password_hash IS 'Password hash (NULL for passwordless accounts using only WebAuthn)';
