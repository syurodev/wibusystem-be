-- GetBySlug: Lấy author theo slug
-- Flow: Repository.GetBySlug() -> pgx.CollectOneRow()
-- Params: $1 = author slug
SELECT id, user_id, name, slug, biography, avatar_url, social_links,
       novel_count, total_chapters, total_views, follower_count,
       is_verified, metadata, version, created_by, updated_by, 
       created_at, updated_at, deleted_at, deleted_by
FROM catalog.authors
WHERE slug = $1 AND deleted_at IS NULL
