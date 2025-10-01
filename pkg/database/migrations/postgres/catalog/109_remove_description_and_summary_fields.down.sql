-- Restore description and summary fields

-- Add back description field to novel_translation
ALTER TABLE novel_translation
ADD COLUMN description TEXT;

-- Add back summary field to novel
ALTER TABLE novel
ADD COLUMN summary JSONB;