package http

import (
	"net/http"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// bnccSkillResponse is the shape returned by listBnccSkills.
type bnccSkillResponse struct {
	ID         int64  `json:"id"`
	Code       string `json:"code"`
	Disciplina string `json:"disciplina"`
	Ano        string `json:"ano"`
	Descricao  string `json:"descricao"`
}

func toBnccSkillResponse(s store.BnccSkill) bnccSkillResponse {
	return bnccSkillResponse{
		ID:         s.ID,
		Code:       s.Code,
		Disciplina: s.Disciplina,
		Ano:        s.Ano,
		Descricao:  s.Descricao,
	}
}

// listBnccSkills handles GET /api/bncc-skills: lists all BNCC skills
// available for lesson authoring, optionally filtered by disciplina/ano
// query params.
func (d Deps) listBnccSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := d.Store.ListBnccSkills(r.Context(), store.ListBnccSkillsParams{
		Column1: r.URL.Query().Get("disciplina"),
		Column2: r.URL.Query().Get("ano"),
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao listar habilidades BNCC"})
		return
	}

	resp := make([]bnccSkillResponse, 0, len(skills))
	for _, s := range skills {
		resp = append(resp, toBnccSkillResponse(s))
	}
	writeJSON(w, http.StatusOK, resp)
}
