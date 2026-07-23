package http

import (
	"net/http"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

type studentAulaResponse struct {
	Ordem     int32  `json:"ordem"`
	Label     string `json:"label"`
	Codigo    string `json:"codigo"`
	Concluido bool   `json:"concluido"`
	Pontos    int32  `json:"pontos"`
	Unlocked  bool   `json:"unlocked"`
}

func (d Deps) studentLessons(w http.ResponseWriter, r *http.Request) {
	studentID, ok := studentIDFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}
	ctx := r.Context()
	student, err := d.Store.GetStudent(ctx, studentID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro"})
		return
	}
	rows, err := d.Store.ListStudentTurmaLessons(ctx, store.ListStudentTurmaLessonsParams{
		TurmaID: student.TurmaID, StudentID: &studentID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar aulas"})
		return
	}
	aulas := make([]studentAulaResponse, 0, len(rows))
	prevConcluido := true // a primeira aula está sempre liberada
	for _, row := range rows {
		codigo := ""
		if row.Codigo != nil {
			codigo = *row.Codigo
		}
		aulas = append(aulas, studentAulaResponse{
			Ordem:     row.Ordem,
			Label:     row.Label,
			Codigo:    codigo,
			Concluido: row.Concluido,
			Pontos:    row.Pontos,
			Unlocked:  prevConcluido,
		})
		prevConcluido = row.Concluido
	}
	writeJSON(w, http.StatusOK, map[string]any{"aulas": aulas})
}
