package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// exportTrailPDF handles GET /api/t/:code/export.pdf: a public,
// unauthenticated PDF export of a published study trail. Like publicTrail,
// the quiz answer key must never appear in the output, so this handler uses
// ListQuestionsPublic (never ListQuestions/correta).
func (d Deps) exportTrailPDF(w http.ResponseWriter, r *http.Request) {
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

	var questoes []publicQuestao
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

	pdfBytes, err := buildTrailPDF(tituloAula, topicRows, questoes)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao gerar pdf"})
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"trilha-%s.pdf\"", code))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}

// buildTrailPDF renders a lightweight PDF for a published trail: a title,
// each topic (título + resumo), and the quiz questions with lettered
// options — WITHOUT the answer key.
func buildTrailPDF(tituloAula string, topicos []store.TrailTopic, questoes []publicQuestao) ([]byte, error) {
	m := maroto.New()

	m.AddRow(14, text.NewCol(12, tituloAula, props.Text{
		Size:  16,
		Style: fontstyle.Bold,
		Align: align.Center,
	}))

	if len(topicos) > 0 {
		m.AddRow(10, text.NewCol(12, "Tópicos", props.Text{
			Size:  13,
			Style: fontstyle.Bold,
		}))
		for _, t := range topicos {
			m.AddAutoRow(text.NewCol(12, fmt.Sprintf("%d. %s", t.Ordem, t.Titulo), props.Text{
				Size:  11,
				Style: fontstyle.Bold,
			}))
			m.AddAutoRow(text.NewCol(12, t.Resumo, props.Text{
				Size: 10,
			}))
		}
	}

	if len(questoes) > 0 {
		m.AddRow(10, text.NewCol(12, "Quiz", props.Text{
			Size:  13,
			Style: fontstyle.Bold,
		}))
		for i, q := range questoes {
			m.AddAutoRow(text.NewCol(12, fmt.Sprintf("%d. %s", i+1, q.Enunciado), props.Text{
				Size:  11,
				Style: fontstyle.Bold,
			}))
			for j, opt := range q.Opcoes {
				letra := string(rune('a' + j))
				m.AddAutoRow(text.NewCol(12, fmt.Sprintf("%s) %s", letra, opt), props.Text{
					Size: 10,
				}))
			}
		}
	}

	doc, err := m.Generate()
	if err != nil {
		return nil, err
	}
	return doc.GetBytes(), nil
}
