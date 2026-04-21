"use client";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { useSettings } from "@/hooks/useSettings";
import { SETTINGS_ENDPOINTS } from "@/lib/api/endpoints";
import type { SystemConfig } from "@/lib/types/settings.types";
import { IconRefresh } from "@tabler/icons-react";
import { InfoHover } from "./info-hover";

const DEFAULTS: SystemConfig = {
  user_cache_ttl_minutes: 1440,
  cleanup_interval_minutes: 5,
  offline_delay_seconds: 2,
  game_leave_delay_seconds: 30,
  analytics_active_days: 1,
  analytics_archive_days: 30,
  dataset_check_enabled: false,
  dataset_check_cron: "0 * * * *",
  version_check_enabled: false,
  session_cleanup_enabled: true,
  session_cleanup_cron: "0 * * * *",
  game_cleanup_enabled: true,
  game_cleanup_cron: "*/5 * * * *",
};

function formatDuration(minutes: number, isSeconds?: boolean): string {
  if (isSeconds) {
    if (minutes < 60) return `${minutes} sec`;
    return `${Math.floor(minutes / 60)} min`;
  }
  if (minutes < 60) return `${minutes} min`;
  if (minutes < 1440) return `${Math.floor(minutes / 60)} hr`;
  return `${Math.floor(minutes / 1440)} day`;
}

export function SystemConfigCard() {
  const { config, updateProperty, save, loading, initialLoading } =
    useSettings<SystemConfig>({
      endpoint: SETTINGS_ENDPOINTS.SYSTEM_CONFIG,
      defaults: DEFAULTS,
      validate: (cfg) => {
        if (
          cfg.user_cache_ttl_minutes <= 0 ||
          cfg.cleanup_interval_minutes <= 0
        ) {
          return "Cache TTL and cleanup interval must be positive";
        }
        return true;
      },
    });

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          System Settings
        </CardTitle>
        <CardDescription>
          Configure cache, cleanup, timeouts, analytics retention, and update
          checkers.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div>
          <p className="text-sm font-medium mb-3 flex items-center gap-2">
            Cache
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="user-cache-ttl">User cache TTL</Label>
                <InfoHover description="How long user data is cached in Redis. Higher values reduce DB load but may show stale data." />
              </div>
              <Input
                id="user-cache-ttl"
                type="number"
                min={1}
                value={config.user_cache_ttl_minutes}
                onChange={(e) =>
                  updateProperty(
                    "user_cache_ttl_minutes",
                    Number(e.target.value),
                  )
                }
                disabled={initialLoading}
              />
              <p className="text-xs text-muted-foreground">
                Current: {formatDuration(config.user_cache_ttl_minutes)}
              </p>
            </div>
          </div>
        </div>

        <Separator />

        <div>
          <p className="text-sm font-medium mb-3 flex items-center gap-2">
            Cleanup
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="cleanup-interval">Cleanup interval</Label>
                <InfoHover description="How often the server runs cleanup tasks (remove abandoned games, stale sessions, etc.)." />
              </div>
              <Input
                id="cleanup-interval"
                type="number"
                min={1}
                value={config.cleanup_interval_minutes}
                onChange={(e) =>
                  updateProperty(
                    "cleanup_interval_minutes",
                    Number(e.target.value),
                  )
                }
                disabled={initialLoading}
              />
              <p className="text-xs text-muted-foreground">
                Current: {formatDuration(config.cleanup_interval_minutes)}
              </p>
            </div>
          </div>
        </div>

        <Separator />

        <div>
          <p className="text-sm font-medium mb-3">WebSocket Delays</p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="offline-delay">Offline delay</Label>
                <InfoHover description="Grace period before marking a user as offline after WebSocket disconnect. Prevents false positives during page refreshes or brief network hiccups." />
              </div>
              <Input
                id="offline-delay"
                type="number"
                min={1}
                value={config.offline_delay_seconds}
                onChange={(e) =>
                  updateProperty(
                    "offline_delay_seconds",
                    Number(e.target.value),
                  )
                }
                disabled={initialLoading}
              />
              <p className="text-xs text-muted-foreground">
                Grace period before marking user offline
              </p>
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="game-leave-delay">Game leave delay</Label>
                <InfoHover description="Grace period before treating a disconnected player as having left the game. In solo mode, the game is cancelled; in multiplayer, the opponent wins by default." />
              </div>
              <Input
                id="game-leave-delay"
                type="number"
                min={1}
                value={config.game_leave_delay_seconds}
                onChange={(e) =>
                  updateProperty(
                    "game_leave_delay_seconds",
                    Number(e.target.value),
                  )
                }
                disabled={initialLoading}
              />
            </div>
          </div>
        </div>

        <Separator />

        <div>
          <p className="text-sm font-medium mb-3">Analytics Data Retention</p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="analytics-active">Active data days</Label>
                <InfoHover description="How long to keep 'active' analytics data (current sessions, online users, etc.). After this period, data is moved to archive." />
              </div>
              <Input
                id="analytics-active"
                type="number"
                min={1}
                value={config.analytics_active_days}
                onChange={(e) =>
                  updateProperty(
                    "analytics_active_days",
                    Number(e.target.value),
                  )
                }
                disabled={initialLoading}
              />
              <p className="text-xs text-muted-foreground">
                How long to keep active session data
              </p>
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="analytics-archive">Archive data days</Label>
                <InfoHover description="How long to keep archived analytics data (historical stats, player history, etc.). After this period, data is permanently deleted." />
              </div>
              <Input
                id="analytics-archive"
                type="number"
                min={1}
                value={config.analytics_archive_days}
                onChange={(e) =>
                  updateProperty(
                    "analytics_archive_days",
                    Number(e.target.value),
                  )
                }
                disabled={initialLoading}
              />
            </div>
          </div>
        </div>

        <Separator />

        <div>
          <p className="text-sm font-medium mb-3 flex items-center gap-2">
            Schedulers
          </p>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <div className="flex items-center gap-2">
                  <Label>Session cleanup</Label>
                  <InfoHover description="Periodically removes expired sessions from the database. Disable only for debugging." />
                </div>
                <p className="text-sm text-muted-foreground">
                  Remove expired sessions at regular intervals
                </p>
              </div>
              <Switch
                checked={config.session_cleanup_enabled}
                onCheckedChange={(checked) =>
                  updateProperty("session_cleanup_enabled", checked)
                }
                disabled={initialLoading}
              />
            </div>
            {config.session_cleanup_enabled && (
              <div className="pl-4 grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Label htmlFor="session-cleanup-cron">
                      Cron expression
                    </Label>
                    <InfoHover description="Unix cron format (minute hour day month day-of-week). Example: '0 * * * *' runs every hour, '*/5 * * * *' runs every 5 minutes." />
                  </div>
                  <Input
                    id="session-cleanup-cron"
                    type="text"
                    placeholder="0 * * * *"
                    value={config.session_cleanup_cron}
                    onChange={(e) =>
                      updateProperty("session_cleanup_cron", e.target.value)
                    }
                    disabled={initialLoading}
                    className="font-mono"
                  />
                  <p className="text-xs text-muted-foreground">
                    Format : <code>min hr dom month dow</code> — ex:{" "}
                    <code>0 * * * *</code> = every hour
                  </p>
                </div>
              </div>
            )}

            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <div className="flex items-center gap-2">
                  <Label>Game cleanup</Label>
                  <InfoHover description="Periodically removes abandoned games that were never completed. Disable only for debugging." />
                </div>
                <p className="text-sm text-muted-foreground">
                  Remove abandoned games at regular intervals
                </p>
              </div>
              <Switch
                checked={config.game_cleanup_enabled}
                onCheckedChange={(checked) =>
                  updateProperty("game_cleanup_enabled", checked)
                }
                disabled={initialLoading}
              />
            </div>
            {config.game_cleanup_enabled && (
              <div className="pl-4 grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Label htmlFor="game-cleanup-cron">Cron expression</Label>
                    <InfoHover description="Unix cron format (minute hour day month day-of-week). Example: '0 * * * *' runs every hour, '*/5 * * * *' runs every 5 minutes." />
                  </div>
                  <Input
                    id="game-cleanup-cron"
                    type="text"
                    placeholder="*/5 * * * *"
                    value={config.game_cleanup_cron}
                    onChange={(e) =>
                      updateProperty("game_cleanup_cron", e.target.value)
                    }
                    disabled={initialLoading}
                    className="font-mono"
                  />
                  <p className="text-xs text-muted-foreground">
                    Format : <code>min hr dom month dow</code> — ex:{" "}
                    <code>*/5 * * * *</code> = every 5 minutes
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>

        <Separator />

        <div>
          <p className="text-sm font-medium mb-3 flex items-center gap-2">
            Update Checkers
          </p>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <div className="flex items-center gap-2">
                  <Label>Dataset update check</Label>
                  <InfoHover description="When enabled, the server periodically checks if new question datasets are available and notifies connected admins." />
                </div>
                <p className="text-sm text-muted-foreground">
                  Automatically check for new dataset versions
                </p>
              </div>
              <Switch
                checked={config.dataset_check_enabled}
                onCheckedChange={(checked) =>
                  updateProperty("dataset_check_enabled", checked)
                }
                disabled={initialLoading}
              />
            </div>
            {config.dataset_check_enabled && (
              <div className="pl-4 grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Label htmlFor="dataset-check-cron">Cron expression</Label>
                    <InfoHover description="Unix cron format (minute hour day month day-of-week). Example: '0 * * * *' runs every hour, '*/5 * * * *' runs every 5 minutes." />
                  </div>
                  <Input
                    id="dataset-check-cron"
                    type="text"
                    placeholder="0 * * * *"
                    value={config.dataset_check_cron}
                    onChange={(e) =>
                      updateProperty("dataset_check_cron", e.target.value)
                    }
                    disabled={initialLoading}
                    className="font-mono"
                  />
                  <p className="text-xs text-muted-foreground">
                    Format : <code>min hr dom month dow</code> — ex:{" "}
                    <code>0 * * * *</code> = every hour
                  </p>
                </div>
              </div>
            )}
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <div className="flex items-center gap-2">
                  <Label>Version update check</Label>
                  <InfoHover description="When enabled, clients periodically check the server version and notify admins when a new version is deployed." />
                </div>
                <p className="text-sm text-muted-foreground">
                  Notify when server version changes
                </p>
              </div>
              <Switch
                checked={config.version_check_enabled}
                onCheckedChange={(checked) =>
                  updateProperty("version_check_enabled", checked)
                }
                disabled={initialLoading}
              />
            </div>
          </div>
        </div>

        <Separator />

        <div className="flex justify-end">
          <Button onClick={() => save()} disabled={loading || initialLoading}>
            {loading && <IconRefresh className="h-4 w-4 mr-2 animate-spin" />}
            Save Changes
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
