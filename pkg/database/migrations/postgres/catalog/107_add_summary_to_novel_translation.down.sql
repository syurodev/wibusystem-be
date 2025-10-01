-- Remove summary field from novel_translation table

ALTER TABLE novel_translation
DROP COLUMN IF EXISTS summary;