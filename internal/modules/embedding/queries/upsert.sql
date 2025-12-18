-- Upsert a novel embedding
-- $1: novel_id
-- $2: embedding (vector)
-- $3: model_version
-- $4: source_hash
INSERT INTO catalog.novel_embeddings (novel_id, embedding, model_version, source_hash)
VALUES ($1, $2, $3, $4)
ON CONFLICT (novel_id) DO UPDATE SET
    embedding = EXCLUDED.embedding,
    model_version = EXCLUDED.model_version,
    source_hash = EXCLUDED.source_hash,
    updated_at = NOW();
