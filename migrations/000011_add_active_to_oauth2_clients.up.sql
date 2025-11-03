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
