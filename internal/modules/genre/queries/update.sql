UPDATE catalog.genres
SET name = $2,
    slug = $3,
    description = $4,
    parent_id = $5,
    is_active = $6,
    updated_by = $7,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
