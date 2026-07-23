"use client";

import { useState } from "react";
import { ChevronUp, ChevronDown, Trash2, Plus, Lock } from "lucide-react";
import type { TurmaAula, LessonSummary } from "@/lib/types";
import {
  listLessons,
  attachTurmaLesson,
  detachTurmaLesson,
  reorderTurmaLessons,
} from "@/lib/api";
import { Button } from "@/components/ui/button";

export function TurmaAulas({
  turmaId,
  aulas,
  onChanged,
}: {
  turmaId: number;
  aulas: TurmaAula[];
  onChanged: () => void;
}) {
  const [picking, setPicking] = useState(false);
  const [options, setOptions] = useState<LessonSummary[] | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function openPicker() {
    setError(null);
    setPicking(true);
    try {
      const all = await listLessons();
      const inTurma = new Set(aulas.map((a) => a.lesson_plan_id));
      setOptions(all.filter((l) => l.status === "pronto" && !inTurma.has(l.id)));
    } catch {
      setError("Não foi possível carregar suas aulas.");
      setPicking(false);
    }
  }

  async function add(lessonId: number) {
    setBusy(true);
    setError(null);
    try {
      await attachTurmaLesson(turmaId, lessonId);
      setPicking(false);
      setOptions(null);
      onChanged();
    } catch {
      setError("Não foi possível adicionar a aula.");
    } finally {
      setBusy(false);
    }
  }

  async function remove(lessonId: number) {
    setBusy(true);
    setError(null);
    try {
      await detachTurmaLesson(turmaId, lessonId);
      onChanged();
    } catch {
      setError("Não foi possível remover a aula.");
    } finally {
      setBusy(false);
    }
  }

  async function move(index: number, dir: -1 | 1) {
    const next = index + dir;
    if (next < 0 || next >= aulas.length) return;
    const ids = aulas.map((a) => a.lesson_plan_id);
    [ids[index], ids[next]] = [ids[next], ids[index]];
    setBusy(true);
    setError(null);
    try {
      await reorderTurmaLessons(turmaId, ids);
      onChanged();
    } catch {
      setError("Não foi possível reordenar as aulas.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">Aulas da turma</h2>
          <p className="text-sm text-muted-foreground">
            O aluno percorre as aulas em ordem — a próxima libera ao concluir a anterior.
          </p>
        </div>
        <Button type="button" size="sm" onClick={openPicker} disabled={busy}>
          <Plus className="size-4" />
          Adicionar aula
        </Button>
      </div>

      {error && (
        <p className="text-sm text-destructive" role="alert">
          {error}
        </p>
      )}

      {aulas.length === 0 ? (
        <p className="rounded-lg border border-dashed bg-muted/30 px-4 py-6 text-center text-sm text-muted-foreground">
          Nenhuma aula ainda. Adicione a primeira aula da sequência.
        </p>
      ) : (
        <ol className="flex flex-col gap-2">
          {aulas.map((a, i) => (
            <li
              key={a.lesson_plan_id}
              className="flex items-center gap-3 rounded-xl border bg-card p-3"
            >
              <span className="flex size-7 shrink-0 items-center justify-center rounded-full bg-primary/10 text-sm font-semibold text-primary">
                {a.ordem}
              </span>
              <span className="flex-1 truncate text-sm font-medium">{a.label}</span>
              <div className="flex items-center gap-1">
                <button
                  type="button"
                  aria-label="Mover para cima"
                  disabled={busy || i === 0}
                  onClick={() => move(i, -1)}
                  className="flex size-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted disabled:opacity-30"
                >
                  <ChevronUp className="size-4" />
                </button>
                <button
                  type="button"
                  aria-label="Mover para baixo"
                  disabled={busy || i === aulas.length - 1}
                  onClick={() => move(i, 1)}
                  className="flex size-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted disabled:opacity-30"
                >
                  <ChevronDown className="size-4" />
                </button>
                <button
                  type="button"
                  aria-label="Remover aula"
                  disabled={busy}
                  onClick={() => remove(a.lesson_plan_id)}
                  className="flex size-8 items-center justify-center rounded-md text-muted-foreground hover:bg-destructive/10 hover:text-destructive disabled:opacity-30"
                >
                  <Trash2 className="size-4" />
                </button>
              </div>
            </li>
          ))}
        </ol>
      )}

      {picking && (
        <div className="rounded-xl border bg-card p-3">
          <div className="mb-2 flex items-center justify-between">
            <span className="text-sm font-medium">Escolha uma aula pronta</span>
            <Button type="button" variant="ghost" size="sm" onClick={() => setPicking(false)}>
              Fechar
            </Button>
          </div>
          {options === null ? (
            <p className="text-sm text-muted-foreground">Carregando…</p>
          ) : options.length === 0 ? (
            <p className="flex items-center gap-2 text-sm text-muted-foreground">
              <Lock className="size-4" />
              Nenhuma aula pronta disponível. Crie e finalize uma aula primeiro.
            </p>
          ) : (
            <ul className="flex flex-col gap-1">
              {options.map((l) => (
                <li key={l.id}>
                  <button
                    type="button"
                    disabled={busy}
                    onClick={() => add(l.id)}
                    className="flex w-full items-center justify-between rounded-lg px-3 py-2 text-left text-sm hover:bg-muted disabled:opacity-50"
                  >
                    <span>Aula #{l.id}</span>
                    <span className="text-xs text-muted-foreground">
                      {l.duracao} min · {l.origem}
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </section>
  );
}
