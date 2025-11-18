-- Create test OAuth2 client for development
-- Client ID: 20000000-0000-0000-0000-000000000001
-- Client Secret: test-secret (hashed)

-- 1. Insert client
INSERT INTO catalog.oauth2_clients (
    id,
    name,
    client_secret_hash,
    is_public,
    is_active,
    created_at,
    updated_at
) VALUES (
    '20000000-0000-0000-0000-000000000001'::uuid,
    'Test Development Client',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', -- bcrypt hash of 'test-secret'
    false,
    true,
    NOW(),
    NOW()
) ON CONFLICT (id) DO UPDATE SET
    is_active = true,
    updated_at = NOW();

-- 2. Insert redirect URIs
INSERT INTO catalog.oauth2_client_redirect_uris (
    client_id,
    redirect_uri
) VALUES
    ('20000000-0000-0000-0000-000000000001'::uuid, 'http://localhost:3000/api/auth/callback'),
    ('20000000-0000-0000-0000-000000000001'::uuid, 'http://localhost:3000/callback'),
    ('20000000-0000-0000-0000-000000000001'::uuid, 'http://127.0.0.1:3000/api/auth/callback')
ON CONFLICT (client_id, redirect_uri) DO NOTHING;

-- 3. Insert grant types
INSERT INTO catalog.oauth2_client_grant_types (
    client_id,
    grant_type
) VALUES
    ('20000000-0000-0000-0000-000000000001'::uuid, 'authorization_code'),
    ('20000000-0000-0000-0000-000000000001'::uuid, 'refresh_token')
ON CONFLICT (client_id, grant_type) DO NOTHING;

-- 4. Insert scopes
INSERT INTO catalog.oauth2_client_scopes (
    client_id,
    scope
) VALUES
    ('20000000-0000-0000-0000-000000000001'::uuid, 'openid'),
    ('20000000-0000-0000-0000-000000000001'::uuid, 'profile'),
    ('20000000-0000-0000-0000-000000000001'::uuid, 'email'),
    ('20000000-0000-0000-0000-000000000001'::uuid, 'offline_access')
ON CONFLICT (client_id, scope) DO NOTHING;

-- Verify
SELECT
    c.id,
    c.name,
    c.is_active,
    array_agg(DISTINCT r.redirect_uri) as redirect_uris,
    array_agg(DISTINCT s.scope) as scopes,
    array_agg(DISTINCT g.grant_type) as grant_types
FROM catalog.oauth2_clients c
LEFT JOIN catalog.oauth2_client_redirect_uris r ON c.id = r.client_id
LEFT JOIN catalog.oauth2_client_scopes s ON c.id = s.client_id
LEFT JOIN catalog.oauth2_client_grant_types g ON c.id = g.client_id
WHERE c.id = '20000000-0000-0000-0000-000000000001'::uuid
GROUP BY c.id, c.name, c.is_active;
