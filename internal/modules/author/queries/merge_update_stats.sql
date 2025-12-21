-- MergeUpdateStats: Cập nhật stats của target author sau merge
-- Flow: Repository.Merge() step 3
-- Params:
--   $1 = target_id (UUID)
--   $2 = source_ids (UUID[])
--   $3 = merged_by (UUID) - user thực hiện merge
UPDATE catalog.authors
SET total_views = total_views + (
        SELECT COALESCE(SUM(total_views), 0) FROM catalog.authors WHERE id = ANY($2::uuid[])
    ),
    novel_count = (SELECT COUNT(*) FROM catalog.novel_authors WHERE author_id = $1),
    updated_by = $3,
    updated_at = NOW()
WHERE id = $1
