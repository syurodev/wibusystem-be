-- MergeSoftDelete: Soft delete các source authors sau khi merge
-- Flow: Repository.Merge() step 4 (final)
-- Params:
--   $1 = source_ids (UUID[])
--   $2 = deleted_by (UUID) - user thực hiện merge
UPDATE catalog.authors
SET deleted_at = NOW(),
    deleted_by = $2
WHERE id = ANY($1::uuid[])
