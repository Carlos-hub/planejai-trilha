import { PenLine, Sparkles, Wand2, ArrowRight } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

export type LessonMode = "manual" | "ia" | "enhance";

interface ModeOption {
  mode: LessonMode;
  title: string;
  description: string;
  icon: LucideIcon;
  recommended?: boolean;
}

const MODE_OPTIONS: ModeOption[] = [
  {
    mode: "ia",
    title: "Gerar com IA",
    description:
      "Escolha a habilidade BNCC e a duração — a IA monta plano, trilha e quiz.",
    icon: Sparkles,
    recommended: true,
  },
  {
    mode: "enhance",
    title: "Aprimorar rascunho",
    description:
      "Escreva um esboço e deixe a IA revisar e enriquecer o conteúdo.",
    icon: Wand2,
  },
  {
    mode: "manual",
    title: "Criar manualmente",
    description:
      "Monte o plano, os tópicos e o quiz do zero, com controle total.",
    icon: PenLine,
  },
];

export function ModePicker({ onPick }: { onPick: (mode: LessonMode) => void }) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
      {MODE_OPTIONS.map(({ mode, title, description, icon: Icon, recommended }) => (
        <button
          key={mode}
          type="button"
          onClick={() => onPick(mode)}
          className={cn(
            "group relative flex flex-col gap-3 rounded-2xl border bg-card p-5 text-left transition-all",
            "hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-md hover:shadow-primary/5",
            "focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-ring/40"
          )}
        >
          {recommended && (
            <span className="absolute right-4 top-4 rounded-full bg-amber/25 px-2 py-0.5 text-[11px] font-semibold text-amber-foreground">
              Recomendado
            </span>
          )}
          <div className="flex size-11 items-center justify-center rounded-xl bg-brand-muted text-primary">
            <Icon className="size-5" />
          </div>
          <div className="flex-1">
            <h3 className="font-[family-name:var(--font-display)] text-base font-semibold tracking-tight">
              {title}
            </h3>
            <p className="mt-1 text-sm text-muted-foreground">{description}</p>
          </div>
          <span className="flex items-center gap-1 text-sm font-medium text-primary">
            Começar
            <ArrowRight className="size-4 transition-transform group-hover:translate-x-0.5" />
          </span>
        </button>
      ))}
    </div>
  );
}
