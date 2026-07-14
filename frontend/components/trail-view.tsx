"use client";

import { useState } from "react";
import type { PublicTrail } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";

export interface TrailViewProps {
  trail: PublicTrail;
  /** Called once the student advances past the last tópico into the quiz. */
  onFinishTopics: () => void;
}

/**
 * Focus-mode trail viewer: shows one tópico at a time with prev/next
 * navigation and a progress bar. Once the last tópico is passed, it calls
 * onFinishTopics — the actual quiz UI is built in Task 9.2 (QuizRunner);
 * this component only owns the topic-browsing phase.
 */
export function TrailView({ trail, onFinishTopics }: TrailViewProps) {
  const topicos = trail.topicos;
  const [index, setIndex] = useState(0);

  const total = topicos.length;
  const current = topicos[index];
  const isFirst = index === 0;
  const isLast = index === total - 1;
  const progressValue = total > 0 ? ((index + 1) / total) * 100 : 0;

  if (total === 0) {
    return (
      <Card>
        <CardContent className="py-8 text-center text-muted-foreground">
          Esta trilha ainda não tem tópicos.
          <div className="mt-4">
            <Button onClick={onFinishTopics}>Ir para o quiz</Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  function handleNext() {
    if (isLast) {
      onFinishTopics();
      return;
    }
    setIndex((i) => Math.min(i + 1, total - 1));
  }

  function handlePrev() {
    setIndex((i) => Math.max(i - 1, 0));
  }

  return (
    <div className="flex flex-col gap-4">
      <Progress value={progressValue}>
        <div className="flex w-full items-center justify-between text-sm text-muted-foreground">
          <span>
            Tópico {index + 1} de {total}
          </span>
        </div>
      </Progress>

      <Card>
        <CardHeader>
          <CardTitle>{current.titulo}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="whitespace-pre-wrap text-sm leading-relaxed text-foreground">
            {current.resumo}
          </p>
        </CardContent>
      </Card>

      <div className="flex items-center justify-between gap-2">
        <Button variant="outline" onClick={handlePrev} disabled={isFirst}>
          Anterior
        </Button>
        <Button onClick={handleNext}>{isLast ? "Ir para o quiz" : "Próximo"}</Button>
      </div>
    </div>
  );
}
