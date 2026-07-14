"use client";

import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import type { Me } from "@/lib/types";

export interface UseMeResult {
  me: Me | null;
  loading: boolean;
  error: Error | null;
}

/**
 * Fetches the current authenticated teacher (GET /api/me). `me` stays null
 * while loading or when the request fails (e.g. 401 when not logged in).
 */
export function useMe(): UseMeResult {
  const [me, setMe] = useState<Me | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    apiFetch<Me>("/api/me")
      .then((data) => {
        if (!cancelled) {
          setMe(data);
          setError(null);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setMe(null);
          setError(err instanceof Error ? err : new Error(String(err)));
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return { me, loading, error };
}
