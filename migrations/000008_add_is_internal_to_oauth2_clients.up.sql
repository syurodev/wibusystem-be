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
