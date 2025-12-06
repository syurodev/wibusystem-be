-- Migration: Create Core Tables (Organizations, Users, Permissions, Roles)
-- Description: Tạo các bảng cốt lõi cho hệ thống multi-organization với RBAC
-- Author: System
-- Created: 2025-10-26

-- Enable UUID extension nếu chưa có (PostgreSQL 18 đã có sẵn uuidv7)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- Table: users
-- Description: Bảng toàn cục cho tất cả tài khoản người dùng
-- =====================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    password_hash VARCHAR(255) NOT NULL,

    -- Profile information
    full_name VARCHAR(255),
    avatar_url TEXT,
    phone VARCHAR(50),

    -- Account status
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, suspended, deleted

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT users_status_check CHECK (status IN ('active', 'suspended', 'deleted'))
);

-- Indexes cho users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_last_login_at ON users(last_login_at);

-- Comment
COMMENT ON TABLE users IS 'Bảng toàn cục chứa tất cả người dùng trong hệ thống';
COMMENT ON COLUMN users.email_verified IS 'Trạng thái xác thực email';
COMMENT ON COLUMN users.password_hash IS 'Mật khẩu đã được hash (bcrypt hoặc argon2)';

-- =====================================================
-- Table: organizations
-- Description: Lưu trữ thông tin về các tổ chức (organizations/translation teams)
-- =====================================================
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, suspended, archived
    settings JSONB,

    -- Translation team fields (merged from migration 000014)
    description JSONB, -- Plate editor JSON output
    avatar_url VARCHAR(1000),

    -- Translation capabilities
    is_recruiting BOOLEAN NOT NULL DEFAULT FALSE,
    can_translate BOOLEAN NOT NULL DEFAULT TRUE,
    can_proofread BOOLEAN NOT NULL DEFAULT TRUE,
    can_edit BOOLEAN NOT NULL DEFAULT TRUE,

    -- Statistics (auto-updated by application)
    member_count INTEGER NOT NULL DEFAULT 0 CHECK (member_count >= 0),
    active_projects INTEGER NOT NULL DEFAULT 0 CHECK (active_projects >= 0),
    completed_translations INTEGER NOT NULL DEFAULT 0 CHECK (completed_translations >= 0),

    -- Metadata & Audit
    metadata JSONB DEFAULT '{}',
    created_by UUID REFERENCES users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    deleted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT organizations_status_check CHECK (status IN ('active', 'suspended', 'archived'))
);

-- Indexes
CREATE INDEX idx_organizations_slug ON organizations(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_organizations_status ON organizations(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_organizations_created_at ON organizations(created_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_organizations_recruiting ON organizations(is_recruiting) WHERE is_recruiting = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_organizations_metadata ON organizations USING GIN(metadata) WHERE deleted_at IS NULL;
CREATE INDEX idx_organizations_description ON organizations USING GIN(description) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE organizations IS 'Bảng lưu trữ thông tin về các tổ chức (organizations/translation teams) trong hệ thống';
COMMENT ON COLUMN organizations.slug IS 'URL-friendly identifier cho organization';
COMMENT ON COLUMN organizations.settings IS 'Cấu hình tùy chỉnh cho organization (JSONB format)';
COMMENT ON COLUMN organizations.description IS 'Mô tả organization (Plate editor JSON output)';
COMMENT ON COLUMN organizations.can_translate IS 'Quyền dịch nội dung';
COMMENT ON COLUMN organizations.can_proofread IS 'Quyền hiệu đính bản dịch';
COMMENT ON COLUMN organizations.can_edit IS 'Quyền chỉnh sửa bản dịch';

-- =====================================================
-- Type: permission_scope
-- Description: Phạm vi của permission (global hoặc organization-specific)
-- =====================================================
CREATE TYPE permission_scope AS ENUM ('global', 'organization');

COMMENT ON TYPE permission_scope IS 'Phạm vi của permission: global (toàn hệ thống) hoặc organization (trong organization)';

-- =====================================================
-- Table: permissions
-- Description: Danh sách master của tất cả các quyền trong hệ thống
-- =====================================================
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL, -- e.g., 'user:view_self', 'content:create_anime'
    scope permission_scope NOT NULL,
    description TEXT,
    resource VARCHAR(100), -- e.g., 'user', 'content', 'organization'
    action VARCHAR(100), -- e.g., 'view', 'create', 'update', 'delete'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT permissions_name_format_check CHECK (name ~ '^[a-z_]+:[a-z_]+$')
);

-- Indexes cho permissions
CREATE INDEX idx_permissions_scope ON permissions(scope);
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_name ON permissions(name);

-- Comment
COMMENT ON TABLE permissions IS 'Danh sách master của tất cả permissions trong hệ thống';
COMMENT ON COLUMN permissions.name IS 'Tên unique của permission, format: resource:action (e.g., user:create)';
COMMENT ON COLUMN permissions.scope IS 'Phạm vi áp dụng: global hoặc tenant';
COMMENT ON COLUMN permissions.resource IS 'Resource type mà permission này áp dụng';
COMMENT ON COLUMN permissions.action IS 'Hành động được phép (view, create, update, delete, etc.)';

-- =====================================================
-- Type: role_scope
-- Description: Phạm vi của role (global hoặc organization-specific)
-- =====================================================
CREATE TYPE role_scope AS ENUM ('global', 'organization');

COMMENT ON TYPE role_scope IS 'Phạm vi của role: global (toàn hệ thống) hoặc organization (trong organization)';

-- =====================================================
-- Table: roles
-- Description: Danh sách master của tất cả các vai trò
-- =====================================================
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL, -- e.g., 'SUPER_ADMIN', 'ORGANIZATION_ADMIN', 'USER'
    slug VARCHAR(255) UNIQUE NOT NULL, -- URL-friendly version
    scope role_scope NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE nếu là system role (không thể xóa)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT roles_name_format_check CHECK (name ~ '^[A-Z_]+$')
);

-- Indexes cho roles
CREATE INDEX idx_roles_scope ON roles(scope);
CREATE INDEX idx_roles_slug ON roles(slug);
CREATE INDEX idx_roles_is_system ON roles(is_system);

-- Comment
COMMENT ON TABLE roles IS 'Danh sách master của tất cả roles trong hệ thống';
COMMENT ON COLUMN roles.name IS 'Tên unique của role (UPPER_SNAKE_CASE)';
COMMENT ON COLUMN roles.slug IS 'URL-friendly identifier cho role';
COMMENT ON COLUMN roles.scope IS 'Phạm vi áp dụng: global hoặc tenant';
COMMENT ON COLUMN roles.is_system IS 'System role không thể xóa hoặc sửa đổi';

-- =====================================================
-- Trigger: Updated timestamp
-- Description: Tự động update updated_at khi record thay đổi
-- =====================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger cho các tables có updated_at
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- Seed Data: Default Permissions
-- Description: Tạo các permissions cơ bản cho hệ thống
-- =====================================================

-- Global Permissions (quản lý toàn hệ thống)
INSERT INTO permissions (name, scope, description, resource, action) VALUES
    ('system:manage_all', 'global', 'Quản lý toàn bộ hệ thống', 'system', 'manage_all'),
    ('organization:create', 'global', 'Tạo organization mới', 'organization', 'create'),
    ('organization:view_all', 'global', 'Xem tất cả organizations', 'organization', 'view_all'),
    ('organization:delete', 'global', 'Xóa organization', 'organization', 'delete'),
    ('user:view_all', 'global', 'Xem tất cả users trong hệ thống', 'user', 'view_all');

-- Organization-level Permissions (trong phạm vi organization)
INSERT INTO permissions (name, scope, description, resource, action) VALUES
    ('organization:manage', 'organization', 'Quản lý cấu hình organization', 'organization', 'manage'),
    ('user:view', 'organization', 'Xem danh sách users trong organization', 'user', 'view'),
    ('user:create', 'organization', 'Tạo user mới trong organization', 'user', 'create'),
    ('user:update', 'organization', 'Cập nhật thông tin user', 'user', 'update'),
    ('user:delete', 'organization', 'Xóa user khỏi organization', 'user', 'delete'),
    ('role:view', 'organization', 'Xem danh sách roles', 'role', 'view'),
    ('role:manage', 'organization', 'Quản lý roles trong organization', 'role', 'manage'),
    ('content:view', 'organization', 'Xem nội dung', 'content', 'view'),
    ('content:create', 'organization', 'Tạo nội dung mới', 'content', 'create'),
    ('content:update', 'organization', 'Cập nhật nội dung', 'content', 'update'),
    ('content:delete', 'organization', 'Xóa nội dung', 'content', 'delete');

-- =====================================================
-- Seed Data: Default Roles
-- Description: Tạo các roles cơ bản cho hệ thống
-- =====================================================

-- Global Roles (toàn hệ thống)
INSERT INTO roles (name, slug, scope, description, is_system) VALUES
    ('SUPER_ADMIN', 'super-admin', 'global', 'Quản trị viên tối cao của toàn hệ thống', TRUE),
    ('PLATFORM_ADMIN', 'platform-admin', 'global', 'Quản trị viên nền tảng', TRUE);

-- Organization-level Roles
INSERT INTO roles (name, slug, scope, description, is_system) VALUES
    ('ORGANIZATION_ADMIN', 'organization-admin', 'organization', 'Quản trị viên của organization', TRUE),
    ('ORGANIZATION_MANAGER', 'organization-manager', 'organization', 'Người quản lý trong organization', TRUE),
    ('ORGANIZATION_MEMBER', 'organization-member', 'organization', 'Thành viên thông thường trong organization', TRUE),
    ('ORGANIZATION_VIEWER', 'organization-viewer', 'organization', 'Người xem (chỉ đọc) trong organization', TRUE);

COMMENT ON TABLE permissions IS 'Seeded với permissions cơ bản cho system và organization operations';
COMMENT ON TABLE roles IS 'Seeded với roles cơ bản: SUPER_ADMIN, PLATFORM_ADMIN, ORGANIZATION_ADMIN, ORGANIZATION_MANAGER, ORGANIZATION_MEMBER, ORGANIZATION_VIEWER';
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
-- Migration: Create OAuth2 Tables
-- Description: Tạo các bảng cho OAuth 2.0 Authorization Server (Fosite-compatible)
-- Author: System
-- Created: 2025-10-26

-- =====================================================
-- Table: oauth2_clients
-- Description: Lưu trữ thông tin về OAuth 2.0 clients đã đăng ký
-- =====================================================
CREATE TABLE oauth2_clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_name VARCHAR(255) NOT NULL,

    -- Secret (hashed)
    secret_hash VARCHAR(255) NOT NULL,

    -- URIs và configurations (stored as arrays)
    redirect_uris TEXT[] NOT NULL DEFAULT '{}',
    grant_types TEXT[] NOT NULL DEFAULT '{}',
    response_types TEXT[] NOT NULL DEFAULT '{}',
    scopes TEXT[] NOT NULL DEFAULT '{}',

    -- Client type
    is_public BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE for public clients (mobile, SPA)

    -- Multi-organization support
    -- NULL = global/first-party client (e.g., admin dashboard)
    -- NOT NULL = organization-specific client
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,

    -- Metadata
    owner_user_id UUID REFERENCES users(id) ON DELETE SET NULL, -- User who created this client
    logo_url TEXT,
    terms_of_service_url TEXT,
    policy_url TEXT,
    client_uri TEXT, -- Homepage của client application

    -- Security settings
    token_endpoint_auth_method VARCHAR(50) NOT NULL DEFAULT 'client_secret_basic',
    -- Options: client_secret_basic, client_secret_post, none (for public clients)

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT oauth2_clients_auth_method_check CHECK (
        token_endpoint_auth_method IN ('client_secret_basic', 'client_secret_post', 'client_secret_jwt', 'private_key_jwt', 'none')
    ),
    CONSTRAINT oauth2_clients_public_check CHECK (
        (is_public = TRUE AND token_endpoint_auth_method = 'none') OR
        (is_public = FALSE)
    )
);

-- Indexes
CREATE INDEX idx_oauth2_clients_organization_id ON oauth2_clients(organization_id);
CREATE INDEX idx_oauth2_clients_owner_user_id ON oauth2_clients(owner_user_id);
CREATE INDEX idx_oauth2_clients_is_public ON oauth2_clients(is_public);

-- Trigger for updated_at
CREATE TRIGGER update_oauth2_clients_updated_at BEFORE UPDATE ON oauth2_clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comment
COMMENT ON TABLE oauth2_clients IS 'Lưu trữ OAuth 2.0 clients đã đăng ký - hỗ trợ cả global và organization-specific clients';
COMMENT ON COLUMN oauth2_clients.organization_id IS 'NULL = global client (first-party), NOT NULL = organization-specific client (third-party)';
COMMENT ON COLUMN oauth2_clients.is_public IS 'TRUE cho public clients (mobile apps, SPAs) không có secret';
COMMENT ON COLUMN oauth2_clients.secret_hash IS 'Hashed client secret (bcrypt hoặc argon2)';
COMMENT ON COLUMN oauth2_clients.redirect_uris IS 'Danh sách redirect URIs được phép (OAuth 2.0 callback)';
COMMENT ON COLUMN oauth2_clients.grant_types IS 'Grant types được phép: authorization_code, refresh_token, client_credentials, etc.';

-- =====================================================
-- Type: session_type
-- Description: Các loại session/token được lưu trữ
-- =====================================================
CREATE TYPE session_type AS ENUM (
    'authorize_code',    -- Authorization codes
    'access_token',      -- Access tokens
    'refresh_token',     -- Refresh tokens
    'pkce',             -- PKCE sessions
    'openid'            -- OpenID Connect sessions
);

COMMENT ON TYPE session_type IS 'Loại session/token trong OAuth 2.0 flow';

-- =====================================================
-- Table: oauth2_sessions
-- Description: Bảng chung lưu trữ tất cả OAuth 2.0 sessions và tokens
-- =====================================================
CREATE TABLE oauth2_sessions (
    -- Primary identifier
    signature VARCHAR(255) PRIMARY KEY,

    -- Session metadata
    request_id VARCHAR(255) NOT NULL,
    session_type session_type NOT NULL,

    -- Status
    active BOOLEAN NOT NULL DEFAULT TRUE,

    -- Session data (Fosite Requester serialized to JSON)
    session_data JSONB NOT NULL,

    -- Expiration
    expires_at TIMESTAMPTZ NOT NULL,

    -- References
    client_id UUID NOT NULL REFERENCES oauth2_clients(id) ON DELETE CASCADE,
    subject_id UUID REFERENCES users(id) ON DELETE CASCADE, -- NULL for client_credentials grant

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_oauth2_sessions_request_id ON oauth2_sessions(request_id);
CREATE INDEX idx_oauth2_sessions_session_type ON oauth2_sessions(session_type);
CREATE INDEX idx_oauth2_sessions_expires_at ON oauth2_sessions(expires_at);
CREATE INDEX idx_oauth2_sessions_client_id ON oauth2_sessions(client_id);
CREATE INDEX idx_oauth2_sessions_subject_id ON oauth2_sessions(subject_id);
CREATE INDEX idx_oauth2_sessions_active ON oauth2_sessions(active);

-- Index cho cleanup expired sessions
CREATE INDEX idx_oauth2_sessions_cleanup ON oauth2_sessions(expires_at, active);

-- GIN index cho JSONB queries (nếu cần query vào session_data)
CREATE INDEX idx_oauth2_sessions_session_data ON oauth2_sessions USING GIN (session_data);

-- Comment
COMMENT ON TABLE oauth2_sessions IS 'Bảng chung lưu trữ tất cả OAuth 2.0 sessions, codes, và tokens';
COMMENT ON COLUMN oauth2_sessions.signature IS 'Unique signature của token/code (thường là hash)';
COMMENT ON COLUMN oauth2_sessions.request_id IS 'ID của authorization request';
COMMENT ON COLUMN oauth2_sessions.session_data IS 'Fosite Requester object serialized thành JSON';
COMMENT ON COLUMN oauth2_sessions.subject_id IS 'User ID (NULL cho client_credentials grant)';

-- =====================================================
-- Table: oauth2_jti_blacklist
-- Description: Blacklist cho revoked tokens (JWT Token Identifier)
-- =====================================================
CREATE TABLE oauth2_jti_blacklist (
    signature VARCHAR(255) PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index
CREATE INDEX idx_oauth2_jti_blacklist_expires_at ON oauth2_jti_blacklist(expires_at);

-- Comment
COMMENT ON TABLE oauth2_jti_blacklist IS 'Blacklist cho revoked tokens - ngăn chặn replay attacks';
COMMENT ON COLUMN oauth2_jti_blacklist.signature IS 'Token signature đã bị revoke';
COMMENT ON COLUMN oauth2_jti_blacklist.expires_at IS 'Thời điểm token hết hạn (có thể xóa khỏi blacklist sau đó)';

-- =====================================================
-- Cleanup Function: Xóa expired sessions
-- Description: Function để cleanup sessions và blacklist entries đã hết hạn
-- =====================================================
CREATE OR REPLACE FUNCTION cleanup_expired_oauth2_data()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete expired sessions
    WITH deleted AS (
        DELETE FROM oauth2_sessions
        WHERE expires_at < NOW()
        RETURNING 1
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;

    -- Delete expired blacklist entries
    DELETE FROM oauth2_jti_blacklist
    WHERE expires_at < NOW();

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_expired_oauth2_data IS 'Cleanup expired OAuth2 sessions và blacklist entries. Nên chạy định kỳ (cron job).';

-- =====================================================
-- Helper Functions: OAuth2 Operations
-- =====================================================

-- Function: Kiểm tra token có bị revoked không
CREATE OR REPLACE FUNCTION is_token_revoked(
    p_signature VARCHAR
)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM oauth2_jti_blacklist
        WHERE signature = p_signature
            AND expires_at > NOW()
    );
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION is_token_revoked IS 'Kiểm tra xem token có trong blacklist không';

-- Function: Revoke token
CREATE OR REPLACE FUNCTION revoke_token(
    p_signature VARCHAR,
    p_expires_at TIMESTAMPTZ
)
RETURNS VOID AS $$
BEGIN
    -- Add to blacklist
    INSERT INTO oauth2_jti_blacklist (signature, expires_at)
    VALUES (p_signature, p_expires_at)
    ON CONFLICT (signature) DO NOTHING;

    -- Mark session as inactive
    UPDATE oauth2_sessions
    SET active = FALSE
    WHERE signature = p_signature;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION revoke_token IS 'Revoke một token bằng cách thêm vào blacklist và mark session inactive';

-- Function: Revoke all tokens của một user trong một client
CREATE OR REPLACE FUNCTION revoke_user_client_tokens(
    p_user_id UUID,
    p_client_id UUID
)
RETURNS INTEGER AS $$
DECLARE
    revoked_count INTEGER;
BEGIN
    -- Mark sessions as inactive
    WITH updated AS (
        UPDATE oauth2_sessions
        SET active = FALSE
        WHERE subject_id = p_user_id
            AND client_id = p_client_id
            AND active = TRUE
        RETURNING signature, expires_at
    )
    -- Add to blacklist
    INSERT INTO oauth2_jti_blacklist (signature, expires_at)
    SELECT signature, expires_at FROM updated
    ON CONFLICT (signature) DO NOTHING;

    GET DIAGNOSTICS revoked_count = ROW_COUNT;
    RETURN revoked_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION revoke_user_client_tokens IS 'Revoke tất cả tokens của một user cho một client cụ thể';

-- Function: Revoke all tokens của một user (logout từ tất cả clients)
CREATE OR REPLACE FUNCTION revoke_all_user_tokens(
    p_user_id UUID
)
RETURNS INTEGER AS $$
DECLARE
    revoked_count INTEGER;
BEGIN
    WITH updated AS (
        UPDATE oauth2_sessions
        SET active = FALSE
        WHERE subject_id = p_user_id
            AND active = TRUE
        RETURNING signature, expires_at
    )
    INSERT INTO oauth2_jti_blacklist (signature, expires_at)
    SELECT signature, expires_at FROM updated
    ON CONFLICT (signature) DO NOTHING;

    GET DIAGNOSTICS revoked_count = ROW_COUNT;
    RETURN revoked_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION revoke_all_user_tokens IS 'Revoke tất cả tokens của một user (global logout)';

-- =====================================================
-- Seed Data: Demo OAuth2 Client
-- Description: Tạo một demo client cho testing
-- =====================================================

-- Insert a demo first-party client (tenant_id = NULL)
INSERT INTO oauth2_clients (
    id,
    client_name,
    secret_hash,
    redirect_uris,
    grant_types,
    response_types,
    scopes,
    is_public,
    organization_id,
    token_endpoint_auth_method
) VALUES (
    gen_random_uuid(),
    'System Admin Dashboard',
    '$2a$10$demo.hash.for.testing.only', -- In production, use proper bcrypt hash
    ARRAY['http://localhost:3000/callback', 'http://localhost:3000/auth/callback'],
    ARRAY['authorization_code', 'refresh_token'],
    ARRAY['code'],
    ARRAY['openid', 'profile', 'email', 'offline_access'],
    FALSE, -- Confidential client
    NULL, -- Global first-party client
    'client_secret_basic'
);

COMMENT ON TABLE oauth2_clients IS 'Seeded với demo admin dashboard client';
-- Migration: Create Schemas and Move Tables
-- Description: Tạo multi-schema architecture và di chuyển tables vào đúng schemas
-- Author: System
-- Created: 2025-10-26

-- =====================================================
-- Create Schemas
-- Description: Tạo các schemas cho từng domain
-- =====================================================

CREATE SCHEMA IF NOT EXISTS identify;
COMMENT ON SCHEMA identify IS 'Authentication, Authorization, User Management, và OAuth 2.0';

CREATE SCHEMA IF NOT EXISTS catalog;
COMMENT ON SCHEMA catalog IS 'Product Catalog, Inventory, Content Management';

CREATE SCHEMA IF NOT EXISTS community;
COMMENT ON SCHEMA community IS 'Social Features, Posts, Comments, Reactions';

CREATE SCHEMA IF NOT EXISTS payment;
COMMENT ON SCHEMA payment IS 'Payment Processing, Transactions, Invoicing';

-- =====================================================
-- Move Tables to Identify Schema
-- Description: Di chuyển tất cả tables hiện tại vào identify schema
-- =====================================================

-- Core tables
ALTER TABLE organizations SET SCHEMA identify;
ALTER TABLE users SET SCHEMA identify;
ALTER TABLE permissions SET SCHEMA identify;
ALTER TABLE roles SET SCHEMA identify;

-- RBAC relation tables
ALTER TABLE role_permissions SET SCHEMA identify;
ALTER TABLE user_organization_memberships SET SCHEMA identify;
ALTER TABLE user_organization_roles SET SCHEMA identify;
ALTER TABLE user_global_roles SET SCHEMA identify;

-- OAuth2 tables
ALTER TABLE oauth2_clients SET SCHEMA identify;
ALTER TABLE oauth2_sessions SET SCHEMA identify;
ALTER TABLE oauth2_jti_blacklist SET SCHEMA identify;

-- =====================================================
-- Update Search Path (Optional)
-- Description: Set default search path để không cần qualify schema name
-- =====================================================

-- Option 1: Set cho database (affects all connections)
-- ALTER DATABASE system_dev SET search_path TO identify, public;

-- Option 2: Set cho specific user
-- ALTER USER system_dev SET search_path TO identify, public;

-- Note: Application code nên explicitly qualify schema names
-- Example: SELECT * FROM identify.users
-- Nhưng có thể set search_path để backward compatible

-- =====================================================
-- Verify Schema Migration
-- Description: Query để verify tables đã được move
-- =====================================================

-- Query này sẽ fail nếu tables chưa được move vào identify schema
DO $$
DECLARE
    table_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO table_count
    FROM information_schema.tables
    WHERE table_schema = 'identify';

    IF table_count < 11 THEN
        RAISE EXCEPTION 'Schema migration incomplete. Expected at least 11 tables in identify schema, found %', table_count;
    END IF;

    RAISE NOTICE 'Schema migration successful. % tables in identify schema', table_count;
END $$;

COMMENT ON SCHEMA identify IS 'Migration 000004: All authentication/authorization tables moved to this schema';
-- Migration: Move Helper Functions to Identify Schema
-- Description: Di chuyển tất cả helper functions vào identify schema để đồng nhất với tables
-- Author: System
-- Created: 2025-10-26

-- =====================================================
-- Move RBAC Helper Functions
-- Description: Di chuyển functions kiểm tra permissions
-- =====================================================

ALTER FUNCTION user_has_organization_permission(UUID, UUID, VARCHAR) SET SCHEMA identify;
ALTER FUNCTION user_has_global_permission(UUID, VARCHAR) SET SCHEMA identify;
ALTER FUNCTION get_user_organization_permissions(UUID, UUID) SET SCHEMA identify;
ALTER FUNCTION get_user_global_permissions(UUID) SET SCHEMA identify;

-- =====================================================
-- Move OAuth2 Helper Functions
-- Description: Di chuyển functions quản lý OAuth2 tokens
-- =====================================================

ALTER FUNCTION cleanup_expired_oauth2_data() SET SCHEMA identify;
ALTER FUNCTION is_token_revoked(VARCHAR) SET SCHEMA identify;
ALTER FUNCTION revoke_token(VARCHAR, TIMESTAMPTZ) SET SCHEMA identify;
ALTER FUNCTION revoke_user_client_tokens(UUID, UUID) SET SCHEMA identify;
ALTER FUNCTION revoke_all_user_tokens(UUID) SET SCHEMA identify;

-- =====================================================
-- Verify Function Migration
-- Description: Kiểm tra tất cả functions đã được move
-- =====================================================

DO $$
DECLARE
    func_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO func_count
    FROM pg_proc p
    JOIN pg_namespace n ON p.pronamespace = n.oid
    WHERE n.nspname = 'identify'
        AND p.proname IN (
            'user_has_organization_permission',
            'user_has_global_permission',
            'get_user_organization_permissions',
            'get_user_global_permissions',
            'cleanup_expired_oauth2_data',
            'is_token_revoked',
            'revoke_token',
            'revoke_user_client_tokens',
            'revoke_all_user_tokens'
        );

    IF func_count < 9 THEN
        RAISE EXCEPTION 'Function migration incomplete. Expected 9 functions in identify schema, found %', func_count;
    END IF;

    RAISE NOTICE 'Function migration successful. % functions in identify schema', func_count;
END $$;

COMMENT ON SCHEMA identify IS 'Migration 000005: All helper functions moved to this schema';
-- Migration: Fix Function Search Paths
-- Description: Set search_path cho functions trong identify schema để tìm được tables
-- Author: System
-- Created: 2025-10-26

-- =====================================================
-- Set Search Path for RBAC Functions
-- Description: Functions cần tìm tables trong identify schema
-- =====================================================

ALTER FUNCTION identify.user_has_organization_permission(UUID, UUID, VARCHAR) SET search_path = identify, public;
ALTER FUNCTION identify.user_has_global_permission(UUID, VARCHAR) SET search_path = identify, public;
ALTER FUNCTION identify.get_user_organization_permissions(UUID, UUID) SET search_path = identify, public;
ALTER FUNCTION identify.get_user_global_permissions(UUID) SET search_path = identify, public;

-- =====================================================
-- Set Search Path for OAuth2 Functions
-- Description: Functions cần tìm tables trong identify schema
-- =====================================================

ALTER FUNCTION identify.cleanup_expired_oauth2_data() SET search_path = identify, public;
ALTER FUNCTION identify.is_token_revoked(VARCHAR) SET search_path = identify, public;
ALTER FUNCTION identify.revoke_token(VARCHAR, TIMESTAMPTZ) SET search_path = identify, public;
ALTER FUNCTION identify.revoke_user_client_tokens(UUID, UUID) SET search_path = identify, public;
ALTER FUNCTION identify.revoke_all_user_tokens(UUID) SET search_path = identify, public;

-- =====================================================
-- Verify Function Configuration
-- Description: Kiểm tra search_path đã được set
-- =====================================================

DO $$
DECLARE
    func_count INTEGER;
BEGIN
    -- Count functions with correct search_path
    SELECT COUNT(*) INTO func_count
    FROM pg_proc p
    JOIN pg_namespace n ON p.pronamespace = n.oid
    WHERE n.nspname = 'identify'
        AND p.proname IN (
            'user_has_organization_permission',
            'user_has_global_permission',
            'get_user_organization_permissions',
            'get_user_global_permissions',
            'cleanup_expired_oauth2_data',
            'is_token_revoked',
            'revoke_token',
            'revoke_user_client_tokens',
            'revoke_all_user_tokens'
        )
        AND p.proconfig IS NOT NULL;  -- proconfig chứa SET options

    IF func_count < 9 THEN
        RAISE WARNING 'Some functions may not have search_path set. Found % functions with configuration', func_count;
    END IF;

    RAISE NOTICE 'Function search_path configuration complete. % functions configured', func_count;
END $$;

COMMENT ON SCHEMA identify IS 'Migration 000006: Function search paths configured for identify schema';
-- Migration: Create OAuth2 Consents Table
-- Description: Tạo bảng lưu trữ user consents cho OAuth2 clients
-- Author: System
-- Created: 2025-11-01

-- =====================================================
-- Table: oauth2_consents
-- Description: Lưu trữ thông tin về quyền mà user đã cấp cho các OAuth2 clients
-- =====================================================
CREATE TABLE identify.oauth2_consents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- User và Client
    user_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES identify.oauth2_clients(id) ON DELETE CASCADE,

    -- Scopes được cấp phép
    granted_scopes TEXT[] NOT NULL DEFAULT '{}',

    -- Status
    revoked BOOLEAN NOT NULL DEFAULT FALSE,

    -- Timestamps
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ, -- NULL = không bao giờ hết hạn (persistent consent)

    -- Metadata
    consent_method VARCHAR(50) NOT NULL DEFAULT 'explicit', -- explicit, implicit, remembered
    ip_address INET,
    user_agent TEXT,

    -- Constraints
    CONSTRAINT oauth2_consents_user_client_unique UNIQUE (user_id, client_id),
    CONSTRAINT oauth2_consents_method_check CHECK (
        consent_method IN ('explicit', 'implicit', 'remembered')
    )
);

-- Indexes
CREATE INDEX idx_oauth2_consents_user_id ON identify.oauth2_consents(user_id);
CREATE INDEX idx_oauth2_consents_client_id ON identify.oauth2_consents(client_id);
CREATE INDEX idx_oauth2_consents_revoked ON identify.oauth2_consents(revoked);
CREATE INDEX idx_oauth2_consents_granted_at ON identify.oauth2_consents(granted_at);
CREATE INDEX idx_oauth2_consents_expires_at ON identify.oauth2_consents(expires_at);

-- Index cho cleanup expired consents
CREATE INDEX idx_oauth2_consents_cleanup ON identify.oauth2_consents(expires_at, revoked)
WHERE expires_at IS NOT NULL;

-- Trigger for updated_at (nếu cần track updates)
-- Note: Bảng này không có updated_at vì consent là immutable - chỉ có thể revoke

-- Comments
COMMENT ON TABLE identify.oauth2_consents IS 'Lưu trữ user consents cho OAuth2 clients - quản lý quyền truy cập';
COMMENT ON COLUMN identify.oauth2_consents.user_id IS 'User đã cấp quyền';
COMMENT ON COLUMN identify.oauth2_consents.client_id IS 'Client được cấp quyền';
COMMENT ON COLUMN identify.oauth2_consents.granted_scopes IS 'Danh sách scopes đã được user chấp thuận';
COMMENT ON COLUMN identify.oauth2_consents.revoked IS 'TRUE nếu consent đã bị thu hồi';
COMMENT ON COLUMN identify.oauth2_consents.consent_method IS 'Phương thức consent: explicit (user clicked allow), implicit (trusted first-party), remembered (previous consent)';
COMMENT ON COLUMN identify.oauth2_consents.expires_at IS 'Thời điểm consent hết hạn (NULL = persistent)';

-- =====================================================
-- Functions: Consent Management
-- =====================================================

-- Function: Cleanup expired consents
CREATE OR REPLACE FUNCTION identify.cleanup_expired_consents()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    WITH deleted AS (
        DELETE FROM identify.oauth2_consents
        WHERE expires_at IS NOT NULL
            AND expires_at < NOW()
            AND revoked = FALSE
        RETURNING 1
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION identify.cleanup_expired_consents IS 'Xóa các consents đã hết hạn - nên chạy định kỳ';

-- Function: Revoke consent
CREATE OR REPLACE FUNCTION identify.revoke_consent(
    p_user_id UUID,
    p_client_id UUID
)
RETURNS BOOLEAN AS $$
DECLARE
    rows_affected INTEGER;
BEGIN
    UPDATE identify.oauth2_consents
    SET revoked = TRUE,
        revoked_at = NOW()
    WHERE user_id = p_user_id
        AND client_id = p_client_id
        AND revoked = FALSE;

    GET DIAGNOSTICS rows_affected = ROW_COUNT;
    RETURN rows_affected > 0;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION identify.revoke_consent IS 'Thu hồi consent của user cho một client';

-- Function: Revoke all consents for a user
CREATE OR REPLACE FUNCTION identify.revoke_all_user_consents(
    p_user_id UUID
)
RETURNS INTEGER AS $$
DECLARE
    rows_affected INTEGER;
BEGIN
    UPDATE identify.oauth2_consents
    SET revoked = TRUE,
        revoked_at = NOW()
    WHERE user_id = p_user_id
        AND revoked = FALSE;

    GET DIAGNOSTICS rows_affected = ROW_COUNT;
    RETURN rows_affected;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION identify.revoke_all_user_consents IS 'Thu hồi tất cả consents của một user';

-- Function: Get active consent
CREATE OR REPLACE FUNCTION identify.get_active_consent(
    p_user_id UUID,
    p_client_id UUID
)
RETURNS TABLE (
    id UUID,
    granted_scopes TEXT[],
    granted_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        c.id,
        c.granted_scopes,
        c.granted_at,
        c.expires_at
    FROM identify.oauth2_consents c
    WHERE c.user_id = p_user_id
        AND c.client_id = p_client_id
        AND c.revoked = FALSE
        AND (c.expires_at IS NULL OR c.expires_at > NOW());
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION identify.get_active_consent IS 'Lấy active consent của user cho một client';

-- =====================================================
-- Seed Data: Demo Consent (Optional)
-- =====================================================

-- Note: Không seed consents vì chúng nên được tạo thông qua OAuth2 flow
-- Migration: Add is_internal field to oauth2_clients
-- Description: Thêm trường is_internal để phân biệt client nội bộ và bên ngoài
-- Author: System
-- Created: 2025-11-02

-- =====================================================
-- Add is_internal column
-- Description:
--   TRUE = Internal client (có quyền truy cập đầy đủ, ít giới hạn)
--   FALSE = External client (bị giới hạn chức năng, rate limit cao hơn)
-- =====================================================
ALTER TABLE identify.oauth2_clients
ADD COLUMN is_internal BOOLEAN NOT NULL DEFAULT FALSE;

-- Index for filtering internal/external clients
CREATE INDEX idx_oauth2_clients_is_internal ON identify.oauth2_clients(is_internal);

-- Comment
COMMENT ON COLUMN identify.oauth2_clients.is_internal IS 'TRUE = Internal client (full access), FALSE = External client (limited features)';

-- =====================================================
-- Update existing clients
-- Description: Mark clients with organization_id = NULL as internal (first-party clients)
-- =====================================================
UPDATE identify.oauth2_clients
SET is_internal = TRUE
WHERE organization_id IS NULL;

COMMENT ON COLUMN identify.oauth2_clients.is_internal IS 'TRUE = Internal/first-party client (full access), FALSE = External/third-party client (limited features). Internal clients typically have organization_id = NULL.';
-- Migration: Create email verification tokens table
-- Description: Tạo bảng lưu trữ token để xác thực email khi đăng ký
-- Author: System
-- Created: 2025-11-02

-- =====================================================
-- Table: email_verification_tokens
-- Description: Lưu trữ tokens để verify email address
-- =====================================================
CREATE TABLE identify.email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_email_verification_tokens_user_id ON identify.email_verification_tokens(user_id);
CREATE INDEX idx_email_verification_tokens_token ON identify.email_verification_tokens(token);
CREATE INDEX idx_email_verification_tokens_expires_at ON identify.email_verification_tokens(expires_at);

-- Comment
COMMENT ON TABLE identify.email_verification_tokens IS 'Lưu trữ tokens để xác thực email address khi đăng ký hoặc thay đổi email';
COMMENT ON COLUMN identify.email_verification_tokens.token IS 'Random token gửi qua email (hashed hoặc plain - tùy implementation)';
COMMENT ON COLUMN identify.email_verification_tokens.expires_at IS 'Token hết hạn sau 24 giờ';
COMMENT ON COLUMN identify.email_verification_tokens.used_at IS 'NULL = chưa sử dụng, NOT NULL = đã verify thành công';

-- =====================================================
-- Cleanup Function: Xóa expired tokens
-- =====================================================
CREATE OR REPLACE FUNCTION identify.cleanup_expired_verification_tokens()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    WITH deleted AS (
        DELETE FROM identify.email_verification_tokens
        WHERE expires_at < NOW()
        RETURNING 1
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION identify.cleanup_expired_verification_tokens IS 'Cleanup expired verification tokens. Nên chạy định kỳ (cron job).';
-- Migration: Create password reset tokens table
-- Description: Tạo bảng lưu trữ token để reset password
-- Author: System
-- Created: 2025-11-02

-- =====================================================
-- Table: password_reset_tokens
-- Description: Lưu trữ tokens để reset password
-- =====================================================
CREATE TABLE identify.password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_password_reset_tokens_user_id ON identify.password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_token ON identify.password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_expires_at ON identify.password_reset_tokens(expires_at);

-- Comment
COMMENT ON TABLE identify.password_reset_tokens IS 'Lưu trữ tokens để reset password';
COMMENT ON COLUMN identify.password_reset_tokens.token IS 'Random token gửi qua email';
COMMENT ON COLUMN identify.password_reset_tokens.expires_at IS 'Token hết hạn sau 1 giờ';
COMMENT ON COLUMN identify.password_reset_tokens.used_at IS 'NULL = chưa sử dụng, NOT NULL = đã reset thành công';

-- =====================================================
-- Cleanup Function: Xóa expired tokens
-- =====================================================
CREATE OR REPLACE FUNCTION identify.cleanup_expired_password_reset_tokens()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    WITH deleted AS (
        DELETE FROM identify.password_reset_tokens
        WHERE expires_at < NOW()
        RETURNING 1
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION identify.cleanup_expired_password_reset_tokens IS 'Cleanup expired password reset tokens. Nên chạy định kỳ (cron job).';
-- Migration: Add active column to oauth2_clients
-- Description: Thêm cột active để đánh dấu client có hoạt động hay không
-- Author: System
-- Created: 2025-11-02

-- =====================================================
-- Add active column
-- Description:
--   TRUE = Active client (có thể sử dụng)
--   FALSE = Inactive/Disabled client (bị vô hiệu hóa)
-- =====================================================
ALTER TABLE identify.oauth2_clients
ADD COLUMN active BOOLEAN NOT NULL DEFAULT TRUE;

-- Index for filtering active/inactive clients
CREATE INDEX idx_oauth2_clients_active ON identify.oauth2_clients(active);

-- Comment
COMMENT ON COLUMN identify.oauth2_clients.active IS 'TRUE = Active client, FALSE = Inactive/Disabled client';
-- =====================================================
-- Migration 000012: Core Novel System Tables
-- Description: novels, volumes, chapters with audit fields
-- =====================================================

-- Create catalog schema if not exists
CREATE SCHEMA IF NOT EXISTS catalog;

-- =====================================================
-- ENUM TYPES
-- =====================================================

-- Novel status
CREATE TYPE catalog.novel_status AS ENUM (
    'draft',
    'ongoing',
    'completed',
    'hiatus',
    'dropped'
);

-- Chapter status
CREATE TYPE catalog.chapter_status AS ENUM (
    'draft',
    'published',
    'scheduled'
);

-- =====================================================
-- NOVELS TABLE (Top Level)
-- =====================================================
CREATE TABLE catalog.novels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic info
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL UNIQUE,

    -- Ownership (Polymorphic - validated in application)
    owner_type VARCHAR(20) NOT NULL CHECK (owner_type IN ('user', 'tenant')),
    owner_id UUID NOT NULL,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Content
    synopsis JSONB, -- Synopsis in ORIGINAL language only

    -- Images
    cover_image_url VARCHAR(1000),
    thumbnail_url VARCHAR(1000),

    -- Original info
    original_language VARCHAR(10) NOT NULL, -- ISO 639-1: vi, en, zh, ja, ko
    original_title VARCHAR(500),

    -- Status
    status catalog.novel_status NOT NULL DEFAULT 'draft',

    -- Statistics (auto-updated by application)
    total_volumes INTEGER NOT NULL DEFAULT 0,
    total_chapters INTEGER NOT NULL DEFAULT 0,
    total_words BIGINT NOT NULL DEFAULT 0,
    view_count BIGINT NOT NULL DEFAULT 0,
    favorite_count INTEGER NOT NULL DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0.00 CHECK (rating_average >= 0 AND rating_average <= 5),
    rating_count INTEGER NOT NULL DEFAULT 0,

    -- Additional metadata
    metadata JSONB DEFAULT '{}',

    -- Dates
    first_published_at TIMESTAMP WITH TIME ZONE,
    last_chapter_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,

    -- Audit timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL
);

-- Indexes for novels
CREATE INDEX idx_novels_owner ON catalog.novels(owner_type, owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_created_by ON catalog.novels(created_by) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_status ON catalog.novels(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_original_language ON catalog.novels(original_language) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_slug ON catalog.novels(slug);
CREATE INDEX idx_novels_synopsis ON catalog.novels USING GIN(synopsis) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_metadata ON catalog.novels USING GIN(metadata) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_created_at ON catalog.novels(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_rating ON catalog.novels(rating_average DESC) WHERE status = 'ongoing' AND deleted_at IS NULL;
CREATE INDEX idx_novels_views ON catalog.novels(view_count DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.novels IS 'Top-level table storing novel information with polymorphic ownership';
COMMENT ON COLUMN catalog.novels.owner_type IS 'Polymorphic owner type: user or tenant';
COMMENT ON COLUMN catalog.novels.owner_id IS 'Reference to users.id OR tenants.id based on owner_type (validated in application)';
COMMENT ON COLUMN catalog.novels.created_by IS 'User who created the novel (never changes)';
COMMENT ON COLUMN catalog.novels.updated_by IS 'User who last updated the novel';
COMMENT ON COLUMN catalog.novels.version IS 'Version number, auto-incremented on each update';
COMMENT ON COLUMN catalog.novels.synopsis IS 'Synopsis in ORIGINAL language only. Translations go to novel_synopsis_translations';

-- =====================================================
-- VOLUMES TABLE (Middle Level)
-- =====================================================
CREATE TABLE catalog.volumes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- Basic info
    volume_number INTEGER NOT NULL CHECK (volume_number > 0),
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL,
    description TEXT,

    -- Images
    cover_image_url VARCHAR(1000),

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Statistics (auto-updated by application)
    chapter_count INTEGER NOT NULL DEFAULT 0 CHECK (chapter_count >= 0),
    word_count BIGINT NOT NULL DEFAULT 0 CHECK (word_count >= 0),

    -- Ordering
    display_order INTEGER NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,

    -- Dates
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    UNIQUE(novel_id, volume_number),
    UNIQUE(novel_id, slug)
);

-- Indexes for volumes
CREATE INDEX idx_volumes_novel_id ON catalog.volumes(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_volumes_display_order ON catalog.volumes(novel_id, display_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_volumes_published ON catalog.volumes(novel_id, published_at DESC) WHERE is_published = TRUE AND deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.volumes IS 'Middle-level table organizing chapters into volumes';
COMMENT ON COLUMN catalog.volumes.chapter_count IS 'Auto-updated by application when chapters change';

-- =====================================================
-- CHAPTERS TABLE (Bottom Level)
-- =====================================================
CREATE TABLE catalog.chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    volume_id UUID REFERENCES catalog.volumes(id) ON DELETE SET NULL,

    -- Basic info
    chapter_number INTEGER NOT NULL CHECK (chapter_number > 0),
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL,

    -- Content in ORIGINAL language only (JSONB)
    content JSONB NOT NULL,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Metrics
    word_count INTEGER NOT NULL DEFAULT 0 CHECK (word_count >= 0),
    character_count INTEGER NOT NULL DEFAULT 0 CHECK (character_count >= 0),

    -- Access control
    is_free BOOLEAN NOT NULL DEFAULT TRUE,
    price DECIMAL(10,2) DEFAULT 0.00 CHECK (price >= 0),
    currency VARCHAR(3) DEFAULT 'VND',

    -- Status
    status catalog.chapter_status NOT NULL DEFAULT 'draft',

    -- Statistics
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,

    -- Ordering
    display_order INTEGER NOT NULL,

    -- Author notes (JSONB)
    author_notes JSONB,

    -- Dates
    published_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    UNIQUE(novel_id, chapter_number),
    UNIQUE(novel_id, slug)
);

-- Indexes for chapters
CREATE INDEX idx_chapters_novel_id ON catalog.chapters(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_volume_id ON catalog.chapters(volume_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_status ON catalog.chapters(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_display_order ON catalog.chapters(novel_id, display_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_published ON catalog.chapters(published_at DESC) WHERE status = 'published' AND deleted_at IS NULL;
CREATE INDEX idx_chapters_scheduled ON catalog.chapters(scheduled_at ASC) WHERE status = 'scheduled' AND deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.chapters IS 'Bottom-level table storing chapter content';
COMMENT ON COLUMN catalog.chapters.content IS 'Chapter content in ORIGINAL language only. Translations go to chapter_translations';
COMMENT ON COLUMN catalog.chapters.volume_id IS 'Nullable - chapters can exist without belonging to a volume';

-- =====================================================
-- TRIGGER FUNCTIONS (Minimal - only version increment)
-- =====================================================

-- Function to increment version and update timestamp
CREATE OR REPLACE FUNCTION catalog.increment_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.version := OLD.version + 1;
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION catalog.update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Novels triggers
CREATE TRIGGER trg_novels_version
    BEFORE UPDATE ON catalog.novels
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Volumes triggers
CREATE TRIGGER trg_volumes_version
    BEFORE UPDATE ON catalog.volumes
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Chapters triggers
CREATE TRIGGER trg_chapters_version
    BEFORE UPDATE ON catalog.chapters
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();
-- =====================================================
-- Migration 000013: Ownership System
-- Description: Ownership transfers and exclusive rights reporting
-- =====================================================

-- =====================================================
-- ENUM TYPES
-- =====================================================

-- Transfer status
CREATE TYPE catalog.transfer_status AS ENUM (
    'pending',
    'approved',
    'rejected',
    'cancelled'
);

-- Report status
CREATE TYPE catalog.report_status AS ENUM (
    'pending',
    'under_review',
    'resolved',
    'rejected'
);

-- =====================================================
-- OWNERSHIP TRANSFERS TABLE
-- =====================================================
CREATE TABLE catalog.ownership_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- Source ownership
    from_owner_type VARCHAR(20) NOT NULL CHECK (from_owner_type IN ('user', 'organization')),
    from_owner_id UUID NOT NULL,

    -- Target ownership
    to_owner_type VARCHAR(20) NOT NULL CHECK (to_owner_type IN ('user', 'organization')),
    to_owner_id UUID NOT NULL,

    -- Transfer info
    status catalog.transfer_status NOT NULL DEFAULT 'pending',
    reason TEXT,

    -- Admin review (required for 2-way transfers)
    requires_approval BOOLEAN NOT NULL DEFAULT FALSE,
    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    review_notes TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    -- Ensure from and to are different
    CHECK (
        from_owner_type != to_owner_type OR
        from_owner_id != to_owner_id
    )
);

-- Indexes for ownership_transfers
CREATE INDEX idx_ownership_transfers_novel_id ON catalog.ownership_transfers(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_ownership_transfers_from ON catalog.ownership_transfers(from_owner_type, from_owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_ownership_transfers_to ON catalog.ownership_transfers(to_owner_type, to_owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_ownership_transfers_status ON catalog.ownership_transfers(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_ownership_transfers_pending_approval ON catalog.ownership_transfers(status, requires_approval) WHERE requires_approval = TRUE AND status = 'pending' AND deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.ownership_transfers IS 'Tracks ownership transfer requests between users and organizations';
COMMENT ON COLUMN catalog.ownership_transfers.from_owner_type IS 'Source owner type: user or organization';
COMMENT ON COLUMN catalog.ownership_transfers.from_owner_id IS 'Source owner UUID (validated in application)';
COMMENT ON COLUMN catalog.ownership_transfers.to_owner_type IS 'Target owner type: user or organization';
COMMENT ON COLUMN catalog.ownership_transfers.to_owner_id IS 'Target owner UUID (validated in application)';
COMMENT ON COLUMN catalog.ownership_transfers.requires_approval IS 'TRUE for 2-way transfers (user<->organization or organization<->organization), FALSE for 1-way (user->organization)';

-- =====================================================
-- EXCLUSIVE TRANSLATION REPORTS TABLE
-- =====================================================
CREATE TABLE catalog.exclusive_translation_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- ISO 639-1

    -- Reporting organization
    reporting_organization_id UUID NOT NULL, -- References identify.organizations

    -- Reported organization (claiming exclusive rights)
    reported_organization_id UUID NOT NULL, -- References identify.organizations

    -- Report details
    reason TEXT NOT NULL,
    evidence_urls TEXT[], -- Array of URLs to evidence
    status catalog.report_status NOT NULL DEFAULT 'pending',

    -- Admin review
    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    review_notes TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,

    -- Resolution
    resolution TEXT,
    resolved_at TIMESTAMP WITH TIME ZONE,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    -- Ensure reporting and reported organizations are different
    CHECK (reporting_organization_id != reported_organization_id)
);

-- Indexes for exclusive_translation_reports
CREATE INDEX idx_exclusive_reports_novel_id ON catalog.exclusive_translation_reports(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_exclusive_reports_language ON catalog.exclusive_translation_reports(language) WHERE deleted_at IS NULL;
CREATE INDEX idx_exclusive_reports_reporting_organization ON catalog.exclusive_translation_reports(reporting_organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_exclusive_reports_reported_organization ON catalog.exclusive_translation_reports(reported_organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_exclusive_reports_status ON catalog.exclusive_translation_reports(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_exclusive_reports_pending ON catalog.exclusive_translation_reports(status) WHERE status = 'pending' AND deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.exclusive_translation_reports IS 'Reports for organizations claiming exclusive translation rights';
COMMENT ON COLUMN catalog.exclusive_translation_reports.reporting_organization_id IS 'Organization filing the report';
COMMENT ON COLUMN catalog.exclusive_translation_reports.reported_organization_id IS 'Organization being reported for claiming exclusive rights';
COMMENT ON COLUMN catalog.exclusive_translation_reports.evidence_urls IS 'URLs to evidence supporting the report';

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Ownership transfers triggers
CREATE TRIGGER trg_ownership_transfers_version
    BEFORE UPDATE ON catalog.ownership_transfers
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Exclusive reports triggers
CREATE TRIGGER trg_exclusive_reports_version
    BEFORE UPDATE ON catalog.exclusive_translation_reports
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();
-- Migration: Delete Translation Teams (Merged into Organizations)
-- Description: This migration is no longer needed as translation_teams have been merged into organizations table
-- The functionality from this migration has been incorporated into:
-- - Migration 000001: organizations table (merged from translation_teams)
-- - Migration 000002: user_organization_memberships (merged from team_members)
--
-- This file is kept empty to maintain migration numbering sequence.
-- The corresponding features are now in the identify.organizations table.

-- No operations needed - features merged into earlier migrations
-- =====================================================
-- Migration 000015: Synopsis Translations
-- Description: Translations and contributions for novel synopsis
-- =====================================================

-- =====================================================
-- ENUM TYPES
-- =====================================================

-- Translation status
CREATE TYPE catalog.translation_status AS ENUM (
    'draft',
    'pending_review',
    'published',
    'rejected'
);

-- Contribution status
CREATE TYPE catalog.contribution_status AS ENUM (
    'pending',
    'accepted',
    'rejected'
);

-- =====================================================
-- NOVEL SYNOPSIS TRANSLATIONS TABLE
-- =====================================================
CREATE TABLE catalog.novel_synopsis_translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- ISO 639-1: en, vi, zh, ja, ko, etc.

    -- Translation content
    synopsis JSONB NOT NULL,

    -- Organization assignment (optional - can be contributed by individuals)
    organization_id UUID REFERENCES identify.organizations(id) ON DELETE SET NULL,

    -- Status
    status catalog.translation_status NOT NULL DEFAULT 'draft',

    -- Quality metrics
    quality_score DECIMAL(3,2) DEFAULT 0.00 CHECK (quality_score >= 0 AND quality_score <= 5),
    reviewer_rating DECIMAL(3,2) DEFAULT 0.00 CHECK (reviewer_rating >= 0 AND reviewer_rating <= 5),

    -- Statistics (auto-updated by application)
    contribution_count INTEGER NOT NULL DEFAULT 0 CHECK (contribution_count >= 0),
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,

    -- Review
    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    review_notes TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    -- One translation per novel per language
    UNIQUE(novel_id, language)
);

-- Indexes for novel_synopsis_translations
CREATE INDEX idx_synopsis_translations_novel_id ON catalog.novel_synopsis_translations(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_translations_language ON catalog.novel_synopsis_translations(language) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_translations_organization_id ON catalog.novel_synopsis_translations(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_translations_status ON catalog.novel_synopsis_translations(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_translations_published ON catalog.novel_synopsis_translations(novel_id, language, published_at DESC) WHERE status = 'published' AND deleted_at IS NULL;
CREATE INDEX idx_synopsis_translations_content ON catalog.novel_synopsis_translations USING GIN(synopsis) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.novel_synopsis_translations IS 'Translated synopsis for novels';
COMMENT ON COLUMN catalog.novel_synopsis_translations.synopsis IS 'Translated synopsis content in JSONB format';
COMMENT ON COLUMN catalog.novel_synopsis_translations.organization_id IS 'Optional organization responsible for this translation';
COMMENT ON COLUMN catalog.novel_synopsis_translations.quality_score IS 'Aggregate quality score from community (0-5 scale)';
COMMENT ON COLUMN catalog.novel_synopsis_translations.reviewer_rating IS 'Quality rating from official reviewers (0-5 scale)';

-- =====================================================
-- SYNOPSIS TRANSLATION CONTRIBUTIONS TABLE
-- =====================================================
CREATE TABLE catalog.synopsis_translation_contributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    synopsis_translation_id UUID NOT NULL REFERENCES catalog.novel_synopsis_translations(id) ON DELETE CASCADE,
    contributor_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,

    -- Contribution details
    contribution_type VARCHAR(50) NOT NULL, -- 'translation', 'proofread', 'edit', 'review'
    contribution_notes TEXT,
    status catalog.contribution_status NOT NULL DEFAULT 'pending',

    -- Changes (optional - for tracking what was changed)
    changes JSONB,

    -- Quality metrics
    quality_score DECIMAL(3,2) DEFAULT 0.00 CHECK (quality_score >= 0 AND quality_score <= 5),

    -- Review
    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    review_notes TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    contributed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL
);

-- Indexes for synopsis_translation_contributions
CREATE INDEX idx_synopsis_contributions_translation ON catalog.synopsis_translation_contributions(synopsis_translation_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_contributions_contributor ON catalog.synopsis_translation_contributions(contributor_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_contributions_type ON catalog.synopsis_translation_contributions(contribution_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_contributions_status ON catalog.synopsis_translation_contributions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_contributions_contributed ON catalog.synopsis_translation_contributions(contributed_at DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.synopsis_translation_contributions IS 'Tracks individual contributions to synopsis translations';
COMMENT ON COLUMN catalog.synopsis_translation_contributions.contribution_type IS 'Type of contribution: translation, proofread, edit, review';
COMMENT ON COLUMN catalog.synopsis_translation_contributions.changes IS 'JSONB documenting what was changed in this contribution';

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Synopsis translations triggers
CREATE TRIGGER trg_synopsis_translations_version
    BEFORE UPDATE ON catalog.novel_synopsis_translations
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Synopsis contributions triggers
CREATE TRIGGER trg_synopsis_contributions_version
    BEFORE UPDATE ON catalog.synopsis_translation_contributions
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();
-- =====================================================
-- Migration 000016: Chapter Translations
-- Description: Translations, contributions, and version history for chapters
-- =====================================================

-- =====================================================
-- CHAPTER TRANSLATIONS TABLE
-- =====================================================
CREATE TABLE catalog.chapter_translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id UUID NOT NULL REFERENCES catalog.chapters(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- ISO 639-1: en, vi, zh, ja, ko, etc.

    -- Translation content
    content JSONB NOT NULL,
    title VARCHAR(500) NOT NULL,

    -- Organization assignment (optional - can be contributed by individuals)
    organization_id UUID REFERENCES identify.organizations(id) ON DELETE SET NULL,

    -- Status
    status catalog.translation_status NOT NULL DEFAULT 'draft',

    -- Quality metrics
    quality_score DECIMAL(3,2) DEFAULT 0.00 CHECK (quality_score >= 0 AND quality_score <= 5),
    reviewer_rating DECIMAL(3,2) DEFAULT 0.00 CHECK (reviewer_rating >= 0 AND reviewer_rating >= 0),

    -- Metrics
    word_count INTEGER NOT NULL DEFAULT 0 CHECK (word_count >= 0),
    character_count INTEGER NOT NULL DEFAULT 0 CHECK (character_count >= 0),

    -- Statistics (auto-updated by application)
    contribution_count INTEGER NOT NULL DEFAULT 0 CHECK (contribution_count >= 0),
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,

    -- Review
    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    review_notes TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,

    -- Author/Translator notes
    translator_notes JSONB,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    -- One translation per chapter per language
    UNIQUE(chapter_id, language)
);

-- Indexes for chapter_translations
CREATE INDEX idx_chapter_translations_chapter_id ON catalog.chapter_translations(chapter_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_language ON catalog.chapter_translations(language) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_organization_id ON catalog.chapter_translations(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_status ON catalog.chapter_translations(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_published ON catalog.chapter_translations(chapter_id, language, published_at DESC) WHERE status = 'published' AND deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_content ON catalog.chapter_translations USING GIN(content) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_views ON catalog.chapter_translations(view_count DESC) WHERE status = 'published' AND deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.chapter_translations IS 'Translated chapter content';
COMMENT ON COLUMN catalog.chapter_translations.content IS 'Translated chapter content in JSONB format';
COMMENT ON COLUMN catalog.chapter_translations.organization_id IS 'Optional organization responsible for this translation';
COMMENT ON COLUMN catalog.chapter_translations.quality_score IS 'Aggregate quality score from community (0-5 scale)';
COMMENT ON COLUMN catalog.chapter_translations.reviewer_rating IS 'Quality rating from official reviewers (0-5 scale)';

-- =====================================================
-- TRANSLATION CONTRIBUTIONS TABLE
-- =====================================================
CREATE TABLE catalog.translation_contributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_translation_id UUID NOT NULL REFERENCES catalog.chapter_translations(id) ON DELETE CASCADE,
    contributor_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,

    -- Contribution details
    contribution_type VARCHAR(50) NOT NULL, -- 'translation', 'proofread', 'edit', 'review', 'typeset'
    contribution_notes TEXT,
    status catalog.contribution_status NOT NULL DEFAULT 'pending',

    -- Changes (optional - stores metadata about changes, NOT full content)
    changes JSONB,

    -- Quality metrics
    quality_score DECIMAL(3,2) DEFAULT 0.00 CHECK (quality_score >= 0 AND quality_score <= 5),

    -- Review
    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    review_notes TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    contributed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL
);

-- Indexes for translation_contributions
CREATE INDEX idx_translation_contributions_chapter ON catalog.translation_contributions(chapter_translation_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_contributor ON catalog.translation_contributions(contributor_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_type ON catalog.translation_contributions(contribution_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_status ON catalog.translation_contributions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_contributed ON catalog.translation_contributions(contributed_at DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.translation_contributions IS 'Tracks individual contributions to chapter translations';
COMMENT ON COLUMN catalog.translation_contributions.contribution_type IS 'Type of contribution: translation, proofread, edit, review, typeset';
COMMENT ON COLUMN catalog.translation_contributions.changes IS 'JSONB documenting metadata about changes (NOT full content to save space)';

-- =====================================================
-- TRANSLATION HISTORY TABLE
-- =====================================================
CREATE TABLE catalog.translation_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_translation_id UUID NOT NULL REFERENCES catalog.chapter_translations(id) ON DELETE CASCADE,

    -- Version tracking
    version_number INTEGER NOT NULL CHECK (version_number > 0),

    -- Metadata (NOT full content - to save space)
    title VARCHAR(500),
    word_count INTEGER,
    character_count INTEGER,
    status catalog.translation_status,

    -- Change summary
    change_summary TEXT,
    changed_fields JSONB, -- Array of field names that changed

    -- Who made this version
    changed_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(chapter_translation_id, version_number)
);

-- Indexes for translation_history
CREATE INDEX idx_translation_history_chapter ON catalog.translation_history(chapter_translation_id);
CREATE INDEX idx_translation_history_version ON catalog.translation_history(chapter_translation_id, version_number DESC);
CREATE INDEX idx_translation_history_changed_by ON catalog.translation_history(changed_by);
CREATE INDEX idx_translation_history_created ON catalog.translation_history(created_at DESC);

-- Comments
COMMENT ON TABLE catalog.translation_history IS 'Version control for chapter translations (metadata only, NOT full content)';
COMMENT ON COLUMN catalog.translation_history.changed_fields IS 'JSONB array of field names that changed in this version';
COMMENT ON COLUMN catalog.translation_history.change_summary IS 'Human-readable summary of what changed';

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Chapter translations triggers
CREATE TRIGGER trg_chapter_translations_version
    BEFORE UPDATE ON catalog.chapter_translations
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Translation contributions triggers
CREATE TRIGGER trg_translation_contributions_version
    BEFORE UPDATE ON catalog.translation_contributions
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();
-- =====================================================
-- Migration 000017: Audit History Tables
-- Description: History tables for novels, volumes, and chapters
-- Note: These tables store METADATA only, NOT full content snapshots
-- Logging is done at APPLICATION layer, NOT via database triggers
-- =====================================================

-- =====================================================
-- ENUM TYPES
-- =====================================================

-- Action type
CREATE TYPE catalog.audit_action AS ENUM (
    'created',
    'updated',
    'deleted',
    'restored',
    'published',
    'unpublished',
    'transferred'
);

-- =====================================================
-- NOVEL HISTORY TABLE
-- =====================================================
CREATE TABLE catalog.novel_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- Version tracking
    version_number INTEGER NOT NULL CHECK (version_number > 0),

    -- Action
    action catalog.audit_action NOT NULL,

    -- Metadata (NOT full content)
    title VARCHAR(500),
    slug VARCHAR(500),
    status catalog.novel_status,
    owner_type VARCHAR(20),
    owner_id UUID,

    -- Statistics snapshot
    total_volumes INTEGER,
    total_chapters INTEGER,
    total_words BIGINT,

    -- Change tracking
    changed_fields JSONB, -- Array of field names that changed: ["title", "status", "cover_image_url"]
    change_summary TEXT, -- Human-readable summary

    -- Who made this change
    changed_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,

    -- Request context (rich application context)
    request_id VARCHAR(100),
    ip_address INET,
    user_agent TEXT,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(novel_id, version_number)
);

-- Indexes for novel_history
CREATE INDEX idx_novel_history_novel ON catalog.novel_history(novel_id);
CREATE INDEX idx_novel_history_version ON catalog.novel_history(novel_id, version_number DESC);
CREATE INDEX idx_novel_history_action ON catalog.novel_history(action);
CREATE INDEX idx_novel_history_changed_by ON catalog.novel_history(changed_by);
CREATE INDEX idx_novel_history_created ON catalog.novel_history(created_at DESC);
CREATE INDEX idx_novel_history_request ON catalog.novel_history(request_id) WHERE request_id IS NOT NULL;

-- Comments
COMMENT ON TABLE catalog.novel_history IS 'Audit log for novel changes (metadata only, logged at application layer)';
COMMENT ON COLUMN catalog.novel_history.changed_fields IS 'JSONB array of field names that changed';
COMMENT ON COLUMN catalog.novel_history.change_summary IS 'Human-readable description of changes';
COMMENT ON COLUMN catalog.novel_history.request_id IS 'Request ID for tracing';

-- =====================================================
-- VOLUME HISTORY TABLE
-- =====================================================
CREATE TABLE catalog.volume_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    volume_id UUID NOT NULL REFERENCES catalog.volumes(id) ON DELETE CASCADE,
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- Version tracking
    version_number INTEGER NOT NULL CHECK (version_number > 0),

    -- Action
    action catalog.audit_action NOT NULL,

    -- Metadata (NOT full content)
    title VARCHAR(500),
    slug VARCHAR(500),
    volume_number INTEGER,
    is_published BOOLEAN,

    -- Statistics snapshot
    chapter_count INTEGER,
    word_count BIGINT,

    -- Change tracking
    changed_fields JSONB, -- Array of field names that changed
    change_summary TEXT,

    -- Who made this change
    changed_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,

    -- Request context
    request_id VARCHAR(100),
    ip_address INET,
    user_agent TEXT,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(volume_id, version_number)
);

-- Indexes for volume_history
CREATE INDEX idx_volume_history_volume ON catalog.volume_history(volume_id);
CREATE INDEX idx_volume_history_novel ON catalog.volume_history(novel_id);
CREATE INDEX idx_volume_history_version ON catalog.volume_history(volume_id, version_number DESC);
CREATE INDEX idx_volume_history_action ON catalog.volume_history(action);
CREATE INDEX idx_volume_history_changed_by ON catalog.volume_history(changed_by);
CREATE INDEX idx_volume_history_created ON catalog.volume_history(created_at DESC);

-- Comments
COMMENT ON TABLE catalog.volume_history IS 'Audit log for volume changes (metadata only, logged at application layer)';
COMMENT ON COLUMN catalog.volume_history.changed_fields IS 'JSONB array of field names that changed';

-- =====================================================
-- CHAPTER HISTORY TABLE
-- =====================================================
CREATE TABLE catalog.chapter_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id UUID NOT NULL REFERENCES catalog.chapters(id) ON DELETE CASCADE,
    volume_id UUID REFERENCES catalog.volumes(id) ON DELETE SET NULL,
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- Version tracking
    version_number INTEGER NOT NULL CHECK (version_number > 0),

    -- Action
    action catalog.audit_action NOT NULL,

    -- Metadata (NOT full content - content stored separately if needed)
    title VARCHAR(500),
    slug VARCHAR(500),
    chapter_number INTEGER,
    status catalog.chapter_status,

    -- Metrics snapshot
    word_count INTEGER,
    character_count INTEGER,

    -- Change tracking
    changed_fields JSONB, -- Array of field names that changed
    change_summary TEXT,
    content_changed BOOLEAN DEFAULT FALSE, -- Flag indicating if content JSONB changed

    -- Who made this change
    changed_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,

    -- Request context
    request_id VARCHAR(100),
    ip_address INET,
    user_agent TEXT,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(chapter_id, version_number)
);

-- Indexes for chapter_history
CREATE INDEX idx_chapter_history_chapter ON catalog.chapter_history(chapter_id);
CREATE INDEX idx_chapter_history_volume ON catalog.chapter_history(volume_id);
CREATE INDEX idx_chapter_history_novel ON catalog.chapter_history(novel_id);
CREATE INDEX idx_chapter_history_version ON catalog.chapter_history(chapter_id, version_number DESC);
CREATE INDEX idx_chapter_history_action ON catalog.chapter_history(action);
CREATE INDEX idx_chapter_history_changed_by ON catalog.chapter_history(changed_by);
CREATE INDEX idx_chapter_history_created ON catalog.chapter_history(created_at DESC);
CREATE INDEX idx_chapter_history_content_changed ON catalog.chapter_history(chapter_id, content_changed) WHERE content_changed = TRUE;

-- Comments
COMMENT ON TABLE catalog.chapter_history IS 'Audit log for chapter changes (metadata only, logged at application layer)';
COMMENT ON COLUMN catalog.chapter_history.changed_fields IS 'JSONB array of field names that changed';
COMMENT ON COLUMN catalog.chapter_history.content_changed IS 'Flag indicating if content JSONB changed (actual content NOT stored here to save space)';
COMMENT ON COLUMN catalog.chapter_history.change_summary IS 'Human-readable description like "Updated chapter content and title"';
-- =====================================================
-- Migration 000018: Supporting Tables
-- Description: Genres, authors, artists, translators, and their associations
-- =====================================================

-- =====================================================
-- ENUM TYPES
-- =====================================================

-- Contributor role types
CREATE TYPE catalog.author_role AS ENUM (
    'original_author',
    'co_author',
    'ghostwriter'
);

CREATE TYPE catalog.artist_role AS ENUM (
    'cover_artist',
    'illustrator',
    'character_designer'
);

CREATE TYPE catalog.translator_role AS ENUM (
    'translator',
    'localizer',
    'adapter'
);

-- =====================================================
-- GENRES TABLE
-- =====================================================
CREATE TABLE catalog.genres (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic info
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,

    -- Hierarchy (optional parent for sub-genres)
    parent_id UUID REFERENCES catalog.genres(id) ON DELETE SET NULL,

    -- Display
    icon VARCHAR(100), -- Icon name or emoji
    color VARCHAR(7), -- Hex color code
    display_order INTEGER NOT NULL DEFAULT 0,

    -- Statistics (auto-updated by application)
    novel_count INTEGER NOT NULL DEFAULT 0 CHECK (novel_count >= 0),
    active_readers BIGINT NOT NULL DEFAULT 0 CHECK (active_readers >= 0),
    total_views BIGINT NOT NULL DEFAULT 0 CHECK (total_views >= 0),

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL
);

-- Indexes for genres
CREATE INDEX idx_genres_slug ON catalog.genres(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_genres_parent ON catalog.genres(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_genres_display_order ON catalog.genres(display_order) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.genres IS 'Genre definitions for novels';
COMMENT ON COLUMN catalog.genres.parent_id IS 'Optional parent genre for hierarchical organization';

-- =====================================================
-- NOVEL_GENRES JUNCTION TABLE
-- =====================================================
CREATE TABLE catalog.novel_genres (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES catalog.genres(id) ON DELETE CASCADE,

    -- Ordering
    display_order INTEGER NOT NULL DEFAULT 0,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(novel_id, genre_id)
);

-- Indexes for novel_genres
CREATE INDEX idx_novel_genres_novel ON catalog.novel_genres(novel_id);
CREATE INDEX idx_novel_genres_genre ON catalog.novel_genres(genre_id);

-- Comments
COMMENT ON TABLE catalog.novel_genres IS 'Junction table linking novels to genres';

-- =====================================================
-- AUTHORS TABLE
-- =====================================================
CREATE TABLE catalog.authors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic info
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL UNIQUE,
    bio TEXT,
    avatar_url VARCHAR(1000),

    -- Optional link to user account
    user_id UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    -- Statistics (auto-updated by application)
    novel_count INTEGER NOT NULL DEFAULT 0 CHECK (novel_count >= 0),
    total_views BIGINT NOT NULL DEFAULT 0,

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL
);

-- Indexes for authors
CREATE INDEX idx_authors_slug ON catalog.authors(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_authors_user_id ON catalog.authors(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_authors_metadata ON catalog.authors USING GIN(metadata) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.authors IS 'Author information';
COMMENT ON COLUMN catalog.authors.user_id IS 'Optional link to user account if author is registered';

-- =====================================================
-- NOVEL_AUTHORS JUNCTION TABLE
-- =====================================================
CREATE TABLE catalog.novel_authors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES catalog.authors(id) ON DELETE CASCADE,

    -- Role
    role catalog.author_role NOT NULL DEFAULT 'original_author',

    -- Ordering
    display_order INTEGER NOT NULL DEFAULT 0,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(novel_id, author_id)
);

-- Indexes for novel_authors
CREATE INDEX idx_novel_authors_novel ON catalog.novel_authors(novel_id);
CREATE INDEX idx_novel_authors_author ON catalog.novel_authors(author_id);
CREATE INDEX idx_novel_authors_role ON catalog.novel_authors(novel_id, role);

-- Comments
COMMENT ON TABLE catalog.novel_authors IS 'Junction table linking novels to authors';

-- =====================================================
-- ARTISTS TABLE
-- =====================================================
CREATE TABLE catalog.artists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic info
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL UNIQUE,
    bio TEXT,
    avatar_url VARCHAR(1000),
    portfolio_url VARCHAR(1000),

    -- Optional link to user account
    user_id UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    -- Statistics (auto-updated by application)
    novel_count INTEGER NOT NULL DEFAULT 0 CHECK (novel_count >= 0),

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL
);

-- Indexes for artists
CREATE INDEX idx_artists_slug ON catalog.artists(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_artists_user_id ON catalog.artists(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_artists_metadata ON catalog.artists USING GIN(metadata) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.artists IS 'Artist information (cover artists, illustrators, etc.)';

-- =====================================================
-- NOVEL_ARTISTS JUNCTION TABLE
-- =====================================================
CREATE TABLE catalog.novel_artists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    artist_id UUID NOT NULL REFERENCES catalog.artists(id) ON DELETE CASCADE,

    -- Role
    role catalog.artist_role NOT NULL DEFAULT 'cover_artist',

    -- Ordering
    display_order INTEGER NOT NULL DEFAULT 0,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(novel_id, artist_id, role)
);

-- Indexes for novel_artists
CREATE INDEX idx_novel_artists_novel ON catalog.novel_artists(novel_id);
CREATE INDEX idx_novel_artists_artist ON catalog.novel_artists(artist_id);
CREATE INDEX idx_novel_artists_role ON catalog.novel_artists(novel_id, role);

-- Comments
COMMENT ON TABLE catalog.novel_artists IS 'Junction table linking novels to artists';

-- =====================================================
-- TRANSLATORS TABLE
-- =====================================================
CREATE TABLE catalog.translators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic info
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL UNIQUE,
    bio TEXT,
    avatar_url VARCHAR(1000),

    -- Link to user account (usually required for translators)
    user_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,

    -- Languages
    source_languages VARCHAR(10)[] NOT NULL, -- Array of ISO 639-1 codes
    target_languages VARCHAR(10)[] NOT NULL, -- Array of ISO 639-1 codes

    -- Statistics (auto-updated by application)
    translation_count INTEGER NOT NULL DEFAULT 0 CHECK (translation_count >= 0),
    total_words_translated BIGINT NOT NULL DEFAULT 0,
    quality_score DECIMAL(3,2) DEFAULT 0.00 CHECK (quality_score >= 0 AND quality_score <= 5),

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL
);

-- Indexes for translators
CREATE INDEX idx_translators_slug ON catalog.translators(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_translators_user_id ON catalog.translators(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_translators_source_langs ON catalog.translators USING GIN(source_languages) WHERE deleted_at IS NULL;
CREATE INDEX idx_translators_target_langs ON catalog.translators USING GIN(target_languages) WHERE deleted_at IS NULL;
CREATE INDEX idx_translators_quality ON catalog.translators(quality_score DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.translators IS 'Translator information';
COMMENT ON COLUMN catalog.translators.source_languages IS 'Array of languages the translator can translate from';
COMMENT ON COLUMN catalog.translators.target_languages IS 'Array of languages the translator can translate to';

-- =====================================================
-- NOVEL_TRANSLATORS JUNCTION TABLE
-- =====================================================
CREATE TABLE catalog.novel_translators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    translator_id UUID NOT NULL REFERENCES catalog.translators(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- Target language

    -- Role
    role catalog.translator_role NOT NULL DEFAULT 'translator',

    -- Statistics (auto-updated by application)
    chapters_translated INTEGER NOT NULL DEFAULT 0 CHECK (chapters_translated >= 0),

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(novel_id, translator_id, language)
);

-- Indexes for novel_translators
CREATE INDEX idx_novel_translators_novel ON catalog.novel_translators(novel_id);
CREATE INDEX idx_novel_translators_translator ON catalog.novel_translators(translator_id);
CREATE INDEX idx_novel_translators_language ON catalog.novel_translators(novel_id, language);

-- Comments
COMMENT ON TABLE catalog.novel_translators IS 'Junction table linking novels to translators';
COMMENT ON COLUMN catalog.novel_translators.language IS 'Target language for this translator on this novel';

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Genres triggers
CREATE TRIGGER trg_genres_version
    BEFORE UPDATE ON catalog.genres
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Authors triggers
CREATE TRIGGER trg_authors_version
    BEFORE UPDATE ON catalog.authors
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Artists triggers
CREATE TRIGGER trg_artists_version
    BEFORE UPDATE ON catalog.artists
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Translators triggers
CREATE TRIGGER trg_translators_version
    BEFORE UPDATE ON catalog.translators
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();
-- Add is_active column to genres table
ALTER TABLE catalog.genres
ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

-- Create index for active genres filtering
CREATE INDEX idx_genres_is_active ON catalog.genres(is_active) WHERE deleted_at IS NULL;

-- Comment
COMMENT ON COLUMN catalog.genres.is_active IS 'Whether the genre is active and visible to users';
-- Add missing columns to authors table
ALTER TABLE catalog.authors
ADD COLUMN total_chapters INTEGER NOT NULL DEFAULT 0 CHECK (total_chapters >= 0),
ADD COLUMN follower_count INTEGER NOT NULL DEFAULT 0 CHECK (follower_count >= 0),
ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT false;

-- Create indexes for filtering
CREATE INDEX idx_authors_is_verified ON catalog.authors(is_verified) WHERE deleted_at IS NULL;
CREATE INDEX idx_authors_follower_count ON catalog.authors(follower_count DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON COLUMN catalog.authors.total_chapters IS 'Total number of chapters written by this author across all novels';
COMMENT ON COLUMN catalog.authors.follower_count IS 'Number of followers for this author';
COMMENT ON COLUMN catalog.authors.is_verified IS 'Whether the author has been verified by the platform';
-- Add missing columns to artists table
ALTER TABLE catalog.artists
ADD COLUMN specialization VARCHAR(100),
ADD COLUMN artwork_count INTEGER NOT NULL DEFAULT 0 CHECK (artwork_count >= 0),
ADD COLUMN follower_count INTEGER NOT NULL DEFAULT 0 CHECK (follower_count >= 0),
ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT false;

-- Create indexes for filtering
CREATE INDEX idx_artists_specialization ON catalog.artists(specialization) WHERE deleted_at IS NULL;
CREATE INDEX idx_artists_is_verified ON catalog.artists(is_verified) WHERE deleted_at IS NULL;
CREATE INDEX idx_artists_follower_count ON catalog.artists(follower_count DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON COLUMN catalog.artists.specialization IS 'Artist specialization (e.g., cover_artist, illustrator, manga_artist)';
COMMENT ON COLUMN catalog.artists.artwork_count IS 'Total number of artworks created by this artist';
COMMENT ON COLUMN catalog.artists.follower_count IS 'Number of followers for this artist';
COMMENT ON COLUMN catalog.artists.is_verified IS 'Whether the artist has been verified by the platform';
-- Fix authors table: rename bio to biography and add social_links
ALTER TABLE catalog.authors
RENAME COLUMN bio TO biography;

ALTER TABLE catalog.authors
ADD COLUMN social_links JSONB DEFAULT '{}';

COMMENT ON COLUMN catalog.authors.biography IS 'Author biography in JSONB format';
COMMENT ON COLUMN catalog.authors.social_links IS 'Social media links in JSONB format (e.g., {"facebook": "...", "twitter": "..."})';

-- Fix artists table: rename bio to biography and add social_links
ALTER TABLE catalog.artists
RENAME COLUMN bio TO biography;

ALTER TABLE catalog.artists
ADD COLUMN social_links JSONB DEFAULT '{}';

COMMENT ON COLUMN catalog.artists.biography IS 'Artist biography in JSONB format';
COMMENT ON COLUMN catalog.artists.social_links IS 'Social media links in JSONB format (e.g., {"facebook": "...", "twitter": "..."})';
-- Migration: Add Settings to Users
-- Description: Thêm cột settings (JSONB) vào bảng users để lưu user preferences
-- Author: System
-- Created: 2025-11-19

ALTER TABLE identify.users ADD COLUMN settings JSONB DEFAULT '{}'::jsonb;

COMMENT ON COLUMN identify.users.settings IS 'User preferences (theme, language, notifications, etc.)';
-- Migration: Create WebAuthn Credentials Table
-- Description: Tạo bảng lưu trữ passkey/WebAuthn credentials cho passwordless authentication
-- Author: System
-- Created: 2025-11-22

-- =====================================================
-- Table: webauthn_credentials
-- Description: Lưu trữ WebAuthn/FIDO2 credentials cho passwordless authentication
-- =====================================================
CREATE TABLE identify.webauthn_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,

    -- WebAuthn credential data
    credential_id TEXT UNIQUE NOT NULL, -- Base64URL encoded credential ID từ authenticator
    public_key BYTEA NOT NULL, -- Public key của credential (binary format)

    -- Attestation information
    attestation_type VARCHAR(50) NOT NULL DEFAULT 'none', -- none, indirect, direct
    aaguid BYTEA, -- Authenticator AAGUID (16 bytes)

    -- Security and tracking
    sign_count INTEGER NOT NULL DEFAULT 0, -- Counter để phát hiện cloned authenticators

    -- Credential metadata
    transports TEXT[], -- Supported transports: usb, nfc, ble, internal, hybrid
    backup_eligible BOOLEAN NOT NULL DEFAULT FALSE, -- Có thể backup credential không
    backup_state BOOLEAN NOT NULL DEFAULT FALSE, -- Credential đã được backup chưa

    -- User-friendly information
    credential_name VARCHAR(255), -- User-defined name (e.g., "iPhone 15 Pro", "YubiKey 5")

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ, -- Lần cuối credential được sử dụng để authentication

    -- Foreign key constraint
    CONSTRAINT fk_webauthn_credentials_user
        FOREIGN KEY (user_id)
        REFERENCES identify.users(id)
        ON DELETE CASCADE,

    -- Check constraints
    CONSTRAINT webauthn_credentials_attestation_type_check
        CHECK (attestation_type IN ('none', 'indirect', 'direct'))
);

-- Indexes
CREATE INDEX idx_webauthn_credentials_user_id ON identify.webauthn_credentials(user_id);
CREATE INDEX idx_webauthn_credentials_credential_id ON identify.webauthn_credentials(credential_id);
CREATE INDEX idx_webauthn_credentials_created_at ON identify.webauthn_credentials(created_at);
CREATE INDEX idx_webauthn_credentials_last_used_at ON identify.webauthn_credentials(last_used_at);

-- Comments
COMMENT ON TABLE identify.webauthn_credentials IS 'Bảng lưu trữ WebAuthn/FIDO2 credentials cho passwordless authentication';
COMMENT ON COLUMN identify.webauthn_credentials.credential_id IS 'Unique identifier của credential từ authenticator (Base64URL encoded)';
COMMENT ON COLUMN identify.webauthn_credentials.public_key IS 'Public key của credential trong binary format';
COMMENT ON COLUMN identify.webauthn_credentials.attestation_type IS 'Loại attestation: none (no attestation), indirect (anonymized), direct (full)';
COMMENT ON COLUMN identify.webauthn_credentials.aaguid IS 'Authenticator AAGUID (16 bytes) - identifies authenticator model';
COMMENT ON COLUMN identify.webauthn_credentials.sign_count IS 'Counter tăng dần mỗi lần authentication - dùng để phát hiện cloned authenticators';
COMMENT ON COLUMN identify.webauthn_credentials.transports IS 'Supported transports (usb, nfc, ble, internal, hybrid)';
COMMENT ON COLUMN identify.webauthn_credentials.backup_eligible IS 'Credential có thể được backup (multi-device credentials)';
COMMENT ON COLUMN identify.webauthn_credentials.backup_state IS 'Credential hiện đang được backup hay không';
COMMENT ON COLUMN identify.webauthn_credentials.credential_name IS 'User-defined name để nhận diện credential';
COMMENT ON COLUMN identify.webauthn_credentials.last_used_at IS 'Timestamp lần cuối credential được sử dụng để authentication';

-- =====================================================
-- Table: webauthn_sessions
-- Description: Lưu trữ temporary sessions cho WebAuthn registration/authentication flow
-- =====================================================
CREATE TABLE identify.webauthn_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID, -- NULL for registration of new users, set for authentication

    -- Session data
    challenge TEXT NOT NULL, -- Base64URL encoded challenge
    session_type VARCHAR(50) NOT NULL, -- registration, authentication

    -- Session metadata
    user_agent TEXT, -- Browser/device information
    ip_address VARCHAR(45), -- IPv4 or IPv6

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '5 minutes', -- Sessions expire after 5 minutes

    -- Foreign key constraint (optional, as user_id can be NULL for registration)
    CONSTRAINT fk_webauthn_sessions_user
        FOREIGN KEY (user_id)
        REFERENCES identify.users(id)
        ON DELETE CASCADE,

    -- Check constraints
    CONSTRAINT webauthn_sessions_session_type_check
        CHECK (session_type IN ('registration', 'authentication'))
);

-- Indexes
CREATE INDEX idx_webauthn_sessions_user_id ON identify.webauthn_sessions(user_id);
CREATE INDEX idx_webauthn_sessions_challenge ON identify.webauthn_sessions(challenge);
CREATE INDEX idx_webauthn_sessions_expires_at ON identify.webauthn_sessions(expires_at);
CREATE INDEX idx_webauthn_sessions_created_at ON identify.webauthn_sessions(created_at);

-- Comments
COMMENT ON TABLE identify.webauthn_sessions IS 'Bảng lưu trữ temporary sessions cho WebAuthn registration/authentication flow';
COMMENT ON COLUMN identify.webauthn_sessions.challenge IS 'Random challenge string (Base64URL encoded) dùng cho ceremony';
COMMENT ON COLUMN identify.webauthn_sessions.session_type IS 'Loại session: registration (đăng ký credential mới) hoặc authentication (xác thực)';
COMMENT ON COLUMN identify.webauthn_sessions.expires_at IS 'Session timeout (default 5 minutes)';

-- =====================================================
-- Update users table to support passwordless accounts
-- =====================================================
-- Make password_hash optional to support passwordless accounts
ALTER TABLE identify.users ALTER COLUMN password_hash DROP NOT NULL;

-- Add comment
COMMENT ON COLUMN identify.users.password_hash IS 'Password hash (NULL for passwordless accounts using only WebAuthn)';
-- Migration: Convert all tables to use UUID v7
-- Description: Convert from gen_random_uuid() (UUID v4) to uuidv7() (UUID v7)
-- Author: System
-- Created: 2025-11-22
-- Note: Only updates tables in identify schema (novel schema will be added in future)

-- Schema: identify
ALTER TABLE identify.users ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE identify.roles ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE identify.permissions ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE identify.oauth2_clients ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE identify.oauth2_consents ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE identify.email_verification_tokens ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE identify.password_reset_tokens ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE identify.webauthn_credentials ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE identify.webauthn_sessions ALTER COLUMN id SET DEFAULT uuidv7();
-- =====================================================
-- Migration 000026: Remove created_by and role from novel_authors and novel_artists
-- Description: Simplify novel_authors and novel_artists tables
-- =====================================================

-- Remove role and created_by from novel_authors
ALTER TABLE catalog.novel_authors
    DROP COLUMN IF EXISTS role,
    DROP COLUMN IF EXISTS created_by,
    DROP COLUMN IF EXISTS created_at;

-- Remove role and created_by from novel_artists
ALTER TABLE catalog.novel_artists
    DROP COLUMN IF EXISTS role,
    DROP COLUMN IF EXISTS created_by,
    DROP COLUMN IF EXISTS created_at;

-- Update unique constraint for novel_artists (remove role from constraint)
ALTER TABLE catalog.novel_artists
    DROP CONSTRAINT IF EXISTS novel_artists_novel_id_artist_id_role_key;

ALTER TABLE catalog.novel_artists
    ADD CONSTRAINT novel_artists_novel_id_artist_id_key UNIQUE (novel_id, artist_id);

-- Drop role-related indexes
DROP INDEX IF EXISTS catalog.idx_novel_authors_role;
DROP INDEX IF EXISTS catalog.idx_novel_artists_role;
-- Enable pg_trgm extension for fuzzy search and ILIKE support
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create GIN indexes using gin_trgm_ops for fast ILIKE searches
-- This supports multi-language and partial matching better than simple FTS
CREATE INDEX idx_novels_title_trgm ON catalog.novels USING GIN (title gin_trgm_ops);
CREATE INDEX idx_novels_original_title_trgm ON catalog.novels USING GIN (original_title gin_trgm_ops);
-- =====================================================
-- Migration 000028: Rename Novel Tables
-- Description: Rename tables to add 'novel_' prefix for consistency
-- =====================================================

-- Rename tables
-- PostgreSQL automatically renames dependent objects (indexes, constraints, triggers)
ALTER TABLE catalog.volumes RENAME TO novel_volumes;
ALTER TABLE catalog.volume_history RENAME TO novel_volume_histories;
ALTER TABLE catalog.chapters RENAME TO novel_chapters;
ALTER TABLE catalog.chapter_translations RENAME TO novel_chapter_translations;
ALTER TABLE catalog.chapter_history RENAME TO novel_chapter_histories;
-- =====================================================
-- Migration 000029: Novel Organization Assignments
-- Description: Assigns organizations to translate specific novels
-- Merged from migration 000014 novel_team_assignments
-- =====================================================

-- Enum type for assignment_status (merged from 000014)
-- Use IF NOT EXISTS in case it was created by old migration 000014
DO $$ BEGIN
    CREATE TYPE catalog.assignment_status AS ENUM (
        'active',
        'inactive',
        'suspended'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- =====================================================
-- NOVEL ORGANIZATION ASSIGNMENTS TABLE
-- =====================================================
CREATE TABLE catalog.novel_organization_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES identify.organizations(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- ISO 639-1: en, vi, zh, ja, ko, etc.

    -- Assignment details
    status catalog.assignment_status NOT NULL DEFAULT 'active',
    has_exclusive_rights BOOLEAN NOT NULL DEFAULT FALSE,

    -- Statistics (auto-updated by application)
    chapters_translated INTEGER NOT NULL DEFAULT 0 CHECK (chapters_translated >= 0),
    chapters_proofread INTEGER NOT NULL DEFAULT 0 CHECK (chapters_proofread >= 0),

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    -- One organization per novel per language (can have multiple if no exclusive rights)
    UNIQUE(novel_id, language, organization_id)
);

-- Indexes for novel_organization_assignments
CREATE INDEX idx_novel_organization_assignments_novel ON catalog.novel_organization_assignments(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_novel_organization_assignments_organization ON catalog.novel_organization_assignments(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_novel_organization_assignments_language ON catalog.novel_organization_assignments(novel_id, language) WHERE deleted_at IS NULL;
CREATE INDEX idx_novel_organization_assignments_exclusive ON catalog.novel_organization_assignments(novel_id, language, has_exclusive_rights) WHERE has_exclusive_rights = TRUE AND status = 'active' AND deleted_at IS NULL;
CREATE INDEX idx_novel_organization_assignments_status ON catalog.novel_organization_assignments(status) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.novel_organization_assignments IS 'Assigns organizations to translate specific novels in specific languages';
COMMENT ON COLUMN catalog.novel_organization_assignments.has_exclusive_rights IS 'If TRUE, this organization claims exclusive translation rights (can be challenged via reports)';
COMMENT ON COLUMN catalog.novel_organization_assignments.status IS 'Assignment status: active, inactive, or suspended';

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Novel organization assignments triggers
CREATE TRIGGER trg_novel_organization_assignments_version
    BEFORE UPDATE ON catalog.novel_organization_assignments
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();
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
    ('tenant:manage_member', 'organization', 'Manage tenant members', 'organization', 'manage_member'),
    ('tenant:assign_permission', 'organization', 'Assign permissions in tenant', 'organization', 'assign_permission'),
    ('tenant:update_info', 'organization', 'Update tenant information', 'organization', 'update_info'),
    ('tenant:view_stats', 'organization', 'View tenant statistics', 'organization', 'view_stats'),
    ('tenant:billing_manage', 'organization', 'Manage tenant billing', 'organization', 'billing_manage'),

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

    -- Master Data Management in Tenant
    ('character:manage', 'organization', 'Manage characters in tenant', 'character', 'manage'),
    ('creator:manage', 'organization', 'Manage creators in tenant', 'creator', 'manage'),
    ('genre:manage', 'organization', 'Manage genres in tenant', 'genre', 'manage'),
    ('relation:manage', 'organization', 'Manage relations in tenant', 'relation', 'manage'),

    -- Content Publishing
    ('content:publish', 'organization', 'Publish content', 'content', 'publish'),
    ('content:unpublish', 'organization', 'Unpublish content', 'content', 'unpublish'),
    ('analytics:view', 'organization', 'View analytics', 'analytics', 'view')
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
-- Script: Seed Admin User and OAuth2 Clients
-- Description: Seed admin user và internal OAuth2 clients
-- Author: System
-- Created: 2025-11-17

-- =====================================================
-- Install pgcrypto extension for password hashing
-- =====================================================
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =====================================================
-- Seed Admin User
-- =====================================================

-- Admin User: syuro.dev@gmail.com / Vv19082001@#
-- Using bcrypt with default cost (automatically determined by crypt)
INSERT INTO identify.users (
    id,
    email,
    email_verified,
    password_hash,
    full_name,
    avatar_url,
    status
) VALUES (
    '019af19d-2030-718d-b880-13cf816dbeac'::uuid,
    'syuro.dev@gmail.com',
    true,
    crypt('Vv19082001@#', gen_salt('bf')), -- bcrypt hash
    'System Administrator',
    'https://i.pravatar.cc/150?img=99',
    'active'
) ON CONFLICT (email) DO UPDATE SET
    password_hash = crypt('Vv19082001@#', gen_salt('bf')),
    full_name = EXCLUDED.full_name,
    email_verified = EXCLUDED.email_verified,
    status = EXCLUDED.status;

-- Assign SUPER_ADMIN role to admin user
INSERT INTO identify.user_global_roles (user_id, role_id)
SELECT
    '019af19d-2030-718d-b880-13cf816dbeac'::uuid,
    id
FROM identify.roles
WHERE name = 'SUPER_ADMIN'
ON CONFLICT DO NOTHING;

-- =====================================================
-- Seed OAuth2 Client for Wibutime Web (Internal)
-- =====================================================

INSERT INTO identify.oauth2_clients (
    id,
    client_name,
    secret_hash,
    redirect_uris,
    grant_types,
    response_types,
    scopes,
    is_public,
    organization_id,
    token_endpoint_auth_method,
    client_uri,
    logo_url,
    is_internal
) VALUES (
    '019af198-446b-748b-a2e1-53d9e326ed02'::uuid,
    'Wibutime Web (Internal)',
    crypt('7c3a2c6bc5b0ebf093bf677b39e71e7b994931bcfcd044e0d71cb19765aba7d7', gen_salt('bf')), -- bcrypt hash
    ARRAY[
        'https://wibutime.io.vn/api/auth/callback',
        'http://localhost:3000/api/auth/callback',
    ],
    ARRAY['authorization_code', 'refresh_token'],
    ARRAY['code'],
    ARRAY[
        'openid',
        'profile',
        'email',
        'offline_access',
        'read:content',
        'write:content',
        'internal',
        'admin:system'
    ],
    false, -- Confidential client (has secret)
    NULL,  -- Global/first-party client
    'client_secret_post',
    'https://wibutime.com',
    'https://wibutime.com/logo.png',
    true   -- Internal/first-party client
) ON CONFLICT (id) DO UPDATE SET
    client_name = EXCLUDED.client_name,
    secret_hash = EXCLUDED.secret_hash,
    redirect_uris = EXCLUDED.redirect_uris,
    grant_types = EXCLUDED.grant_types,
    response_types = EXCLUDED.response_types,
    scopes = EXCLUDED.scopes,
    is_internal = EXCLUDED.is_internal,
    client_uri = EXCLUDED.client_uri;

-- =====================================================
-- Display Seed Summary
-- =====================================================

SELECT '
========================================
🔐 ADMIN USER CREATED
========================================

Email: syuro.dev@gmail.com
Password: Vv19082001@#
Role: SUPER_ADMIN
User ID: 019af19d-2030-718d-b880-13cf816dbeac
Status: Active

========================================
🔑 WIBUTIME WEB CLIENT
========================================

Client Name: Wibutime Web (Internal)
Client ID: 20000000-0000-0000-0000-000000000001
Client Secret: wibutime-web-secret-2025
Type: Confidential (Internal/First-party)

Redirect URIs:
  - http://localhost:3000/auth/callback
  - http://localhost:3000/callback
  - https://wibutime.com/auth/callback
  - https://wibutime.com/callback
  - https://app.wibutime.com/auth/callback
  - https://app.wibutime.com/callback

Grant Types:
  - authorization_code
  - refresh_token

Scopes:
  - openid
  - profile
  - email
  - offline_access
  - read:content
  - write:content
  - admin:system

Authentication Method: client_secret_post

========================================
' as "Seed Summary";

-- Verify admin user and role assignment
SELECT
    u.email,
    u.full_name,
    u.status,
    r.name as role,
    r.scope
FROM identify.users u
JOIN identify.user_global_roles ugr ON u.id = ugr.user_id
JOIN identify.roles r ON ugr.role_id = r.id
WHERE u.email = 'syuro.dev@gmail.com';

-- Verify OAuth2 client
SELECT
    client_name,
    is_public,
    is_internal,
    array_length(redirect_uris, 1) as redirect_uris_count,
    array_length(scopes, 1) as scopes_count,
    token_endpoint_auth_method
FROM identify.oauth2_clients
WHERE client_name = 'Wibutime Web (Internal)';
