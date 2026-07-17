"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Download, Printer, Upload, UserPlus } from "lucide-react";
import { addStudent, getTurma, importStudentsCSV } from "@/lib/api";
import type { ImportedStudent } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

type Aluno = { id: number; nome: string; usuario: string; matricula: string | null };

export default function TurmaDetailPage() {
  const params = useParams<{ id: string }>();
  const turmaId = Number(params.id);

  const [nome, setNome] = useState("");
  const [alunos, setAlunos] = useState<Aluno[]>([]);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [criados, setCriados] = useState<ImportedStudent[]>([]);
  const [importing, setImporting] = useState(false);
  const [importError, setImportError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [addNome, setAddNome] = useState("");
  const [addMatricula, setAddMatricula] = useState("");
  const [addingStudent, setAddingStudent] = useState(false);
  const [addError, setAddError] = useState<string | null>(null);

  const reload = useCallback(() => {
    getTurma(turmaId)
      .then((data) => {
        setNome(data.turma.nome);
        setAlunos(data.alunos);
        setLoadError(null);
      })
      .catch(() => {
        setLoadError("Não foi possível carregar esta turma.");
      });
  }, [turmaId]);

  useEffect(() => {
    reload();
  }, [reload]);

  async function onFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setImporting(true);
    setImportError(null);
    try {
      const text = await file.text();
      const res = await importStudentsCSV(turmaId, text);
      setCriados(res.criados);
      reload();
    } catch {
      setImportError("Não foi possível importar o arquivo. Verifique o formato do CSV.");
    } finally {
      setImporting(false);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  }

  async function onAddStudent(e: React.FormEvent) {
    e.preventDefault();
    const nome = addNome.trim();
    if (!nome) {
      setAddError("Informe o nome do aluno.");
      return;
    }
    setAddingStudent(true);
    setAddError(null);
    try {
      const created = await addStudent(turmaId, nome, addMatricula);
      setCriados((prev) => [...prev, created]);
      setAddNome("");
      setAddMatricula("");
      reload();
    } catch {
      setAddError("Não foi possível adicionar o aluno.");
    } finally {
      setAddingStudent(false);
    }
  }

  function downloadCredentials() {
    const rows = [
      ["nome", "usuario", "senha"],
      ...criados.map((c) => [c.nome, c.usuario, c.senha]),
    ];
    const csv = rows.map((r) => r.map((v) => `"${v.replace(/"/g, '""')}"`).join(",")).join("\n");
    const url = URL.createObjectURL(new Blob([csv], { type: "text/csv;charset=utf-8" }));
    const a = document.createElement("a");
    a.href = url;
    a.download = `credenciais-${nome || turmaId}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }

  if (loadError) {
    return (
      <div className="mx-auto max-w-4xl p-4 sm:p-6">
        <p className="text-sm text-destructive" role="alert">
          {loadError}
        </p>
      </div>
    );
  }

  return (
    <div className="mx-auto flex max-w-4xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <div className="flex flex-col gap-2">
        <Link
          href="/turmas"
          className="inline-flex w-fit items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-4" />
          Turmas
        </Link>
        <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">
          {nome || "Carregando..."}
        </h1>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Importar alunos (CSV)</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          <p className="text-sm text-muted-foreground">
            Colunas: <code>nome</code> (obrigatória), <code>matricula</code> (opcional).
          </p>
          <div>
            <Button
              type="button"
              variant="outline"
              disabled={importing}
              onClick={() => fileInputRef.current?.click()}
            >
              <Upload className="size-4" />
              {importing ? "Importando..." : "Selecionar arquivo CSV"}
            </Button>
            <input
              ref={fileInputRef}
              type="file"
              accept=".csv,text/csv"
              onChange={onFile}
              className="hidden"
            />
          </div>
          {importError && (
            <p className="text-sm text-destructive" role="alert">
              {importError}
            </p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Adicionar aluno manualmente</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={onAddStudent} className="flex flex-col gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="add-nome">Nome</Label>
                <Input
                  id="add-nome"
                  value={addNome}
                  onChange={(e) => setAddNome(e.target.value)}
                  placeholder="Nome do aluno"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="add-matricula">Matrícula (opcional)</Label>
                <Input
                  id="add-matricula"
                  value={addMatricula}
                  onChange={(e) => setAddMatricula(e.target.value)}
                  placeholder="Matrícula"
                />
              </div>
            </div>
            <div>
              <Button type="submit" disabled={addingStudent}>
                <UserPlus className="size-4" />
                {addingStudent ? "Adicionando..." : "Adicionar aluno"}
              </Button>
            </div>
            {addError && (
              <p className="text-sm text-destructive" role="alert">
                {addError}
              </p>
            )}
          </form>
        </CardContent>
      </Card>

      {criados.length > 0 && (
        <Card className="print:shadow-none">
          <CardHeader>
            <CardTitle>Credenciais geradas</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <p className="rounded-lg border border-amber/40 bg-amber/10 px-4 py-3 text-sm">
              <strong>Guarde estas credenciais agora.</strong> As senhas são exibidas apenas
              nesta tela e não poderão ser recuperadas depois.
            </p>
            <div className="flex flex-wrap gap-2 print:hidden">
              <Button type="button" variant="outline" onClick={downloadCredentials}>
                <Download className="size-4" />
                Baixar credenciais (CSV)
              </Button>
              <Button type="button" variant="outline" onClick={() => window.print()}>
                <Printer className="size-4" />
                Imprimir
              </Button>
            </div>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Nome</TableHead>
                  <TableHead>Usuário</TableHead>
                  <TableHead>Senha</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {criados.map((c) => (
                  <TableRow key={c.usuario}>
                    <TableCell>{c.nome}</TableCell>
                    <TableCell>{c.usuario}</TableCell>
                    <TableCell className="font-mono">{c.senha}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Alunos ({alunos.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {alunos.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              Nenhum aluno ainda. Adicione manualmente ou importe uma planilha CSV.
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Nome</TableHead>
                  <TableHead>Usuário</TableHead>
                  <TableHead>Matrícula</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {alunos.map((a) => (
                  <TableRow key={a.id}>
                    <TableCell>{a.nome}</TableCell>
                    <TableCell className="font-mono">{a.usuario}</TableCell>
                    <TableCell>{a.matricula ?? "—"}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
