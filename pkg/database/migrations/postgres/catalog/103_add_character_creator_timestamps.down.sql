-- Reverse the character and creator table timestamp additions
DROP INDEX IF EXISTS idx_creator_name;
DROP INDEX IF EXISTS idx_character_name;

ALTER TABLE creator DROP CONSTRAINT IF EXISTS creator_name_unique;
ALTER TABLE character DROP CONSTRAINT IF EXISTS character_name_unique;

ALTER TABLE creator DROP COLUMN IF EXISTS updated_at;
ALTER TABLE creator DROP COLUMN IF EXISTS created_at;

ALTER TABLE character DROP COLUMN IF EXISTS updated_at;
ALTER TABLE character DROP COLUMN IF EXISTS created_at;