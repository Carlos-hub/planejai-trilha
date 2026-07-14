package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

func (d Deps) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Senha string }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "json inválido"})
		return
	}
	u, err := d.Store.GetUserByEmail(r.Context(), in.Email)
	if err != nil || !auth.CheckPassword(u.SenhaHash, in.Senha) {
		writeJSON(w, 401, map[string]string{"error": "credenciais inválidas"})
		return
	}
	sid, _ := auth.NewSessionID()
	exp := time.Now().Add(auth.SessionTTL)
	if _, err := d.Store.CreateSession(r.Context(), store.CreateSessionParams{
		ID: sid, UserID: u.ID, ExpiresAt: pgTime(exp),
	}); err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro"})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "sid", Value: sid, Path: "/", HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Expires: exp,
	})
	writeJSON(w, 200, map[string]any{"id": u.ID, "email": u.Email, "nome": u.Nome})
}

func (d Deps) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("sid"); err == nil {
		_ = d.Store.DeleteSession(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "sid", Value: "", Path: "/", MaxAge: -1})
	w.WriteHeader(204)
}

func (d Deps) me(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFromContext(r)
	u, err := d.Store.GetUserByID(r.Context(), uid)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "não encontrado"})
		return
	}
	writeJSON(w, 200, map[string]any{"id": u.ID, "email": u.Email, "nome": u.Nome})
}
