INSERT INTO catalog.novel_genres (id, novel_id, genre_id, display_order, created_by, created_at)
VALUES (gen_random_uuid(), $1, $2, 0, $3, NOW())
ON CONFLICT (novel_id, genre_id) DO UPDATE
SET display_order = EXCLUDED.display_order
