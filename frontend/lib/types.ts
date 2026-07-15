// TypeScript interfaces mirroring backend JSON responses (field names are
// pt-BR, copied exactly from the Go response structs in backend/internal/http).

// --- auth (auth_handlers.go: login/register/me) ---
export interface Me {
  id: number;
  email: string;
  nome: string;
}

// --- lesson content shared shapes (lesson_handlers.go: lessonResponse.Trilha/Plano) ---
export interface Topico {
  titulo: string;
  resumo: string;
}

export interface Questao {
  enunciado: string;
  opcoes: string[];
  correta: number;
}

export interface Quiz {
  questoes: Questao[];
}

export interface Trilha {
  topicos: Topico[];
  quiz: Quiz;
}

export interface Plano {
  objetivos: string;
  metodologia: string;
  recursos: string;
  avaliacao: string;
}

// --- lesson_handlers.go: lessonSummary (GET /api/lessons list) ---
export interface LessonSummary {
  id: number;
  origem: string;
  status: string;
  duracao: number;
  created_at: string;
  updated_at: string;
}

// --- lesson_handlers.go: lessonResponse (lesson detail, owner-only, includes correta) ---
export interface Lesson {
  id: number;
  bncc_skill_id: number | null;
  duracao: number;
  origem: string;
  status: string;
  created_at: string;
  updated_at: string;
  plano: Plano;
  atividade: string;
  trilha: Trilha;
}

// --- bncc_handlers.go: bnccSkillResponse (GET /api/bncc-skills) ---
export interface BnccSkill {
  id: number;
  code: string;
  disciplina: string;
  ano: string;
  assunto: string;
  descricao: string;
}

// --- lesson_handlers.go: publishTrailResponse (POST /api/trails/:id/publish) ---
export interface PublishResult {
  codigo: string;
  publica_url: string;
}

// --- lesson_handlers.go: tentativaStats / trailStatsResponse (GET /api/trails/:id/stats) ---
export interface Tentativa {
  nome_aluno: string;
  pontos: number;
  concluido_em: string | null;
}

export interface TrailStats {
  total_alunos: number;
  concluidos: number;
  media_pontos: number;
  tentativas: Tentativa[];
}

// --- public_handlers.go: publicQuestao (GET /api/t/:code) — no "correta" field ---
export interface PublicQuestao {
  id: number;
  enunciado: string;
  opcoes: string[];
}

export interface PublicQuiz {
  questoes: PublicQuestao[];
}

export interface PublicTopico {
  ordem: number;
  titulo: string;
  resumo: string;
}

// --- public_handlers.go: publicTrailResponse (GET /api/t/:code) ---
export interface PublicTrail {
  titulo_aula: string;
  topicos: PublicTopico[];
  quiz: PublicQuiz;
}

// --- public_handlers.go: startAttemptResponse (POST /api/t/:code/attempt) ---
export interface AttemptStart {
  attempt_id: number;
}

// --- public_handlers.go: submitAnswersResponse (POST /api/attempts/:id/answers) ---
export interface AttemptResult {
  pontos: number;
  acertos: number;
  total: number;
}
