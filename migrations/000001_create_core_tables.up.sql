-- Migration: Create Core Tables (Tenants, Users, Permissions, Roles)
-- Description: Tạo các bảng cốt lõi cho hệ thống multi-tenant với RBAC
-- Author: System
-- Created: 2025-10-26

-- Enable UUID extension nếu chưa có (PostgreSQL 18 đã có sẵn uuidv7)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- Table: tenants
-- Description: Lưu trữ thông tin về các tổ chức/khách hàng (tenant)
-- =====================================================
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Sử dụng gen_random_uuid() cho compatibility
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL, -- URL-friendly identifier (e.g., "acme-corp")
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, suspended, archived
    settings JSONB, -- Flexible settings cho mỗi tenant
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT tenants_status_check CHECK (status IN ('active', 'suspended', 'archived'))
);

-- Indexes cho tenants
CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_created_at ON tenants(created_at);

-- Comment
COMMENT ON TABLE tenants IS 'Bảng lưu trữ thông tin về các tổ chức/khách hàng trong hệ thống multi-tenant';
COMMENT ON COLUMN tenants.slug IS 'URL-friendly identifier cho tenant, dùng trong subdomain hoặc path';
COMMENT ON COLUMN tenants.settings IS 'Cấu hình tùy chỉnh cho tenant (JSONB format)';

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
    CONSTRAINT users_email_check CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
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
-- Type: permission_scope
-- Description: Phạm vi của permission (global hoặc tenant-specific)
-- =====================================================
CREATE TYPE permission_scope AS ENUM ('global', 'tenant');

COMMENT ON TYPE permission_scope IS 'Phạm vi của permission: global (toàn hệ thống) hoặc tenant (trong tenant)';

-- =====================================================
-- Table: permissions
-- Description: Danh sách master của tất cả các quyền trong hệ thống
-- =====================================================
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL, -- e.g., 'user:view_self', 'content:create_anime'
    scope permission_scope NOT NULL,
    description TEXT,
    resource VARCHAR(100), -- e.g., 'user', 'content', 'tenant'
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
-- Description: Phạm vi của role (global hoặc tenant-specific)
-- =====================================================
CREATE TYPE role_scope AS ENUM ('global', 'tenant');

COMMENT ON TYPE role_scope IS 'Phạm vi của role: global (toàn hệ thống) hoặc tenant (trong tenant)';

-- =====================================================
-- Table: roles
-- Description: Danh sách master của tất cả các vai trò
-- =====================================================
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL, -- e.g., 'SUPER_ADMIN', 'TENANT_ADMIN', 'USER'
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
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants
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
    ('tenant:create', 'global', 'Tạo tenant mới', 'tenant', 'create'),
    ('tenant:view_all', 'global', 'Xem tất cả tenants', 'tenant', 'view_all'),
    ('tenant:delete', 'global', 'Xóa tenant', 'tenant', 'delete'),
    ('user:view_all', 'global', 'Xem tất cả users trong hệ thống', 'user', 'view_all');

-- Tenant-level Permissions (trong phạm vi tenant)
INSERT INTO permissions (name, scope, description, resource, action) VALUES
    ('tenant:manage', 'tenant', 'Quản lý cấu hình tenant', 'tenant', 'manage'),
    ('user:view', 'tenant', 'Xem danh sách users trong tenant', 'user', 'view'),
    ('user:create', 'tenant', 'Tạo user mới trong tenant', 'user', 'create'),
    ('user:update', 'tenant', 'Cập nhật thông tin user', 'user', 'update'),
    ('user:delete', 'tenant', 'Xóa user khỏi tenant', 'user', 'delete'),
    ('role:view', 'tenant', 'Xem danh sách roles', 'role', 'view'),
    ('role:manage', 'tenant', 'Quản lý roles trong tenant', 'role', 'manage'),
    ('content:view', 'tenant', 'Xem nội dung', 'content', 'view'),
    ('content:create', 'tenant', 'Tạo nội dung mới', 'content', 'create'),
    ('content:update', 'tenant', 'Cập nhật nội dung', 'content', 'update'),
    ('content:delete', 'tenant', 'Xóa nội dung', 'content', 'delete');

-- =====================================================
-- Seed Data: Default Roles
-- Description: Tạo các roles cơ bản cho hệ thống
-- =====================================================

-- Global Roles (toàn hệ thống)
INSERT INTO roles (name, slug, scope, description, is_system) VALUES
    ('SUPER_ADMIN', 'super-admin', 'global', 'Quản trị viên tối cao của toàn hệ thống', TRUE),
    ('PLATFORM_ADMIN', 'platform-admin', 'global', 'Quản trị viên nền tảng', TRUE);

-- Tenant-level Roles
INSERT INTO roles (name, slug, scope, description, is_system) VALUES
    ('TENANT_ADMIN', 'tenant-admin', 'tenant', 'Quản trị viên của tenant', TRUE),
    ('TENANT_MANAGER', 'tenant-manager', 'tenant', 'Người quản lý trong tenant', TRUE),
    ('TENANT_USER', 'tenant-user', 'tenant', 'Người dùng thông thường trong tenant', TRUE),
    ('TENANT_VIEWER', 'tenant-viewer', 'tenant', 'Người xem (chỉ đọc) trong tenant', TRUE);

COMMENT ON TABLE permissions IS 'Seeded với permissions cơ bản cho system và tenant operations';
COMMENT ON TABLE roles IS 'Seeded với roles cơ bản: SUPER_ADMIN, PLATFORM_ADMIN, TENANT_ADMIN, TENANT_MANAGER, TENANT_USER, TENANT_VIEWER';
