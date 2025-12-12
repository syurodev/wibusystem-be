-- Update: Cập nhật thông tin novel
UPDATE catalog.novels
SET title = $2,
    slug = $3,
    synopsis = $4,
    cover_image_url = $5,
    thumbnail_url = $6,
    status = $7,
    is_oneshot = $8,
    original_language = $9,
    original_title = $10,
    metadata = $11,
    first_published_at = $12,
    completed_at = $13,
    updated_by = $14,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
