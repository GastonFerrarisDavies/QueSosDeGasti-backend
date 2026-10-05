-- Se ejecuta automáticamente la primera vez que se crea el volumen de Postgres
-- (docker-entrypoint-initdb.d). Es idempotente: también se puede aplicar a una
-- base existente con
--   docker compose exec -T db psql -U $POSTGRES_USER -d $POSTGRES_DB < db/init.sql

CREATE EXTENSION IF NOT EXISTS vector;

-- Parte vectorial: el texto del test de cada persona y su embedding.
-- text-embedding-3-small devuelve 1536 dimensiones (EMBEDDING_DIMENSIONS).
-- Si cambiás ese valor, ajustá vector(N) y volvé a correr el seed.
CREATE TABLE IF NOT EXISTS person_embeddings (
    name      TEXT         PRIMARY KEY,
    test_text TEXT         NOT NULL,
    embedding vector(1536) NOT NULL
);

-- Índice HNSW para búsquedas por distancia del coseno (operador <=>).
CREATE INDEX IF NOT EXISTS person_embeddings_embedding_cosine_idx
    ON person_embeddings USING hnsw (embedding vector_cosine_ops);

-- Parte relacional: lo que se muestra como resultado.
CREATE TABLE IF NOT EXISTS person_info (
    name        TEXT PRIMARY KEY REFERENCES person_embeddings (name) ON DELETE CASCADE,
    description TEXT NOT NULL,
    phrase      TEXT NOT NULL
);
