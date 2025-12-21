UPDATE catalog.novel_chapters
SET status = 'scheduled',
    scheduled_at = $2
WHERE id = $1 AND deleted_at IS NULL
