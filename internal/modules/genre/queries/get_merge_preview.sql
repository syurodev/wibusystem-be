-- Get merge preview: novels affected by merge
SELECT DISTINCT n.id, n.title, n.slug, n.cover_image_url
FROM catalog.novels n
JOIN catalog.novel_genres ng ON n.id = ng.novel_id
WHERE ng.genre_id = ANY($1::uuid[])
AND n.deleted_at IS NULL
ORDER BY n.title ASC
