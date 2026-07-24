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

// attachTurmaLesson handles POST /api/turmas/:id/lessons: attaches an
// already-ready lesson plan to the turma, assigning it the next ordem and
// ensuring its trail has a public code.
func (d Deps) attachTurmaLesson(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}
	turma, ok := d.loadOwnedTurma(w, r, userID)
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

// detachTurmaLesson handles DELETE /api/turmas/:id/lessons/:lessonId: removes
// the lesson from the turma and renumbers the remaining aulas so ordem stays
// contiguous, all inside a single transaction.
func (d Deps) detachTurmaLesson(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}
	turma, ok := d.loadOwnedTurma(w, r, userID)
	if !ok {
		return
	}
	lessonID, err := strconv.ParseInt(chi.URLParam(r, "lessonId"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "aula não encontrada"})
		return
	}
	ctx := r.Context()
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro"})
		return
	}
	defer tx.Rollback(ctx)
	q := d.Store.WithTx(tx)
	if err := q.DetachTurmaLesson(ctx, store.DetachTurmaLessonParams{TurmaID: turma.ID, LessonPlanID: lessonID}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao remover aula"})
		return
	}
	if err := q.RenumberTurmaLessons(ctx, turma.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao reordenar"})
		return
	}
	if err := tx.Commit(ctx); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao salvar"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// reorderTurmaLessons handles PATCH /api/turmas/:id/lessons: renumbers the
// turma's aulas to match the given ordered_ids, rejecting when the submitted
// set does not exactly match the turma's current aulas.
func (d Deps) reorderTurmaLessons(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}
	turma, ok := d.loadOwnedTurma(w, r, userID)
	if !ok {
		return
	}
	var in struct {
		OrderedIDs []int64 `json:"ordered_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "json inválido"})
		return
	}
	ctx := r.Context()
	current, err := d.Store.ListTurmaLessons(ctx, turma.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro"})
		return
	}
	// O conjunto enviado deve ser exatamente o conjunto atual (mesmos ids).
	want := map[int64]bool{}
	for _, row := range current {
		want[row.LessonPlanID] = true
	}
	if len(in.OrderedIDs) != len(current) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "lista de aulas não confere"})
		return
	}
	seen := map[int64]bool{}
	for _, id := range in.OrderedIDs {
		if !want[id] || seen[id] {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "lista de aulas não confere"})
			return
		}
		seen[id] = true
	}
	if err := d.Store.SetTurmaLessonOrder(ctx, store.SetTurmaLessonOrderParams{
		TurmaID: turma.ID, LessonPlanIds: in.OrderedIDs,
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao reordenar"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
