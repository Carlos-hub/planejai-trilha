"use client";

import { useEffect, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Sparkles, Route, Users } from "lucide-react";
import { apiFetch } from "@/lib/api";
import type { Me } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { BrandLockup, BrandMark } from "@/components/brand-mark";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [senha, setSenha] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [justRegistered, setJustRegistered] = useState(false);

  // Show a success banner when arriving from signup (?cadastro=1). Read from
  // the URL client-side to avoid a Suspense boundary around useSearchParams.
  // One-shot read on mount (starts false to keep SSR/hydration in sync), so the
  // set-state-in-effect rule doesn't apply here.
  useEffect(() => {
    if (new URLSearchParams(window.location.search).get("cadastro") === "1") {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setJustRegistered(true);
    }
  }, []);

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      await apiFetch<Me>("/api/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, senha }),
      });
      router.push("/");
    } catch {
      setError("E-mail ou senha incorretos. Tente novamente.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="grid min-h-screen lg:grid-cols-2">
      {/* Brand panel */}
      <div className="relative hidden flex-col justify-between overflow-hidden bg-sidebar p-10 text-sidebar-foreground lg:flex">
        <div className="absolute -right-24 -top-24 size-80 rounded-full bg-sidebar-primary/20 blur-3xl" />
        <div className="absolute -bottom-24 -left-16 size-72 rounded-full bg-white/5 blur-3xl" />
        <div className="relative text-sidebar-accent-foreground">
          <BrandLockup />
        </div>
        <div className="relative">
          <BrandMark className="size-10 text-sidebar-primary" />
          <h2 className="mt-5 max-w-sm font-[family-name:var(--font-display)] text-3xl font-semibold leading-tight text-sidebar-accent-foreground">
            Planeje uma vez. Ensine o aluno inteiro.
          </h2>
          <p className="mt-3 max-w-sm text-sm text-sidebar-foreground/70">
            Aulas alinhadas à BNCC que viram trilhas de estudo com quiz
            autocorrigido — sem trabalho duplicado.
          </p>
          <ul className="mt-8 flex flex-col gap-3 text-sm text-sidebar-foreground/80">
            <Feature icon={Sparkles} text="Plano de aula gerado com IA" />
            <Feature icon={Route} text="Trilha do aluno em um clique" />
            <Feature icon={Users} text="Acompanhe a turma em tempo real" />
          </ul>
        </div>
        <p className="relative text-xs text-sidebar-foreground/50">
          PlanejAI · Trilha — Hackathon 6FSDT
        </p>
      </div>

      {/* Form panel */}
      <div className="flex items-center justify-center p-6">
        <div className="w-full max-w-sm">
          <div className="mb-8 lg:hidden">
            <BrandLockup markClassName="text-primary" />
          </div>

          <h1 className="text-2xl font-semibold tracking-tight">
            Entrar
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Acesse sua conta de professor.
          </p>

          {justRegistered && (
            <p
              className="mt-6 rounded-lg border border-primary/30 bg-primary/5 px-3 py-2 text-sm text-primary"
              role="status"
            >
              Conta criada! Faça login para começar.
            </p>
          )}

          <form onSubmit={handleSubmit} className="mt-8 flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="email">E-mail</Label>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                placeholder="voce@escola.edu.br"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="senha">Senha</Label>
              <Input
                id="senha"
                type="password"
                autoComplete="current-password"
                placeholder="••••••••"
                required
                value={senha}
                onChange={(e) => setSenha(e.target.value)}
              />
            </div>

            {error && (
              <p
                className="rounded-lg border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive"
                role="alert"
              >
                {error}
              </p>
            )}

            <Button
              type="submit"
              size="lg"
              className="mt-1 w-full"
              disabled={submitting}
            >
              {submitting ? "Entrando…" : "Entrar"}
            </Button>
          </form>

          <p className="mt-6 text-center text-sm text-muted-foreground">
            Não tem conta?{" "}
            <Link
              href="/cadastro"
              className="font-medium text-primary underline underline-offset-2"
            >
              Criar conta
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}

function Feature({
  icon: Icon,
  text,
}: {
  icon: typeof Sparkles;
  text: string;
}) {
  return (
    <li className="flex items-center gap-3">
      <span className="flex size-8 items-center justify-center rounded-lg bg-sidebar-accent/60 text-sidebar-primary">
        <Icon className="size-4" />
      </span>
      {text}
    </li>
  );
}
