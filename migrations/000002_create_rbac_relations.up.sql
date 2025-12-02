-- Migration: Create RBAC Relation Tables
-- Description: Tạo các bảng quan hệ cho RBAC và multi-organization user management
-- Author: System
-- Created: 2025-10-26

-- =====================================================
-- ENUM TYPES
-- =====================================================

-- Organization member role (merged from migration 000014 team_member_role)
CREATE TYPE organization_member_role AS ENUM (
    'leader',
    'translator',
    'proofreader',
    'editor',
    'member'
);

COMMENT ON TYPE organization_member_role IS 'Vai trò của member trong organization';

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
-- Table: user_organization_memberships
-- Description: Liên kết user với organization (user có thể thuộc nhiều organizations)
-- Merged with team_members from migration 000014
-- =====================================================
CREATE TABLE user_organization_memberships (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Membership status
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, pending_invite, suspended

    -- Organization member role (from translation teams)
    role organization_member_role NOT NULL DEFAULT 'member',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    -- Statistics (auto-updated by application, from team_members)
    contribution_count INTEGER NOT NULL DEFAULT 0 CHECK (contribution_count >= 0),
    quality_score DECIMAL(3,2) DEFAULT 0.00 CHECK (quality_score >= 0 AND quality_score <= 5),

    -- Metadata
    metadata JSONB DEFAULT '{}',
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    invited_at TIMESTAMPTZ,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    left_at TIMESTAMPTZ,

    -- Audit fields
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    deleted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- Composite primary key
    PRIMARY KEY (user_id, organization_id),

    -- Constraints
    CONSTRAINT user_organization_memberships_status_check CHECK (status IN ('active', 'pending_invite', 'suspended'))
);

-- Indexes
CREATE INDEX idx_user_organization_memberships_user_id ON user_organization_memberships(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_organization_memberships_organization_id ON user_organization_memberships(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_organization_memberships_status ON user_organization_memberships(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_organization_memberships_invited_by ON user_organization_memberships(invited_by) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_organization_memberships_role ON user_organization_memberships(organization_id, role) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_organization_memberships_active ON user_organization_memberships(organization_id, is_active) WHERE is_active = TRUE AND deleted_at IS NULL;

-- Trigger for updated_at
CREATE TRIGGER update_user_organization_memberships_updated_at BEFORE UPDATE ON user_organization_memberships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comment
COMMENT ON TABLE user_organization_memberships IS 'Bảng liên kết users với organizations - một user có thể thuộc nhiều organizations. Merged with team_members.';
COMMENT ON COLUMN user_organization_memberships.status IS 'Trạng thái membership: active (đang hoạt động), pending_invite (chờ chấp nhận), suspended (bị đình chỉ)';
COMMENT ON COLUMN user_organization_memberships.invited_by IS 'User đã mời thành viên này vào organization';
COMMENT ON COLUMN user_organization_memberships.role IS 'Vai trò trong organization: leader, translator, proofreader, editor, member';
COMMENT ON COLUMN user_organization_memberships.quality_score IS 'Điểm chất lượng trung bình từ reviewers (thang điểm 0-5)';

-- =====================================================
-- Table: user_organization_roles
-- Description: Gán RBAC roles cho user trong context của một organization cụ thể
-- =====================================================
CREATE TABLE user_organization_roles (
    user_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Composite primary key
    PRIMARY KEY (user_id, organization_id, role_id),

    -- Foreign key compound constraint
    FOREIGN KEY (user_id, organization_id)
        REFERENCES user_organization_memberships(user_id, organization_id)
        ON DELETE CASCADE

    -- Note: Scope validation phải được thực hiện ở application level
    -- hoặc trigger vì PostgreSQL không cho phép subquery trong CHECK constraint
);

-- Indexes
CREATE INDEX idx_user_organization_roles_user_id ON user_organization_roles(user_id);
CREATE INDEX idx_user_organization_roles_organization_id ON user_organization_roles(organization_id);
CREATE INDEX idx_user_organization_roles_role_id ON user_organization_roles(role_id);
CREATE INDEX idx_user_organization_roles_user_organization ON user_organization_roles(user_id, organization_id);

-- Comment
COMMENT ON TABLE user_organization_roles IS 'Gán RBAC roles cho user trong context của một organization cụ thể. QUAN TRỌNG: Application phải đảm bảo chỉ gán organization-scoped roles';

-- =====================================================
-- Table: user_global_roles
-- Description: Gán global roles cho user (độc lập với organization)
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
COMMENT ON TABLE user_global_roles IS 'Gán global roles cho user, độc lập với organization. QUAN TRỌNG: Application phải đảm bảo chỉ gán global-scoped roles';

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

-- ORGANIZATION_ADMIN: Tất cả organization permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'ORGANIZATION_ADMIN'),
    id
FROM permissions
WHERE scope = 'organization';

-- ORGANIZATION_MANAGER: Một số organization permissions (không bao gồm role:manage)
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'ORGANIZATION_MANAGER'),
    id
FROM permissions
WHERE scope = 'organization'
    AND name IN (
        'user:view', 'user:create', 'user:update',
        'content:view', 'content:create', 'content:update', 'content:delete'
    );

-- ORGANIZATION_MEMBER: Permissions cơ bản
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'ORGANIZATION_MEMBER'),
    id
FROM permissions
WHERE name IN ('content:view', 'content:create', 'content:update');

-- ORGANIZATION_VIEWER: Chỉ view
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'ORGANIZATION_VIEWER'),
    id
FROM permissions
WHERE name IN ('content:view', 'user:view');

-- =====================================================
-- Helper Functions: Permission Check
-- Description: Functions hỗ trợ kiểm tra permissions
-- =====================================================

-- Function: Kiểm tra user có permission trong organization không
CREATE OR REPLACE FUNCTION user_has_organization_permission(
    p_user_id UUID,
    p_organization_id UUID,
    p_permission_name VARCHAR
)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM user_organization_roles uor
        JOIN role_permissions rp ON uor.role_id = rp.role_id
        JOIN permissions p ON rp.permission_id = p.id
        WHERE uor.user_id = p_user_id
            AND uor.organization_id = p_organization_id
            AND p.name = p_permission_name
    );
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION user_has_organization_permission IS 'Kiểm tra xem user có permission cụ thể trong organization không';

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

-- Function: Lấy tất cả permissions của user trong organization
CREATE OR REPLACE FUNCTION get_user_organization_permissions(
    p_user_id UUID,
    p_organization_id UUID
)
RETURNS TABLE(permission_name VARCHAR, permission_description TEXT) AS $$
BEGIN
    RETURN QUERY
    SELECT DISTINCT p.name, p.description
    FROM user_organization_roles uor
    JOIN role_permissions rp ON uor.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE uor.user_id = p_user_id
        AND uor.organization_id = p_organization_id
    ORDER BY p.name;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION get_user_organization_permissions IS 'Lấy danh sách tất cả permissions của user trong organization';

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
