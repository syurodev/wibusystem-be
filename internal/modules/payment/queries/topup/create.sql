INSERT INTO payment.topup_orders (
    user_id, package_id, order_code, coin_amount, base_coin_amount, bonus_coin_amount,
    vnd_amount, bank_name, bank_account, account_name, expired_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, created_at, updated_at
