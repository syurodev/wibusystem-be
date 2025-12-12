-- IncrementViewCount: Tăng view count của novel
UPDATE catalog.novels
SET view_count = view_count + 1,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
