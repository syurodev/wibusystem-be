INSERT INTO identify.users (
    id, email, email_verified, password_hash, full_name,
    avatar_url, phone, status, settings
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)
