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
import type { ELOConfig } from "@/lib/types/settings.types";
import { IconRefresh } from "@tabler/icons-react";
import { InfoHover } from "./info-hover";

const DEFAULTS: ELOConfig = {
  k_factor_low_games: 32,
  k_factor_high_games: 16,
  k_factor_threshold: 30,
  min_rating: 0,
  max_rating: 9999,
};

export function EloConfigCard() {
  const { config, updateProperty, save, loading, initialLoading } =
    useSettings<ELOConfig>({
      endpoint: SETTINGS_ENDPOINTS.ELO_CONFIG,
      defaults: DEFAULTS,
      validate: (cfg) => {
        if (
          cfg.k_factor_low_games <= 0 ||
          cfg.k_factor_high_games <= 0 ||
          cfg.k_factor_threshold <= 0
        ) {
          return "K-factor values and threshold must be positive";
        }
        if (cfg.min_rating < 0 || cfg.max_rating <= cfg.min_rating) {
          return "min_rating must be >= 0 and max_rating must be > min_rating";
        }
        return true;
      },
    });

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          ELO Rating Settings
        </CardTitle>
        <CardDescription>
          Configure the ELO rating system parameters for competitive matches.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div>
          <div className="flex items-center gap-2 mb-3">
            <p className="text-sm font-medium">K-Factor</p>
            <InfoHover description="The K-factor determines how much a player's rating changes after a game. Higher values mean faster rating changes." />
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="elo-k-low">K-Factor (new players)</Label>
                <InfoHover description="K-factor for players with less than the threshold number of games. Higher values allow new players to climb or fall faster." />
              </div>
              <Input
                id="elo-k-low"
                type="number"
                min={1}
                max={100}
                step={1}
                value={config.k_factor_low_games}
                onChange={(e) =>
                  updateProperty("k_factor_low_games", Number(e.target.value))
                }
                disabled={initialLoading}
              />
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="elo-k-high">K-Factor (experienced)</Label>
                <InfoHover description="K-factor for players with the threshold number of games or more. Lower values mean rating changes are more stable." />
              </div>
              <Input
                id="elo-k-high"
                type="number"
                min={1}
                max={100}
                step={1}
                value={config.k_factor_high_games}
                onChange={(e) =>
                  updateProperty("k_factor_high_games", Number(e.target.value))
                }
                disabled={initialLoading}
              />
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="elo-k-threshold">Games threshold</Label>
                <InfoHover description="Number of games a player must play before switching from the new player K-factor to the experienced K-factor." />
              </div>
              <Input
                id="elo-k-threshold"
                type="number"
                min={1}
                max={1000}
                step={1}
                value={config.k_factor_threshold}
                onChange={(e) =>
                  updateProperty("k_factor_threshold", Number(e.target.value))
                }
                disabled={initialLoading}
              />
            </div>
          </div>
        </div>

        <Separator />

        <div>
          <div className="flex items-center gap-2 mb-3">
            <p className="text-sm font-medium">Rating bounds</p>
            <InfoHover description="Minimum and maximum possible ELO ratings. Players cannot go below min_rating or above max_rating." />
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="elo-min-rating">Minimum rating</Label>
                <InfoHover description="Lowest possible ELO rating. New players start here. No player can fall below this rating." />
              </div>
              <Input
                id="elo-min-rating"
                type="number"
                min={0}
                step={100}
                value={config.min_rating}
                onChange={(e) =>
                  updateProperty("min_rating", Number(e.target.value))
                }
                disabled={initialLoading}
              />
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="elo-max-rating">Maximum rating</Label>
                <InfoHover description="Highest possible ELO rating. No player can rise above this rating, even with consistent wins." />
              </div>
              <Input
                id="elo-max-rating"
                type="number"
                min={1}
                step={100}
                value={config.max_rating}
                onChange={(e) =>
                  updateProperty("max_rating", Number(e.target.value))
                }
                disabled={initialLoading}
              />
            </div>
          </div>
        </div>

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
