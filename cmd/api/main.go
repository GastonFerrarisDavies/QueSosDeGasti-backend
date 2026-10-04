package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/config"
	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/database"
	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/embeddings"
	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/handlers"
	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/repository"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("connecting to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	embedder := embeddings.NewOllamaClient(cfg.OllamaURL, cfg.EmbeddingModel)
	profileHandler := handlers.NewProfileHandler(repository.NewProfileRepository(pool), embedder, cfg.RequestTimeout)

	app := fiber.New(fiber.Config{AppName: "QueSosDeGasti API"})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.AllowedOrigins,
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		if err := pool.Ping(c.UserContext()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "db unavailable"})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api")
	api.Post("/profiles", profileHandler.CreateProfile)
	api.Post("/test/match", profileHandler.Match)

	go func() {
		<-ctx.Done()
		_ = app.Shutdown()
	}()

	slog.Info("starting api", "port", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
