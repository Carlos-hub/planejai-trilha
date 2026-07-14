import Link from "next/link";
import { PenLine, Sparkles, Wand2, Users, ArrowRight } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { LessonSummary } from "@/lib/types";
import { cn } from "@/lib/utils";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

const ORIGEM: Record<string, { label: string; icon: LucideIcon }> = {
  manual: { label: "Manual", icon: PenLine },
  ia: { label: "Gerada com IA", icon: Sparkles },
  ia_aprimorado: { label: "Aprimorada com IA", icon: Wand2 },
};

const STATUS: Record<string, { label: string; dot: string; chip: string }> = {
  rascunho: {
    label: "Rascunho",
    dot: "bg-muted-foreground",
    chip: "bg-muted text-muted-foreground",
  },
  pronto: {
    label: "Pronta",
    dot: "bg-emerald-500",
    chip: "bg-emerald-50 text-emerald-700",
  },
  falha: {
    label: "Falha",
    dot: "bg-red-500",
    chip: "bg-red-50 text-red-700",
  },
};

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString("pt-BR", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    });
  } catch {
    return iso;
  }
}

export function LessonCard({ lesson }: { lesson: LessonSummary }) {
  const origem = ORIGEM[lesson.origem] ?? { label: lesson.origem, icon: PenLine };
  const status = STATUS[lesson.status] ?? {
    label: lesson.status,
    dot: "bg-muted-foreground",
    chip: "bg-muted text-muted-foreground",
  };
  const OrigemIcon = origem.icon;

  return (
    <Card className="group flex flex-col gap-4 p-5 transition-all hover:-translate-y-0.5 hover:border-primary/30 hover:shadow-md hover:shadow-primary/5">
      <div className="flex items-start justify-between gap-3">
        <div className="flex size-10 items-center justify-center rounded-xl bg-brand-muted text-primary">
          <OrigemIcon className="size-5" />
        </div>
        <span
          className={cn(
            "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium",
            status.chip
          )}
        >
          <span className={cn("size-1.5 rounded-full", status.dot)} />
          {status.label}
        </span>
      </div>

      <div className="min-w-0">
        <h3 className="font-[family-name:var(--font-display)] text-base font-semibold tracking-tight">
          Aula #{lesson.id}
        </h3>
        <p className="mt-0.5 text-sm text-muted-foreground">
          {origem.label} · {lesson.duracao} min
        </p>
        <p className="mt-0.5 text-xs text-muted-foreground/80">
          Criada em {formatDate(lesson.created_at)}
        </p>
      </div>

      <div className="mt-auto flex items-center gap-2 pt-1">
        <Button
          size="sm"
          className="flex-1"
          render={<Link href={`/lessons/${lesson.id}`} />}
        >
          Abrir
          <ArrowRight className="size-4" />
        </Button>
        <Button
          variant="outline"
          size="sm"
          render={<Link href={`/lessons/${lesson.id}/stats`} />}
        >
          <Users className="size-4" />
          Turma
        </Button>
      </div>
    </Card>
  );
}
