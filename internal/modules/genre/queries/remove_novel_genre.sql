DELETE FROM catalog.novel_genres
WHERE novel_id = $1 AND genre_id = $2
