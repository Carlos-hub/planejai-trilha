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
