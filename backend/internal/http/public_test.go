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
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

func TestPublicTrail(t *testing.T) {
	d := testDeps(t)
	ctx := context.Background()

	h, _ := auth.HashPassword("segredo123")
	owner, err := d.Store.CreateUser(ctx, store.CreateUserParams{
		Email: "public-owner@x.com", SenhaHash: h, Nome: "Prof Dono",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM users WHERE id=$1", owner.ID) })

	srv := httptest.NewServer(NewRouter(d))
	defer srv.Close()

	ownerClient := loginClient(t, srv, "public-owner@x.com", "segredo123")

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

	// Not yet published: unknown code should 404 (use a bogus code first).
	notFoundResp, err := http.Get(srv.URL + "/api/t/TR-ZZZZ")
	if err != nil {
		t.Fatal(err)
	}
	defer notFoundResp.Body.Close()
	if notFoundResp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown code want 404 got %d", notFoundResp.StatusCode)
	}

	// Publish the trail.
	publishResp, err := ownerClient.Post(srv.URL+"/api/trails/"+idStr+"/publish", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer publishResp.Body.Close()
	if publishResp.StatusCode != http.StatusOK {
		t.Fatalf("publish want 200 got %d", publishResp.StatusCode)
	}
	var pub struct {
		Codigo string `json:"codigo"`
	}
	if err := json.NewDecoder(publishResp.Body).Decode(&pub); err != nil {
		t.Fatal(err)
	}

	// Public read with NO auth cookie.
	resp, err := http.Get(srv.URL + "/api/t/" + pub.Codigo)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("public trail want 200 got %d", resp.StatusCode)
	}

	bodyBytes := make([]byte, 0)
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		bodyBytes = append(bodyBytes, buf[:n]...)
		if err != nil {
			break
		}
	}
	bodyString := string(bodyBytes)

	if strings.Contains(bodyString, "correta") {
		t.Fatalf("public trail response MUST NOT contain 'correta' answer key; body=%s", bodyString)
	}

	var parsed struct {
		TituloAula string `json:"titulo_aula"`
		Topicos    []struct {
			Ordem  int    `json:"ordem"`
			Titulo string `json:"titulo"`
			Resumo string `json:"resumo"`
		} `json:"topicos"`
		Quiz struct {
			Questoes []struct {
				ID        int64    `json:"id"`
				Enunciado string   `json:"enunciado"`
				Opcoes    []string `json:"opcoes"`
			} `json:"questoes"`
		} `json:"quiz"`
	}
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		t.Fatalf("decode public trail response: %v; body=%s", err, bodyString)
	}

	if parsed.TituloAula == "" {
		t.Fatalf("expected non-empty titulo_aula")
	}
	if len(parsed.Topicos) != 1 || parsed.Topicos[0].Titulo != "Topico 1" || parsed.Topicos[0].Resumo != "resumo 1" {
		t.Fatalf("unexpected topicos: %+v", parsed.Topicos)
	}
	if len(parsed.Quiz.Questoes) != 1 {
		t.Fatalf("expected 1 questao got %d", len(parsed.Quiz.Questoes))
	}
	q := parsed.Quiz.Questoes[0]
	if q.ID == 0 || q.Enunciado != "Pergunta?" || len(q.Opcoes) != 3 {
		t.Fatalf("unexpected questao: %+v", q)
	}

	// Unknown code -> 404.
	unknownResp, err := http.Get(srv.URL + "/api/t/TR-9999")
	if err != nil {
		t.Fatal(err)
	}
	defer unknownResp.Body.Close()
	if unknownResp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown code want 404 got %d", unknownResp.StatusCode)
	}
}
