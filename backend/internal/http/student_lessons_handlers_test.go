package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// loginStudent creates a student in turmaID and returns their session cookie.
func loginStudent(t *testing.T, d Deps, turmaID int64, usuario string) (*http.Cookie, int64) {
	t.Helper()
	hash, _ := auth.HashPassword("senha123")
	s, err := d.Store.CreateStudent(context.Background(), store.CreateStudentParams{
		TurmaID: turmaID, Nome: "Aluno", Usuario: usuario, SenhaHash: hash,
	})
	if err != nil {
		t.Fatal(err)
	}
	sid, _ := auth.NewSessionID()
	_, err = d.Store.CreateStudentSession(context.Background(), store.CreateStudentSessionParams{
		ID: sid, StudentID: s.ID, ExpiresAt: pgTime(timePlus24h()),
	})
	if err != nil {
		t.Fatal(err)
	}
	return &http.Cookie{Name: "student_sid", Value: sid}, s.ID
}

func TestStudentLessonsGating(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	suffix := itoa(time.Now().UnixNano())
	cookie := loginProfessor(t, d, "stlessons-"+suffix+"@t.com")
	uid := userIDFromCookie(t, d, cookie)
	turmaID := seedTurma(t, d, uid, "6A-stlessons-"+suffix)
	l1 := seedReadyLesson(t, d, uid)
	l2 := seedReadyLesson(t, d, uid)
	for _, id := range []int64{l1, l2} {
		b, _ := json.Marshal(map[string]any{"lesson_plan_id": id})
		req := httptest.NewRequest("POST", "/api/turmas/"+itoa(turmaID)+"/lessons", bytes.NewReader(b))
		req.AddCookie(cookie)
		r.ServeHTTP(httptest.NewRecorder(), req)
	}

	stCookie, studentID := loginStudent(t, d, turmaID, "aluno-stlessons-"+suffix)

	// antes de concluir nada: aula 1 unlocked, aula 2 locked
	aulas := fetchStudentAulas(t, r, stCookie)
	if len(aulas) != 2 || !aulas[0].Unlocked || aulas[1].Unlocked {
		t.Fatalf("initial gating wrong: %+v", aulas)
	}

	// concluir a aula 1: attempt com concluido_em na trilha de l1
	trail1, _ := d.Store.GetTrailByLesson(context.Background(), l1)
	att, _ := d.Store.CreateAttempt(context.Background(), store.CreateAttemptParams{
		StudyTrailID: trail1.ID, NomeAluno: "Aluno", StudentID: &studentID,
	})
	if _, err := d.Store.CompleteAttempt(context.Background(), store.CompleteAttemptParams{ID: att.ID, Pontos: 3}); err != nil {
		t.Fatal(err)
	}

	// agora aula 2 libera
	aulas = fetchStudentAulas(t, r, stCookie)
	if !aulas[1].Unlocked || !aulas[0].Concluido {
		t.Fatalf("after completion gating wrong: %+v", aulas)
	}
}

type studentAula struct {
	Ordem     int    `json:"ordem"`
	Unlocked  bool   `json:"unlocked"`
	Concluido bool   `json:"concluido"`
	Codigo    string `json:"codigo"`
}

func fetchStudentAulas(t *testing.T, r http.Handler, c *http.Cookie) []studentAula {
	t.Helper()
	req := httptest.NewRequest("GET", "/api/student/lessons", nil)
	req.AddCookie(c)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("student lessons status=%d body=%s", w.Code, w.Body)
	}
	var resp struct {
		Aulas []studentAula `json:"aulas"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Aulas
}
