-- =====================================================
-- Migration 000037: Create Novel Embeddings Table
-- Description: Store vector embeddings for novel similarity search
-- =====================================================

CREATE TABLE catalog.novel_embeddings (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL UNIQUE REFERENCES catalog.novels(id) ON DELETE CASCADE,
    
    -- Vector embedding (384 dimensions for all-MiniLM-L6-v2)
    embedding vector(384),
    
    -- Metadata for tracking
    model_version VARCHAR(100) NOT NULL DEFAULT 'all-MiniLM-L6-v2',
    source_hash VARCHAR(64),  -- SHA256 of input text to detect changes
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- HNSW index for fast approximate nearest neighbor search
-- Using cosine distance (vector_cosine_ops)
CREATE INDEX idx_novel_embeddings_hnsw 
ON catalog.novel_embeddings 
USING hnsw (embedding vector_cosine_ops);

-- Index for looking up by novel_id (already UNIQUE constraint creates one, but explicit for clarity)
CREATE INDEX idx_novel_embeddings_novel_id 
ON catalog.novel_embeddings(novel_id);

-- Comments
COMMENT ON TABLE catalog.novel_embeddings IS 'Vector embeddings for novels used in similarity search';
COMMENT ON COLUMN catalog.novel_embeddings.embedding IS '384-dimensional vector from all-MiniLM-L6-v2 model';
COMMENT ON COLUMN catalog.novel_embeddings.source_hash IS 'SHA256 hash of text used to generate embedding, for detecting stale embeddings';
COMMENT ON COLUMN catalog.novel_embeddings.model_version IS 'Model used to generate embedding, for tracking and regeneration';
