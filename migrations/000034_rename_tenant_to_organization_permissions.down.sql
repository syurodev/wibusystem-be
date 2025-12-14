-- Rollback migration 000034
-- =====================================================
-- 3. Rename back to tenant
-- =====================================================

UPDATE identify.permissions SET 
    name = 'tenant:manage_member',
    resource = 'organization',
    action = 'manage_member',
    description = 'Manage tenant members'
WHERE name = 'organization:manage_member';

UPDATE identify.permissions SET 
    name = 'tenant:assign_permission',
    resource = 'organization',
    action = 'assign_permission',
    description = 'Assign permissions in tenant'
WHERE name = 'organization:assign_role';

UPDATE identify.permissions SET 
    name = 'tenant:update_info',
    resource = 'organization',
    action = 'update_info',
    description = 'Update tenant information'
WHERE name = 'organization:update_info';

UPDATE identify.permissions SET 
    name = 'tenant:view_stats',
    resource = 'organization',
    action = 'view_stats',
    description = 'View tenant statistics'
WHERE name = 'organization:view_stats';

UPDATE identify.permissions SET 
    name = 'tenant:billing_manage',
    resource = 'organization',
    action = 'billing_manage',
    description = 'Manage tenant billing'
WHERE name = 'organization:billing_manage';
