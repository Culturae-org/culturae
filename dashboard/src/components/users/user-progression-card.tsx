"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import type { AdminUser } from "@/lib/types/user.types";
import {
  IconClock,
  IconFlame,
  IconTarget,
  IconTrophy,
} from "@tabler/icons-react";

function formatPlayTime(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`;
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  return m > 0 ? `${h}h ${m}m` : `${h}h`;
}

function formatDate(dateString: string) {
  return new Date(dateString).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

interface UserProgressionCardProps {
  user: AdminUser;
}

export function UserProgressionCard({ user }: UserProgressionCardProps) {
  const gs = user.game_stats;

  const getLevelColor = (level: string) => {
    const colors: Record<string, string> = {
      Beginner:
        "text-slate-600 bg-slate-100 dark:text-slate-400 dark:bg-slate-800",
      Intermediate:
        "text-blue-600 bg-blue-100 dark:text-blue-400 dark:bg-blue-900",
      Pro: "text-purple-600 bg-purple-100 dark:text-purple-400 dark:bg-purple-900",
      Expert:
        "text-orange-600 bg-orange-100 dark:text-orange-400 dark:bg-orange-900",
      Legend:
        "text-yellow-600 bg-yellow-100 dark:text-yellow-400 dark:bg-yellow-900",
    };
    return colors[level] ?? colors.Beginner;
  };

  const winRate =
    gs && gs.total_games > 0
      ? Math.round((gs.games_won / gs.total_games) * 100)
      : 0;

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center">Progression</CardTitle>
      </CardHeader>
      <CardContent className="space-y-5">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-muted-foreground text-xs mb-1">Rank</p>
            <Badge
              variant="outline"
              className={`text-sm ${getLevelColor(user.rank)}`}
            >
              <IconTrophy className="h-3.5 w-3.5 mr-1" />
              {user.rank || "N/A"}
            </Badge>
          </div>
          <div className="text-right">
            <p className="text-muted-foreground text-xs mb-1">Experience</p>
            <p className="text-xl font-bold">
              {(user.experience ?? 0).toLocaleString()} XP
            </p>
          </div>
        </div>

        <Separator />

        <div>
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium flex items-center">ELO Rating</p>
            <p className="text-2xl font-bold">{user.elo_rating ?? 1000}</p>
          </div>
          <p className="text-xs text-muted-foreground">
            {user.elo_games_played ?? 0} ranked game
            {user.elo_games_played !== 1 ? "s" : ""} played
          </p>
        </div>

        <Separator />

        {gs && gs.total_games > 0 && (
          <div>
            <div className="flex items-center justify-between mb-2">
              <p className="text-sm font-medium flex items-center gap-1.5">
                <IconTarget className="h-4 w-4" />
                Win Rate
              </p>
              <p className="text-lg font-bold">{winRate}%</p>
            </div>
            <div className="h-2 w-full rounded-full bg-muted overflow-hidden">
              <div
                className="h-full rounded-full bg-primary transition-all"
                style={{ width: `${winRate}%` }}
              />
            </div>
            <div className="flex justify-between mt-1 text-xs text-muted-foreground">
              <span>{gs.games_won} W</span>
              <span>{gs.games_lost} L</span>
            </div>
          </div>
        )}

        {gs && (
          <div className="flex items-center justify-between p-3 rounded-lg border">
            <div className="flex items-center gap-2">
              <IconFlame className="h-4 w-4 text-orange-500" />
              <div>
                <p className="text-sm font-medium">Day Streak</p>
                <p className="text-xs text-muted-foreground">
                  Best: {gs.best_day_streak}
                </p>
              </div>
            </div>
            <p className="text-2xl font-bold">{gs.day_streak}</p>
          </div>
        )}

        {gs && (
          <div className="flex items-center justify-between p-3 rounded-lg border">
            <div className="flex items-center gap-2">
              <IconClock className="h-4 w-4" />
              <div>
                <p className="text-sm font-medium">Total Play Time</p>
                {gs.last_game_at && (
                  <p className="text-xs text-muted-foreground">
                    Last game: {formatDate(gs.last_game_at)}
                  </p>
                )}
              </div>
            </div>
            <p className="text-xl font-bold">{formatPlayTime(gs.play_time)}</p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
