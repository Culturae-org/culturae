"use client";

import { Card, CardContent } from "@/components/ui/card";
import type { GameDetail } from "@/lib/types/games.types";
import { IconExternalLink } from "@tabler/icons-react";
import { Link } from "react-router";

function formatDate(dateString: string) {
  return new Date(dateString).toLocaleString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function scoreModLabel(timeBonus?: boolean): string {
  if (timeBonus === undefined) return "—";
  return timeBonus ? "Time bonus" : "Classic";
}

interface GameOverviewCardProps {
  game: GameDetail;
}

export function GameOverviewCard({ game }: GameOverviewCardProps) {
  return (
    <Card className="border-0 dark:border">
      <CardContent className="pt-4 pb-4">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-3 text-sm">
          <div className="space-y-3">
            <div>
              <div className="text-xs text-muted-foreground mb-0.5">Created</div>
              <div>{formatDate(game.created_at)}</div>
            </div>
            {game.started_at && (
              <div>
                <div className="text-xs text-muted-foreground mb-0.5">Started</div>
                <div>{formatDate(game.started_at)}</div>
              </div>
            )}
            {game.completed_at && (
              <div>
                <div className="text-xs text-muted-foreground mb-0.5">Completed</div>
                <div>{formatDate(game.completed_at)}</div>
              </div>
            )}
          </div>
          <div className="space-y-3">
            <div>
              <div className="text-xs text-muted-foreground mb-0.5">Score mode</div>
              <div>{scoreModLabel(game.time_bonus)}</div>
            </div>
            {game.points_per_correct !== undefined && (
              <div>
                <div className="text-xs text-muted-foreground mb-0.5">
                  Points per correct
                </div>
                <div>{game.points_per_correct}</div>
              </div>
            )}
            {game.template_id && (
              <div>
                <div className="text-xs text-muted-foreground mb-0.5">Template</div>
                <Link
                  to={`/game-templates?view=${game.template_id}`}
                  className="flex items-center gap-1 hover:underline"
                >
                  <span className="font-mono text-xs opacity-70">
                    {game.template_id.slice(0, 8)}…
                  </span>
                  <IconExternalLink className="h-3 w-3" />
                </Link>
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
