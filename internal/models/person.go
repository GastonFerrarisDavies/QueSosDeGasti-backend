package models

// TotalQuestions es la cantidad de respuestas que espera /api/test/match.
const TotalQuestions = 20

// PersonEmbedding es una fila de person_embeddings.
type PersonEmbedding struct {
	Name      string
	TestText  string
	Embedding []float32
}

// PersonInfo es una fila de person_info y también la respuesta de /api/test/match.
type PersonInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Phrase      string `json:"phrase"`
}

type MatchRequest struct {
	Answers []string `json:"answers"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
