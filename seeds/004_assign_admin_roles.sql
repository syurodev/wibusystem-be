-- Seed: Assign Admin Roles
-- Description: Assign global roles to admin users
-- Dependencies: 002_users_and_tenants.sql, 003_permissions_and_roles.sql

-- Assign SUPER_ADMIN role to admin@wibusystem.dev
DO $$
DECLARE
    admin_user_id UUID;
    super_admin_role_id UUID;
BEGIN
    -- Get admin user ID
    SELECT id INTO admin_user_id
    FROM identity.users
    WHERE email = 'admin@wibusystem.dev';

    -- Get SUPER_ADMIN role ID
    SELECT id INTO super_admin_role_id
    FROM identity.global_roles
    WHERE name = 'SUPER_ADMIN';

    -- Assign SUPER_ADMIN role to admin user
    IF admin_user_id IS NOT NULL AND super_admin_role_id IS NOT NULL THEN
        INSERT INTO identity.user_global_roles (
            user_id,
            role_id,
            assigned_by
        ) VALUES (
            admin_user_id,
            super_admin_role_id,
            admin_user_id  -- Self-assigned for initial setup
        ) ON CONFLICT (user_id, role_id) DO NOTHING;

        RAISE NOTICE '✅ SUPER_ADMIN role assigned to admin@wibusystem.dev';
    ELSE
        IF admin_user_id IS NULL THEN
            RAISE WARNING '⚠️  Admin user not found';
        END IF;
        IF super_admin_role_id IS NULL THEN
            RAISE WARNING '⚠️  SUPER_ADMIN role not found';
        END IF;
    END IF;
END $$;

-- Assign USER role to regular users
DO $$
DECLARE
    user_role_id UUID;
    user_record RECORD;
BEGIN
    -- Get USER role ID
    SELECT id INTO user_role_id
    FROM identity.global_roles
    WHERE name = 'USER';

    IF user_role_id IS NOT NULL THEN
        -- Assign USER role to all regular users who don't have any global roles yet
        FOR user_record IN
            SELECT u.id, u.email
            FROM identity.users u
            WHERE u.email IN ('user1@wibusystem.dev', 'user2@wibusystem.dev', 'pending@wibusystem.dev')
            AND NOT EXISTS (
                SELECT 1 FROM identity.user_global_roles ugr
                WHERE ugr.user_id = u.id
            )
        LOOP
            INSERT INTO identity.user_global_roles (
                user_id,
                role_id,
                assigned_by
            )
            SELECT
                user_record.id,
                user_role_id,
                (SELECT id FROM identity.users WHERE email = 'admin@wibusystem.dev')
            ON CONFLICT (user_id, role_id) DO NOTHING;

            RAISE NOTICE '✅ USER role assigned to %', user_record.email;
        END LOOP;
    ELSE
        RAISE WARNING '⚠️  USER role not found';
    END IF;
END $$;

-- Display assigned roles
\echo ''
\echo '╔══════════════════════════════════════════════════════════════════════╗'
\echo '║                  Admin Roles Assigned Successfully                  ║'
\echo '╚══════════════════════════════════════════════════════════════════════╝'
\echo ''
\echo '👑 User Global Roles:'
\echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'

SELECT
    u.email,
    u.display_name,
    r.name as role,
    ugr.assigned_at
FROM identity.user_global_roles ugr
JOIN identity.users u ON ugr.user_id = u.id
JOIN identity.global_roles r ON ugr.role_id = r.id
ORDER BY r.name, u.email;

\echo ''
\echo '📊 Role Summary:'
\echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'

SELECT
    r.name as role,
    COUNT(ugr.user_id) as user_count
FROM identity.global_roles r
LEFT JOIN identity.user_global_roles ugr ON r.id = ugr.role_id
GROUP BY r.name
ORDER BY r.name;

\echo ''
