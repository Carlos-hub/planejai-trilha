// Command bnccseed regenerates seed/bncc.json from the public BNCC API
// (https://api.bncc.dev). It walks the full EF+EM habilidades catalog and maps
// each record onto the shape consumed by internal/seed.BNCC, so the app keeps
// seeding from a committed file at boot (offline-safe) while the catalog itself
// stays sourced from the official dataset.
//
// Usage: go run ./cmd/bnccseed [-out seed/bncc.json]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)

const apiBase = "https://api.bncc.dev/v1"

// skill mirrors the seed JSON shape read by internal/seed.BNCC.
type skill struct {
	Code       string `json:"code"`
	Etapa      string `json:"etapa"`
	Disciplina string `json:"disciplina"`
	Anos       []int  `json:"anos"`
	Assunto    string `json:"assunto"`
	Descricao  string `json:"descricao"`
}

// apiItem is the subset of a /v1/habilidades item we consume. EF and EM records
// differ: EF carries componente/anos/organizacao/objetosConhecimento, while EM
// carries area/competenciasEspecificas and omits anos/organizacao.
type apiItem struct {
	Codigo     string `json:"codigo"`
	Etapa      string `json:"etapa"`
	Texto      string `json:"texto"`
	Anos       []int  `json:"anos"`
	Componente *named `json:"componente"`
	Area       *named `json:"area"`
	Organizacao *struct {
		Tipo  string         `json:"tipo"`
		Nomes map[string]any `json:"nomes"`
	} `json:"organizacao"`
	ObjetosConhecimento []named `json:"objetosConhecimento"`
	CompetenciasEspecificas []struct {
		Numero int `json:"numero"`
	} `json:"competenciasEspecificas"`
}

type named struct {
	Nome string `json:"nome"`
}

type apiPage struct {
	Total  int       `json:"total"`
	Pagina int       `json:"pagina"`
	Limite int       `json:"limite"`
	Itens  []apiItem `json:"itens"`
}

func main() {
	out := flag.String("out", "seed/bncc.json", "output path for the generated seed file")
	flag.Parse()

	items, err := fetchAll()
	if err != nil {
		log.Fatalf("fetch: %v", err)
	}

	skills := make([]skill, 0, len(items))
	for _, it := range items {
		skills = append(skills, mapItem(it))
	}
	// Stable order so the committed file has a clean, reviewable diff.
	sort.Slice(skills, func(i, j int) bool { return skills[i].Code < skills[j].Code })

	buf, err := json.MarshalIndent(skills, "", "  ")
	if err != nil {
		log.Fatalf("marshal: %v", err)
	}
	buf = append(buf, '\n')
	if err := os.WriteFile(*out, buf, 0o644); err != nil {
		log.Fatalf("write %s: %v", *out, err)
	}
	log.Printf("wrote %d skills to %s", len(skills), *out)
}

// fetchAll pages through /v1/habilidades until the whole catalog is collected.
func fetchAll() ([]apiItem, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	const limit = 50
	var all []apiItem
	for page := 1; ; page++ {
		url := fmt.Sprintf("%s/habilidades?limite=%d&pagina=%d", apiBase, limit, page)
		resp, err := client.Get(url)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("page %d: status %d", page, resp.StatusCode)
		}
		var p apiPage
		err = json.NewDecoder(resp.Body).Decode(&p)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if len(p.Itens) == 0 {
			break
		}
		all = append(all, p.Itens...)
		if p.Total > 0 && len(all) >= p.Total {
			break
		}
	}
	return all, nil
}

// mapItem projects an API record onto our seed schema.
func mapItem(it apiItem) skill {
	return skill{
		Code:       it.Codigo,
		Etapa:      it.Etapa,
		Disciplina: disciplinaOf(it),
		Anos:       anosOf(it),
		Assunto:    assuntoOf(it),
		Descricao:  it.Texto,
	}
}

// disciplinaOf prefers the curricular componente (EF), falling back to the área
// de conhecimento (EM, which has no componente).
func disciplinaOf(it apiItem) string {
	if it.Componente != nil && it.Componente.Nome != "" {
		return it.Componente.Nome
	}
	if it.Area != nil {
		return it.Area.Nome
	}
	return ""
}

// anosOf returns the school years a skill spans. EF records carry them directly;
// EM records omit them, but every EM code (EM13…) covers the three séries.
func anosOf(it apiItem) []int {
	if len(it.Anos) > 0 {
		return it.Anos
	}
	if it.Etapa == "EM" {
		return []int{1, 2, 3}
	}
	return []int{}
}

// assuntoOf derives a short topic label. EF exposes it via organizacao.nomes
// (prática de linguagem / unidade temática) or objetos de conhecimento; EM has
// neither, so we fall back to the competência específica number.
func assuntoOf(it apiItem) string {
	if it.Organizacao != nil {
		for _, k := range []string{"praticaLinguagem", "unidadeTematica", "praticaCultural", "eixo"} {
			if s, ok := it.Organizacao.Nomes[k].(string); ok && s != "" {
				return s
			}
		}
		for _, v := range it.Organizacao.Nomes {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	if len(it.ObjetosConhecimento) > 0 && it.ObjetosConhecimento[0].Nome != "" {
		return it.ObjetosConhecimento[0].Nome
	}
	if len(it.CompetenciasEspecificas) > 0 {
		return fmt.Sprintf("Competência específica %d", it.CompetenciasEspecificas[0].Numero)
	}
	return ""
}
