package lesson

import "testing"

func TestParseLessonData(t *testing.T) {
	raw := []byte(`{"plano":{"objetivos":"o","metodologia":"m","recursos":"r","avaliacao":"a"},"atividade":"at","trilha":{"topicos":[{"titulo":"t","resumo":"r"}],"quiz":{"questoes":[{"enunciado":"e","opcoes":["x","y"],"correta":1}]}}}`)
	d, err := ParseLessonData(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Trilha.Topicos) != 1 || d.Trilha.Quiz.Questoes[0].Correta != 1 {
		t.Fatal("parse mismatch")
	}
}

func TestParseRejectsBadCorreta(t *testing.T) {
	raw := []byte(`{"plano":{},"atividade":"a","trilha":{"topicos":[{"titulo":"t","resumo":"r"}],"quiz":{"questoes":[{"enunciado":"e","opcoes":["x","y"],"correta":5}]}}}`)
	if _, err := ParseLessonData(raw); err == nil {
		t.Fatal("want validation error")
	}
}

func TestParseRejectsZeroTopicos(t *testing.T) {
	raw := []byte(`{"plano":{},"atividade":"a","trilha":{"topicos":[],"quiz":{"questoes":[{"enunciado":"e","opcoes":["x","y"],"correta":0}]}}}`)
	if _, err := ParseLessonData(raw); err == nil {
		t.Fatal("want validation error for zero topicos")
	}
}

func TestParseRejectsZeroQuestoes(t *testing.T) {
	raw := []byte(`{"plano":{},"atividade":"a","trilha":{"topicos":[{"titulo":"t","resumo":"r"}],"quiz":{"questoes":[]}}}`)
	if _, err := ParseLessonData(raw); err == nil {
		t.Fatal("want validation error for zero questoes")
	}
}

func TestParseRejectsInsufficientOpcoes(t *testing.T) {
	raw := []byte(`{"plano":{},"atividade":"a","trilha":{"topicos":[{"titulo":"t","resumo":"r"}],"quiz":{"questoes":[{"enunciado":"e","opcoes":["x"],"correta":0}]}}}`)
	if _, err := ParseLessonData(raw); err == nil {
		t.Fatal("want validation error for < 2 opcoes")
	}
}
