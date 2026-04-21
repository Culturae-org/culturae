"use client";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import {
  type TimeRange,
  useDashboardGamesData,
} from "@/hooks/useDashboardGamesData";
import { IconRefresh } from "@tabler/icons-react";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";

const chartConfig = {
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

const GAME_MODES = [
  { value: "all", label: "All Modes" },
  { value: "solo", label: "Solo" },
  { value: "1v1", label: "1v1" },
  { value: "tournament", label: "Tournament" },
  { value: "team", label: "Team" },
];

function ChartSkeleton() {
  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <Skeleton className="h-5 w-32" />
        <Skeleton className="h-4 w-48" />
        <CardAction>
          <Skeleton className="h-8 w-48" />
        </CardAction>
      </CardHeader>
      <CardContent>
        <Skeleton className="h-[300px] w-full" />
      </CardContent>
    </Card>
  );
}

export function GameChart() {
  const {
    data,
    loading,
    refreshing,
    timeRange,
    gameMode,
    setTimeRange,
    setGameMode,
    refresh,
  } = useDashboardGamesData();

  if (loading) {
    return <ChartSkeleton />;
  }

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle>Games Trend</CardTitle>
        <CardDescription>Daily games over time</CardDescription>
        <CardAction>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => refresh()}
              disabled={refreshing || loading}
            >
              <IconRefresh
                className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`}
              />
            </Button>
            <Select
              value={timeRange}
              onValueChange={(val) => setTimeRange(val as TimeRange)}
            >
              <SelectTrigger className="w-[130px]" size="sm">
                <SelectValue placeholder="Range" />
              </SelectTrigger>
              <SelectContent className="rounded-xl">
                <SelectItem value="7d" className="rounded-lg">
                  Last 7 days
                </SelectItem>
                <SelectItem value="30d" className="rounded-lg">
                  Last 30 days
                </SelectItem>
                <SelectItem value="90d" className="rounded-lg">
                  Last 90 days
                </SelectItem>
                <SelectItem value="all" className="rounded-lg">
                  All time
                </SelectItem>
              </SelectContent>
            </Select>
            <Select value={gameMode} onValueChange={setGameMode}>
              <SelectTrigger className="w-[150px]" size="sm">
                <SelectValue placeholder="Mode" />
              </SelectTrigger>
              <SelectContent className="rounded-xl">
                {GAME_MODES.map((mode) => (
                  <SelectItem
                    key={mode.value}
                    value={mode.value}
                    className="rounded-lg"
                  >
                    {mode.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </CardAction>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[300px] w-full">
          <AreaChart
            data={data}
            margin={{ top: 10, right: 10, left: 0, bottom: 0 }}
          >
            <CartesianGrid vertical={false} className="stroke-muted" />
            <XAxis
              dataKey="date"
              type="category"
              allowDuplicatedCategory={false}
              interval={Math.floor((data.length - 1) / 10)}
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tickFormatter={(value) =>
                new Date(value).toLocaleDateString("en-US", {
                  month: "short",
                  day: "numeric",
                })
              }
              className="text-xs"
            />
            <YAxis className="text-xs" />
            <ChartTooltip content={<ChartTooltipContent indicator="line" />} />
            <Area
              type="monotone"
              dataKey="total_games"
              stackId="1"
              stroke={chartConfig.total_games.color}
              fill={chartConfig.total_games.color}
              fillOpacity={0.3}
            />
            <Area
              type="monotone"
              dataKey="completed_games"
              stackId="2"
              stroke={chartConfig.completed_games.color}
              fill={chartConfig.completed_games.color}
              fillOpacity={0.3}
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
