-- Migration: Create password reset tokens table
-- Description: Tạo bảng lưu trữ token để reset password
-- Author: System
-- Created: 2025-11-02

-- =====================================================
-- Table: password_reset_tokens
-- Description: Lưu trữ tokens để reset password
-- =====================================================
CREATE TABLE identify.password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_password_reset_tokens_user_id ON identify.password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_token ON identify.password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_expires_at ON identify.password_reset_tokens(expires_at);

-- Comment
COMMENT ON TABLE identify.password_reset_tokens IS 'Lưu trữ tokens để reset password';
COMMENT ON COLUMN identify.password_reset_tokens.token IS 'Random token gửi qua email';
COMMENT ON COLUMN identify.password_reset_tokens.expires_at IS 'Token hết hạn sau 1 giờ';
COMMENT ON COLUMN identify.password_reset_tokens.used_at IS 'NULL = chưa sử dụng, NOT NULL = đã reset thành công';

-- =====================================================
-- Cleanup Function: Xóa expired tokens
-- =====================================================
CREATE OR REPLACE FUNCTION identify.cleanup_expired_password_reset_tokens()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    WITH deleted AS (
        DELETE FROM identify.password_reset_tokens
        WHERE expires_at < NOW()
        RETURNING 1
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION identify.cleanup_expired_password_reset_tokens IS 'Cleanup expired password reset tokens. Nên chạy định kỳ (cron job).';
