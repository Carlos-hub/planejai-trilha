"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { apiFetch } from "@/lib/api";
import type { Lesson, PublishResult } from "@/lib/types";
import { LessonEditor, type LessonContent } from "@/components/lesson-editor";
import { Button } from "@/components/ui/button";
import { ShareLinks } from "@/components/share-links";

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
  const [publishing, setPublishing] = useState(false);
  const [publishResult, setPublishResult] = useState<PublishResult | null>(null);
  const [publishError, setPublishError] = useState<string | null>(null);

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

  async function handlePublish() {
    setPublishing(true);
    setPublishError(null);
    try {
      const result = await apiFetch<PublishResult>(`/api/trails/${lessonId}/publish`, {
        method: "POST",
      });
      setPublishResult(result);
    } catch {
      setPublishError("Não foi possível publicar a trilha. Tente novamente.");
    } finally {
      setPublishing(false);
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

  const STATUS_LABELS: Record<string, string> = {
    rascunho: "Rascunho",
    pronto: "Pronta",
    falha: "Falha",
  };
  const ORIGEM_LABELS: Record<string, string> = {
    manual: "Manual",
    ia: "Gerada com IA",
    ia_aprimorado: "Aprimorada com IA",
  };

  return (
    <div className="mx-auto flex max-w-4xl flex-col gap-6 px-4 py-6 pb-28 sm:px-6 sm:py-8">
      <div className="flex flex-col gap-2">
        <Link
          href="/"
          className="inline-flex w-fit items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-4" />
          Painel
        </Link>
        <div className="flex flex-wrap items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">
            Aula #{lesson.id}
          </h1>
          <span className="inline-flex items-center gap-1.5 rounded-full bg-secondary px-2.5 py-1 text-xs font-medium text-secondary-foreground">
            {STATUS_LABELS[lesson.status] ?? lesson.status}
          </span>
          <span className="text-xs text-muted-foreground">
            {ORIGEM_LABELS[lesson.origem] ?? lesson.origem}
          </span>
        </div>
      </div>

      {actionError && (
        <p
          className="rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive"
          role="alert"
        >
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

      <div className="flex flex-col gap-3 rounded-xl border p-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div>
            <h2 className="text-lg font-semibold">Publicar trilha</h2>
            {lesson.status !== "pronto" && !publishResult && (
              <p className="text-sm text-muted-foreground">
                Salve o conteúdo da aula e deixe o status como &quot;Pronto&quot; antes de publicar.
              </p>
            )}
          </div>
          <div className="flex gap-2">
            <Button
              type="button"
              onClick={handlePublish}
              disabled={publishing || lesson.status !== "pronto"}
            >
              {publishing ? "Publicando..." : "Publicar trilha"}
            </Button>
            <Button type="button" variant="outline" render={<Link href={`/lessons/${lesson.id}/stats`} />}>
              Ver turma
            </Button>
          </div>
        </div>

        {publishError && (
          <p className="text-sm text-destructive" role="alert">
            {publishError}
          </p>
        )}

        {publishResult && (
          <ShareLinks codigo={publishResult.codigo} publicaUrl={publishResult.publica_url} />
        )}
      </div>
    </div>
  );
}
