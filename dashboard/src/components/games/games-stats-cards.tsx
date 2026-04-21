"use client";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useGamesCharts } from "@/hooks/useGamesCharts";
import {
  IconClock,
  IconGoGame,
  IconTrendingDown,
  IconTrendingUp,
  IconTrophy,
  IconUsers,
} from "@tabler/icons-react";

export function GamesStatsCards() {
  const { data, loading, error } = useGamesCharts();

  if (loading) {
    const _skeletonCount = 8;
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {[0, 1, 2, 3, 4, 5, 6, 7].map((i) => (
          <Card key={`stat-skel-${i}`}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-4 w-4" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-8 w-16 mb-1" />
              <Skeleton className="h-3 w-32" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (error || !data) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error || "No data available"}</AlertDescription>
      </Alert>
    );
  }

  const formatDuration = (seconds: number) => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = Math.floor(seconds % 60);
    return `${minutes}m ${remainingSeconds}s`;
  };

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card className="border-0 dark:border">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Total Games</CardTitle>
          <IconGoGame className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {data.stats.total_games.toLocaleString()}
          </div>
          <p className="text-xs text-muted-foreground">All games created</p>
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Active Games</CardTitle>
          <IconTrendingUp className="h-4 w-4 " />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {data.stats.active_games.toLocaleString()}
          </div>
          <p className="text-xs text-muted-foreground">Currently running</p>
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Completed Games</CardTitle>
          <IconTrophy className="h-4 w-4" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {data.stats.completed_games.toLocaleString()}
          </div>
          <p className="text-xs text-muted-foreground">Successfully finished</p>
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Abandoned Games</CardTitle>
          <IconTrendingDown className="h-4 w-4" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {data.stats.abandoned_games.toLocaleString()}
          </div>
          <p className="text-xs text-muted-foreground">Left unfinished</p>
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Total Players</CardTitle>
          <IconUsers className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {data.stats.total_players.toLocaleString()}
          </div>
          <p className="text-xs text-muted-foreground">Across all games</p>
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">
            Avg Game Duration
          </CardTitle>
          <IconClock className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {formatDuration(data.stats.avg_game_duration || 0)}
          </div>
          <p className="text-xs text-muted-foreground">For completed games</p>
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">
            Avg Players/Game
          </CardTitle>
          <IconUsers className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {data.stats.avg_players_per_game
              ? data.stats.avg_players_per_game.toFixed(1)
              : "N/A"}
          </div>
          <p className="text-xs text-muted-foreground">Players per game</p>
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Popular Mode</CardTitle>
          <IconGoGame className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold capitalize">
            {data.stats.most_popular_mode || "N/A"}
          </div>
          <p className="text-xs text-muted-foreground">Most played game mode</p>
        </CardContent>
      </Card>
    </div>
  );
}
