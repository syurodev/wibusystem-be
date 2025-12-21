UPDATE catalog.novel_volumes
SET volume_number = $2,
    title = $3,
    slug = $4,
    description = $5,
    cover_image_url = $6,
    display_order = $7,
    is_published = $8
WHERE id = $1 AND deleted_at IS NULL
