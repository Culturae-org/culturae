"use client";

import { Badge } from "@/components/ui/badge";
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
import type { WebSocketConfig } from "@/lib/types/settings.types";
import { IconPlus, IconRefresh, IconX } from "@tabler/icons-react";
import * as React from "react";
import { InfoHover } from "./info-hover";

const DEFAULTS: WebSocketConfig = {
  write_wait_seconds: 10,
  pong_wait_seconds: 60,
  max_message_size_kb: 512,
  allowed_origins: [],
  reconnect_grace_period_seconds: 180,
  message_rate_limit: 0,
  message_rate_window_seconds: 60,
};

export function WebSocketConfigCard() {
  const { config, updateProperty, save, loading, initialLoading } =
    useSettings<WebSocketConfig>({
      endpoint: SETTINGS_ENDPOINTS.WEBSOCKET,
      defaults: DEFAULTS,
      validate: (cfg) => {
        if (
          cfg.write_wait_seconds <= 0 ||
          cfg.pong_wait_seconds <= 0 ||
          cfg.max_message_size_kb <= 0
        ) {
          return "Write wait, pong wait and max message size must be positive";
        }
        if (cfg.reconnect_grace_period_seconds <= 0) {
          return "Reconnect grace period must be positive";
        }
        if (
          cfg.message_rate_limit > 0 &&
          cfg.message_rate_window_seconds <= 0
        ) {
          return "Rate limit window must be positive when rate limiting is enabled";
        }
        return true;
      },
    });

  const [originInput, setOriginInput] = React.useState("");

  const addOrigin = React.useCallback(() => {
    const trimmed = originInput.trim();
    if (!trimmed) return;
    if (config.allowed_origins.includes(trimmed)) {
      setOriginInput("");
      return;
    }
    updateProperty("allowed_origins", [...config.allowed_origins, trimmed]);
    setOriginInput("");
  }, [originInput, config.allowed_origins, updateProperty]);

  const removeOrigin = React.useCallback(
    (origin: string) => {
      updateProperty(
        "allowed_origins",
        config.allowed_origins.filter((o) => o !== origin),
      );
    },
    [config.allowed_origins, updateProperty],
  );

  const rateLimitEnabled = config.message_rate_limit > 0;

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle>WebSocket Settings</CardTitle>
        <CardDescription>
          Configure WebSocket connection parameters. Changes apply to new
          connections only.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Connection timings */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Label htmlFor="ws-write-wait">Write wait</Label>
              <InfoHover description="Maximum time the server will wait when writing a message to the client. If the write blocks longer than this, the connection is closed." />
            </div>
            <Input
              id="ws-write-wait"
              type="number"
              min={1}
              value={config.write_wait_seconds}
              onChange={(e) =>
                updateProperty("write_wait_seconds", Number(e.target.value))
              }
              disabled={initialLoading}
            />
            <p className="text-xs text-muted-foreground">seconds</p>
          </div>
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Label htmlFor="ws-pong-wait">Pong wait</Label>
              <InfoHover description="Time the server waits for a pong response after sending a ping. If no pong is received, the connection is considered dead." />
            </div>
            <Input
              id="ws-pong-wait"
              type="number"
              min={1}
              value={config.pong_wait_seconds}
              onChange={(e) =>
                updateProperty("pong_wait_seconds", Number(e.target.value))
              }
              disabled={initialLoading}
            />
            <p className="text-xs text-muted-foreground">seconds</p>
          </div>
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Label htmlFor="ws-max-size">Max message size</Label>
              <InfoHover description="Maximum size allowed for incoming WebSocket messages. Messages larger than this will be rejected." />
            </div>
            <Input
              id="ws-max-size"
              type="number"
              min={1}
              value={config.max_message_size_kb}
              onChange={(e) =>
                updateProperty("max_message_size_kb", Number(e.target.value))
              }
              disabled={initialLoading}
            />
            <p className="text-xs text-muted-foreground">KB</p>
          </div>
        </div>

        <Separator />

        <div className="space-y-2 max-w-xs">
          <div className="flex items-center gap-2">
            <Label htmlFor="ws-reconnect-grace">Reconnect grace period</Label>
            <InfoHover description="Time the server waits before removing a disconnected player from an in-progress game. During this window the player can reconnect and resume." />
          </div>
          <Input
            id="ws-reconnect-grace"
            type="number"
            min={1}
            value={config.reconnect_grace_period_seconds}
            onChange={(e) =>
              updateProperty(
                "reconnect_grace_period_seconds",
                Number(e.target.value),
              )
            }
            disabled={initialLoading}
          />
          <p className="text-xs text-muted-foreground">seconds</p>
        </div>

        <Separator />

        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <div className="flex items-center gap-2">
                <Label>Message rate limiting</Label>
                <InfoHover description="Limits the number of WebSocket messages a single client can send within a time window. Excess messages are dropped with an error ack." />
              </div>
              <p className="text-xs text-muted-foreground">
                {rateLimitEnabled
                  ? `${config.message_rate_limit} messages / ${config.message_rate_window_seconds}s per client`
                  : "Disabled"}
              </p>
            </div>
            <Switch
              checked={rateLimitEnabled}
              onCheckedChange={(checked) =>
                updateProperty("message_rate_limit", checked ? 60 : 0)
              }
              disabled={initialLoading}
            />
          </div>
          {rateLimitEnabled && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pl-1">
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Label htmlFor="ws-rate-limit">Max messages</Label>
                  <InfoHover description="Maximum number of messages a client can send within the time window. Excess messages are rejected." />
                </div>
                <Input
                  id="ws-rate-limit"
                  type="number"
                  min={1}
                  value={config.message_rate_limit}
                  onChange={(e) =>
                    updateProperty("message_rate_limit", Number(e.target.value))
                  }
                  disabled={initialLoading}
                />
                <p className="text-xs text-muted-foreground">
                  messages per window
                </p>
              </div>
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Label htmlFor="ws-rate-window">Window</Label>
                  <InfoHover description="Time window for rate limit counting. The max message limit is enforced within this period." />
                </div>
                <Input
                  id="ws-rate-window"
                  type="number"
                  min={1}
                  value={config.message_rate_window_seconds}
                  onChange={(e) =>
                    updateProperty(
                      "message_rate_window_seconds",
                      Number(e.target.value),
                    )
                  }
                  disabled={initialLoading}
                />
                <p className="text-xs text-muted-foreground">seconds</p>
              </div>
            </div>
          )}
        </div>

        <Separator />

        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <Label>Allowed origins</Label>
            <InfoHover description="Origins allowed to connect via WebSocket. Leave empty to allow all origins (*). Native clients (mobile/desktop) are always allowed regardless of this list." />
          </div>
          {config.allowed_origins.length === 0 ? (
            <p className="text-xs text-muted-foreground">
              All origins allowed (*)
            </p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {config.allowed_origins.map((origin) => (
                <Badge key={origin} variant="secondary" className="gap-1 pr-1">
                  {origin}
                  <button
                    type="button"
                    onClick={() => removeOrigin(origin)}
                    className="ml-1 rounded-sm hover:bg-muted"
                    disabled={initialLoading}
                  >
                    <IconX className="h-3 w-3" />
                  </button>
                </Badge>
              ))}
            </div>
          )}
          <div className="flex gap-2">
            <Input
              placeholder="https://app.example.com"
              value={originInput}
              onChange={(e) => setOriginInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  addOrigin();
                }
              }}
              disabled={initialLoading}
              className="max-w-sm"
            />
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={addOrigin}
              disabled={initialLoading || !originInput.trim()}
            >
              <IconPlus className="h-4 w-4" />
              Add
            </Button>
          </div>
        </div>

        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Ping period: {Math.floor((config.pong_wait_seconds * 9) / 10)}s
            &nbsp;·&nbsp; Max message:{" "}
            {(config.max_message_size_kb / 1024).toFixed(2)} MB
          </p>
          <Button onClick={() => save()} disabled={loading || initialLoading}>
            {loading && <IconRefresh className="h-4 w-4 mr-2 animate-spin" />}
            Save Changes
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
