-- Migration: Add Settings to Users
-- Description: Thêm cột settings (JSONB) vào bảng users để lưu user preferences
-- Author: System
-- Created: 2025-11-19

ALTER TABLE identify.users ADD COLUMN settings JSONB DEFAULT '{}'::jsonb;

COMMENT ON COLUMN identify.users.settings IS 'User preferences (theme, language, notifications, etc.)';
