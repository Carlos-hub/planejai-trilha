"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { LayoutDashboard, PlusCircle, LogOut, Menu, X } from "lucide-react";
import { useMe } from "@/lib/hooks";
import { apiFetch } from "@/lib/api";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { BrandLockup } from "@/components/brand-mark";

const NAV = [
  { href: "/", label: "Painel", icon: LayoutDashboard, match: (p: string) => p === "/" },
  {
    href: "/lessons/new",
    label: "Nova aula",
    icon: PlusCircle,
    match: (p: string) => p.startsWith("/lessons/new"),
  },
];

export default function TeacherLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { me, loading, error } = useMe();
  const router = useRouter();
  const pathname = usePathname();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const isLoginPage = pathname === "/login";

  useEffect(() => {
    if (!loading && error && !isLoginPage) {
      router.push("/login");
    }
  }, [loading, error, isLoginPage, router]);

  // Close the mobile drawer whenever the route changes.
  useEffect(() => {
    setDrawerOpen(false);
  }, [pathname]);

  // The login page renders standalone, without the shell or auth guard.
  if (isLoginPage) {
    return <>{children}</>;
  }

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center text-sm text-muted-foreground">
        Carregando…
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

  const initials = me.nome.trim().slice(0, 1).toUpperCase() || "P";

  const nav = (
    <nav className="flex flex-col gap-1">
      {NAV.map(({ href, label, icon: Icon, match }) => {
        const active = match(pathname);
        return (
          <Link
            key={href}
            href={href}
            className={cn(
              "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
              active
                ? "bg-sidebar-accent text-sidebar-accent-foreground"
                : "text-sidebar-foreground/70 hover:bg-sidebar-accent/60 hover:text-sidebar-accent-foreground"
            )}
          >
            <Icon className="size-4 shrink-0" />
            {label}
          </Link>
        );
      })}
    </nav>
  );

  const userBlock = (
    <div className="flex items-center gap-3 rounded-lg bg-sidebar-accent/40 px-3 py-2.5">
      <div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-sidebar-primary font-semibold text-sidebar-primary-foreground">
        {initials}
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-sidebar-accent-foreground">
          {me.nome}
        </p>
        <p className="truncate text-xs text-sidebar-foreground/60">{me.email}</p>
      </div>
      <button
        type="button"
        onClick={handleLogout}
        title="Sair"
        aria-label="Sair"
        className="flex size-8 items-center justify-center rounded-md text-sidebar-foreground/70 transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
      >
        <LogOut className="size-4" />
      </button>
    </div>
  );

  const sidebarInner = (
    <div className="flex h-full flex-col gap-6 bg-sidebar bg-gradient-to-b from-white/[0.04] to-black/[0.06] p-4 text-sidebar-foreground">
      <div className="px-2 pt-2 text-sidebar-accent-foreground">
        <BrandLockup />
        <p className="mt-1 pl-9 text-xs text-sidebar-foreground/60">
          Área do professor
        </p>
      </div>
      <div className="flex-1">{nav}</div>
      {userBlock}
    </div>
  );

  return (
    <div className="min-h-screen bg-background">
      {/* Desktop sidebar */}
      <aside className="fixed inset-y-0 left-0 z-30 hidden w-64 border-r border-sidebar-border lg:block">
        {sidebarInner}
      </aside>

      {/* Mobile top bar */}
      <header className="sticky top-0 z-20 flex items-center justify-between border-b bg-background/85 px-4 py-2.5 backdrop-blur lg:hidden">
        <BrandLockup markClassName="text-primary" />
        <button
          type="button"
          onClick={() => setDrawerOpen(true)}
          aria-label="Abrir menu"
          className="flex size-9 items-center justify-center rounded-md border text-foreground"
        >
          <Menu className="size-5" />
        </button>
      </header>

      {/* Mobile drawer */}
      {drawerOpen && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <div
            className="absolute inset-0 bg-black/50"
            onClick={() => setDrawerOpen(false)}
          />
          <div className="absolute inset-y-0 left-0 w-72 max-w-[85%] shadow-xl">
            <button
              type="button"
              onClick={() => setDrawerOpen(false)}
              aria-label="Fechar menu"
              className="absolute right-3 top-3 z-10 flex size-8 items-center justify-center rounded-md text-sidebar-foreground/70 hover:bg-sidebar-accent"
            >
              <X className="size-5" />
            </button>
            {sidebarInner}
          </div>
        </div>
      )}

      <main className="lg:pl-64">
        <div className="pai-fade-up">{children}</div>
      </main>
    </div>
  );
}
