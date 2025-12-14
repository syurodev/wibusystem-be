INSERT INTO catalog.genres (
    id, name, slug, description, parent_id,
    is_active, created_by, created_at, updated_at,
    anime_count, manga_count
) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW(), 0, 0)
