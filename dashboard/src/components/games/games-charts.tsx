"use client";

import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { Skeleton } from "@/components/ui/skeleton";
import { apiGet } from "@/lib/api-client";
import { GAMES_ENDPOINTS } from "@/lib/api/endpoints";
import * as React from "react";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  XAxis,
  YAxis,
} from "recharts";

const _modeChartConfig = {
  count: {
    label: "Games",
    color: "var(--chart-1)",
  },
} satisfies ChartConfig;

const _dailyChartConfig = {
  total_games: {
    label: "Total Games",
    color: "var(--chart-1)",
  },
  completed_games: {
    label: "Completed",
    color: "var(--chart-2)",
  },
  cancelled_games: {
    label: "Cancelled",
    color: "var(--chart-3)",
  },
} satisfies ChartConfig;

const playersChartConfig = {
  total_players: {
    label: "Players",
    color: "var(--chart-1)",
  },
} satisfies ChartConfig;

const completionChartConfig = {
  completion_rate: {
    label: "Completion Rate",
    color: "var(--chart-1)",
  },
} satisfies ChartConfig;

function SimpleBarChart({
  data,
  title,
}: { data: Array<{ label: string; value: number }>; title: string }) {
  const maxValue = Math.max(...data.map((d) => d.value));

  return (
    <div className="space-y-4">
      <h4 className="text-sm font-medium">{title}</h4>
      <div className="space-y-2">
        {data.map((item) => (
          <div key={item.label} className="flex items-center gap-2">
            <div className="w-20 text-xs text-muted-foreground truncate">
              {item.label}
            </div>
            <div className="flex-1 bg-muted rounded-full h-2">
              <div
                className="bg-primary rounded-full h-2 transition-all duration-300"
                style={{ width: `${(item.value / maxValue) * 100}%` }}
              />
            </div>
            <div className="w-12 text-xs text-right">{item.value}</div>
          </div>
        ))}
      </div>
    </div>
  );
}

function SimpleLineChart({
  data,
  title,
}: {
  data: Array<{ date: string; total_games: number; completed_games: number }>;
  title: string;
}) {
  const maxValue = Math.max(
    ...data.flatMap((d) => [d.total_games, d.completed_games]),
  );

  return (
    <div className="space-y-4">
      <h4 className="text-sm font-medium">{title}</h4>
      <div className="space-y-2">
        {data.slice(-7).map((item) => (
          <div key={item.date} className="flex items-center gap-2">
            <div className="w-16 text-xs text-muted-foreground">
              {new Date(item.date).toLocaleDateString("en-US", {
                month: "short",
                day: "numeric",
              })}
            </div>
            <div className="flex-1 flex gap-1">
              <div className="flex-1 bg-muted rounded h-2">
                <div
                  className="rounded h-2 transition-all duration-300"
                  style={{
                    width: `${(item.total_games / maxValue) * 100}%`,
                    backgroundColor: "var(--chart-1)",
                  }}
                  title={`Total: ${item.total_games}`}
                />
              </div>
              <div className="flex-1 bg-muted rounded h-2">
                <div
                  className="rounded h-2 transition-all duration-300"
                  style={{
                    width: `${(item.completed_games / maxValue) * 100}%`,
                    backgroundColor: "var(--chart-2)",
                  }}
                  title={`Completed: ${item.completed_games}`}
                />
              </div>
            </div>
            <div className="w-16 text-xs text-right">
              {item.total_games}/{item.completed_games}
            </div>
          </div>
        ))}
      </div>
      <div className="flex gap-4 text-xs text-muted-foreground">
        <div className="flex items-center gap-1">
          <div
            className="w-3 h-3 rounded"
            style={{ backgroundColor: "var(--chart-1)" }}
          />
          Total Games
        </div>
        <div className="flex items-center gap-1">
          <div
            className="w-3 h-3 rounded"
            style={{ backgroundColor: "var(--chart-2)" }}
          />
          Completed Games
        </div>
      </div>
    </div>
  );
}

export function GamesCharts() {
  const [modeStats, setModeStats] = React.useState<
    Array<{ mode: string; count: number }>
  >([]);
  const [dailyStats, setDailyStats] = React.useState<
    Array<{
      date: string;
      total_games: number;
      completed_games: number;
      cancelled_games: number;
      total_players: number;
    }>
  >([]);
  const [performanceStats, setPerformanceStats] = React.useState<{
    avg_game_duration_seconds: number;
    avg_questions_per_game: number;
    avg_players_per_game: number;
    total_questions_used: number;
    most_popular_mode: string;
  } | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    const fetchChartsData = async () => {
      try {
        setLoading(true);

        const modeResponse = await apiGet(GAMES_ENDPOINTS.MODE_STATS);
        if (modeResponse.ok) {
          const modeData = await modeResponse.json();
          const modeObj = modeData.data ?? modeData;
          const modeArray = Object.entries(modeObj).map(([mode, count]) => ({
            mode: mode.charAt(0).toUpperCase() + mode.slice(1),
            count: count as number,
          }));
          setModeStats(modeArray);
        }

        const dailyResponse = await apiGet(GAMES_ENDPOINTS.DAILY_STATS);
        if (dailyResponse.ok) {
          const dailyData = await dailyResponse.json();
          setDailyStats(dailyData.data ?? dailyData);
        }

        const perfResponse = await apiGet(GAMES_ENDPOINTS.PERFORMANCE_STATS);
        if (perfResponse.ok) {
          const perfData = await perfResponse.json();
          setPerformanceStats(perfData.data ?? perfData);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "An error occurred");
      } finally {
        setLoading(false);
      }
    };

    fetchChartsData();
  }, []);

  if (loading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {[0, 1, 2, 3, 4, 5].map((i) => (
          <Card className="border-0 dark:border" key={`chart-skel-${i}`}>
            <CardHeader>
              <Skeleton className="h-4 w-32" />
              <Skeleton className="h-3 w-48" />
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {[0, 1, 2, 3].map((j) => (
                  <Skeleton
                    key={`chart-skel-inner-${j}`}
                    className="h-4 w-full"
                  />
                ))}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="grid gap-4 md:grid-cols-2">
        <Card className="md:col-span-2 lg:col-span-3">
          <CardContent className="flex items-center justify-center h-32">
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          </CardContent>
        </Card>
      </div>
    );
  }

  const formattedDailyData = dailyStats.map((day) => ({
    ...day,
    date: new Date(day.date).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
    }),
    completion_rate:
      day.total_games > 0
        ? Math.round((day.completed_games / day.total_games) * 100)
        : 0,
  }));

  const topModes = [...modeStats].sort((a, b) => b.count - a.count).slice(0, 5);

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Game Modes Distribution</CardTitle>
          <CardDescription>Popularity of different game modes</CardDescription>
        </CardHeader>
        <CardContent>
          {modeStats.length > 0 ? (
            <SimpleBarChart
              data={modeStats.map((stat) => ({
                label: stat.mode,
                value: stat.count,
              }))}
              title=""
            />
          ) : (
            <div className="text-center text-muted-foreground py-8">
              No data available
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Daily Game Activity</CardTitle>
          <CardDescription>
            Games created and completed over the last 7 days
          </CardDescription>
        </CardHeader>
        <CardContent>
          {dailyStats.length > 0 ? (
            <SimpleLineChart data={dailyStats} title="" />
          ) : (
            <div className="text-center text-muted-foreground py-8">
              No data available
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Performance Metrics</CardTitle>
          <CardDescription>
            Key performance indicators for games
          </CardDescription>
        </CardHeader>
        <CardContent>
          {performanceStats ? (
            <div className="grid gap-4 grid-cols-2">
              <div className="space-y-2">
                <div className="text-2xl font-bold">
                  {Math.floor(performanceStats.avg_game_duration_seconds / 60)}m{" "}
                  {Math.floor(performanceStats.avg_game_duration_seconds % 60)}s
                </div>
                <div className="text-sm text-muted-foreground">
                  Avg Duration
                </div>
              </div>
              <div className="space-y-2">
                <div className="text-2xl font-bold">
                  {performanceStats.avg_questions_per_game.toFixed(1)}
                </div>
                <div className="text-sm text-muted-foreground">
                  Questions/Game
                </div>
              </div>
              <div className="space-y-2">
                <div className="text-2xl font-bold">
                  {performanceStats.avg_players_per_game.toFixed(1)}
                </div>
                <div className="text-sm text-muted-foreground">
                  Players/Game
                </div>
              </div>
              <div className="space-y-2">
                <div className="text-2xl font-bold">
                  {performanceStats.total_questions_used.toLocaleString()}
                </div>
                <div className="text-sm text-muted-foreground">
                  Questions Used
                </div>
              </div>
            </div>
          ) : (
            <div className="text-center text-muted-foreground py-8">
              No performance data available
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Most Popular Mode</CardTitle>
          <CardDescription>Top game mode</CardDescription>
        </CardHeader>
        <CardContent>
          {performanceStats?.most_popular_mode ? (
            <div className="flex flex-col items-center justify-center h-[250px]">
              <div className="text-4xl font-bold text-primary capitalize">
                {performanceStats.most_popular_mode}
              </div>
              <div className="text-sm text-muted-foreground mt-2">
                {topModes[0]?.count ?? 0} games played
              </div>
            </div>
          ) : (
            <div className="text-center text-muted-foreground py-8">
              No data available
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Players Over Time</CardTitle>
          <CardDescription>Total players per day</CardDescription>
        </CardHeader>
        <CardContent>
          {dailyStats.length > 0 ? (
            <ChartContainer
              config={playersChartConfig}
              className="min-h-[250px] w-full"
            >
              <AreaChart data={formattedDailyData} accessibilityLayer>
                <CartesianGrid vertical={false} className="stroke-muted" />
                <XAxis
                  dataKey="date"
                  tickLine={false}
                  tickMargin={10}
                  axisLine={false}
                />
                <YAxis tickLine={false} tickMargin={10} axisLine={false} />
                <ChartTooltip
                  content={<ChartTooltipContent indicator="line" />}
                />
                <Area
                  type="monotone"
                  dataKey="total_players"
                  stroke="var(--color-total_players)"
                  fill="var(--color-total_players)"
                  fillOpacity={0.3}
                />
              </AreaChart>
            </ChartContainer>
          ) : (
            <div className="text-center text-muted-foreground py-8">
              No data available
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Completion Rate</CardTitle>
          <CardDescription>Percentage of games completed</CardDescription>
        </CardHeader>
        <CardContent>
          {dailyStats.length > 0 ? (
            <ChartContainer
              config={completionChartConfig}
              className="min-h-[250px] w-full"
            >
              <BarChart data={formattedDailyData} accessibilityLayer>
                <CartesianGrid vertical={false} className="stroke-muted" />
                <XAxis
                  dataKey="date"
                  tickLine={false}
                  tickMargin={10}
                  axisLine={false}
                />
                <YAxis
                  tickLine={false}
                  tickMargin={10}
                  axisLine={false}
                  domain={[0, 100]}
                  tickFormatter={(value) => `${value}%`}
                />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Bar
                  dataKey="completion_rate"
                  fill="var(--color-completion_rate)"
                  radius={4}
                />
              </BarChart>
            </ChartContainer>
          ) : (
            <div className="text-center text-muted-foreground py-8">
              No data available
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Quick Stats</CardTitle>
          <CardDescription>Summary statistics</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 grid-cols-2">
            <div className="space-y-2 p-4 border rounded-lg">
              <div className="text-3xl font-bold">
                {dailyStats.reduce((sum, d) => sum + d.total_games, 0)}
              </div>
              <div className="text-sm text-muted-foreground">
                Total Games (7d)
              </div>
            </div>
            <div className="space-y-2 p-4 border rounded-lg">
              <div className="text-3xl font-bold">
                {dailyStats.reduce((sum, d) => sum + d.completed_games, 0)}
              </div>
              <div className="text-sm text-muted-foreground">
                Completed (7d)
              </div>
            </div>
            <div className="space-y-2 p-4 border rounded-lg">
              <div className="text-3xl font-bold">
                {dailyStats.reduce((sum, d) => sum + d.total_players, 0)}
              </div>
              <div className="text-sm text-muted-foreground">
                Total Players (7d)
              </div>
            </div>
            <div className="space-y-2 p-4 border rounded-lg">
              <div className="text-3xl font-bold">{modeStats.length}</div>
              <div className="text-sm text-muted-foreground">Active Modes</div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
