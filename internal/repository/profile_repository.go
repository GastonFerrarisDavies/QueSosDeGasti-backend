package repository

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/models"
)

var ErrNoProfiles = errors.New("no profiles loaded")

type ProfileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) Create(ctx context.Context, name, description string, embedding []float32) (models.Profile, error) {
	const q = `
		INSERT INTO profiles (name, description, embedding)
		VALUES ($1, $2, $3::vector)
		RETURNING id, name, description, created_at`

	var p models.Profile
	err := r.db.QueryRow(ctx, q, name, description, toVectorLiteral(embedding)).
		Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt)
	return p, err
}

// FindClosest devuelve el perfil con menor distancia del coseno (<=>) al vector dado.
func (r *ProfileRepository) FindClosest(ctx context.Context, embedding []float32) (models.ProfileMatch, error) {
	const q = `
		SELECT id, name, description, created_at, 1 - (embedding <=> $1::vector) AS similarity
		FROM profiles
		ORDER BY embedding <=> $1::vector
		LIMIT 1`

	var m models.ProfileMatch
	err := r.db.QueryRow(ctx, q, toVectorLiteral(embedding)).
		Scan(&m.ID, &m.Name, &m.Description, &m.CreatedAt, &m.Similarity)
	if errors.Is(err, pgx.ErrNoRows) {
		return m, ErrNoProfiles
	}
	return m, err
}

// toVectorLiteral serializa el vector al formato de texto de pgvector: "[0.1,0.2,...]".
func toVectorLiteral(v []float32) string {
	var b strings.Builder
	b.Grow(len(v) * 12)
	b.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(f), 'f', -1, 32))
	}
	b.WriteByte(']')
	return b.String()
}
