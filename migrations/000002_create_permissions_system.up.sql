-- Migration: 000002_create_permissions_system
-- Description: Creates global and tenant-level permissions and roles system
-- Schema: identity

BEGIN;

SET search_path TO identity, public;

-- ============================================================================
-- GLOBAL PERMISSIONS & ROLES
-- ============================================================================

-- Global permissions table (platform-wide permissions)
CREATE TABLE identity.global_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_global_permissions_key ON identity.global_permissions(key);

COMMENT ON TABLE identity.global_permissions IS 'Global permissions applicable across the platform';
COMMENT ON COLUMN identity.global_permissions.key IS 'Permission key (e.g., auth:login, user:view_self)';

-- Global roles table (system-wide roles like SUPER_ADMIN, ADMIN, USER, GUEST)
CREATE TABLE identity.global_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT TRUE, -- System roles cannot be deleted
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_global_roles_name ON identity.global_roles(name);

COMMENT ON TABLE identity.global_roles IS 'Global roles with platform-wide access control';
COMMENT ON COLUMN identity.global_roles.is_system IS 'System roles cannot be modified or deleted';

-- Global role-permission mapping
CREATE TABLE identity.global_role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_global_role_permissions_role ON identity.global_role_permissions(role_id);
CREATE INDEX idx_global_role_permissions_permission ON identity.global_role_permissions(permission_id);

COMMENT ON TABLE identity.global_role_permissions IS 'Mapping between global roles and permissions';

-- User global roles (assign global roles to users)
CREATE TABLE identity.user_global_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    assigned_by UUID,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_user_global_roles_user ON identity.user_global_roles(user_id);
CREATE INDEX idx_user_global_roles_role ON identity.user_global_roles(role_id);

COMMENT ON TABLE identity.user_global_roles IS 'Assigns global roles to users';

-- ============================================================================
-- TENANT-LEVEL PERMISSIONS
-- ============================================================================

-- Tenant permissions (permissions specific to tenant operations)
CREATE TABLE identity.permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_permissions_key ON identity.permissions(key);

COMMENT ON TABLE identity.permissions IS 'Tenant-level permissions for content and tenant management';
COMMENT ON COLUMN identity.permissions.key IS 'Permission key (e.g., content:create_anime, tenant:manage_member)';

-- Tenant member permissions (direct permission assignment)
-- Note: We keep the existing JSONB permissions column in tenant_members for flexibility
-- This table provides a more structured alternative
CREATE TABLE identity.tenant_member_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_member_id UUID NOT NULL, -- References tenant_members(id)
    permission_id UUID NOT NULL,
    granted_by UUID,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_member_id, permission_id)
);

CREATE INDEX idx_tenant_member_permissions_member ON identity.tenant_member_permissions(tenant_member_id);
CREATE INDEX idx_tenant_member_permissions_permission ON identity.tenant_member_permissions(permission_id);

COMMENT ON TABLE identity.tenant_member_permissions IS 'Direct permission assignments to tenant members';

-- ============================================================================
-- FOREIGN KEY CONSTRAINTS
-- ============================================================================

-- Global role permissions
ALTER TABLE identity.global_role_permissions
    ADD CONSTRAINT fk_global_role_permissions_role
    FOREIGN KEY (role_id)
    REFERENCES identity.global_roles(id)
    ON DELETE CASCADE;

ALTER TABLE identity.global_role_permissions
    ADD CONSTRAINT fk_global_role_permissions_permission
    FOREIGN KEY (permission_id)
    REFERENCES identity.global_permissions(id)
    ON DELETE CASCADE;

-- User global roles
ALTER TABLE identity.user_global_roles
    ADD CONSTRAINT fk_user_global_roles_user
    FOREIGN KEY (user_id)
    REFERENCES identity.users(id)
    ON DELETE CASCADE;

ALTER TABLE identity.user_global_roles
    ADD CONSTRAINT fk_user_global_roles_role
    FOREIGN KEY (role_id)
    REFERENCES identity.global_roles(id)
    ON DELETE CASCADE;

ALTER TABLE identity.user_global_roles
    ADD CONSTRAINT fk_user_global_roles_assigned_by
    FOREIGN KEY (assigned_by)
    REFERENCES identity.users(id)
    ON DELETE SET NULL;

-- Tenant member permissions
ALTER TABLE identity.tenant_member_permissions
    ADD CONSTRAINT fk_tenant_member_permissions_member
    FOREIGN KEY (tenant_member_id)
    REFERENCES identity.tenant_members(id)
    ON DELETE CASCADE;

ALTER TABLE identity.tenant_member_permissions
    ADD CONSTRAINT fk_tenant_member_permissions_permission
    FOREIGN KEY (permission_id)
    REFERENCES identity.permissions(id)
    ON DELETE CASCADE;

ALTER TABLE identity.tenant_member_permissions
    ADD CONSTRAINT fk_tenant_member_permissions_granted_by
    FOREIGN KEY (granted_by)
    REFERENCES identity.users(id)
    ON DELETE SET NULL;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Update trigger for global_permissions
CREATE TRIGGER update_global_permissions_updated_at
    BEFORE UPDATE ON identity.global_permissions
    FOR EACH ROW
    EXECUTE FUNCTION identity.update_updated_at_column();

-- Update trigger for global_roles
CREATE TRIGGER update_global_roles_updated_at
    BEFORE UPDATE ON identity.global_roles
    FOR EACH ROW
    EXECUTE FUNCTION identity.update_updated_at_column();

-- Update trigger for permissions
CREATE TRIGGER update_permissions_updated_at
    BEFORE UPDATE ON identity.permissions
    FOR EACH ROW
    EXECUTE FUNCTION identity.update_updated_at_column();

-- ============================================================================
-- HELPER VIEWS
-- ============================================================================

-- View: User effective global permissions
CREATE OR REPLACE VIEW identity.user_effective_global_permissions AS
SELECT DISTINCT
    u.id as user_id,
    u.email,
    gp.id as permission_id,
    gp.key as permission_key,
    gp.description as permission_description,
    gr.name as role_name
FROM identity.users u
JOIN identity.user_global_roles ugr ON u.id = ugr.user_id
JOIN identity.global_roles gr ON ugr.role_id = gr.id
JOIN identity.global_role_permissions grp ON gr.id = grp.role_id
JOIN identity.global_permissions gp ON grp.permission_id = gp.id;

COMMENT ON VIEW identity.user_effective_global_permissions IS 'Effective global permissions for each user through their roles';

-- View: Tenant member effective permissions
CREATE OR REPLACE VIEW identity.tenant_member_effective_permissions AS
SELECT DISTINCT
    tm.id as tenant_member_id,
    tm.tenant_id,
    tm.user_id,
    u.email,
    p.id as permission_id,
    p.key as permission_key,
    p.description as permission_description,
    'direct' as source
FROM identity.tenant_members tm
JOIN identity.users u ON tm.user_id = u.id
JOIN identity.tenant_member_permissions tmp ON tm.id = tmp.tenant_member_id
JOIN identity.permissions p ON tmp.permission_id = p.id;

COMMENT ON VIEW identity.tenant_member_effective_permissions IS 'Effective tenant permissions for each member';

COMMIT;
