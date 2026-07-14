# PlanejAI + Trilha — Rebuild em Go + NextJS — Design

**Data:** 2026-07-13
**Contexto:** Hackathon 6FSDT (FIAP/Postech) — tema "Auxílio aos professores do ensino público". Reconstrução do MVP (antes Laravel/PHP) em Go (backend) + NextJS (frontend). Produto e regras de negócio inalterados; muda somente a stack. Substitui o design de 2026-06-29 (`2026-06-29-planejai-trilha-design.md`), que fica como referência histórica.

## Resumo Executivo

Planejador de aulas alinhado à BNCC (lado professor) que gera trilhas de estudo focadas para os alunos a partir do mesmo plano. Diferencial: o professor planeja uma vez, e o aluno recebe conteúdo derivado, sem trabalho duplicado. Engine única de geração via LLM (agnóstica a marca), exposta por API Go e consumida por frontend NextJS.

## Problema

Professores do ensino público gastam horas planejando aulas e ainda produzem material de estudo separado para alunos. Ferramentas com IA geram planos, mas não vinculam o plano ao material do aluno — trabalho duplicado.

## Solução

Uma API que, a partir de `{disciplina, ano/série, habilidade BNCC, duração}`, devolve plano de aula + atividade + trilha do aluno (tópicos + quiz autocorrigido). O professor pode criar tudo manualmente, deixar a IA gerar tudo, ou fornecer um rascunho para a IA aprimorar. O professor publica a trilha; o aluno acessa por código curto, sem login.

## Stack

- **Backend:** Go — `chi` + `net/http` (router), `sqlc` + `pgx` (acesso a dados type-safe), `goose` (migrations), `langchaingo` (LLM agnóstico), sessões server-side (tabela em Postgres) para auth do professor.
- **Frontend:** NextJS App Router + TypeScript + Tailwind + shadcn/ui.
- **Banco:** PostgreSQL.
- **Infra:** `docker-compose` (postgres + go-api + next). Nada instalado no host. Monorepo.
- **IA agnóstica a marca:** interface `lesson.Generator` + adapter LangChainGo. Provider/modelo via env (`LLM_PROVIDER`, `LLM_MODEL`); default Anthropic `claude-opus-4-8`. Structured output via JSON schema. Domínio depende só da interface — trocar de IA é mudança de env.
- BNCC carregada como seed (JSON estático), sem fonte externa em runtime.

## Layout do monorepo

```
Hackathon/
├─ backend/
│  ├─ cmd/api/main.go            ← entrypoint, wiring, graceful shutdown
│  ├─ internal/
│  │  ├─ http/                   ← chi router, handlers, middleware (session auth)
│  │  ├─ store/                  ← sqlc gerado + queries/*.sql
│  │  ├─ lesson/                 ← interface Generator + adapter langchaingo
│  │  ├─ auth/                   ← sessões pg-backed, hash de senha (bcrypt)
│  │  └─ domain/                 ← models, geração de código de trilha, scoring do quiz
│  ├─ migrations/                ← goose *.sql
│  ├─ seed/bncc.json             ← habilidades BNCC estáticas
│  ├─ sqlc.yaml
│  ├─ go.mod
│  └─ Dockerfile
├─ frontend/
│  ├─ app/
│  │  ├─ (teacher)/              ← login, dashboard, editor de aula, stats de turma (auth)
│  │  └─ t/[code]/               ← trilha pública do aluno + quiz (sem auth)
│  ├─ components/ui/             ← shadcn
│  ├─ lib/api.ts                 ← fetch tipado para a API Go
│  ├─ package.json
│  └─ Dockerfile
├─ docs/  spec/                  ← mantidos
└─ docker-compose.yml
```

## Modelo de dados (PostgreSQL)

```
users              professor — id, email, senha_hash, nome
sessions           id, user_id, expires_at            ← auth server-side
bncc_skills (seed) code, disciplina, ano, descricao
lesson_plans       user_id, bncc_skill_id, duracao_min, origem,
                   objetivos, metodologia, recursos, avaliacao,
                   atividade, status
study_trails       lesson_plan_id, codigo (único), publicada_em
trail_topics       study_trail_id, ordem, titulo, resumo
quizzes            study_trail_id
quiz_questions     quiz_id, enunciado, opcoes(jsonb), correta(int)
student_attempts   study_trail_id, nome_aluno, pontos, concluido_em
attempt_answers    student_attempt_id, quiz_question_id, escolhida, correta(bool)
```

- `lesson_plans.origem`: `manual | ia | ia_aprimorado`.
- `lesson_plans.status`: `rascunho | pronto | falha`.
- Sem tabela de aluno — acesso por código + nome (login simplificado, permitido pelas regras do hackathon).
- Nomes de campos em português (consistência com o domínio original).

## Modos de autoria

O professor escolhe por aula:

1. **Manual** — escreve plano/atividade/trilha/quiz à mão (CRUD completo, sem IA). `origem=manual`.
2. **IA completa** — gera tudo a partir da habilidade BNCC. `origem=ia`.
3. **IA aprimora** — professor fornece rascunho (plano e/ou atividades), IA melhora/expande e devolve versão editável que o professor aceita/ajusta. `origem=ia_aprimorado`.

Qualquer saída da IA é editável inline pelo professor antes de publicar.

## API

```
Auth (professor, cookie de sessão httpOnly):
  POST /api/auth/login          {email, senha} → set-cookie
  POST /api/auth/logout
  GET  /api/me

BNCC:
  GET  /api/bncc-skills         ?disciplina=&ano=   (busca/filtro)

Aulas (professor, protegido):
  POST  /api/lessons            manual create (sem IA)               → lesson
  POST  /api/lessons/generate   {disciplina, ano, bncc_skill_id, duracao}
                                 → Generator.Generate() [1 chamada LLM, schema] → lesson
  POST  /api/lessons/:id/enhance  → Generator.Enhance(draft atual)  → lesson aprimorada
  PATCH /api/lessons/:id        edição manual de qualquer campo
  GET   /api/lessons            lista do professor
  GET   /api/lessons/:id

Trilha:
  POST /api/trails/:id/publish  → gera código curto (ex: TR-7K2P)
  GET  /api/trails/:id/stats    → dashboard de turma (quem concluiu, acertos)

Aluno (público, sem auth):
  GET  /api/t/:code                         → trilha + tópicos + quiz (sem gabarito)
  POST /api/t/:code/attempt   {nome}        → attempt id
  POST /api/attempts/:id/answers {answers}  → correção server-side, pontos, progresso
  GET  /api/t/:code/export.pdf              → PDF leve (gerado no Go)
```

- Gabarito (`correta`) nunca vai para o cliente do aluno; correção é server-side.
- Link `wa.me` (WhatsApp) montado no client a partir da URL pública da trilha.
- PDF gerado no backend Go (biblioteca `maroto` ou `gofpdf`).

## Contrato de IA (agnóstico a marca)

Domínio depende da interface, não do SDK:

```go
type Generator interface {
    Generate(ctx context.Context, skill BnccSkill, duracaoMin int) (LessonData, error)
    Enhance(ctx context.Context, draft LessonData, skill BnccSkill) (LessonData, error)
}
```

Adapter LangChainGo (única unidade que conhece a IA); provider/modelo de env. Structured output forçado por JSON schema:

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

`Generate` cria do zero a partir da habilidade; `Enhance` recebe o rascunho no prompt e devolve versão melhorada com o mesmo schema. Trocar de IA = mudar `LLM_PROVIDER`/`LLM_MODEL`.

## Fluxo core

```
Professor (NextJS, autenticado)
  escolhe modo: manual | IA completa | IA aprimora
    → POST /api/lessons[/generate|/:id/enhance]
    → grava lesson_plan + trilha (status=rascunho, origem=<modo>)
  edita inline qualquer campo (PATCH)
  botão "Publicar trilha" → código curto (TR-XXXX)
       ↓
Aluno (NextJS público, /t/TR-XXXX)
  digita nome → vê trilha (tópicos + resumo, um por vez) → quiz autocorrigido
  gamificação: pontos + barra de progresso
       ↓
Professor: dashboard de turma (quem completou, acertos)
Export: PDF leve (Go) + link wa.me (WhatsApp)
```

## Tratamento de erro e testes

- LLM falha/timeout → `lesson_plans.status='falha'`, professor re-tenta ou cai no modo manual.
- Unit: mock da interface `Generator`; testa scoring do quiz e parse de `LessonData` a partir de fixture JSON.
- Integração: queries sqlc testadas contra Postgres dockerizado (migrations goose aplicadas no setup).
- Feature: publicar trilha → acessar por código → submeter quiz → assert pontos corretos.
- Frontend: server components buscam a API Go; páginas do aluno públicas, páginas do professor guardadas por middleware de sessão.

## Escopo do MVP

Inclui: fluxo professor→aluno completo, três modos de autoria (manual / IA completa / IA aprimora), edição manual pós-geração, gamificação do quiz, export PDF/WhatsApp, dashboard de turma.

Cortado (→ pitch "Próximos Passos"): multi-turma, contas de aluno reais, analytics avançado. Export é o primeiro candidato a corte se o tempo apertar.

## Impacto esperado

Reduz horas de planejamento (uma geração cobre aula + material do aluno) e dá foco ao aluno (um tópico por vez + quiz). Mensurável no pitch: tempo de planejamento, taxa de conclusão de trilha.
