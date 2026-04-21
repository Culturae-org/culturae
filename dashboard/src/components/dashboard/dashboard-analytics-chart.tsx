"use client";

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
import { useDashboardGamesData } from "@/hooks/useDashboardGamesData";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";

const chartConfig = {
  total_games: {
    label: "Games",
    color: "hsl(var(--chart-1))",
  },
} satisfies ChartConfig;

function ChartSkeleton() {
  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <Skeleton className="h-5 w-32" />
        <Skeleton className="h-4 w-48" />
      </CardHeader>
      <CardContent>
        <Skeleton className="h-[300px] w-full" />
      </CardContent>
    </Card>
  );
}

export function DashboardAnalyticsChart() {
  const { data: gamesData, loading } = useDashboardGamesData();

  if (loading) {
    return <ChartSkeleton />;
  }

  const data = gamesData;

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle>Games Trend</CardTitle>
        <CardDescription>Daily games over time</CardDescription>
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
              stroke={chartConfig.total_games.color}
              fill={chartConfig.total_games.color}
              fillOpacity={0.3}
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
