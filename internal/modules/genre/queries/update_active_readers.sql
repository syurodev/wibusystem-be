UPDATE catalog.genres
SET active_readers = $2
WHERE id = $1
