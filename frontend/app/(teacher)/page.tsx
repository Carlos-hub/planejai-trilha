"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { apiFetch } from "@/lib/api";
import type { LessonSummary } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { LessonCard } from "@/components/lesson-card";

export default function DashboardPage() {
  const [lessons, setLessons] = useState<LessonSummary[] | null>(null);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    apiFetch<LessonSummary[]>("/api/lessons")
      .then((data) => {
        if (!cancelled) {
          setLessons(data);
          setError(null);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err : new Error(String(err)));
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const loading = lessons === null && !error;

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6 p-4 sm:p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">Minhas aulas</h1>
        <Button render={<Link href="/lessons/new" />}>Nova aula</Button>
      </div>

      {loading && (
        <p className="text-sm text-muted-foreground">Carregando aulas...</p>
      )}

      {error && (
        <p className="text-sm text-destructive" role="alert">
          Não foi possível carregar suas aulas. Tente novamente mais tarde.
        </p>
      )}

      {lessons && lessons.length === 0 && (
        <div className="flex flex-col items-center gap-3 rounded-xl border border-dashed p-12 text-center">
          <p className="text-sm text-muted-foreground">
            Nenhuma aula ainda — crie a primeira
          </p>
          <Button render={<Link href="/lessons/new" />}>Nova aula</Button>
        </div>
      )}

      {lessons && lessons.length > 0 && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {lessons.map((lesson) => (
            <LessonCard key={lesson.id} lesson={lesson} />
          ))}
        </div>
      )}
    </div>
  );
}
