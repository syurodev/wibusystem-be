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
-- Seed Organization Permissions
-- =====================================================

INSERT INTO identify.permissions (name, scope, description, resource, action) VALUES
    -- Organization Management (renamed from tenant:*)
    ('organization:manage_member', 'organization', 'Manage organization members', 'organization', 'manage_member'),
    ('organization:assign_role', 'organization', 'Assign roles to members in organization', 'organization', 'assign_role'),
    ('organization:update_info', 'organization', 'Update organization information', 'organization', 'update_info'),
    ('organization:view_stats', 'organization', 'View organization statistics', 'organization', 'view_stats'),
    ('organization:billing_manage', 'organization', 'Manage organization billing', 'organization', 'billing_manage'),

    -- Organization Member Management (NEW)
    ('organization:invite_member', 'organization', 'Invite new members to organization', 'organization', 'invite_member'),
    ('organization:approve_invite', 'organization', 'Approve pending member invites', 'organization', 'approve_invite'),
    ('organization:kick_member', 'organization', 'Remove members from organization', 'organization', 'kick_member'),

    -- Organization Settings Management (NEW)
    ('organization:update_settings', 'organization', 'Update organization settings', 'organization', 'update_settings'),

    -- Organization Report Management (NEW)
    ('organization:respond_report', 'organization', 'Respond to reports about organization', 'organization', 'respond_report'),
    ('organization:view_reports', 'organization', 'View reports about organization', 'organization', 'view_reports'),

    -- Content Management (Anime)
    ('content:create_anime', 'organization', 'Create anime content', 'content', 'create_anime'),
    ('content:update_anime', 'organization', 'Update anime content', 'content', 'update_anime'),
    ('content:delete_anime', 'organization', 'Delete anime content', 'content', 'delete_anime'),

    -- Content Management (Manga)
    ('content:create_manga', 'organization', 'Create manga content', 'content', 'create_manga'),
    ('content:update_manga', 'organization', 'Update manga content', 'content', 'update_manga'),
    ('content:delete_manga', 'organization', 'Delete manga content', 'content', 'delete_manga'),

    -- Content Management (Novel)
    ('content:create_novel', 'organization', 'Create novel content', 'content', 'create_novel'),
    ('content:update_novel', 'organization', 'Update novel content', 'content', 'update_novel'),
    ('content:delete_novel', 'organization', 'Delete novel content', 'content', 'delete_novel'),

    -- Anime Management
    ('anime:episode_create', 'organization', 'Create anime episodes', 'anime', 'episode_create'),
    ('anime:episode_update', 'organization', 'Update anime episodes', 'anime', 'episode_update'),
    ('anime:episode_delete', 'organization', 'Delete anime episodes', 'anime', 'episode_delete'),
    ('anime:season_create', 'organization', 'Create anime seasons', 'anime', 'season_create'),
    ('anime:season_update', 'organization', 'Update anime seasons', 'anime', 'season_update'),
    ('anime:season_delete', 'organization', 'Delete anime seasons', 'anime', 'season_delete'),

    -- Manga Management
    ('manga:chapter_create', 'organization', 'Create manga chapters', 'manga', 'chapter_create'),
    ('manga:chapter_update', 'organization', 'Update manga chapters', 'manga', 'chapter_update'),
    ('manga:chapter_delete', 'organization', 'Delete manga chapters', 'manga', 'chapter_delete'),
    ('manga:volume_create', 'organization', 'Create manga volumes', 'manga', 'volume_create'),
    ('manga:volume_update', 'organization', 'Update manga volumes', 'manga', 'volume_update'),
    ('manga:volume_delete', 'organization', 'Delete manga volumes', 'manga', 'volume_delete'),

    -- Novel Management
    ('novel:chapter_create', 'organization', 'Create novel chapters', 'novel', 'chapter_create'),
    ('novel:chapter_update', 'organization', 'Update novel chapters', 'novel', 'chapter_update'),
    ('novel:chapter_delete', 'organization', 'Delete novel chapters', 'novel', 'chapter_delete'),
    ('novel:volume_create', 'organization', 'Create novel volumes', 'novel', 'volume_create'),
    ('novel:volume_update', 'organization', 'Update novel volumes', 'novel', 'volume_update'),
    ('novel:volume_delete', 'organization', 'Delete novel volumes', 'novel', 'volume_delete'),

    -- Master Data Management in Organization
    ('character:manage', 'organization', 'Manage characters in organization', 'character', 'manage'),
    ('creator:manage', 'organization', 'Manage creators in organization', 'creator', 'manage'),
    ('genre:manage', 'organization', 'Manage genres in organization', 'genre', 'manage'),
    ('relation:manage', 'organization', 'Manage relations in organization', 'relation', 'manage'),

    -- Content Publishing
    ('content:publish', 'organization', 'Publish content', 'content', 'publish'),
    ('content:unpublish', 'organization', 'Unpublish content', 'content', 'unpublish'),
    ('analytics:view', 'organization', 'View analytics', 'analytics', 'view')
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    resource = EXCLUDED.resource,
    action = EXCLUDED.action;

-- Global permission for reporting organizations
INSERT INTO identify.permissions (name, scope, description, resource, action) VALUES
    ('organization:report', 'global', 'Report an organization', 'organization', 'report')
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description;

-- =====================================================
-- Seed Global Roles
-- =====================================================

INSERT INTO identify.roles (name, slug, scope, description, is_system) VALUES
    ('SUPER_ADMIN', 'super-admin', 'global', 'Super Administrator with full system access', TRUE),
    ('ADMIN', 'admin', 'global', 'Administrator with system management access', TRUE),
    ('MODERATOR', 'moderator', 'global', 'Moderator for content and user management', TRUE),
    ('CREATOR', 'creator', 'global', 'Content creator with publishing permissions', TRUE),
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
        'creator:view', 'genre:view', 'relation:view',
        -- Organization
        'organization:report'
    )
ON CONFLICT DO NOTHING;

-- CREATOR: User permissions + creator:create/update
INSERT INTO identify.role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM identify.roles WHERE name = 'CREATOR'),
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
        -- Master data viewing + contribute
        'character:view', 'character:contribute', 'character:contribute_update_self',
        'creator:view', 'genre:view', 'relation:view',
        -- Creator specific
        'creator:create', 'creator:update',
        -- Organization
        'organization:report'
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
    COUNT(*) FILTER (WHERE scope = 'organization') as tenant_count
FROM identify.permissions;

SELECT
    '✅ Roles Seeded' as status,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE scope = 'global') as global_count,
    COUNT(*) FILTER (WHERE scope = 'organization') as tenant_count
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
