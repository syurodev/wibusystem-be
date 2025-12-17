-- GetNovelFullBySlug: Lấy toàn bộ dữ liệu novel trong một query
-- Sử dụng CTEs và JSON aggregation để giảm round-trips
WITH novel_data AS (
    SELECT n.id, n.title, n.slug, n.synopsis, n.cover_image_url, n.thumbnail_url,
           n.status, n.is_oneshot, n.original_language, n.original_title,
           n.owner_id, n.owner_type,
           COALESCE(u.full_name, '') as owner_display_name,
           COALESCE(u.email, '') as owner_username,
           u.avatar_url as owner_avatar_url,
           n.total_volumes, n.total_chapters, n.total_words, n.view_count,
           n.favorite_count, n.rating_average, n.rating_count, n.metadata,
           n.first_published_at, n.last_chapter_at, n.completed_at,
           n.created_by, n.updated_by, n.deleted_by,
           n.created_at, n.updated_at, n.deleted_at
    FROM catalog.novels n
    LEFT JOIN identify.users u ON n.owner_type = 'user' AND n.owner_id = u.id
    WHERE n.slug = $1 AND n.deleted_at IS NULL
),
genres_data AS (
    SELECT json_agg(json_build_object(
        'id', g.id,
        'name', g.name,
        'slug', g.slug
    ) ORDER BY ng.display_order) as genres
    FROM catalog.genres g
    INNER JOIN catalog.novel_genres ng ON g.id = ng.genre_id
    INNER JOIN novel_data nd ON ng.novel_id = nd.id
),
authors_data AS (
    SELECT json_agg(json_build_object(
        'id', a.id,
        'name', a.name,
        'slug', a.slug
    ) ORDER BY na.display_order) as authors
    FROM catalog.authors a
    INNER JOIN catalog.novel_authors na ON a.id = na.author_id
    INNER JOIN novel_data nd ON na.novel_id = nd.id
),
artists_data AS (
    SELECT json_agg(json_build_object(
        'id', a.id,
        'name', a.name,
        'slug', a.slug
    ) ORDER BY nart.display_order) as artists
    FROM catalog.artists a
    INNER JOIN catalog.novel_artists nart ON a.id = nart.artist_id
    INNER JOIN novel_data nd ON nart.novel_id = nd.id
),
volumes_data AS (
    SELECT json_agg(json_build_object(
        'id', v.id,
        'volume_number', v.volume_number,
        'title', v.title,
        'slug', v.slug,
        'cover_image_url', v.cover_image_url,
        'display_order', v.display_order,
        'is_published', v.is_published,
        'published_at', v.published_at
    ) ORDER BY v.display_order, v.volume_number) as volumes
    FROM catalog.novel_volumes v
    INNER JOIN novel_data nd ON v.novel_id = nd.id
    WHERE v.deleted_at IS NULL AND v.is_published = true
),
chapters_data AS (
    SELECT json_agg(json_build_object(
        'id', c.id,
        'volume_id', c.volume_id,
        'chapter_number', c.chapter_number,
        'title', c.title,
        'slug', c.slug,
        'display_order', c.display_order,
        'status', c.status,
        'published_at', c.published_at
    ) ORDER BY c.display_order, c.chapter_number) as chapters
    FROM catalog.novel_chapters c
    INNER JOIN novel_data nd ON c.novel_id = nd.id
    WHERE c.deleted_at IS NULL AND c.status = 'published'
)
SELECT 
    nd.*,
    COALESCE(gd.genres, '[]'::json) as genres_json,
    COALESCE(ad.authors, '[]'::json) as authors_json,
    COALESCE(artd.artists, '[]'::json) as artists_json,
    COALESCE(vd.volumes, '[]'::json) as volumes_json,
    COALESCE(cd.chapters, '[]'::json) as chapters_json
FROM novel_data nd
CROSS JOIN genres_data gd
CROSS JOIN authors_data ad
CROSS JOIN artists_data artd
CROSS JOIN volumes_data vd
CROSS JOIN chapters_data cd
