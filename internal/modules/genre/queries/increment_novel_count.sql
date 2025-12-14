UPDATE catalog.genres
SET novel_count = novel_count + $2,
    updated_at = NOW()
WHERE id = $1
