-- Migration: Add owner and admin to organization_member_role
ALTER TYPE organization_member_role ADD VALUE IF NOT EXISTS 'owner';
ALTER TYPE organization_member_role ADD VALUE IF NOT EXISTS 'admin';
