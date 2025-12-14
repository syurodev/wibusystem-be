-- Merge step 1: Move novels from source genres to target genre
-- Only move if novel doesn't already have the target genre
UPDATE catalog.novel_genres
SET genre_id = $1
WHERE genre_id = ANY($2::uuid[])
AND novel_id NOT IN (
    SELECT novel_id FROM catalog.novel_genres WHERE genre_id = $1
)
