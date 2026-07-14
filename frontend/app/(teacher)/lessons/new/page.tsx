"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { apiFetch } from "@/lib/api";
import type { Lesson } from "@/lib/types";
import { ModePicker, type LessonMode } from "@/components/mode-picker";
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

  async function handleManualSave() {
    setSaving(true);
    setError(null);
    try {
      const lesson = await apiFetch<Lesson>("/api/lessons", {
        method: "POST",
        body: JSON.stringify({
          bncc_skill_id: bnccSkillId,
          duracao,
          plano: content.plano,
          atividade: content.atividade,
          trilha: content.trilha,
        }),
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
    try {
      const lesson = await apiFetch<Lesson>("/api/lessons/generate", {
        method: "POST",
        body: JSON.stringify({ bncc_skill_id: bnccSkillId, duracao }),
      });
      router.push(`/lessons/${lesson.id}`);
    } catch {
      setError("Falha ao gerar a aula com IA. Tente novamente.");
      setSaving(false);
    }
  }

  async function handleEnhance() {
    setSaving(true);
    setError(null);
    try {
      const lesson = await apiFetch<Lesson>("/api/lessons", {
        method: "POST",
        body: JSON.stringify({
          bncc_skill_id: bnccSkillId,
          duracao,
          plano: content.plano,
          atividade: content.atividade,
          trilha: content.trilha,
        }),
      });
      await apiFetch<Lesson>(`/api/lessons/${lesson.id}/enhance`, { method: "POST" });
      router.push(`/lessons/${lesson.id}`);
    } catch {
      setError("Não foi possível aprimorar o rascunho. Tente novamente.");
      setSaving(false);
    }
  }

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6 p-4 sm:p-6">
      <h1 className="text-2xl font-semibold tracking-tight">Nova aula</h1>

      {error && (
        <p className="text-sm text-destructive" role="alert">
          {error}
        </p>
      )}

      {mode === null && <ModePicker onPick={setMode} />}

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
          extraActions={
            <Button type="button" variant="outline" onClick={() => setMode(null)} disabled={saving}>
              Voltar
            </Button>
          }
        />
      )}

      {mode === "ia" && (
        <div className="flex flex-col gap-4 rounded-xl border p-4 sm:p-6">
          <p className="text-sm text-muted-foreground">
            Escolha a habilidade BNCC e a duração da aula. A IA vai gerar o plano completo,
            a trilha de tópicos e o quiz — tudo editável depois.
          </p>
          <LessonEditor
            value={content}
            onChange={setContent}
            bnccSkillId={bnccSkillId}
            onBnccSkillIdChange={setBnccSkillId}
            duracao={duracao}
            onDuracaoChange={setDuracao}
            onSave={handleGenerate}
            saving={saving}
            saveLabel={saving ? "Gerando com IA..." : "Gerar com IA"}
            extraActions={
              <Button type="button" variant="outline" onClick={() => setMode(null)} disabled={saving}>
                Voltar
              </Button>
            }
          />
        </div>
      )}

      {mode === "enhance" && (
        <div className="flex flex-col gap-4">
          <p className="text-sm text-muted-foreground">
            Preencha o rascunho da aula. Ao salvar, a IA vai revisar e aprimorar o conteúdo
            automaticamente — tudo continuará editável.
          </p>
          <LessonEditor
            value={content}
            onChange={setContent}
            bnccSkillId={bnccSkillId}
            onBnccSkillIdChange={setBnccSkillId}
            duracao={duracao}
            onDuracaoChange={setDuracao}
            onSave={handleEnhance}
            saving={saving}
            saveLabel={saving ? "Aprimorando com IA..." : "Salvar e aprimorar com IA"}
            extraActions={
              <Button type="button" variant="outline" onClick={() => setMode(null)} disabled={saving}>
                Voltar
              </Button>
            }
          />
        </div>
      )}
    </div>
  );
}
