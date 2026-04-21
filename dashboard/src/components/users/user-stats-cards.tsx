"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { AdminUser } from "@/lib/types/user.types";

function formatPlayTime(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`;
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  return m > 0 ? `${h}h ${m}m` : `${h}h`;
}

interface StatCardProps {
  label: string;
  value: string | number;
  sub?: string;
}

function StatCard({ label, value, sub }: StatCardProps) {
  return (
    <Card className="border-0 dark:border">
      <CardHeader className="pb-1 pt-3 px-4">
        <CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
          {label}
        </CardTitle>
      </CardHeader>
      <CardContent className="px-4 pb-3">
        <div className="text-2xl font-bold">{value}</div>
        {sub && <p className="text-xs text-muted-foreground mt-0.5">{sub}</p>}
      </CardContent>
    </Card>
  );
}

interface UserStatsCardsProps {
  user: AdminUser;
}

export function UserStatsCards({ user }: UserStatsCardsProps) {
  const gs = user.game_stats;
  if (!gs) return null;

  const winRate =
    gs.total_games > 0 ? Math.round((gs.games_won / gs.total_games) * 100) : 0;

  return (
    <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-3">
      <StatCard label="Total Games" value={gs.total_games.toLocaleString()} />
      <StatCard label="Wins" value={gs.games_won.toLocaleString()} />
      <StatCard label="Losses" value={gs.games_lost.toLocaleString()} />
      <StatCard label="Win Rate" value={`${winRate}%`} />
      <StatCard
        label="Day Streak"
        value={gs.day_streak}
        sub={`Best: ${gs.best_day_streak}`}
      />
      <StatCard label="Total Score" value={gs.total_score.toLocaleString()} />
      <StatCard
        label="Avg Score"
        value={Math.round(gs.average_score).toLocaleString()}
      />
      <StatCard label="Play Time" value={formatPlayTime(gs.play_time)} />
    </div>
  );
}
