"use client";

import { apiGet } from "@/lib/api-client";
import { GAMES_ENDPOINTS } from "@/lib/api/endpoints";
import type {
  GameDailyStats,
  GameModeStats,
  GameStats,
} from "@/lib/types/games.types";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

interface GamesChartsData {
  stats: GameStats;
  modes: GameModeStats[];
  daily: GameDailyStats[];
}

function parseModes(raw: unknown): GameModeStats[] {
  if (Array.isArray(raw)) return raw as GameModeStats[];
  if (raw && typeof raw === "object") {
    return Object.entries(raw as Record<string, number>).map(
      ([mode, count]) => ({ mode, count }),
    );
  }
  return [];
}

export function useGamesCharts() {
  const [data, setData] = useState<GamesChartsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetch = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [statsRes, modesRes, dailyRes] = await Promise.all([
        apiGet(GAMES_ENDPOINTS.STATS),
        apiGet(GAMES_ENDPOINTS.MODE_STATS),
        apiGet(GAMES_ENDPOINTS.DAILY_STATS),
      ]);

      if (!statsRes.ok || !modesRes.ok || !dailyRes.ok) {
        throw new Error(
          `HTTP Error: stats=${statsRes.status}, modes=${modesRes.status}, daily=${dailyRes.status}`,
        );
      }

      const statsData = await statsRes.json();
      const modesData = await modesRes.json();
      const dailyData = await dailyRes.json();

      const stats = statsData.data ?? statsData;
      const modesRaw = modesData.data ?? modesData;
      const daily = dailyData.data ?? dailyData;

      if (!stats || !Array.isArray(daily)) {
        throw new Error(
          "Invalid response: missing required data (stats or daily)",
        );
      }

      setData({ stats, modes: parseModes(modesRaw), daily });
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to fetch games charts";
      setError(errorMessage);
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { data, loading, error, refresh: fetch };
}

export function useGamesStats() {
  const [stats, setStats] = useState<GameStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetch = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await apiGet(GAMES_ENDPOINTS.STATS);
      if (response.ok) {
        const result = await response.json();
        setStats(result.data ?? result);
      } else {
        throw new Error(`HTTP ${response.status}`);
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to fetch games stats";
      setError(errorMessage);
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { stats, loading, error, refresh: fetch };
}

export function useGamesOverview() {
  const [overview, setOverview] = useState<GameStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetch = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await apiGet(GAMES_ENDPOINTS.LIST);
      if (response.ok) {
        const result = await response.json();
        setOverview(result.data ?? result);
      } else {
        throw new Error(`HTTP ${response.status}`);
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to fetch games overview";
      setError(errorMessage);
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { overview, loading, error, refresh: fetch };
}
