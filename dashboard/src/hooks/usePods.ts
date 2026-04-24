"use client";

import { apiGet } from "@/lib/api-client";
import { PODS_ENDPOINTS } from "@/lib/api/endpoints";
import type { PodsDiscovery } from "@/lib/types/metrics.types";
import { useCallback, useEffect, useMemo, useState } from "react";

export function usePods(refreshInterval = 5000) {
  const [podsDiscovery, setPodsDiscovery] = useState<PodsDiscovery | null>(null);
  const [loading, setLoading] = useState(true);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [error, setError] = useState<string | null>(null);

  const loadPods = useCallback(async () => {
    try {
      setError(null);
      const response = await apiGet(PODS_ENDPOINTS.LIST);
      if (response.ok) {
        const data = await response.json();
        setPodsDiscovery(data.data ?? data);
        setLastUpdated(new Date());
      } else {
        setError(`HTTP ${response.status}`);
      }
    } catch (err) {
      console.error("Error loading pods:", err);
      setError("Failed to load pods");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadPods();

    let interval: ReturnType<typeof setInterval> | null = setInterval(loadPods, refreshInterval);

    const handleVisibilityChange = () => {
      if (document.hidden) {
        if (interval) {
          clearInterval(interval);
          interval = null;
        }
      } else {
        loadPods();
        interval = setInterval(loadPods, refreshInterval);
      }
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      if (interval) clearInterval(interval);
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, [loadPods, refreshInterval]);

  const mainPods = useMemo(
    () => podsDiscovery?.pods.filter((p) => p.pod_type === "main") ?? [],
    [podsDiscovery]
  );

  const headlessPods = useMemo(
    () => podsDiscovery?.pods.filter((p) => p.pod_type === "headless") ?? [],
    [podsDiscovery]
  );

  const totalActiveGames = useMemo(
    () => podsDiscovery?.pods.reduce((acc, p) => acc + (p.active_games ?? 0), 0) ?? 0,
    [podsDiscovery]
  );

  const totalLoadScore = useMemo(
    () =>
      podsDiscovery?.pods.reduce(
        (acc, p) => acc + p.connected_clients + (p.active_games ?? 0) * 2,
        0
      ) ?? 0,
    [podsDiscovery]
  );

  return {
    podsDiscovery,
    loading,
    lastUpdated,
    error,
    mainPods,
    headlessPods,
    totalActiveGames,
    totalLoadScore,
    refresh: loadPods,
  };
}
