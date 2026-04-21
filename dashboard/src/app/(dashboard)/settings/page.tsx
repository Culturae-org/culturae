"use client";

import { AuthConfigCard } from "@/components/settings/auth-config-card";
import { AvatarConfigCard } from "@/components/settings/avatar-config-card";
import { CacheManagementCard } from "@/components/settings/cache-management-card";
import { EloConfigCard } from "@/components/settings/elo-config-card";
import { GameConfigCard } from "@/components/settings/game-config-card";
import { GameCountdownConfigCard } from "@/components/settings/game-countdown-config-card";
import { MaintenanceModeCard } from "@/components/settings/maintenance-mode-card";
import { PlatformInfoCard } from "@/components/settings/platform-info-card";
import { RateLimitCard } from "@/components/settings/rate-limit-card";
import { SystemConfigCard } from "@/components/settings/system-config-card";
import { WebSocketConfigCard } from "@/components/settings/websocket-config-card";
import { XpConfigCard } from "@/components/settings/xp-config-card";
import { useVersionCheck } from "@/hooks/useVersionCheck";
import { apiGet } from "@/lib/api-client";
import { SETTINGS_ENDPOINTS } from "@/lib/api/endpoints";
import * as React from "react";
import { useEffect } from "react";
import { useSearchParams } from "react-router";

export default function SettingsPage() {
  const [maintenanceMode, setMaintenanceMode] = React.useState(false);
  const [maintenanceLoading, setMaintenanceLoading] = React.useState(false);
  const [cacheLoading, setCacheLoading] = React.useState(false);
  const [initialLoading, setInitialLoading] = React.useState(true);
  const { version, buildTime, environment } = useVersionCheck();
  const [_searchParams] = useSearchParams();

  React.useEffect(() => {
    async function load() {
      try {
        const maintRes = await apiGet(SETTINGS_ENDPOINTS.MAINTENANCE).catch(
          () => null,
        );
        if (maintRes?.ok) {
          const data = await maintRes.json();
          setMaintenanceMode(!!data.enabled);
        }
      } catch {
      } finally {
        setInitialLoading(false);
      }
    }
    load();
  }, []);

  useEffect(() => {
    const hash = window.location.hash.slice(1);
    if (hash) {
      const element = document.getElementById(hash);
      if (element) {
        setTimeout(() => {
          element.scrollIntoView({ behavior: "smooth" });
        }, 100);
      }
    }
  }, []);

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <h1 className="text-3xl font-bold flex items-center">Settings</h1>
        <p className="text-muted-foreground mt-1">
          Server configuration and administration
        </p>
      </div>

      <div className="px-4 lg:px-6">
        <ul className="flex flex-wrap gap-x-6 gap-y-2 text-sm text-muted-foreground">
          {[
            { label: "Maintenance", id: "maintenance" },
            { label: "Cache", id: "cache" },
            { label: "Rate Limit", id: "rate-limit" },
            { label: "Authentication", id: "auth" },
            { label: "WebSocket", id: "websocket" },
            { label: "Avatar", id: "avatar" },
            { label: "XP", id: "xp" },
            { label: "ELO", id: "elo" },
            { label: "Game TTL", id: "game" },
            { label: "Game Countdown", id: "game-countdown" },
            { label: "System", id: "system" },
            { label: "Platform", id: "platform" },
          ].map((item) => (
            <li key={item.id} className="flex items-center gap-2">
              <span className="h-1 w-1 rounded-full bg-muted-foreground/40" />
              <a
                href={`#${item.id}`}
                className="hover:text-foreground hover:underline underline-offset-4 transition-colors"
              >
                {item.label}
              </a>
            </li>
          ))}
        </ul>
      </div>

      <div className="px-4 lg:px-6 space-y-6">
        <div id="maintenance" className="scroll-mt-20">
          <MaintenanceModeCard
            maintenanceMode={maintenanceMode}
            loading={maintenanceLoading}
            initialLoading={initialLoading}
            onMaintenanceModeChange={setMaintenanceMode}
            onLoadingChange={setMaintenanceLoading}
          />
        </div>

        <div id="cache" className="scroll-mt-20">
          <CacheManagementCard
            loading={cacheLoading}
            onLoadingChange={setCacheLoading}
          />
        </div>

        <div id="rate-limit" className="scroll-mt-20">
          <RateLimitCard />
        </div>

        <div id="auth" className="scroll-mt-20">
          <AuthConfigCard />
        </div>

        <div id="websocket" className="scroll-mt-20">
          <WebSocketConfigCard />
        </div>

        <div id="avatar" className="scroll-mt-20">
          <AvatarConfigCard />
        </div>

        <div id="xp" className="scroll-mt-20">
          <XpConfigCard />
        </div>

        <div id="elo" className="scroll-mt-20">
          <EloConfigCard />
        </div>

        <div id="game" className="scroll-mt-20">
          <GameConfigCard />
        </div>

        <div id="game-countdown" className="scroll-mt-20">
          <GameCountdownConfigCard />
        </div>

        <div id="system" className="scroll-mt-20">
          <SystemConfigCard />
        </div>

        <div id="platform" className="scroll-mt-20">
          <PlatformInfoCard
            version={version}
            buildTime={buildTime}
            environment={environment}
          />
        </div>
      </div>
    </div>
  );
}
