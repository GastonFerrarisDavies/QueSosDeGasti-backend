-- Se ejecuta automáticamente la primera vez que se crea el volumen de Postgres
-- (docker-entrypoint-initdb.d). Si cambiás este archivo, borrá el volumen:
--   docker compose down -v

CREATE EXTENSION IF NOT EXISTS vector;

-- all-minilm (all-MiniLM-L6-v2) genera vectores de 384 dimensiones.
CREATE TABLE IF NOT EXISTS profiles (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL,
    embedding   vector(384) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Índice HNSW para búsquedas por distancia del coseno (operador <=>).
CREATE INDEX IF NOT EXISTS profiles_embedding_cosine_idx
    ON profiles USING hnsw (embedding vector_cosine_ops);
