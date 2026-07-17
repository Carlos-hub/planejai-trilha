"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft, Wand2 } from "lucide-react";
import { apiFetch } from "@/lib/api";
import type { Lesson } from "@/lib/types";
import { ModePicker, type LessonMode } from "@/components/mode-picker";
import { GenerateForm } from "@/components/generate-form";
import {
  LessonEditor,
  EMPTY_LESSON_CONTENT,
  type LessonContent,
} from "@/components/lesson-editor";
import { Button } from "@/components/ui/button";

export default function NewLessonPage() {
  const router = useRouter();
  const [mode, setMode] = useState<LessonMode | null>(null);
  const [content, setContent] = useState<LessonContent>(EMPTY_LESSON_CONTENT);
  const [bnccSkillId, setBnccSkillId] = useState<number | null>(null);
  const [duracao, setDuracao] = useState(50);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [needsToken, setNeedsToken] = useState(false);

  function body() {
    return JSON.stringify({
      bncc_skill_id: bnccSkillId,
      duracao,
      plano: content.plano,
      atividade: content.atividade,
      trilha: content.trilha,
    });
  }

  async function handleManualSave() {
    setSaving(true);
    setError(null);
    setNeedsToken(false);
    try {
      const lesson = await apiFetch<Lesson>("/api/lessons", {
        method: "POST",
        body: body(),
      });
      router.push(`/lessons/${lesson.id}`);
    } catch {
      setError("Não foi possível salvar a aula. Tente novamente.");
      setSaving(false);
    }
  }

  async function handleGenerate() {
    if (bnccSkillId === null) {
      setError("Selecione uma habilidade BNCC antes de gerar.");
      return;
    }
    setSaving(true);
    setError(null);
    setNeedsToken(false);
    try {
      const lesson = await apiFetch<Lesson>("/api/lessons/generate", {
        method: "POST",
        body: JSON.stringify({ bncc_skill_id: bnccSkillId, duracao }),
      });
      router.push(`/lessons/${lesson.id}`);
    } catch (err) {
      if (err instanceof Error && err.message.startsWith("API 503")) {
        setNeedsToken(true);
      } else {
        setError("Falha ao gerar a aula com IA. Tente novamente.");
      }
      setSaving(false);
    }
  }

  async function handleEnhance() {
    setSaving(true);
    setError(null);
    setNeedsToken(false);
    try {
      const lesson = await apiFetch<Lesson>("/api/lessons", {
        method: "POST",
        body: body(),
      });
      await apiFetch<Lesson>(`/api/lessons/${lesson.id}/enhance`, {
        method: "POST",
      });
      router.push(`/lessons/${lesson.id}`);
    } catch (err) {
      if (err instanceof Error && err.message.startsWith("API 503")) {
        setNeedsToken(true);
      } else {
        setError("Não foi possível aprimorar o rascunho. Tente novamente.");
      }
      setSaving(false);
    }
  }

  const backAction = (
    <Button
      type="button"
      variant="outline"
      onClick={() => setMode(null)}
      disabled={saving}
    >
      Trocar modo
    </Button>
  );

  const subtitle =
    mode === null
      ? "Como você quer criar esta aula?"
      : mode === "ia"
        ? "Gerar com IA"
        : mode === "enhance"
          ? "Aprimorar rascunho com IA"
          : "Criar manualmente";

  return (
    <div className="mx-auto flex max-w-4xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <div className="flex flex-col gap-2">
        <Link
          href="/"
          className="inline-flex w-fit items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-4" />
          Painel
        </Link>
        <div>
          <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">
            Nova aula
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">{subtitle}</p>
        </div>
      </div>

      {needsToken && (
        <p
          className="rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive"
          role="alert"
        >
          Configure seu token de IA no{" "}
          <Link href="/perfil" className="font-medium underline underline-offset-2">
            Perfil
          </Link>{" "}
          para gerar com IA.
        </p>
      )}

      {error && (
        <p
          className="rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive"
          role="alert"
        >
          {error}
        </p>
      )}

      {mode === null && <ModePicker onPick={setMode} />}

      {mode === "ia" && (
        <GenerateForm
          bnccSkillId={bnccSkillId}
          onBnccSkillIdChange={setBnccSkillId}
          duracao={duracao}
          onDuracaoChange={setDuracao}
          onGenerate={handleGenerate}
          onBack={() => setMode(null)}
          generating={saving}
        />
      )}

      {mode === "manual" && (
        <LessonEditor
          value={content}
          onChange={setContent}
          bnccSkillId={bnccSkillId}
          onBnccSkillIdChange={setBnccSkillId}
          duracao={duracao}
          onDuracaoChange={setDuracao}
          onSave={handleManualSave}
          saving={saving}
          saveLabel="Criar aula"
          extraActions={backAction}
        />
      )}

      {mode === "enhance" && (
        <>
          <div className="flex items-start gap-3 rounded-xl border border-primary/20 bg-brand-muted/50 px-4 py-3">
            <Wand2 className="mt-0.5 size-5 shrink-0 text-primary" />
            <p className="text-sm text-foreground/80">
              Escreva um rascunho — pode ser incompleto. Ao salvar, a IA revisa e
              enriquece o conteúdo, e você continua editando.
            </p>
          </div>
          <LessonEditor
            value={content}
            onChange={setContent}
            bnccSkillId={bnccSkillId}
            onBnccSkillIdChange={setBnccSkillId}
            duracao={duracao}
            onDuracaoChange={setDuracao}
            onSave={handleEnhance}
            saving={saving}
            saveLabel={saving ? "Aprimorando…" : "Salvar e aprimorar com IA"}
            extraActions={backAction}
          />
        </>
      )}
    </div>
  );
}
