UPDATE catalog.novel_volumes
SET is_published = false
WHERE id = $1 AND deleted_at IS NULL
