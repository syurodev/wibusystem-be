-- MergeUpdateStats: Cập nhật novel_count của target artist sau merge
-- Flow: Repository.Merge() step 3
-- Params:
--   $1 = target_id (UUID)
--   $2 = merged_by (UUID) - user thực hiện merge
UPDATE catalog.artists
SET novel_count = (SELECT COUNT(*) FROM catalog.novel_artists WHERE artist_id = $1),
    updated_by = $2,
    updated_at = NOW()
WHERE id = $1
