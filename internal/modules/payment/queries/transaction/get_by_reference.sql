SELECT id, user_id, type, coin_amount, vnd_amount, balance_after,
       reference_type, reference_id, creator_id, creator_revenue_vnd, platform_revenue_vnd,
       description, metadata, created_at
FROM payment.transactions
WHERE reference_type = $1 AND reference_id = $2
LIMIT 1
