package lesson

import (
	"fmt"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// systemPrompt instructs the model to reply with strict JSON only, matching
// the LessonData schema, in Portuguese.
const systemPrompt = `Você é um assistente pedagógico especializado em planejamento de aulas alinhadas à BNCC.

Responda SEMPRE e SOMENTE com um JSON válido, estrito, sem markdown, sem crases, sem comentários e sem texto antes ou depois. Não use blocos de código (` + "```" + `). A resposta deve ser exclusivamente o objeto JSON.

O JSON deve seguir exatamente este formato:
{
  "plano": {
    "objetivos": "string",
    "metodologia": "string",
    "recursos": "string",
    "avaliacao": "string"
  },
  "atividade": "string",
  "trilha": {
    "topicos": [
      {"titulo": "string", "resumo": "string"}
    ],
    "quiz": {
      "questoes": [
        {"enunciado": "string", "opcoes": ["string", "string", "string", "string"], "correta": 0}
      ]
    }
  }
}

Regras:
- Todo o conteúdo textual deve estar em português do Brasil.
- "correta" é o índice (base 0) da opção correta dentro do array "opcoes".
- Cada questão deve ter exatamente 4 opções.
- Inclua pelo menos 1 tópico em "trilha.topicos" e pelo menos 1 questão em "trilha.quiz.questoes".
- Não inclua nenhum campo além dos especificados.`

// generateUserPrompt builds the user prompt for generating a new lesson from
// a BNCC skill and a duration in minutes.
func generateUserPrompt(skill store.BnccSkill, duracaoMin int) string {
	return fmt.Sprintf(`Crie um plano de aula completo para a seguinte habilidade da BNCC:

Código: %s
Disciplina: %s
Ano: %s
Descrição da habilidade: %s

Duração da aula: %d minutos

Gere o plano de aula, uma atividade prática e uma trilha de aprendizagem (com tópicos e um quiz de fixação), seguindo rigorosamente o esquema JSON definido.`, skill.Code, skill.Disciplina, skill.AnoLabel(), skill.Descricao, duracaoMin)
}

// enhanceUserPrompt builds the user prompt for enhancing an existing draft
// lesson (given as JSON) for a BNCC skill.
func enhanceUserPrompt(skill store.BnccSkill, draftJSON string) string {
	return fmt.Sprintf(`Aprimore o seguinte rascunho de plano de aula, mantendo o alinhamento com a habilidade da BNCC abaixo. Melhore a clareza, profundidade pedagógica e qualidade das questões, mas preserve a estrutura e a intenção original do rascunho.

Código: %s
Disciplina: %s
Ano: %s
Descrição da habilidade: %s

Rascunho atual (JSON):
%s

Responda com o JSON completo e aprimorado, seguindo rigorosamente o mesmo esquema.`, skill.Code, skill.Disciplina, skill.AnoLabel(), skill.Descricao, draftJSON)
}
