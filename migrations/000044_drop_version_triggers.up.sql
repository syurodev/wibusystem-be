-- Migration: Drop version triggers from novels and chapters
-- Version increment will be handled in application code instead of database triggers
-- This prevents unwanted version increments when only updating view_count or other non-critical fields

DROP TRIGGER IF EXISTS trg_novels_version ON catalog.novels;
DROP TRIGGER IF EXISTS trg_chapters_version ON catalog.novel_chapters;

-- Also drop the function if it's no longer used by any other trigger
-- Check if function is used elsewhere first before dropping
-- DROP FUNCTION IF EXISTS catalog.increment_version();
