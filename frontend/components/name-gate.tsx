"use client";

import { useState, type FormEvent } from "react";
import { apiFetch } from "@/lib/api";
import type { AttemptStart } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

export interface NameGateProps {
  code: string;
  onStarted: (attemptId: number, nome: string) => void;
}

export function attemptStorageKey(code: string) {
  return `trilha:${code}:attempt`;
}

export function NameGate({ code, onStarted }: NameGateProps) {
  const [nome, setNome] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmed = nome.trim();
    if (!trimmed) {
      setError("Digite seu nome para começar.");
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      const result = await apiFetch<AttemptStart>(`/api/t/${code}/attempt`, {
        method: "POST",
        body: JSON.stringify({ nome: trimmed }),
      });
      try {
        window.localStorage.setItem(
          attemptStorageKey(code),
          JSON.stringify({ attemptId: result.attempt_id, nome: trimmed })
        );
      } catch {
        // localStorage unavailable (private mode, etc.) — not fatal.
      }
      onStarted(result.attempt_id, trimmed);
    } catch {
      setError("Não foi possível iniciar a trilha. Tente novamente.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Card className="mx-auto w-full max-w-sm">
      <CardHeader>
        <CardTitle>Bem-vindo(a)!</CardTitle>
        <CardDescription>Digite seu nome para começar a trilha de estudos.</CardDescription>
      </CardHeader>
      <CardContent>
        <form className="flex flex-col gap-3" onSubmit={handleSubmit}>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="nome">Seu nome</Label>
            <Input
              id="nome"
              value={nome}
              onChange={(event) => setNome(event.target.value)}
              placeholder="Ex: Maria Silva"
              autoFocus
              disabled={submitting}
            />
          </div>
          {error ? <p className="text-sm text-destructive">{error}</p> : null}
          <Button type="submit" disabled={submitting} className="w-full">
            {submitting ? "Iniciando..." : "Começar"}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
