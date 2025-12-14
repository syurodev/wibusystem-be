UPDATE catalog.genres
SET deleted_at = NOW(),
    deleted_by = $2
WHERE id = $1 AND deleted_at IS NULL
