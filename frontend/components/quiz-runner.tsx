"use client";

import { useState } from "react";
import { apiFetch } from "@/lib/api";
import type { AttemptResult, PublicQuestao } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export interface QuizRunnerProps {
  questoes: PublicQuestao[];
  attemptId: number;
  code: string;
  publicaUrl?: string;
}

/**
 * Self-graded quiz: shows one questão at a time, collects the chosen option
 * index per questão id, and submits everything to
 * POST /api/attempts/:id/answers on the last question. Never sends
 * "correta" — the client only ever knows the chosen index.
 */
export function QuizRunner({ questoes, attemptId, code, publicaUrl }: QuizRunnerProps) {
  const [index, setIndex] = useState(0);
  const [answers, setAnswers] = useState<Record<number, number>>({});
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [result, setResult] = useState<AttemptResult | null>(null);

  const total = questoes.length;

  if (total === 0) {
    return (
      <Card>
        <CardContent className="py-8 text-center text-muted-foreground">
          Esta trilha ainda não tem quiz.
        </CardContent>
      </Card>
    );
  }

  if (result) {
    const percent = result.total > 0 ? (result.acertos / result.total) * 100 : 0;
    const shareUrl = publicaUrl ?? (typeof window !== "undefined" ? window.location.href : "");
    const whatsappHref = `https://wa.me/?text=${encodeURIComponent(
      `Concluí a trilha e fiz ${result.pontos} pontos! ${shareUrl}`
    )}`;
    const pdfHref = `${API_BASE}/api/t/${code}/export.pdf`;

    return (
      <Card>
        <CardHeader>
          <CardTitle>Quiz concluído!</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-6">
          <div className="text-center">
            <p className="text-6xl font-bold tracking-tight text-primary">{result.pontos}</p>
            <p className="text-sm text-muted-foreground">pontos</p>
          </div>

          <Progress value={percent}>
            <div className="flex w-full items-center justify-between text-sm text-muted-foreground">
              <span>
                {result.acertos} de {result.total} corretas
              </span>
              <span>{Math.round(percent)}%</span>
            </div>
          </Progress>

          <p className="text-center text-sm font-medium">
            {percent === 100
              ? "Mandou muito bem! Gabaritou o quiz. 🎉"
              : percent >= 60
                ? "Muito bom! Continue estudando para chegar aos 100%. 💪"
                : "Bom começo! Revise os tópicos e tente aprender mais. 📚"}
          </p>

          <div className="flex flex-wrap justify-center gap-2">
            <Button
              type="button"
              variant="outline"
              render={<a href={pdfHref} target="_blank" rel="noopener noreferrer" />}
            >
              Baixar PDF
            </Button>
            <Button
              type="button"
              variant="outline"
              render={<a href={whatsappHref} target="_blank" rel="noopener noreferrer" />}
            >
              Compartilhar no WhatsApp
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  const current = questoes[index];
  const isLast = index === total - 1;
  const chosen = answers[current.id];
  const progressValue = ((index + 1) / total) * 100;

  function handleSelect(optionIndex: number) {
    setAnswers((prev) => ({ ...prev, [current.id]: optionIndex }));
  }

  async function handleNext() {
    if (chosen === undefined) return;

    if (!isLast) {
      setIndex((i) => Math.min(i + 1, total - 1));
      return;
    }

    setSubmitting(true);
    setSubmitError(null);
    try {
      const payload = {
        answers: questoes
          .filter((q) => answers[q.id] !== undefined || q.id === current.id)
          .map((q) => ({
            quiz_question_id: q.id,
            escolhida: q.id === current.id ? chosen : answers[q.id],
          })),
      };
      const data = await apiFetch<AttemptResult>(`/api/attempts/${attemptId}/answers`, {
        method: "POST",
        body: JSON.stringify(payload),
      });
      setResult(data);
    } catch {
      setSubmitError("Não foi possível enviar suas respostas. Tente novamente.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <Progress value={progressValue}>
        <div className="flex w-full items-center justify-between text-sm text-muted-foreground">
          <span>
            Questão {index + 1} de {total}
          </span>
        </div>
      </Progress>

      <Card>
        <CardHeader>
          <CardTitle className="text-base font-medium leading-relaxed">
            {current.enunciado}
          </CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-2">
          {current.opcoes.map((opcao, optionIndex) => {
            const isSelected = chosen === optionIndex;
            return (
              <button
                key={optionIndex}
                type="button"
                onClick={() => handleSelect(optionIndex)}
                aria-pressed={isSelected}
                className={`flex items-center gap-3 rounded-lg border px-4 py-3 text-left text-sm transition-colors ${
                  isSelected
                    ? "border-primary bg-primary/10 font-medium"
                    : "border-border hover:bg-muted"
                }`}
              >
                <span
                  className={`flex h-5 w-5 shrink-0 items-center justify-center rounded-full border text-xs ${
                    isSelected ? "border-primary bg-primary text-primary-foreground" : "border-muted-foreground"
                  }`}
                >
                  {isSelected ? "✓" : ""}
                </span>
                {opcao}
              </button>
            );
          })}
        </CardContent>
      </Card>

      {submitError ? <p className="text-sm text-destructive">{submitError}</p> : null}

      <div className="flex items-center justify-end gap-2">
        <Button onClick={handleNext} disabled={chosen === undefined || submitting}>
          {submitting ? "Enviando..." : isLast ? "Enviar respostas" : "Próxima"}
        </Button>
      </div>
    </div>
  );
}
