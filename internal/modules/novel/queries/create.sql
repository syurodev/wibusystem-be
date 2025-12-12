-- Create: Tạo novel mới
INSERT INTO catalog.novels (
    id, title, slug, synopsis, cover_image_url, thumbnail_url,
    status, is_oneshot, original_language, original_title, metadata,
    owner_id, owner_type,
    created_by, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
