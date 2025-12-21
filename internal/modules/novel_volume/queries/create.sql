INSERT INTO catalog.novel_volumes (
    id, novel_id, volume_number, title, slug, description,
    cover_image_url, display_order, is_published, created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
