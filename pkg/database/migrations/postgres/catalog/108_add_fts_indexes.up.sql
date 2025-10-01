-- Add Full Text Search (FTS) indexes for better search performance
-- Using 'simple' configuration for language-agnostic search

-- FTS index for novel_translation.title
CREATE INDEX idx_novel_translation_title_fts ON novel_translation
USING gin(to_tsvector('simple', title));

-- FTS index for novel.keywords
CREATE INDEX idx_novel_keywords_fts ON novel
USING gin(to_tsvector('simple', keywords));

-- FTS index for novel.meta_description
CREATE INDEX idx_novel_meta_description_fts ON novel
USING gin(to_tsvector('simple', meta_description));