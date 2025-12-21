-- GetNovelAuthors: Lấy danh sách authors của một novel
-- Flow: Repository.GetNovelAuthors() -> scan rows manually
-- Params: $1 = novel UUID
-- Returns: author info + display_order từ junction table
SELECT a.id, a.user_id, a.name, a.slug, a.biography, a.avatar_url, a.social_links,
       a.novel_count, a.total_chapters, a.total_views, a.follower_count,
       a.is_verified, a.created_by, a.updated_by, a.created_at, a.updated_at, 
       a.deleted_at, a.deleted_by,
       na.display_order
FROM catalog.authors a
INNER JOIN catalog.novel_authors na ON a.id = na.author_id
WHERE na.novel_id = $1 AND a.deleted_at IS NULL
ORDER BY na.display_order ASC
