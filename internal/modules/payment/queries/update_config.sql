UPDATE payment.configurations
SET value = $1, updated_by = $2, updated_at = NOW()
WHERE key = $3
