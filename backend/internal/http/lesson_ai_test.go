package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/lesson"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// testBnccSkill returns an existing seeded BNCC skill id, inserting one if
// the seed table happens to be empty (e.g. a fresh test DB).
func testBnccSkill(t *testing.T, d Deps) int64 {
	t.Helper()
	ctx := context.Background()
	skills, err := d.Store.ListBnccSkills(ctx, store.ListBnccSkillsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) > 0 {
		return skills[0].ID
	}

	if err := d.Store.InsertBnccSkill(ctx, store.InsertBnccSkillParams{
		Code:       "TEST01",
		Disciplina: "Matemática",
		Etapa:      "EF",
		Anos:       []int32{5},
		Descricao:  "Habilidade de teste",
	}); err != nil {
		t.Fatal(err)
	}
	skills, err = d.Store.ListBnccSkills(ctx, store.ListBnccSkillsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) == 0 {
		t.Fatal("expected at least one bncc skill after insert")
	}
	return skills[0].ID
}

func TestLessonGenerateAndEnhance(t *testing.T) {
	d := testDeps(t)
	d.Gen = &lesson.MockGenerator{}
	ctx := context.Background()

	h, _ := auth.HashPassword("segredo123")
	u1, err := d.Store.CreateUser(ctx, store.CreateUserParams{
		Email: "lesson-ai@x.com", SenhaHash: h, Nome: "Prof IA",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM users WHERE id=$1", u1.ID) })

	skillID := testBnccSkill(t, d)

	srv := httptest.NewServer(NewRouter(d))
	defer srv.Close()

	client := loginClient(t, srv, "lesson-ai@x.com", "segredo123")

	// POST /api/lessons/generate
	genBody := `{"bncc_skill_id":` + strconv.FormatInt(skillID, 10) + `,"duracao":30}`
	genResp, err := client.Post(srv.URL+"/api/lessons/generate", "application/json", strings.NewReader(genBody))
	if err != nil {
		t.Fatal(err)
	}
	defer genResp.Body.Close()
	if genResp.StatusCode != http.StatusCreated {
		t.Fatalf("generate want 201 got %d", genResp.StatusCode)
	}
	var generated lessonResponse
	if err := json.NewDecoder(genResp.Body).Decode(&generated); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM lesson_plans WHERE id=$1", generated.ID) })

	if generated.Origem != "ia" {
		t.Fatalf("want origem ia, got %s", generated.Origem)
	}
	if generated.Status != "pronto" {
		t.Fatalf("want status pronto, got %s", generated.Status)
	}
	if len(generated.Trilha.Topicos) == 0 {
		t.Fatalf("expected topicos in generated lesson")
	}
	if len(generated.Trilha.Quiz.Questoes) == 0 {
		t.Fatalf("expected questoes in generated lesson")
	}

	// POST /api/lessons/:id/enhance
	idStr := strconv.FormatInt(generated.ID, 10)
	enhResp, err := client.Post(srv.URL+"/api/lessons/"+idStr+"/enhance", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	defer enhResp.Body.Close()
	if enhResp.StatusCode != http.StatusOK {
		t.Fatalf("enhance want 200 got %d", enhResp.StatusCode)
	}
	var enhanced lessonResponse
	if err := json.NewDecoder(enhResp.Body).Decode(&enhanced); err != nil {
		t.Fatal(err)
	}
	if enhanced.Origem != "ia_aprimorado" {
		t.Fatalf("want origem ia_aprimorado, got %s", enhanced.Origem)
	}
	if !strings.HasSuffix(enhanced.Plano.Objetivos, "(aprimorado)") {
		t.Fatalf("expected objetivos to be enhanced, got %s", enhanced.Plano.Objetivos)
	}

	// generate with no Gen configured -> 503
	noGenDeps := d
	noGenDeps.Gen = nil
	srv2 := httptest.NewServer(NewRouter(noGenDeps))
	defer srv2.Close()
	client2 := loginClient(t, srv2, "lesson-ai@x.com", "segredo123")
	noGenResp, err := client2.Post(srv2.URL+"/api/lessons/generate", "application/json", strings.NewReader(genBody))
	if err != nil {
		t.Fatal(err)
	}
	defer noGenResp.Body.Close()
	if noGenResp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("want 503 when Gen is nil, got %d", noGenResp.StatusCode)
	}
}
