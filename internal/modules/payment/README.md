# Payment Module

Module xử lý thanh toán, Coin wallet, Creator subscription, và revenue management.

## Tổng Quan

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         PAYMENT MODULE                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────────────────┐   │
│  │    Coin      │    │ Subscription │    │    Creator Wallet        │   │
│  │   Wallet     │    │   Plans      │    │    (Revenue + Payout)    │   │
│  └──────────────┘    └──────────────┘    └──────────────────────────┘   │
│         │                   │                        │                   │
│         ▼                   ▼                        ▼                   │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                      Transactions Log                            │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  External: SePay (Bank Transfer QR)                                      │
└─────────────────────────────────────────────────────────────────────────┘
```

## Configuration (Database-based)

Configurations được lưu trong database, có thể thay đổi runtime qua Admin API.

### Config Table Schema

```sql
CREATE TABLE payment.configurations (
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

-- Seed default configurations
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
('subscription.grace_period_days', '3', 'number', 'Số ngày ân hạn sau khi hết hạn', FALSE);
```

### Configuration List

| Key                          | Type   | Default | Description               |
| ---------------------------- | ------ | ------- | ------------------------- |
| `sepay.api_key`              | string | -       | SePay API key (sensitive) |
| `sepay.bank_account`         | string | -       | Số tài khoản (sensitive)  |
| `sepay.bank_name`            | string | MB      | Tên ngân hàng             |
| `sepay.account_name`         | string | -       | Tên chủ tài khoản         |
| `coin.to_vnd_rate`           | number | 1000    | 1 Coin = ? VND            |
| `coin.min_chapter_price`     | number | 0.5     | Giá chapter tối thiểu     |
| `coin.min_topup`             | number | 20      | Coin tối thiểu khi nạp    |
| `revenue.creator_percent`    | number | 80      | % cho creator             |
| `revenue.platform_percent`   | number | 20      | % cho platform            |
| `revenue.hold_days`          | number | 7       | Ngày giữ pending          |
| `payout.min_amount`          | number | 100000  | Min rút tiền (VND)        |
| `subscription.reminder_days` | number | 3       | Nhắc trước N ngày         |

### Configuration APIs (Admin Only)

| Method | Endpoint                            | Description         |
| ------ | ----------------------------------- | ------------------- |
| GET    | `/api/v1/admin/payment/config`      | Lấy tất cả configs  |
| GET    | `/api/v1/admin/payment/config/:key` | Lấy config theo key |
| PUT    | `/api/v1/admin/payment/config/:key` | Cập nhật config     |
| POST   | `/api/v1/admin/payment/config`      | Tạo config mới      |

### API Examples

**Get all configs:**

```json
// GET /api/v1/admin/payment/config

{
  "configs": [
    {
      "key": "coin.to_vnd_rate",
      "value": "1000",
      "value_type": "number",
      "description": "Tỷ lệ quy đổi: 1 Coin = ? VND"
    },
    {
      "key": "sepay.api_key",
      "value": "***", // Hidden because is_sensitive = true
      "value_type": "string",
      "description": "SePay API key cho webhook validation"
    }
  ]
}
```

**Update config:**

```json
// PUT /api/v1/admin/payment/config/revenue.creator_percent

{
  "value": "85"
}

// Response
{
  "key": "revenue.creator_percent",
  "value": "85",
  "old_value": "80",
  "updated_at": "2025-12-18T18:00:00Z"
}
```

### Config Service Usage

```go
// Get config value
rate := configService.GetNumber("coin.to_vnd_rate")  // 1000
bankName := configService.GetString("sepay.bank_name")  // "MB"
processingDays := configService.GetJSON("payout.processing_days")  // ["friday"]

// With caching (Redis recommended)
rate := configService.GetNumberCached("coin.to_vnd_rate", 5*time.Minute)
```

---

## 1. Coin Wallet System

### 1.1. Quy Tắc

- **1 Coin = 1.000 VND**
- Coin hỗ trợ số lẻ: `DECIMAL(12,2)` (VD: 0.5 coin, 2.25 coin)
- Minimum: 0.5 Coin = 500 VND

### 1.2. Gói Nạp Coin

| Package | Coin | Price (VND) | Bonus |
| ------- | ---- | ----------- | ----- |
| Gói 20  | 20   | 20,000      | 0%    |
| Gói 50  | 50   | 50,000      | 0%    |
| Gói 100 | 100  | 100,000     | +5%   |
| Gói 200 | 200  | 200,000     | +10%  |
| Gói 500 | 500  | 500,000     | +15%  |

### 1.3. Top-up Flow (SePay)

```
User                    Server                  SePay                   Bank
  │                        │                       │                      │
  │  1. Chọn gói nạp       │                       │                      │
  │───────────────────────►│                       │                      │
  │                        │                       │                      │
  │  2. QR Code + order_id │                       │                      │
  │◄───────────────────────│                       │                      │
  │                        │                       │                      │
  │  3. Chuyển khoản       │                       │                      │
  │────────────────────────┼───────────────────────┼─────────────────────►│
  │                        │                       │                      │
  │                        │  4. Webhook callback  │                      │
  │                        │◄──────────────────────│                      │
  │                        │                       │                      │
  │                        │  5. Validate + Cộng coin                     │
  │                        │──────────┐            │                      │
  │                        │          │            │                      │
  │                        │◄─────────┘            │                      │
  │                        │                       │                      │
  │  6. Notify success     │                       │                      │
  │◄───────────────────────│                       │                      │
```

**Nội dung chuyển khoản format:** `WIBU <order_code>`

Ví dụ: `WIBU NAP123456`

---

## 2. Creator Subscription Plans

### 2.1. Các Gói

| Plan             | Price          | Features                                 |
| ---------------- | -------------- | ---------------------------------------- |
| **Free**         | 0 Coin         | Chỉ đăng nội dung miễn phí               |
| **Creator Lite** | 50 Coin/tháng  | Freemium (đặt giá chapter)               |
| **Creator Pro**  | 200 Coin/tháng | + Rental + Bán series + Bảo vệ bản quyền |

### 2.2. Creator Lite Features

- ✅ Đăng tác phẩm không giới hạn
- ✅ Đặt giá từng chapter (Freemium)
- ✅ 80% doanh thu từ bán chapter
- ✅ Analytics cơ bản
- ❌ Không có Rental
- ❌ Không bán toàn series

### 2.3. Creator Pro Features

- ✅ Tất cả quyền Creator Lite
- ✅ Cho thuê nội dung (Rental: 1 ngày/tuần/tháng)
- ✅ Bán toàn bộ series
- ✅ Watermark tên user trên nội dung
- ✅ Chống copy-paste text
- ✅ Ưu tiên xử lý DMCA takedown
- ✅ Badge "Verified Creator" ✓
- ✅ Analytics nâng cao

### 2.4. Subscription Logic

- **Payment**: Trừ Coin từ user wallet
- **Auto-renewal**: Mặc định BẬT, có thể toggle OFF
- **Upgrade**: Tính chênh lệch theo ngày còn lại
- **Downgrade**: Hiệu lực từ kỳ tiếp theo
- **Cancel**: Giữ quyền đến hết kỳ hiện tại

### 2.5. Upgrade Calculation

```
Ví dụ: Lite → Pro, đã dùng 15/30 ngày

Giá trị còn lại Lite: 50 × (15/30) = 25 Coin
Giá Pro: 200 Coin
Cần trả thêm: 200 - 25 = 175 Coin
```

---

## 3. Chapter Purchase

### 3.1. Revenue Split

```
Chapter Price: 5 Coin = 5,000 VND

┌─────────────────────────────────────────┐
│ User pays: 5 Coin                       │
├─────────────────────────────────────────┤
│ Creator receives: 4,000 VND (80%)       │
│ Platform receives: 1,000 VND (20%)      │
└─────────────────────────────────────────┘
```

### 3.2. Purchase Flow

```sql
BEGIN TRANSACTION;

-- 1. Lock user wallet
SELECT coin_balance FROM payment.user_wallets
WHERE user_id = ? FOR UPDATE;

-- 2. Check balance
IF coin_balance < chapter_price THEN ROLLBACK;

-- 3. Deduct coins
UPDATE payment.user_wallets
SET coin_balance = coin_balance - 5 WHERE user_id = ?;

-- 4. Add creator revenue (80% in VND)
UPDATE payment.creator_wallets
SET available_balance = available_balance + 4000
WHERE user_id = creator_id;

-- 5. Log transaction
INSERT INTO payment.transactions (...);

-- 6. Unlock chapter
INSERT INTO payment.unlocked_chapters (...);

COMMIT;
```

---

## 4. Creator Wallet & Payout

### 4.1. Balance Types

| Balance             | Description                |
| ------------------- | -------------------------- |
| `available_balance` | Có thể rút ngay            |
| `pending_balance`   | Chờ xác nhận (7 ngày hold) |
| `frozen_balance`    | Đã yêu cầu rút, đang xử lý |

### 4.2. Payout Flow

1. Creator yêu cầu rút tiền
2. Server kiểm tra điều kiện (min 100,000 VND)
3. Tiền chuyển từ `available` → `frozen`
4. Admin xem danh sách yêu cầu
5. Admin chuyển khoản thủ công
6. Admin nhập mã giao dịch ngân hàng → Confirm
7. Server update status → Email creator

---

## 5. Database Schema

### 5.1. Tables

| Table                           | Description                       |
| ------------------------------- | --------------------------------- |
| `payment.subscription_plans`    | Định nghĩa gói subscription       |
| `payment.creator_subscriptions` | Subscription hiện tại của creator |
| `payment.subscription_history`  | Lịch sử thay đổi subscription     |
| `payment.user_wallets`          | Ví Coin của user                  |
| `payment.creator_wallets`       | Ví doanh thu creator (VND)        |
| `payment.transactions`          | Log giao dịch                     |
| `payment.coin_packages`         | Gói nạp coin                      |
| `payment.unlocked_chapters`     | Chapters đã mua                   |
| `payment.payout_requests`       | Yêu cầu rút tiền                  |

### 5.2. Key Data Types

```sql
-- Coin amounts: DECIMAL(12,2) - supports 0.5, 1.25, etc.
coin_balance DECIMAL(12,2)

-- VND amounts: DECIMAL(14,2)
available_balance DECIMAL(14,2)
```

---

## 6. API Endpoints

### 6.1. Coin APIs

| Method | Endpoint                    | Description        |
| ------ | --------------------------- | ------------------ |
| GET    | `/api/v1/payment/packages`  | Danh sách gói nạp  |
| POST   | `/api/v1/payment/topup`     | Tạo đơn nạp coin   |
| GET    | `/api/v1/payment/topup/:id` | Trạng thái đơn nạp |
| POST   | `/api/webhook/sepay`        | SePay webhook      |

### 6.2. Subscription APIs

| Method | Endpoint                          | Description           |
| ------ | --------------------------------- | --------------------- |
| GET    | `/api/v1/subscription/plans`      | Danh sách gói         |
| GET    | `/api/v1/subscription/me`         | Subscription hiện tại |
| POST   | `/api/v1/subscription/subscribe`  | Đăng ký               |
| POST   | `/api/v1/subscription/upgrade`    | Nâng cấp              |
| POST   | `/api/v1/subscription/downgrade`  | Hạ cấp                |
| POST   | `/api/v1/subscription/cancel`     | Hủy                   |
| PUT    | `/api/v1/subscription/auto-renew` | Toggle auto-renew     |

### 6.3. Purchase APIs

| Method | Endpoint                      | Description      |
| ------ | ----------------------------- | ---------------- |
| POST   | `/api/v1/chapters/:id/unlock` | Mua chapter      |
| GET    | `/api/v1/chapters/:id/access` | Kiểm tra quyền   |
| GET    | `/api/v1/me/unlocked`         | Danh sách đã mua |

### 6.4. Creator Wallet APIs

| Method | Endpoint                         | Description        |
| ------ | -------------------------------- | ------------------ |
| GET    | `/api/v1/creator/wallet`         | Thông tin ví       |
| GET    | `/api/v1/creator/wallet/history` | Lịch sử doanh thu  |
| POST   | `/api/v1/creator/wallet/payout`  | Yêu cầu rút tiền   |
| PUT    | `/api/v1/creator/wallet/bank`    | Cập nhật bank info |

### 6.5. Admin APIs

| Method | Endpoint                             | Description           |
| ------ | ------------------------------------ | --------------------- |
| GET    | `/api/v1/admin/payouts`              | Danh sách yêu cầu rút |
| POST   | `/api/v1/admin/payouts/:id/complete` | Xác nhận đã chuyển    |
| POST   | `/api/v1/admin/payouts/:id/reject`   | Từ chối               |

---

## 7. Cron Jobs

| Job                           | Schedule    | Description                           |
| ----------------------------- | ----------- | ------------------------------------- |
| `ProcessSubscriptionRenewals` | Daily 00:00 | Gia hạn subscriptions hết hạn         |
| `SendExpiryReminders`         | Daily 09:00 | Email nhắc 3 ngày trước hết hạn       |
| `ExpireSubscriptions`         | Daily 00:05 | Mark expired nếu không renew được     |
| `SendLowBalanceAlerts`        | Daily 09:00 | Email cảnh báo coin không đủ          |
| `ProcessPendingRevenue`       | Daily 00:00 | Chuyển pending → available sau 7 ngày |

---

## 8. Module Structure

```
internal/modules/payment/
├── README.md                 # This file
├── handler.go                # HTTP handlers
├── routes.go                 # Route definitions
├── service.go                # Business logic
├── repository.go             # Database operations
├── service_interfaces.go     # Interface definitions
├── usecase_interfaces.go     # UseCase interfaces
│
├── usecases/
│   ├── topup.go              # Top-up coin
│   ├── subscribe.go          # Subscribe to plan
│   ├── upgrade.go            # Upgrade plan
│   ├── purchase_chapter.go   # Buy chapter
│   └── payout.go             # Creator payout
│
├── sepay/
│   ├── client.go             # SePay API client
│   └── webhook.go            # Webhook handler
│
├── cron/
│   ├── renewals.go           # Auto-renewal job
│   ├── reminders.go          # Email reminders
│   └── revenue.go            # Revenue processing
│
└── queries/
    ├── create_wallet.sql
    ├── get_subscription.sql
    └── ...
```

---

## 9. Security Checklist

- [ ] Validate SePay webhook API key
- [ ] Use `FOR UPDATE` lock trong transactions
- [ ] Idempotency key cho webhook (tránh duplicate)
- [ ] Rate limiting cho topup API
- [ ] Encrypt bank account info
- [ ] Audit log cho mọi financial operations

---

## 10. References

- [SePay API Documentation](https://docs.sepay.vn)
- [Implementation Plan](/docs/implementation_plan.md)
