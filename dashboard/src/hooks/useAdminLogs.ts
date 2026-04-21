"use client";

import { logsService } from "@/lib/services/logs.service";
import { useCallback, useState } from "react";

export function useAdminLogs() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchAdminLogs = useCallback(
    async (
      page: number,
      limit: number,
      action?: string,
      resource?: string,
      status?: string,
    ) => {
      setLoading(true);
      setError(null);

      try {
        const params: Record<string, string | number | boolean> = {
          page,
          limit,
        };
        if (action) params.action = action;
        if (resource && resource !== "all") params.resource = resource;
        if (status && status !== "all") {
          params.is_success = status === "success";
        }

        const result = await logsService.getAdminActionLogs(params);

        return {
          logs: result.data,
          total_count: result.total,
          total_pages: result.total_pages,
        };
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to fetch admin logs";
        setError(message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  return {
    loading,
    error,
    fetchAdminLogs,
  };
}
