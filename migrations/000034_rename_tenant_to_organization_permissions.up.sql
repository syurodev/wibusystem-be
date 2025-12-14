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
