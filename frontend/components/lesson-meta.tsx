"use client";

import { useEffect, useMemo, useState } from "react";
import { BookMarked, Clock } from "lucide-react";
import { apiFetch } from "@/lib/api";
import type { BnccSkill } from "@/lib/types";
import { cn } from "@/lib/utils";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const DURACOES = [30, 50, 90];

const selectClass =
  "h-10 w-full rounded-lg border border-input bg-card px-3 text-sm outline-none transition-colors focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/40 disabled:opacity-50";

export function LessonMeta({
  bnccSkillId,
  onBnccSkillIdChange,
  duracao,
  onDuracaoChange,
}: {
  bnccSkillId: number | null;
  onBnccSkillIdChange: (id: number | null) => void;
  duracao: number;
  onDuracaoChange: (d: number) => void;
}) {
  const [skills, setSkills] = useState<BnccSkill[] | null>(null);
  const [disciplina, setDisciplina] = useState("");
  const [ano, setAno] = useState("");
  const [assunto, setAssunto] = useState("");

  useEffect(() => {
    let cancelled = false;
    apiFetch<BnccSkill[]>("/api/bncc-skills")
      .then((data) => !cancelled && setSkills(data))
      .catch(() => !cancelled && setSkills([]));
    return () => {
      cancelled = true;
    };
  }, []);

  const disciplinas = useMemo(
    () => [...new Set((skills ?? []).map((s) => s.disciplina))].sort(),
    [skills]
  );
  const anos = useMemo(
    () => [...new Set((skills ?? []).map((s) => s.ano))].sort(),
    [skills]
  );

  // Assuntos available narrow to the current disciplina/ano selection.
  const assuntos = useMemo(
    () =>
      [
        ...new Set(
          (skills ?? [])
            .filter(
              (s) =>
                (!disciplina || s.disciplina === disciplina) &&
                (!ano || s.ano === ano)
            )
            .map((s) => s.assunto)
            .filter(Boolean)
        ),
      ].sort(),
    [skills, disciplina, ano]
  );

  // Drop the assunto filter if it no longer applies to the current selection.
  useEffect(() => {
    if (assunto && !assuntos.includes(assunto)) setAssunto("");
  }, [assunto, assuntos]);

  const filtered = useMemo(
    () =>
      (skills ?? []).filter(
        (s) =>
          (!disciplina || s.disciplina === disciplina) &&
          (!ano || s.ano === ano) &&
          (!assunto || s.assunto === assunto)
      ),
    [skills, disciplina, ano, assunto]
  );

  const selected = (skills ?? []).find((s) => s.id === bnccSkillId) ?? null;

  return (
    <div className="flex flex-col gap-5">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <Field label="Disciplina">
          <select
            className={selectClass}
            value={disciplina}
            onChange={(e) => setDisciplina(e.target.value)}
          >
            <option value="">Todas as disciplinas</option>
            {disciplinas.map((d) => (
              <option key={d} value={d}>
                {d}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Ano / série">
          <select
            className={selectClass}
            value={ano}
            onChange={(e) => setAno(e.target.value)}
          >
            <option value="">Todos os anos</option>
            {anos.map((a) => (
              <option key={a} value={a}>
                {a}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Assunto">
          <select
            className={selectClass}
            value={assunto}
            onChange={(e) => setAssunto(e.target.value)}
            disabled={assuntos.length === 0}
          >
            <option value="">Todos os assuntos</option>
            {assuntos.map((a) => (
              <option key={a} value={a}>
                {a}
              </option>
            ))}
          </select>
        </Field>
      </div>

      <Field
        label="Habilidade BNCC"
        icon={<BookMarked className="size-3.5" />}
        htmlFor="bncc-skill"
      >
        <select
          id="bncc-skill"
          className={selectClass}
          value={bnccSkillId ?? ""}
          onChange={(e) =>
            onBnccSkillIdChange(
              e.target.value === "" ? null : Number(e.target.value)
            )
          }
        >
          <option value="">
            {skills === null ? "Carregando…" : "Selecione uma habilidade"}
          </option>
          {filtered.map((s) => (
            <option key={s.id} value={s.id}>
              {s.code} · {s.assunto} · {s.ano}
            </option>
          ))}
        </select>
      </Field>

      {selected && (
        <div className="rounded-lg border border-primary/20 bg-brand-muted/50 px-4 py-3">
          <div className="flex flex-wrap items-center gap-2">
            <span className="rounded bg-primary/10 px-1.5 py-0.5 text-xs font-semibold text-primary">
              {selected.code}
            </span>
            <span className="text-xs text-muted-foreground">
              {selected.disciplina} · {selected.ano} · {selected.assunto}
            </span>
          </div>
          <p className="mt-2 text-sm text-foreground/80">{selected.descricao}</p>
        </div>
      )}

      <Field label="Duração da aula" icon={<Clock className="size-3.5" />}>
        <div className="flex flex-wrap items-center gap-2">
          {DURACOES.map((d) => (
            <button
              key={d}
              type="button"
              onClick={() => onDuracaoChange(d)}
              className={cn(
                "rounded-full border px-3.5 py-1.5 text-sm font-medium transition-colors",
                duracao === d
                  ? "border-primary bg-primary text-primary-foreground"
                  : "border-input bg-card text-muted-foreground hover:border-primary/40"
              )}
            >
              {d} min
            </button>
          ))}
          <div className="flex items-center gap-1.5">
            <Input
              type="number"
              min={0}
              value={duracao}
              onChange={(e) => onDuracaoChange(Number(e.target.value))}
              className="w-20"
              aria-label="Duração personalizada em minutos"
            />
            <span className="text-sm text-muted-foreground">min</span>
          </div>
        </div>
      </Field>
    </div>
  );
}

function Field({
  label,
  icon,
  htmlFor,
  children,
}: {
  label: string;
  icon?: React.ReactNode;
  htmlFor?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <Label
        htmlFor={htmlFor}
        className="flex items-center gap-1.5 text-muted-foreground"
      >
        {icon}
        {label}
      </Label>
      {children}
    </div>
  );
}
