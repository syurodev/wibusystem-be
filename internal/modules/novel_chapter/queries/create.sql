INSERT INTO catalog.novel_chapters (
    id, novel_id, volume_id, chapter_number, title, slug, content,
    word_count, character_count, is_free, price, currency, status,
    display_order, author_notes, scheduled_at, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
