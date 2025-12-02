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
