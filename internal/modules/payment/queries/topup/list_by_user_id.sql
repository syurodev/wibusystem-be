SELECT id, user_id, package_id, order_code, coin_amount, base_coin_amount, bonus_coin_amount,
       vnd_amount, status, sepay_transaction_id, sepay_content, bank_name, bank_account, account_name,
       completed_at, expired_at, created_at, updated_at,
       COUNT(*) OVER() AS total_count
FROM payment.topup_orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
