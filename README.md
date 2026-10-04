# QueSosDeGasti — Backend

API en Go (Fiber + pgx) que usa embeddings de Ollama (`all-minilm`, all-MiniLM-L6-v2) y pgvector para encontrar a qué conocido de Gasti te parecés.

## Estructura

```
cmd/api/main.go                 # Entry point: config, DB, rutas
internal/config/                # Variables de entorno
internal/database/              # Pool pgx con reintentos
internal/embeddings/            # Cliente HTTP de Ollama (/api/embed)
internal/models/                # DTOs y entidades
internal/repository/            # Queries SQL (INSERT y búsqueda por coseno)
internal/handlers/              # Handlers HTTP
db/init.sql                     # Extensión vector + tabla profiles
```

## Levantar en local

```bash
cp .env.example .env    # opcional
docker compose up --build
```

La primera vez, el servicio `llm` descarga el modelo (~45 MB); la API espera a que esté listo antes de arrancar.
Los datos de Postgres quedan en `./data/postgres`. Para recrear la base desde `init.sql`: `docker compose down && rm -rf data/`.

## Endpoints

### `POST /api/profiles`

```bash
curl -X POST http://localhost:8080/api/profiles \
  -H "Content-Type: application/json" \
  -d '{"name":"Juan","description":"Le encanta el asado, es fanático del fútbol y siempre llega tarde."}'
```

### `POST /api/test/match`

Recibe exactamente 20 respuestas. Se concatenan en un párrafo, se vectorizan y se busca el perfil más cercano por distancia del coseno (`<=>`).

```bash
curl -X POST http://localhost:8080/api/test/match \
  -H "Content-Type: application/json" \
  -d '{"answers":["...", "...", "... (20 en total)"]}'
```

Respuesta:

```json
{ "id": 1, "name": "Juan", "description": "...", "created_at": "...", "similarity": 0.71 }
```

### `GET /health`

## Notas

- `all-MiniLM-L6-v2` está entrenado principalmente en inglés y trunca la entrada a ~256 tokens. Funciona con español, pero si la precisión no alcanza, se puede cambiar `EMBEDDING_MODEL` por un modelo multilingüe (p. ej. `paraphrase-multilingual`, 768 dims) ajustando `vector(384)` en `init.sql`.
- Las descripciones de perfiles conviene escribirlas en el mismo estilo que las respuestas del test (primera persona, frases descriptivas) para que los vectores sean comparables.
