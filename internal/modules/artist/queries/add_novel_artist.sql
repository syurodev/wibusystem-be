-- AddNovelArtist: Thêm artist cho novel
-- Flow: Repository.AddNovelArtist() -> ON CONFLICT upsert
-- Params:
--   $1 = novel_id (UUID)
--   $2 = artist_id (UUID)
--   $3 = display_order (int)
INSERT INTO catalog.novel_artists (novel_id, artist_id, display_order)
VALUES ($1, $2, $3)
ON CONFLICT (novel_id, artist_id) DO UPDATE
SET display_order = EXCLUDED.display_order
