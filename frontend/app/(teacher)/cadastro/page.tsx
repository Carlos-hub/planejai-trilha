"use client";

import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Sparkles, Route, Users } from "lucide-react";
import { register } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { BrandLockup, BrandMark } from "@/components/brand-mark";

export default function CadastroPage() {
  const router = useRouter();
  const [nome, setNome] = useState("");
  const [email, setEmail] = useState("");
  const [senha, setSenha] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      await register(nome, email, senha);
      router.push("/login?cadastro=1");
    } catch (err) {
      if (err instanceof Error && err.message.startsWith("API 409")) {
        setError("Este e-mail já está cadastrado. Tente entrar.");
      } else {
        setError("Não foi possível criar a conta. Tente novamente.");
      }
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
            Comece a planejar em minutos.
          </h2>
          <p className="mt-3 max-w-sm text-sm text-sidebar-foreground/70">
            Crie sua conta de professor e transforme aulas da BNCC em trilhas de
            estudo com quiz autocorrigido.
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

          <h1 className="text-2xl font-semibold tracking-tight">Criar conta</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Cadastre-se como professor.
          </p>

          <form onSubmit={handleSubmit} className="mt-8 flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="nome">Nome</Label>
              <Input
                id="nome"
                autoComplete="name"
                placeholder="Seu nome"
                required
                value={nome}
                onChange={(e) => setNome(e.target.value)}
              />
            </div>
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
                autoComplete="new-password"
                placeholder="••••••••"
                required
                minLength={8}
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
              {submitting ? "Criando…" : "Criar conta"}
            </Button>
          </form>

          <p className="mt-6 text-center text-sm text-muted-foreground">
            Já tem conta?{" "}
            <Link
              href="/login"
              className="font-medium text-primary underline underline-offset-2"
            >
              Entrar
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
    <li className="flex items-center gap-2.5">
      <span className="flex size-7 items-center justify-center rounded-md bg-sidebar-primary/15 text-sidebar-primary">
        <Icon className="size-4" />
      </span>
      {text}
    </li>
  );
}
