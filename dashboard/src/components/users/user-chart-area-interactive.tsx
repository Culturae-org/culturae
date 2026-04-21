"use client";

import { Alert, AlertDescription } from "@/components/ui/alert";
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
import { useUsersStats } from "@/hooks";
import { useIsMobile } from "@/hooks/useMobile";
import { IconRefresh } from "@tabler/icons-react";
import * as React from "react";
import { Area, AreaChart, CartesianGrid, XAxis } from "recharts";
export const description = "An interactive area chart";

type TimeRange = "7d" | "30d" | "90d" | "all";

function ChartSkeleton() {
  return (
    <Card className="w-full @container/card border-0 dark:border">
      <CardHeader>
        <Skeleton className="h-6 w-48" />
        <Skeleton className="h-4 w-64" />
        <CardAction>
          <Skeleton className="h-8 w-40" />
        </CardAction>
      </CardHeader>
      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
        <div className="aspect-auto h-[250px] w-full flex flex-col justify-end gap-2">
          <div className="flex items-end justify-between h-full gap-1 px-4">
            {[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11].map((i) => (
              <Skeleton
                key={`chart-skel-${i}`}
                className="flex-1 rounded-t-sm"
                style={{
                  height: `${Math.random() * 60 + 20}%`,
                }}
              />
            ))}
          </div>
          <div className="flex justify-between px-4">
            {[0, 1, 2, 3, 4, 5].map((i) => (
              <Skeleton key={`label-skel-${i}`} className="h-3 w-12" />
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function UsersChart() {
  const isMobile = useIsMobile();
  const { fetchCreationDates } = useUsersStats();

  const [timeRange, setTimeRange] = React.useState<TimeRange>(
    isMobile ? "7d" : "30d",
  );
  const [chartData, setChartData] = React.useState<
    Array<{ date: string; users: number }>
  >([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [refreshing, setRefreshing] = React.useState(false);

  const fetchUserData = React.useCallback(
    async (isRefresh = false) => {
      try {
        if (isRefresh) {
          setRefreshing(true);
        } else {
          setLoading(true);
        }
        setError(null);

        let startDate: string | undefined;
        let endDate: string | undefined;
        const now = new Date();

        if (timeRange !== "all") {
          let daysToSubtract = 30;
          if (timeRange === "7d") daysToSubtract = 7;
          else if (timeRange === "90d") daysToSubtract = 90;

          const startDateCalc = new Date(now);
          startDateCalc.setDate(startDateCalc.getDate() - daysToSubtract);
          startDate = startDateCalc.toISOString().split("T")[0];
          endDate = now.toISOString().split("T")[0];
        }

        const dates = await fetchCreationDates(startDate, endDate);

        const statsByDate: Record<string, number> = {};

        for (const dateString of dates) {
          const date = new Date(dateString);
          const dateKey = date.toISOString().split("T")[0];
          statsByDate[dateKey] = (statsByDate[dateKey] || 0) + 1;
        }

        const filteredData = Object.entries(statsByDate)
          .sort(([a], [b]) => a.localeCompare(b))
          .map(([date, count]) => ({
            date,
            users: count,
          }));

        setChartData(filteredData);
      } catch (err) {
        console.error("Error fetching user data:", err);
        setError("Failed to load user statistics");
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [fetchCreationDates, timeRange],
  );

  React.useEffect(() => {
    fetchUserData();
  }, [fetchUserData]);

  const chartConfig = {
    users: {
      label: "New Users",
      color: "var(--primary)",
    },
  } satisfies ChartConfig;

  if (loading && !refreshing) {
    return <ChartSkeleton />;
  }

  if (error) {
    return (
      <Card className="w-full @container/card border-0 dark:border">
        <CardHeader>
          <CardTitle>User Registration Trends</CardTitle>
          <CardDescription>Error loading data</CardDescription>
        </CardHeader>
        <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
          <div className="aspect-auto h-[250px] w-full flex items-center justify-center">
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full @container/card border-0 dark:border">
      <CardHeader>
        <CardTitle>User Registration Trends</CardTitle>
        <CardDescription>
          <span className="hidden @[540px]/card:block">
            Daily user registrations over time
          </span>
          <span className="@[540px]/card:hidden">User registrations</span>
        </CardDescription>
        <CardAction>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => fetchUserData(true)}
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
              <SelectTrigger className="w-[140px]" size="sm">
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
          </div>
        </CardAction>
      </CardHeader>
      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6 relative">
        {chartData.length === 0 ? (
          <div className="aspect-auto h-[250px] w-full flex items-center justify-center text-muted-foreground">
            No user registration data available
          </div>
        ) : (
          <ChartContainer
            config={chartConfig}
            className="aspect-auto h-[250px] w-full"
          >
            <AreaChart data={chartData}>
              <defs>
                <linearGradient id="fillUsers" x1="0" y1="0" x2="0" y2="1">
                  <stop
                    offset="5%"
                    stopColor="var(--color-users)"
                    stopOpacity={1.0}
                  />
                  <stop
                    offset="95%"
                    stopColor="var(--color-users)"
                    stopOpacity={0.1}
                  />
                </linearGradient>
              </defs>
              <CartesianGrid vertical={false} />
              <XAxis
                dataKey="date"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                minTickGap={32}
                tickFormatter={(value) => {
                  const date = new Date(value);
                  return date.toLocaleDateString("en-US", {
                    month: "short",
                    day: "numeric",
                  });
                }}
              />
              <ChartTooltip
                cursor={false}
                content={
                  <ChartTooltipContent
                    labelFormatter={(value) => {
                      return new Date(value).toLocaleDateString("en-US", {
                        month: "short",
                        day: "numeric",
                        year: "numeric",
                      });
                    }}
                    indicator="line"
                  />
                }
              />
              <Area
                dataKey="users"
                type="natural"
                fill="url(#fillUsers)"
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
