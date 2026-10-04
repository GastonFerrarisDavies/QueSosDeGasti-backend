package models

import "time"

// TotalQuestions es la cantidad de respuestas que espera /api/test/match.
const TotalQuestions = 20

type Profile struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProfileMatch es un perfil junto con su similitud coseno (1 - distancia).
type ProfileMatch struct {
	Profile
	Similarity float64 `json:"similarity"`
}

type CreateProfileRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MatchRequest struct {
	Answers []string `json:"answers"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
