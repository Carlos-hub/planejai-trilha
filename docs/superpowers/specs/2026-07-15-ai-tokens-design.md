# Design — Token de IA por Professor (Subsistema B, iteração 1)

**Data:** 2026-07-15
**Status:** Aprovado (aguardando review do spec)
**Escopo:** Segundo dos três subsistemas. A (turmas + alunos) concluído (PR #3). Esta é a **iteração 1 de B**: apenas o token cadastrado pelo professor. Token de escola + papel Admin ficam para uma iteração/sub-spec futura de B.

---

## Contexto

Hoje a geração com IA usa um único `lesson.Generator` (`d.Gen`) construído no startup a partir de variáveis de ambiente (`LLM_PROVIDER`, `LLM_MODEL`, `ANTHROPIC_API_KEY`) — ver `internal/lesson/langchain.go` (`NewLangChainGenerator`) e `internal/http/lesson_handlers.go` (`generateLesson`/`enhanceLesson`, que fazem nil-check de `d.Gen` → 503). Só o provider `anthropic` está ligado. O provider é escolhido uma vez, global para toda a plataforma.

Esta iteração troca esse modelo global por **resolução por request**: cada professor cadastra o próprio token de IA e escolhe o provider; a geração usa o token do professor logado. O token é **criptografado em repouso** para que um vazamento do banco não exponha credenciais.

## Objetivo

Permitir que o professor:
1. Cadastre/atualize o próprio token de IA escolhendo um provider entre Claude, GPT, Gemini, Deepseek, Llama.
2. Veja se já tem token configurado (e qual provider), sem que o token em texto seja devolvido.
3. Remova o token.
4. Gere/aprimore planos usando o próprio token; sem token → geração indisponível (503).

## Não-objetivos (fora do escopo desta iteração)

- Entidade **Escola** + papel **Admin** + token de escola. (Precedência decidida para o futuro: Escola > Professor > Plataforma — registrada, não implementada.)
- Fallback para token de plataforma quando o professor não tem token (decisão: **503**, sem fallback).
- Seleção de modelo editável pelo professor (modelo é **fixo por provider**).
- Rotação de chave / KMS gerenciado (chave única via env).
- Validação online do token no momento do save (test-call). O save apenas armazena.

---

## Segurança — criptografia em repouso

Novo pacote `internal/secret`.

- **Cifra:** AES-256-GCM (autenticada). Nonce aleatório de 12 bytes por operação.
- **Chave-mestra:** variável de ambiente `TOKEN_ENC_KEY`, contendo 32 bytes codificados em base64 (`base64.StdEncoding`). Decodificação deve resultar em exatamente 32 bytes, senão erro de inicialização.
- **Interface:**
  ```go
  package secret

  type Box struct { /* holds the AEAD */ }

  // NewBox decodes a base64 32-byte key and builds an AES-256-GCM Box.
  func NewBox(base64Key string) (*Box, error)

  // Seal encrypts plaintext, returning (ciphertext, nonce).
  func (b *Box) Seal(plaintext []byte) (ciphertext, nonce []byte, err error)

  // Open decrypts ciphertext given its nonce.
  func (b *Box) Open(ciphertext, nonce []byte) (plaintext []byte, err error)
  ```
- Se `TOKEN_ENC_KEY` estiver ausente/ inválida na subida da API, `d.Secret` fica `nil`; nesse estado, salvar token → 500 com mensagem clara e a geração continua 503. A API não deve entrar em pânico por falta da chave (degrada).
- O plaintext do token **nunca** é logado, **nunca** retornado por nenhum endpoint. Persistimos apenas `token_ciphertext` + `token_nonce`.

## Modelo de dados

Nova migration `backend/migrations/00005_ai_tokens.sql` (goose up/down).

```sql
-- +goose Up
CREATE TABLE ai_tokens (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  provider TEXT NOT NULL CHECK (provider IN ('anthropic','openai','googleai','deepseek','llama')),
  token_ciphertext BYTEA NOT NULL,
  token_nonce BYTEA NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE ai_tokens;
```

- Um token por professor (`user_id UNIQUE`); atualizar = upsert.
- `provider` é o identificador interno; os rótulos de UI (Claude/GPT/Gemini/Deepseek/Llama) mapeiam para esses valores.

Queries sqlc (`internal/store/queries/ai_tokens.sql`):
```sql
-- name: UpsertAIToken :one
INSERT INTO ai_tokens (user_id, provider, token_ciphertext, token_nonce)
VALUES ($1,$2,$3,$4)
ON CONFLICT (user_id) DO UPDATE
  SET provider=EXCLUDED.provider,
      token_ciphertext=EXCLUDED.token_ciphertext,
      token_nonce=EXCLUDED.token_nonce,
      updated_at=now()
RETURNING *;
-- name: GetAIToken :one
SELECT * FROM ai_tokens WHERE user_id=$1;
-- name: DeleteAIToken :exec
DELETE FROM ai_tokens WHERE user_id=$1;
```

## Factory multi-provider

Refatorar `internal/lesson/langchain.go`: extrair a construção do generator para uma função parametrizada, sem depender de env.

```go
// NewGeneratorForProvider builds a Generator for the given provider using the
// supplied API key. Model is fixed per provider (see defaultModels). Only
// internal/lesson imports provider-specific packages.
func NewGeneratorForProvider(ctx context.Context, provider, apiKey string) (Generator, error)
```

Mapa de modelos default (ajustável num único lugar):

| Rótulo UI | provider | backend langchaingo | modelo default |
|-----------|----------|---------------------|----------------|
| Claude | `anthropic` | `llms/anthropic` | `claude-opus-4-8` |
| GPT | `openai` | `llms/openai` | `gpt-4o` |
| Gemini | `googleai` | `llms/googleai` | `gemini-2.0-flash` |
| Deepseek | `deepseek` | `llms/openai` com `WithBaseURL("https://api.deepseek.com")` | `deepseek-chat` |
| Llama | `llama` | `llms/openai` com `WithBaseURL("https://api.groq.com/openai/v1")` | `llama-3.3-70b-versatile` |

- `anthropic`: `anthropic.New(anthropic.WithModel(model), anthropic.WithToken(apiKey))`.
- `openai` / `deepseek` / `llama`: `openai.New(openai.WithModel(model), openai.WithToken(apiKey), [openai.WithBaseURL(...)])`.
- `googleai`: `googleai.New(ctx, googleai.WithAPIKey(apiKey), googleai.WithDefaultModel(model))`.
- Provider desconhecido → erro.
- `NewLangChainGenerator()` (env-based) pode ser removido ou mantido só para um caminho de bootstrap opcional; a geração por request não o usa mais. A remoção é preferível para não deixar dois caminhos; verificar usos antes.

Nota de dependências: adiciona os pacotes `github.com/tmc/langchaingo/llms/openai` e `.../llms/googleai` (já usamos `.../llms/anthropic`). Rodar `go mod tidy`; se a rede bloquear o download e o módulo não estiver em cache, isso é um risco a sinalizar na implementação.

## Resolução por request

Em `internal/http`:

- `Deps` ganha:
  - `Secret *secret.Box` (pode ser nil se a chave não estiver configurada).
  - `NewGen func(ctx context.Context, provider, apiKey string) (lesson.Generator, error)` — default aponta para `lesson.NewGeneratorForProvider`; testes injetam um mock que registra o provider pedido.
  - O campo `Gen lesson.Generator` global é removido (ou deixado só para compatibilidade de testes antigos, se necessário — preferir remover e ajustar os testes).
- Helper:
  ```go
  // generatorForUser loads the professor's stored AI token, decrypts it, and
  // builds a Generator. Returns a sentinel that callers map to 503 when no
  // token is configured or the secret box is unavailable.
  func (d Deps) generatorForUser(ctx context.Context, userID int64) (lesson.Generator, error)
  ```
  - Sem `d.Secret` → erro → 503.
  - `GetAIToken` retorna `pgx.ErrNoRows` → 503 "configure seu token de IA".
  - Decripta com `d.Secret.Open`; erro → 500.
  - `d.NewGen(ctx, row.Provider, plaintext)` → Generator.
- `generateLesson` e `enhanceLesson` substituem o uso de `d.Gen` por `generatorForUser`; se ele retornar o sentinel de "sem token", respondem 503 com mensagem clara.

## API — token do professor (auth de professor)

- `PUT /api/me/ai-token` — body `{ "provider": "...", "token": "..." }`.
  - Valida `provider` no conjunto permitido (`anthropic|openai|googleai|deepseek|llama`), senão 400.
  - `token` não vazio, senão 400.
  - `d.Secret` nil → 500 "criptografia de token indisponível".
  - `Seal(token)` → `UpsertAIToken`. Resposta 200 `{ "provider": "...", "configured": true }` (nunca o token).
- `GET /api/me/ai-token` — 200 `{ "configured": bool, "provider": "..."|null }`. Nunca inclui o token.
- `DELETE /api/me/ai-token` — remove; 204.

Rotas registradas no grupo de professor (`RequireAuth`) em `router.go`.

## Frontend

- Página de perfil/dados do professor (rota nova sob `(teacher)`, ex.: `/perfil`), com uma seção "Token de IA":
  - Dropdown de provider com rótulos Claude / GPT / Gemini / Deepseek / Llama (value = provider interno).
  - Campo de token (`type=password`), botão **Salvar** (`PUT`).
  - Mostra estado atual (`GET`): "Configurado — provider X" ou "Nenhum token"; botão **Remover** (`DELETE`).
  - Deixa claro que o token não é exibido novamente após salvo.
- Link "Perfil" (ou ícone) no nav do professor.
- Onde a geração é acionada (fluxo de criação de aula), tratar 503 com mensagem orientando o professor a cadastrar o token em Perfil.

## Testes

- `internal/secret`: round-trip `Seal`/`Open`; nonces diferentes em dois `Seal` do mesmo plaintext; `Open` com chave errada falha; `NewBox` rejeita chave que não decodifica para 32 bytes.
- `internal/lesson`: `NewGeneratorForProvider` constrói sem erro para cada provider (com uma chave fake — não faz chamada de rede); provider desconhecido → erro.
- `internal/store`: upsert de `ai_tokens` (segundo upsert troca provider/ciphertext, mantém uma linha por user).
- `internal/http`:
  - `PUT /api/me/ai-token` provider inválido → 400; sucesso → 200 `{provider,configured:true}` e o token não aparece na resposta; persiste ciphertext (não o plaintext).
  - `GET` após save → `{configured:true, provider}`; nunca o token.
  - `DELETE` → 204; `GET` depois → `{configured:false}`.
  - `generateLesson` sem token → 503; com token → usa `d.NewGen` com o provider salvo (mock factory registra o provider recebido e devolve um generator mock).

## Camada Go (padrões existentes)

- Queries sqlc em `internal/store/queries/ai_tokens.sql` → `sqlc generate`. `BYTEA` vira `[]byte` no Go.
- Handlers em `internal/http/ai_token_handlers.go`; rotas em `router.go` no grupo `RequireAuth`.
- `internal/secret/secret.go` novo, sem dependências externas além da stdlib (`crypto/aes`, `crypto/cipher`, `crypto/rand`, `encoding/base64`).
- `cmd/api/main.go`: construir `secret.NewBox(os.Getenv("TOKEN_ENC_KEY"))` (degradando para nil se ausente) e o `NewGen` default; injetar em `Deps`; remover a construção do `Gen` global env-based.

## Riscos / decisões abertas para a implementação

- Download dos módulos langchaingo `openai`/`googleai` pode falhar se a rede estiver bloqueada e não houver cache — sinalizar se ocorrer.
- Modelos default por provider são sensíveis ao tempo; ficam num único mapa e são facilmente ajustáveis.
- `googleai.New` exige `ctx`; por isso a factory recebe `ctx`.
- Remoção do `Gen` global exige ajustar/rever os testes existentes que o injetavam (`internal/http` testes de generate/enhance).
