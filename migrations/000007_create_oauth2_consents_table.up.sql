-- Migration: Create OAuth2 Consents Table
-- Description: Tạo bảng lưu trữ user consents cho OAuth2 clients
-- Author: System
-- Created: 2025-11-01

-- =====================================================
-- Table: oauth2_consents
-- Description: Lưu trữ thông tin về quyền mà user đã cấp cho các OAuth2 clients
-- =====================================================
CREATE TABLE identify.oauth2_consents (
    id UUID PRIMARY KEY DEFAULT uuidv7(),

    -- User và Client
    user_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES identify.oauth2_clients(id) ON DELETE CASCADE,

    -- Scopes được cấp phép
    granted_scopes TEXT[] NOT NULL DEFAULT '{}',

    -- Status
    revoked BOOLEAN NOT NULL DEFAULT FALSE,

    -- Timestamps
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ, -- NULL = không bao giờ hết hạn (persistent consent)

    -- Metadata
    consent_method VARCHAR(50) NOT NULL DEFAULT 'explicit', -- explicit, implicit, remembered
    ip_address INET,
    user_agent TEXT,

    -- Constraints
    CONSTRAINT oauth2_consents_user_client_unique UNIQUE (user_id, client_id),
    CONSTRAINT oauth2_consents_method_check CHECK (
        consent_method IN ('explicit', 'implicit', 'remembered')
    )
);

-- Indexes
CREATE INDEX idx_oauth2_consents_user_id ON identify.oauth2_consents(user_id);
CREATE INDEX idx_oauth2_consents_client_id ON identify.oauth2_consents(client_id);
CREATE INDEX idx_oauth2_consents_revoked ON identify.oauth2_consents(revoked);
CREATE INDEX idx_oauth2_consents_granted_at ON identify.oauth2_consents(granted_at);
CREATE INDEX idx_oauth2_consents_expires_at ON identify.oauth2_consents(expires_at);

-- Index cho cleanup expired consents
CREATE INDEX idx_oauth2_consents_cleanup ON identify.oauth2_consents(expires_at, revoked)
WHERE expires_at IS NOT NULL;

-- Trigger for updated_at (nếu cần track updates)
-- Note: Bảng này không có updated_at vì consent là immutable - chỉ có thể revoke

-- Comments
COMMENT ON TABLE identify.oauth2_consents IS 'Lưu trữ user consents cho OAuth2 clients - quản lý quyền truy cập';
COMMENT ON COLUMN identify.oauth2_consents.user_id IS 'User đã cấp quyền';
COMMENT ON COLUMN identify.oauth2_consents.client_id IS 'Client được cấp quyền';
COMMENT ON COLUMN identify.oauth2_consents.granted_scopes IS 'Danh sách scopes đã được user chấp thuận';
COMMENT ON COLUMN identify.oauth2_consents.revoked IS 'TRUE nếu consent đã bị thu hồi';
COMMENT ON COLUMN identify.oauth2_consents.consent_method IS 'Phương thức consent: explicit (user clicked allow), implicit (trusted first-party), remembered (previous consent)';
COMMENT ON COLUMN identify.oauth2_consents.expires_at IS 'Thời điểm consent hết hạn (NULL = persistent)';

-- =====================================================
-- Functions: Consent Management
-- =====================================================

-- Function: Cleanup expired consents
CREATE OR REPLACE FUNCTION identify.cleanup_expired_consents()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    WITH deleted AS (
        DELETE FROM identify.oauth2_consents
        WHERE expires_at IS NOT NULL
            AND expires_at < NOW()
            AND revoked = FALSE
        RETURNING 1
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION identify.cleanup_expired_consents IS 'Xóa các consents đã hết hạn - nên chạy định kỳ';

-- Function: Revoke consent
CREATE OR REPLACE FUNCTION identify.revoke_consent(
    p_user_id UUID,
    p_client_id UUID
)
RETURNS BOOLEAN AS $$
DECLARE
    rows_affected INTEGER;
BEGIN
    UPDATE identify.oauth2_consents
    SET revoked = TRUE,
        revoked_at = NOW()
    WHERE user_id = p_user_id
        AND client_id = p_client_id
        AND revoked = FALSE;

    GET DIAGNOSTICS rows_affected = ROW_COUNT;
    RETURN rows_affected > 0;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION identify.revoke_consent IS 'Thu hồi consent của user cho một client';

-- Function: Revoke all consents for a user
CREATE OR REPLACE FUNCTION identify.revoke_all_user_consents(
    p_user_id UUID
)
RETURNS INTEGER AS $$
DECLARE
    rows_affected INTEGER;
BEGIN
    UPDATE identify.oauth2_consents
    SET revoked = TRUE,
        revoked_at = NOW()
    WHERE user_id = p_user_id
        AND revoked = FALSE;

    GET DIAGNOSTICS rows_affected = ROW_COUNT;
    RETURN rows_affected;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION identify.revoke_all_user_consents IS 'Thu hồi tất cả consents của một user';

-- Function: Get active consent
CREATE OR REPLACE FUNCTION identify.get_active_consent(
    p_user_id UUID,
    p_client_id UUID
)
RETURNS TABLE (
    id UUID,
    granted_scopes TEXT[],
    granted_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        c.id,
        c.granted_scopes,
        c.granted_at,
        c.expires_at
    FROM identify.oauth2_consents c
    WHERE c.user_id = p_user_id
        AND c.client_id = p_client_id
        AND c.revoked = FALSE
        AND (c.expires_at IS NULL OR c.expires_at > NOW());
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION identify.get_active_consent IS 'Lấy active consent của user cho một client';

-- =====================================================
-- Seed Data: Demo Consent (Optional)
-- =====================================================

-- Note: Không seed consents vì chúng nên được tạo thông qua OAuth2 flow
