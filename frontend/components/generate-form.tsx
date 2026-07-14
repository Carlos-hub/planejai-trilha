"use client";

import { Sparkles, ArrowLeft, Loader2 } from "lucide-react";
import { LessonMeta } from "@/components/lesson-meta";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";

export function GenerateForm({
  bnccSkillId,
  onBnccSkillIdChange,
  duracao,
  onDuracaoChange,
  onGenerate,
  onBack,
  generating,
}: {
  bnccSkillId: number | null;
  onBnccSkillIdChange: (id: number | null) => void;
  duracao: number;
  onDuracaoChange: (d: number) => void;
  onGenerate: () => void;
  onBack: () => void;
  generating: boolean;
}) {
  return (
    <Card className="flex flex-col gap-6 p-6">
      <div className="flex items-start gap-3">
        <div className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-brand-muted text-primary">
          <Sparkles className="size-5" />
        </div>
        <div>
          <h2 className="font-[family-name:var(--font-display)] text-lg font-semibold">
            Gerar aula com IA
          </h2>
          <p className="text-sm text-muted-foreground">
            Escolha a habilidade e a duração. A IA monta o plano, a trilha de
            tópicos e o quiz — você revisa e edita tudo depois.
          </p>
        </div>
      </div>

      <LessonMeta
        bnccSkillId={bnccSkillId}
        onBnccSkillIdChange={onBnccSkillIdChange}
        duracao={duracao}
        onDuracaoChange={onDuracaoChange}
      />

      {generating && (
        <p className="flex items-center gap-2 rounded-lg bg-brand-muted/60 px-4 py-3 text-sm text-primary">
          <Loader2 className="size-4 animate-spin" />
          Gerando o plano, a trilha e o quiz… isso pode levar alguns segundos.
        </p>
      )}

      <div className="flex items-center justify-between gap-2 border-t pt-4">
        <Button
          type="button"
          variant="ghost"
          onClick={onBack}
          disabled={generating}
        >
          <ArrowLeft className="size-4" />
          Voltar
        </Button>
        <Button
          type="button"
          size="lg"
          onClick={onGenerate}
          disabled={generating || bnccSkillId === null}
        >
          {generating ? (
            <>
              <Loader2 className="size-4 animate-spin" />
              Gerando…
            </>
          ) : (
            <>
              <Sparkles className="size-4" />
              Gerar aula
            </>
          )}
        </Button>
      </div>
    </Card>
  );
}
