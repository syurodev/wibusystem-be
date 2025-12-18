-- Get embedding by novel_id
-- $1: novel_id
SELECT id, novel_id, embedding, model_version, source_hash, created_at, updated_at
FROM catalog.novel_embeddings
WHERE novel_id = $1;
