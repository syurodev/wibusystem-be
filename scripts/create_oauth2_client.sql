-- Script to create OAuth2 client
-- Usage:
--   psql -U system_dev -d system_dev -f scripts/create_oauth2_client.sql
-- Or via docker:
--   docker exec -i system_dev psql -U system_dev -d system_dev < scripts/create_oauth2_client.sql

\set client_id 'wibusystem-web-app'
\set client_name 'WibuSystem Web Application'
\set client_secret 'your-super-secret-key-change-this-in-production'

-- Generate bcrypt hash for client secret
-- Note: In production, generate this using bcrypt with proper salt
-- This is a placeholder - replace with actual bcrypt hash
\set client_secret_hash '$2a$10$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ'

INSERT INTO identity.oauth2_clients (
    id,
    client_secret_hash,
    redirect_uris,
    grant_types,
    response_types,
    scopes,
    audience,
    public,
    client_name,
    client_uri,
    logo_uri,
    token_endpoint_auth_method
) VALUES (
    :'client_id',
    :'client_secret_hash',
    ARRAY['http://localhost:3000/auth/callback', 'http://localhost:3000/callback'],
    ARRAY['authorization_code', 'refresh_token'],
    ARRAY['code'],
    ARRAY['openid', 'profile', 'email', 'offline_access'],
    ARRAY['wibusystem-api'],
    false,  -- not a public client (requires secret)
    :'client_name',
    'http://localhost:3000',
    NULL,
    'client_secret_basic'
)
ON CONFLICT (id) DO UPDATE SET
    client_name = EXCLUDED.client_name,
    redirect_uris = EXCLUDED.redirect_uris,
    grant_types = EXCLUDED.grant_types,
    updated_at = CURRENT_TIMESTAMP;

-- Display the created client
SELECT
    id,
    client_name,
    redirect_uris,
    grant_types,
    scopes,
    public,
    created_at
FROM identity.oauth2_clients
WHERE id = :'client_id';

\echo 'OAuth2 client created successfully!'
\echo 'Client ID: ' :'client_id'
\echo 'Note: Store the client secret securely!'
