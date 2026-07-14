package lesson

import (
	"context"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// MockGenerator is a mock implementation of Generator that returns canned LessonData.
// Used in tests and as a placeholder for handler development.
type MockGenerator struct{}

// Generate returns a canned valid LessonData.
func (m *MockGenerator) Generate(ctx context.Context, skill store.BnccSkill, duracaoMin int) (LessonData, error) {
	return LessonData{
		Plano: Plano{
			Objetivos:   "Desenvolver compreensão de " + skill.Descricao,
			Metodologia: "Aula expositiva com exemplos práticos",
			Recursos:    "Quadro, projetor, material impresso",
			Avaliacao:   "Prova escrita e atividades práticas",
		},
		Atividade: "Realizar leitura e análise de textos",
		Trilha: Trilha{
			Topicos: []Topico{
				{
					Titulo: "Introdução ao tópico",
					Resumo: "Conceitos fundamentais relacionados a " + skill.Descricao,
				},
				{
					Titulo: "Desenvolvimento",
					Resumo: "Aprofundamento dos conceitos apresentados",
				},
			},
			Quiz: Quiz{
				Questoes: []Questao{
					{
						Enunciado: "Qual é o conceito principal abordado nesta aula?",
						Opcoes:    []string{"Opção A incorreta", "Opção B correta", "Opção C incorreta"},
						Correta:   1,
					},
					{
						Enunciado: "Como aplicar o conceito aprendido?",
						Opcoes:    []string{"Forma A", "Forma B", "Forma C correta"},
						Correta:   2,
					},
				},
			},
		},
	}, nil
}

// Enhance returns the draft with a minor enhancement (appends " (aprimorado)" to objetivos).
// The result is still valid LessonData.
func (m *MockGenerator) Enhance(ctx context.Context, draft LessonData, skill store.BnccSkill) (LessonData, error) {
	enhanced := draft
	enhanced.Plano.Objetivos = draft.Plano.Objetivos + " (aprimorado)"
	return enhanced, nil
}
