package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// seedStudent creates a professor, turma, and one student with a known password.
func seedStudent(t *testing.T, d Deps, usuario, senha string) store.Student {
	t.Helper()
	ctx := context.Background()
	u, err := d.Store.CreateUser(ctx, store.CreateUserParams{Email: usuario + "-prof@t.com", SenhaHash: "x", Nome: "P"})
	if err != nil {
		t.Fatal(err)
	}
	turma, err := d.Store.CreateTurma(ctx, store.CreateTurmaParams{UserID: u.ID, Nome: "T", Etapa: "EF", Anos: []int32{6}})
	if err != nil {
		t.Fatal(err)
	}
	hash, _ := auth.HashPassword(senha)
	s, err := d.Store.CreateStudent(ctx, store.CreateStudentParams{TurmaID: turma.ID, Nome: "Aluno", Usuario: usuario, SenhaHash: hash})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestStudentLogin(t *testing.T) {
	d := testDeps(t)
	seedStudent(t, d, "aluno.login.aa", "segredo123")
	r := NewRouter(d)

	body, _ := json.Marshal(map[string]string{"usuario": "aluno.login.aa", "senha": "segredo123"})
	req := httptest.NewRequest("POST", "/api/student/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login status = %d body=%s", w.Code, w.Body)
	}
	if len(w.Result().Cookies()) == 0 || w.Result().Cookies()[0].Name != "student_sid" {
		t.Fatalf("expected student_sid cookie")
	}

	bad, _ := json.Marshal(map[string]string{"usuario": "aluno.login.aa", "senha": "errada"})
	req2 := httptest.NewRequest("POST", "/api/student/login", bytes.NewReader(bad))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("bad login status = %d", w2.Code)
	}
}
