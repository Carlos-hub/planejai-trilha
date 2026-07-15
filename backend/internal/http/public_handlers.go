package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/Carlos-hub/planejai/backend/internal/domain"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// publicTopic is a single trail topic exposed on the public trail read.
type publicTopic struct {
	Ordem  int32  `json:"ordem"`
	Titulo string `json:"titulo"`
	Resumo string `json:"resumo"`
}

// publicQuestao is a single quiz question exposed on the public trail read.
// It intentionally has NO field for the answer key ("correta") — this must
// never be sent to students.
type publicQuestao struct {
	ID        int64    `json:"id"`
	Enunciado string   `json:"enunciado"`
	Opcoes    []string `json:"opcoes"`
}

type publicQuiz struct {
	Questoes []publicQuestao `json:"questoes"`
}

// publicTrailResponse is the shape returned by GET /api/t/:code. No field
// here may carry the quiz answer key.
type publicTrailResponse struct {
	TituloAula string        `json:"titulo_aula"`
	Topicos    []publicTopic `json:"topicos"`
	Quiz       publicQuiz    `json:"quiz"`
}

// publicTrail handles GET /api/t/:code: a public, unauthenticated read of a
// published study trail. The quiz answer key must never appear in the
// response, so this handler uses ListQuestionsPublic (never ListQuestions).
func (d Deps) publicTrail(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	ctx := r.Context()

	trail, err := d.Store.GetTrailByCode(ctx, &code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "trilha não encontrada"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar trilha"})
		return
	}
	if !trail.PublicadaEm.Valid {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "trilha não encontrada"})
		return
	}
	if _, status := d.resolveStudentForTrail(r, trail); status != 0 {
		writeJSON(w, status, map[string]string{"error": "acesso restrito à turma"})
		return
	}

	tituloAula := "Trilha de Estudos"
	lp, err := d.Store.GetLessonPlan(ctx, trail.LessonPlanID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar aula"})
		return
	}
	if err == nil && lp.BnccSkillID != nil {
		skill, err := d.Store.GetBnccSkill(ctx, *lp.BnccSkillID)
		if err == nil {
			tituloAula = skill.Disciplina + " · " + skill.AnoLabel()
		}
	}

	topicRows, err := d.Store.ListTopics(ctx, trail.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar tópicos"})
		return
	}
	topicos := make([]publicTopic, 0, len(topicRows))
	for _, t := range topicRows {
		topicos = append(topicos, publicTopic{Ordem: t.Ordem, Titulo: t.Titulo, Resumo: t.Resumo})
	}

	questoes := []publicQuestao{}
	quiz, err := d.Store.GetQuizByTrail(ctx, trail.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar quiz"})
		return
	}
	if err == nil {
		rows, err := d.Store.ListQuestionsPublic(ctx, quiz.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar questões"})
			return
		}
		for _, q := range rows {
			var opcoes []string
			if err := json.Unmarshal(q.Opcoes, &opcoes); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar questões"})
				return
			}
			questoes = append(questoes, publicQuestao{ID: q.ID, Enunciado: q.Enunciado, Opcoes: opcoes})
		}
	}

	writeJSON(w, http.StatusOK, publicTrailResponse{
		TituloAula: tituloAula,
		Topicos:    topicos,
		Quiz:       publicQuiz{Questoes: questoes},
	})
}

// startAttemptRequest is the body for POST /api/t/:code/attempt.
type startAttemptRequest struct {
	Nome string `json:"nome"`
}

// startAttemptResponse is the response for POST /api/t/:code/attempt.
type startAttemptResponse struct {
	AttemptID int64 `json:"attempt_id"`
}

// startAttempt handles POST /api/t/:code/attempt: an unauthenticated
// student starts an attempt on a published trail.
func (d Deps) startAttempt(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	ctx := r.Context()

	var req startAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "corpo inválido"})
		return
	}

	trail, err := d.Store.GetTrailByCode(ctx, &code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "trilha não encontrada"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar trilha"})
		return
	}
	if !trail.PublicadaEm.Valid {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "trilha não encontrada"})
		return
	}

	student, status := d.resolveStudentForTrail(r, trail)
	if status != 0 {
		writeJSON(w, status, map[string]string{"error": "acesso restrito à turma"})
		return
	}
	if student == nil && strings.TrimSpace(req.Nome) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "nome é obrigatório"})
		return
	}
	var studentID *int64
	nome := req.Nome
	if student != nil {
		studentID = &student.ID
		nome = student.Nome
	}

	attempt, err := d.Store.CreateAttempt(ctx, store.CreateAttemptParams{
		StudyTrailID: trail.ID,
		NomeAluno:    nome,
		StudentID:    studentID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao iniciar tentativa"})
		return
	}

	writeJSON(w, http.StatusCreated, startAttemptResponse{AttemptID: attempt.ID})
}

// submitAnswer is a single answer sent by the client. It intentionally has
// no field for correctness — the client only ever sends the option it
// chose; correctness is always computed server-side from ListQuestions.
type submitAnswer struct {
	QuizQuestionID int64 `json:"quiz_question_id"`
	Escolhida      int   `json:"escolhida"`
}

// submitAnswersRequest is the body for POST /api/attempts/:id/answers.
type submitAnswersRequest struct {
	Answers []submitAnswer `json:"answers"`
}

// submitAnswersResponse is the response for POST /api/attempts/:id/answers.
type submitAnswersResponse struct {
	Pontos  int `json:"pontos"`
	Acertos int `json:"acertos"`
	Total   int `json:"total"`
}

// submitAnswers handles POST /api/attempts/:id/answers: an unauthenticated
// student submits their chosen options for an attempt. Scoring is computed
// entirely server-side from the quiz's stored answer key (ListQuestions);
// no correctness data from the client request is ever trusted.
func (d Deps) submitAnswers(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	ctx := r.Context()

	attemptID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "tentativa não encontrada"})
		return
	}

	var req submitAnswersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "corpo inválido"})
		return
	}

	attempt, err := d.Store.GetAttempt(ctx, attemptID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "tentativa não encontrada"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar tentativa"})
		return
	}

	if attempt.ConcluidoEm.Valid {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "tentativa já concluída"})
		return
	}

	quiz, err := d.Store.GetQuizByTrail(ctx, attempt.StudyTrailID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "quiz não encontrado"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar quiz"})
		return
	}

	questions, err := d.Store.ListQuestions(ctx, quiz.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar questões"})
		return
	}

	// correct is the server-side answer key: quiz_question_id -> correct
	// option index. This is never derived from client input.
	correct := make(map[int64]int, len(questions))
	for _, q := range questions {
		correct[q.ID] = int(q.Correta)
	}

	// answers maps quiz_question_id -> chosen option index, built solely
	// from the client's selections (never trust a "correta" field even if
	// present in the request body).
	answers := make(map[int64]int, len(req.Answers))
	for _, a := range req.Answers {
		answers[a.QuizQuestionID] = a.Escolhida
	}

	pontos, acertos, total := domain.Score(answers, correct)

	for _, q := range questions {
		escolhida, answered := answers[q.ID]
		if !answered {
			escolhida = -1
		}
		correta := answered && escolhida == int(q.Correta)
		if err := d.Store.InsertAnswer(ctx, store.InsertAnswerParams{
			StudentAttemptID: attempt.ID,
			QuizQuestionID:   q.ID,
			Escolhida:        int32(escolhida),
			Correta:          correta,
		}); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao salvar respostas"})
			return
		}
	}

	if _, err := d.Store.CompleteAttempt(ctx, store.CompleteAttemptParams{
		ID:     attempt.ID,
		Pontos: int32(pontos),
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao concluir tentativa"})
		return
	}

	writeJSON(w, http.StatusOK, submitAnswersResponse{Pontos: pontos, Acertos: acertos, Total: total})
}

// resolveStudentForTrail enforces turma gating. For an ungated trail
// (TurmaID == nil) it returns (nil, 0). For a gated trail it requires a
// valid student session whose student belongs to the trail's turma:
// status 401 = not logged in / invalid session, 403 = wrong turma.
func (d Deps) resolveStudentForTrail(r *http.Request, trail store.StudyTrail) (*store.Student, int) {
	if trail.TurmaID == nil {
		return nil, 0
	}
	c, err := r.Cookie(studentCookie)
	if err != nil {
		return nil, http.StatusUnauthorized
	}
	sess, err := d.Store.GetStudentSession(r.Context(), c.Value)
	if err != nil {
		return nil, http.StatusUnauthorized
	}
	s, err := d.Store.GetStudent(r.Context(), sess.StudentID)
	if err != nil {
		return nil, http.StatusUnauthorized
	}
	if s.TurmaID != *trail.TurmaID {
		return nil, http.StatusForbidden
	}
	return &s, 0
}
