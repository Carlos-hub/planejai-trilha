package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

const studentCookie = "student_sid"

func setStudentSessionCookie(w http.ResponseWriter, sid string, exp time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name: studentCookie, Value: sid, Path: "/", HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Expires: exp,
	})
}

func (d Deps) studentLogin(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Usuario string `json:"usuario"`
		Senha   string `json:"senha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "json inválido"})
		return
	}
	s, err := d.Store.GetStudentByUsuario(r.Context(), in.Usuario)
	if err != nil || !auth.CheckPassword(s.SenhaHash, in.Senha) {
		writeJSON(w, 401, map[string]string{"error": "credenciais inválidas"})
		return
	}
	sid, _ := auth.NewSessionID()
	exp := time.Now().Add(auth.SessionTTL)
	if _, err := d.Store.CreateStudentSession(r.Context(), store.CreateStudentSessionParams{
		ID: sid, StudentID: s.ID, ExpiresAt: pgTime(exp),
	}); err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro"})
		return
	}
	setStudentSessionCookie(w, sid, exp)
	writeJSON(w, 200, map[string]any{"id": s.ID, "nome": s.Nome, "usuario": s.Usuario, "turma_id": s.TurmaID})
}

func (d Deps) studentLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(studentCookie); err == nil {
		_ = d.Store.DeleteStudentSession(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: studentCookie, Value: "", Path: "/", MaxAge: -1})
	w.WriteHeader(204)
}

func (d Deps) studentChangePassword(w http.ResponseWriter, r *http.Request) {
	sid, ok := studentIDFromContext(r)
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "não autenticado"})
		return
	}
	var in struct {
		SenhaAtual string `json:"senha_atual"`
		SenhaNova  string `json:"senha_nova"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "json inválido"})
		return
	}
	if len(in.SenhaNova) < 6 {
		writeJSON(w, 400, map[string]string{"error": "senha nova muito curta"})
		return
	}
	s, err := d.Store.GetStudent(r.Context(), sid)
	if err != nil || !auth.CheckPassword(s.SenhaHash, in.SenhaAtual) {
		writeJSON(w, 401, map[string]string{"error": "senha atual incorreta"})
		return
	}
	hash, err := auth.HashPassword(in.SenhaNova)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro"})
		return
	}
	if err := d.Store.UpdateStudentPassword(r.Context(), store.UpdateStudentPasswordParams{ID: sid, SenhaHash: hash}); err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro"})
		return
	}
	w.WriteHeader(204)
}
