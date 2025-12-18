INSERT INTO payment.configurations (key, value, value_type, description, is_sensitive, updated_by, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    value_type = EXCLUDED.value_type,
    description = COALESCE(EXCLUDED.description, payment.configurations.description),
    is_sensitive = EXCLUDED.is_sensitive,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW()
RETURNING id, key, value, value_type, description, is_sensitive, created_at, updated_at, updated_by;
