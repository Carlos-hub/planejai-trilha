package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

var trailCodeRe = regexp.MustCompile(`^TR-[A-Z2-9]{4}$`)

func TestPublishTrail(t *testing.T) {
	d := testDeps(t)
	ctx := context.Background()

	h, _ := auth.HashPassword("segredo123")
	owner, err := d.Store.CreateUser(ctx, store.CreateUserParams{
		Email: "publish-owner@x.com", SenhaHash: h, Nome: "Prof Dono",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM users WHERE id=$1", owner.ID) })

	other, err := d.Store.CreateUser(ctx, store.CreateUserParams{
		Email: "publish-other@x.com", SenhaHash: h, Nome: "Prof Outro",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM users WHERE id=$1", other.ID) })

	srv := httptest.NewServer(NewRouter(d))
	defer srv.Close()

	ownerClient := loginClient(t, srv, "publish-owner@x.com", "segredo123")
	otherClient := loginClient(t, srv, "publish-other@x.com", "segredo123")

	// Lesson without content: status stays "rascunho", publish must 400.
	draftLP, err := d.Store.CreateLessonPlan(ctx, store.CreateLessonPlanParams{
		UserID:     owner.ID,
		DuracaoMin: 30,
		Origem:     "manual",
		Status:     "rascunho",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM lesson_plans WHERE id=$1", draftLP.ID) })

	draftIDStr := strconv.FormatInt(draftLP.ID, 10)
	notReadyResp, err := ownerClient.Post(srv.URL+"/api/trails/"+draftIDStr+"/publish", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer notReadyResp.Body.Close()
	if notReadyResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("publish draft want 400 got %d", notReadyResp.StatusCode)
	}

	// Lesson with content via createLesson: status becomes "pronto" and a
	// trail is created.
	createBody := `{
		"duracao": 45,
		"plano": {"objetivos":"objetivos","metodologia":"metodologia","recursos":"recursos","avaliacao":"avaliacao"},
		"atividade": "atividade",
		"trilha": {
			"topicos": [{"titulo":"Topico 1","resumo":"resumo 1"}],
			"quiz": {"questoes": [{"enunciado":"Pergunta?","opcoes":["a","b","c"],"correta":1}]}
		}
	}`
	createResp, err := ownerClient.Post(srv.URL+"/api/lessons", "application/json", strings.NewReader(createBody))
	if err != nil {
		t.Fatal(err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create want 201 got %d", createResp.StatusCode)
	}
	var created lessonResponse
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM lesson_plans WHERE id=$1", created.ID) })

	idStr := strconv.FormatInt(created.ID, 10)

	// Non-owner gets 404.
	otherResp, err := otherClient.Post(srv.URL+"/api/trails/"+idStr+"/publish", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer otherResp.Body.Close()
	if otherResp.StatusCode != http.StatusNotFound {
		t.Fatalf("non-owner publish want 404 got %d", otherResp.StatusCode)
	}

	// Owner publishes successfully.
	resp, err := ownerClient.Post(srv.URL+"/api/trails/"+idStr+"/publish", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("publish want 200 got %d", resp.StatusCode)
	}
	var publishResp struct {
		Codigo     string `json:"codigo"`
		PublicaURL string `json:"publica_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&publishResp); err != nil {
		t.Fatal(err)
	}
	if !trailCodeRe.MatchString(publishResp.Codigo) {
		t.Fatalf("codigo %q does not match TR-XXXX pattern", publishResp.Codigo)
	}
	if !strings.HasSuffix(publishResp.PublicaURL, "/t/"+publishResp.Codigo) {
		t.Fatalf("publica_url %q does not end with /t/%s", publishResp.PublicaURL, publishResp.Codigo)
	}

	trail, err := d.Store.GetTrailByCode(ctx, &publishResp.Codigo)
	if err != nil {
		t.Fatalf("GetTrailByCode(%s): %v", publishResp.Codigo, err)
	}
	if trail.LessonPlanID != created.ID {
		t.Fatalf("trail lesson_plan_id want %d got %d", created.ID, trail.LessonPlanID)
	}
	if !trail.PublicadaEm.Valid {
		t.Fatalf("expected publicada_em to be set")
	}
}
