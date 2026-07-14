"use client";

import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import type { BnccSkill, Plano, Questao, Topico, Trilha } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "@/components/ui/card";

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

function emptyTopico(): Topico {
  return { titulo: "", resumo: "" };
}

function emptyQuestao(): Questao {
  return { enunciado: "", opcoes: ["", ""], correta: 0 };
}

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
  const [skills, setSkills] = useState<BnccSkill[] | null>(null);

  useEffect(() => {
    let cancelled = false;
    apiFetch<BnccSkill[]>("/api/bncc-skills")
      .then((data) => {
        if (!cancelled) setSkills(data);
      })
      .catch(() => {
        if (!cancelled) setSkills([]);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  function updatePlano(field: keyof Plano, text: string) {
    onChange({ ...value, plano: { ...value.plano, [field]: text } });
  }

  function updateAtividade(text: string) {
    onChange({ ...value, atividade: text });
  }

  function updateTopico(index: number, field: keyof Topico, text: string) {
    const topicos = value.trilha.topicos.map((t, i) =>
      i === index ? { ...t, [field]: text } : t
    );
    onChange({ ...value, trilha: { ...value.trilha, topicos } });
  }

  function addTopico() {
    onChange({
      ...value,
      trilha: { ...value.trilha, topicos: [...value.trilha.topicos, emptyTopico()] },
    });
  }

  function removeTopico(index: number) {
    onChange({
      ...value,
      trilha: {
        ...value.trilha,
        topicos: value.trilha.topicos.filter((_, i) => i !== index),
      },
    });
  }

  function updateQuestao(index: number, patch: Partial<Questao>) {
    const questoes = value.trilha.quiz.questoes.map((q, i) =>
      i === index ? { ...q, ...patch } : q
    );
    onChange({ ...value, trilha: { ...value.trilha, quiz: { questoes } } });
  }

  function addQuestao() {
    onChange({
      ...value,
      trilha: {
        ...value.trilha,
        quiz: { questoes: [...value.trilha.quiz.questoes, emptyQuestao()] },
      },
    });
  }

  function removeQuestao(index: number) {
    onChange({
      ...value,
      trilha: {
        ...value.trilha,
        quiz: { questoes: value.trilha.quiz.questoes.filter((_, i) => i !== index) },
      },
    });
  }

  function updateOpcao(qIndex: number, oIndex: number, text: string) {
    const questao = value.trilha.quiz.questoes[qIndex];
    const opcoes = questao.opcoes.map((o, i) => (i === oIndex ? text : o));
    updateQuestao(qIndex, { opcoes });
  }

  function addOpcao(qIndex: number) {
    const questao = value.trilha.quiz.questoes[qIndex];
    updateQuestao(qIndex, { opcoes: [...questao.opcoes, ""] });
  }

  function removeOpcao(qIndex: number, oIndex: number) {
    const questao = value.trilha.quiz.questoes[qIndex];
    const opcoes = questao.opcoes.filter((_, i) => i !== oIndex);
    const correta = questao.correta >= opcoes.length ? 0 : questao.correta;
    updateQuestao(qIndex, { opcoes, correta });
  }

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader>
          <CardTitle>Dados gerais</CardTitle>
          <CardDescription>Habilidade BNCC e duração da aula.</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4 sm:flex-row">
          <div className="flex flex-1 flex-col gap-1.5">
            <Label htmlFor="bncc-skill">Habilidade BNCC</Label>
            <select
              id="bncc-skill"
              className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
              value={bnccSkillId ?? ""}
              onChange={(e) =>
                onBnccSkillIdChange(e.target.value === "" ? null : Number(e.target.value))
              }
            >
              <option value="">Selecione...</option>
              {(skills ?? []).map((skill) => (
                <option key={skill.id} value={skill.id}>
                  {skill.code} · {skill.disciplina} · {skill.ano}
                </option>
              ))}
            </select>
          </div>
          <div className="flex w-full flex-col gap-1.5 sm:w-40">
            <Label htmlFor="duracao">Duração (min)</Label>
            <Input
              id="duracao"
              type="number"
              min={0}
              value={duracao}
              onChange={(e) => onDuracaoChange(Number(e.target.value))}
            />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Plano de aula</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="objetivos">Objetivos</Label>
            <Textarea
              id="objetivos"
              value={value.plano.objetivos}
              onChange={(e) => updatePlano("objetivos", e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="metodologia">Metodologia</Label>
            <Textarea
              id="metodologia"
              value={value.plano.metodologia}
              onChange={(e) => updatePlano("metodologia", e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="recursos">Recursos</Label>
            <Textarea
              id="recursos"
              value={value.plano.recursos}
              onChange={(e) => updatePlano("recursos", e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="avaliacao">Avaliação</Label>
            <Textarea
              id="avaliacao"
              value={value.plano.avaliacao}
              onChange={(e) => updatePlano("avaliacao", e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="atividade">Atividade</Label>
            <Textarea
              id="atividade"
              value={value.atividade}
              onChange={(e) => updateAtividade(e.target.value)}
            />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Tópicos da trilha</CardTitle>
          <CardDescription>Os alunos percorrem estes tópicos em ordem.</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          {value.trilha.topicos.map((topico, index) => (
            <div key={index} className="flex flex-col gap-2 rounded-lg border p-3">
              <div className="flex items-center justify-between gap-2">
                <span className="text-xs font-medium text-muted-foreground">
                  Tópico {index + 1}
                </span>
                <Button
                  type="button"
                  variant="ghost"
                  size="xs"
                  onClick={() => removeTopico(index)}
                >
                  Remover
                </Button>
              </div>
              <Input
                placeholder="Título"
                value={topico.titulo}
                onChange={(e) => updateTopico(index, "titulo", e.target.value)}
              />
              <Textarea
                placeholder="Resumo"
                value={topico.resumo}
                onChange={(e) => updateTopico(index, "resumo", e.target.value)}
              />
            </div>
          ))}
          <Button type="button" variant="outline" size="sm" onClick={addTopico}>
            + Adicionar tópico
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Quiz</CardTitle>
          <CardDescription>Selecione a opção correta de cada questão.</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          {value.trilha.quiz.questoes.map((questao, qIndex) => (
            <div key={qIndex} className="flex flex-col gap-2 rounded-lg border p-3">
              <div className="flex items-center justify-between gap-2">
                <span className="text-xs font-medium text-muted-foreground">
                  Questão {qIndex + 1}
                </span>
                <Button
                  type="button"
                  variant="ghost"
                  size="xs"
                  onClick={() => removeQuestao(qIndex)}
                >
                  Remover
                </Button>
              </div>
              <Textarea
                placeholder="Enunciado"
                value={questao.enunciado}
                onChange={(e) => updateQuestao(qIndex, { enunciado: e.target.value })}
              />
              <div className="flex flex-col gap-2">
                {questao.opcoes.map((opcao, oIndex) => (
                  <div key={oIndex} className="flex items-center gap-2">
                    <input
                      type="radio"
                      name={`correta-${qIndex}`}
                      checked={questao.correta === oIndex}
                      onChange={() => updateQuestao(qIndex, { correta: oIndex })}
                      aria-label={`Opção ${oIndex + 1} correta`}
                    />
                    <Input
                      placeholder={`Opção ${oIndex + 1}`}
                      value={opcao}
                      onChange={(e) => updateOpcao(qIndex, oIndex, e.target.value)}
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="xs"
                      disabled={questao.opcoes.length <= 2}
                      onClick={() => removeOpcao(qIndex, oIndex)}
                    >
                      Remover
                    </Button>
                  </div>
                ))}
                <Button
                  type="button"
                  variant="outline"
                  size="xs"
                  className="self-start"
                  onClick={() => addOpcao(qIndex)}
                >
                  + Adicionar opção
                </Button>
              </div>
            </div>
          ))}
          <Button type="button" variant="outline" size="sm" onClick={addQuestao}>
            + Adicionar questão
          </Button>
        </CardContent>
      </Card>

      <div className="flex items-center justify-end gap-2">
        {extraActions}
        <Button onClick={onSave} disabled={saving}>
          {saving ? "Salvando..." : saveLabel}
        </Button>
      </div>
    </div>
  );
}
