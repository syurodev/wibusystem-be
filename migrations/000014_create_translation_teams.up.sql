-- Migration: Delete Translation Teams (Merged into Organizations)
-- Description: This migration is no longer needed as translation_teams have been merged into organizations table
-- The functionality from this migration has been incorporated into:
-- - Migration 000001: organizations table (merged from translation_teams)
-- - Migration 000002: user_organization_memberships (merged from team_members)
--
-- This file is kept empty to maintain migration numbering sequence.
-- The corresponding features are now in the identify.organizations table.

-- No operations needed - features merged into earlier migrations
