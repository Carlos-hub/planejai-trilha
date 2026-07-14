"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useMe } from "@/lib/hooks";
import { apiFetch } from "@/lib/api";
import { Button } from "@/components/ui/button";

export default function TeacherLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { me, loading, error } = useMe();
  const router = useRouter();
  const pathname = usePathname();
  const isLoginPage = pathname === "/login";

  useEffect(() => {
    if (!loading && error && !isLoginPage) {
      router.push("/login");
    }
  }, [loading, error, isLoginPage, router]);

  // The login page renders standalone, without the top bar or auth guard.
  if (isLoginPage) {
    return <>{children}</>;
  }

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center text-sm text-muted-foreground">
        Carregando...
      </div>
    );
  }

  if (error || !me) {
    return null;
  }

  async function handleLogout() {
    try {
      await apiFetch<void>("/api/auth/logout", { method: "POST" });
    } finally {
      router.push("/login");
    }
  }

  return (
    <div className="flex min-h-screen flex-col">
      <header className="flex items-center justify-between border-b px-4 py-2.5">
        <span className="text-sm font-medium">PlanejAI+Trilha</span>
        <div className="flex items-center gap-3">
          <span className="text-sm text-muted-foreground">{me.nome}</span>
          <Button variant="outline" size="sm" onClick={handleLogout}>
            Sair
          </Button>
        </div>
      </header>
      <main className="flex-1">{children}</main>
    </div>
  );
}
