"use client";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
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
import { gameTemplatesService } from "@/lib/services/game-templates.service";
import type { GameConfig } from "@/lib/types/settings.types";
import { IconRefresh, IconRestore } from "@tabler/icons-react";
import * as React from "react";
import { toast } from "sonner";
import { InfoHover } from "./info-hover";

const DEFAULTS: GameConfig = {
  active_ttl_minutes: 1440,
  finished_ttl_minutes: 120,
};

export function GameConfigCard() {
  const { config, updateProperty, save, reset, loading, initialLoading } =
    useSettings<GameConfig>({
      endpoint: SETTINGS_ENDPOINTS.GAME_CONFIG,
      defaults: DEFAULTS,
      validate: (cfg) => {
        if (cfg.active_ttl_minutes <= 0 || cfg.finished_ttl_minutes <= 0) {
          return "Both TTL values must be positive";
        }
        return true;
      },
    });

  const [resetConfirmOpen, setResetConfirmOpen] = React.useState(false);
  const [resetting, setResetting] = React.useState(false);

  const handleSeedTemplates = async () => {
    setResetting(true);
    try {
      const { created } = await gameTemplatesService.seedDefaultTemplates();
      toast.success(
        `Templates seeded — ${created} default template${created !== 1 ? "s" : ""} added`,
      );
    } catch {
      toast.error("Failed to seed game templates");
    } finally {
      setResetting(false);
      setResetConfirmOpen(false);
    }
  };

  const formatTime = (minutes: number) => {
    if (minutes < 60) return `${minutes} min`;
    if (minutes < 1440) return `${Math.floor(minutes / 60)} hr`;
    return `${Math.floor(minutes / 1440)} day`;
  };

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          Game TTL Settings
        </CardTitle>
        <CardDescription>
          Configure how long game data is kept in Redis before automatic
          cleanup.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Label htmlFor="game-active-ttl">Active games TTL</Label>
              <InfoHover description="How long games in waiting or in_progress status are kept in Redis before cleanup." />
            </div>
            <Input
              id="game-active-ttl"
              type="number"
              min={1}
              value={config.active_ttl_minutes}
              onChange={(e) =>
                updateProperty("active_ttl_minutes", Number(e.target.value))
              }
              disabled={initialLoading}
            />
            <p className="text-xs text-muted-foreground font-medium">
              Current: {formatTime(config.active_ttl_minutes)}
            </p>
          </div>
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Label htmlFor="game-finished-ttl">Finished games TTL</Label>
              <InfoHover description="How long completed, cancelled, or abandoned games are kept before cleanup." />
            </div>
            <Input
              id="game-finished-ttl"
              type="number"
              min={1}
              value={config.finished_ttl_minutes}
              onChange={(e) =>
                updateProperty("finished_ttl_minutes", Number(e.target.value))
              }
              disabled={initialLoading}
            />
            <p className="text-xs text-muted-foreground font-medium">
              Current: {formatTime(config.finished_ttl_minutes)}
            </p>
          </div>
        </div>

        <Separator />

        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Active: {formatTime(config.active_ttl_minutes)} · Finished:{" "}
            {formatTime(config.finished_ttl_minutes)}
          </p>
          <Button onClick={() => save()} disabled={loading || initialLoading}>
            {loading && <IconRefresh className="h-4 w-4 mr-2 animate-spin" />}
            Save Changes
          </Button>
        </div>

        <Separator />

        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium">Game templates</p>
            <p className="text-sm text-muted-foreground">
              Re-create missing built-in templates (keeps custom templates
              intact).
            </p>
          </div>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setResetConfirmOpen(true)}
          >
            <IconRestore className="h-4 w-4 mr-2" />
            Seed defaults
          </Button>
        </div>
      </CardContent>

      <AlertDialog open={resetConfirmOpen} onOpenChange={setResetConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Seed missing default templates?</AlertDialogTitle>
            <AlertDialogDescription>
              This will add any missing built-in templates without affecting
              your custom templates. No data will be deleted.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={resetting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleSeedTemplates}
              disabled={resetting}
            >
              {resetting ? "Seeding…" : "Seed templates"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Card>
  );
}
