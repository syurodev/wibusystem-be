UPDATE catalog.novel_volumes v
SET 
    chapter_count = COALESCE((
        SELECT COUNT(*)
        FROM catalog.novel_chapters c
        WHERE c.volume_id = v.id AND c.deleted_at IS NULL
    ), 0),
    word_count = COALESCE((
        SELECT SUM(word_count)
        FROM catalog.novel_chapters c
        WHERE c.volume_id = v.id AND c.deleted_at IS NULL
    ), 0)
WHERE v.id = $1 AND v.deleted_at IS NULL
