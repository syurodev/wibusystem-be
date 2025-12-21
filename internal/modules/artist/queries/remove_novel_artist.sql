-- RemoveNovelArtist: Xóa artist khỏi novel
-- Flow: Repository.RemoveNovelArtist()
-- Params:
--   $1 = novel_id (UUID)
--   $2 = artist_id (UUID)
--   $3 = role (string)
DELETE FROM catalog.novel_artists
WHERE novel_id = $1 AND artist_id = $2 AND role = $3
