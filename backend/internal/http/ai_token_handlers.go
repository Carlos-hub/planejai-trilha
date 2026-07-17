package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

var allowedProviders = map[string]bool{
	"anthropic": true, "openai": true, "googleai": true, "deepseek": true, "llama": true,
}

func (d Deps) putAIToken(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	var in struct {
		Provider string `json:"provider"`
		Token    string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "json inválido"})
		return
	}
	if !allowedProviders[in.Provider] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "provider inválido"})
		return
	}
	if in.Token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token é obrigatório"})
		return
	}
	if d.Secret == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "criptografia de token indisponível"})
		return
	}
	ct, nonce, err := d.Secret.Seal([]byte(in.Token))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao cifrar token"})
		return
	}
	if _, err := d.Store.UpsertAIToken(r.Context(), store.UpsertAITokenParams{
		UserID: userID, Provider: in.Provider, TokenCiphertext: ct, TokenNonce: nonce,
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao salvar token"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"provider": in.Provider, "configured": true})
}

func (d Deps) getAIToken(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	tok, err := d.Store.GetAIToken(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusOK, map[string]any{"configured": false, "provider": nil})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"configured": true, "provider": tok.Provider})
}

func (d Deps) deleteAIToken(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	if err := d.Store.DeleteAIToken(r.Context(), userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
