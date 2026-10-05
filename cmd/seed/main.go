// Comando seed: vectoriza internal/seed/profiles.json y carga ambas tablas.
//
//	go run ./cmd/seed
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/config"
	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/database"
	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/embeddings"
	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/repository"
	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/seed"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.OpenAIAPIKey == "" {
		slog.Error("OPENAI_API_KEY is required")
		os.Exit(1)
	}

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("connecting to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	embedder := embeddings.NewOpenAIClient(cfg.OpenAIEmbeddingsURL, cfg.OpenAIAPIKey, cfg.EmbeddingModel, cfg.EmbeddingDimensions)
	if err := seed.Run(ctx, repository.NewPersonRepository(pool), embedder); err != nil {
		slog.Error("seeding", "err", err)
		os.Exit(1)
	}
}
