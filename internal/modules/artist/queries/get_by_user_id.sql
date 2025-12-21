-- GetByUserID: Lấy artist theo user ID (linking với user account)
-- Flow: Repository.GetByUserID() -> pgx.CollectOneRow()
-- Params: $1 = user UUID
SELECT id, user_id, name, slug, biography, avatar_url, social_links,
       specialization, portfolio_url, novel_count, artwork_count, follower_count,
       is_verified, metadata, version, created_by, updated_by, 
       created_at, updated_at, deleted_at, deleted_by
FROM catalog.artists
WHERE user_id = $1 AND deleted_at IS NULL
