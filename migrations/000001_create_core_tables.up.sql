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
    id UUID PRIMARY KEY DEFAULT uuidv7(),
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
    id UUID PRIMARY KEY DEFAULT uuidv7(),
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
    id UUID PRIMARY KEY DEFAULT uuidv7(),
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
    id UUID PRIMARY KEY DEFAULT uuidv7(),
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
