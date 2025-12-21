UPDATE identify.users
SET last_login_at = NOW()
WHERE id = $1
