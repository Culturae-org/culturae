"use client";

import { apiGet } from "@/lib/api-client";
import { GAMES_ENDPOINTS } from "@/lib/api/endpoints";
import { useCallback, useEffect, useState } from "react";
export type TimeRange = "7d" | "30d" | "90d" | "all";

interface GameChartDataPoint {
  date: string;
  total_games: number;
  completed_games: number;
  cancelled_games: number;
  total_players?: number;
}

interface UseDashboardGamesDataResult {
  data: GameChartDataPoint[];
  loading: boolean;
  refreshing: boolean;
  timeRange: TimeRange;
  gameMode: string;
  setTimeRange: (range: TimeRange) => void;
  setGameMode: (mode: string) => void;
  refresh: () => void;
}

export function useDashboardGamesData(): UseDashboardGamesDataResult {
  const [data, setData] = useState<GameChartDataPoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [timeRange, setTimeRange] = useState<TimeRange>("30d");
  const [gameMode, setGameMode] = useState<string>("all");

  const fetchData = useCallback(
    async (isRefresh = false) => {
      try {
        if (isRefresh) {
          setRefreshing(true);
        } else {
          setLoading(true);
        }

        const params = new URLSearchParams();
        const now = new Date();
        let startDateCalc = new Date(now);
        const endDateCalc = now;

        if (timeRange !== "all") {
          let daysToSubtract = 30;
          if (timeRange === "7d") daysToSubtract = 7;
          else if (timeRange === "90d") daysToSubtract = 90;

          startDateCalc = new Date(now);
          startDateCalc.setDate(startDateCalc.getDate() - daysToSubtract);
          params.set("start_date", startDateCalc.toISOString().split("T")[0]);
          params.set("end_date", now.toISOString().split("T")[0]);
        } else {
          startDateCalc = new Date(now);
          startDateCalc.setDate(startDateCalc.getDate() - 30);
          params.set("start_date", startDateCalc.toISOString().split("T")[0]);
          params.set("end_date", now.toISOString().split("T")[0]);
        }

        if (gameMode !== "all") {
          params.set("mode", gameMode);
        }

        const url = `${GAMES_ENDPOINTS.DAILY_STATS}?${params.toString()}`;
        const res = await apiGet(url);
        if (res.ok) {
          const json = await res.json();
          const rawData = json.data ?? json;

          const dataByDate: Record<string, GameChartDataPoint> = {};
          for (const item of rawData) {
            dataByDate[item.date] = item;
          }

          const filledData: GameChartDataPoint[] = [];
          const current = new Date(startDateCalc);
          const end = new Date(endDateCalc);
          while (current <= end) {
            const dateKey = current.toISOString().split("T")[0];
            filledData.push(
              dataByDate[dateKey] || {
                date: dateKey,
                total_games: 0,
                completed_games: 0,
                cancelled_games: 0,
                total_players: 0,
              },
            );
            current.setDate(current.getDate() + 1);
          }

          setData(filledData);
        }
      } catch (e) {
        console.error("Failed to fetch games chart:", e);
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [timeRange, gameMode],
  );

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  return {
    data,
    loading,
    refreshing,
    timeRange,
    gameMode,
    setTimeRange,
    setGameMode,
    refresh: () => fetchData(true),
  };
}
