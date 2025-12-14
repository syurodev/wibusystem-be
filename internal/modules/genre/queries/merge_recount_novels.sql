-- Merge step 4: Recalculate novel_count for target genre
UPDATE catalog.genres
SET novel_count = (SELECT COUNT(*) FROM catalog.novel_genres WHERE genre_id = $1)
WHERE id = $1
