-- Migration: Create email verification tokens table
-- Description: Tạo bảng lưu trữ token để xác thực email khi đăng ký
-- Author: System
-- Created: 2025-11-02

-- =====================================================
-- Table: email_verification_tokens
-- Description: Lưu trữ tokens để verify email address
-- =====================================================
CREATE TABLE identify.email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_email_verification_tokens_user_id ON identify.email_verification_tokens(user_id);
CREATE INDEX idx_email_verification_tokens_token ON identify.email_verification_tokens(token);
CREATE INDEX idx_email_verification_tokens_expires_at ON identify.email_verification_tokens(expires_at);

-- Comment
COMMENT ON TABLE identify.email_verification_tokens IS 'Lưu trữ tokens để xác thực email address khi đăng ký hoặc thay đổi email';
COMMENT ON COLUMN identify.email_verification_tokens.token IS 'Random token gửi qua email (hashed hoặc plain - tùy implementation)';
COMMENT ON COLUMN identify.email_verification_tokens.expires_at IS 'Token hết hạn sau 24 giờ';
COMMENT ON COLUMN identify.email_verification_tokens.used_at IS 'NULL = chưa sử dụng, NOT NULL = đã verify thành công';

-- =====================================================
-- Cleanup Function: Xóa expired tokens
-- =====================================================
CREATE OR REPLACE FUNCTION identify.cleanup_expired_verification_tokens()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    WITH deleted AS (
        DELETE FROM identify.email_verification_tokens
        WHERE expires_at < NOW()
        RETURNING 1
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION identify.cleanup_expired_verification_tokens IS 'Cleanup expired verification tokens. Nên chạy định kỳ (cron job).';
