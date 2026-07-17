"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { Plus, Users } from "lucide-react";
import { listTurmas, createTurma } from "@/lib/api";
import type { Turma } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";

export default function TurmasPage() {
  const [turmas, setTurmas] = useState<Turma[] | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [nome, setNome] = useState("");
  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);

  const reload = useCallback(() => {
    listTurmas()
      .then((data) => {
        setTurmas(data);
        setLoadError(null);
      })
      .catch(() => {
        setLoadError("Não foi possível carregar as turmas.");
      });
  }, []);

  useEffect(() => {
    reload();
  }, [reload]);

  async function onCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!nome.trim()) return;
    setCreating(true);
    setCreateError(null);
    try {
      await createTurma({ nome: nome.trim() });
      setNome("");
      reload();
    } catch {
      setCreateError("Não foi possível criar a turma. Tente novamente.");
    } finally {
      setCreating(false);
    }
  }

  const loading = turmas === null && !loadError;

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-8 px-4 py-6 sm:px-6 sm:py-8">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">Turmas</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Gerencie suas turmas e importe alunos por planilha.
        </p>
      </header>

      <Card>
        <CardContent>
          <form onSubmit={onCreate} className="flex flex-col gap-3 sm:flex-row sm:items-end">
            <div className="flex-1">
              <label htmlFor="nome-turma" className="mb-1 block text-sm font-medium">
                Nome da turma
              </label>
              <Input
                id="nome-turma"
                value={nome}
                onChange={(e) => setNome(e.target.value)}
                placeholder="Ex: 6º ano A"
              />
            </div>
            <Button type="submit" disabled={creating || !nome.trim()}>
              <Plus className="size-4" />
              {creating ? "Criando..." : "Criar turma"}
            </Button>
          </form>
          {createError && (
            <p className="mt-2 text-sm text-destructive" role="alert">
              {createError}
            </p>
          )}
        </CardContent>
      </Card>

      <section className="flex flex-col gap-4">
        {loading && (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {[0, 1, 2].map((i) => (
              <div key={i} className="h-24 animate-pulse rounded-xl border bg-muted/40" />
            ))}
          </div>
        )}

        {loadError && (
          <p
            className="rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive"
            role="alert"
          >
            {loadError}
          </p>
        )}

        {turmas && turmas.length === 0 && (
          <div className="flex flex-col items-center gap-4 rounded-2xl border border-dashed bg-card/50 px-6 py-16 text-center">
            <div className="flex size-12 items-center justify-center rounded-2xl bg-brand-muted text-primary">
              <Users className="size-6" />
            </div>
            <div>
              <p className="font-medium">Nenhuma turma ainda</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Crie sua primeira turma para importar alunos.
              </p>
            </div>
          </div>
        )}

        {turmas && turmas.length > 0 && (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {turmas.map((t) => (
              <Link key={t.id} href={`/turmas/${t.id}`}>
                <Card className="transition-colors hover:bg-muted/40">
                  <CardContent className="flex items-center gap-3">
                    <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-brand-muted text-primary">
                      <Users className="size-4" />
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-medium">{t.nome}</p>
                      {t.etapa && (
                        <p className="truncate text-xs text-muted-foreground">{t.etapa}</p>
                      )}
                    </div>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
