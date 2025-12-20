-- Rollback: Remove check constraint (cannot revert data changes)
ALTER TABLE catalog.novels DROP CONSTRAINT IF EXISTS novels_owner_type_check;
