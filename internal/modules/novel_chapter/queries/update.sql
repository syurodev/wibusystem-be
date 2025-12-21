UPDATE catalog.novel_chapters
SET volume_id = $2,
    chapter_number = $3,
    title = $4,
    slug = $5,
    content = $6,
    word_count = $7,
    character_count = $8,
    is_free = $9,
    price = $10,
    currency = $11,
    status = $12,
    display_order = $13,
    author_notes = $14,
    scheduled_at = $15,
    updated_by = $16
WHERE id = $1 AND deleted_at IS NULL
