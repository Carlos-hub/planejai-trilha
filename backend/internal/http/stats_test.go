package http

import (
	"bytes"
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

// TestTrailStats publishes a trail, creates two student attempts with
// different scores (at least one completed), and verifies the class
// dashboard stats endpoint aggregates them correctly and enforces ownership.
func TestTrailStats(t *testing.T) {
	d := testDeps(t)
	ctx := context.Background()

	h, _ := auth.HashPassword("segredo123")
	owner, err := d.Store.CreateUser(ctx, store.CreateUserParams{
		Email: "stats-owner@x.com", SenhaHash: h, Nome: "Prof Dono",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM users WHERE id=$1", owner.ID) })

	other, err := d.Store.CreateUser(ctx, store.CreateUserParams{
		Email: "stats-other@x.com", SenhaHash: h, Nome: "Prof Outro",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM users WHERE id=$1", other.ID) })

	srv := httptest.NewServer(NewRouter(d))
	defer srv.Close()

	ownerClient := loginClient(t, srv, "stats-owner@x.com", "segredo123")
	otherClient := loginClient(t, srv, "stats-other@x.com", "segredo123")

	// Quiz with 2 questions: correct option index 0 for Q1, 2 for Q2.
	createBody := `{
		"duracao": 45,
		"plano": {"objetivos":"objetivos","metodologia":"metodologia","recursos":"recursos","avaliacao":"avaliacao"},
		"atividade": "atividade",
		"trilha": {
			"topicos": [{"titulo":"Topico 1","resumo":"resumo 1"}],
			"quiz": {"questoes": [
				{"enunciado":"Pergunta 1?","opcoes":["certa","errada1","errada2"],"correta":0},
				{"enunciado":"Pergunta 2?","opcoes":["errada1","errada2","certa"],"correta":2}
			]}
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

	// Fetch the public trail to learn quiz_question_id values.
	pubResp, err := http.Get(srv.URL + "/api/t/" + pub.Codigo)
	if err != nil {
		t.Fatal(err)
	}
	defer pubResp.Body.Close()
	var pubTrail struct {
		Quiz struct {
			Questoes []struct {
				ID        int64    `json:"id"`
				Enunciado string   `json:"enunciado"`
				Opcoes    []string `json:"opcoes"`
			} `json:"questoes"`
		} `json:"quiz"`
	}
	if err := json.NewDecoder(pubResp.Body).Decode(&pubTrail); err != nil {
		t.Fatal(err)
	}
	if len(pubTrail.Quiz.Questoes) != 2 {
		t.Fatalf("expected 2 questoes got %d", len(pubTrail.Quiz.Questoes))
	}
	q1 := pubTrail.Quiz.Questoes[0]
	q2 := pubTrail.Quiz.Questoes[1]

	startAttempt := func(nome string) int64 {
		startBody := `{"nome":"` + nome + `"}`
		resp, err := http.Post(srv.URL+"/api/t/"+pub.Codigo+"/attempt", "application/json", strings.NewReader(startBody))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("start attempt want 201 got %d", resp.StatusCode)
		}
		var started struct {
			AttemptID int64 `json:"attempt_id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&started); err != nil {
			t.Fatal(err)
		}
		return started.AttemptID
	}

	submitAnswers := func(attemptID int64, escolhidaQ1, escolhidaQ2 int) {
		payload := map[string]any{
			"answers": []map[string]any{
				{"quiz_question_id": q1.ID, "escolhida": escolhidaQ1},
				{"quiz_question_id": q2.ID, "escolhida": escolhidaQ2},
			},
		}
		b, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		attemptIDStr := strconv.FormatInt(attemptID, 10)
		resp, err := http.Post(srv.URL+"/api/attempts/"+attemptIDStr+"/answers", "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("submit answers want 200 got %d", resp.StatusCode)
		}
	}

	// Attempt 1: both correct -> 20 points, completed.
	attempt1ID := startAttempt("Aluno Um")
	t.Cleanup(func() {
		d.Pool.Exec(ctx, "DELETE FROM attempt_answers WHERE student_attempt_id=$1", attempt1ID)
		d.Pool.Exec(ctx, "DELETE FROM student_attempts WHERE id=$1", attempt1ID)
	})
	submitAnswers(attempt1ID, 0, 2)

	// Attempt 2: one correct -> 10 points, completed.
	attempt2ID := startAttempt("Aluno Dois")
	t.Cleanup(func() {
		d.Pool.Exec(ctx, "DELETE FROM attempt_answers WHERE student_attempt_id=$1", attempt2ID)
		d.Pool.Exec(ctx, "DELETE FROM student_attempts WHERE id=$1", attempt2ID)
	})
	submitAnswers(attempt2ID, 0, 0)

	// Non-owner gets 404.
	otherResp, err := otherClient.Get(srv.URL + "/api/trails/" + idStr + "/stats")
	if err != nil {
		t.Fatal(err)
	}
	defer otherResp.Body.Close()
	if otherResp.StatusCode != http.StatusNotFound {
		t.Fatalf("non-owner stats want 404 got %d", otherResp.StatusCode)
	}

	// Owner fetches stats.
	statsResp, err := ownerClient.Get(srv.URL + "/api/trails/" + idStr + "/stats")
	if err != nil {
		t.Fatal(err)
	}
	defer statsResp.Body.Close()
	if statsResp.StatusCode != http.StatusOK {
		t.Fatalf("owner stats want 200 got %d", statsResp.StatusCode)
	}

	var stats struct {
		TotalAlunos int     `json:"total_alunos"`
		Concluidos  int     `json:"concluidos"`
		MediaPontos float64 `json:"media_pontos"`
		Tentativas  []struct {
			NomeAluno   string  `json:"nome_aluno"`
			Pontos      int32   `json:"pontos"`
			ConcluidoEm *string `json:"concluido_em"`
		} `json:"tentativas"`
	}
	if err := json.NewDecoder(statsResp.Body).Decode(&stats); err != nil {
		t.Fatal(err)
	}

	if stats.TotalAlunos != 2 {
		t.Fatalf("total_alunos want 2 got %d", stats.TotalAlunos)
	}
	if stats.Concluidos != 2 {
		t.Fatalf("concluidos want 2 got %d", stats.Concluidos)
	}
	if stats.MediaPontos != 15 {
		t.Fatalf("media_pontos want 15 got %v", stats.MediaPontos)
	}
	if len(stats.Tentativas) != 2 {
		t.Fatalf("tentativas length want 2 got %d", len(stats.Tentativas))
	}
	for _, tv := range stats.Tentativas {
		if tv.ConcluidoEm == nil {
			t.Fatalf("expected concluido_em to be set for tentativa %q", tv.NomeAluno)
		}
	}

	// Unknown lesson id -> 404.
	notFoundResp, err := ownerClient.Get(srv.URL + "/api/trails/9999999/stats")
	if err != nil {
		t.Fatal(err)
	}
	defer notFoundResp.Body.Close()
	if notFoundResp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown lesson stats want 404 got %d", notFoundResp.StatusCode)
	}
}
