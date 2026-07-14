"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { apiFetch } from "@/lib/api";
import type { PublicTrail } from "@/lib/types";
import { NameGate, attemptStorageKey } from "@/components/name-gate";
import { TrailView } from "@/components/trail-view";
import { QuizRunner } from "@/components/quiz-runner";

type Phase = "name" | "topics" | "quiz";

interface StoredAttempt {
  attemptId: number;
  nome: string;
}

function readStoredAttempt(code: string): StoredAttempt | null {
  try {
    const raw = window.localStorage.getItem(attemptStorageKey(code));
    if (!raw) return null;
    const parsed = JSON.parse(raw);
    if (typeof parsed?.attemptId === "number" && typeof parsed?.nome === "string") {
      return parsed;
    }
    return null;
  } catch {
    return null;
  }
}

export default function PublicTrailPage() {
  const params = useParams<{ code: string }>();
  const code = params.code;

  const [trail, setTrail] = useState<PublicTrail | null>(null);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);

  const [phase, setPhase] = useState<Phase>("name");
  const [attemptId, setAttemptId] = useState<number | null>(null);
  const [nome, setNome] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setNotFound(false);
    setLoadError(null);

    apiFetch<PublicTrail>(`/api/t/${code}`)
      .then((data) => {
        if (cancelled) return;
        setTrail(data);
        const stored = readStoredAttempt(code);
        if (stored) {
          setAttemptId(stored.attemptId);
          setNome(stored.nome);
          setPhase("topics");
        } else {
          setPhase("name");
        }
      })
      .catch((err) => {
        if (cancelled) return;
        const message = err instanceof Error ? err.message : "";
        if (message.startsWith("API 404")) {
          setNotFound(true);
        } else {
          setLoadError("Não foi possível carregar esta trilha. Tente novamente mais tarde.");
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [code]);

  const handleStarted = useCallback((newAttemptId: number, newNome: string) => {
    setAttemptId(newAttemptId);
    setNome(newNome);
    setPhase("topics");
  }, []);

  const handleFinishTopics = useCallback(() => {
    setPhase("quiz");
  }, []);

  if (loading) {
    return (
      <main className="mx-auto flex min-h-svh w-full max-w-2xl flex-col items-center justify-center gap-2 px-4 py-8">
        <p className="text-sm text-muted-foreground">Carregando trilha...</p>
      </main>
    );
  }

  if (notFound) {
    return (
      <main className="mx-auto flex min-h-svh w-full max-w-2xl flex-col items-center justify-center gap-2 px-4 py-8 text-center">
        <h1 className="font-heading text-xl font-medium">Trilha não encontrada</h1>
        <p className="text-sm text-muted-foreground">
          Verifique se o link ou código está correto.
        </p>
      </main>
    );
  }

  if (loadError || !trail) {
    return (
      <main className="mx-auto flex min-h-svh w-full max-w-2xl flex-col items-center justify-center gap-2 px-4 py-8 text-center">
        <h1 className="font-heading text-xl font-medium">Algo deu errado</h1>
        <p className="text-sm text-muted-foreground">
          {loadError ?? "Não foi possível carregar esta trilha."}
        </p>
      </main>
    );
  }

  return (
    <main className="mx-auto flex min-h-svh w-full max-w-2xl flex-col gap-6 px-4 py-8">
      <header className="flex flex-col gap-1">
        <p className="text-sm text-muted-foreground">Trilha de estudos</p>
        <h1 className="font-heading text-xl font-medium">{trail.titulo_aula}</h1>
        {nome ? <p className="text-sm text-muted-foreground">Olá, {nome}!</p> : null}
      </header>

      {phase === "name" ? <NameGate code={code} onStarted={handleStarted} /> : null}

      {phase === "topics" ? (
        <TrailView trail={trail} onFinishTopics={handleFinishTopics} />
      ) : null}

      {phase === "quiz" && attemptId != null ? (
        <QuizRunner
          questoes={trail.quiz.questoes}
          attemptId={attemptId}
          code={code}
          publicaUrl={typeof window !== "undefined" ? window.location.href : undefined}
        />
      ) : null}
    </main>
  );
}
