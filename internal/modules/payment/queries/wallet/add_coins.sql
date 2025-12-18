UPDATE payment.user_wallets
SET 
    coin_balance = coin_balance + $1,
    total_deposited = total_deposited + $2,
    updated_at = NOW()
WHERE user_id = $3
RETURNING id, user_id, coin_balance, total_deposited, total_spent, total_subscription_spent, created_at, updated_at
