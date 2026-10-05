# QueSosDeGasti — Backend

API en Go (Fiber + pgx) que usa embeddings de OpenAI (`text-embedding-3-small`, 1536 dims) y pgvector para encontrar a qué conocido de Gasti te parecés.

## Estructura

```
cmd/api/main.go                 # Entry point de la API: config, DB, rutas
cmd/seed/main.go                # Comando que carga los perfiles en la base
internal/config/                # Variables de entorno
internal/database/              # Pool pgx con reintentos
internal/embeddings/            # Cliente HTTP de OpenAI (/v1/embeddings)
internal/models/                # DTOs y entidades
internal/repository/            # Queries SQL (seed y búsqueda por coseno)
internal/handlers/              # Handler de /api/test/match
internal/seed/profiles.json     # Respuestas concatenadas de cada persona (se vectorizan)
internal/seed/descriptions.json # Descripción y frase de cada persona (se muestran)
db/init.sql                     # Extensión vector + tablas person_embeddings y person_info
```

## Base de datos

- `person_embeddings (name PK, test_text, embedding vector(1536))`: la parte vectorial.
- `person_info (name PK/FK → person_embeddings, description, phrase)`: lo que se devuelve como resultado.

## Levantar en local

```bash
cp .env.example .env    # completar OPENAI_API_KEY
docker compose up --build
```

`init.sql` corre solo la primera vez que se crea `./data/postgres`. Es idempotente, así que para aplicarlo a una base existente:

```bash
docker compose exec -T db sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"' < db/init.sql
```

### Cargar los perfiles (seed)

Vectoriza cada `description` de `profiles.json` y sincroniza ambas tablas en una transacción (inserta, actualiza y borra a quien ya no esté en los JSON). Ambos archivos deben tener exactamente los mismos `name`. Volvé a correrlo cada vez que los edites:

```bash
# Con las variables de .env cargadas y la DB expuesta en localhost:5432
DATABASE_URL="postgres://USER:PASS@localhost:5432/quesosdegasti?sslmode=disable" go run ./cmd/seed
```

Variables de embeddings (en `.env`):

| Variable | Default | Descripción |
|---|---|---|
| `OPENAI_API_KEY` | — (obligatoria) | API key de OpenAI |
| `OPENAI_EMBEDDINGS_URL` | `https://api.openai.com/v1/embeddings` | Endpoint al que se envía la petición |
| `EMBEDDING_MODEL` | `text-embedding-3-small` | Modelo de embeddings |
| `EMBEDDING_DIMENSIONS` | `1536` | Dimensiones pedidas; debe coincidir con `vector(N)` en `init.sql` |

## Endpoints

### `POST /api/test/match`

Recibe exactamente 20 respuestas. Se unen con `". "` (mismo formato que `profiles.json`), se vectorizan y se busca la persona más cercana por distancia del coseno (`<=>`).

```bash
curl -X POST http://localhost:8080/api/test/match \
  -H "Content-Type: application/json" \
  -d '{"answers":["...", "...", "... (20 en total)"]}'
```

Respuesta:

```json
{ "name": "Emir", "description": "...", "phrase": "..." }
```

### `GET /health`

## Notas

- Si cambiás `EMBEDDING_DIMENSIONS` o el modelo, ajustá `vector(1536)` en `init.sql`, recreá las tablas y volvé a correr el seed (los vectores de modelos distintos no son comparables).
- Los textos de `profiles.json` tienen que estar escritos con las mismas opciones de `q.json` del frontend para que los vectores sean comparables.
