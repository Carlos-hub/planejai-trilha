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

// TestAttemptScoring publishes a trail with a known quiz answer key, starts
// a student attempt, submits answers with a known number correct, and
// verifies the score is computed entirely server-side.
func TestAttemptScoring(t *testing.T) {
	d := testDeps(t)
	ctx := context.Background()

	h, _ := auth.HashPassword("segredo123")
	owner, err := d.Store.CreateUser(ctx, store.CreateUserParams{
		Email: "attempt-owner@x.com", SenhaHash: h, Nome: "Prof Dono",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM users WHERE id=$1", owner.ID) })

	srv := httptest.NewServer(NewRouter(d))
	defer srv.Close()

	ownerClient := loginClient(t, srv, "attempt-owner@x.com", "segredo123")

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

	// Fetch the public trail to learn the quiz_question_id values (the
	// client never sees "correta", only ids/enunciado/opcoes).
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

	// Start an attempt (no auth cookie — public endpoint).
	startBody := `{"nome":"Aluno Teste"}`
	startResp, err := http.Post(srv.URL+"/api/t/"+pub.Codigo+"/attempt", "application/json", strings.NewReader(startBody))
	if err != nil {
		t.Fatal(err)
	}
	defer startResp.Body.Close()
	if startResp.StatusCode != http.StatusCreated {
		t.Fatalf("start attempt want 201 got %d", startResp.StatusCode)
	}
	var startedAttempt struct {
		AttemptID int64 `json:"attempt_id"`
	}
	if err := json.NewDecoder(startResp.Body).Decode(&startedAttempt); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		d.Pool.Exec(ctx, "DELETE FROM attempt_answers WHERE student_attempt_id=$1", startedAttempt.AttemptID)
		d.Pool.Exec(ctx, "DELETE FROM student_attempts WHERE id=$1", startedAttempt.AttemptID)
	})
	attemptIDStr := strconv.FormatInt(startedAttempt.AttemptID, 10)

	// Unknown trail code -> 404 on start.
	badStartResp, err := http.Post(srv.URL+"/api/t/TR-ZZZZ/attempt", "application/json", strings.NewReader(startBody))
	if err != nil {
		t.Fatal(err)
	}
	defer badStartResp.Body.Close()
	if badStartResp.StatusCode != http.StatusNotFound {
		t.Fatalf("start attempt unknown code want 404 got %d", badStartResp.StatusCode)
	}

	// Empty nome -> 400.
	emptyNomeResp, err := http.Post(srv.URL+"/api/t/"+pub.Codigo+"/attempt", "application/json", strings.NewReader(`{"nome":""}`))
	if err != nil {
		t.Fatal(err)
	}
	defer emptyNomeResp.Body.Close()
	if emptyNomeResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("start attempt empty nome want 400 got %d", emptyNomeResp.StatusCode)
	}

	// Submit answers: Q1 correct (0), Q2 wrong (0 instead of 2). Also try
	// to smuggle a "correta" field in the answer payload — it must be
	// ignored entirely since the JSON struct has no such field mapped.
	submitPayload := map[string]any{
		"answers": []map[string]any{
			{"quiz_question_id": q1.ID, "escolhida": 0, "correta": true},
			{"quiz_question_id": q2.ID, "escolhida": 0, "correta": true},
		},
	}
	submitBytes, err := json.Marshal(submitPayload)
	if err != nil {
		t.Fatal(err)
	}
	submitResp, err := http.Post(srv.URL+"/api/attempts/"+attemptIDStr+"/answers", "application/json", bytes.NewReader(submitBytes))
	if err != nil {
		t.Fatal(err)
	}
	defer submitResp.Body.Close()
	if submitResp.StatusCode != http.StatusOK {
		t.Fatalf("submit answers want 200 got %d", submitResp.StatusCode)
	}
	var result struct {
		Pontos  int `json:"pontos"`
		Acertos int `json:"acertos"`
		Total   int `json:"total"`
	}
	if err := json.NewDecoder(submitResp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Total != 2 {
		t.Fatalf("total want 2 got %d", result.Total)
	}
	if result.Acertos != 1 {
		t.Fatalf("acertos want 1 got %d (client-supplied 'correta':true for both must be ignored)", result.Acertos)
	}
	if result.Pontos != 10*result.Acertos {
		t.Fatalf("pontos want %d got %d", 10*result.Acertos, result.Pontos)
	}

	// Verify attempt is marked concluded with the persisted score.
	attempt, err := d.Store.GetAttempt(ctx, startedAttempt.AttemptID)
	if err != nil {
		t.Fatalf("GetAttempt: %v", err)
	}
	if !attempt.ConcluidoEm.Valid {
		t.Fatalf("expected concluido_em to be set")
	}
	if attempt.Pontos != int32(result.Pontos) {
		t.Fatalf("persisted pontos want %d got %d", result.Pontos, attempt.Pontos)
	}

	// Submitting answers for an unknown attempt id -> 404.
	badSubmitResp, err := http.Post(srv.URL+"/api/attempts/9999999/answers", "application/json", strings.NewReader(`{"answers":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	defer badSubmitResp.Body.Close()
	if badSubmitResp.StatusCode != http.StatusNotFound {
		t.Fatalf("submit unknown attempt want 404 got %d", badSubmitResp.StatusCode)
	}

	// Re-submitting answers on an already-completed attempt must be
	// rejected with 409 and must not mutate pontos or insert duplicate
	// answer rows (no unique constraint on attempt_answers guards this).
	resubmitPayload := map[string]any{
		"answers": []map[string]any{
			{"quiz_question_id": q1.ID, "escolhida": 0},
			{"quiz_question_id": q2.ID, "escolhida": 2},
		},
	}
	resubmitBytes, err := json.Marshal(resubmitPayload)
	if err != nil {
		t.Fatal(err)
	}
	resubmitResp, err := http.Post(srv.URL+"/api/attempts/"+attemptIDStr+"/answers", "application/json", bytes.NewReader(resubmitBytes))
	if err != nil {
		t.Fatal(err)
	}
	defer resubmitResp.Body.Close()
	if resubmitResp.StatusCode != http.StatusConflict {
		t.Fatalf("resubmit completed attempt want 409 got %d", resubmitResp.StatusCode)
	}
	var resubmitErr struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resubmitResp.Body).Decode(&resubmitErr); err != nil {
		t.Fatal(err)
	}
	if resubmitErr.Error != "tentativa já concluída" {
		t.Fatalf("want error 'tentativa já concluída' got %q", resubmitErr.Error)
	}

	// pontos must be unchanged even though the resubmit payload would have
	// scored both questions correct (would have changed pontos if applied).
	attemptAfterResubmit, err := d.Store.GetAttempt(ctx, startedAttempt.AttemptID)
	if err != nil {
		t.Fatalf("GetAttempt after resubmit: %v", err)
	}
	if attemptAfterResubmit.Pontos != attempt.Pontos {
		t.Fatalf("pontos changed after rejected resubmit: want %d got %d", attempt.Pontos, attemptAfterResubmit.Pontos)
	}

	// no duplicate answer rows should have been inserted.
	var answerCount int
	if err := d.Pool.QueryRow(ctx, "SELECT count(*) FROM attempt_answers WHERE student_attempt_id=$1", startedAttempt.AttemptID).Scan(&answerCount); err != nil {
		t.Fatalf("count attempt_answers: %v", err)
	}
	if answerCount != result.Total {
		t.Fatalf("want %d answer rows (no duplicates), got %d", result.Total, answerCount)
	}
}
