-- Migration: Add CREATOR role and user columns for creator stats
-- Version: 000032

-- =====================================================
-- Add new columns to users table
-- =====================================================

ALTER TABLE identify.users 
ADD COLUMN IF NOT EXISTS display_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS username VARCHAR(100),
ADD COLUMN IF NOT EXISTS bio JSONB,
ADD COLUMN IF NOT EXISTS is_verified BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS follower_count INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS works_count INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS last_content_updated_at TIMESTAMP WITH TIME ZONE;

-- Unique constraint for username
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON identify.users(username) WHERE username IS NOT NULL;

-- Index for listing creators
CREATE INDEX IF NOT EXISTS idx_users_is_verified ON identify.users(is_verified);
CREATE INDEX IF NOT EXISTS idx_users_last_content_updated_at ON identify.users(last_content_updated_at);
CREATE INDEX IF NOT EXISTS idx_users_follower_count ON identify.users(follower_count);

-- =====================================================
-- Add CREATOR role
-- =====================================================

INSERT INTO identify.roles (name, slug, scope, description, is_system) 
VALUES ('CREATOR', 'creator', 'global', 'Content creator with publishing permissions', TRUE)
ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description;

-- =====================================================
-- Assign permissions to CREATOR role
-- (USER permissions + creator:create/update)
-- =====================================================

INSERT INTO identify.role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM identify.roles WHERE name = 'CREATOR'),
    id 
FROM identify.permissions
WHERE scope = 'global' AND name IN (
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
    'creator:create', 'creator:update'
)
ON CONFLICT DO NOTHING;
