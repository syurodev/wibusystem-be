UPDATE catalog.novel_chapters
SET status = 'published',
    published_at = COALESCE(published_at, NOW()),
    scheduled_at = NULL
WHERE id = $1 AND deleted_at IS NULL
