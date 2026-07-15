import type { Student } from "./types";

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
