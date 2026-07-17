# Design — Turmas + Contas de Aluno (Subsistema A)

**Data:** 2026-07-15
**Status:** Aprovado (aguardando review do spec)
**Escopo:** Primeiro de três subsistemas independentes. B (Escola + Admin + tokens de IA multi-provider) e C (notas por tipo + PDF com QRCode) virão como specs próprios.

---

## Contexto

Hoje o modelo é: `User` (professor) → `LessonPlan` → `StudyTrail` (publicada com código curto) → `TrailTopic` + `Quiz`/`QuizQuestion` → `StudentAttempt` (anônimo, só `nome_aluno`) + `AttemptAnswer`.

Não existe conceito de **turma** nem **conta de aluno**. Aluno acessa `/t/CÓDIGO` sem login e a tentativa grava apenas o nome digitado.

Este subsistema adiciona turmas geridas pelo professor, contas de aluno com login/senha (criadas via import CSV) e atividades restritas (gated) aos alunos matriculados, mantendo o fluxo anônimo por código para trails avulsas.

## Objetivo

Permitir que o professor:
1. Crie turmas.
2. Importe alunos via CSV; o sistema gera credenciais (usuário + senha inicial) e as devolve uma única vez para o professor distribuir.
3. Atribua uma atividade a uma turma, tornando-a acessível apenas por alunos logados daquela turma, com a tentativa vinculada à identidade do aluno.

## Não-objetivos (fora do escopo A)

- Notas/pesos por tipo de atividade (tarefa / tema de casa / prova) → Subsistema C.
- Geração de PDF de prova + QRCode (gabarito / submissão) → Subsistema C.
- Entidade Escola, papel Admin, tokens de IA por escola/professor, seleção de provider (Gemini/Claude/GPT/Llama/Deepseek) → Subsistema B.
  - **Requisito de segurança já fixado para B (não perder):** os tokens/credenciais de acesso à IA são **obrigatoriamente criptografados em repouso** pelo backend Go. Em caso de vazamento do banco, o token permanece protegido. Diretrizes: cifra simétrica autenticada (ex.: AES-256-GCM); chave-mestra vem de env/KMS, **nunca** armazenada no banco; persistir apenas ciphertext + nonce; plaintext do token nunca é logado nem retornado em GET (apenas write-only / mascarado na UI).
- Auto-cadastro de aluno / add avulso de aluno (só import CSV nesta entrega).

---

## Modelo de dados

Nova migration `backend/migrations/00004_turmas_alunos.sql` (goose up/down).

### Novas tabelas

```sql
CREATE TABLE turmas (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  nome TEXT NOT NULL,
  etapa TEXT NOT NULL DEFAULT '',
  anos INT[] NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE students (
  id BIGSERIAL PRIMARY KEY,
  turma_id BIGINT NOT NULL REFERENCES turmas(id) ON DELETE CASCADE,
  nome TEXT NOT NULL,
  usuario TEXT UNIQUE NOT NULL,
  senha_hash TEXT NOT NULL,
  matricula TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE student_sessions (
  id TEXT PRIMARY KEY,
  student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Alterações em tabelas existentes

```sql
ALTER TABLE study_trails
  ADD COLUMN turma_id BIGINT REFERENCES turmas(id) ON DELETE SET NULL;

ALTER TABLE student_attempts
  ADD COLUMN student_id BIGINT REFERENCES students(id) ON DELETE SET NULL;
```

### Regras / invariantes

- Aluno pertence a **exatamente uma** turma. O mesmo aluno físico em duas turmas = duas contas distintas.
- `students.usuario` é **único global**.
- `study_trails.turma_id`:
  - `NULL` → comportamento atual: acesso anônimo por `codigo`.
  - setado → atividade **gated**: só alunos logados daquela turma acessam.
- `student_attempts.student_id`:
  - preenchido quando o aluno está logado (atividade gated);
  - `nome_aluno` continua sendo gravado (denormalizado) para display e compatibilidade com tentativas anônimas.
- `ON DELETE SET NULL` em `turma_id`/`student_id` preserva atividades publicadas e histórico de tentativas se a turma/aluno for removido.

---

## Autenticação do aluno

Reusa `internal/auth` (`HashPassword`, `CheckPassword`, `NewSessionID`, `SessionTTL`) — nada novo em cripto.

- **Geração de `usuario`:** slug do nome (minúsculo, sem acento, sem espaço) + sufixo aleatório curto (ex.: `maria.silva.k7x2`). Em colisão, regenera o sufixo até ser único.
- **Senha inicial:** aleatória legível (~8 chars, alfanumérico sem caracteres ambíguos). Retornada em **texto puro apenas uma vez**, no response do import. Persistimos somente `senha_hash`.
- **Login:** `POST /api/student/login` `{usuario, senha}` → cria `student_session`, seta cookie `student_session` (HttpOnly, mesmos atributos do cookie de professor). Middleware `requireStudent` resolve a sessão → `student`.
- **Sessões separadas:** cookie e tabela distintos dos do professor, para não misturar papéis. Um request pode carregar sessão de professor OU de aluno, nunca ambos com significado cruzado.
- **Troca de senha (opcional, a qualquer hora):** `POST /api/student/password` `{senha_atual, senha_nova}` (auth aluno). Não é forçada no primeiro acesso.
- **Logout:** `POST /api/student/logout` deleta a sessão.

---

## Endpoints — professor (auth de professor + ownership da turma)

- `POST /api/turmas` `{nome, etapa?, anos?}` → cria turma do professor logado.
- `GET /api/turmas` → lista turmas do professor.
- `GET /api/turmas/:id` → turma + lista de alunos (sem senha; sem hash).
- `PATCH /api/turmas/:id` `{nome?, etapa?, anos?}` → edita metadados.
- `DELETE /api/turmas/:id` → remove turma (cascade em students/sessions).
- `POST /api/turmas/:id/students/import` — corpo CSV (multipart ou `text/csv`).
  - Colunas: `nome` (obrigatória), `matricula` (opcional). Header case-insensitive.
  - Para cada linha: gera `usuario` + senha inicial, insere `student`.
  - **Response:** `[{ nome, matricula, usuario, senha }]` com a senha em texto puro (única exibição), para o professor exportar/imprimir.
  - Erros de linha (nome vazio) retornam relatório por linha; linhas válidas ainda são inseridas (import parcial) — decisão a validar na implementação, default: aborta tudo se qualquer linha for inválida (transação única) para simplicidade e previsibilidade.

Todos os endpoints de turma verificam que `turma.user_id == professor logado`; caso contrário `404` (não vaza existência).

## Endpoints — aluno

- `POST /api/student/login`
- `POST /api/student/logout`
- `POST /api/student/password`
- (leitura da trail gated reusa o handler público existente, agora com checagem de sessão de aluno)

## Publicação com turma

O fluxo de publish da atividade ganha um campo opcional `turma_id`. Ao publicar:
- `turma_id` setado → grava em `study_trails.turma_id`; valida ownership (turma pertence ao professor).
- ausente → publica como trail avulsa (código anônimo), como hoje.

## Fluxo do aluno (gated)

Ao acessar `/t/CÓDIGO`:
1. Resolve a trail pelo código.
2. Se `trail.turma_id` é `NULL` → fluxo anônimo atual (inalterado).
3. Se setado:
   - sem `student_session` válida → redireciona para `/aluno/login`.
   - com sessão, mas aluno não pertence a `trail.turma_id` → `403`.
   - ok → renderiza trail; ao submeter, a tentativa grava `student_id` e `nome_aluno` da conta.
4. Tentativa anônima em atividade gated é negada.

---

## Frontend

### Professor
- `/turmas` — lista de turmas + criar turma.
- `/turmas/[id]` — detalhe: metadados, lista de alunos, **upload de CSV**, e tabela de credenciais geradas com **download/print** (a senha em texto só aparece aqui, logo após o import).
- Seletor de turma (opcional) no passo de publicação da atividade.

### Aluno
- `/aluno/login` — usuário + senha.
- Página de trail gated (reusa a UI da trail; exige login quando gated).
- (Opcional) troca de senha na área do aluno.

---

## Camada Go (padrões existentes)

- Queries sqlc em `internal/store/queries/`: `turmas.sql`, `students.sql`, `student_sessions.sql` → `sqlc generate`.
- Handlers em `internal/http/`: `turma_handlers.go`, `student_handlers.go`; rotas registradas em `router.go`.
- Middleware `requireStudent` em `middleware.go`, espelhando `requireUser` existente.
- CSV parseado com `encoding/csv` da stdlib.

---

## Testes (Go, espelhando `*_handlers_test.go`)

- Turma CRUD: criar, listar só as próprias, ownership em GET/PATCH/DELETE (404 p/ turma de outro professor).
- Import CSV: gera `usuario` único; senha retornada uma vez; hash persistido, texto puro não; import de N linhas cria N alunos.
- Login do aluno: credencial válida → sessão; inválida → 401.
- Troca de senha: senha_atual errada → 401; sucesso muda o hash.
- Acesso gated: aluno da turma acessa; aluno de outra turma → 403; anônimo → redirect/401; trail sem turma continua anônima.
- Attempt: submissão gated grava `student_id` correto.

---

## Riscos / decisões abertas para a implementação

- Import parcial vs. transação única em CSV inválido → default: transação única (aborta tudo).
- Formato exato do slug de `usuario` e tamanho do sufixo → ajustável; garantir unicidade por retry.
- Atributos de cookie (`Secure`, `SameSite`) devem seguir os já usados pelo cookie do professor no ambiente.
