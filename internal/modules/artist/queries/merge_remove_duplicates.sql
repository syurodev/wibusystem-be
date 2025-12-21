-- MergeRemoveDuplicates: Xóa các novel assignments còn lại của source artists
-- Flow: Repository.Merge() step 2
-- Params: $1 = source_ids (UUID[])
-- Logic: Sau khi move xong, xóa các records còn lại (duplicates)
DELETE FROM catalog.novel_artists
WHERE artist_id = ANY($1::uuid[])
