package http

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

func TestLoginFlow(t *testing.T) {
	d := testDeps(t)
	ctx := context.Background()
	// seed a user
	h, _ := auth.HashPassword("segredo123")
	u, err := d.Store.CreateUser(ctx, store.CreateUserParams{
		Email: "prof@x.com", SenhaHash: h, Nome: "Prof",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM users WHERE id=$1", u.ID) })

	srv := httptest.NewServer(NewRouter(d))
	defer srv.Close()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	body := `{"email":"prof@x.com","senha":"segredo123"}`
	resp, err := client.Post(srv.URL+"/api/auth/login", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("login want 200 got %d", resp.StatusCode)
	}

	me, _ := client.Get(srv.URL + "/api/me")
	if me.StatusCode != 200 {
		t.Fatalf("me want 200 got %d", me.StatusCode)
	}
}

// TestRegisterDoesNotAutoLogin verifies signup creates the account and returns
// 201 without establishing a session (the user must then log in).
func TestRegisterDoesNotAutoLogin(t *testing.T) {
	d := testDeps(t)
	ctx := context.Background()
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM users WHERE email=$1", "novo@x.com") })

	srv := httptest.NewServer(NewRouter(d))
	defer srv.Close()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	body := `{"email":"novo@x.com","senha":"segredo123","nome":"Novo Prof"}`
	resp, err := client.Post(srv.URL+"/api/auth/register", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("register want 201 got %d", resp.StatusCode)
	}
	// No session cookie should have been set by register.
	if len(resp.Cookies()) != 0 {
		t.Fatalf("register set %d cookies, want 0 (no auto-login)", len(resp.Cookies()))
	}
	// The jar has no session, so a protected route is unauthorized.
	me, _ := client.Get(srv.URL + "/api/me")
	if me.StatusCode != 401 {
		t.Fatalf("me after register want 401 got %d", me.StatusCode)
	}
	// The account exists and can log in normally.
	login := `{"email":"novo@x.com","senha":"segredo123"}`
	lr, err := client.Post(srv.URL+"/api/auth/login", "application/json", strings.NewReader(login))
	if err != nil {
		t.Fatal(err)
	}
	if lr.StatusCode != 200 {
		t.Fatalf("login after register want 200 got %d", lr.StatusCode)
	}
}
