-- Minimal seed for Identity database
-- Creates one admin user (Syuro) with bcrypt password

-- Enable pgcrypto for bcrypt hashing (bf)
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Insert admin user if not exists
INSERT INTO users (email, username, display_name, is_admin, created_at)
SELECT 'syuro.dev@gmail.com', 'syurodev', 'Syuro', TRUE, NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE email = 'syuro.dev@gmail.com'
);

-- Insert password credential for the admin (type=password)
INSERT INTO credentials (user_id, type, provider, identifier, secret_hash, created_at)
SELECT u.id, 'password'::auth_type, NULL, u.email, crypt('Vv19082001@#', gen_salt('bf', 12)), NOW()
FROM users u
WHERE u.email = 'syuro.dev@gmail.com'
  AND NOT EXISTS (
    SELECT 1 FROM credentials c WHERE c.user_id = u.id AND c.type = 'password'::auth_type
  );

-- Note: Additional sample tenants/roles are intentionally omitted.
