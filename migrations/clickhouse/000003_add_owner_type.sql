-- Migration 000003: Add owner_type for creator/org analytics
-- Description: Adds owner_type column to distinguish between user (creator) and org

-- Add owner_type column to view_events
ALTER TABLE view_events
    ADD COLUMN IF NOT EXISTS owner_type LowCardinality(String) DEFAULT '';
