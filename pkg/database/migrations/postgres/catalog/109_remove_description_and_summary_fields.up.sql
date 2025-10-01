-- Remove description field from novel_translation and summary field from novel
-- Consolidate to use only summary field in novel_translation

-- Remove description field from novel_translation (keep only title and summary)
ALTER TABLE novel_translation
DROP COLUMN IF EXISTS description;

-- Remove summary field from novel (now handled by novel_translation)
ALTER TABLE novel
DROP COLUMN IF EXISTS summary;