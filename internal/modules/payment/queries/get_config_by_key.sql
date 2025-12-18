SELECT
    id, key, value, value_type, description, is_sensitive,
    created_at, updated_at, updated_by
FROM payment.configurations
WHERE key = $1
