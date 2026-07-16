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

func aiTokenSetup(t *testing.T, email string) (Deps, *http.Cookie) {
	t.Helper()
	d := testDeps(t)
	// give Deps a secret box (reuse withAI's box path by sealing directly is not
	// needed here; putAIToken needs d.Secret set)
	setTestSecret(t, &d)
	h, _ := auth.HashPassword("segredo123")
	u, err := d.Store.CreateUser(context.Background(), store.CreateUserParams{Email: email, SenhaHash: h, Nome: "P"})
	if err != nil {
		t.Fatal(err)
	}
	sid, _ := auth.NewSessionID()
	if _, err := d.Store.CreateSession(context.Background(), store.CreateSessionParams{ID: sid, UserID: u.ID, ExpiresAt: pgTime(timePlus24h())}); err != nil {
		t.Fatal(err)
	}
	return d, &http.Cookie{Name: "sid", Value: sid}
}

func TestAITokenCRUD(t *testing.T) {
	d, cookie := aiTokenSetup(t, "ai-token-crud@t.com")
	r := NewRouter(d)

	// invalid provider → 400
	bad, _ := json.Marshal(map[string]string{"provider": "nope", "token": "k"})
	req := httptest.NewRequest("PUT", "/api/me/ai-token", bytes.NewReader(bad))
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad provider = %d, want 400", w.Code)
	}

	// valid PUT → 200, token never echoed
	ok, _ := json.Marshal(map[string]string{"provider": "anthropic", "token": "sk-secret-xyz"})
	req2 := httptest.NewRequest("PUT", "/api/me/ai-token", bytes.NewReader(ok))
	req2.AddCookie(cookie)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("put = %d body=%s", w2.Code, w2.Body)
	}
	if bytes.Contains(w2.Body.Bytes(), []byte("sk-secret-xyz")) {
		t.Fatal("PUT response leaked the token")
	}

	// GET → configured:true, provider, no token
	req3 := httptest.NewRequest("GET", "/api/me/ai-token", nil)
	req3.AddCookie(cookie)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK || !bytes.Contains(w3.Body.Bytes(), []byte("anthropic")) {
		t.Fatalf("get = %d body=%s", w3.Code, w3.Body)
	}
	if bytes.Contains(w3.Body.Bytes(), []byte("sk-secret-xyz")) {
		t.Fatal("GET leaked the token")
	}
	var got struct {
		Configured bool `json:"configured"`
	}
	json.Unmarshal(w3.Body.Bytes(), &got)
	if !got.Configured {
		t.Fatal("want configured true")
	}

	// DELETE → 204, then GET configured:false
	req4 := httptest.NewRequest("DELETE", "/api/me/ai-token", nil)
	req4.AddCookie(cookie)
	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, req4)
	if w4.Code != http.StatusNoContent {
		t.Fatalf("delete = %d", w4.Code)
	}
	req5 := httptest.NewRequest("GET", "/api/me/ai-token", nil)
	req5.AddCookie(cookie)
	w5 := httptest.NewRecorder()
	r.ServeHTTP(w5, req5)
	if !bytes.Contains(w5.Body.Bytes(), []byte("false")) {
		t.Fatalf("after delete get = %s", w5.Body)
	}
}
