-- Migration: Rename tenant permissions to organization and add new permissions
-- Version: 000034

-- =====================================================
-- 1. Rename existing tenant permissions to organization
-- =====================================================

UPDATE identify.permissions SET 
    name = 'organization:manage_member',
    resource = 'organization',
    action = 'manage_member',
    description = 'Manage organization members'
WHERE name = 'tenant:manage_member';

UPDATE identify.permissions SET 
    name = 'organization:assign_role',
    resource = 'organization',
    action = 'assign_role',
    description = 'Assign roles to members in organization'
WHERE name = 'tenant:assign_permission';

UPDATE identify.permissions SET 
    name = 'organization:update_info',
    resource = 'organization',
    action = 'update_info',
    description = 'Update organization information'
WHERE name = 'tenant:update_info';

UPDATE identify.permissions SET 
    name = 'organization:view_stats',
    resource = 'organization',
    action = 'view_stats',
    description = 'View organization statistics'
WHERE name = 'tenant:view_stats';

UPDATE identify.permissions SET 
    name = 'organization:billing_manage',
    resource = 'organization',
    action = 'billing_manage',
    description = 'Manage organization billing'
WHERE name = 'tenant:billing_manage';

-- =====================================================
-- 2. Add new organization permissions (scoped to organization)
-- =====================================================

INSERT INTO identify.permissions (name, scope, description, resource, action) VALUES
    -- Member Management
    ('organization:invite_member', 'organization', 'Invite new members to organization', 'organization', 'invite_member'),
    ('organization:approve_invite', 'organization', 'Approve pending member invites', 'organization', 'approve_invite'),
    ('organization:kick_member', 'organization', 'Remove members from organization', 'organization', 'kick_member'),
    
    -- Settings Management
    ('organization:update_settings', 'organization', 'Update organization settings', 'organization', 'update_settings'),
    
    -- Report Management
    ('organization:respond_report', 'organization', 'Respond to reports about organization', 'organization', 'respond_report'),
    ('organization:view_reports', 'organization', 'View reports about organization', 'organization', 'view_reports')
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    resource = EXCLUDED.resource,
    action = EXCLUDED.action;

-- =====================================================
-- 3. Add global permission for reporting organizations
-- =====================================================

INSERT INTO identify.permissions (name, scope, description, resource, action) VALUES
    ('organization:report', 'global', 'Report an organization', 'organization', 'report')
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description;

-- Add organization:report to USER role
INSERT INTO identify.role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM identify.roles WHERE name = 'USER'),
    (SELECT id FROM identify.permissions WHERE name = 'organization:report')
ON CONFLICT DO NOTHING;

-- Add organization:report to CREATOR role
INSERT INTO identify.role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM identify.roles WHERE name = 'CREATOR'),
    (SELECT id FROM identify.permissions WHERE name = 'organization:report')
ON CONFLICT DO NOTHING;

-- =====================================================
-- 4. Display summary
-- =====================================================

SELECT 
    '✅ Permissions Updated' as status,
    COUNT(*) FILTER (WHERE name LIKE 'organization:%') as organization_permissions
FROM identify.permissions;
