UPDATE payment.topup_orders
SET status = $1,
    sepay_transaction_id = COALESCE($2, sepay_transaction_id),
    sepay_content = COALESCE($3, sepay_content),
    completed_at = $4,
    updated_at = NOW()
WHERE id = $5
