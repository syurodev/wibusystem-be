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
