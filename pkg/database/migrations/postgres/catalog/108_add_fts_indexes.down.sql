-- Remove Full Text Search (FTS) indexes

DROP INDEX IF EXISTS idx_novel_meta_description_fts;
DROP INDEX IF EXISTS idx_novel_keywords_fts;
DROP INDEX IF EXISTS idx_novel_translation_title_fts;