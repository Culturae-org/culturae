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
import type { RankDefinition, XPConfig } from "@/lib/types/settings.types";
import {
  IconPlus,
  IconRefresh,
  IconTrash,
} from "@tabler/icons-react";
import type * as React from "react";
import { InfoHover } from "./info-hover";

const DEFAULTS: XPConfig = {
  base_xp: 2000,
  growth_rate: 1.5,
  solo_multiplier: 0.5,
  onevone_multiplier: 1.0,
  multi_multiplier: 1.0,
  winner_bonus: 100,
  ranks: [
    { name: "Beginner", min_level: 0 },
    { name: "Intermediate", min_level: 5 },
    { name: "Pro", min_level: 10 },
    { name: "Expert", min_level: 15 },
    { name: "Legend", min_level: 20 },
  ],
};

export function XpConfigCard() {
  const {
    config,
    setConfig,
    save: baseSave,
    loading,
    initialLoading,
  } = useSettings<XPConfig>({
    endpoint: SETTINGS_ENDPOINTS.XP_CONFIG,
    defaults: DEFAULTS,
    validate: (cfg) => {
      if (cfg.base_xp <= 0 || cfg.growth_rate <= 0) {
        return "base_xp and growth_rate must be positive";
      }
      if (cfg.ranks.length === 0) {
        return "At least one rank definition is required";
      }
      const emptyRank = cfg.ranks.find((r) => !r.name.trim());
      if (emptyRank) {
        return "All rank names must be filled in";
      }
      return true;
    },
  });

  const setNum =
    (key: keyof Omit<XPConfig, "ranks">) =>
    (e: React.ChangeEvent<HTMLInputElement>) =>
      setConfig((prev) => ({ ...prev, [key]: Number(e.target.value) }));

  const updateRank = (
    index: number,
    field: keyof RankDefinition,
    value: string,
  ) => {
    setConfig((prev) => {
      const ranks = [...prev.ranks];
      ranks[index] = {
        ...ranks[index],
        [field]: field === "min_level" ? Number(value) : value,
      };
      return { ...prev, ranks };
    });
  };

  const addRank = () => {
    setConfig((prev) => ({
      ...prev,
      ranks: [...prev.ranks, { name: "", min_level: 0 }],
    }));
  };

  const removeRank = (index: number) => {
    setConfig((prev) => ({
      ...prev,
      ranks: prev.ranks.filter((_, i) => i !== index),
    }));
  };

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          XP &amp; Level Settings
        </CardTitle>
        <CardDescription>
          Configure XP gain, level formula, multipliers per game mode, and rank
          thresholds.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div>
          <div className="flex items-center gap-2 mb-3">
            <p className="text-sm font-medium">Level formula</p>
            <InfoHover description="Formula used to calculate player level from total XP. Higher growth rate means slower leveling progression." />
          </div>
          <p className="text-xs text-muted-foreground mb-3">
            level = floor( log<sub>growth_rate</sub>(totalXP / base_xp + 1) )
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="xp-base">Base XP</Label>
                <InfoHover description="Base XP for the leveling formula. Approximate XP needed to reach level 1. Higher values make the game harder to level up." />
              </div>
              <Input
                id="xp-base"
                type="number"
                min={1}
                step={100}
                value={config.base_xp}
                onChange={setNum("base_xp")}
                disabled={initialLoading}
              />
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="xp-growth">Growth rate</Label>
                <InfoHover description="Logarithm base for the level formula. Higher values make leveling slower (e.g., 2.0 = much harder than 1.5)." />
              </div>
              <Input
                id="xp-growth"
                type="number"
                min={1.01}
                step={0.05}
                value={config.growth_rate}
                onChange={setNum("growth_rate")}
                disabled={initialLoading}
              />
            </div>
          </div>
        </div>

        <Separator />

        <div>
          <div className="flex items-center gap-2 mb-3">
            <p className="text-sm font-medium">XP multipliers per game mode</p>
            <InfoHover description="XP earned is multiplied by these values based on game mode. Set to 0 to disable XP gain for a mode." />
          </div>
          <div className="flex flex-wrap gap-4 items-start">
            {(
              [
                {
                  key: "solo_multiplier",
                  label: "Solo",
                  tooltip:
                    "XP multiplier for solo games. Use 0 to disable XP gains.",
                },
                {
                  key: "onevone_multiplier",
                  label: "1v1",
                  tooltip:
                    "XP multiplier for 1v1 games. This applies before the winner bonus.",
                },
                {
                  key: "multi_multiplier",
                  label: "Multiplayer",
                  tooltip:
                    "XP multiplier for multiplayer games with 3+ players.",
                },
              ] as {
                key: keyof Omit<XPConfig, "ranks">;
                label: string;
                tooltip: string;
              }[]
            ).map(({ key, label, tooltip }) => (
              <div key={key} className="space-y-2 w-32">
                <div className="flex items-center gap-2">
                  <Label htmlFor={`xp-${key}`}>{label}</Label>
                  <InfoHover description={tooltip} />
                </div>
                <Input
                  id={`xp-${key}`}
                  type="number"
                  min={0}
                  step={0.1}
                  value={config[key] as number}
                  onChange={setNum(key)}
                  disabled={initialLoading}
                />
              </div>
            ))}
            <div className="space-y-2 w-40">
              <div className="flex items-center gap-2">
                <Label htmlFor="xp-winner-bonus">1v1 winner bonus</Label>
                <InfoHover description="Flat XP bonus awarded to the winner in 1v1 games (on top of score-based XP)." />
              </div>
              <Input
                id="xp-winner-bonus"
                type="number"
                min={0}
                step={10}
                value={config.winner_bonus}
                onChange={setNum("winner_bonus")}
                disabled={initialLoading}
              />
            </div>
          </div>
        </div>

        <Separator />

        <div>
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <p className="text-sm font-medium">Rank thresholds</p>
              <InfoHover description="Define rank names and the minimum level required to achieve them. The highest matching threshold is used." />
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={addRank}
              disabled={initialLoading}
            >
              <IconPlus className="h-4 w-4 mr-2" />
              Add rank
            </Button>
          </div>
          <p className="text-xs text-muted-foreground mb-3">
            Each rank applies to players whose level ≥ min_level. The highest
            matching threshold wins.
          </p>
          <div className="space-y-2">
            <div className="grid grid-cols-[240px_120px_40px] gap-2 text-xs font-medium text-muted-foreground px-1">
              <span>Rank name</span>
              <span>Min level</span>
              <span />
            </div>
            {config.ranks.map((rank, i) => (
              <div
                key={rank.name || `rank-${i}`}
                className="grid grid-cols-[240px_120px_40px] gap-2 items-center"
              >
                <Input
                  value={rank.name}
                  onChange={(e) => updateRank(i, "name", e.target.value)}
                  placeholder="e.g. Legend"
                  disabled={initialLoading}
                />
                <Input
                  type="number"
                  min={0}
                  value={rank.min_level}
                  onChange={(e) => updateRank(i, "min_level", e.target.value)}
                  disabled={initialLoading}
                />
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => removeRank(i)}
                  disabled={initialLoading || config.ranks.length <= 1}
                  className="text-muted-foreground hover:text-destructive"
                >
                  <IconTrash className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </div>
        </div>

        <div className="flex justify-end">
          <Button
            onClick={() => baseSave()}
            disabled={loading || initialLoading}
          >
            {loading && <IconRefresh className="h-4 w-4 mr-2 animate-spin" />}
            Save Changes
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
