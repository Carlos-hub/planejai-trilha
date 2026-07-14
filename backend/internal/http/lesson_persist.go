package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Carlos-hub/planejai/backend/internal/lesson"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// saveLessonContent persists the generated LessonData (plano, atividade, trilha
// with topicos and quiz) for a lesson plan, transactionally. It replaces any
// existing topics/questions for the lesson's trail and marks the lesson as
// 'pronto' on success.
func (d Deps) saveLessonContent(ctx context.Context, lessonID int64, data lesson.LessonData) error {
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := d.Store.WithTx(tx)

	trail, err := q.GetTrailByLesson(ctx, lessonID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			trail, err = q.CreateTrail(ctx, lessonID)
			if err != nil {
				return fmt.Errorf("create trail: %w", err)
			}
		} else {
			return fmt.Errorf("get trail: %w", err)
		}
	}

	if err := q.DeleteTopics(ctx, trail.ID); err != nil {
		return fmt.Errorf("delete topics: %w", err)
	}
	for i, topico := range data.Trilha.Topicos {
		if err := q.InsertTopic(ctx, store.InsertTopicParams{
			StudyTrailID: trail.ID,
			Ordem:        int32(i + 1),
			Titulo:       topico.Titulo,
			Resumo:       topico.Resumo,
		}); err != nil {
			return fmt.Errorf("insert topic %d: %w", i, err)
		}
	}

	quiz, err := q.CreateQuiz(ctx, trail.ID)
	if err != nil {
		return fmt.Errorf("create quiz: %w", err)
	}

	if err := q.DeleteQuestionsByQuiz(ctx, quiz.ID); err != nil {
		return fmt.Errorf("delete questions: %w", err)
	}
	for i, questao := range data.Trilha.Quiz.Questoes {
		opcoesJSON, err := json.Marshal(questao.Opcoes)
		if err != nil {
			return fmt.Errorf("marshal opcoes %d: %w", i, err)
		}
		if err := q.InsertQuestion(ctx, store.InsertQuestionParams{
			QuizID:    quiz.ID,
			Ordem:     int32(i + 1),
			Enunciado: questao.Enunciado,
			Opcoes:    opcoesJSON,
			Correta:   int32(questao.Correta),
		}); err != nil {
			return fmt.Errorf("insert question %d: %w", i, err)
		}
	}

	current, err := q.GetLessonPlan(ctx, lessonID)
	if err != nil {
		return fmt.Errorf("get lesson plan: %w", err)
	}

	if _, err := q.UpdateLessonPlan(ctx, store.UpdateLessonPlanParams{
		ID:          lessonID,
		BnccSkillID: current.BnccSkillID,
		DuracaoMin:  current.DuracaoMin,
		Objetivos:   data.Plano.Objetivos,
		Metodologia: data.Plano.Metodologia,
		Recursos:    data.Plano.Recursos,
		Avaliacao:   data.Plano.Avaliacao,
		Atividade:   data.Atividade,
		Origem:      current.Origem,
	}); err != nil {
		return fmt.Errorf("update lesson plan: %w", err)
	}

	if err := q.SetLessonStatus(ctx, store.SetLessonStatusParams{
		ID:     lessonID,
		Status: "pronto",
	}); err != nil {
		return fmt.Errorf("set lesson status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
