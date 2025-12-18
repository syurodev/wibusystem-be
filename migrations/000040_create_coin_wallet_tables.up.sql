-- Migration: Create Coin Wallet Tables
-- Description: Tạo các bảng cho hệ thống Coin Wallet, Coin Packages, và Top-up Orders
-- Author: System
-- Created: 2025-12-18

-- =====================================================
-- 1. User Coin Wallets
-- =====================================================

CREATE TABLE IF NOT EXISTS payment.user_wallets (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID UNIQUE NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,
    
    -- Coin balance (supports decimals: 0.5 coin = 500 VND)
    coin_balance DECIMAL(12,2) NOT NULL DEFAULT 0 CHECK (coin_balance >= 0),
    
    -- Stats
    total_deposited DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_spent DECIMAL(12,2) NOT NULL DEFAULT 0,
    
    -- Subscription spending
    total_subscription_spent DECIMAL(12,2) NOT NULL DEFAULT 0,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_wallets_user_id ON payment.user_wallets(user_id);

COMMENT ON TABLE payment.user_wallets IS 'Ví Coin của user';
COMMENT ON COLUMN payment.user_wallets.coin_balance IS 'Số dư Coin hiện tại (1 Coin = 1000 VND)';

-- =====================================================
-- 2. Coin Packages (Gói nạp tiền)
-- =====================================================

CREATE TABLE IF NOT EXISTS payment.coin_packages (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    
    name VARCHAR(100) NOT NULL,           -- "Gói 100 Coin"
    slug VARCHAR(50) UNIQUE NOT NULL,     -- "package-100"
    
    coin_amount DECIMAL(12,2) NOT NULL,   -- 100.00
    price_vnd DECIMAL(14,2) NOT NULL,     -- 100000.00
    bonus_percent INTEGER NOT NULL DEFAULT 0, -- 5 = 5%
    
    is_popular BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    display_order INTEGER NOT NULL DEFAULT 0,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_coin_packages_active ON payment.coin_packages(is_active, display_order);

COMMENT ON TABLE payment.coin_packages IS 'Các gói nạp Coin';

-- =====================================================
-- 3. Seed Default Coin Packages
-- =====================================================

INSERT INTO payment.coin_packages (name, slug, coin_amount, price_vnd, bonus_percent, is_popular, display_order) VALUES
('Gói 20 Coin', 'package-20', 20, 20000, 0, FALSE, 1),
('Gói 50 Coin', 'package-50', 50, 50000, 0, FALSE, 2),
('Gói 100 Coin', 'package-100', 100, 100000, 5, TRUE, 3),
('Gói 200 Coin', 'package-200', 200, 200000, 10, FALSE, 4),
('Gói 500 Coin', 'package-500', 500, 500000, 15, FALSE, 5)
ON CONFLICT (slug) DO NOTHING;

-- =====================================================
-- 4. Top-up Orders
-- =====================================================

CREATE TYPE payment.topup_status AS ENUM (
    'pending',      -- Chờ thanh toán
    'success',      -- Đã thanh toán thành công
    'expired',      -- Hết hạn
    'cancelled',    -- User hủy
    'failed'        -- Lỗi xử lý
);

CREATE TABLE IF NOT EXISTS payment.topup_orders (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,
    package_id UUID NOT NULL REFERENCES payment.coin_packages(id),
    
    -- Order info
    order_code VARCHAR(20) UNIQUE NOT NULL,  -- "NAP" + 8 random chars
    
    -- Amount
    coin_amount DECIMAL(12,2) NOT NULL,       -- Coin sẽ nhận (bao gồm bonus)
    base_coin_amount DECIMAL(12,2) NOT NULL,  -- Coin gốc từ package
    bonus_coin_amount DECIMAL(12,2) NOT NULL DEFAULT 0, -- Bonus
    vnd_amount DECIMAL(14,2) NOT NULL,        -- Số tiền VND cần chuyển
    
    -- Status
    status payment.topup_status NOT NULL DEFAULT 'pending',
    
    -- SePay info
    sepay_transaction_id VARCHAR(100),        -- ID giao dịch từ SePay (for idempotency)
    sepay_content VARCHAR(255),               -- Nội dung chuyển khoản nhận được
    
    -- Bank info snapshot (từ config lúc tạo order)
    bank_name VARCHAR(100),
    bank_account VARCHAR(50),
    account_name VARCHAR(100),
    
    -- Timestamps
    completed_at TIMESTAMPTZ,                 -- Thời điểm hoàn thành
    expired_at TIMESTAMPTZ NOT NULL,          -- Thời điểm hết hạn (15 phút sau tạo)
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_topup_orders_user_id ON payment.topup_orders(user_id);
CREATE INDEX idx_topup_orders_status ON payment.topup_orders(status);
CREATE INDEX idx_topup_orders_order_code ON payment.topup_orders(order_code);
CREATE INDEX idx_topup_orders_expired ON payment.topup_orders(expired_at) WHERE status = 'pending';
CREATE UNIQUE INDEX idx_topup_orders_sepay_tx ON payment.topup_orders(sepay_transaction_id) WHERE sepay_transaction_id IS NOT NULL;

COMMENT ON TABLE payment.topup_orders IS 'Đơn nạp tiền qua SePay';
COMMENT ON COLUMN payment.topup_orders.order_code IS 'Mã đơn = NAP + 8 ký tự random. User chuyển khoản với nội dung này';

-- =====================================================
-- 5. Transactions Log (Unified)
-- =====================================================

CREATE TYPE payment.transaction_type AS ENUM (
    'topup',              -- Nạp coin
    'purchase_chapter',   -- Mua chapter
    'purchase_series',    -- Mua series
    'rental',             -- Thuê nội dung
    'subscription',       -- Thanh toán subscription
    'refund',             -- Hoàn tiền
    'admin_adjustment'    -- Admin điều chỉnh
);

CREATE TABLE IF NOT EXISTS payment.transactions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,
    
    -- Type
    type payment.transaction_type NOT NULL,
    
    -- Amount
    coin_amount DECIMAL(12,2) NOT NULL,       -- + nạp, - tiêu
    vnd_amount DECIMAL(14,2),                 -- VND tương ứng (nullable cho subscription)
    
    -- Balance after transaction
    balance_after DECIMAL(12,2) NOT NULL,
    
    -- Reference
    reference_type VARCHAR(50),               -- 'topup_order', 'chapter', 'subscription', etc.
    reference_id UUID,                        -- ID của entity liên quan
    
    -- Creator revenue (for purchases)
    creator_id UUID REFERENCES identify.users(id),
    creator_revenue_vnd DECIMAL(14,2),        -- 80% của vnd_amount
    platform_revenue_vnd DECIMAL(14,2),       -- 20% của vnd_amount
    
    -- Metadata
    description TEXT,
    metadata JSONB,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_id ON payment.transactions(user_id);
CREATE INDEX idx_transactions_type ON payment.transactions(type);
CREATE INDEX idx_transactions_created_at ON payment.transactions(created_at);
CREATE INDEX idx_transactions_reference ON payment.transactions(reference_type, reference_id);

COMMENT ON TABLE payment.transactions IS 'Log tất cả giao dịch coin';
