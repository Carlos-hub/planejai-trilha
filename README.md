# PlanejAI + Trilha

Planejador de aulas alinhado à **BNCC** para professores, que gera automaticamente uma **trilha de estudo do aluno** (acesso por código, quiz autocorrigido e gamificação) a partir de uma única geração via IA.

> MVP do Hackathon FIAP — Pós-Tech (6FSDT).

---

## Visão geral

O professor escolhe uma habilidade da BNCC (ou uma unidade do currículo, incluindo matérias extras como Educação Financeira) e, com um clique, a IA produz o **plano de aula** e uma **trilha do aluno** (tópicos + quiz). A trilha pode ser publicada de duas formas: com um **código curto** de acesso anônimo (sem login), ou **atrelada a uma turma**, caso em que o aluno precisa fazer login com usuário e senha para responder. O professor gerencia suas **turmas** e importa **contas de aluno** via CSV, com credenciais geradas automaticamente. O professor acompanha a turma num dashboard.

### Fluxo do MVP

```
Professor → login → cria turma → importa alunos via CSV (nome, matrícula opcional)
         → sistema gera usuário + senha por aluno → professor baixa/imprime credenciais (mostradas 1x)
Professor → cria plano (BNCC/currículo + duração) → "Gerar com IA"
         → publica trilha (atrelada a uma turma OU anônima por código curto)

Trilha atrelada a turma:
Aluno    → /aluno/login (usuário + senha) → estuda tópicos → faz quiz → vê pontuação
         → aluno pode trocar sua própria senha a qualquer momento

Trilha anônima (sem turma):
Aluno    → abre /t/CÓDIGO → estuda tópicos → faz quiz → vê pontuação

Professor → dashboard da turma (tentativas, média, conclusões)
```

---

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Backend | Go 1.26, API REST |
| Frontend | Next.js, TypeScript, React |
| Banco | PostgreSQL 16 |
| Runtime | Docker Compose |
| IA (agnóstica) | Structured output via provider (Anthropic, OpenAI, etc.) |

---

## Arquitetura

Arquitetura de monorepo com backend Go (API REST) e frontend Next.js separados, comunicando-se via HTTP. O backend orquestra a geração de planos com IA (agnóstica por provider). O frontend renderiza painel do professor e páginas públicas do aluno.

```
┌──────────────────────────────────────────────────┐
│              Docker Compose                       │
├──────────────────────────────────────────────────┤
│  ┌──────────┐      ┌──────────┐    ┌──────────┐ │
│  │ Postgres │      │   API    │    │   Web    │ │
│  │    :5432 │      │   :8080  │    │  :3000   │ │
│  │ (db)     │      │   (Go)   │    │ (Next.js)│ │
│  └──────────┘      └──────────┘    └──────────┘ │
│       ↑                 ↑                ↓        │
│       └─────────────────┴────────────────┘       │
│                  HTTP API                        │
└──────────────────────────────────────────────────┘
```

---

## Quickstart

### Pré-requisitos

- Docker e Docker Compose
- Go 1.26 (opcional, para desenvolvimento local)
- Node.js 20+ (opcional, para desenvolvimento local)
- `curl` e `jq` (para rodar `scripts/smoke.sh`)

### Executar com Docker Compose

```bash
# Clonar o repositório
git clone <url> planejai
cd planejai

# Copiar variáveis de ambiente
cp .env.example .env

# Iniciar serviços (db, api, web)
docker compose up

# App em http://localhost:3000
# API em http://localhost:8080
# Postgres em localhost:5432
```

> Docker faz o build automático de `backend/` (Go + Dockerfile) e `frontend/` (Next.js + Dockerfile).

> A API roda as migrações do banco (via [goose](https://github.com/pressly/goose)) e o seed do professor demo automaticamente na inicialização — não é preciso rodar `goose` manualmente para subir um banco limpo.

### Smoke test (ponta a ponta)

Valida o fluxo completo (login → criar aula manual → publicar trilha → leitura pública → tentativa do aluno → pontuação → dashboard → export PDF) sem depender da IA (evita o endpoint `/generate`, que exige um token de IA real configurado no perfil do professor):

```bash
cp .env.example .env
bash scripts/smoke.sh
```

O script sobe `db`, `api` e `web` via `docker compose up -d --build`, espera o banco e a API ficarem saudáveis e roda as asserções via `curl` + `jq`. Ao final, imprime um resumo com o código da trilha, a pontuação obtida e `PDF OK`, e sai com código 0 em caso de sucesso.

Se o build do `web` estiver lento ou não for necessário (só quer validar a API), rode apenas `db` + `api`:

```bash
SMOKE_SKIP_WEB=1 bash scripts/smoke.sh
```

### Configurar a IA

A geração com IA usa um **token por professor**, não mais uma chave global no `.env`. Cada professor cadastra o próprio token na página **`/perfil`** (`PUT/GET/DELETE /api/me/ai-token`), escolhendo um provider entre:

| Provider  | Modelo (fixo) |
|-----------|---------------|
| Claude    | `claude-opus-4-8` |
| GPT       | `gpt-4o` |
| Gemini    | `gemini-2.0-flash` |
| Deepseek  | `deepseek-chat` |
| Llama     | `llama-3.3-70b-versatile` |

O token é **criptografado em repouso** (AES-256-GCM) antes de ser salvo no banco; nenhum endpoint devolve o token em texto plano depois de salvo. A chave mestra dessa criptografia vem da env `TOKEN_ENC_KEY` (veja abaixo) — **sem ela, salvar um token falha e a geração fica indisponível**.

Se o professor tentar gerar um plano/trilha **sem ter configurado um token**, a API responde **`503`** pedindo para configurar o token no perfil (não há fallback de provider da plataforma).

> A antiga configuração global via `.env` (`LLM_PROVIDER` / `LLM_MODEL` / `ANTHROPIC_API_KEY`) foi **removida**: essas variáveis não existem mais e não influenciam a geração.

---

## Desenvolvimento local

### Backend (Go)

```bash
cd backend

# Instalar dependências
go mod download

# Rodar servidor (requer DATABASE_URL + env vars)
PORT=8080 go run ./cmd/api

# Testes
go test ./...
```

### Frontend (Next.js)

```bash
cd frontend

# Instalar dependências
npm install

# Rodar servidor (requer NEXT_PUBLIC_API_URL)
NEXT_PUBLIC_API_URL=http://localhost:8080 npm run dev

# Build
npm run build
```

---

## Variáveis de Ambiente

Consulte `.env.example`:

- `DATABASE_URL`: conexão PostgreSQL (formato libpq)
- `TOKEN_ENC_KEY`: chave mestra (base64 de 32 bytes, ex. `openssl rand -base64 32`) usada para criptografar (AES-256-GCM) o token de IA de cada professor em repouso. Obrigatória para salvar token e para a geração com IA funcionar.
- `SESSION_SECRET`: chave para sessões (mín. 32 bytes)
- `PORT`: porta do backend (default 8080)
- `NEXT_PUBLIC_API_URL`: URL da API vista pelo frontend (ex. `http://localhost:8080`)

---

## Estrutura do Projeto

```
planejai/
├── backend/                    # Go API
│   ├── cmd/api/                # entry point
│   ├── internal/               # código privado do módulo
│   ├── go.mod / go.sum         # dependências Go
│   └── Dockerfile              # build Go
├── frontend/                   # Next.js
│   ├── app/                    # app router
│   ├── components/             # React components
│   ├── package.json            # dependências npm
│   └── Dockerfile              # build Next.js
├── docker-compose.yml          # orquestração (db, api, web)
├── .env.example                # template de variáveis
└── README.md                   # este arquivo
```

---

## Funcionalidades

- **Currículo BNCC completo** por série e trimestre — do 1º ano do Fundamental à 3ª série do Médio.
- **Matérias extras** (fora da BNCC): Educação Financeira, Projeto de Vida, Empreendedorismo, Educação Digital e Tecnologia, Cidadania/Ética/Direitos Humanos.
- **Geração com IA** do plano de aula + trilha (structured output, provider configurável por professor).
- **Token de IA por professor**, configurado em `/perfil` e criptografado em repouso (AES-256-GCM); sem token configurado, a geração retorna `503`.
- **Turmas**: o professor cria, lista, edita e remove turmas (`/turmas`, `/turmas/[id]`).
- **Contas de aluno via import CSV**: upload de CSV (`nome` obrigatória, `matricula` opcional) numa turma gera usuário + senha inicial por aluno (apenas o hash bcrypt é armazenado); as credenciais em texto plano são exibidas **uma única vez** (download em CSV ou impressão) para o professor distribuir. O aluno pode trocar sua senha depois.
- **Trilha do aluno** por código curto, sem login — continua funcionando normalmente para trilhas não atreladas a turma.
- **Atividade atrelada à turma**: ao publicar, o professor pode vincular a trilha a uma turma; nesse caso o aluno precisa logar em `/aluno/login` (usuário + senha) e estar matriculado na turma para responder, e a tentativa fica associada à identidade do aluno.
- **Quiz autocorrigido** com pontuação e barra de progresso (gamificação).
- **Dashboard da turma** (tentativas, média de pontos, conclusões).
- **Export PDF** da trilha + link de compartilhamento no WhatsApp.

### Controle de acesso

- **Senhas** (professor e aluno) são armazenadas apenas como hash **bcrypt**; a senha inicial do aluno é exibida em texto plano **uma única vez**, no retorno do import CSV, e nunca é devolvida por nenhum outro endpoint.
- **Sessões** de professor e de aluno são separadas (cookies `sid` × `student_sid`, tabelas distintas), sem cruzamento de papéis.
- **Trilha atrelada a turma** é bloqueada para acesso anônimo ou de aluno de outra turma em **todos** os caminhos de leitura e resposta: página da trilha, `export.pdf` e submissão de tentativa. Só o aluno matriculado e logado acessa; a submissão de uma tentativa só é aceita do próprio aluno dono da tentativa.
- **Trilha sem turma** mantém o fluxo anônimo por código curto, inalterado.
- **Propriedade**: um professor só enxerga/edita as próprias turmas e só vincula trilha a turma que possui (caso contrário `404`, sem vazar existência).

### Próximos passos (não implementado)

- Notas/avaliação qualitativa por atividade (além da pontuação do quiz).
- Geração de PDF de credenciais com QR Code.
- Contas/token de administração para gestão em nível de escola.

---

## Referências

- [Backend README](./backend/README.md) (quando criado)
- [Frontend README](./frontend/README.md) (quando criado)
- [Especificação do Projeto](./spec/)
