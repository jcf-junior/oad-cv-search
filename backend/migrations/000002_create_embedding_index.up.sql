CREATE INDEX candidates_embedding_idx
    ON candidates
    USING hnsw (embedding vector_cosine_ops);
