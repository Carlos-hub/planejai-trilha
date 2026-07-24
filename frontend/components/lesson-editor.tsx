"use client";

import {
  Target,
  Route,
  ListChecks,
  Plus,
  Trash2,
  Check,
  Footprints,
  BookOpen,
  ClipboardCheck,
  Wrench,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { Plano, Questao, Topico, Trilha } from "@/lib/types";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card } from "@/components/ui/card";
import { LessonMeta } from "@/components/lesson-meta";

export interface LessonContent {
  plano: Plano;
  atividade: string;
  trilha: Trilha;
}

export const EMPTY_LESSON_CONTENT: LessonContent = {
  plano: { objetivos: "", metodologia: "", recursos: "", avaliacao: "" },
  atividade: "",
  trilha: { topicos: [], quiz: { questoes: [] } },
};

const emptyTopico = (): Topico => ({ titulo: "", resumo: "" });
const emptyQuestao = (): Questao => ({ enunciado: "", opcoes: ["", ""], correta: 0 });

type PlanoStep = {
  /** stable id, used as React key and in tests */
  id: string;
  /** where the value lives: the plano object or the top-level atividade field */
  source: "plano" | "atividade";
  planoKey?: keyof Plano;
  icon: LucideIcon;
  label: string;
  /** short chip shown next to the label, e.g. the amber "fecha a aula" */
  badge?: string;
  hint: string;
  example: string;
  placeholder: string;
  /** brand steps are the teacher's planning; the amber step is student-facing */
  tone: "brand" | "amber";
};

/**
 * The Plano de aula as a sequential, numbered flow. Each step tells the teacher
 * WHAT to write and offers a real example they can drop in. The last step
 * (Atividade) lives on `content.atividade`, not inside `plano`, and is the only
 * part the student sees — hence the distinct amber tone.
 */
export const PLANO_STEPS: PlanoStep[] = [
  {
    id: "objetivos",
    source: "plano",
    planoKey: "objetivos",
    icon: Target,
    label: "Objetivos",
    hint: "O que o aluno deve saber ou ser capaz de fazer ao terminar a aula.",
    example:
      "Ao fim da aula, o aluno deve identificar as causas da Revolução Industrial e relacioná-las com mudanças sociais.",
    placeholder: "Escreva o objetivo da aula…",
    tone: "brand",
  },
  {
    id: "metodologia",
    source: "plano",
    planoKey: "metodologia",
    icon: Footprints,
    label: "Metodologia",
    hint: "Como você vai conduzir a aula, passo a passo, com o tempo de cada momento.",
    example:
      "Abertura com pergunta disparadora (10 min), exposição dialogada (20 min), atividade em duplas (15 min), fechamento (5 min).",
    placeholder: "Descreva como a aula acontece, do início ao fim…",
    tone: "brand",
  },
  {
    id: "recursos",
    source: "plano",
    planoKey: "recursos",
    icon: BookOpen,
    label: "Recursos",
    hint: "Materiais, ferramentas e referências que você vai usar.",
    example:
      "Projetor, mapa histórico impresso, capítulo 4 do livro, vídeo de 5 min sobre a linha de produção.",
    placeholder: "Liste os materiais e referências…",
    tone: "brand",
  },
  {
    id: "avaliacao",
    source: "plano",
    planoKey: "avaliacao",
    icon: ClipboardCheck,
    label: "Avaliação",
    hint: "Como você vai verificar se os objetivos foram atingidos.",
    example:
      "Participação na discussão inicial + desempenho no quiz ao final da trilha.",
    placeholder: "Como o aprendizado será avaliado…",
    tone: "brand",
  },
  {
    id: "atividade",
    source: "atividade",
    icon: Wrench,
    label: "Atividade prática",
    badge: "fecha a aula",
    hint: "Uma tarefa mão-na-massa para o aluno fixar o conteúdo.",
    example:
      "Em duplas, montar uma linha do tempo com 5 invenções da Revolução Industrial e seu impacto no trabalho.",
    placeholder: "Descreva a atividade prática…",
    tone: "amber",
  },
];

const LETTERS = "ABCDEFGH";

interface LessonEditorProps {
  value: LessonContent;
  onChange: (value: LessonContent) => void;
  bnccSkillId: number | null;
  onBnccSkillIdChange: (id: number | null) => void;
  duracao: number;
  onDuracaoChange: (duracao: number) => void;
  onSave: () => void;
  saving?: boolean;
  saveLabel?: string;
  extraActions?: React.ReactNode;
}

export function LessonEditor({
  value,
  onChange,
  bnccSkillId,
  onBnccSkillIdChange,
  duracao,
  onDuracaoChange,
  onSave,
  saving = false,
  saveLabel = "Salvar",
  extraActions,
}: LessonEditorProps) {
  const { topicos } = value.trilha;
  const { questoes } = value.trilha.quiz;

  function updatePlano(field: keyof Plano, text: string) {
    onChange({ ...value, plano: { ...value.plano, [field]: text } });
  }
  function stepValue(step: PlanoStep): string {
    return step.source === "plano" ? value.plano[step.planoKey!] : value.atividade;
  }
  function setStepValue(step: PlanoStep, text: string) {
    if (step.source === "plano") updatePlano(step.planoKey!, text);
    else onChange({ ...value, atividade: text });
  }
  function fillExample(step: PlanoStep) {
    // Only seed an empty field — never clobber what the teacher already wrote.
    if (stepValue(step).trim()) return;
    setStepValue(step, step.example);
  }
  function updateTopico(index: number, field: keyof Topico, text: string) {
    onChange({
      ...value,
      trilha: {
        ...value.trilha,
        topicos: topicos.map((t, i) => (i === index ? { ...t, [field]: text } : t)),
      },
    });
  }
  function addTopico() {
    onChange({
      ...value,
      trilha: { ...value.trilha, topicos: [...topicos, emptyTopico()] },
    });
  }
  function removeTopico(index: number) {
    onChange({
      ...value,
      trilha: { ...value.trilha, topicos: topicos.filter((_, i) => i !== index) },
    });
  }
  function updateQuestao(index: number, patch: Partial<Questao>) {
    onChange({
      ...value,
      trilha: {
        ...value.trilha,
        quiz: {
          questoes: questoes.map((q, i) => (i === index ? { ...q, ...patch } : q)),
        },
      },
    });
  }
  function addQuestao() {
    onChange({
      ...value,
      trilha: { ...value.trilha, quiz: { questoes: [...questoes, emptyQuestao()] } },
    });
  }
  function removeQuestao(index: number) {
    onChange({
      ...value,
      trilha: {
        ...value.trilha,
        quiz: { questoes: questoes.filter((_, i) => i !== index) },
      },
    });
  }
  function updateOpcao(qIndex: number, oIndex: number, text: string) {
    updateQuestao(qIndex, {
      opcoes: questoes[qIndex].opcoes.map((o, i) => (i === oIndex ? text : o)),
    });
  }
  function addOpcao(qIndex: number) {
    updateQuestao(qIndex, { opcoes: [...questoes[qIndex].opcoes, ""] });
  }
  function removeOpcao(qIndex: number, oIndex: number) {
    const opcoes = questoes[qIndex].opcoes.filter((_, i) => i !== oIndex);
    const correta = questoes[qIndex].correta >= opcoes.length ? 0 : questoes[qIndex].correta;
    updateQuestao(qIndex, { opcoes, correta });
  }

  return (
    <div className="flex flex-col gap-6 pb-24">
      <Section
        icon={<Target className="size-4" />}
        title="Sobre a aula"
        description="Vincule uma habilidade da BNCC e defina a duração."
      >
        <LessonMeta
          bnccSkillId={bnccSkillId}
          onBnccSkillIdChange={onBnccSkillIdChange}
          duracao={duracao}
          onDuracaoChange={onDuracaoChange}
        />
      </Section>

      <Section
        icon={<ListChecks className="size-4" />}
        title="Plano de aula"
        description="O planejamento do professor — não aparece para o aluno. Preencha na ordem."
      >
        <PlanoProgress
          filled={PLANO_STEPS.filter((s) => stepValue(s).trim().length > 0).length}
          total={PLANO_STEPS.length}
        />
        <ol className="relative mt-5 flex flex-col">
          {PLANO_STEPS.map((step, index) => {
            const text = stepValue(step);
            const filled = text.trim().length > 0;
            const last = index === PLANO_STEPS.length - 1;
            const amber = step.tone === "amber";
            const Icon = step.icon;
            return (
              <li
                key={step.id}
                data-step={step.id}
                data-filled={filled}
                className="relative flex gap-4 pb-5 last:pb-0"
              >
                {!last && (
                  <span
                    aria-hidden
                    className="absolute bottom-1 left-[17px] top-9 w-0.5 bg-border"
                  />
                )}
                <StepCircle n={index + 1} filled={filled} amber={amber} />
                <div className="flex flex-1 flex-col gap-1.5">
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="flex items-center gap-1.5 text-sm font-semibold">
                      <Icon
                        className={cn(
                          "size-4",
                          amber ? "text-amber-foreground" : "text-primary"
                        )}
                      />
                      {step.label}
                    </h3>
                    {step.badge && (
                      <span
                        className={cn(
                          "rounded-full px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide",
                          amber
                            ? "bg-amber/20 text-amber-foreground"
                            : "bg-muted text-muted-foreground"
                        )}
                      >
                        {step.badge}
                      </span>
                    )}
                  </div>
                  <p className="text-sm text-muted-foreground">{step.hint}</p>
                  <button
                    type="button"
                    onClick={() => fillExample(step)}
                    disabled={filled}
                    title={filled ? undefined : "Usar como ponto de partida"}
                    className="flex items-start gap-1.5 rounded-lg border border-dashed bg-muted/50 px-2.5 py-1.5 text-left text-xs text-muted-foreground transition-colors hover:enabled:border-primary/40 hover:enabled:bg-muted disabled:opacity-60"
                  >
                    <span
                      className={cn(
                        "shrink-0 font-semibold",
                        amber ? "text-amber-foreground" : "text-primary"
                      )}
                    >
                      Ex:
                    </span>
                    <span>{step.example}</span>
                  </button>
                  <Textarea
                    rows={3}
                    placeholder={step.placeholder}
                    value={text}
                    onChange={(e) => setStepValue(step, e.target.value)}
                    className={cn(
                      filled && "border-emerald-300 bg-emerald-50/40"
                    )}
                  />
                </div>
              </li>
            );
          })}
        </ol>
      </Section>

      <Section
        icon={<Route className="size-4" />}
        title="Trilha do aluno"
        description="Os tópicos que o aluno percorre, um de cada vez."
        action={
          <Button type="button" variant="outline" size="sm" onClick={addTopico}>
            <Plus className="size-4" />
            Tópico
          </Button>
        }
      >
        {topicos.length === 0 ? (
          <EmptyHint text="Nenhum tópico ainda. Adicione o primeiro passo da trilha." />
        ) : (
          <div className="flex flex-col gap-3">
            {topicos.map((topico, index) => (
              <div
                key={index}
                className="flex gap-3 rounded-xl border bg-card p-3"
              >
                <StepBadge n={index + 1} />
                <div className="flex flex-1 flex-col gap-2">
                  <Input
                    placeholder="Título do tópico"
                    value={topico.titulo}
                    onChange={(e) => updateTopico(index, "titulo", e.target.value)}
                  />
                  <Textarea
                    rows={2}
                    placeholder="Resumo — o que o aluno lê neste tópico"
                    value={topico.resumo}
                    onChange={(e) => updateTopico(index, "resumo", e.target.value)}
                  />
                </div>
                <IconButton label="Remover tópico" onClick={() => removeTopico(index)} />
              </div>
            ))}
          </div>
        )}
      </Section>

      <Section
        icon={<ListChecks className="size-4" />}
        title="Quiz"
        description="Marque a alternativa correta de cada questão."
        action={
          <Button type="button" variant="outline" size="sm" onClick={addQuestao}>
            <Plus className="size-4" />
            Questão
          </Button>
        }
      >
        {questoes.length === 0 ? (
          <EmptyHint text="Nenhuma questão ainda. Adicione a primeira pergunta do quiz." />
        ) : (
          <div className="flex flex-col gap-4">
            {questoes.map((questao, qIndex) => (
              <div key={qIndex} className="rounded-xl border bg-card p-4">
                <div className="mb-3 flex items-center justify-between gap-2">
                  <span className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                    Questão {qIndex + 1}
                  </span>
                  <IconButton
                    label="Remover questão"
                    onClick={() => removeQuestao(qIndex)}
                  />
                </div>
                <Textarea
                  rows={2}
                  placeholder="Enunciado da pergunta"
                  value={questao.enunciado}
                  onChange={(e) => updateQuestao(qIndex, { enunciado: e.target.value })}
                />
                <div className="mt-3 flex flex-col gap-2">
                  {questao.opcoes.map((opcao, oIndex) => {
                    const correct = questao.correta === oIndex;
                    return (
                      <div
                        key={oIndex}
                        className={cn(
                          "flex items-center gap-2 rounded-lg border px-2 py-1.5 transition-colors",
                          correct
                            ? "border-emerald-300 bg-emerald-50/60"
                            : "border-transparent"
                        )}
                      >
                        <label
                          className="cursor-pointer"
                          title={correct ? "Alternativa correta" : "Marcar como correta"}
                        >
                          <input
                            type="radio"
                            name={`correta-${qIndex}`}
                            className="sr-only"
                            checked={correct}
                            onChange={() => updateQuestao(qIndex, { correta: oIndex })}
                          />
                          <span
                            className={cn(
                              "flex size-7 items-center justify-center rounded-md text-sm font-semibold transition-colors",
                              correct
                                ? "bg-emerald-500 text-white"
                                : "bg-muted text-muted-foreground hover:bg-primary/10 hover:text-primary"
                            )}
                          >
                            {correct ? <Check className="size-4" /> : LETTERS[oIndex]}
                          </span>
                        </label>
                        <Input
                          placeholder={`Alternativa ${LETTERS[oIndex]}`}
                          value={opcao}
                          onChange={(e) => updateOpcao(qIndex, oIndex, e.target.value)}
                          className="border-0 bg-transparent shadow-none focus-visible:ring-0"
                        />
                        <IconButton
                          label="Remover alternativa"
                          onClick={() => removeOpcao(qIndex, oIndex)}
                          disabled={questao.opcoes.length <= 2}
                        />
                      </div>
                    );
                  })}
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="mt-2"
                  onClick={() => addOpcao(qIndex)}
                >
                  <Plus className="size-4" />
                  Adicionar alternativa
                </Button>
              </div>
            ))}
          </div>
        )}
      </Section>

      {/* Sticky save bar */}
      <div className="fixed inset-x-0 bottom-0 z-20 border-t bg-background/90 backdrop-blur lg:pl-64">
        <div className="mx-auto flex max-w-4xl items-center justify-end gap-2 px-4 py-3 sm:px-6">
          {extraActions}
          <Button size="lg" onClick={onSave} disabled={saving}>
            {saving ? "Salvando…" : saveLabel}
          </Button>
        </div>
      </div>
    </div>
  );
}

function Section({
  icon,
  title,
  description,
  action,
  children,
}: {
  icon: React.ReactNode;
  title: string;
  description?: string;
  action?: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <Card className="p-5 sm:p-6">
      <div className="mb-4 flex items-start justify-between gap-3">
        <div className="flex items-start gap-3">
          <div className="flex size-9 items-center justify-center rounded-lg bg-brand-muted text-primary">
            {icon}
          </div>
          <div>
            <h2 className="font-[family-name:var(--font-display)] text-base font-semibold tracking-tight">
              {title}
            </h2>
            {description && (
              <p className="text-sm text-muted-foreground">{description}</p>
            )}
          </div>
        </div>
        {action}
      </div>
      {children}
    </Card>
  );
}

function PlanoProgress({ filled, total }: { filled: number; total: number }) {
  const pct = total === 0 ? 0 : Math.round((filled / total) * 100);
  return (
    <div className="flex items-center gap-3" aria-hidden>
      <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
        <div
          className="h-full rounded-full bg-gradient-to-r from-primary to-chart-2 transition-[width] duration-300"
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="text-xs font-medium tabular-nums text-muted-foreground">
        {filled} de {total} etapas
      </span>
    </div>
  );
}

function StepCircle({
  n,
  filled,
  amber,
}: {
  n: number;
  filled: boolean;
  amber: boolean;
}) {
  return (
    <div
      className={cn(
        "z-[1] flex size-9 shrink-0 items-center justify-center rounded-full border-2 bg-card text-sm font-bold transition-colors",
        filled && !amber && "border-primary bg-primary text-primary-foreground",
        filled && amber && "border-amber bg-amber text-amber-foreground",
        !filled && amber && "border-amber text-amber-foreground",
        !filled && !amber && "border-primary text-primary"
      )}
    >
      {filled ? <Check className="size-4" /> : n}
    </div>
  );
}

function StepBadge({ n }: { n: number }) {
  return (
    <div className="flex size-7 shrink-0 items-center justify-center rounded-full bg-primary/10 text-sm font-semibold text-primary">
      {n}
    </div>
  );
}

function IconButton({
  label,
  onClick,
  disabled,
}: {
  label: string;
  onClick: () => void;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      aria-label={label}
      title={label}
      className="flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive disabled:pointer-events-none disabled:opacity-30"
    >
      <Trash2 className="size-4" />
    </button>
  );
}

function EmptyHint({ text }: { text: string }) {
  return (
    <p className="rounded-lg border border-dashed bg-muted/30 px-4 py-6 text-center text-sm text-muted-foreground">
      {text}
    </p>
  );
}
