SELECT id, user_id, type, coin_amount, vnd_amount, balance_after,
       reference_type, reference_id, creator_id, creator_revenue_vnd, platform_revenue_vnd,
       description, metadata, created_at,
       COUNT(*) OVER() AS total_count
FROM payment.transactions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
