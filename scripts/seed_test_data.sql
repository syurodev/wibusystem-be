-- Script: Seed Test Data for OAuth2 Testing
-- Description: Tạo test users và OAuth2 clients để test OAuth2 flow
-- Author: System
-- Created: 2025-11-02

-- =====================================================
-- Seed Test Users
-- =====================================================

-- Test User 1: john@example.com / password123
-- Password hash generated with bcrypt cost 12
INSERT INTO identify.users (
    id,
    email,
    email_verified,
    password_hash,
    full_name,
    avatar_url,
    status
) VALUES (
    '00000000-0000-0000-0000-000000000001'::uuid,
    'john@example.com',
    true,
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIRe8Jvjm6', -- password123
    'John Doe',
    'https://i.pravatar.cc/150?img=1',
    'active'
) ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    full_name = EXCLUDED.full_name;

-- Test User 2: jane@example.com / password456
INSERT INTO identify.users (
    id,
    email,
    email_verified,
    password_hash,
    full_name,
    avatar_url,
    status
) VALUES (
    '00000000-0000-0000-0000-000000000002'::uuid,
    'jane@example.com',
    true,
    '$2a$12$rZ5c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIRe8Jvjm6', -- password456
    'Jane Smith',
    'https://i.pravatar.cc/150?img=2',
    'active'
) ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    full_name = EXCLUDED.full_name;

-- Test User 3: test@example.com / test123
INSERT INTO identify.users (
    id,
    email,
    email_verified,
    password_hash,
    full_name,
    avatar_url,
    status
) VALUES (
    '00000000-0000-0000-0000-000000000003'::uuid,
    'test@example.com',
    true,
    '$2a$12$sZ5c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIRe8Jvjm6', -- test123
    'Test User',
    'https://i.pravatar.cc/150?img=3',
    'active'
) ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    full_name = EXCLUDED.full_name;

-- =====================================================
-- Seed Test OAuth2 Clients
-- =====================================================

-- Test Client 1: Web Application (Confidential Client)
INSERT INTO identify.oauth2_clients (
    id,
    client_name,
    secret_hash,
    redirect_uris,
    grant_types,
    response_types,
    scopes,
    is_public,
    tenant_id,
    token_endpoint_auth_method,
    client_uri,
    logo_url
) VALUES (
    '10000000-0000-0000-0000-000000000001'::uuid,
    'Test Web Application',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIRe8Jvjm6', -- secret: test-client-secret
    ARRAY[
        'http://localhost:3000/callback',
        'http://localhost:3000/auth/callback',
        'https://oauth.pstmn.io/v1/callback'
    ],
    ARRAY['authorization_code', 'refresh_token'],
    ARRAY['code'],
    ARRAY['openid', 'profile', 'email', 'offline_access'],
    false, -- Confidential client
    NULL,  -- Global client
    'client_secret_basic',
    'http://localhost:3000',
    'https://via.placeholder.com/150/0000FF/FFFFFF?text=Web'
) ON CONFLICT (id) DO UPDATE SET
    client_name = EXCLUDED.client_name,
    redirect_uris = EXCLUDED.redirect_uris,
    scopes = EXCLUDED.scopes;

-- Test Client 2: Mobile App (Public Client with PKCE)
INSERT INTO identify.oauth2_clients (
    id,
    client_name,
    secret_hash,
    redirect_uris,
    grant_types,
    response_types,
    scopes,
    is_public,
    tenant_id,
    token_endpoint_auth_method,
    client_uri,
    logo_url
) VALUES (
    '10000000-0000-0000-0000-000000000002'::uuid,
    'Test Mobile App',
    '', -- Public client, no secret
    ARRAY[
        'myapp://callback',
        'http://localhost:8080/callback'
    ],
    ARRAY['authorization_code', 'refresh_token'],
    ARRAY['code'],
    ARRAY['openid', 'profile', 'email', 'offline_access'],
    true, -- Public client
    NULL,
    'none', -- No client authentication
    'myapp://home',
    'https://via.placeholder.com/150/00FF00/FFFFFF?text=Mobile'
) ON CONFLICT (id) DO UPDATE SET
    client_name = EXCLUDED.client_name,
    redirect_uris = EXCLUDED.redirect_uris,
    scopes = EXCLUDED.scopes;

-- Test Client 3: Single Page Application (Public Client)
INSERT INTO identify.oauth2_clients (
    id,
    client_name,
    secret_hash,
    redirect_uris,
    grant_types,
    response_types,
    scopes,
    is_public,
    tenant_id,
    token_endpoint_auth_method,
    client_uri,
    logo_url
) VALUES (
    '10000000-0000-0000-0000-000000000003'::uuid,
    'Test SPA Application',
    '',
    ARRAY[
        'http://localhost:5173/callback',
        'http://localhost:5173/auth/callback'
    ],
    ARRAY['authorization_code', 'refresh_token'],
    ARRAY['code'],
    ARRAY['openid', 'profile', 'email'],
    true,
    NULL,
    'none',
    'http://localhost:5173',
    'https://via.placeholder.com/150/FF0000/FFFFFF?text=SPA'
) ON CONFLICT (id) DO UPDATE SET
    client_name = EXCLUDED.client_name,
    redirect_uris = EXCLUDED.redirect_uris,
    scopes = EXCLUDED.scopes;

-- Test Client 4: Server-to-Server (Client Credentials)
INSERT INTO identify.oauth2_clients (
    id,
    client_name,
    secret_hash,
    redirect_uris,
    grant_types,
    response_types,
    scopes,
    is_public,
    tenant_id,
    token_endpoint_auth_method,
    client_uri,
    logo_url
) VALUES (
    '10000000-0000-0000-0000-000000000004'::uuid,
    'Test API Service',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIRe8Jvjm6',
    ARRAY[]::text[], -- No redirect URIs for client_credentials
    ARRAY['client_credentials'],
    ARRAY[],
    ARRAY['api:read', 'api:write'],
    false,
    NULL,
    'client_secret_basic',
    NULL,
    'https://via.placeholder.com/150/FFA500/FFFFFF?text=API'
) ON CONFLICT (id) DO UPDATE SET
    client_name = EXCLUDED.client_name,
    grant_types = EXCLUDED.grant_types;

-- =====================================================
-- Display Seed Data Summary
-- =====================================================

SELECT
    '✅ Test Users Created' as status,
    COUNT(*) as count
FROM identify.users
WHERE email LIKE '%@example.com';

SELECT
    '✅ Test Clients Created' as status,
    COUNT(*) as count
FROM identify.oauth2_clients
WHERE client_name LIKE 'Test%';

-- Display Test Credentials
SELECT '
========================================
🔐 TEST CREDENTIALS
========================================

📧 Test Users:
--------------
Email: john@example.com
Password: password123
User ID: 00000000-0000-0000-0000-000000000001

Email: jane@example.com
Password: password456
User ID: 00000000-0000-0000-0000-000000000002

Email: test@example.com
Password: test123
User ID: 00000000-0000-0000-0000-000000000003

🔑 Test OAuth2 Clients:
-----------------------
1. Web Application (Confidential)
   Client ID: 10000000-0000-0000-0000-000000000001
   Client Secret: test-client-secret
   Redirect URI: http://localhost:3000/callback

2. Mobile App (Public + PKCE)
   Client ID: 10000000-0000-0000-0000-000000000002
   No Secret (public client)
   Redirect URI: myapp://callback

3. SPA Application (Public)
   Client ID: 10000000-0000-0000-0000-000000000003
   No Secret (public client)
   Redirect URI: http://localhost:5173/callback

4. API Service (Client Credentials)
   Client ID: 10000000-0000-0000-0000-000000000004
   Client Secret: test-client-secret
   Grant Type: client_credentials

========================================
' as "Test Credentials";
