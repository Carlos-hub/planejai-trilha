package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

type turmaInput struct {
	Nome  string  `json:"nome"`
	Etapa string  `json:"etapa"`
	Anos  []int32 `json:"anos"`
}

func (d Deps) loadOwnedTurma(w http.ResponseWriter, r *http.Request, userID int64) (store.Turma, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "turma não encontrada"})
		return store.Turma{}, false
	}
	turma, err := d.Store.GetTurma(r.Context(), id)
	if err != nil || turma.UserID != userID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "turma não encontrada"})
		return store.Turma{}, false
	}
	return turma, true
}

func (d Deps) createTurma(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	var in turmaInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Nome == "" {
		writeJSON(w, 400, map[string]string{"error": "nome é obrigatório"})
		return
	}
	if in.Anos == nil {
		in.Anos = []int32{}
	}
	turma, err := d.Store.CreateTurma(r.Context(), store.CreateTurmaParams{
		UserID: userID, Nome: in.Nome, Etapa: in.Etapa, Anos: in.Anos,
	})
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro ao criar turma"})
		return
	}
	writeJSON(w, 201, turma)
}

func (d Deps) listTurmas(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	turmas, err := d.Store.ListTurmasByUser(r.Context(), userID)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro ao listar turmas"})
		return
	}
	writeJSON(w, 200, turmas)
}

// studentPublic is a student without any secret fields.
type studentPublic struct {
	ID        int64   `json:"id"`
	Nome      string  `json:"nome"`
	Usuario   string  `json:"usuario"`
	Matricula *string `json:"matricula"`
}

func (d Deps) getTurma(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	turma, ok := d.loadOwnedTurma(w, r, userID)
	if !ok {
		return
	}
	students, err := d.Store.ListStudentsByTurma(r.Context(), turma.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, 500, map[string]string{"error": "erro ao carregar alunos"})
		return
	}
	pub := make([]studentPublic, 0, len(students))
	for _, s := range students {
		pub = append(pub, studentPublic{ID: s.ID, Nome: s.Nome, Usuario: s.Usuario, Matricula: s.Matricula})
	}
	writeJSON(w, 200, map[string]any{"turma": turma, "alunos": pub})
}

func (d Deps) patchTurma(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	turma, ok := d.loadOwnedTurma(w, r, userID)
	if !ok {
		return
	}
	var in turmaInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "json inválido"})
		return
	}
	if in.Nome == "" {
		in.Nome = turma.Nome
	}
	if in.Anos == nil {
		in.Anos = turma.Anos
	}
	if in.Etapa == "" {
		in.Etapa = turma.Etapa
	}
	updated, err := d.Store.UpdateTurma(r.Context(), store.UpdateTurmaParams{
		ID: turma.ID, Nome: in.Nome, Etapa: in.Etapa, Anos: in.Anos,
	})
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro ao atualizar turma"})
		return
	}
	writeJSON(w, 200, updated)
}

func (d Deps) deleteTurma(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	turma, ok := d.loadOwnedTurma(w, r, userID)
	if !ok {
		return
	}
	if err := d.Store.DeleteTurma(r.Context(), turma.ID); err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro ao remover turma"})
		return
	}
	w.WriteHeader(204)
}
