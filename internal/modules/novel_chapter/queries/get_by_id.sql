SELECT id, novel_id, volume_id, chapter_number, title, slug, content,
       word_count, character_count, is_free, price, currency, status,
       view_count, like_count, comment_count, display_order, author_notes,
       published_at, scheduled_at, created_at, updated_at, deleted_at,
       created_by, updated_by, deleted_by, version
FROM catalog.novel_chapters
WHERE id = $1 AND deleted_at IS NULL
