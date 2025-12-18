INSERT INTO payment.transactions (
    user_id, type, coin_amount, vnd_amount, balance_after,
    reference_type, reference_id, creator_id, creator_revenue_vnd, platform_revenue_vnd,
    description, metadata
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id, created_at
