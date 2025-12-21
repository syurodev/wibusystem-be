-- GetNovelArtists: Lấy danh sách artists của một novel
-- Flow: Repository.GetNovelArtists() -> scan rows manually
-- Params: $1 = novel UUID
-- Returns: artist info + display_order từ junction table
SELECT a.id, a.user_id, a.name, a.slug, a.biography, a.avatar_url, a.social_links,
       a.specialization, a.novel_count, a.artwork_count, a.follower_count,
       a.is_verified, a.created_by, a.updated_by, a.created_at, a.updated_at, 
       a.deleted_at, a.deleted_by,
       na.display_order
FROM catalog.artists a
INNER JOIN catalog.novel_artists na ON a.id = na.artist_id
WHERE na.novel_id = $1 AND a.deleted_at IS NULL
ORDER BY na.display_order ASC
