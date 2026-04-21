"use client";

import { logsService } from "@/lib/services/logs.service";
import { useCallback, useState } from "react";

export function useApiLogs() {
  const [timestampsLoading, setTimestampsLoading] = useState(false);
  const [timestampsError, setTimestampsError] = useState<string | null>(null);
  const [statsLoading, setStatsLoading] = useState(false);
  const [statsError, setStatsError] = useState<string | null>(null);

  const fetchRequestTimestamps = useCallback(
    async (
      method?: string,
      statusCode?: string,
      startDate?: Date,
      endDate?: Date,
    ) => {
      setTimestampsLoading(true);
      setTimestampsError(null);

      try {
        const params: Record<string, string> = {};
        if (method && method !== "all") params.method = method;
        if (statusCode && statusCode !== "all") params.status_code = statusCode;
        if (startDate) params.start_date = startDate.toISOString();
        if (endDate) params.end_date = endDate.toISOString();

        return await logsService.getAPIRequestTimestamps(params);
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to fetch timestamps";
        setTimestampsError(message);
        throw err;
      } finally {
        setTimestampsLoading(false);
      }
    },
    [],
  );

  const fetchRequestStats = useCallback(
    async (startDate?: Date, endDate?: Date) => {
      setStatsLoading(true);
      setStatsError(null);

      try {
        const params: Record<string, string> = {};
        if (startDate) params.start_date = startDate.toISOString();
        if (endDate) params.end_date = endDate.toISOString();

        return await logsService.getAPIRequestStats(params);
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to fetch stats";
        setStatsError(message);
        throw err;
      } finally {
        setStatsLoading(false);
      }
    },
    [],
  );

  return {
    timestampsLoading,
    timestampsError,
    statsLoading,
    statsError,
    fetchRequestTimestamps,
    fetchRequestStats,
  };
}
