package handlers

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/embeddings"
	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/models"
	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/repository"
)

type MatchHandler struct {
	repo     *repository.PersonRepository
	embedder embeddings.Embedder
	timeout  time.Duration
}

func NewMatchHandler(repo *repository.PersonRepository, embedder embeddings.Embedder, timeout time.Duration) *MatchHandler {
	return &MatchHandler{repo: repo, embedder: embedder, timeout: timeout}
}

// Match maneja POST /api/test/match.
func (h *MatchHandler) Match(c *fiber.Ctx) error {
	var req models.MatchRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid JSON body")
	}
	if len(req.Answers) != models.TotalQuestions {
		return badRequest(c, "exactly 20 answers are required")
	}

	answers := make([]string, 0, len(req.Answers))
	for _, a := range req.Answers {
		a = strings.TrimSpace(a)
		if a == "" {
			return badRequest(c, "answers cannot be empty")
		}
		answers = append(answers, a)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), h.timeout)
	defer cancel()

	vec, err := h.embedder.Embed(ctx, JoinAnswers(answers))
	if err != nil {
		slog.Error("embedding answers", "err", err)
		return c.Status(fiber.StatusBadGateway).JSON(models.ErrorResponse{Error: "could not generate embedding"})
	}

	match, err := h.repo.FindClosest(ctx, vec)
	if errors.Is(err, repository.ErrNoProfiles) {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{Error: "no profiles loaded yet"})
	}
	if err != nil {
		slog.Error("finding closest person", "err", err)
		return internalError(c)
	}

	return c.JSON(match)
}

// JoinAnswers arma el párrafo a vectorizar: "Respuesta 1. Respuesta 2. ... Respuesta 20."
// Es el mismo formato que los textos de seed/profiles.json, para que los
// vectores sean comparables.
func JoinAnswers(answers []string) string {
	parts := make([]string, 0, len(answers))
	for _, a := range answers {
		parts = append(parts, strings.TrimRight(strings.TrimSpace(a), "."))
	}
	return strings.Join(parts, ". ") + "."
}

func badRequest(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: msg})
}

func internalError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "internal server error"})
}
