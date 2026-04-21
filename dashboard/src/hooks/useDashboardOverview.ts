"use client";

import { apiGet } from "@/lib/api-client";
import { DATASETS_ENDPOINTS, GAMES_ENDPOINTS, REPORTS_ENDPOINTS } from "@/lib/api/endpoints";

import type { GameStats } from "@/lib/types/games.types";
import { useCallback, useEffect, useState } from "react";

interface DashboardOverviewData {
  games: {
    total: number;
    active: number;
    completed: number;
    abandoned: number;
    popularMode: string | null;
  };
  datasets: {
    questions: number;
    geography: number;
  };
  reports: {
    pending: number;
  };
}

const DEFAULT_DATA: DashboardOverviewData = {
  games: { total: 0, active: 0, completed: 0, abandoned: 0, popularMode: null },
  datasets: { questions: 0, geography: 0 },
  reports: { pending: 0 },
};

export function useDashboardOverview() {
  const [data, setData] = useState<DashboardOverviewData>(DEFAULT_DATA);
  const [loading, setLoading] = useState(true);

  const fetch = useCallback(async () => {
    setLoading(true);
    try {
      const [gamesRes, questionDatasetsRes, geoDatasetsRes, reportsRes] =
        await Promise.allSettled([
          apiGet(GAMES_ENDPOINTS.STATS),
          apiGet(DATASETS_ENDPOINTS.LIST_QUESTIONS("question")),
          apiGet(DATASETS_ENDPOINTS.LIST_QUESTIONS("geography")),
          apiGet(REPORTS_ENDPOINTS.LIST),
        ]);

      let games = DEFAULT_DATA.games;
      if (gamesRes.status === "fulfilled" && gamesRes.value.ok) {
        const json = await gamesRes.value.json();
        const s: GameStats = json.data ?? json;
        games = {
          total: s.total_games,
          active: s.active_games,
          completed: s.completed_games,
          abandoned: s.abandoned_games,
          popularMode: s.most_popular_mode,
        };
      }

      let questionCount = 0;
      if (questionDatasetsRes.status === "fulfilled" && questionDatasetsRes.value.ok) {
        const json = await questionDatasetsRes.value.json();
        questionCount = Array.isArray(json.data) ? json.data.length : 0;
      }

      let geoCount = 0;
      if (geoDatasetsRes.status === "fulfilled" && geoDatasetsRes.value.ok) {
        const json = await geoDatasetsRes.value.json();
        geoCount = Array.isArray(json.data) ? json.data.length : 0;
      }

      let pendingReports = 0;
      if (reportsRes.status === "fulfilled" && reportsRes.value.ok) {
        const json = await reportsRes.value.json();
        pendingReports = json.total ?? json.data?.length ?? 0;
      }

      setData({
        games,
        datasets: { questions: questionCount, geography: geoCount },
        reports: { pending: pendingReports },
      });
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { data, loading };
}
