-- Seed: OAuth2 Clients
-- Description: Initialize default OAuth2 clients for development
-- Dependencies: pgcrypto extension for bcrypt hashing

-- Enable pgcrypto extension for password hashing
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Insert default OAuth2 clients
-- Using crypt() function with bcrypt for password hashing

-- 1. Web Application Client (Confidential)
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
    token_endpoint_auth_method
) VALUES (
    'wibusystem-web',
    crypt('wibusystem-web-secret-dev', gen_salt('bf', 10)), -- bcrypt with cost 10
    ARRAY['http://localhost:3000/auth/callback', 'http://localhost:3000/api/auth/callback'],
    ARRAY['authorization_code', 'refresh_token'],
    ARRAY['code'],
    ARRAY['openid', 'profile', 'email', 'offline_access'],
    ARRAY['wibusystem-api'],
    false,
    'WibuSystem Web Application',
    'http://localhost:3000',
    'client_secret_basic'
) ON CONFLICT (id) DO UPDATE SET
    client_name = EXCLUDED.client_name,
    redirect_uris = EXCLUDED.redirect_uris,
    grant_types = EXCLUDED.grant_types,
    scopes = EXCLUDED.scopes,
    updated_at = CURRENT_TIMESTAMP;

-- 2. Mobile Application Client (Confidential)
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
    token_endpoint_auth_method
) VALUES (
    'wibusystem-mobile',
    crypt('wibusystem-mobile-secret-dev', gen_salt('bf', 10)),
    ARRAY['wibusystem://oauth/callback', 'http://localhost:3000/mobile/callback'],
    ARRAY['authorization_code', 'refresh_token', 'password'],
    ARRAY['code'],
    ARRAY['openid', 'profile', 'email', 'offline_access'],
    ARRAY['wibusystem-api'],
    false,
    'WibuSystem Mobile Application',
    'client_secret_basic'
) ON CONFLICT (id) DO UPDATE SET
    client_name = EXCLUDED.client_name,
    redirect_uris = EXCLUDED.redirect_uris,
    grant_types = EXCLUDED.grant_types,
    scopes = EXCLUDED.scopes,
    updated_at = CURRENT_TIMESTAMP;

-- 3. Single Page Application Client (Public - No Secret)
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
    token_endpoint_auth_method
) VALUES (
    'wibusystem-spa',
    NULL, -- Public clients don't have secrets
    ARRAY['http://localhost:3000/callback', 'http://localhost:3000/silent-renew'],
    ARRAY['authorization_code', 'refresh_token'],
    ARRAY['code'],
    ARRAY['openid', 'profile', 'email', 'offline_access'],
    ARRAY['wibusystem-api'],
    true,
    'WibuSystem Single Page Application',
    'http://localhost:3000',
    'none' -- Public clients use 'none' auth method
) ON CONFLICT (id) DO UPDATE SET
    client_name = EXCLUDED.client_name,
    redirect_uris = EXCLUDED.redirect_uris,
    grant_types = EXCLUDED.grant_types,
    scopes = EXCLUDED.scopes,
    public = EXCLUDED.public,
    updated_at = CURRENT_TIMESTAMP;

-- 4. Backend Service Client (Machine-to-Machine)
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
    token_endpoint_auth_method
) VALUES (
    'wibusystem-service',
    crypt('wibusystem-service-secret-dev', gen_salt('bf', 10)),
    ARRAY[]::text[], -- M2M doesn't need redirect URIs
    ARRAY['client_credentials'],
    ARRAY[]::text[],
    ARRAY['api:read', 'api:write', 'admin:access'],
    ARRAY['wibusystem-api'],
    false,
    'WibuSystem Backend Service',
    'client_secret_basic'
) ON CONFLICT (id) DO UPDATE SET
    client_name = EXCLUDED.client_name,
    grant_types = EXCLUDED.grant_types,
    scopes = EXCLUDED.scopes,
    updated_at = CURRENT_TIMESTAMP;

-- Display seeded clients
SELECT
    id,
    client_name,
    public,
    grant_types,
    scopes,
    CASE
        WHEN client_secret_hash IS NOT NULL THEN '***HASHED***'
        ELSE 'NULL (public client)'
    END as secret_status,
    created_at
FROM identity.oauth2_clients
ORDER BY created_at;

-- Output client credentials for development
\echo ''
\echo '╔══════════════════════════════════════════════════════════════════════╗'
\echo '║                  OAuth2 Clients Seeded Successfully                 ║'
\echo '╚══════════════════════════════════════════════════════════════════════╝'
\echo ''
\echo '📝 Development Credentials (DO NOT use in production!):'
\echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'
\echo ''
\echo '1. Web Application:'
\echo '   Client ID:     wibusystem-web'
\echo '   Client Secret: wibusystem-web-secret-dev'
\echo '   Type:          Confidential'
\echo ''
\echo '2. Mobile Application:'
\echo '   Client ID:     wibusystem-mobile'
\echo '   Client Secret: wibusystem-mobile-secret-dev'
\echo '   Type:          Confidential'
\echo ''
\echo '3. Single Page App:'
\echo '   Client ID:     wibusystem-spa'
\echo '   Client Secret: (none - public client)'
\echo '   Type:          Public'
\echo ''
\echo '4. Backend Service:'
\echo '   Client ID:     wibusystem-service'
\echo '   Client Secret: wibusystem-service-secret-dev'
\echo '   Type:          Confidential (M2M)'
\echo ''
\echo '⚠️  WARNING: These are DEVELOPMENT credentials only!'
\echo '   Change them before deploying to production.'
\echo ''
