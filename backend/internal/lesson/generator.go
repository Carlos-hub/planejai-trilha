package lesson

import (
	"context"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// Generator interface for creating and enhancing lesson data.
type Generator interface {
	// Generate creates a new LessonData from a BNCC skill and duration.
	Generate(ctx context.Context, skill store.BnccSkill, duracaoMin int) (LessonData, error)

	// Enhance improves an existing LessonData draft based on a BNCC skill.
	Enhance(ctx context.Context, draft LessonData, skill store.BnccSkill) (LessonData, error)
}
