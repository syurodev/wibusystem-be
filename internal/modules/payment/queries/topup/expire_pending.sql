UPDATE payment.topup_orders
SET status = 'expired', updated_at = NOW()
WHERE status = 'pending' AND expired_at < NOW()
