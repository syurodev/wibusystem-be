INSERT INTO catalog.novel_genres (id, novel_id, genre_id, display_order, created_by, created_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
