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
  const [etapa, setEtapa] = useState("");
  const [disciplina, setDisciplina] = useState("");
  const [ano, setAno] = useState("");
  const [assunto, setAssunto] = useState("");
  const [query, setQuery] = useState("");

  useEffect(() => {
    let cancelled = false;
    apiFetch<BnccSkill[]>("/api/bncc-skills")
      .then((data) => !cancelled && setSkills(data))
      .catch(() => !cancelled && setSkills([]));
    return () => {
      cancelled = true;
    };
  }, []);

  // Disciplinas narrow to the selected etapa (EF and EM don't share them).
  const disciplinas = useMemo(
    () =>
      [
        ...new Set(
          (skills ?? [])
            .filter((s) => !etapa || s.etapa === etapa)
            .map((s) => s.disciplina)
        ),
      ].sort(),
    [skills, etapa]
  );

  // Anos available for the current etapa (a skill can span multiple anos).
  const anos = useMemo(
    () =>
      [
        ...new Set(
          (skills ?? [])
            .filter((s) => !etapa || s.etapa === etapa)
            .flatMap((s) => s.anos)
        ),
      ].sort((a, b) => a - b),
    [skills, etapa]
  );

  // A stored filter can go stale when a broader one changes (switching etapa
  // drops its disciplinas/anos). Derive the effective value during render rather
  // than resetting state in an effect — same result, no cascading re-render.
  const effDisciplina =
    disciplina && disciplinas.includes(disciplina) ? disciplina : "";
  const effAno = ano && anos.includes(Number(ano)) ? ano : "";

  // Matérias (assuntos) for the current etapa/disciplina/ano. The catalog
  // repeats each assunto across many codes and years; collapse to a distinct,
  // deduped list so the teacher picks the matéria once — the ano select scopes
  // which year's habilidades appear below.
  const assuntos = useMemo(
    () =>
      [
        ...new Set(
          (skills ?? [])
            .filter(
              (s) =>
                (!etapa || s.etapa === etapa) &&
                (!effDisciplina || s.disciplina === effDisciplina) &&
                (!effAno || s.anos.includes(Number(effAno)))
            )
            .map((s) => s.assunto)
            .filter((a) => a.trim().length > 0)
        ),
      ].sort((a, b) => a.localeCompare(b, "pt-BR")),
    [skills, etapa, effDisciplina, effAno]
  );

  const effAssunto = assunto && assuntos.includes(assunto) ? assunto : "";

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return (skills ?? []).filter(
      (s) =>
        (!etapa || s.etapa === etapa) &&
        (!effDisciplina || s.disciplina === effDisciplina) &&
        (!effAno || s.anos.includes(Number(effAno))) &&
        (!effAssunto || s.assunto === effAssunto) &&
        (!q ||
          s.code.toLowerCase().includes(q) ||
          s.disciplina.toLowerCase().includes(q) ||
          s.assunto.toLowerCase().includes(q) ||
          s.descricao.toLowerCase().includes(q))
    );
  }, [skills, etapa, effDisciplina, effAno, effAssunto, query]);

  // Full catalog is large; cap the rendered options and nudge to refine.
  const LIMIT = 300;
  const shown = filtered.slice(0, LIMIT);
  const overflow = filtered.length - shown.length;

  const etapas = [
    { value: "", label: "Todas" },
    { value: "EF", label: "Fundamental" },
    { value: "EM", label: "Médio" },
  ];

  return (
    <div className="flex flex-col gap-5">
      <Field label="Etapa de ensino">
        <div className="flex flex-wrap gap-2">
          {etapas.map((e) => (
            <button
              key={e.value}
              type="button"
              onClick={() => setEtapa(e.value)}
              className={cn(
                "rounded-full border px-3.5 py-1.5 text-sm font-medium transition-colors",
                etapa === e.value
                  ? "border-primary bg-primary text-primary-foreground"
                  : "border-input bg-card text-muted-foreground hover:border-primary/40"
              )}
            >
              {e.label}
            </button>
          ))}
        </div>
      </Field>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Field label="Ano">
          <select
            className={selectClass}
            value={effAno}
            onChange={(e) => setAno(e.target.value)}
          >
            <option value="">Todos os anos</option>
            {anos.map((n) => (
              <option key={n} value={n}>
                {n}º ano
              </option>
            ))}
          </select>
        </Field>
        <Field label="Disciplina">
          <select
            className={selectClass}
            value={effDisciplina}
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
      </div>
      <Field label="Matéria / assunto" htmlFor="bncc-assunto">
        <select
          id="bncc-assunto"
          className={selectClass}
          value={effAssunto}
          onChange={(e) => setAssunto(e.target.value)}
          disabled={skills === null || assuntos.length === 0}
        >
          <option value="">
            {skills === null
              ? "Carregando…"
              : `Todas as matérias (${assuntos.length})`}
          </option>
          {assuntos.map((a) => (
            <option key={a} value={a}>
              {a}
            </option>
          ))}
        </select>
      </Field>
      <Field label="Buscar habilidade / tópico">
        <Input
          type="search"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="ex.: frações, leitura, EF67LP08…"
        />
      </Field>

      <Field label="Habilidade BNCC" icon={<BookMarked className="size-3.5" />}>
        {skills === null ? (
          <p className="text-sm text-muted-foreground">Carregando…</p>
        ) : filtered.length === 0 ? (
          <p className="rounded-lg border border-dashed bg-muted/30 px-4 py-6 text-center text-sm text-muted-foreground">
            Nenhuma habilidade encontrada. Ajuste os filtros acima.
          </p>
        ) : (
          <>
            <p className="mb-1.5 text-xs text-muted-foreground">
              {filtered.length}{" "}
              {filtered.length === 1 ? "habilidade" : "habilidades"} · toque para
              selecionar
            </p>
            <div
              role="listbox"
              aria-label="Habilidades BNCC"
              className="flex max-h-80 flex-col gap-1.5 overflow-y-auto rounded-lg border bg-card p-1.5"
            >
              {shown.map((s) => {
                const active = s.id === bnccSkillId;
                return (
                  <button
                    key={s.id}
                    type="button"
                    role="option"
                    aria-selected={active}
                    onClick={() => onBnccSkillIdChange(active ? null : s.id)}
                    className={cn(
                      "flex flex-col gap-1 rounded-md border px-3 py-2 text-left transition-colors",
                      active
                        ? "border-primary bg-brand-muted/60"
                        : "border-transparent hover:bg-muted"
                    )}
                  >
                    <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
                      <span
                        className={cn(
                          "rounded px-1.5 py-0.5 text-xs font-semibold",
                          active
                            ? "bg-primary text-primary-foreground"
                            : "bg-primary/10 text-primary"
                        )}
                      >
                        {s.code}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {s.disciplina} · {s.ano_label}
                      </span>
                      {s.assunto && (
                        <span className="rounded-full bg-muted px-1.5 py-0.5 text-[11px] font-medium text-muted-foreground">
                          {s.assunto}
                        </span>
                      )}
                    </div>
                    <p className="text-sm leading-snug text-foreground/90">
                      {s.descricao}
                    </p>
                  </button>
                );
              })}
            </div>
            {overflow > 0 && (
              <p className="mt-1.5 text-xs text-muted-foreground">
                Mostrando {LIMIT} de {filtered.length}. Refine os filtros para
                ver o restante.
              </p>
            )}
          </>
        )}
      </Field>

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
