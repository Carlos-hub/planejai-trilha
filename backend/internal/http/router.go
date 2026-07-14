package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Carlos-hub/planejai/backend/internal/lesson"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// Deps holds handler dependencies; extended in later tasks.
type Deps struct {
	Store         *store.Queries
	Pool          *pgxpool.Pool
	SessionSecret string
	Gen           lesson.Generator
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/login", d.login)
		r.Post("/auth/register", d.register)
		r.Post("/auth/logout", d.logout)
		r.Get("/t/{code}", d.publicTrail)
		r.Get("/t/{code}/export.pdf", d.exportTrailPDF)
		r.Post("/t/{code}/attempt", d.startAttempt)
		r.Post("/attempts/{id}/answers", d.submitAnswers)
		r.Group(func(r chi.Router) {
			r.Use(d.RequireAuth)
			r.Get("/me", d.me)
			r.Get("/bncc-skills", d.listBnccSkills)
			r.Post("/lessons", d.createLesson)
			r.Get("/lessons", d.listLessons)
			r.Get("/lessons/{id}", d.getLesson)
			r.Patch("/lessons/{id}", d.patchLesson)
			r.Post("/lessons/generate", d.generateLesson)
			r.Post("/lessons/{id}/enhance", d.enhanceLesson)
			r.Post("/trails/{id}/publish", d.publishTrail)
			r.Get("/trails/{id}/stats", d.trailStats)
		})
	})
	return r
}

// corsMiddleware allows the frontend origin (http://localhost:3000) to make
// credentialed requests (cookies) to the API.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// pgTime converts a time.Time into a valid pgtype.Timestamptz.
func pgTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
