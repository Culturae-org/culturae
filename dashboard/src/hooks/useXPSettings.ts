"use client";

import { apiGet } from "@/lib/api-client";
import { SETTINGS_ENDPOINTS } from "@/lib/api/endpoints";
import type { RankDefinition, XPConfig } from "@/lib/types/settings.types";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

const DEFAULT_RANKS: RankDefinition[] = [
  { name: "Beginner", min_level: 0 },
  { name: "Intermediate", min_level: 5 },
  { name: "Pro", min_level: 10 },
  { name: "Expert", min_level: 15 },
  { name: "Legend", min_level: 20 },
];

interface UseXPSettingsReturn {
  ranks: RankDefinition[];
  config: XPConfig | null;
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
}

export function useXPSettings(): UseXPSettingsReturn {
  const [ranks, setRanks] = useState<RankDefinition[]>(DEFAULT_RANKS);
  const [config, setConfig] = useState<XPConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchSettings = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiGet(SETTINGS_ENDPOINTS.XP_CONFIG);
      if (res.ok) {
        const json = await res.json();
        const data: XPConfig = json.data ?? json;
        setConfig(data);
        setRanks(
          data.ranks && data.ranks.length > 0 ? data.ranks : DEFAULT_RANKS,
        );
      } else {
        setError("Failed to fetch XP configuration");
        setRanks(DEFAULT_RANKS);
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to fetch XP configuration";
      setError(errorMessage);
      setRanks(DEFAULT_RANKS);
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchSettings();
  }, [fetchSettings]);

  return {
    ranks,
    config,
    loading,
    error,
    refresh: fetchSettings,
  };
}
