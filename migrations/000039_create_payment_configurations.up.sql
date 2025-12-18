-- Migration: Create Payment Configuration Table
-- Description: Bảng lưu trữ cấu hình payment module, có thể thay đổi runtime qua Admin API
-- Author: System
-- Created: 2025-12-18

-- =====================================================
-- Create Payment Configuration Table
-- =====================================================

CREATE TABLE IF NOT EXISTS payment.configurations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    value_type VARCHAR(20) NOT NULL DEFAULT 'string',  -- string, number, boolean, json
    description TEXT,
    is_sensitive BOOLEAN NOT NULL DEFAULT FALSE,       -- Ẩn value trong API response
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES identify.users(id)
);

-- Index for fast key lookup
CREATE INDEX idx_payment_config_key ON payment.configurations(key);

-- Comment
COMMENT ON TABLE payment.configurations IS 'Cấu hình động cho payment module, có thể thay đổi qua Admin API';
COMMENT ON COLUMN payment.configurations.value_type IS 'Kiểu dữ liệu: string, number, boolean, json';
COMMENT ON COLUMN payment.configurations.is_sensitive IS 'TRUE = ẩn value trong API response (API key, bank account...)';

-- =====================================================
-- Seed Default Configurations
-- =====================================================

INSERT INTO payment.configurations (key, value, value_type, description, is_sensitive) VALUES
-- SePay Settings
('sepay.api_key', '', 'string', 'SePay API key cho webhook validation', TRUE),
('sepay.bank_account', '', 'string', 'Số tài khoản ngân hàng nhận tiền', TRUE),
('sepay.bank_name', 'MB', 'string', 'Tên ngân hàng', FALSE),
('sepay.account_name', '', 'string', 'Tên chủ tài khoản', FALSE),

-- Coin Settings
('coin.to_vnd_rate', '1000', 'number', 'Tỷ lệ quy đổi: 1 Coin = ? VND', FALSE),
('coin.min_chapter_price', '0.5', 'number', 'Giá chapter tối thiểu (Coin)', FALSE),
('coin.min_topup', '20', 'number', 'Số coin tối thiểu khi nạp', FALSE),

-- Revenue Settings  
('revenue.creator_percent', '80', 'number', '% doanh thu cho creator', FALSE),
('revenue.platform_percent', '20', 'number', '% doanh thu cho platform', FALSE),
('revenue.hold_days', '7', 'number', 'Số ngày giữ pending trước khi available', FALSE),

-- Payout Settings
('payout.min_amount', '100000', 'number', 'Số tiền tối thiểu để rút (VND)', FALSE),
('payout.processing_days', '["friday"]', 'json', 'Các ngày xử lý payout trong tuần', FALSE),

-- Subscription Settings
('subscription.reminder_days', '3', 'number', 'Nhắc trước bao nhiêu ngày hết hạn', FALSE),
('subscription.grace_period_days', '3', 'number', 'Số ngày ân hạn sau khi hết hạn', FALSE),

-- Topup Settings
('topup.expiry_minutes', '15', 'number', 'Thời gian hết hạn đơn nạp (phút)', FALSE),
('topup.order_prefix', 'NAP', 'string', 'Prefix cho mã đơn nạp', FALSE)

ON CONFLICT (key) DO NOTHING;
