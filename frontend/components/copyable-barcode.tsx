"use client";

import { useState, useCallback, useRef, useEffect } from "react";

export function CopyableBarcode({ code }: { code: string }) {
  const [toast, setToast] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => () => {
    if (timer.current) clearTimeout(timer.current);
  }, []);

  const copy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(code);
    } catch {
      // fallback for non-secure contexts
      const ta = document.createElement("textarea");
      ta.value = code;
      ta.style.position = "fixed";
      ta.style.opacity = "0";
      document.body.appendChild(ta);
      ta.select();
      document.execCommand("copy");
      document.body.removeChild(ta);
    }
    setToast(true);
    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => setToast(false), 2500);
  }, [code]);

  return (
    <>
      <code
        role="button"
        tabIndex={0}
        onClick={copy}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            copy();
          }
        }}
        title="Clique para copiar o código de barras"
        className="block text-xs font-mono break-all bg-[var(--bg-soft)] rounded px-3 py-2 cursor-pointer transition-colors hover:brightness-95 active:brightness-90 focus:outline-none focus:ring-2 focus:ring-offset-1"
      >
        {code}
      </code>

      {toast && (
        <div
          role="status"
          aria-live="polite"
          className="fixed bottom-4 right-4 z-50 rounded-md bg-[var(--bg-soft)] border px-4 py-2 text-sm shadow-lg animate-in fade-in slide-in-from-bottom-2"
        >
          Código de barras copiado ✓
        </div>
      )}
    </>
  );
}
