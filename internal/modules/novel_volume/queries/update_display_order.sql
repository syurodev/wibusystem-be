UPDATE catalog.novel_volumes
SET display_order = $2
WHERE id = $1 AND deleted_at IS NULL
