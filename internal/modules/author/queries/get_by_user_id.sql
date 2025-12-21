-- GetByUserID: Lấy author theo user ID (linking với user account)
-- Flow: Repository.GetByUserID() -> pgx.CollectOneRow()
-- Params: $1 = user UUID
SELECT id, user_id, name, slug, biography, avatar_url, social_links,
       novel_count, total_chapters, total_views, follower_count,
       is_verified, metadata, version, created_by, updated_by, 
       created_at, updated_at, deleted_at, deleted_by
FROM catalog.authors
WHERE user_id = $1 AND deleted_at IS NULL
