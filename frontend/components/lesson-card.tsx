import Link from "next/link";
import type { LessonSummary } from "@/lib/types";
import { cn } from "@/lib/utils";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";

const ORIGEM_LABELS: Record<string, string> = {
  manual: "Manual",
  ia: "IA",
  ia_aprimorado: "IA aprimorada",
};

const STATUS_LABELS: Record<string, string> = {
  rascunho: "Rascunho",
  pronto: "Pronto",
  falha: "Falha",
};

const STATUS_CLASSES: Record<string, string> = {
  rascunho: "bg-muted text-muted-foreground",
  pronto: "bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-300",
  falha: "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300",
};

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString("pt-BR", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
    });
  } catch {
    return iso;
  }
}

export function LessonCard({ lesson }: { lesson: LessonSummary }) {
  const origemLabel = ORIGEM_LABELS[lesson.origem] ?? lesson.origem;
  const statusLabel = STATUS_LABELS[lesson.status] ?? lesson.status;
  const statusClass = STATUS_CLASSES[lesson.status] ?? "bg-muted text-muted-foreground";

  return (
    <Card>
      <CardHeader>
        <CardTitle>Aula #{lesson.id}</CardTitle>
        <CardDescription>
          {origemLabel} · {lesson.duracao} min · criada em {formatDate(lesson.created_at)}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <span
          className={cn(
            "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium",
            statusClass
          )}
        >
          {statusLabel}
        </span>
      </CardContent>
      <CardFooter className="gap-2">
        <Button
          variant="default"
          size="sm"
          render={<Link href={`/lessons/${lesson.id}`} />}
        >
          Abrir
        </Button>
        <Button
          variant="outline"
          size="sm"
          render={<Link href={`/lessons/${lesson.id}/stats`} />}
        >
          Ver turma
        </Button>
      </CardFooter>
    </Card>
  );
}
