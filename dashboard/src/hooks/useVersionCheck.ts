"use client";

import { apiGet } from "@/lib/api-client";
import { INFO_ENDPOINT, SETTINGS_ENDPOINTS } from "@/lib/api/endpoints";
import { useCallback, useEffect, useState } from "react";

interface InfoResponse {
  service: string;
  version: string;
  build_time: string;
  environment: string;
}

interface VersionStatusResponse {
  current_version: string;
  latest_version: string;
  is_up_to_date: boolean;
  checked_at: string | null;
}

export function useVersionCheck() {
  const [version, setVersion] = useState<string | null>(null);
  const [buildTime, setBuildTime] = useState<string | null>(null);
  const [environment, setEnvironment] = useState<string | null>(null);
  const [updateAvailable, setUpdateAvailable] = useState(false);
  const [latestVersion, setLatestVersion] = useState<string | null>(null);

  const applyVersionStatus = useCallback((data: VersionStatusResponse) => {
    setUpdateAvailable(!data.is_up_to_date);
    if (!data.is_up_to_date && data.latest_version) {
      setLatestVersion(data.latest_version);
    }
  }, []);

  useEffect(() => {
    apiGet(INFO_ENDPOINT)
      .then((res) => res.ok ? res.json() : null)
      .then((json) => {
        if (!json) return;
        const data: InfoResponse = json.data;
        setVersion(data.version);
        setBuildTime(data.build_time);
        setEnvironment(data.environment);
      })
      .catch(() => {});

    apiGet(SETTINGS_ENDPOINTS.VERSION_STATUS)
      .then((res) => res.ok ? res.json() : null)
      .then((json) => {
        if (!json) return;
        applyVersionStatus(json.data);
      })
      .catch(() => {});
  }, [applyVersionStatus]);

  useEffect(() => {
    const handler = (e: Event) => {
      const data = (e as CustomEvent).detail;
      if (data?.type !== "admin_notification" || data?.event !== "version_status_updated") return;
      applyVersionStatus(data.data as VersionStatusResponse);
    };
    window.addEventListener("realtime-message", handler);
    return () => window.removeEventListener("realtime-message", handler);
  }, [applyVersionStatus]);

  return { version, buildTime, environment, updateAvailable, latestVersion };
}
