SELECT novel_id, genre_id
FROM catalog.novel_genres
WHERE novel_id = ANY($1)
