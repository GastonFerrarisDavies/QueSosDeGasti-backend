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

type ProfileHandler struct {
	repo     *repository.ProfileRepository
	embedder embeddings.Embedder
	timeout  time.Duration
}

func NewProfileHandler(repo *repository.ProfileRepository, embedder embeddings.Embedder, timeout time.Duration) *ProfileHandler {
	return &ProfileHandler{repo: repo, embedder: embedder, timeout: timeout}
}

// CreateProfile maneja POST /api/profiles.
func (h *ProfileHandler) CreateProfile(c *fiber.Ctx) error {
	var req models.CreateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid JSON body")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" || req.Description == "" {
		return badRequest(c, "name and description are required")
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), h.timeout)
	defer cancel()

	vec, err := h.embedder.Embed(ctx, req.Description)
	if err != nil {
		slog.Error("embedding profile", "err", err)
		return c.Status(fiber.StatusBadGateway).JSON(models.ErrorResponse{Error: "could not generate embedding"})
	}

	profile, err := h.repo.Create(ctx, req.Name, req.Description, vec)
	if err != nil {
		slog.Error("saving profile", "err", err)
		return internalError(c)
	}

	return c.Status(fiber.StatusCreated).JSON(profile)
}

// Match maneja POST /api/test/match.
func (h *ProfileHandler) Match(c *fiber.Ctx) error {
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
		if !strings.HasSuffix(a, ".") {
			a += "."
		}
		answers = append(answers, a)
	}
	paragraph := strings.Join(answers, " ")

	ctx, cancel := context.WithTimeout(c.UserContext(), h.timeout)
	defer cancel()

	vec, err := h.embedder.Embed(ctx, paragraph)
	if err != nil {
		slog.Error("embedding answers", "err", err)
		return c.Status(fiber.StatusBadGateway).JSON(models.ErrorResponse{Error: "could not generate embedding"})
	}

	match, err := h.repo.FindClosest(ctx, vec)
	if errors.Is(err, repository.ErrNoProfiles) {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{Error: "no profiles loaded yet"})
	}
	if err != nil {
		slog.Error("finding closest profile", "err", err)
		return internalError(c)
	}

	return c.JSON(match)
}

func badRequest(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: msg})
}

func internalError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "internal server error"})
}
