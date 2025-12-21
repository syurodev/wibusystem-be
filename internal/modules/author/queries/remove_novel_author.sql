-- RemoveNovelAuthor: Xóa author khỏi novel
-- Flow: Repository.RemoveNovelAuthor()
-- Params:
--   $1 = novel_id (UUID)
--   $2 = author_id (UUID)
DELETE FROM catalog.novel_authors
WHERE novel_id = $1 AND author_id = $2
