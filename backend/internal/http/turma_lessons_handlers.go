package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Carlos-hub/planejai/backend/internal/domain"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// ensureTrailCodigo returns the trail's public code, generating one (with
// retry on unique-collision) if it has not been published yet.
func (d Deps) ensureTrailCodigo(ctx context.Context, trail store.StudyTrail) (string, error) {
	if trail.Codigo != nil && *trail.Codigo != "" {
		return *trail.Codigo, nil
	}
	for attempt := 0; attempt < maxPublishAttempts; attempt++ {
		candidate := domain.NewTrailCode()
		published, err := d.Store.PublishTrail(ctx, store.PublishTrailParams{ID: trail.ID, Codigo: &candidate})
		if err == nil {
			return *published.Codigo, nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			continue
		}
		return "", err
	}
	return "", errors.New("não foi possível gerar um código único")
}

// ownedTurma loads a turma and confirms it belongs to userID. Writes 404 and
// returns ok=false otherwise.
func (d Deps) ownedTurma(w http.ResponseWriter, r *http.Request, userID int64) (store.Turma, bool) {
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

// attachTurmaLesson handles POST /api/turmas/:id/lessons: attaches an
// already-ready lesson plan to the turma, assigning it the next ordem and
// ensuring its trail has a public code.
func (d Deps) attachTurmaLesson(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}
	turma, ok := d.ownedTurma(w, r, userID)
	if !ok {
		return
	}
	var in struct {
		LessonPlanID int64 `json:"lesson_plan_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "json inválido"})
		return
	}
	ctx := r.Context()

	lp, err := d.Store.GetLessonPlan(ctx, in.LessonPlanID)
	if err != nil || lp.UserID != userID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "aula não encontrada"})
		return
	}
	if lp.Status != "pronto" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "aula ainda não está pronta"})
		return
	}

	trail, err := d.Store.GetTrailByLesson(ctx, lp.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "aula sem trilha"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar trilha"})
		return
	}
	if _, err := d.ensureTrailCodigo(ctx, trail); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao publicar trilha"})
		return
	}

	row, err := d.Store.AttachTurmaLesson(ctx, store.AttachTurmaLessonParams{
		TurmaID: turma.ID, LessonPlanID: lp.ID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "aula já está nesta turma"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao anexar aula"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"lesson_plan_id": row.LessonPlanID, "ordem": row.Ordem})
}
