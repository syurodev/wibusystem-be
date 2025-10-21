-- Seed: Users and Tenants
-- Description: Initialize default users and tenants for development
-- Dependencies: 001_oauth2_clients.sql (pgcrypto extension)

-- Ensure pgcrypto is enabled
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Insert demo users
-- Password: 'password123' (hashed with bcrypt)

-- 1. Admin User
INSERT INTO identity.users (
    id,
    email,
    email_verified,
    password_hash,
    display_name,
    avatar_url,
    status
) VALUES (
    gen_random_uuid(),
    'admin@wibusystem.dev',
    true,
    crypt('password123', gen_salt('bf', 10)),
    'System Administrator',
    'https://ui-avatars.com/api/?name=Admin&background=0D8ABC&color=fff',
    'active'
) ON CONFLICT (email) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    status = EXCLUDED.status,
    updated_at = CURRENT_TIMESTAMP
RETURNING id, email, display_name;

-- Store admin user ID for later use
DO $$
DECLARE
    admin_user_id UUID;
BEGIN
    -- Get admin user ID
    SELECT id INTO admin_user_id
    FROM identity.users
    WHERE email = 'admin@wibusystem.dev';

    -- Create admin's tenant if not exists
    INSERT INTO identity.tenants (
        id,
        name,
        slug,
        description,
        logo_url,
        status,
        owner_id,
        settings,
        metadata
    ) VALUES (
        gen_random_uuid(),
        'WibuSystem HQ',
        'wibusystem-hq',
        'Official WibuSystem Headquarters',
        'https://ui-avatars.com/api/?name=WibuSystem&background=6366f1&color=fff',
        'active',
        admin_user_id,
        jsonb_build_object(
            'features', jsonb_build_object(
                'api_access', true,
                'advanced_analytics', true,
                'custom_branding', true
            ),
            'limits', jsonb_build_object(
                'max_members', 100,
                'max_projects', 50,
                'storage_gb', 100
            )
        ),
        jsonb_build_object(
            'industry', 'Technology',
            'company_size', 'startup'
        )
    ) ON CONFLICT (slug) DO UPDATE SET
        name = EXCLUDED.name,
        description = EXCLUDED.description,
        updated_at = CURRENT_TIMESTAMP;

    -- Add admin as tenant member
    INSERT INTO identity.tenant_members (
        tenant_id,
        user_id,
        role,
        permissions
    ) SELECT
        t.id,
        admin_user_id,
        'owner',
        jsonb_build_array(
            'tenant:read', 'tenant:write', 'tenant:delete',
            'member:read', 'member:write', 'member:delete',
            'settings:read', 'settings:write'
        )
    FROM identity.tenants t
    WHERE t.slug = 'wibusystem-hq'
    ON CONFLICT (tenant_id, user_id) DO UPDATE SET
        role = EXCLUDED.role,
        permissions = EXCLUDED.permissions,
        updated_at = CURRENT_TIMESTAMP;
END $$;

-- 2. Regular User 1
INSERT INTO identity.users (
    id,
    email,
    email_verified,
    password_hash,
    display_name,
    avatar_url,
    status
) VALUES (
    gen_random_uuid(),
    'user1@wibusystem.dev',
    true,
    crypt('password123', gen_salt('bf', 10)),
    'John Doe',
    'https://ui-avatars.com/api/?name=John+Doe&background=10b981&color=fff',
    'active'
) ON CONFLICT (email) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    status = EXCLUDED.status,
    updated_at = CURRENT_TIMESTAMP;

-- 3. Regular User 2
INSERT INTO identity.users (
    id,
    email,
    email_verified,
    password_hash,
    display_name,
    avatar_url,
    status
) VALUES (
    gen_random_uuid(),
    'user2@wibusystem.dev',
    true,
    crypt('password123', gen_salt('bf', 10)),
    'Jane Smith',
    'https://ui-avatars.com/api/?name=Jane+Smith&background=f59e0b&color=fff',
    'active'
) ON CONFLICT (email) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    status = EXCLUDED.status,
    updated_at = CURRENT_TIMESTAMP;

-- 4. Pending User (not verified)
INSERT INTO identity.users (
    id,
    email,
    email_verified,
    password_hash,
    display_name,
    avatar_url,
    status
) VALUES (
    gen_random_uuid(),
    'pending@wibusystem.dev',
    false,
    crypt('password123', gen_salt('bf', 10)),
    'Pending User',
    'https://ui-avatars.com/api/?name=Pending&background=6b7280&color=fff',
    'pending_verification'
) ON CONFLICT (email) DO UPDATE SET
    status = EXCLUDED.status,
    updated_at = CURRENT_TIMESTAMP;

-- Create additional tenants
DO $$
DECLARE
    user1_id UUID;
    user2_id UUID;
    tenant_id UUID;
BEGIN
    -- Get user IDs
    SELECT id INTO user1_id FROM identity.users WHERE email = 'user1@wibusystem.dev';
    SELECT id INTO user2_id FROM identity.users WHERE email = 'user2@wibusystem.dev';

    -- Create user1's personal tenant
    INSERT INTO identity.tenants (
        id,
        name,
        slug,
        description,
        status,
        owner_id,
        settings
    ) VALUES (
        gen_random_uuid(),
        'John''s Workspace',
        'john-workspace',
        'Personal workspace for John Doe',
        'active',
        user1_id,
        jsonb_build_object(
            'features', jsonb_build_object(
                'api_access', true,
                'advanced_analytics', false
            ),
            'limits', jsonb_build_object(
                'max_members', 10,
                'max_projects', 5,
                'storage_gb', 10
            )
        )
    ) ON CONFLICT (slug) DO NOTHING
    RETURNING id INTO tenant_id;

    -- Add user1 as member of their own tenant
    IF tenant_id IS NOT NULL THEN
        INSERT INTO identity.tenant_members (
            tenant_id,
            user_id,
            role,
            permissions
        ) VALUES (
            tenant_id,
            user1_id,
            'owner',
            jsonb_build_array('tenant:read', 'tenant:write', 'member:read', 'member:write')
        ) ON CONFLICT ON CONSTRAINT tenant_members_tenant_id_user_id_key DO NOTHING;
    END IF;

    -- Create shared team tenant
    INSERT INTO identity.tenants (
        id,
        name,
        slug,
        description,
        status,
        owner_id,
        settings,
        metadata
    ) VALUES (
        gen_random_uuid(),
        'Team Collaboration',
        'team-collab',
        'Shared workspace for team members',
        'trial',
        user1_id,
        jsonb_build_object(
            'features', jsonb_build_object(
                'api_access', true,
                'advanced_analytics', true
            ),
            'limits', jsonb_build_object(
                'max_members', 20,
                'max_projects', 10,
                'storage_gb', 50
            )
        ),
        jsonb_build_object(
            'trial_ends_at', (CURRENT_TIMESTAMP + INTERVAL '30 days')::text
        )
    ) ON CONFLICT (slug) DO NOTHING
    RETURNING id INTO tenant_id;

    -- Add members to team tenant
    IF tenant_id IS NOT NULL THEN
        -- Owner
        INSERT INTO identity.tenant_members (tenant_id, user_id, role, permissions, invited_by)
        VALUES (
            tenant_id,
            user1_id,
            'owner',
            jsonb_build_array('tenant:read', 'tenant:write', 'tenant:delete', 'member:read', 'member:write', 'member:delete'),
            user1_id
        ) ON CONFLICT ON CONSTRAINT tenant_members_tenant_id_user_id_key DO NOTHING;

        -- Admin member
        INSERT INTO identity.tenant_members (tenant_id, user_id, role, permissions, invited_by)
        VALUES (
            tenant_id,
            user2_id,
            'admin',
            jsonb_build_array('tenant:read', 'member:read', 'member:write', 'settings:read'),
            user1_id
        ) ON CONFLICT ON CONSTRAINT tenant_members_tenant_id_user_id_key DO NOTHING;
    END IF;
END $$;

-- Display seeded data
\echo ''
\echo '╔══════════════════════════════════════════════════════════════════════╗'
\echo '║                Users and Tenants Seeded Successfully                ║'
\echo '╚══════════════════════════════════════════════════════════════════════╝'
\echo ''
\echo '👥 Seeded Users:'
\echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'

SELECT
    email,
    display_name,
    status,
    email_verified,
    created_at
FROM identity.users
ORDER BY created_at;

\echo ''
\echo '🏢 Seeded Tenants:'
\echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'

SELECT
    t.name,
    t.slug,
    t.status,
    u.email as owner_email,
    t.created_at
FROM identity.tenants t
JOIN identity.users u ON t.owner_id = u.id
ORDER BY t.created_at;

\echo ''
\echo '🔐 Test Credentials (All users):'
\echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'
\echo '   Email:    admin@wibusystem.dev'
\echo '   Password: password123'
\echo ''
\echo '   Email:    user1@wibusystem.dev'
\echo '   Password: password123'
\echo ''
\echo '   Email:    user2@wibusystem.dev'
\echo '   Password: password123'
\echo ''
\echo '⚠️  WARNING: These are DEVELOPMENT credentials only!'
\echo ''
