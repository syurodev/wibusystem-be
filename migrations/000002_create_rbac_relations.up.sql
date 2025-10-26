-- Migration: Create RBAC Relation Tables
-- Description: Tạo các bảng quan hệ cho RBAC và multi-tenant user management
-- Author: System
-- Created: 2025-10-26

-- =====================================================
-- Table: role_permissions
-- Description: Bảng nối many-to-many giữa roles và permissions
-- =====================================================
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Composite primary key
    PRIMARY KEY (role_id, permission_id)
);

-- Indexes
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Comment
COMMENT ON TABLE role_permissions IS 'Bảng nối many-to-many liên kết roles với permissions';
COMMENT ON COLUMN role_permissions.role_id IS 'ID của role';
COMMENT ON COLUMN role_permissions.permission_id IS 'ID của permission được gán cho role';

-- =====================================================
-- Table: user_tenant_memberships
-- Description: Liên kết user với tenant (user có thể thuộc nhiều tenants)
-- =====================================================
CREATE TABLE user_tenant_memberships (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Membership status
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, pending_invite, suspended

    -- Metadata
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    invited_at TIMESTAMPTZ,
    joined_at TIMESTAMPTZ DEFAULT NOW(),

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Composite primary key
    PRIMARY KEY (user_id, tenant_id),

    -- Constraints
    CONSTRAINT user_tenant_memberships_status_check CHECK (status IN ('active', 'pending_invite', 'suspended'))
);

-- Indexes
CREATE INDEX idx_user_tenant_memberships_user_id ON user_tenant_memberships(user_id);
CREATE INDEX idx_user_tenant_memberships_tenant_id ON user_tenant_memberships(tenant_id);
CREATE INDEX idx_user_tenant_memberships_status ON user_tenant_memberships(status);
CREATE INDEX idx_user_tenant_memberships_invited_by ON user_tenant_memberships(invited_by);

-- Trigger for updated_at
CREATE TRIGGER update_user_tenant_memberships_updated_at BEFORE UPDATE ON user_tenant_memberships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comment
COMMENT ON TABLE user_tenant_memberships IS 'Bảng liên kết users với tenants - một user có thể thuộc nhiều tenants';
COMMENT ON COLUMN user_tenant_memberships.status IS 'Trạng thái membership: active (đang hoạt động), pending_invite (chờ chấp nhận), suspended (bị đình chỉ)';
COMMENT ON COLUMN user_tenant_memberships.invited_by IS 'User đã mời thành viên này vào tenant';

-- =====================================================
-- Table: user_tenant_roles
-- Description: Gán roles cho user trong context của một tenant cụ thể
-- =====================================================
CREATE TABLE user_tenant_roles (
    user_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Composite primary key
    PRIMARY KEY (user_id, tenant_id, role_id),

    -- Foreign key compound constraint
    FOREIGN KEY (user_id, tenant_id)
        REFERENCES user_tenant_memberships(user_id, tenant_id)
        ON DELETE CASCADE

    -- Note: Scope validation phải được thực hiện ở application level
    -- hoặc trigger vì PostgreSQL không cho phép subquery trong CHECK constraint
);

-- Indexes
CREATE INDEX idx_user_tenant_roles_user_id ON user_tenant_roles(user_id);
CREATE INDEX idx_user_tenant_roles_tenant_id ON user_tenant_roles(tenant_id);
CREATE INDEX idx_user_tenant_roles_role_id ON user_tenant_roles(role_id);
CREATE INDEX idx_user_tenant_roles_user_tenant ON user_tenant_roles(user_id, tenant_id);

-- Comment
COMMENT ON TABLE user_tenant_roles IS 'Gán roles cho user trong context của một tenant cụ thể. QUAN TRỌNG: Application phải đảm bảo chỉ gán tenant-scoped roles';

-- =====================================================
-- Table: user_global_roles
-- Description: Gán global roles cho user (độc lập với tenant)
-- =====================================================
CREATE TABLE user_global_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Composite primary key
    PRIMARY KEY (user_id, role_id)

    -- Note: Scope validation phải được thực hiện ở application level
    -- hoặc trigger vì PostgreSQL không cho phép subquery trong CHECK constraint
);

-- Indexes
CREATE INDEX idx_user_global_roles_user_id ON user_global_roles(user_id);
CREATE INDEX idx_user_global_roles_role_id ON user_global_roles(role_id);

-- Comment
COMMENT ON TABLE user_global_roles IS 'Gán global roles cho user, độc lập với tenant. QUAN TRỌNG: Application phải đảm bảo chỉ gán global-scoped roles';

-- =====================================================
-- Seed Data: Gán permissions cho roles
-- Description: Mapping permissions vào các roles đã tạo
-- =====================================================

-- SUPER_ADMIN: Tất cả permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'SUPER_ADMIN'),
    id
FROM permissions;

-- PLATFORM_ADMIN: Global permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'PLATFORM_ADMIN'),
    id
FROM permissions
WHERE scope = 'global';

-- TENANT_ADMIN: Tất cả tenant permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'TENANT_ADMIN'),
    id
FROM permissions
WHERE scope = 'tenant';

-- TENANT_MANAGER: Một số tenant permissions (không bao gồm role:manage)
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'TENANT_MANAGER'),
    id
FROM permissions
WHERE scope = 'tenant'
    AND name IN (
        'user:view', 'user:create', 'user:update',
        'content:view', 'content:create', 'content:update', 'content:delete'
    );

-- TENANT_USER: Permissions cơ bản
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'TENANT_USER'),
    id
FROM permissions
WHERE name IN ('content:view', 'content:create', 'content:update');

-- TENANT_VIEWER: Chỉ view
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'TENANT_VIEWER'),
    id
FROM permissions
WHERE name IN ('content:view', 'user:view');

-- =====================================================
-- Helper Functions: Permission Check
-- Description: Functions hỗ trợ kiểm tra permissions
-- =====================================================

-- Function: Kiểm tra user có permission trong tenant không
CREATE OR REPLACE FUNCTION user_has_tenant_permission(
    p_user_id UUID,
    p_tenant_id UUID,
    p_permission_name VARCHAR
)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM user_tenant_roles utr
        JOIN role_permissions rp ON utr.role_id = rp.role_id
        JOIN permissions p ON rp.permission_id = p.id
        WHERE utr.user_id = p_user_id
            AND utr.tenant_id = p_tenant_id
            AND p.name = p_permission_name
    );
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION user_has_tenant_permission IS 'Kiểm tra xem user có permission cụ thể trong tenant không';

-- Function: Kiểm tra user có global permission không
CREATE OR REPLACE FUNCTION user_has_global_permission(
    p_user_id UUID,
    p_permission_name VARCHAR
)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM user_global_roles ugr
        JOIN role_permissions rp ON ugr.role_id = rp.role_id
        JOIN permissions p ON rp.permission_id = p.id
        WHERE ugr.user_id = p_user_id
            AND p.name = p_permission_name
    );
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION user_has_global_permission IS 'Kiểm tra xem user có global permission cụ thể không';

-- Function: Lấy tất cả permissions của user trong tenant
CREATE OR REPLACE FUNCTION get_user_tenant_permissions(
    p_user_id UUID,
    p_tenant_id UUID
)
RETURNS TABLE(permission_name VARCHAR, permission_description TEXT) AS $$
BEGIN
    RETURN QUERY
    SELECT DISTINCT p.name, p.description
    FROM user_tenant_roles utr
    JOIN role_permissions rp ON utr.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE utr.user_id = p_user_id
        AND utr.tenant_id = p_tenant_id
    ORDER BY p.name;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION get_user_tenant_permissions IS 'Lấy danh sách tất cả permissions của user trong tenant';

-- Function: Lấy tất cả global permissions của user
CREATE OR REPLACE FUNCTION get_user_global_permissions(
    p_user_id UUID
)
RETURNS TABLE(permission_name VARCHAR, permission_description TEXT) AS $$
BEGIN
    RETURN QUERY
    SELECT DISTINCT p.name, p.description
    FROM user_global_roles ugr
    JOIN role_permissions rp ON ugr.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE ugr.user_id = p_user_id
    ORDER BY p.name;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION get_user_global_permissions IS 'Lấy danh sách tất cả global permissions của user';

COMMENT ON TABLE role_permissions IS 'Seeded với permission mappings cho tất cả default roles';
