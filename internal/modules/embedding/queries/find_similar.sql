-- Find similar novels by embedding distance
-- $1: source embedding (vector)
-- $2: source novel_id (to exclude)
-- $3: limit
SELECT 
    ne.novel_id,
    ne.embedding <=> $1 AS distance
FROM catalog.novel_embeddings ne
JOIN catalog.novels n ON n.id = ne.novel_id
WHERE ne.novel_id != $2 
  AND n.deleted_at IS NULL
ORDER BY distance
LIMIT $3;
