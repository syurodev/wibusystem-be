UPDATE catalog.novel_volumes
SET is_published = true,
    published_at = COALESCE(published_at, NOW())
WHERE id = $1 AND deleted_at IS NULL
