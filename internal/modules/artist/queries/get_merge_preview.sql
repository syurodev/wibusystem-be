-- GetMergePreview: Lấy danh sách novels sẽ bị ảnh hưởng khi merge
-- Flow: Repository.GetMergePreview()
-- Params: $1 = source_ids (UUID[])
-- Returns: Distinct novels thuộc về các source artists
SELECT DISTINCT n.id, n.title, n.slug, n.cover_image_url
FROM catalog.novels n
JOIN catalog.novel_artists na ON n.id = na.novel_id
WHERE na.artist_id = ANY($1::uuid[])
AND n.deleted_at IS NULL
ORDER BY n.title ASC
