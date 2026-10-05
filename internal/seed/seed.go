// Package seed carga los perfiles de profiles.json y descriptions.json en la base.
package seed

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/embeddings"
	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/models"
	"github.com/GastonFerrarisDavies/QueSosDeGasti-backend/internal/repository"
)

// Los JSON se embeben en el binario: el seed funciona sin depender del
// directorio desde el que se ejecuta.
var (
	//go:embed profiles.json
	profilesJSON []byte
	//go:embed descriptions.json
	descriptionsJSON []byte
)

type profile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Run vectoriza cada perfil y sincroniza person_embeddings y person_info.
// Es idempotente: se puede volver a correr después de editar los JSON.
func Run(ctx context.Context, repo *repository.PersonRepository, embedder embeddings.Embedder) error {
	var profiles []profile
	if err := json.Unmarshal(profilesJSON, &profiles); err != nil {
		return fmt.Errorf("parse profiles.json: %w", err)
	}
	var infos []models.PersonInfo
	if err := json.Unmarshal(descriptionsJSON, &infos); err != nil {
		return fmt.Errorf("parse descriptions.json: %w", err)
	}
	if err := validate(profiles, infos); err != nil {
		return err
	}

	people := make([]models.PersonEmbedding, 0, len(profiles))
	for _, p := range profiles {
		vec, err := embedder.Embed(ctx, p.Description)
		if err != nil {
			return fmt.Errorf("embed %q: %w", p.Name, err)
		}
		slog.Info("embedded profile", "name", p.Name, "dims", len(vec))
		people = append(people, models.PersonEmbedding{Name: p.Name, TestText: p.Description, Embedding: vec})
	}

	if err := repo.ReplaceAll(ctx, people, infos); err != nil {
		return fmt.Errorf("save people: %w", err)
	}
	slog.Info("seed completed", "people", len(people))
	return nil
}

// validate exige que ambos JSON tengan exactamente los mismos nombres: un
// perfil sin info nunca saldría en el match (JOIN) y una info sin perfil
// violaría la FK.
func validate(profiles []profile, infos []models.PersonInfo) error {
	inProfiles := make(map[string]bool, len(profiles))
	for _, p := range profiles {
		if strings.TrimSpace(p.Name) == "" || strings.TrimSpace(p.Description) == "" {
			return fmt.Errorf("profiles.json: name and description are required (%q)", p.Name)
		}
		if inProfiles[p.Name] {
			return fmt.Errorf("profiles.json: duplicated name %q", p.Name)
		}
		inProfiles[p.Name] = true
	}

	inInfos := make(map[string]bool, len(infos))
	for _, i := range infos {
		if !inProfiles[i.Name] {
			return fmt.Errorf("descriptions.json: %q is not in profiles.json", i.Name)
		}
		if inInfos[i.Name] {
			return fmt.Errorf("descriptions.json: duplicated name %q", i.Name)
		}
		inInfos[i.Name] = true
	}
	for name := range inProfiles {
		if !inInfos[name] {
			return fmt.Errorf("descriptions.json: missing %q", name)
		}
	}
	return nil
}
