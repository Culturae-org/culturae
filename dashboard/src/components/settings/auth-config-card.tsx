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
import { useSettings } from "@/hooks/useSettings";
import { SETTINGS_ENDPOINTS } from "@/lib/api/endpoints";
import type { AuthConfig } from "@/lib/types/settings.types";
import { IconRefresh } from "@tabler/icons-react";
import { InfoHover } from "./info-hover";

const DEFAULTS: AuthConfig = {
  access_token_ttl_minutes: 15,
  refresh_token_ttl_days: 7,
  session_ttl_days: 30,
  max_concurrent_sessions: 5,
  failed_login_attempts: 5,
  login_lockout_minutes: 15,
};

export function AuthConfigCard() {
  const { config, updateProperty, save, loading, initialLoading } =
    useSettings<AuthConfig>({
      endpoint: SETTINGS_ENDPOINTS.AUTH_CONFIG,
      defaults: DEFAULTS,
      validate: (cfg) => {
        if (
          cfg.access_token_ttl_minutes <= 0 ||
          cfg.refresh_token_ttl_days <= 0 ||
          cfg.session_ttl_days <= 0
        ) {
          return "All TTL values must be positive";
        }
        if (cfg.max_concurrent_sessions <= 0) {
          return "max_concurrent_sessions must be positive";
        }
        if (cfg.failed_login_attempts <= 0 || cfg.login_lockout_minutes <= 0) {
          return "failed_login_attempts and login_lockout_minutes must be positive";
        }
        return true;
      },
    });

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center">
          Authentication Settings
        </CardTitle>
        <CardDescription>
          Configure token lifetimes, session management, and login security
          policies.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div>
          <div className="flex items-center gap-2 mb-3">
            <p className="text-sm font-medium">Token Lifetimes</p>
            <InfoHover description="Shorter lifetimes improve security but require more frequent refreshes. Longer lifetimes reduce load but are less secure." />
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="access-token-ttl">Access token lifetime</Label>
                <InfoHover description="How long an access token remains valid. Shorter values require frequent refreshes." />
              </div>
              <Input
                id="access-token-ttl"
                type="number"
                min={1}
                max={1440}
                value={config.access_token_ttl_minutes}
                onChange={(e) =>
                  updateProperty(
                    "access_token_ttl_minutes",
                    Number(e.target.value),
                  )
                }
                disabled={initialLoading}
              />
              <p className="text-xs text-muted-foreground">Minutes</p>
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="refresh-token-ttl">
                  Refresh token lifetime
                </Label>
                <InfoHover description="How long a refresh token remains valid. After this period, users must re-authenticate." />
              </div>
              <Input
                id="refresh-token-ttl"
                type="number"
                min={1}
                max={365}
                value={config.refresh_token_ttl_days}
                onChange={(e) =>
                  updateProperty(
                    "refresh_token_ttl_days",
                    Number(e.target.value),
                  )
                }
                disabled={initialLoading}
              />
              <p className="text-xs text-muted-foreground">Days</p>
            </div>
          </div>
        </div>

        <Separator />

        <div>
          <div className="flex items-center gap-2 mb-3">
            <p className="text-sm font-medium">Session Management</p>
            <InfoHover description="Configure how long sessions are active and how many concurrent sessions per user." />
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="session-ttl">Session lifetime</Label>
                <InfoHover description="How long a session remains active without user activity. After this period, the user must log in again." />
              </div>
              <Input
                id="session-ttl"
                type="number"
                min={1}
                max={365}
                value={config.session_ttl_days}
                onChange={(e) =>
                  updateProperty("session_ttl_days", Number(e.target.value))
                }
                disabled={initialLoading}
              />
              <p className="text-xs text-muted-foreground">Days</p>
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="max-sessions">Max concurrent sessions</Label>
                <InfoHover description="Maximum number of simultaneous sessions per user across devices. Oldest session is invalidated when limit is exceeded." />
              </div>
              <Input
                id="max-sessions"
                type="number"
                min={1}
                max={20}
                value={config.max_concurrent_sessions}
                onChange={(e) =>
                  updateProperty(
                    "max_concurrent_sessions",
                    Number(e.target.value),
                  )
                }
                disabled={initialLoading}
              />
              <p className="text-xs text-muted-foreground">Sessions per user</p>
            </div>
          </div>
        </div>

        <Separator />

        <div>
          <div className="flex items-center gap-2 mb-3">
            <p className="text-sm font-medium">Login Security</p>
            <InfoHover description="Configure brute force protection policies." />
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="failed-attempts">
                  Failed attempts before lockout
                </Label>
                <InfoHover description="Number of consecutive failed login attempts before the account is temporarily locked." />
              </div>
              <Input
                id="failed-attempts"
                type="number"
                min={1}
                max={50}
                value={config.failed_login_attempts}
                onChange={(e) =>
                  updateProperty(
                    "failed_login_attempts",
                    Number(e.target.value),
                  )
                }
                disabled={initialLoading}
              />
              <p className="text-xs text-muted-foreground">Attempts</p>
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="lockout-duration">Lockout duration</Label>
                <InfoHover description="How long an account is locked after exceeding failed login attempts." />
              </div>
              <Input
                id="lockout-duration"
                type="number"
                min={1}
                max={1440}
                value={config.login_lockout_minutes}
                onChange={(e) =>
                  updateProperty(
                    "login_lockout_minutes",
                    Number(e.target.value),
                  )
                }
                disabled={initialLoading}
              />
              <p className="text-xs text-muted-foreground">Minutes</p>
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
