package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/Carlos-hub/planejai/backend/internal/lesson"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// lessonRequest is the JSON DTO accepted by create and patch endpoints.
type lessonRequest struct {
	BnccSkillID *int64        `json:"bncc_skill_id"`
	Duracao     int           `json:"duracao"`
	Plano       lesson.Plano  `json:"plano"`
	Atividade   string        `json:"atividade"`
	Trilha      lesson.Trilha `json:"trilha"`
}

func (req lessonRequest) toLessonData() lesson.LessonData {
	return lesson.LessonData{
		Plano:     req.Plano,
		Atividade: req.Atividade,
		Trilha:    req.Trilha,
	}
}

// lessonSummary is the shape returned by the list endpoint.
type lessonSummary struct {
	ID        int64  `json:"id"`
	Origem    string `json:"origem"`
	Status    string `json:"status"`
	Duracao   int32  `json:"duracao"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toLessonSummary(lp store.LessonPlan) lessonSummary {
	return lessonSummary{
		ID:        lp.ID,
		Origem:    lp.Origem,
		Status:    lp.Status,
		Duracao:   lp.DuracaoMin,
		CreatedAt: lp.CreatedAt.Time.Format(timeLayout),
		UpdatedAt: lp.UpdatedAt.Time.Format(timeLayout),
	}
}

const timeLayout = "2006-01-02T15:04:05Z07:00"

// lessonResponse is the full detail shape returned by create/get/patch.
type lessonResponse struct {
	ID          int64         `json:"id"`
	BnccSkillID *int64        `json:"bncc_skill_id"`
	Duracao     int32         `json:"duracao"`
	Origem      string        `json:"origem"`
	Status      string        `json:"status"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
	Plano       lesson.Plano  `json:"plano"`
	Atividade   string        `json:"atividade"`
	Trilha      lesson.Trilha `json:"trilha"`
}

// buildLessonResponse assembles the full lesson detail, including topicos and
// questoes (with the correct answer) from the lesson's trail/quiz, if any.
func (d Deps) buildLessonResponse(r *http.Request, lp store.LessonPlan) (lessonResponse, error) {
	ctx := r.Context()
	resp := lessonResponse{
		ID:          lp.ID,
		BnccSkillID: lp.BnccSkillID,
		Duracao:     lp.DuracaoMin,
		Origem:      lp.Origem,
		Status:      lp.Status,
		CreatedAt:   lp.CreatedAt.Time.Format(timeLayout),
		UpdatedAt:   lp.UpdatedAt.Time.Format(timeLayout),
		Plano: lesson.Plano{
			Objetivos:   lp.Objetivos,
			Metodologia: lp.Metodologia,
			Recursos:    lp.Recursos,
			Avaliacao:   lp.Avaliacao,
		},
		Atividade: lp.Atividade,
	}

	trail, err := d.Store.GetTrailByLesson(ctx, lp.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return resp, nil
		}
		return lessonResponse{}, err
	}

	topics, err := d.Store.ListTopics(ctx, trail.ID)
	if err != nil {
		return lessonResponse{}, err
	}
	for _, t := range topics {
		resp.Trilha.Topicos = append(resp.Trilha.Topicos, lesson.Topico{
			Titulo: t.Titulo,
			Resumo: t.Resumo,
		})
	}

	quiz, err := d.Store.GetQuizByTrail(ctx, trail.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return resp, nil
		}
		return lessonResponse{}, err
	}

	questions, err := d.Store.ListQuestions(ctx, quiz.ID)
	if err != nil {
		return lessonResponse{}, err
	}
	for _, q := range questions {
		var opcoes []string
		if err := json.Unmarshal(q.Opcoes, &opcoes); err != nil {
			return lessonResponse{}, err
		}
		resp.Trilha.Quiz.Questoes = append(resp.Trilha.Quiz.Questoes, lesson.Questao{
			Enunciado: q.Enunciado,
			Opcoes:    opcoes,
			Correta:   int(q.Correta),
		})
	}

	return resp, nil
}

// hasContent reports whether the request carries any lesson content (topicos
// or questoes) that should be persisted via saveLessonContent.
func (req lessonRequest) hasContent() bool {
	return len(req.Trilha.Topicos) > 0 || len(req.Trilha.Quiz.Questoes) > 0
}

// createLesson handles POST /api/lessons: manual creation of a lesson plan
// scoped to the authenticated user.
func (d Deps) createLesson(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}

	var req lessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "corpo inválido"})
		return
	}

	lp, err := d.Store.CreateLessonPlan(r.Context(), store.CreateLessonPlanParams{
		UserID:      userID,
		BnccSkillID: req.BnccSkillID,
		DuracaoMin:  int32(req.Duracao),
		Origem:      "manual",
		Status:      "rascunho",
		Objetivos:   req.Plano.Objetivos,
		Metodologia: req.Plano.Metodologia,
		Recursos:    req.Plano.Recursos,
		Avaliacao:   req.Plano.Avaliacao,
		Atividade:   req.Atividade,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao criar aula"})
		return
	}

	if req.hasContent() {
		if err := d.saveLessonContent(r.Context(), lp.ID, req.toLessonData()); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao salvar conteúdo"})
			return
		}
		lp, err = d.Store.GetLessonPlan(r.Context(), lp.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar aula"})
			return
		}
	}

	resp, err := d.buildLessonResponse(r, lp)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao montar resposta"})
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// listLessons handles GET /api/lessons: lists lessons owned by the
// authenticated user.
func (d Deps) listLessons(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}

	lessons, err := d.Store.ListLessonPlansByUser(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao listar aulas"})
		return
	}

	summaries := make([]lessonSummary, 0, len(lessons))
	for _, lp := range lessons {
		summaries = append(summaries, toLessonSummary(lp))
	}
	writeJSON(w, http.StatusOK, summaries)
}

// loadOwnedLesson fetches the lesson plan by id and verifies it belongs to
// userID. Returns ok=false (and writes a 404) if not found or not owned;
// callers should return immediately when ok is false.
func (d Deps) loadOwnedLesson(w http.ResponseWriter, r *http.Request, userID int64) (store.LessonPlan, bool) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "aula não encontrada"})
		return store.LessonPlan{}, false
	}

	lp, err := d.Store.GetLessonPlan(r.Context(), id)
	if err != nil || lp.UserID != userID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "aula não encontrada"})
		return store.LessonPlan{}, false
	}
	return lp, true
}

// getLesson handles GET /api/lessons/:id.
func (d Deps) getLesson(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}

	lp, ok := d.loadOwnedLesson(w, r, userID)
	if !ok {
		return
	}

	resp, err := d.buildLessonResponse(r, lp)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao montar resposta"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// patchLesson handles PATCH /api/lessons/:id: updates plan fields and, if
// content is provided, replaces the trail's topics/quiz.
func (d Deps) patchLesson(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}

	lp, ok := d.loadOwnedLesson(w, r, userID)
	if !ok {
		return
	}

	var req lessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "corpo inválido"})
		return
	}

	bnccSkillID := lp.BnccSkillID
	if req.BnccSkillID != nil {
		bnccSkillID = req.BnccSkillID
	}
	duracao := lp.DuracaoMin
	if req.Duracao != 0 {
		duracao = int32(req.Duracao)
	}
	objetivos := lp.Objetivos
	if req.Plano.Objetivos != "" {
		objetivos = req.Plano.Objetivos
	}
	metodologia := lp.Metodologia
	if req.Plano.Metodologia != "" {
		metodologia = req.Plano.Metodologia
	}
	recursos := lp.Recursos
	if req.Plano.Recursos != "" {
		recursos = req.Plano.Recursos
	}
	avaliacao := lp.Avaliacao
	if req.Plano.Avaliacao != "" {
		avaliacao = req.Plano.Avaliacao
	}
	atividade := lp.Atividade
	if req.Atividade != "" {
		atividade = req.Atividade
	}

	updated, err := d.Store.UpdateLessonPlan(r.Context(), store.UpdateLessonPlanParams{
		ID:          lp.ID,
		BnccSkillID: bnccSkillID,
		DuracaoMin:  duracao,
		Objetivos:   objetivos,
		Metodologia: metodologia,
		Recursos:    recursos,
		Avaliacao:   avaliacao,
		Atividade:   atividade,
		Origem:      lp.Origem,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao atualizar aula"})
		return
	}

	if req.hasContent() {
		data := req.toLessonData()
		// saveLessonContent overwrites plano/atividade too; use the merged
		// values just persisted so a partial patch doesn't blank them out.
		data.Plano = lesson.Plano{
			Objetivos:   objetivos,
			Metodologia: metodologia,
			Recursos:    recursos,
			Avaliacao:   avaliacao,
		}
		data.Atividade = atividade
		if err := d.saveLessonContent(r.Context(), lp.ID, data); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao salvar conteúdo"})
			return
		}
		updated, err = d.Store.GetLessonPlan(r.Context(), lp.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar aula"})
			return
		}
	}

	resp, err := d.buildLessonResponse(r, updated)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao montar resposta"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
