package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// seedReadyLesson creates a lesson_plan (status pronto) with a trail for userID.
func seedReadyLesson(t *testing.T, d Deps, userID int64) int64 {
	t.Helper()
	ctx := context.Background()
	lp, err := d.Store.CreateLessonPlan(ctx, store.CreateLessonPlanParams{
		UserID: userID, DuracaoMin: 50, Origem: "ia", Status: "pronto",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Store.CreateTrail(ctx, lp.ID); err != nil {
		t.Fatal(err)
	}
	return lp.ID
}

func seedTurma(t *testing.T, d Deps, userID int64, nome string) int64 {
	turma, err := d.Store.CreateTurma(context.Background(), store.CreateTurmaParams{
		UserID: userID, Nome: nome, Etapa: "EF", Anos: []int32{6},
	})
	if err != nil {
		t.Fatal(err)
	}
	return turma.ID
}

func TestAttachTurmaLesson(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookie := loginProfessor(t, d, "attach-lesson-task3@t.com")
	uid := userIDFromCookie(t, d, cookie)
	turmaID := seedTurma(t, d, uid, "6A-attach")
	lessonID := seedReadyLesson(t, d, uid)

	body, _ := json.Marshal(map[string]any{"lesson_plan_id": lessonID})
	req := httptest.NewRequest("POST", "/api/turmas/"+itoa(turmaID)+"/lessons", bytes.NewReader(body))
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("attach status=%d body=%s", w.Code, w.Body)
	}
	var got struct {
		LessonPlanID int64 `json:"lesson_plan_id"`
		Ordem        int   `json:"ordem"`
	}
	json.Unmarshal(w.Body.Bytes(), &got)
	if got.LessonPlanID != lessonID || got.Ordem != 1 {
		t.Fatalf("got %+v, want lesson=%d ordem=1", got, lessonID)
	}

	// duplicata → 409
	req2 := httptest.NewRequest("POST", "/api/turmas/"+itoa(turmaID)+"/lessons", bytes.NewReader(body))
	req2.AddCookie(cookie)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Fatalf("duplicate status=%d, want 409", w2.Code)
	}
}

func userIDFromCookie(t *testing.T, d Deps, c *http.Cookie) int64 {
	t.Helper()
	s, err := d.Store.GetSession(context.Background(), c.Value)
	if err != nil {
		t.Fatal(err)
	}
	return s.UserID
}

func TestGetTurmaIncludesAulas(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookie := loginProfessor(t, d, "task4-include-aulas@t.com")
	uid := userIDFromCookie(t, d, cookie)
	turmaID := seedTurma(t, d, uid, "6A-getaulas")
	lessonID := seedReadyLesson(t, d, uid)

	body, _ := json.Marshal(map[string]any{"lesson_plan_id": lessonID})
	att := httptest.NewRequest("POST", "/api/turmas/"+itoa(turmaID)+"/lessons", bytes.NewReader(body))
	att.AddCookie(cookie)
	r.ServeHTTP(httptest.NewRecorder(), att)

	req := httptest.NewRequest("GET", "/api/turmas/"+itoa(turmaID), nil)
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get turma status=%d body=%s", w.Code, w.Body)
	}
	var resp struct {
		Aulas []struct {
			LessonPlanID int64  `json:"lesson_plan_id"`
			Ordem        int    `json:"ordem"`
			Status       string `json:"status"`
			Label        string `json:"label"`
		} `json:"aulas"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Aulas) != 1 || resp.Aulas[0].LessonPlanID != lessonID {
		t.Fatalf("aulas=%+v, want 1 with lesson=%d", resp.Aulas, lessonID)
	}
}
