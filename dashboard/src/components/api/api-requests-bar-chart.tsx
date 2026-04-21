"use client";

import { IconRefresh } from "@tabler/icons-react";
import { TrendingUp } from "lucide-react";
import * as React from "react";
import { Bar, BarChart, CartesianGrid, XAxis } from "recharts";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
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
import { useApiLogs } from "@/hooks/useApiLogs";

const chartConfig = {
  requests: {
    label: "Requests",
    color: "var(--chart-2)",
  },
} satisfies ChartConfig;

type TimeRange = "1h" | "4h" | "6h" | "12h" | "24h" | "7d" | "30d";

interface ApiRequestsBarChartProps {
  className?: string;
  viewMode?: "method" | "status";
}

export function ApiRequestsBarChart({
  className,
  viewMode = "method",
}: ApiRequestsBarChartProps) {
  const { fetchRequestStats } = useApiLogs();
  const [chartData, setChartData] = React.useState<
    Array<{ label: string; requests: number }>
  >([]);
  const [refreshing, setRefreshing] = React.useState(false);
  const [timeRange, setTimeRange] = React.useState<TimeRange>("24h");

  const getDateRange = React.useCallback(
    (range: TimeRange): { start?: Date; end?: Date } => {
      const now = new Date();
      const end = now;

      switch (range) {
        case "1h":
          return { start: new Date(now.getTime() - 60 * 60 * 1000), end };
        case "4h":
          return { start: new Date(now.getTime() - 4 * 60 * 60 * 1000), end };
        case "6h":
          return { start: new Date(now.getTime() - 6 * 60 * 60 * 1000), end };
        case "12h":
          return { start: new Date(now.getTime() - 12 * 60 * 60 * 1000), end };
        case "24h":
          return { start: new Date(now.getTime() - 24 * 60 * 60 * 1000), end };
        case "7d":
          return {
            start: new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000),
            end,
          };
        case "30d":
          return {
            start: new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000),
            end,
          };
        default:
          return {};
      }
    },
    [],
  );

  const fetchData = React.useCallback(async () => {
    try {
      setRefreshing(true);
      const { start, end } = getDateRange(timeRange);
      const stats = await fetchRequestStats(start, end);

      let formattedData: Array<{ label: string; requests: number }> = [];

      if (viewMode === "method") {
        formattedData = Object.entries(stats.requests_by_method || {}).map(
          ([method, count]) => ({
            label: method,
            requests: Number(count),
          }),
        );
      } else {
        formattedData = Object.entries(stats.requests_by_status || {}).map(
          ([status, count]) => ({
            label: status,
            requests: Number(count),
          }),
        );
      }

      formattedData.sort((a, b) => b.requests - a.requests);
      setChartData(formattedData);
    } catch (err) {
      console.error("Error fetching API request data:", err);
    } finally {
      setRefreshing(false);
    }
  }, [fetchRequestStats, viewMode, timeRange, getDateRange]);

  React.useEffect(() => {
    fetchData();
  }, [fetchData]);

  const totalRequests = chartData.reduce((sum, item) => sum + item.requests, 0);

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="capitalize">
              API Requests by {viewMode}
            </CardTitle>
            <CardDescription>
              {totalRequests.toLocaleString()} total requests
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="icon"
              onClick={() => fetchData()}
              disabled={refreshing}
              className="h-8 w-8"
            >
              <IconRefresh
                className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`}
              />
            </Button>
            <Select
              value={timeRange}
              onValueChange={(val) => setTimeRange(val as TimeRange)}
            >
              <SelectTrigger className="w-[110px] h-8">
                <SelectValue placeholder="Range" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="1h">Last 1h</SelectItem>
                <SelectItem value="4h">Last 4h</SelectItem>
                <SelectItem value="6h">Last 6h</SelectItem>
                <SelectItem value="12h">Last 12h</SelectItem>
                <SelectItem value="24h">Last 24h</SelectItem>
                <SelectItem value="7d">Last 7d</SelectItem>
                <SelectItem value="30d">Last 30d</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig}>
          <BarChart accessibilityLayer data={chartData}>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="label"
              tickLine={false}
              tickMargin={10}
              axisLine={false}
            />
            <ChartTooltip
              cursor={false}
              content={<ChartTooltipContent hideLabel />}
            />
            <Bar dataKey="requests" fill="var(--color-requests)" radius={8} />
          </BarChart>
        </ChartContainer>
      </CardContent>
      <CardFooter className="flex-col items-start gap-2 text-sm">
        <div className="flex gap-2 leading-none font-medium">
          Distribution by {viewMode} <TrendingUp className="h-4 w-4" />
        </div>
        <div className="leading-none text-muted-foreground">
          Showing request data for the last{" "}
          {timeRange === "1h"
            ? "hour"
            : timeRange === "24h"
              ? "24 hours"
              : timeRange}
        </div>
      </CardFooter>
    </Card>
  );
}
