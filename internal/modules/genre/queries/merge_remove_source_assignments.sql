-- Merge step 2: Remove all source genre assignments
DELETE FROM catalog.novel_genres
WHERE genre_id = ANY($1::uuid[])
