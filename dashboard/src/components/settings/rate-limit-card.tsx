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
import type { RateLimitConfig } from "@/lib/types/settings.types";
import { IconRefresh } from "@tabler/icons-react";
import { InfoHover } from "./info-hover";

const DEFAULTS: RateLimitConfig = {
  enabled: false,
  apply_to_admin: false,
  max_requests: 100,
  window_seconds: 60,
};

export function RateLimitCard() {
  const { config, updateProperty, save, loading, initialLoading } =
    useSettings<RateLimitConfig>({
      endpoint: SETTINGS_ENDPOINTS.RATE_LIMIT,
      defaults: DEFAULTS,
      validate: (cfg) => {
        if (cfg.enabled && (cfg.max_requests <= 0 || cfg.window_seconds <= 0)) {
          return "Max requests and window must be positive when rate limiting is enabled";
        }
        return true;
      },
    });

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">Rate Limiting</CardTitle>
        <CardDescription>
          Configure the global API rate limiting policy. Admin routes are always
          exempt. Changes take effect immediately.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <div className="flex items-center gap-2">
              <Label>Enable rate limiting</Label>
              <InfoHover description="When enabled, the API will limit the number of requests per IP address. This protects against brute force attacks and abuse." />
            </div>
            <p className="text-xs text-muted-foreground">
              Limit the number of API requests per IP address
            </p>
          </div>
          <Switch
            checked={config.enabled}
            onCheckedChange={(v) => updateProperty("enabled", v)}
            disabled={initialLoading}
          />
        </div>

        {config.enabled && (
          <>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <div className="flex items-center gap-2">
                  <Label>Apply to admin routes</Label>
                  <InfoHover description="When enabled, rate limiting also applies to authenticated admin API calls. When disabled, admins bypass rate limits." />
                </div>
                <p className="text-xs text-muted-foreground">
                  Also rate-limit authenticated admin API calls
                </p>
              </div>
              <Switch
                checked={config.apply_to_admin}
                onCheckedChange={(v) => updateProperty("apply_to_admin", v)}
              />
            </div>
            <Separator />
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Label htmlFor="max-requests">Max requests per window</Label>
                  <InfoHover description="Maximum number of API requests allowed from a single IP within the time window." />
                </div>
                <Input
                  id="max-requests"
                  type="number"
                  min={1}
                  value={config.max_requests}
                  onChange={(e) =>
                    updateProperty("max_requests", Number(e.target.value))
                  }
                />
                <p className="text-xs text-muted-foreground">
                  Number of requests allowed per time window
                </p>
              </div>
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Label htmlFor="window-seconds">Window (seconds)</Label>
                  <InfoHover description="Time window duration for rate limit counting. Requests exceeding the limit within this period are blocked." />
                </div>
                <Input
                  id="window-seconds"
                  type="number"
                  min={1}
                  value={config.window_seconds}
                  onChange={(e) =>
                    updateProperty("window_seconds", Number(e.target.value))
                  }
                />
              </div>
            </div>
          </>
        )}

        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            {config.enabled
              ? `Active — ${config.max_requests} requests / ${config.window_seconds}s`
              : "Disabled — no request limiting"}
          </p>
          <Button onClick={() => save()} disabled={loading}>
            {loading ? (
              <IconRefresh className="h-4 w-4 mr-2 animate-spin" />
            ) : null}
            Save Changes
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
