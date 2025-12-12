-- GetByAuthorID: Lấy danh sách novel theo author ID
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
INNER JOIN catalog.novel_authors na ON n.id = na.novel_id
LEFT JOIN identify.users u ON n.owner_type = 'user' AND n.owner_id = u.id
WHERE na.author_id = $1 AND n.deleted_at IS NULL
ORDER BY n.created_at DESC
LIMIT $2 OFFSET $3
