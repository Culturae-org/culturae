"use client";

import { Button } from "@/components/ui/button";
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
import { useUsersStats } from "@/hooks/useUsers";
import { IconRefresh, IconUsers } from "@tabler/icons-react";
import * as React from "react";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";

const chartConfig = {
  users: {
    label: "New Users",
    color: "hsl(var(--chart-1))",
  },
} satisfies ChartConfig;

interface UserCreationChartProps {
  className?: string;
}

export function UserCreationChart({ className }: UserCreationChartProps) {
  const { fetchCreationDates, loading, error } = useUsersStats();
  const [chartData, setChartData] = React.useState<
    Array<{ date: string; users: number }>
  >([]);
  const [refreshing, setRefreshing] = React.useState(false);

  const fetchUserCreationData = React.useCallback(
    async (isRefresh = false) => {
      try {
        if (isRefresh) {
          setRefreshing(true);
        }

        const dates = await fetchCreationDates();

        const dateCount: Record<string, number> = {};
        for (const dateString of dates) {
          const date = new Date(dateString).toISOString().split("T")[0];
          dateCount[date] = (dateCount[date] || 0) + 1;
        }

        const formattedData = Object.entries(dateCount)
          .map(([date, count]) => ({
            date: new Date(date).toLocaleDateString("en-US", {
              month: "short",
              day: "numeric",
              year: "numeric",
            }),
            users: count,
            fullDate: date,
          }))
          .sort(
            (a, b) =>
              new Date(a.fullDate).getTime() - new Date(b.fullDate).getTime(),
          );

        setChartData(formattedData);
      } catch (err) {
        console.error("Error fetching user creation data:", err);
      } finally {
        setRefreshing(false);
      }
    },
    [fetchCreationDates],
  );

  React.useEffect(() => {
    fetchUserCreationData();
  }, [fetchUserCreationData]);

  const handleRefresh = () => {
    fetchUserCreationData(true);
  };

  if (loading) {
    return (
      <Card className={className}>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <div className="space-y-1">
            <Skeleton className="h-5 w-32" />
            <Skeleton className="h-4 w-48" />
          </div>
          <Skeleton className="h-8 w-8" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-64 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className={className}>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <div className="space-y-1">
            <CardTitle className="text-base font-medium">
              User Creation Trends
            </CardTitle>
            <CardDescription>
              Daily user registration statistics
            </CardDescription>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={refreshing}
          >
            <IconRefresh
              className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`}
            />
          </Button>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <p className="text-muted-foreground mb-2">
                Failed to load chart data
              </p>
              <Button onClick={handleRefresh} variant="outline" size="sm">
                Try Again
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="space-y-1">
          <CardTitle className="text-base font-medium flex items-center gap-2">
            <IconUsers className="h-4 w-4" />
            User Creation Trends
          </CardTitle>
          <CardDescription>
            Daily user registration statistics ({chartData.length} days tracked)
          </CardDescription>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={handleRefresh}
          disabled={refreshing}
        >
          <IconRefresh
            className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`}
          />
        </Button>
      </CardHeader>
      <CardContent>
        {chartData.length === 0 ? (
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <IconUsers className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">
                No user creation data available
              </p>
            </div>
          </div>
        ) : (
          <ChartContainer config={chartConfig} className="h-64 w-full">
            <AreaChart
              data={chartData}
              margin={{
                left: 12,
                right: 12,
              }}
            >
              <CartesianGrid vertical={false} />
              <XAxis
                dataKey="date"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                tickFormatter={(value) => value.slice(0, 6)}
              />
              <YAxis tickLine={false} axisLine={false} tickMargin={8} />
              <ChartTooltip
                cursor={false}
                content={<ChartTooltipContent indicator="line" />}
              />
              <Area
                dataKey="users"
                type="natural"
                fill="var(--color-users)"
                fillOpacity={0.4}
                stroke="var(--color-users)"
                stackId="a"
              />
            </AreaChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  );
}
