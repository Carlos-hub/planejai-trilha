"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Lock, ArrowRight, CheckCircle2 } from "lucide-react";
import { getStudentLessons } from "@/lib/api";
import type { StudentAula } from "@/lib/types";

export default function AlunoHomePage() {
  const [aulas, setAulas] = useState<StudentAula[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  useEffect(() => {
    getStudentLessons()
      .then((data) => setAulas(data.aulas))
      .catch((err) => {
        if (err instanceof Error && err.message.startsWith("API 401")) {
          router.push("/aluno/login?next=/aluno");
          return;
        }
        setError("Não foi possível carregar suas aulas.");
      });
  }, [router]);

  return (
    <div className="mx-auto flex max-w-2xl flex-col gap-6 px-4 py-8">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight">Minhas aulas</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Conclua cada aula para liberar a próxima.
        </p>
      </header>

      {error && (
        <p className="text-sm text-destructive" role="alert">
          {error}
        </p>
      )}

      {aulas === null && !error && (
        <p className="text-sm text-muted-foreground">Carregando…</p>
      )}

      {aulas && aulas.length === 0 && (
        <p className="rounded-lg border border-dashed bg-muted/30 px-4 py-6 text-center text-sm text-muted-foreground">
          Sua turma ainda não tem aulas. Volte mais tarde.
        </p>
      )}

      {aulas && aulas.length > 0 && (
        <ol className="flex flex-col gap-2">
          {aulas.map((a) => {
            const state = !a.codigo
              ? "locked"
              : a.concluido
                ? "done"
                : a.unlocked
                  ? "open"
                  : "locked";
            const inner = (
              <div
                className={`flex items-center gap-3 rounded-xl border p-4 ${
                  state === "locked"
                    ? "bg-muted/40 opacity-70"
                    : "bg-card hover:border-primary/40"
                }`}
              >
                <span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary/10 text-sm font-semibold text-primary">
                  {a.ordem}
                </span>
                <div className="flex-1">
                  <p className="text-sm font-medium">{a.label}</p>
                  {state === "done" && (
                    <p className="text-xs text-muted-foreground">
                      Concluída · {a.pontos} pontos
                    </p>
                  )}
                  {state === "locked" && (
                    <p className="text-xs text-muted-foreground">
                      Conclua a aula anterior para liberar.
                    </p>
                  )}
                </div>
                {state === "done" && (
                  <CheckCircle2 className="size-5 text-emerald-500" />
                )}
                {state === "open" && <ArrowRight className="size-5 text-primary" />}
                {state === "locked" && (
                  <Lock className="size-5 text-muted-foreground" />
                )}
              </div>
            );
            return (
              <li key={a.ordem}>
                {state === "locked" ? (
                  inner
                ) : (
                  <Link href={`/t/${a.codigo}`}>{inner}</Link>
                )}
              </li>
            );
          })}
        </ol>
      )}
    </div>
  );
}
