package lesson

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

func TestMockGeneratorGenerate(t *testing.T) {
	mock := &MockGenerator{}
	skill := store.BnccSkill{
		ID:         1,
		Code:       "EF01LP01",
		Disciplina: "Português",
		Ano:        "1º ano",
		Descricao:  "Reconhecer a função social de textos que circulam em campos da vida social dos quais participa cotidianamente",
	}

	ld, err := mock.Generate(context.Background(), skill, 45)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify it's valid LessonData by re-serializing through ParseLessonData
	raw, err := json.Marshal(ld)
	if err != nil {
		t.Fatalf("failed to marshal LessonData: %v", err)
	}

	_, err = ParseLessonData(raw)
	if err != nil {
		t.Fatalf("ParseLessonData failed on Generate output: %v", err)
	}
}

func TestMockGeneratorEnhance(t *testing.T) {
	mock := &MockGenerator{}
	skill := store.BnccSkill{
		ID:         1,
		Code:       "EF01LP01",
		Disciplina: "Português",
		Ano:        "1º ano",
		Descricao:  "Test skill",
	}

	// Create a valid draft
	draft := LessonData{
		Plano: Plano{
			Objetivos:   "Entender textos",
			Metodologia: "Aula expositiva",
			Recursos:    "Quadro",
			Avaliacao:   "Prova",
		},
		Atividade: "Ler textos",
		Trilha: Trilha{
			Topicos: []Topico{
				{
					Titulo: "Introdução",
					Resumo: "Resumo do tópico",
				},
			},
			Quiz: Quiz{
				Questoes: []Questao{
					{
						Enunciado: "O que é um texto?",
						Opcoes:    []string{"A", "B"},
						Correta:   0,
					},
				},
			},
		},
	}

	enhanced, err := mock.Enhance(context.Background(), draft, skill)
	if err != nil {
		t.Fatalf("Enhance failed: %v", err)
	}

	// Verify enhanced is valid LessonData by re-serializing through ParseLessonData
	raw, err := json.Marshal(enhanced)
	if err != nil {
		t.Fatalf("failed to marshal enhanced LessonData: %v", err)
	}

	_, err = ParseLessonData(raw)
	if err != nil {
		t.Fatalf("ParseLessonData failed on Enhance output: %v", err)
	}
}
