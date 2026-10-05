package config

import (
	"os"
	"strconv"
	"time"
)

// Config agrupa la configuración de la API, leída desde variables de entorno.
type Config struct {
	Port                string
	DatabaseURL         string
	OpenAIEmbeddingsURL string
	OpenAIAPIKey        string
	EmbeddingModel      string
	EmbeddingDimensions int
	AllowedOrigins      string
	RequestTimeout      time.Duration
}

func Load() Config {
	return Config{
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://gasti:gasti@localhost:5432/quesosdegasti?sslmode=disable"),
		OpenAIEmbeddingsURL: getEnv("OPENAI_EMBEDDINGS_URL", "https://api.openai.com/v1/embeddings"),
		OpenAIAPIKey:        getEnv("OPENAI_API_KEY", ""),
		EmbeddingModel:      getEnv("EMBEDDING_MODEL", "text-embedding-3-small"),
		// Debe coincidir con vector(N) en db/init.sql.
		EmbeddingDimensions: getEnvInt("EMBEDDING_DIMENSIONS", 1536),
		AllowedOrigins:      getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		RequestTimeout:      30 * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if n, err := strconv.Atoi(getEnv(key, "")); err == nil {
		return n
	}
	return fallback
}
