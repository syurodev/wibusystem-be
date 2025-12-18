-- =====================================================
-- Migration 000037: Drop Novel Embeddings Table (Down)
-- =====================================================

DROP INDEX IF EXISTS catalog.idx_novel_embeddings_novel_id;
DROP INDEX IF EXISTS catalog.idx_novel_embeddings_hnsw;
DROP TABLE IF EXISTS catalog.novel_embeddings;
