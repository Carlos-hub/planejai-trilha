"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Plus, BookOpen, CheckCircle2, PencilLine } from "lucide-react";
import type { LucideIcon } from "lucide-react";
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
  const total = lessons?.length ?? 0;
  const prontas = lessons?.filter((l) => l.status === "pronto").length ?? 0;
  const rascunhos = lessons?.filter((l) => l.status === "rascunho").length ?? 0;

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-8 px-4 py-6 sm:px-6 sm:py-8">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">
            Painel
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Planeje aulas e acompanhe as trilhas dos seus alunos.
          </p>
        </div>
        <Button size="lg" render={<Link href="/lessons/new" />}>
          <Plus className="size-4" />
          Nova aula
        </Button>
      </header>

      {!error && (
        <div className="grid grid-cols-3 gap-3 sm:gap-4">
          <StatCard label="Aulas" value={total} icon={BookOpen} tone="brand" />
          <StatCard
            label="Prontas"
            value={prontas}
            icon={CheckCircle2}
            tone="amber"
          />
          <StatCard
            label="Rascunhos"
            value={rascunhos}
            icon={PencilLine}
            tone="muted"
          />
        </div>
      )}

      <section className="flex flex-col gap-4">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
          Minhas aulas
        </h2>

        {loading && (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {[0, 1, 2].map((i) => (
              <div
                key={i}
                className="h-44 animate-pulse rounded-xl border bg-muted/40"
              />
            ))}
          </div>
        )}

        {error && (
          <p
            className="rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive"
            role="alert"
          >
            Não foi possível carregar suas aulas. Tente novamente mais tarde.
          </p>
        )}

        {lessons && lessons.length === 0 && (
          <div className="flex flex-col items-center gap-4 rounded-2xl border border-dashed bg-card/50 px-6 py-16 text-center">
            <div className="flex size-12 items-center justify-center rounded-2xl bg-brand-muted text-primary">
              <BookOpen className="size-6" />
            </div>
            <div>
              <p className="font-medium">Nenhuma aula ainda</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Crie sua primeira aula — manual ou gerada com IA.
              </p>
            </div>
            <Button render={<Link href="/lessons/new" />}>
              <Plus className="size-4" />
              Criar primeira aula
            </Button>
          </div>
        )}

        {lessons && lessons.length > 0 && (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {lessons.map((lesson) => (
              <LessonCard key={lesson.id} lesson={lesson} />
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

function StatCard({
  label,
  value,
  icon: Icon,
  tone,
}: {
  label: string;
  value: number;
  icon: LucideIcon;
  tone: "brand" | "amber" | "muted";
}) {
  const tones = {
    brand: "bg-brand-muted text-primary",
    amber: "bg-amber/20 text-amber-foreground",
    muted: "bg-muted text-muted-foreground",
  } as const;
  return (
    <div className="flex flex-col gap-2 rounded-xl border bg-card p-4">
      <div
        className={`flex size-8 items-center justify-center rounded-lg ${tones[tone]}`}
      >
        <Icon className="size-4" />
      </div>
      <div className="text-2xl font-semibold tracking-tight tabular-nums">
        {value}
      </div>
      <div className="text-xs font-medium text-muted-foreground">{label}</div>
    </div>
  );
}
