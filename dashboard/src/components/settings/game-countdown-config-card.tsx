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
import { Skeleton } from "@/components/ui/skeleton";
import { useSettings } from "@/hooks/useSettings";
import { SETTINGS_ENDPOINTS } from "@/lib/api/endpoints";
import type { CountdownConfig } from "@/lib/types/settings.types";
import { InfoHover } from "./info-hover";

const DEFAULTS: CountdownConfig = {
  pre_game_countdown_seconds: 3,
  reconnect_grace_period_seconds: 30,
};

export function GameCountdownConfigCard() {
  const { config, updateProperty, save, loading, initialLoading } =
    useSettings<CountdownConfig>({
      endpoint: SETTINGS_ENDPOINTS.GAME_COUNTDOWN_CONFIG,
      defaults: DEFAULTS,
      validate: (cfg) => {
        if (cfg.pre_game_countdown_seconds <= 0)
          return "Pre-game countdown must be positive";
        if (cfg.reconnect_grace_period_seconds <= 0)
          return "Reconnect grace period must be positive";
        return true;
      },
    });

  if (initialLoading) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Game Countdown Settings</CardTitle>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-20 w-full" />
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          Game Countdown Settings
          <InfoHover description="Timers controlling game start and reconnection behaviour." />
        </CardTitle>
        <CardDescription>
          Configure the pre-game countdown and the reconnect grace period for
          multiplayer games.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="space-y-4">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Label htmlFor="pre-game-countdown">Pre-Game Countdown</Label>
              <InfoHover description="Time the server waits before starting the game after all players are ready. Gives players a moment to prepare." />
            </div>
            <Input
              id="pre-game-countdown"
              type="number"
              min="1"
              value={config.pre_game_countdown_seconds}
              onChange={(e) =>
                updateProperty(
                  "pre_game_countdown_seconds",
                  Number(e.target.value),
                )
              }
              disabled={loading}
            />
            <p className="text-xs text-muted-foreground">
              Seconds — delay between "game ready" and the first question.
            </p>
          </div>

          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Label htmlFor="reconnect-grace">Reconnect Grace Period</Label>
              <InfoHover description="How long the game is paused when a player disconnects. The client displays a countdown and attempts auto-reconnect. If the player does not reconnect in time, the remaining player wins." />
            </div>
            <Input
              id="reconnect-grace"
              type="number"
              min="1"
              value={config.reconnect_grace_period_seconds}
              onChange={(e) =>
                updateProperty(
                  "reconnect_grace_period_seconds",
                  Number(e.target.value),
                )
              }
              disabled={loading}
            />
            <p className="text-xs text-muted-foreground">
              Seconds — game is paused and client shows a countdown. Default: 30s.
            </p>
          </div>
        </div>

        <Separator />

        <div className="flex gap-2">
          <Button onClick={() => save()} disabled={loading} className="gap-2">
            {loading ? "Saving..." : "Save Changes"}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
