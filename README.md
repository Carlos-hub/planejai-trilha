# PlanejAI + Trilha

Planejador de aulas alinhado à **BNCC** para professores, que gera automaticamente uma **trilha de estudo do aluno** (acesso por código, quiz autocorrigido e gamificação) a partir de uma única geração via IA.

> MVP do Hackathon FIAP — Pós-Tech (6FSDT).

---

## Visão geral

O professor escolhe uma habilidade da BNCC (ou uma unidade do currículo, incluindo matérias extras como Educação Financeira) e, com um clique, a IA produz o **plano de aula** e uma **trilha do aluno** (tópicos + quiz). A trilha é publicada com um **código curto**; o aluno acessa sem login, estuda os tópicos, faz o quiz autocorrigido e ganha pontos. O professor acompanha a turma num dashboard.

### Fluxo do MVP

```
Professor → login → cria plano (BNCC/currículo + duração)
         → "Gerar com IA" → publica trilha
Aluno    → abre /t/CÓDIGO → estuda tópicos → faz quiz → vê pontuação
Professor → dashboard da turma (tentativas, média, conclusões)
```

---

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Backend | Go 1.23, API REST |
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
- Go 1.23 (opcional, para desenvolvimento local)
- Node.js 20+ (opcional, para desenvolvimento local)

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

### Configurar a IA

No `.env`, defina o provider, modelo e chave:

```dotenv
LLM_PROVIDER=anthropic
LLM_MODEL=claude-opus-4-8
ANTHROPIC_API_KEY=sk-ant-xxx
```

Para outro provider (OpenAI, Gemini, etc.), mude `LLM_PROVIDER` e `LLM_MODEL`, informe a chave correspondente.

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
- `LLM_PROVIDER`: provider de IA (`anthropic`, `openai`, etc.)
- `LLM_MODEL`: modelo dentro do provider
- `ANTHROPIC_API_KEY`: chave de autenticação (ou equivalente do provider)
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
- **Geração com IA** do plano de aula + trilha (structured output, provider configurável).
- **Trilha do aluno** por código curto, sem login.
- **Quiz autocorrigido** com pontuação e barra de progresso (gamificação).
- **Dashboard da turma** (tentativas, média de pontos, conclusões).
- **Export PDF** da trilha + link de compartilhamento no WhatsApp.

---

## Referências

- [Backend README](./backend/README.md) (quando criado)
- [Frontend README](./frontend/README.md) (quando criado)
- [Especificação do Projeto](./spec/)
