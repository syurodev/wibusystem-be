UPDATE catalog.genres
SET total_views = total_views + $2
WHERE id = $1
