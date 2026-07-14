package http

import (
	"context"
	"testing"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/lesson"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

func TestSaveLessonContent(t *testing.T) {
	d := testDeps(t)
	ctx := context.Background()

	h, _ := auth.HashPassword("segredo123")
	u, err := d.Store.CreateUser(ctx, store.CreateUserParams{
		Email: "persist-test@x.com", SenhaHash: h, Nome: "Prof Persist",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM users WHERE id=$1", u.ID) })

	lp, err := d.Store.CreateLessonPlan(ctx, store.CreateLessonPlanParams{
		UserID:      u.ID,
		BnccSkillID: nil,
		DuracaoMin:  50,
		Origem:      "manual",
		Status:      "rascunho",
		Objetivos:   "objetivos iniciais",
		Metodologia: "metodologia inicial",
		Recursos:    "recursos iniciais",
		Avaliacao:   "avaliacao inicial",
		Atividade:   "atividade inicial",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM lesson_plans WHERE id=$1", lp.ID) })

	skill := store.BnccSkill{Descricao: "leitura e interpretação de texto"}
	data, err := (&lesson.MockGenerator{}).Generate(ctx, skill, 50)
	if err != nil {
		t.Fatal(err)
	}

	if err := d.saveLessonContent(ctx, lp.ID, data, ""); err != nil {
		t.Fatalf("saveLessonContent: %v", err)
	}

	trail, err := d.Store.GetTrailByLesson(ctx, lp.ID)
	if err != nil {
		t.Fatalf("expected trail to exist: %v", err)
	}

	topics, err := d.Store.ListTopics(ctx, trail.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(topics) != len(data.Trilha.Topicos) {
		t.Fatalf("want %d topics, got %d", len(data.Trilha.Topicos), len(topics))
	}

	quiz, err := d.Store.GetQuizByTrail(ctx, trail.ID)
	if err != nil {
		t.Fatalf("expected quiz to exist: %v", err)
	}

	questions, err := d.Store.ListQuestions(ctx, quiz.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(questions) != len(data.Trilha.Quiz.Questoes) {
		t.Fatalf("want %d questions, got %d", len(data.Trilha.Quiz.Questoes), len(questions))
	}

	updated, err := d.Store.GetLessonPlan(ctx, lp.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "pronto" {
		t.Fatalf("want status pronto, got %s", updated.Status)
	}
	if updated.Objetivos != data.Plano.Objetivos {
		t.Fatalf("want objetivos %q, got %q", data.Plano.Objetivos, updated.Objetivos)
	}
	if updated.Origem != "manual" {
		t.Fatalf("origem should remain unchanged, got %s", updated.Origem)
	}

	// calling again should replace (not duplicate) topics/questions
	if err := d.saveLessonContent(ctx, lp.ID, data, ""); err != nil {
		t.Fatalf("second saveLessonContent: %v", err)
	}

	// calling with a non-empty origem should override it atomically
	if err := d.saveLessonContent(ctx, lp.ID, data, "ia_aprimorado"); err != nil {
		t.Fatalf("third saveLessonContent with origem override: %v", err)
	}
	overridden, err := d.Store.GetLessonPlan(ctx, lp.ID)
	if err != nil {
		t.Fatal(err)
	}
	if overridden.Origem != "ia_aprimorado" {
		t.Fatalf("want origem ia_aprimorado after override, got %s", overridden.Origem)
	}
	topics2, err := d.Store.ListTopics(ctx, trail.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(topics2) != len(data.Trilha.Topicos) {
		t.Fatalf("want %d topics after second save, got %d", len(data.Trilha.Topicos), len(topics2))
	}
}
