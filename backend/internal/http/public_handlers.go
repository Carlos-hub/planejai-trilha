package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
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

	tituloAula := "Trilha de Estudos"
	lp, err := d.Store.GetLessonPlan(ctx, trail.LessonPlanID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar aula"})
		return
	}
	if err == nil && lp.BnccSkillID != nil {
		skill, err := d.Store.GetBnccSkill(ctx, *lp.BnccSkillID)
		if err == nil {
			tituloAula = skill.Disciplina + " · " + skill.Ano
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
