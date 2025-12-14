-- Merge step 3: Update target genre stats (sum views, active_readers from sources)
UPDATE catalog.genres
SET total_views = total_views + (
        SELECT COALESCE(SUM(total_views), 0) FROM catalog.genres WHERE id = ANY($2::uuid[])
    ),
    active_readers = active_readers + (
        SELECT COALESCE(SUM(active_readers), 0) FROM catalog.genres WHERE id = ANY($2::uuid[])
    ),
    updated_by = $3,
    updated_at = NOW()
WHERE id = $1
