"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { apiFetch } from "@/lib/api";
import type { Lesson } from "@/lib/types";
import { LessonEditor, type LessonContent } from "@/components/lesson-editor";
import { Button } from "@/components/ui/button";

function toContent(lesson: Lesson): LessonContent {
  return { plano: lesson.plano, atividade: lesson.atividade, trilha: lesson.trilha };
}

export default function LessonDetailPage() {
  const params = useParams<{ id: string }>();
  const lessonId = params.id;

  const [lesson, setLesson] = useState<Lesson | null>(null);
  const [content, setContent] = useState<LessonContent | null>(null);
  const [bnccSkillId, setBnccSkillId] = useState<number | null>(null);
  const [duracao, setDuracao] = useState(0);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [enhancing, setEnhancing] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const load = useCallback(() => {
    apiFetch<Lesson>(`/api/lessons/${lessonId}`)
      .then((data) => {
        setLesson(data);
        setContent(toContent(data));
        setBnccSkillId(data.bncc_skill_id);
        setDuracao(data.duracao);
        setLoadError(null);
      })
      .catch(() => {
        setLoadError("Não foi possível carregar esta aula.");
      });
  }, [lessonId]);

  useEffect(() => {
    load();
  }, [load]);

  async function handleSave() {
    if (!content) return;
    setSaving(true);
    setActionError(null);
    try {
      const updated = await apiFetch<Lesson>(`/api/lessons/${lessonId}`, {
        method: "PATCH",
        body: JSON.stringify({
          bncc_skill_id: bnccSkillId,
          duracao,
          plano: content.plano,
          atividade: content.atividade,
          trilha: content.trilha,
        }),
      });
      setLesson(updated);
      setContent(toContent(updated));
      setBnccSkillId(updated.bncc_skill_id);
      setDuracao(updated.duracao);
    } catch {
      setActionError("Não foi possível salvar as alterações. Tente novamente.");
    } finally {
      setSaving(false);
    }
  }

  async function handleEnhance() {
    setEnhancing(true);
    setActionError(null);
    try {
      const updated = await apiFetch<Lesson>(`/api/lessons/${lessonId}/enhance`, {
        method: "POST",
      });
      setLesson(updated);
      setContent(toContent(updated));
      setBnccSkillId(updated.bncc_skill_id);
      setDuracao(updated.duracao);
    } catch {
      setActionError("Não foi possível aprimorar a aula com IA. Tente novamente.");
    } finally {
      setEnhancing(false);
    }
  }

  if (loadError) {
    return (
      <div className="mx-auto max-w-5xl p-4 sm:p-6">
        <p className="text-sm text-destructive" role="alert">
          {loadError}
        </p>
      </div>
    );
  }

  if (!lesson || !content) {
    return (
      <div className="mx-auto max-w-5xl p-4 sm:p-6">
        <p className="text-sm text-muted-foreground">Carregando aula...</p>
      </div>
    );
  }

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6 p-4 sm:p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">Aula #{lesson.id}</h1>
        <span className="text-sm text-muted-foreground">
          Status: {lesson.status} · Origem: {lesson.origem}
        </span>
      </div>

      {actionError && (
        <p className="text-sm text-destructive" role="alert">
          {actionError}
        </p>
      )}

      <LessonEditor
        value={content}
        onChange={setContent}
        bnccSkillId={bnccSkillId}
        onBnccSkillIdChange={setBnccSkillId}
        duracao={duracao}
        onDuracaoChange={setDuracao}
        onSave={handleSave}
        saving={saving}
        saveLabel={saving ? "Salvando..." : "Salvar"}
        extraActions={
          <Button
            type="button"
            variant="outline"
            onClick={handleEnhance}
            disabled={enhancing || saving}
          >
            {enhancing ? "Aprimorando com IA..." : "Aprimorar com IA"}
          </Button>
        }
      />
    </div>
  );
}
