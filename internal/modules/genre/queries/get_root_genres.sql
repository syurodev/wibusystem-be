SELECT
    id, name, slug, description, parent_id,
    is_active, novel_count, anime_count, manga_count, active_readers, total_views,
    created_by, deleted_by, updated_by, created_at, updated_at, deleted_at, version
FROM catalog.genres
WHERE parent_id IS NULL AND deleted_at IS NULL
ORDER BY name ASC
