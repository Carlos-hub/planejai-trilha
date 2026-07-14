package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

func loginClient(t *testing.T, srv *httptest.Server, email, senha string) *http.Client {
	t.Helper()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	body := `{"email":"` + email + `","senha":"` + senha + `"}`
	resp, err := client.Post(srv.URL+"/api/auth/login", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("login want 200 got %d", resp.StatusCode)
	}
	return client
}

func TestLessonCRUD(t *testing.T) {
	d := testDeps(t)
	ctx := context.Background()

	h, _ := auth.HashPassword("segredo123")
	u1, err := d.Store.CreateUser(ctx, store.CreateUserParams{
		Email: "lesson-owner@x.com", SenhaHash: h, Nome: "Prof Dono",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM users WHERE id=$1", u1.ID) })

	u2, err := d.Store.CreateUser(ctx, store.CreateUserParams{
		Email: "lesson-other@x.com", SenhaHash: h, Nome: "Prof Outro",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM users WHERE id=$1", u2.ID) })

	srv := httptest.NewServer(NewRouter(d))
	defer srv.Close()

	client1 := loginClient(t, srv, "lesson-owner@x.com", "segredo123")
	client2 := loginClient(t, srv, "lesson-other@x.com", "segredo123")

	createBody := `{
		"duracao": 45,
		"plano": {"objetivos":"objetivos iniciais","metodologia":"metodologia x","recursos":"recursos y","avaliacao":"avaliacao z"},
		"atividade": "atividade inicial",
		"trilha": {
			"topicos": [{"titulo":"Topico 1","resumo":"resumo 1"}],
			"quiz": {"questoes": [{"enunciado":"Pergunta?","opcoes":["a","b","c"],"correta":1}]}
		}
	}`

	resp, err := client1.Post(srv.URL+"/api/lessons", "application/json", strings.NewReader(createBody))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create want 201 got %d", resp.StatusCode)
	}
	var created lessonResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM lesson_plans WHERE id=$1", created.ID) })

	if created.Origem != "manual" {
		t.Fatalf("want origem manual, got %s", created.Origem)
	}
	if created.Plano.Objetivos != "objetivos iniciais" {
		t.Fatalf("want objetivos iniciais, got %s", created.Plano.Objetivos)
	}
	if len(created.Trilha.Topicos) != 1 {
		t.Fatalf("want 1 topico, got %d", len(created.Trilha.Topicos))
	}

	// GET /api/lessons should include it
	listResp, err := client1.Get(srv.URL + "/api/lessons")
	if err != nil {
		t.Fatal(err)
	}
	defer listResp.Body.Close()
	if listResp.StatusCode != 200 {
		t.Fatalf("list want 200 got %d", listResp.StatusCode)
	}
	var list []lessonSummary
	if err := json.NewDecoder(listResp.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, l := range list {
		if l.ID == created.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected list to contain lesson %d", created.ID)
	}

	// GET /api/lessons/:id
	idStr := strconv.FormatInt(created.ID, 10)
	getResp, err := client1.Get(srv.URL + "/api/lessons/" + idStr)
	if err != nil {
		t.Fatal(err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != 200 {
		t.Fatalf("get want 200 got %d", getResp.StatusCode)
	}
	var got lessonResponse
	if err := json.NewDecoder(getResp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got.Trilha.Topicos) != 1 || len(got.Trilha.Quiz.Questoes) != 1 {
		t.Fatalf("expected topicos/questoes in get response, got %+v", got.Trilha)
	}
	if got.Trilha.Quiz.Questoes[0].Correta != 1 {
		t.Fatalf("expected correta=1 in owner view, got %d", got.Trilha.Quiz.Questoes[0].Correta)
	}

	// PATCH updates objetivos
	patchBody := `{
		"duracao": 45,
		"plano": {"objetivos":"objetivos atualizados","metodologia":"metodologia x","recursos":"recursos y","avaliacao":"avaliacao z"},
		"atividade": "atividade inicial",
		"trilha": {
			"topicos": [{"titulo":"Topico 1","resumo":"resumo 1"}],
			"quiz": {"questoes": [{"enunciado":"Pergunta?","opcoes":["a","b","c"],"correta":1}]}
		}
	}`
	req, _ := http.NewRequest(http.MethodPatch, srv.URL+"/api/lessons/"+idStr, strings.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	patchResp, err := client1.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer patchResp.Body.Close()
	if patchResp.StatusCode != 200 {
		t.Fatalf("patch want 200 got %d", patchResp.StatusCode)
	}

	// GET again reflects update
	getResp2, err := client1.Get(srv.URL + "/api/lessons/" + idStr)
	if err != nil {
		t.Fatal(err)
	}
	defer getResp2.Body.Close()
	var got2 lessonResponse
	if err := json.NewDecoder(getResp2.Body).Decode(&got2); err != nil {
		t.Fatal(err)
	}
	if got2.Plano.Objetivos != "objetivos atualizados" {
		t.Fatalf("want updated objetivos, got %s", got2.Plano.Objetivos)
	}

	// second user gets 404
	otherResp, err := client2.Get(srv.URL + "/api/lessons/" + idStr)
	if err != nil {
		t.Fatal(err)
	}
	defer otherResp.Body.Close()
	if otherResp.StatusCode != http.StatusNotFound {
		t.Fatalf("want 404 for non-owner, got %d", otherResp.StatusCode)
	}

	// second user PATCH also 404
	req2, _ := http.NewRequest(http.MethodPatch, srv.URL+"/api/lessons/"+idStr, strings.NewReader(patchBody))
	req2.Header.Set("Content-Type", "application/json")
	otherPatchResp, err := client2.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer otherPatchResp.Body.Close()
	if otherPatchResp.StatusCode != http.StatusNotFound {
		t.Fatalf("want 404 for non-owner patch, got %d", otherPatchResp.StatusCode)
	}
}
