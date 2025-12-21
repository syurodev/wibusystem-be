UPDATE identify.users
SET email = $2,
    email_verified = $3,
    password_hash = $4,
    full_name = $5,
    avatar_url = $6,
    phone = $7,
    status = $8,
    settings = $9::jsonb,
    display_name = $10,
    username = $11,
    bio = $12::jsonb,
    updated_at = NOW()
WHERE id = $1
