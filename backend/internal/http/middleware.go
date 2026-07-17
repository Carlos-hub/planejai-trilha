package http

import (
	"context"
	"net/http"
)

type ctxKey string

const userIDKey ctxKey = "userID"

func (d Deps) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("sid")
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
			return
		}
		sess, err := d.Store.GetSession(r.Context(), c.Value)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "sessão inválida"})
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, sess.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func userIDFromContext(r *http.Request) (int64, bool) {
	v, ok := r.Context().Value(userIDKey).(int64)
	return v, ok
}

const studentIDKey ctxKey = "studentID"

func (d Deps) RequireStudent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("student_sid")
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
			return
		}
		sess, err := d.Store.GetStudentSession(r.Context(), c.Value)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "sessão inválida"})
			return
		}
		ctx := context.WithValue(r.Context(), studentIDKey, sess.StudentID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func studentIDFromContext(r *http.Request) (int64, bool) {
	v, ok := r.Context().Value(studentIDKey).(int64)
	return v, ok
}
