package http

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

type turmaInput struct {
	Nome  string  `json:"nome"`
	Etapa string  `json:"etapa"`
	Anos  []int32 `json:"anos"`
}

func (d Deps) loadOwnedTurma(w http.ResponseWriter, r *http.Request, userID int64) (store.Turma, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "turma não encontrada"})
		return store.Turma{}, false
	}
	turma, err := d.Store.GetTurma(r.Context(), id)
	if err != nil || turma.UserID != userID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "turma não encontrada"})
		return store.Turma{}, false
	}
	return turma, true
}

func (d Deps) createTurma(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	var in turmaInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Nome == "" {
		writeJSON(w, 400, map[string]string{"error": "nome é obrigatório"})
		return
	}
	if in.Anos == nil {
		in.Anos = []int32{}
	}
	turma, err := d.Store.CreateTurma(r.Context(), store.CreateTurmaParams{
		UserID: userID, Nome: in.Nome, Etapa: in.Etapa, Anos: in.Anos,
	})
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro ao criar turma"})
		return
	}
	writeJSON(w, 201, turma)
}

func (d Deps) listTurmas(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	turmas, err := d.Store.ListTurmasByUser(r.Context(), userID)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro ao listar turmas"})
		return
	}
	writeJSON(w, 200, turmas)
}

// studentPublic is a student without any secret fields.
type studentPublic struct {
	ID        int64   `json:"id"`
	Nome      string  `json:"nome"`
	Usuario   string  `json:"usuario"`
	Matricula *string `json:"matricula"`
}

func (d Deps) getTurma(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	turma, ok := d.loadOwnedTurma(w, r, userID)
	if !ok {
		return
	}
	students, err := d.Store.ListStudentsByTurma(r.Context(), turma.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, 500, map[string]string{"error": "erro ao carregar alunos"})
		return
	}
	pub := make([]studentPublic, 0, len(students))
	for _, s := range students {
		pub = append(pub, studentPublic{ID: s.ID, Nome: s.Nome, Usuario: s.Usuario, Matricula: s.Matricula})
	}
	writeJSON(w, 200, map[string]any{"turma": turma, "alunos": pub})
}

func (d Deps) patchTurma(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	turma, ok := d.loadOwnedTurma(w, r, userID)
	if !ok {
		return
	}
	var in turmaInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "json inválido"})
		return
	}
	if in.Nome == "" {
		in.Nome = turma.Nome
	}
	if in.Anos == nil {
		in.Anos = turma.Anos
	}
	if in.Etapa == "" {
		in.Etapa = turma.Etapa
	}
	updated, err := d.Store.UpdateTurma(r.Context(), store.UpdateTurmaParams{
		ID: turma.ID, Nome: in.Nome, Etapa: in.Etapa, Anos: in.Anos,
	})
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro ao atualizar turma"})
		return
	}
	writeJSON(w, 200, updated)
}

func (d Deps) deleteTurma(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	turma, ok := d.loadOwnedTurma(w, r, userID)
	if !ok {
		return
	}
	if err := d.Store.DeleteTurma(r.Context(), turma.ID); err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro ao remover turma"})
		return
	}
	w.WriteHeader(204)
}

type importedStudent struct {
	Nome      string  `json:"nome"`
	Matricula *string `json:"matricula"`
	Usuario   string  `json:"usuario"`
	Senha     string  `json:"senha"`
}

// createUniqueStudent generates a unique username and inserts one student,
// retrying the random suffix on a unique-violation (Postgres 23505).
func (d Deps) createUniqueStudent(r *http.Request, turmaID int64, nome string, matricula *string) (importedStudent, error) {
	slug := auth.UsernameSlug(nome)
	if slug == "" {
		slug = "aluno"
	}
	senha, err := auth.GenerateInitialPassword()
	if err != nil {
		return importedStudent{}, err
	}
	hash, err := auth.HashPassword(senha)
	if err != nil {
		return importedStudent{}, err
	}
	for attempt := 0; attempt < 8; attempt++ {
		suffix, err := auth.RandomSuffix()
		if err != nil {
			return importedStudent{}, err
		}
		usuario := slug + "." + suffix
		s, err := d.Store.CreateStudent(r.Context(), store.CreateStudentParams{
			TurmaID: turmaID, Nome: nome, Usuario: usuario, SenhaHash: hash, Matricula: matricula,
		})
		if err == nil {
			return importedStudent{Nome: s.Nome, Matricula: s.Matricula, Usuario: s.Usuario, Senha: senha}, nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			continue
		}
		return importedStudent{}, err
	}
	return importedStudent{}, errors.New("não foi possível gerar usuário único")
}

func (d Deps) importStudents(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	turma, ok := d.loadOwnedTurma(w, r, userID)
	if !ok {
		return
	}

	reader := csv.NewReader(r.Body)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "csv vazio ou inválido"})
		return
	}
	nomeIdx, matIdx := -1, -1
	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "nome":
			nomeIdx = i
		case "matricula", "matrícula":
			matIdx = i
		}
	}
	if nomeIdx == -1 {
		writeJSON(w, 400, map[string]string{"error": "coluna 'nome' obrigatória"})
		return
	}

	type parsedRow struct {
		nome      string
		matricula *string
	}
	var rows []parsedRow
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": "csv inválido"})
			return
		}
		if nomeIdx >= len(rec) || strings.TrimSpace(rec[nomeIdx]) == "" {
			writeJSON(w, 400, map[string]string{"error": "todas as linhas precisam de nome"})
			return
		}
		var mat *string
		if matIdx != -1 && matIdx < len(rec) && strings.TrimSpace(rec[matIdx]) != "" {
			v := strings.TrimSpace(rec[matIdx])
			mat = &v
		}
		rows = append(rows, parsedRow{nome: strings.TrimSpace(rec[nomeIdx]), matricula: mat})
	}

	criados := make([]importedStudent, 0, len(rows))
	for _, row := range rows {
		imp, err := d.createUniqueStudent(r, turma.ID, row.nome, row.matricula)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "erro ao criar aluno: " + row.nome})
			return
		}
		criados = append(criados, imp)
	}
	writeJSON(w, 201, map[string]any{"criados": criados})
}
