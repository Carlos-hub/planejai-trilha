package lesson

import (
	"encoding/json"
	"fmt"
)

type LessonData struct {
	Plano     Plano  `json:"plano"`
	Atividade string `json:"atividade"`
	Trilha    Trilha `json:"trilha"`
}

type Plano struct {
	Objetivos   string `json:"objetivos"`
	Metodologia string `json:"metodologia"`
	Recursos    string `json:"recursos"`
	Avaliacao   string `json:"avaliacao"`
}

type Trilha struct {
	Topicos []Topico `json:"topicos"`
	Quiz    Quiz     `json:"quiz"`
}

type Topico struct {
	Titulo string `json:"titulo"`
	Resumo string `json:"resumo"`
}

type Quiz struct {
	Questoes []Questao `json:"questoes"`
}

type Questao struct {
	Enunciado string   `json:"enunciado"`
	Opcoes    []string `json:"opcoes"`
	Correta   int      `json:"correta"`
}

// ParseLessonData unmarshals raw JSON bytes into LessonData and validates:
// - at least 1 topico
// - at least 1 questao
// - each questao has at least 2 opcoes
// - each questao has 0 <= correta < len(opcoes)
func ParseLessonData(raw []byte) (LessonData, error) {
	var ld LessonData
	if err := json.Unmarshal(raw, &ld); err != nil {
		return LessonData{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validate at least 1 topico
	if len(ld.Trilha.Topicos) < 1 {
		return LessonData{}, fmt.Errorf("validation error: must have at least 1 topico, got %d", len(ld.Trilha.Topicos))
	}

	// Validate at least 1 questao
	if len(ld.Trilha.Quiz.Questoes) < 1 {
		return LessonData{}, fmt.Errorf("validation error: must have at least 1 questao, got %d", len(ld.Trilha.Quiz.Questoes))
	}

	// Validate each questao
	for i, q := range ld.Trilha.Quiz.Questoes {
		// Each questao must have at least 2 opcoes
		if len(q.Opcoes) < 2 {
			return LessonData{}, fmt.Errorf("validation error: questao %d must have at least 2 opcoes, got %d", i, len(q.Opcoes))
		}
		// correta must be in valid range [0, len(opcoes))
		if q.Correta < 0 || q.Correta >= len(q.Opcoes) {
			return LessonData{}, fmt.Errorf("validation error: questao %d has invalid correta=%d for %d opcoes (must be 0 <= correta < %d)", i, q.Correta, len(q.Opcoes), len(q.Opcoes))
		}
	}

	return ld, nil
}
