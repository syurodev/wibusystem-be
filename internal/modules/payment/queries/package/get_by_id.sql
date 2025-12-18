SELECT id, name, slug, coin_amount, price_vnd, bonus_percent, is_popular, is_active, display_order, created_at, updated_at
FROM payment.coin_packages
WHERE id = $1
