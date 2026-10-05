package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/models"
)

var ErrNoProfiles = errors.New("no profiles loaded")

type PersonRepository struct {
	db *pgxpool.Pool
}

func NewPersonRepository(db *pgxpool.Pool) *PersonRepository {
	return &PersonRepository{db: db}
}

// ReplaceAll sincroniza ambas tablas con los datos dados en una sola
// transacción: inserta o actualiza cada persona y borra las que ya no están
// (person_info se borra en cascada).
func (r *PersonRepository) ReplaceAll(ctx context.Context, people []models.PersonEmbedding, infos []models.PersonInfo) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	names := make([]string, 0, len(people))
	for _, p := range people {
		names = append(names, p.Name)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM person_embeddings WHERE name <> ALL($1)`, names); err != nil {
		return fmt.Errorf("delete stale people: %w", err)
	}

	for _, p := range people {
		const q = `
			INSERT INTO person_embeddings (name, test_text, embedding)
			VALUES ($1, $2, $3::vector)
			ON CONFLICT (name) DO UPDATE
			SET test_text = EXCLUDED.test_text, embedding = EXCLUDED.embedding`
		if _, err := tx.Exec(ctx, q, p.Name, p.TestText, toVectorLiteral(p.Embedding)); err != nil {
			return fmt.Errorf("upsert embedding %q: %w", p.Name, err)
		}
	}

	for _, i := range infos {
		const q = `
			INSERT INTO person_info (name, description, phrase)
			VALUES ($1, $2, $3)
			ON CONFLICT (name) DO UPDATE
			SET description = EXCLUDED.description, phrase = EXCLUDED.phrase`
		if _, err := tx.Exec(ctx, q, i.Name, i.Description, i.Phrase); err != nil {
			return fmt.Errorf("upsert info %q: %w", i.Name, err)
		}
	}

	return tx.Commit(ctx)
}

// FindClosest devuelve la persona con menor distancia del coseno (<=>) al vector dado.
func (r *PersonRepository) FindClosest(ctx context.Context, embedding []float32) (models.PersonInfo, error) {
	const q = `
		SELECT i.name, i.description, i.phrase
		FROM person_embeddings e
		JOIN person_info i ON e.name = i.name
		ORDER BY e.embedding <=> $1::vector
		LIMIT 1`

	var p models.PersonInfo
	err := r.db.QueryRow(ctx, q, toVectorLiteral(embedding)).Scan(&p.Name, &p.Description, &p.Phrase)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, ErrNoProfiles
	}
	return p, err
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
