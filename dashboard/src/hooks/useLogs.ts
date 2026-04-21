"use client";

import { logsService } from "@/lib/services/logs.service";
import type { UserActionLog } from "@/lib/types/logs.types";
import type {
  ServiceStatusResponse,
  SystemMetrics,
} from "@/lib/types/stats.types";
import type { ConnectionLog, Session } from "@/lib/types/user.types";
import { useCallback, useEffect, useState } from "react";

export function useLogsStats() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchAPIRequestStats = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      return await logsService.getAPIRequestStats();
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to fetch stats";
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchAdminActionStats = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      return await logsService.getAdminActionStats();
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to fetch stats";
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchUserActionStats = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      return await logsService.getUserActionStats();
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to fetch stats";
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    loading,
    error,
    fetchAPIRequestStats,
    fetchAdminActionStats,
    fetchUserActionStats,
  };
}

export function useSystemMonitoring() {
  const [systemMetrics, setSystemMetrics] = useState<SystemMetrics | null>(
    null,
  );
  const [serviceStatus, setServiceStatus] =
    useState<ServiceStatusResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchSystemMetrics = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const metrics = await logsService.getSystemMetrics();
      setSystemMetrics(metrics);
      return metrics;
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to fetch metrics";
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchServiceStatus = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const status = await logsService.getServiceStatus();
      setServiceStatus(status);
      return status;
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to fetch status";
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    systemMetrics,
    serviceStatus,
    loading,
    error,
    fetchSystemMetrics,
    fetchServiceStatus,
    clearError: () => setError(null),
  };
}

export function useUserActionLogs(userId: string) {
  const [logs, setLogs] = useState<UserActionLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchLogs = useCallback(
    async (isRefresh = false) => {
      if (!userId) return;
      try {
        isRefresh ? setRefreshing(true) : setLoading(true);
        setError(null);
        const data = await logsService.getUserActionLogs(userId);
        setLogs(
          data.sort(
            (a, b) =>
              new Date(b.CreatedAt).getTime() - new Date(a.CreatedAt).getTime(),
          ),
        );
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to load action logs",
        );
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [userId],
  );

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const refresh = useCallback(() => fetchLogs(true), [fetchLogs]);

  return { logs, loading, refreshing, error, refresh };
}

export function useUserConnectionLogs(userId: string) {
  const [logs, setLogs] = useState<ConnectionLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchLogs = useCallback(
    async (isRefresh = false) => {
      if (!userId) return;
      try {
        isRefresh ? setRefreshing(true) : setLoading(true);
        setError(null);
        const data = await logsService.getUserConnectionLogs(userId);
        setLogs(
          data.sort(
            (a, b) =>
              new Date(b.created_at).getTime() -
              new Date(a.created_at).getTime(),
          ),
        );
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to load connection logs",
        );
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [userId],
  );

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const refresh = useCallback(() => fetchLogs(true), [fetchLogs]);

  return { logs, loading, refreshing, error, refresh };
}

export function useUserSessions(userId: string) {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchSessions = useCallback(
    async (isRefresh = false) => {
      if (!userId) return;
      try {
        isRefresh ? setRefreshing(true) : setLoading(true);
        setError(null);
        const data = await logsService.getUserActiveSessions(userId);
        setSessions(
          data.sort(
            (a, b) =>
              new Date(b.created_at).getTime() -
              new Date(a.created_at).getTime(),
          ),
        );
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to load sessions",
        );
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [userId],
  );

  useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  const refresh = useCallback(() => fetchSessions(true), [fetchSessions]);

  return { sessions, loading, refreshing, error, refresh };
}
