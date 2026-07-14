"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { apiFetch } from "@/lib/api";
import type { TrailStats } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

function formatDate(iso: string | null): string {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleString("pt-BR", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return iso;
  }
}

export default function LessonStatsPage() {
  const params = useParams<{ id: string }>();
  const lessonId = params.id;

  const [stats, setStats] = useState<TrailStats | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    apiFetch<TrailStats>(`/api/trails/${lessonId}/stats`)
      .then((data) => {
        setStats(data);
        setLoadError(null);
      })
      .catch(() => {
        setLoadError("Não foi possível carregar as estatísticas da turma.");
      });
  }, [lessonId]);

  if (loadError) {
    return (
      <div className="mx-auto max-w-5xl p-4 sm:p-6">
        <p className="text-sm text-destructive" role="alert">
          {loadError}
        </p>
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="mx-auto max-w-5xl p-4 sm:p-6">
        <p className="text-sm text-muted-foreground">Carregando estatísticas...</p>
      </div>
    );
  }

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6 p-4 sm:p-6">
      <h1 className="text-2xl font-semibold tracking-tight">Turma da aula #{lessonId}</h1>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Total de alunos</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold">{stats.total_alunos}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Concluídos</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold">{stats.concluidos}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Média de pontos</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold">{stats.media_pontos.toFixed(1)}</p>
          </CardContent>
        </Card>
      </div>

      {stats.tentativas.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          Nenhuma tentativa registrada ainda para esta trilha.
        </p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Aluno</TableHead>
              <TableHead>Pontos</TableHead>
              <TableHead>Concluído em</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {stats.tentativas.map((tentativa, index) => (
              <TableRow key={`${tentativa.nome_aluno}-${index}`}>
                <TableCell>{tentativa.nome_aluno}</TableCell>
                <TableCell>{tentativa.pontos}</TableCell>
                <TableCell>{formatDate(tentativa.concluido_em)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  );
}
