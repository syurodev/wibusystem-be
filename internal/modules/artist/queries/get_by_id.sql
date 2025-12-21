-- GetByID: Lấy artist theo ID
-- Flow: Repository.GetByID() -> pgx.CollectOneRow()
-- Params: $1 = artist UUID
SELECT id, user_id, name, slug, biography, avatar_url, social_links,
       specialization, portfolio_url, novel_count, artwork_count, follower_count,
       is_verified, metadata, version, created_by, updated_by, 
       created_at, updated_at, deleted_at, deleted_by
FROM catalog.artists
WHERE id = $1 AND deleted_at IS NULL
