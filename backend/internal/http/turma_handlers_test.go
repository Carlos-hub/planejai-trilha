package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// loginProfessor creates a professor and returns an authenticated session cookie.
func loginProfessor(t *testing.T, d Deps, email string) *http.Cookie {
	t.Helper()
	hash, _ := auth.HashPassword("senha123")
	u, err := d.Store.CreateUser(context.Background(), store.CreateUserParams{Email: email, SenhaHash: hash, Nome: "Prof"})
	if err != nil {
		t.Fatal(err)
	}
	sid, _ := auth.NewSessionID()
	_, err = d.Store.CreateSession(context.Background(), store.CreateSessionParams{ID: sid, UserID: u.ID, ExpiresAt: pgTime(timePlus24h())})
	if err != nil {
		t.Fatal(err)
	}
	return &http.Cookie{Name: "sid", Value: sid}
}

func TestTurmaCRUDAndOwnership(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookieA := loginProfessor(t, d, "profA-turma@t.com")
	cookieB := loginProfessor(t, d, "profB-turma@t.com")

	// create
	body, _ := json.Marshal(map[string]any{"nome": "6A", "etapa": "EF", "anos": []int{6}})
	req := httptest.NewRequest("POST", "/api/turmas", bytes.NewReader(body))
	req.AddCookie(cookieA)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", w.Code, w.Body)
	}
	var created struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(w.Body.Bytes(), &created)

	// professor B cannot GET professor A's turma → 404
	req2 := httptest.NewRequest("GET", "/api/turmas/"+itoa(created.ID), nil)
	req2.AddCookie(cookieB)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("cross-owner GET = %d, want 404", w2.Code)
	}
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

func timePlus24h() time.Time { return time.Now().Add(24 * time.Hour) }
