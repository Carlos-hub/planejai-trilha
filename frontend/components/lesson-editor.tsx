"use client";

import {
  Target,
  Route,
  ListChecks,
  Plus,
  Trash2,
  Check,
  NotebookPen,
  Presentation,
  Backpack,
  ClipboardCheck,
  Sparkles,
  type LucideIcon,
} from "lucide-react";
import type { Plano, Questao, Topico, Trilha } from "@/lib/types";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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

const PLANO_FIELDS: {
  key: keyof Plano;
  label: string;
  helper: string;
  placeholder: string;
  icon: LucideIcon;
}[] = [
  {
    key: "objetivos",
    label: "Objetivos",
    icon: Target,
    helper: "O que o aluno deve ser capaz de fazer ao final da aula.",
    placeholder: "Ao final, o aluno será capaz de…",
  },
  {
    key: "metodologia",
    label: "Metodologia",
    icon: Presentation,
    helper: "Como você conduz a aula, do começo ao fim.",
    placeholder: "Comece com…, depois…, para fechar…",
  },
  {
    key: "recursos",
    label: "Recursos",
    icon: Backpack,
    helper: "Materiais e ferramentas que você vai usar.",
    placeholder: "Quadro, cartaz, folhas, música…",
  },
  {
    key: "avaliacao",
    label: "Avaliação",
    icon: ClipboardCheck,
    helper: "Como você verifica se o aluno aprendeu.",
    placeholder: "Observação da participação, produção do aluno…",
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
        icon={<NotebookPen className="size-4" />}
        title="Plano de aula"
        description="Preencha na ordem — cada passo guia o próximo. Não aparece para o aluno."
      >
        <div className="flex flex-col">
          {PLANO_FIELDS.map(({ key, label, helper, placeholder, icon: Icon }, i) => {
            const filled = value.plano[key].trim().length > 0;
            const last = i === PLANO_FIELDS.length - 1;
            return (
              <div key={key} className="flex gap-3 sm:gap-4">
                {/* Progress rail — fills as each step is completed. */}
                <div className="flex flex-col items-center">
                  <div
                    className={cn(
                      "flex size-8 shrink-0 items-center justify-center rounded-full border text-sm font-semibold transition-colors",
                      filled
                        ? "border-primary bg-primary text-primary-foreground"
                        : "border-primary/20 bg-primary/10 text-primary"
                    )}
                  >
                    {filled ? <Check className="size-4" /> : i + 1}
                  </div>
                  {!last && (
                    <div
                      className={cn(
                        "w-px flex-1 transition-colors",
                        filled ? "bg-primary/40" : "bg-border"
                      )}
                    />
                  )}
                </div>
                {/* Field */}
                <div className={cn("flex-1", last ? "pb-1" : "pb-5")}>
                  <div className="flex items-center gap-1.5">
                    <Icon className="size-4 text-primary" />
                    <Label htmlFor={key} className="text-sm font-semibold">
                      {label}
                    </Label>
                  </div>
                  <p className="mb-2 mt-0.5 text-xs text-muted-foreground">{helper}</p>
                  <Textarea
                    id={key}
                    rows={3}
                    placeholder={placeholder}
                    value={value.plano[key]}
                    onChange={(e) => updatePlano(key, e.target.value)}
                  />
                </div>
              </div>
            );
          })}
        </div>

        {/* Atividade — the hands-on task that closes the plan. */}
        <div className="mt-2 rounded-xl border border-dashed bg-brand-muted/40 p-4">
          <div className="flex items-center gap-1.5">
            <Sparkles className="size-4 text-primary" />
            <Label htmlFor="atividade" className="text-sm font-semibold">
              Atividade prática
            </Label>
          </div>
          <p className="mb-2 mt-0.5 text-xs text-muted-foreground">
            Uma tarefa concreta para o aluno aplicar o que aprendeu.
          </p>
          <Textarea
            id="atividade"
            rows={3}
            placeholder="Descreva a atividade que o aluno vai fazer…"
            value={value.atividade}
            onChange={(e) => onChange({ ...value, atividade: e.target.value })}
          />
        </div>
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
