UPDATE catalog.novel_chapters
SET view_count = view_count + 1
WHERE id = $1 AND deleted_at IS NULL
