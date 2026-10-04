package config

import (
	"os"
	"time"
)

// Config agrupa la configuración de la API, leída desde variables de entorno.
type Config struct {
	Port           string
	DatabaseURL    string
	OllamaURL      string
	EmbeddingModel string
	AllowedOrigins string
	RequestTimeout time.Duration
}

func Load() Config {
	return Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://gasti:gasti@localhost:5432/quesosdegasti?sslmode=disable"),
		OllamaURL:      getEnv("OLLAMA_URL", "http://localhost:11434"),
		EmbeddingModel: getEnv("EMBEDDING_MODEL", "all-minilm"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		RequestTimeout: 30 * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
