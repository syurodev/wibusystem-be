-- MergeMoveNovels: Di chuyển novels từ source artists sang target artist
-- Flow: Repository.Merge() step 1
-- Params:
--   $1 = target_id (UUID) - artist đích
--   $2 = source_ids (UUID[]) - mảng các artist nguồn
-- Logic: Chỉ move novels chưa có trong target, tránh duplicate
UPDATE catalog.novel_artists
SET artist_id = $1
WHERE artist_id = ANY($2::uuid[])
AND novel_id NOT IN (
    SELECT novel_id FROM catalog.novel_artists WHERE artist_id = $1
)
