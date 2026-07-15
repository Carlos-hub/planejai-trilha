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

func TestPatchTurmaPreservesEtapa(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookie := loginProfessor(t, d, "prof-patch-etapa@t.com")

	// create with non-empty etapa
	body, _ := json.Marshal(map[string]any{"nome": "8A", "etapa": "EF", "anos": []int{8}})
	req := httptest.NewRequest("POST", "/api/turmas", bytes.NewReader(body))
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", w.Code, w.Body)
	}
	var created struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(w.Body.Bytes(), &created)

	// patch omitting etapa
	patchBody, _ := json.Marshal(map[string]any{"nome": "8B"})
	req2 := httptest.NewRequest("PATCH", "/api/turmas/"+itoa(created.ID), bytes.NewReader(patchBody))
	req2.AddCookie(cookie)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("patch status = %d body=%s", w2.Code, w2.Body)
	}

	// get and verify etapa preserved
	req3 := httptest.NewRequest("GET", "/api/turmas/"+itoa(created.ID), nil)
	req3.AddCookie(cookie)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("get status = %d body=%s", w3.Code, w3.Body)
	}
	var got struct {
		Turma struct {
			Etapa string `json:"etapa"`
			Nome  string `json:"nome"`
		} `json:"turma"`
	}
	if err := json.Unmarshal(w3.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Turma.Etapa != "EF" {
		t.Fatalf("etapa = %q, want %q", got.Turma.Etapa, "EF")
	}
	if got.Turma.Nome != "8B" {
		t.Fatalf("nome = %q, want %q", got.Turma.Nome, "8B")
	}
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

// seedPublishableTrail creates a lesson (with plano+atividade+trilha content,
// via POST /api/lessons) owned by the professor identified by cookie, so its
// status becomes "pronto" and a trail row exists ready to publish. Returns
// the lesson plan id.
func seedPublishableTrail(t *testing.T, d Deps, cookie *http.Cookie) int64 {
	t.Helper()
	r := NewRouter(d)
	body := `{
		"duracao": 45,
		"plano": {"objetivos":"objetivos","metodologia":"metodologia","recursos":"recursos","avaliacao":"avaliacao"},
		"atividade": "atividade",
		"trilha": {
			"topicos": [{"titulo":"Topico 1","resumo":"resumo 1"}],
			"quiz": {"questoes": [{"enunciado":"Pergunta?","opcoes":["a","b","c"],"correta":1}]}
		}
	}`
	req := httptest.NewRequest("POST", "/api/lessons", strings.NewReader(body))
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("seedPublishableTrail: create status = %d body=%s", w.Code, w.Body)
	}
	var created lessonResponse
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	return created.ID
}

func TestPublishWithTurma(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookie := loginProfessor(t, d, "profPub-turma@t.com")

	// create turma owned by cookie
	tb, _ := json.Marshal(map[string]any{"nome": "9C"})
	tr := httptest.NewRequest("POST", "/api/turmas", bytes.NewReader(tb))
	tr.AddCookie(cookie)
	tw := httptest.NewRecorder()
	r.ServeHTTP(tw, tr)
	var turma struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(tw.Body.Bytes(), &turma)

	// publishing to a turma owned by another professor → 404
	other := loginProfessor(t, d, "profOther-turma@t.com")
	trail := seedPublishableTrail(t, d, other) // helper: creates lesson+trail ready to publish, returns lesson id
	pb, _ := json.Marshal(map[string]any{"turma_id": turma.ID})
	pr := httptest.NewRequest("POST", "/api/trails/"+itoa(trail)+"/publish", bytes.NewReader(pb))
	pr.AddCookie(other)
	pw := httptest.NewRecorder()
	r.ServeHTTP(pw, pr)
	if pw.Code != http.StatusNotFound {
		t.Fatalf("publish to foreign turma = %d, want 404", pw.Code)
	}

	// publishing to a turma owned by the caller → 200, and turma_id persisted
	ownTrail := seedPublishableTrail(t, d, cookie)
	ob, _ := json.Marshal(map[string]any{"turma_id": turma.ID})
	or := httptest.NewRequest("POST", "/api/trails/"+itoa(ownTrail)+"/publish", bytes.NewReader(ob))
	or.AddCookie(cookie)
	ow := httptest.NewRecorder()
	r.ServeHTTP(ow, or)
	if ow.Code != http.StatusOK {
		t.Fatalf("publish with own turma = %d, want 200, body=%s", ow.Code, ow.Body)
	}
	var published publishTrailResponse
	if err := json.Unmarshal(ow.Body.Bytes(), &published); err != nil {
		t.Fatal(err)
	}
	dbTrail, err := d.Store.GetTrailByCode(context.Background(), &published.Codigo)
	if err != nil {
		t.Fatalf("GetTrailByCode: %v", err)
	}
	if dbTrail.TurmaID == nil || *dbTrail.TurmaID != turma.ID {
		t.Fatalf("trail.turma_id = %v, want %d", dbTrail.TurmaID, turma.ID)
	}

	// unknown turma_id → 404
	unknownTrail := seedPublishableTrail(t, d, cookie)
	ub, _ := json.Marshal(map[string]any{"turma_id": int64(999999999)})
	ur := httptest.NewRequest("POST", "/api/trails/"+itoa(unknownTrail)+"/publish", bytes.NewReader(ub))
	ur.AddCookie(cookie)
	uw := httptest.NewRecorder()
	r.ServeHTTP(uw, ur)
	if uw.Code != http.StatusNotFound {
		t.Fatalf("publish with unknown turma = %d, want 404", uw.Code)
	}
}

func timePlus24h() time.Time { return time.Now().Add(24 * time.Hour) }

func TestImportStudents(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookie := loginProfessor(t, d, "profImport-turma@t.com")

	// create a turma
	body, _ := json.Marshal(map[string]any{"nome": "6B"})
	cr := httptest.NewRequest("POST", "/api/turmas", bytes.NewReader(body))
	cr.AddCookie(cookie)
	cw := httptest.NewRecorder()
	r.ServeHTTP(cw, cr)
	var created struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(cw.Body.Bytes(), &created)

	csv := "nome,matricula\nAna Clara,1001\nJoão da Silva,1002\n"
	req := httptest.NewRequest("POST", "/api/turmas/"+itoa(created.ID)+"/students/import", bytes.NewReader([]byte(csv)))
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "text/csv")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("import status = %d body=%s", w.Code, w.Body)
	}
	var resp struct {
		Criados []struct {
			Nome, Usuario, Senha string
		} `json:"criados"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Criados) != 2 {
		t.Fatalf("criados = %d, want 2", len(resp.Criados))
	}
	if resp.Criados[0].Senha == "" || resp.Criados[0].Usuario == "" {
		t.Fatalf("missing generated credentials: %+v", resp.Criados[0])
	}
	if resp.Criados[0].Usuario == resp.Criados[1].Usuario {
		t.Fatalf("usuarios must be unique")
	}

	// students actually persisted
	gr := httptest.NewRequest("GET", "/api/turmas/"+itoa(created.ID), nil)
	gr.AddCookie(cookie)
	gw := httptest.NewRecorder()
	r.ServeHTTP(gw, gr)
	if !bytes.Contains(gw.Body.Bytes(), []byte("Ana Clara")) {
		t.Fatalf("student not persisted: %s", gw.Body)
	}
	// plaintext password must never be in a GET
	if bytes.Contains(gw.Body.Bytes(), []byte(resp.Criados[0].Senha)) {
		t.Fatalf("plaintext password leaked in GET")
	}
}
