"use client";

import { useCallback, useEffect, useState } from "react";
import { KeyRound } from "lucide-react";
import { getAIToken, saveAIToken, deleteAIToken } from "@/lib/api";
import { AI_PROVIDER_LABELS, type AIProvider, type AITokenStatus } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const PROVIDERS = Object.keys(AI_PROVIDER_LABELS) as AIProvider[];

export default function PerfilPage() {
  const [status, setStatus] = useState<AITokenStatus | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [provider, setProvider] = useState<AIProvider>("anthropic");
  const [token, setToken] = useState("");
  const [saving, setSaving] = useState(false);
  const [removing, setRemoving] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const reload = useCallback(() => {
    getAIToken()
      .then((s) => {
        setStatus(s);
        if (s.provider) setProvider(s.provider);
        setLoadError(null);
      })
      .catch(() => {
        setLoadError("Não foi possível carregar as informações do token.");
      });
  }, []);

  useEffect(() => {
    reload();
  }, [reload]);

  async function onSave(e: React.FormEvent) {
    e.preventDefault();
    setFormError(null);
    setMessage(null);
    if (!token.trim()) {
      setFormError("Informe o token antes de salvar.");
      return;
    }
    setSaving(true);
    try {
      await saveAIToken(provider, token.trim());
      setToken("");
      setMessage("Token salvo. Por segurança, ele não será exibido novamente.");
      reload();
    } catch {
      setFormError("Não foi possível salvar o token. Verifique e tente novamente.");
    } finally {
      setSaving(false);
    }
  }

  async function onDelete() {
    setFormError(null);
    setMessage(null);
    setRemoving(true);
    try {
      await deleteAIToken();
      setMessage("Token removido.");
      reload();
    } catch {
      setFormError("Não foi possível remover o token. Tente novamente.");
    } finally {
      setRemoving(false);
    }
  }

  return (
    <div className="mx-auto flex max-w-2xl flex-col gap-8 px-4 py-6 sm:px-6 sm:py-8">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">Perfil</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Cadastre seu próprio token de IA para gerar e aprimorar aulas.
        </p>
      </header>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <KeyRound className="size-4 text-primary" />
            Token de IA
          </CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          {loadError && (
            <p className="text-sm text-destructive" role="alert">
              {loadError}
            </p>
          )}

          {!loadError && (
            <p className="text-sm text-muted-foreground">
              {status === null
                ? "Carregando..."
                : status.configured
                  ? `Configurado — ${AI_PROVIDER_LABELS[status.provider as AIProvider]}`
                  : "Nenhum token configurado. A geração com IA fica indisponível até você cadastrar um."}
            </p>
          )}

          <form onSubmit={onSave} className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="provider">Provedor</Label>
              <select
                id="provider"
                value={provider}
                onChange={(e) => setProvider(e.target.value as AIProvider)}
                className="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
              >
                {PROVIDERS.map((p) => (
                  <option key={p} value={p}>
                    {AI_PROVIDER_LABELS[p]}
                  </option>
                ))}
              </select>
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="token">Token</Label>
              <Input
                id="token"
                type="password"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                placeholder="Cole seu token de API"
                autoComplete="off"
              />
              <p className="text-xs text-muted-foreground">
                O token é armazenado de forma segura e não será exibido novamente após salvar.
              </p>
            </div>

            {formError && (
              <p className="text-sm text-destructive" role="alert">
                {formError}
              </p>
            )}
            {message && <p className="text-sm text-primary">{message}</p>}

            <div className="flex flex-wrap gap-2">
              <Button type="submit" disabled={saving || !token.trim()}>
                {saving ? "Salvando..." : "Salvar"}
              </Button>
              {status?.configured && (
                <Button
                  type="button"
                  variant="destructive"
                  onClick={onDelete}
                  disabled={removing}
                >
                  {removing ? "Removendo..." : "Remover"}
                </Button>
              )}
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
