-- Seed: Permissions and Roles
-- Description: Seeds global and tenant permissions with default roles
-- Dependencies: 000002_create_permissions_system.up.sql migration

-- Ensure we're in the right schema
SET search_path TO identity, public;

\echo '📝 Seeding Global Permissions...'

-- ============================================================================
-- GLOBAL PERMISSIONS
-- ============================================================================

-- Auth & User permissions
INSERT INTO identity.global_permissions (key, description) VALUES
    ('auth:login', 'Allow user to log into the platform'),
    ('auth:logout', 'Allow user to logout'),
    ('auth:refresh_token', 'Issue refresh tokens'),
    ('user:view_self', 'View own profile'),
    ('user:update_self', 'Update own profile'),
    ('user:delete_self', 'Delete own account'),
    ('user:change_password', 'Change account password'),
    ('user:2fa_manage', 'Manage two-factor authentication')
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

-- Social / Community permissions
INSERT INTO identity.global_permissions (key, description) VALUES
    ('comment:create', 'Create comments'),
    ('comment:update_self', 'Edit own comments'),
    ('comment:delete_self', 'Delete own comments'),
    ('comment:report', 'Report abusive comments'),
    ('reaction:add', 'React to content'),
    ('review:create', 'Create reviews'),
    ('review:update_self', 'Edit own reviews'),
    ('review:delete_self', 'Delete own reviews'),
    ('follow:content', 'Follow content updates'),
    ('follow:user', 'Follow other users'),
    ('translation:submit', 'Submit community translations'),
    ('translation:update_self', 'Edit own translations'),
    ('translation:vote', 'Vote translations'),
    ('subtitle:contribute', 'Contribute subtitles'),
    ('report:content', 'Report inappropriate content')
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

-- Content viewing permissions
INSERT INTO identity.global_permissions (key, description) VALUES
    ('content:view_public', 'View public content'),
    ('content:view_purchased', 'View purchased/authorized content'),
    ('content:stream_anime', 'Stream anime'),
    ('content:read_manga', 'Read manga'),
    ('content:read_novel', 'Read novel')
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

-- Master data: Character permissions
INSERT INTO identity.global_permissions (key, description) VALUES
    ('character:view', 'View characters'),
    ('character:contribute', 'Contribute edits/suggestions to any character'),
    ('character:contribute_update_self', 'Update or delete own character contributions'),
    ('character:create', 'Create new characters (requires approval if not staff)'),
    ('character:approve', 'Approve new characters or contributions'),
    ('character:reject', 'Reject character contributions'),
    ('character:update', 'Update characters directly (staff only)'),
    ('character:delete', 'Delete characters (staff only)')
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

-- Master data: Creator permissions
INSERT INTO identity.global_permissions (key, description) VALUES
    ('creator:view', 'View creators'),
    ('creator:create', 'Create new creators'),
    ('creator:update', 'Update creators'),
    ('creator:delete', 'Delete creators')
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

-- Master data: Genre permissions
INSERT INTO identity.global_permissions (key, description) VALUES
    ('genre:view', 'View genres'),
    ('genre:create', 'Create genres'),
    ('genre:update', 'Update genres'),
    ('genre:delete', 'Delete genres')
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

-- Master data: Relations permissions
INSERT INTO identity.global_permissions (key, description) VALUES
    ('relation:view', 'View content relations'),
    ('relation:create', 'Create content relations'),
    ('relation:update', 'Update content relations'),
    ('relation:delete', 'Delete content relations')
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

-- Moderation & System permissions
INSERT INTO identity.global_permissions (key, description) VALUES
    ('moderation:content_review', 'Review reported content'),
    ('moderation:user_suspend', 'Suspend users'),
    ('moderation:ban', 'Ban users'),
    ('system:config_manage', 'Manage system configuration'),
    ('system:metrics_view', 'View platform metrics'),
    ('system:audit_view', 'View audit logs'),
    ('support:ticket_manage', 'Manage support tickets')
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

\echo '✅ Global permissions seeded'
\echo ''
\echo '📝 Seeding Global Roles...'

-- ============================================================================
-- GLOBAL ROLES
-- ============================================================================

-- Create roles
INSERT INTO identity.global_roles (name, description, is_system) VALUES
    ('SUPER_ADMIN', 'Full access to all system capabilities', true),
    ('ADMIN', 'Manage users, tenants, and master data', true),
    ('MODERATOR', 'Moderate community content and manage master data', true),
    ('USER', 'Default user role with community & content access', true),
    ('GUEST', 'Unauthenticated visitor with read-only access', true)
ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description;

\echo '✅ Global roles created'
\echo ''
\echo '📝 Assigning permissions to roles...'

-- ============================================================================
-- ROLE-PERMISSION MAPPINGS
-- ============================================================================

-- SUPER_ADMIN: All permissions
INSERT INTO identity.global_role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM identity.global_roles WHERE name = 'SUPER_ADMIN'),
    id
FROM identity.global_permissions
ON CONFLICT DO NOTHING;

-- ADMIN: Management permissions
INSERT INTO identity.global_role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM identity.global_roles WHERE name = 'ADMIN'),
    id
FROM identity.global_permissions
WHERE key IN (
    'auth:login', 'auth:logout', 'auth:refresh_token',
    'user:view_self', 'user:update_self', 'user:delete_self', 'user:change_password', 'user:2fa_manage',
    'comment:report', 'report:content',
    'content:view_public', 'content:view_purchased', 'content:stream_anime', 'content:read_manga', 'content:read_novel',
    'character:view', 'character:create', 'character:approve', 'character:reject', 'character:update', 'character:delete',
    'creator:view', 'creator:create', 'creator:update', 'creator:delete',
    'genre:view', 'genre:create', 'genre:update', 'genre:delete',
    'relation:view', 'relation:create', 'relation:update', 'relation:delete',
    'system:metrics_view', 'system:audit_view', 'support:ticket_manage'
)
ON CONFLICT DO NOTHING;

-- MODERATOR: Moderation permissions
INSERT INTO identity.global_role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM identity.global_roles WHERE name = 'MODERATOR'),
    id
FROM identity.global_permissions
WHERE key IN (
    'auth:login', 'auth:logout', 'auth:refresh_token',
    'user:view_self', 'user:update_self', 'user:change_password', 'user:2fa_manage',
    'comment:report', 'translation:vote', 'report:content',
    'content:view_public', 'content:view_purchased', 'content:stream_anime', 'content:read_manga', 'content:read_novel',
    'moderation:content_review', 'moderation:user_suspend', 'moderation:ban',
    'character:view', 'character:approve', 'character:reject',
    'creator:view', 'genre:view', 'relation:view',
    'system:metrics_view', 'support:ticket_manage'
)
ON CONFLICT DO NOTHING;

-- USER: Default user permissions
INSERT INTO identity.global_role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM identity.global_roles WHERE name = 'USER'),
    id
FROM identity.global_permissions
WHERE key IN (
    'auth:login', 'auth:logout', 'auth:refresh_token',
    'user:view_self', 'user:update_self', 'user:change_password', 'user:2fa_manage',
    'comment:create', 'comment:update_self', 'comment:delete_self', 'comment:report',
    'reaction:add', 'review:create', 'review:update_self', 'review:delete_self',
    'follow:content', 'follow:user',
    'translation:submit', 'translation:update_self', 'translation:vote', 'subtitle:contribute', 'report:content',
    'content:view_public', 'content:view_purchased', 'content:stream_anime', 'content:read_manga', 'content:read_novel',
    'character:view', 'character:contribute', 'character:contribute_update_self',
    'creator:view', 'genre:view', 'relation:view'
)
ON CONFLICT DO NOTHING;

-- GUEST: Read-only permissions
INSERT INTO identity.global_role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM identity.global_roles WHERE name = 'GUEST'),
    id
FROM identity.global_permissions
WHERE key IN ('content:view_public')
ON CONFLICT DO NOTHING;

\echo '✅ Role-permission mappings created'
\echo ''
\echo '📝 Seeding Tenant Permissions...'

-- ============================================================================
-- TENANT PERMISSIONS
-- ============================================================================

-- Tenant management permissions
INSERT INTO identity.permissions (key, description) VALUES
    ('tenant:manage_member', 'Manage tenant members'),
    ('tenant:assign_permission', 'Assign permissions within tenant'),
    ('tenant:update_info', 'Update tenant profile information'),
    ('tenant:view_stats', 'View tenant statistics'),
    ('tenant:billing_manage', 'Manage tenant billing')
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

-- Content creation permissions
INSERT INTO identity.permissions (key, description) VALUES
    ('content:create_anime', 'Create anime entries'),
    ('content:update_anime', 'Update anime entries'),
    ('content:delete_anime', 'Delete anime entries'),
    ('content:create_manga', 'Create manga entries'),
    ('content:update_manga', 'Update manga entries'),
    ('content:delete_manga', 'Delete manga entries'),
    ('content:create_novel', 'Create novel entries'),
    ('content:update_novel', 'Update novel entries'),
    ('content:delete_novel', 'Delete novel entries')
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

-- Episode/Chapter/Volume permissions
INSERT INTO identity.permissions (key, description) VALUES
    ('anime:episode_create', 'Create anime episodes'),
    ('anime:episode_update', 'Update anime episodes'),
    ('anime:episode_delete', 'Delete anime episodes'),
    ('manga:chapter_create', 'Create manga chapters'),
    ('manga:chapter_update', 'Update manga chapters'),
    ('manga:chapter_delete', 'Delete manga chapters'),
    ('novel:chapter_create', 'Create novel chapters'),
    ('novel:chapter_update', 'Update novel chapters'),
    ('novel:chapter_delete', 'Delete novel chapters'),
    ('manga:volume_create', 'Create manga volumes'),
    ('manga:volume_update', 'Update manga volumes'),
    ('manga:volume_delete', 'Delete manga volumes'),
    ('novel:volume_create', 'Create novel volumes'),
    ('novel:volume_update', 'Update novel volumes'),
    ('novel:volume_delete', 'Delete novel volumes'),
    ('anime:season_create', 'Create anime seasons'),
    ('anime:season_update', 'Update anime seasons'),
    ('anime:season_delete', 'Delete anime seasons')
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

-- Tenant-scoped master data permissions
INSERT INTO identity.permissions (key, description) VALUES
    ('character:manage', 'Manage characters in tenant scope'),
    ('creator:manage', 'Manage creators in tenant scope'),
    ('genre:manage', 'Manage genres in tenant scope'),
    ('relation:manage', 'Manage content relations in tenant scope')
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

-- Content publishing & analytics
INSERT INTO identity.permissions (key, description) VALUES
    ('content:publish', 'Publish content'),
    ('content:unpublish', 'Unpublish content'),
    ('analytics:view', 'View content analytics')
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

\echo '✅ Tenant permissions seeded'
\echo ''
\echo '╔══════════════════════════════════════════════════════════════════════╗'
\echo '║           Permissions and Roles Seeded Successfully                 ║'
\echo '╚══════════════════════════════════════════════════════════════════════╝'
\echo ''

-- Display statistics
SELECT
    (SELECT COUNT(*) FROM identity.global_permissions) as global_permissions,
    (SELECT COUNT(*) FROM identity.global_roles) as global_roles,
    (SELECT COUNT(*) FROM identity.global_role_permissions) as role_permission_mappings,
    (SELECT COUNT(*) FROM identity.permissions) as tenant_permissions;

\echo ''
\echo '📊 Permission Distribution by Role:'
\echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'

SELECT
    gr.name as role,
    COUNT(grp.permission_id) as permission_count
FROM identity.global_roles gr
LEFT JOIN identity.global_role_permissions grp ON gr.id = grp.role_id
GROUP BY gr.name
ORDER BY permission_count DESC;

\echo ''
\echo '⚠️  Note: Assign USER role to new users by default'
\echo '    Use SUPER_ADMIN role sparingly - only for system administrators'
\echo ''
