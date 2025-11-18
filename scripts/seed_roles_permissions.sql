-- Script: Seed Roles and Permissions
-- Description: Seed toàn bộ roles và permissions cho hệ thống
-- Author: System
-- Created: 2025-11-17

-- =====================================================
-- Seed Global Permissions
-- =====================================================

-- Auth & User Permissions (Global)
INSERT INTO identify.permissions (name, scope, description, resource, action) VALUES
    -- Auth
    ('auth:login', 'global', 'Login to system', 'auth', 'login'),
    ('auth:logout', 'global', 'Logout from system', 'auth', 'logout'),
    ('auth:refresh_token', 'global', 'Refresh access token', 'auth', 'refresh_token'),

    -- User Self-Management
    ('user:view_self', 'global', 'View own profile', 'user', 'view_self'),
    ('user:update_self', 'global', 'Update own profile', 'user', 'update_self'),
    ('user:delete_self', 'global', 'Delete own account', 'user', 'delete_self'),
    ('user:change_password', 'global', 'Change own password', 'user', 'change_password'),
    ('user:two_fa_manage', 'global', 'Manage 2FA settings', 'user', 'two_fa_manage'),

    -- Social / Community
    ('comment:create', 'global', 'Create comments', 'comment', 'create'),
    ('comment:update_self', 'global', 'Update own comments', 'comment', 'update_self'),
    ('comment:delete_self', 'global', 'Delete own comments', 'comment', 'delete_self'),
    ('comment:report', 'global', 'Report comments', 'comment', 'report'),
    ('reaction:add', 'global', 'Add reactions', 'reaction', 'add'),
    ('review:create', 'global', 'Create reviews', 'review', 'create'),
    ('review:update_self', 'global', 'Update own reviews', 'review', 'update_self'),
    ('review:delete_self', 'global', 'Delete own reviews', 'review', 'delete_self'),
    ('follow:content', 'global', 'Follow content', 'follow', 'content'),
    ('follow:user', 'global', 'Follow users', 'follow', 'user'),
    ('translation:submit', 'global', 'Submit translations', 'translation', 'submit'),
    ('translation:update_self', 'global', 'Update own translations', 'translation', 'update_self'),
    ('translation:vote', 'global', 'Vote on translations', 'translation', 'vote'),
    ('subtitle:contribute', 'global', 'Contribute subtitles', 'subtitle', 'contribute'),
    ('report:content', 'global', 'Report content', 'report', 'content'),

    -- Content Viewing
    ('content:view_public', 'global', 'View public content', 'content', 'view_public'),
    ('content:view_purchased', 'global', 'View purchased content', 'content', 'view_purchased'),
    ('content:stream_anime', 'global', 'Stream anime', 'content', 'stream_anime'),
    ('content:read_manga', 'global', 'Read manga', 'content', 'read_manga'),
    ('content:read_novel', 'global', 'Read novel', 'content', 'read_novel'),

    -- Master Data: Character
    ('character:view', 'global', 'View characters', 'character', 'view'),
    ('character:contribute', 'global', 'Contribute character data', 'character', 'contribute'),
    ('character:contribute_update_self', 'global', 'Update own character contributions', 'character', 'contribute_update_self'),
    ('character:create', 'global', 'Create characters', 'character', 'create'),
    ('character:approve', 'global', 'Approve character contributions', 'character', 'approve'),
    ('character:reject', 'global', 'Reject character contributions', 'character', 'reject'),
    ('character:update', 'global', 'Update characters', 'character', 'update'),
    ('character:delete', 'global', 'Delete characters', 'character', 'delete'),

    -- Master Data: Creator
    ('creator:view', 'global', 'View creators', 'creator', 'view'),
    ('creator:create', 'global', 'Create creators', 'creator', 'create'),
    ('creator:update', 'global', 'Update creators', 'creator', 'update'),
    ('creator:delete', 'global', 'Delete creators', 'creator', 'delete'),

    -- Master Data: Genre
    ('genre:view', 'global', 'View genres', 'genre', 'view'),
    ('genre:create', 'global', 'Create genres', 'genre', 'create'),
    ('genre:update', 'global', 'Update genres', 'genre', 'update'),
    ('genre:delete', 'global', 'Delete genres', 'genre', 'delete'),

    -- Master Data: Relations
    ('relation:view', 'global', 'View relations', 'relation', 'view'),
    ('relation:create', 'global', 'Create relations', 'relation', 'create'),
    ('relation:update', 'global', 'Update relations', 'relation', 'update'),
    ('relation:delete', 'global', 'Delete relations', 'relation', 'delete'),

    -- Moderation & System
    ('moderation:content_review', 'global', 'Review content moderation', 'moderation', 'content_review'),
    ('moderation:user_suspend', 'global', 'Suspend users', 'moderation', 'user_suspend'),
    ('moderation:ban', 'global', 'Ban users', 'moderation', 'ban'),
    ('system:config_manage', 'global', 'Manage system configuration', 'system', 'config_manage'),
    ('system:metrics_view', 'global', 'View system metrics', 'system', 'metrics_view'),
    ('system:audit_view', 'global', 'View audit logs', 'system', 'audit_view'),
    ('support:ticket_manage', 'global', 'Manage support tickets', 'support', 'ticket_manage')
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    resource = EXCLUDED.resource,
    action = EXCLUDED.action;

-- =====================================================
-- Seed Tenant Permissions
-- =====================================================

INSERT INTO identify.permissions (name, scope, description, resource, action) VALUES
    -- Tenant Management
    ('tenant:manage_member', 'tenant', 'Manage tenant members', 'tenant', 'manage_member'),
    ('tenant:assign_permission', 'tenant', 'Assign permissions in tenant', 'tenant', 'assign_permission'),
    ('tenant:update_info', 'tenant', 'Update tenant information', 'tenant', 'update_info'),
    ('tenant:view_stats', 'tenant', 'View tenant statistics', 'tenant', 'view_stats'),
    ('tenant:billing_manage', 'tenant', 'Manage tenant billing', 'tenant', 'billing_manage'),

    -- Content Management (Anime)
    ('content:create_anime', 'tenant', 'Create anime content', 'content', 'create_anime'),
    ('content:update_anime', 'tenant', 'Update anime content', 'content', 'update_anime'),
    ('content:delete_anime', 'tenant', 'Delete anime content', 'content', 'delete_anime'),

    -- Content Management (Manga)
    ('content:create_manga', 'tenant', 'Create manga content', 'content', 'create_manga'),
    ('content:update_manga', 'tenant', 'Update manga content', 'content', 'update_manga'),
    ('content:delete_manga', 'tenant', 'Delete manga content', 'content', 'delete_manga'),

    -- Content Management (Novel)
    ('content:create_novel', 'tenant', 'Create novel content', 'content', 'create_novel'),
    ('content:update_novel', 'tenant', 'Update novel content', 'content', 'update_novel'),
    ('content:delete_novel', 'tenant', 'Delete novel content', 'content', 'delete_novel'),

    -- Anime Management
    ('anime:episode_create', 'tenant', 'Create anime episodes', 'anime', 'episode_create'),
    ('anime:episode_update', 'tenant', 'Update anime episodes', 'anime', 'episode_update'),
    ('anime:episode_delete', 'tenant', 'Delete anime episodes', 'anime', 'episode_delete'),
    ('anime:season_create', 'tenant', 'Create anime seasons', 'anime', 'season_create'),
    ('anime:season_update', 'tenant', 'Update anime seasons', 'anime', 'season_update'),
    ('anime:season_delete', 'tenant', 'Delete anime seasons', 'anime', 'season_delete'),

    -- Manga Management
    ('manga:chapter_create', 'tenant', 'Create manga chapters', 'manga', 'chapter_create'),
    ('manga:chapter_update', 'tenant', 'Update manga chapters', 'manga', 'chapter_update'),
    ('manga:chapter_delete', 'tenant', 'Delete manga chapters', 'manga', 'chapter_delete'),
    ('manga:volume_create', 'tenant', 'Create manga volumes', 'manga', 'volume_create'),
    ('manga:volume_update', 'tenant', 'Update manga volumes', 'manga', 'volume_update'),
    ('manga:volume_delete', 'tenant', 'Delete manga volumes', 'manga', 'volume_delete'),

    -- Novel Management
    ('novel:chapter_create', 'tenant', 'Create novel chapters', 'novel', 'chapter_create'),
    ('novel:chapter_update', 'tenant', 'Update novel chapters', 'novel', 'chapter_update'),
    ('novel:chapter_delete', 'tenant', 'Delete novel chapters', 'novel', 'chapter_delete'),
    ('novel:volume_create', 'tenant', 'Create novel volumes', 'novel', 'volume_create'),
    ('novel:volume_update', 'tenant', 'Update novel volumes', 'novel', 'volume_update'),
    ('novel:volume_delete', 'tenant', 'Delete novel volumes', 'novel', 'volume_delete'),

    -- Master Data Management in Tenant
    ('character:manage', 'tenant', 'Manage characters in tenant', 'character', 'manage'),
    ('creator:manage', 'tenant', 'Manage creators in tenant', 'creator', 'manage'),
    ('genre:manage', 'tenant', 'Manage genres in tenant', 'genre', 'manage'),
    ('relation:manage', 'tenant', 'Manage relations in tenant', 'relation', 'manage'),

    -- Content Publishing
    ('content:publish', 'tenant', 'Publish content', 'content', 'publish'),
    ('content:unpublish', 'tenant', 'Unpublish content', 'content', 'unpublish'),
    ('analytics:view', 'tenant', 'View analytics', 'analytics', 'view')
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    resource = EXCLUDED.resource,
    action = EXCLUDED.action;

-- =====================================================
-- Seed Global Roles
-- =====================================================

INSERT INTO identify.roles (name, slug, scope, description, is_system) VALUES
    ('SUPER_ADMIN', 'super-admin', 'global', 'Super Administrator with full system access', TRUE),
    ('ADMIN', 'admin', 'global', 'Administrator with system management access', TRUE),
    ('MODERATOR', 'moderator', 'global', 'Moderator for content and user management', TRUE),
    ('USER', 'user', 'global', 'Regular user with standard permissions', TRUE),
    ('GUEST', 'guest', 'global', 'Guest user with view-only permissions', TRUE)
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description;

-- =====================================================
-- Map Permissions to Roles
-- =====================================================

-- SUPER_ADMIN: All permissions
INSERT INTO identify.role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM identify.roles WHERE name = 'SUPER_ADMIN'),
    id
FROM identify.permissions
ON CONFLICT DO NOTHING;

-- ADMIN: All global permissions + moderation + system management
INSERT INTO identify.role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM identify.roles WHERE name = 'ADMIN'),
    id
FROM identify.permissions
WHERE scope = 'global'
    AND name IN (
        -- Auth
        'auth:login', 'auth:logout', 'auth:refresh_token',
        -- User self-management
        'user:view_self', 'user:update_self', 'user:change_password', 'user:two_fa_manage',
        -- Master data management
        'character:view', 'character:create', 'character:update', 'character:delete',
        'character:approve', 'character:reject',
        'creator:view', 'creator:create', 'creator:update', 'creator:delete',
        'genre:view', 'genre:create', 'genre:update', 'genre:delete',
        'relation:view', 'relation:create', 'relation:update', 'relation:delete',
        -- Moderation
        'moderation:content_review', 'moderation:user_suspend', 'moderation:ban',
        -- System
        'system:config_manage', 'system:metrics_view', 'system:audit_view',
        'support:ticket_manage',
        -- Content viewing
        'content:view_public', 'content:view_purchased', 'content:stream_anime',
        'content:read_manga', 'content:read_novel'
    )
ON CONFLICT DO NOTHING;

-- MODERATOR: Moderation + content review permissions
INSERT INTO identify.role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM identify.roles WHERE name = 'MODERATOR'),
    id
FROM identify.permissions
WHERE scope = 'global'
    AND name IN (
        -- Auth
        'auth:login', 'auth:logout', 'auth:refresh_token',
        -- User self-management
        'user:view_self', 'user:update_self', 'user:change_password',
        -- Character moderation
        'character:view', 'character:approve', 'character:reject', 'character:update',
        'creator:view', 'genre:view', 'relation:view',
        -- Moderation
        'moderation:content_review', 'moderation:user_suspend',
        -- Content viewing
        'content:view_public', 'content:view_purchased', 'content:stream_anime',
        'content:read_manga', 'content:read_novel',
        -- Community
        'comment:create', 'comment:update_self', 'comment:delete_self',
        'review:create', 'review:update_self', 'review:delete_self'
    )
ON CONFLICT DO NOTHING;

-- USER: Standard user permissions
INSERT INTO identify.role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM identify.roles WHERE name = 'USER'),
    id
FROM identify.permissions
WHERE scope = 'global'
    AND name IN (
        -- Auth
        'auth:login', 'auth:logout', 'auth:refresh_token',
        -- User self-management
        'user:view_self', 'user:update_self', 'user:delete_self',
        'user:change_password', 'user:two_fa_manage',
        -- Social / Community
        'comment:create', 'comment:update_self', 'comment:delete_self', 'comment:report',
        'reaction:add',
        'review:create', 'review:update_self', 'review:delete_self',
        'follow:content', 'follow:user',
        'translation:submit', 'translation:update_self', 'translation:vote',
        'subtitle:contribute',
        'report:content',
        -- Content viewing
        'content:view_public', 'content:view_purchased',
        'content:stream_anime', 'content:read_manga', 'content:read_novel',
        -- Master data viewing
        'character:view', 'character:contribute', 'character:contribute_update_self',
        'creator:view', 'genre:view', 'relation:view'
    )
ON CONFLICT DO NOTHING;

-- GUEST: View-only permissions
INSERT INTO identify.role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM identify.roles WHERE name = 'GUEST'),
    id
FROM identify.permissions
WHERE scope = 'global'
    AND name IN (
        -- Basic viewing
        'content:view_public',
        'character:view',
        'creator:view',
        'genre:view',
        'relation:view'
    )
ON CONFLICT DO NOTHING;

-- =====================================================
-- Display Seed Summary
-- =====================================================

SELECT
    '✅ Permissions Seeded' as status,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE scope = 'global') as global_count,
    COUNT(*) FILTER (WHERE scope = 'tenant') as tenant_count
FROM identify.permissions;

SELECT
    '✅ Roles Seeded' as status,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE scope = 'global') as global_count,
    COUNT(*) FILTER (WHERE scope = 'tenant') as tenant_count
FROM identify.roles;

SELECT
    '✅ Role-Permission Mappings Created' as status,
    COUNT(*) as total_mappings
FROM identify.role_permissions;

-- Display role breakdown
SELECT
    r.name as role,
    r.scope,
    COUNT(rp.permission_id) as permission_count
FROM identify.roles r
LEFT JOIN identify.role_permissions rp ON r.id = rp.role_id
WHERE r.scope = 'global'
GROUP BY r.id, r.name, r.scope
ORDER BY r.name;
