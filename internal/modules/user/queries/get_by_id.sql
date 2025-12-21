SELECT id, email, email_verified, password_hash, full_name, avatar_url,
       phone, status, created_at, updated_at, last_login_at, settings,
       display_name, username, bio, is_verified
FROM identify.users
WHERE id = $1 AND status != 'deleted'
