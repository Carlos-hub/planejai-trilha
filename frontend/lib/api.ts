import type {
  AIProvider,
  AITokenStatus,
  Student,
  Turma,
  ImportedStudent,
  TurmaAula,
  StudentAula,
  LessonSummary,
} from "./types";

const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    credentials: "include",
    headers: { "Content-Type": "application/json", ...(init.headers ?? {}) },
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`API ${res.status}: ${body}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export function studentLogin(usuario: string, senha: string): Promise<Student> {
  return apiFetch<Student>("/api/student/login", {
    method: "POST",
    body: JSON.stringify({ usuario, senha }),
  });
}

// register creates a professor account. It does NOT log the user in — the
// caller sends them to the login page afterward.
export function register(nome: string, email: string, senha: string) {
  return apiFetch<{ id: number; email: string; nome: string }>(
    "/api/auth/register",
    { method: "POST", body: JSON.stringify({ nome, email, senha }) }
  );
}

export const listTurmas = () => apiFetch<Turma[]>("/api/turmas");

export const getTurma = (id: number) =>
  apiFetch<{
    turma: Turma;
    alunos: { id: number; nome: string; usuario: string; matricula: string | null }[];
    aulas: TurmaAula[];
  }>(`/api/turmas/${id}`);

export const createTurma = (input: { nome: string; etapa?: string; anos?: number[] }) =>
  apiFetch<Turma>("/api/turmas", { method: "POST", body: JSON.stringify(input) });

export async function importStudentsCSV(
  id: number,
  csvText: string
): Promise<{ criados: ImportedStudent[] }> {
  return apiFetch<{ criados: ImportedStudent[] }>(`/api/turmas/${id}/students/import`, {
    method: "POST",
    headers: { "Content-Type": "text/csv" },
    body: csvText,
  });
}

export const addStudent = (
  id: number,
  nome: string,
  matricula?: string
): Promise<ImportedStudent> =>
  apiFetch<ImportedStudent>(`/api/turmas/${id}/students`, {
    method: "POST",
    body: JSON.stringify({ nome, matricula: matricula?.trim() ? matricula.trim() : null }),
  });

export const getAIToken = () => apiFetch<AITokenStatus>("/api/me/ai-token");

export const saveAIToken = (provider: AIProvider, token: string) =>
  apiFetch<{ provider: AIProvider; configured: boolean }>("/api/me/ai-token", {
    method: "PUT",
    body: JSON.stringify({ provider, token }),
  });

export const deleteAIToken = () =>
  apiFetch<void>("/api/me/ai-token", { method: "DELETE" });

export const listLessons = () => apiFetch<LessonSummary[]>("/api/lessons");

export const attachTurmaLesson = (turmaId: number, lessonPlanId: number) =>
  apiFetch<{ lesson_plan_id: number; ordem: number }>(
    `/api/turmas/${turmaId}/lessons`,
    { method: "POST", body: JSON.stringify({ lesson_plan_id: lessonPlanId }) }
  );

export const detachTurmaLesson = (turmaId: number, lessonPlanId: number) =>
  apiFetch<void>(`/api/turmas/${turmaId}/lessons/${lessonPlanId}`, {
    method: "DELETE",
  });

export const reorderTurmaLessons = (turmaId: number, orderedIds: number[]) =>
  apiFetch<void>(`/api/turmas/${turmaId}/lessons`, {
    method: "PATCH",
    body: JSON.stringify({ ordered_ids: orderedIds }),
  });

export const getStudentLessons = () =>
  apiFetch<{ aulas: StudentAula[] }>("/api/student/lessons");
