-- AddNovelAuthor: Thêm author cho novel
-- Flow: Repository.AddNovelAuthor() -> ON CONFLICT upsert
-- Params:
--   $1 = novel_id (UUID)
--   $2 = author_id (UUID)
--   $3 = display_order (int)
INSERT INTO catalog.novel_authors (novel_id, author_id, display_order)
VALUES ($1, $2, $3)
ON CONFLICT (novel_id, author_id) DO UPDATE
SET display_order = EXCLUDED.display_order
