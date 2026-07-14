"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export function ShareLinks({ codigo, publicaUrl }: { codigo: string; publicaUrl: string }) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(publicaUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // clipboard unavailable; ignore silently
    }
  }

  const whatsappHref = `https://wa.me/?text=${encodeURIComponent(
    `Acesse a trilha de estudos: ${publicaUrl}`
  )}`;
  const pdfHref = `${API_BASE}/api/t/${codigo}/export.pdf`;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Trilha publicada</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div>
          <p className="text-xs text-muted-foreground">Código da trilha</p>
          <p className="text-3xl font-bold tracking-widest">{codigo}</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button type="button" variant="outline" onClick={handleCopy}>
            {copied ? "Link copiado!" : "Copiar link"}
          </Button>
          <Button
            type="button"
            variant="outline"
            render={<a href={whatsappHref} target="_blank" rel="noopener noreferrer" />}
          >
            Compartilhar no WhatsApp
          </Button>
          <Button
            type="button"
            variant="outline"
            render={<a href={pdfHref} target="_blank" rel="noopener noreferrer" />}
          >
            Baixar PDF
          </Button>
          <Button
            type="button"
            render={<a href={publicaUrl} target="_blank" rel="noopener noreferrer" />}
          >
            Abrir trilha pública
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
