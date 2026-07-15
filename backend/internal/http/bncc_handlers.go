package http

import (
	"net/http"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// bnccSkillResponse is the shape returned by listBnccSkills.
type bnccSkillResponse struct {
	ID         int64   `json:"id"`
	Code       string  `json:"code"`
	Etapa      string  `json:"etapa"`
	Disciplina string  `json:"disciplina"`
	Anos       []int32 `json:"anos"`
	AnoLabel   string  `json:"ano_label"`
	Assunto    string  `json:"assunto"`
	Descricao  string  `json:"descricao"`
}

func toBnccSkillResponse(s store.BnccSkill) bnccSkillResponse {
	anos := s.Anos
	if anos == nil {
		anos = []int32{}
	}
	return bnccSkillResponse{
		ID:         s.ID,
		Code:       s.Code,
		Etapa:      s.Etapa,
		Disciplina: s.Disciplina,
		Anos:       anos,
		AnoLabel:   s.AnoLabel(),
		Assunto:    s.Assunto,
		Descricao:  s.Descricao,
	}
}

// listBnccSkills handles GET /api/bncc-skills: lists all BNCC skills
// available for lesson authoring, optionally filtered by etapa (EF/EM),
// disciplina, and a free-text query (q) over code/disciplina/assunto/descricao.
func (d Deps) listBnccSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := d.Store.ListBnccSkills(r.Context(), store.ListBnccSkillsParams{
		Etapa:      r.URL.Query().Get("etapa"),
		Disciplina: r.URL.Query().Get("disciplina"),
		Q:          r.URL.Query().Get("q"),
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
