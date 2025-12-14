SELECT g.id, g.name, g.slug, g.description, g.parent_id,
       g.is_active, g.novel_count, g.anime_count, g.manga_count, g.active_readers, g.total_views,
       g.created_by, g.updated_by, g.created_at, g.updated_at
FROM catalog.genres g
INNER JOIN catalog.novel_genres ng ON g.id = ng.genre_id
WHERE ng.novel_id = $1 AND g.deleted_at IS NULL
ORDER BY ng.display_order ASC, g.name ASC
