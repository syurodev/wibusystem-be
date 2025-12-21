SELECT v.id, v.novel_id, v.volume_number, v.title, v.slug, v.description,
       v.cover_image_url, v.chapter_count, v.word_count, v.display_order,
       v.is_published, v.published_at, v.created_by, v.updated_by, v.deleted_by,
       v.version, v.created_at, v.updated_at, v.deleted_at,
       n.title as novel_title
FROM catalog.novel_volumes v
LEFT JOIN catalog.novels n ON v.novel_id = n.id
WHERE v.id = $1 AND v.deleted_at IS NULL
