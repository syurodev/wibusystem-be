SELECT id, user_id, coin_balance, total_deposited, total_spent, total_subscription_spent, created_at, updated_at
FROM payment.user_wallets
WHERE user_id = $1
FOR UPDATE
