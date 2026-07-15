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

func TestExportTrailPDF(t *testing.T) {
	d := testDeps(t)
	ctx := context.Background()

	h, _ := auth.HashPassword("segredo123")
	owner, err := d.Store.CreateUser(ctx, store.CreateUserParams{
		Email: "export-owner@x.com", SenhaHash: h, Nome: "Prof Dono",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM users WHERE id=$1", owner.ID) })

	srv := httptest.NewServer(NewRouter(d))
	defer srv.Close()

	ownerClient := loginClient(t, srv, "export-owner@x.com", "segredo123")

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

	// Unpublished trail export -> 404 (unknown code path exercised below too).
	unpublishedResp, err := http.Get(srv.URL + "/api/t/TR-NOPE/export.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer unpublishedResp.Body.Close()
	if unpublishedResp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown code export want 404 got %d", unpublishedResp.StatusCode)
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

	// Export PDF with NO auth cookie (public endpoint).
	resp, err := http.Get(srv.URL + "/api/t/" + pub.Codigo + "/export.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("export want 200 got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/pdf") {
		t.Fatalf("want Content-Type application/pdf, got %q", ct)
	}
	if cd := resp.Header.Get("Content-Disposition"); !strings.Contains(cd, "trilha-"+pub.Codigo+".pdf") {
		t.Fatalf("want Content-Disposition to reference trilha-%s.pdf, got %q", pub.Codigo, cd)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		t.Fatal(err)
	}
	body := buf.Bytes()
	if len(body) < 4 || string(body[:4]) != "%PDF" {
		t.Fatalf("expected body to start with %%PDF, got %q...", body[:min(20, len(body))])
	}

	// Unknown code -> 404.
	unknownResp, err := http.Get(srv.URL + "/api/t/TR-9999/export.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer unknownResp.Body.Close()
	if unknownResp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown code want 404 got %d", unknownResp.StatusCode)
	}
}

// TestExportPDFGated verifies the C1 fix: PDF export of a trail bound to a
// turma must respect the same gate as publicTrail — anonymous requests get
// 401, and only a student enrolled in the trail's turma can download it.
func TestExportPDFGated(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookie := loginProfessor(t, d, "profExportGate-turma@t.com")

	// create a turma owned by cookie
	tb, _ := json.Marshal(map[string]any{"nome": "7B"})
	tr := httptest.NewRequest("POST", "/api/turmas", bytes.NewReader(tb))
	tr.AddCookie(cookie)
	tw := httptest.NewRecorder()
	r.ServeHTTP(tw, tr)
	var turma struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(tw.Body.Bytes(), &turma)

	// publish a trail bound to this turma
	code := publishTrailToTurma(t, d, r, cookie, turma.ID)

	// anonymous export of a gated trail -> 401
	ar := httptest.NewRequest("GET", "/api/t/"+code+"/export.pdf", nil)
	aw := httptest.NewRecorder()
	r.ServeHTTP(aw, ar)
	if aw.Code != http.StatusUnauthorized {
		t.Fatalf("anon gated export = %d, want 401", aw.Code)
	}

	// enrolled student can export -> 200, application/pdf
	hash, _ := auth.HashPassword("aluno123")
	s, _ := d.Store.CreateStudent(context.Background(), store.CreateStudentParams{TurmaID: turma.ID, Nome: "Bia", Usuario: "bia.exportgate.cc", SenhaHash: hash})
	ssid, _ := auth.NewSessionID()
	d.Store.CreateStudentSession(context.Background(), store.CreateStudentSessionParams{ID: ssid, StudentID: s.ID, ExpiresAt: pgTime(timePlus24h())})
	sr := httptest.NewRequest("GET", "/api/t/"+code+"/export.pdf", nil)
	sr.AddCookie(&http.Cookie{Name: "student_sid", Value: ssid})
	sw := httptest.NewRecorder()
	r.ServeHTTP(sw, sr)
	if sw.Code != http.StatusOK {
		t.Fatalf("enrolled gated export = %d, want 200, body=%s", sw.Code, sw.Body)
	}
	if ct := sw.Header().Get("Content-Type"); !strings.Contains(ct, "application/pdf") {
		t.Fatalf("want Content-Type application/pdf, got %q", ct)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
