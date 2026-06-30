# PlanejAI + Trilha — Design

**Data:** 2026-06-29
**Contexto:** Hackathon 6FSDT (FIAP/Postech) — tema "Auxílio aos professores do ensino público". Entrega: vídeo pitch (8min), vídeo MVP (8min), relatório. Critérios: MVP 30%, Problema/Impacto 20%, Inovação 20%, Apresentação 20%, Documentação 10%.

## Resumo Executivo

Planejador de aulas alinhado à BNCC (lado professor) que gera automaticamente trilhas de estudo focadas para os alunos a partir do mesmo plano. Diferencial: o professor planeja uma vez, e o aluno recebe conteúdo derivado daquele planejamento, sem trabalho duplicado. Engine única de geração via Anthropic API, exposta em duas interfaces.

## Problema

Professores do ensino público gastam horas planejando aulas e ainda precisam produzir material de estudo separado para alunos. Ferramentas existentes geram planos de aula com IA, mas não vinculam o plano ao material do aluno — o trabalho é duplicado.

## Solução

Uma engine que recebe `{disciplina, ano/série, habilidade BNCC, duração}` e devolve plano de aula + atividade + trilha do aluno (tópicos + quiz autocorrigido), em uma única chamada. O professor publica a trilha; o aluno acessa por link/código, sem login.

## Stack

- **Laravel** monólito modular + **PostgreSQL**.
- **Filament** — painel do professor (auth própria, CRUD, dashboard).
- **Blade** — páginas públicas do aluno.
- **Anthropic PHP SDK** (`anthropic-ai/sdk`), modelo `claude-opus-4-8`, structured outputs.
- BNCC carregada como seed (JSON estático), sem fonte externa em runtime.

## Arquitetura

```
app/
├─ Services/AiService.php        ← encapsula Anthropic SDK (único ponto de IA)
├─ Domain/Planning/              ← lado professor (Filament resources)
├─ Domain/Learning/              ← lado aluno (Blade público)
└─ Filament/                     ← painel professor (auth própria)
```

Fluxo core:

```
Professor (Filament)
  input {disciplina, ano, habilidade BNCC, duração}
    → AiService::generateLesson()  [1 chamada Anthropic, structured output]
    → grava LessonPlan + Atividade + Trilha (rascunho)
  botão "Publicar trilha" → gera código curto (ex: TR-7K2P)
       ↓
Aluno (Blade público, /t/TR-7K2P)
  digita nome → vê trilha (tópicos + resumo) → quiz autocorrigido
  gamificação: pontos + barra de progresso
       ↓
Professor: dashboard turma (quem completou, acertos)
Export: PDF leve + link wa.me (WhatsApp)
```

## Componentes

| Unidade | Responsabilidade | Depende de |
|---|---|---|
| `AiService` | `generateLesson(input): LessonData` — chama `claude-opus-4-8`, structured output JSON (plano+atividade+trilha+quiz) | Anthropic SDK |
| `BnccSeeder` | Carrega habilidades BNCC de JSON estático | — |
| Planning (Filament) | CRUD plano de aula, botão publicar, dashboard de turma | AiService, models |
| Learning (Blade) | Página de trilha pública por código, quiz, gamificação, export | models |

## Modelo de dados

```
users (professor)              ← Filament auth
bncc_skills (seed)             code, disciplina, ano, descrição
lesson_plans                   user_id, bncc_skill_id, duração, objetivos,
                               metodologia, recursos, avaliação, status
study_trails                   lesson_plan_id, código único, publicada_em
trail_topics                   study_trail_id, ordem, título, resumo
quizzes                        study_trail_id
quiz_questions                 quiz_id, enunciado, opções(json), resposta_correta
student_attempts               study_trail_id, nome_aluno, pontos, concluído_em
attempt_answers                student_attempt_id, quiz_question_id, escolhida, correta
```

Sem tabela de aluno — acesso por código + nome (login simplificado, permitido pelas regras do hackathon).

## Contrato AiService

```php
$client->messages->create(
  model: 'claude-opus-4-8',
  max_tokens: 8000,
  outputConfig: ['format' => ['type' => 'json_schema', 'schema' => $schema]],
  messages: [['role' => 'user', 'content' => $prompt]],
);
```

`$schema` força o JSON de saída:

```json
{
  "plano": { "objetivos": "...", "metodologia": "...", "recursos": "...", "avaliacao": "..." },
  "atividade": "...",
  "trilha": {
    "topicos": [{ "titulo": "...", "resumo": "..." }],
    "quiz": { "questoes": [{ "enunciado": "...", "opcoes": ["..."], "correta": 0 }] }
  }
}
```

Structured outputs garante parse confiável. Assistant prefill foi removido no Opus 4.8 (retorna 400) — usar `output_config.format`, não prefill.

## Tratamento de erro e testes

- IA falha/timeout → `lesson_plans.status = 'falha'`, professor re-tenta. Resposta em ~8k tokens fica abaixo do timeout do SDK; streaming não necessário.
- Testes unit: `AiService` com client Anthropic mockado (fixture JSON).
- Teste feature: publicar trilha → acessar por código → submeter quiz → verificar pontos corretos.

## Escopo do MVP

Inclui: lado professor→aluno completo, gamificação do quiz, export PDF/WhatsApp, dashboard de turma.

Cortado (→ pitch "Próximos Passos"): multi-turma, edição manual do plano pós-geração, contas de aluno reais, analytics avançado. Export é o primeiro candidato a corte se o tempo apertar.

## Impacto esperado

Reduz horas de planejamento do professor (uma geração cobre aula + material do aluno) e dá foco ao aluno (modo um tópico por vez + quiz). Mensurável no pitch: tempo de planejamento, taxa de conclusão de trilha.
