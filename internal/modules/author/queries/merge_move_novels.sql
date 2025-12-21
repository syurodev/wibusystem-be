-- MergeMoveNovels: Di chuyển novels từ source authors sang target author
-- Flow: Repository.Merge() step 1
-- Params:
--   $1 = target_id (UUID) - author đích
--   $2 = source_ids (UUID[]) - mảng các author nguồn
-- Logic: Chỉ move novels chưa có trong target, tránh duplicate
UPDATE catalog.novel_authors
SET author_id = $1
WHERE author_id = ANY($2::uuid[])
AND novel_id NOT IN (
    SELECT novel_id FROM catalog.novel_authors WHERE author_id = $1
)
