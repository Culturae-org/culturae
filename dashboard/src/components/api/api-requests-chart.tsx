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
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { Checkbox } from "@/components/ui/checkbox";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { useApiLogs } from "@/hooks/useApiLogs";
import { cn } from "@/lib/utils";
import { IconActivity, IconFilter, IconRefresh } from "@tabler/icons-react";
import * as React from "react";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";

type TimeGranularity =
  | "minute"
  | "5min"
  | "15min"
  | "hour"
  | "day"
  | "week"
  | "month";
type TimeRange = "1h" | "4h" | "6h" | "12h" | "24h" | "7d" | "30d" | "all";

interface ApiRequestsChartProps {
  className?: string;
}

export function ApiRequestsChart({ className }: ApiRequestsChartProps) {
  const { fetchRequestTimestamps, timestampsLoading: loading, timestampsError: error } = useApiLogs();
  const [chartData, setChartData] = React.useState<
    Array<{ date: string; requests: number; originalDate: Date }>
  >([]);
  const [refreshing, setRefreshing] = React.useState(false);
  const [httpMethod, setHttpMethod] = React.useState<string>("all");
  const [statusCode, setStatusCode] = React.useState<string>("all");
  const [timeRange, setTimeRange] = React.useState<TimeRange>("7d");
  const [isDark, setIsDark] = React.useState(false);

  React.useEffect(() => {
    const checkTheme = () => {
      setIsDark(document.documentElement.classList.contains("dark"));
    };

    checkTheme();

    const observer = new MutationObserver(checkTheme);
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["class"],
    });

    return () => observer.disconnect();
  }, []);

  const granularity = React.useMemo<TimeGranularity>(() => {
    switch (timeRange) {
      case "1h":
        return "minute";
      case "4h":
      case "6h":
        return "5min";
      case "12h":
        return "15min";
      case "24h":
        return "hour";
      case "7d":
        return "day";
      case "30d":
        return "day";
      case "all":
        return "month";
    }
  }, [timeRange]);

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

  const generateTimeBuckets = React.useCallback(
    (start: Date, end: Date, gran: TimeGranularity) => {
      const buckets: Date[] = [];
      const current = new Date(start);

      switch (gran) {
        case "minute":
          current.setSeconds(0, 0);
          break;
        case "5min":
          current.setMinutes(Math.floor(current.getMinutes() / 5) * 5, 0, 0);
          break;
        case "15min":
          current.setMinutes(Math.floor(current.getMinutes() / 15) * 15, 0, 0);
          break;
        case "hour":
          current.setMinutes(0, 0, 0);
          break;
        case "day":
          current.setHours(0, 0, 0, 0);
          break;
        case "week":
          current.setHours(0, 0, 0, 0);
          break;
        case "month":
          current.setDate(1);
          current.setHours(0, 0, 0, 0);
          break;
      }

      while (current <= end) {
        buckets.push(new Date(current));
        switch (gran) {
          case "minute":
            current.setMinutes(current.getMinutes() + 1);
            break;
          case "5min":
            current.setMinutes(current.getMinutes() + 5);
            break;
          case "15min":
            current.setMinutes(current.getMinutes() + 15);
            break;
          case "hour":
            current.setHours(current.getHours() + 1);
            break;
          case "day":
            current.setDate(current.getDate() + 1);
            break;
          case "week":
            current.setDate(current.getDate() + 7);
            break;
          case "month":
            current.setMonth(current.getMonth() + 1);
            break;
        }
      }
      return buckets;
    },
    [],
  );

  const formatDateLabel = React.useCallback(
    (date: Date, gran: TimeGranularity) => {
      switch (gran) {
        case "minute":
        case "5min":
        case "15min":
          return date.toLocaleTimeString("en-US", {
            hour: "2-digit",
            minute: "2-digit",
          });
        case "hour":
          return date.toLocaleTimeString("en-US", {
            hour: "2-digit",
            minute: "2-digit",
          });
        case "day":
          return date.toLocaleDateString("en-US", {
            month: "short",
            day: "numeric",
          });
        case "week":
          return date.toLocaleDateString("en-US", {
            month: "short",
            day: "numeric",
          });
        case "month":
          return date.toLocaleDateString("en-US", {
            month: "short",
            year: "numeric",
          });
        default:
          return date.toISOString();
      }
    },
    [],
  );

  const processData = React.useCallback(
    (timestamps: string[], start: Date, end: Date, gran: TimeGranularity) => {
      const buckets = generateTimeBuckets(start, end, gran);
      const counts = new Map<number, number>();
      for (const b of buckets) {
        counts.set(b.getTime(), 0);
      }

      for (const ts of timestamps) {
        const date = new Date(ts);
        const bucketTime = new Date(date);

        switch (gran) {
          case "minute":
            bucketTime.setSeconds(0, 0);
            break;
          case "5min":
            bucketTime.setMinutes(
              Math.floor(bucketTime.getMinutes() / 5) * 5,
              0,
              0,
            );
            break;
          case "15min":
            bucketTime.setMinutes(
              Math.floor(bucketTime.getMinutes() / 15) * 15,
              0,
              0,
            );
            break;
          case "hour":
            bucketTime.setMinutes(0, 0, 0);
            break;
          case "day":
            bucketTime.setHours(0, 0, 0, 0);
            break;
          case "month":
            bucketTime.setDate(1);
            bucketTime.setHours(0, 0, 0, 0);
            break;
        }

        const t = bucketTime.getTime();

        if (counts.has(t)) {
          counts.set(t, (counts.get(t) || 0) + 1);
        }
      }

      const alignedBuckets: Date[] = [];
      const current = new Date(start);

      switch (gran) {
        case "minute":
          current.setSeconds(0, 0);
          break;
        case "5min":
          current.setMinutes(Math.floor(current.getMinutes() / 5) * 5, 0, 0);
          break;
        case "15min":
          current.setMinutes(Math.floor(current.getMinutes() / 15) * 15, 0, 0);
          break;
        case "hour":
          current.setMinutes(0, 0, 0);
          break;
        case "day":
          current.setHours(0, 0, 0, 0);
          break;
        case "month":
          current.setDate(1);
          current.setHours(0, 0, 0, 0);
          break;
      }

      while (current <= end) {
        alignedBuckets.push(new Date(current));

        switch (gran) {
          case "minute":
            current.setMinutes(current.getMinutes() + 1);
            break;
          case "5min":
            current.setMinutes(current.getMinutes() + 5);
            break;
          case "15min":
            current.setMinutes(current.getMinutes() + 15);
            break;
          case "hour":
            current.setHours(current.getHours() + 1);
            break;
          case "day":
            current.setDate(current.getDate() + 1);
            break;
          case "week":
            current.setDate(current.getDate() + 7);
            break;
          case "month":
            current.setMonth(current.getMonth() + 1);
            break;
        }
      }

      const alignedCounts = new Map<number, number>();
      for (const b of alignedBuckets) {
        alignedCounts.set(b.getTime(), 0);
      }

      for (const ts of timestamps) {
        const date = new Date(ts);
        const bucketTime = new Date(date);

        switch (gran) {
          case "minute":
            bucketTime.setSeconds(0, 0);
            break;
          case "5min":
            bucketTime.setMinutes(
              Math.floor(bucketTime.getMinutes() / 5) * 5,
              0,
              0,
            );
            break;
          case "15min":
            bucketTime.setMinutes(
              Math.floor(bucketTime.getMinutes() / 15) * 15,
              0,
              0,
            );
            break;
          case "hour":
            bucketTime.setMinutes(0, 0, 0);
            break;
          case "day":
            bucketTime.setHours(0, 0, 0, 0);
            break;
          case "month":
            bucketTime.setDate(1);
            bucketTime.setHours(0, 0, 0, 0);
            break;
        }

        const t = bucketTime.getTime();
        if (alignedCounts.has(t)) {
          alignedCounts.set(t, (alignedCounts.get(t) || 0) + 1);
        }
      }

      return alignedBuckets.map((date) => ({
        date: formatDateLabel(date, gran),
        requests: alignedCounts.get(date.getTime()) || 0,
        originalDate: date,
      }));
    },
    [generateTimeBuckets, formatDateLabel],
  );

  const fetchApiRequestData = React.useCallback(
    async (isRefresh = false) => {
      try {
        if (isRefresh) {
          setRefreshing(true);
        }

        const { start, end } = getDateRange(timeRange);
        if (!start || !end) return;

        const timestamps = await fetchRequestTimestamps(
          httpMethod,
          statusCode,
          start,
          end,
        );
        const formattedData = processData(timestamps, start, end, granularity);

        setChartData(formattedData);
      } catch (err) {
        console.error("Error fetching API request data:", err);
      } finally {
        setRefreshing(false);
      }
    },
    [
      fetchRequestTimestamps,
      httpMethod,
      statusCode,
      granularity,
      timeRange,
      getDateRange,
      processData,
    ],
  );

  React.useEffect(() => {
    fetchApiRequestData();
  }, [fetchApiRequestData]);

  const handleRefresh = () => {
    fetchApiRequestData(true);
  };

  if (loading && !refreshing) {
    return (
      <Card className={cn("border-0 dark:border", className)}>
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
      <Card className={cn("border-0 dark:border", className)}>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <div className="space-y-1">
            <CardTitle className="text-base font-medium">
              API Request Volume
            </CardTitle>
            <CardDescription>API request statistics</CardDescription>
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
    <Card className={cn("border-0 dark:border", className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="space-y-1">
          <CardTitle className="text-base font-medium flex items-center">
            API Request Volume
          </CardTitle>
          <CardDescription>
            {chartData.length} data points • {granularity} view
          </CardDescription>
        </div>
        <div className="flex items-center gap-2 flex-wrap">
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
          <Select
            value={timeRange}
            onValueChange={(val) => setTimeRange(val as TimeRange)}
          >
            <SelectTrigger className="w-[110px]" size="sm">
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
              <SelectItem value="all">All time</SelectItem>
            </SelectContent>
          </Select>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm">
                <IconFilter className="h-4 w-4" />
                Filter
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
              <div className="px-2 py-1.5 text-sm font-semibold">
                Method Filter
              </div>
              <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                <Checkbox
                  checked={httpMethod === "GET"}
                  onCheckedChange={(checked) =>
                    setHttpMethod(checked ? "GET" : "all")
                  }
                  className="mr-2"
                />
                GET
              </DropdownMenuItem>
              <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                <Checkbox
                  checked={httpMethod === "POST"}
                  onCheckedChange={(checked) =>
                    setHttpMethod(checked ? "POST" : "all")
                  }
                  className="mr-2"
                />
                POST
              </DropdownMenuItem>
              <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                <Checkbox
                  checked={httpMethod === "PUT"}
                  onCheckedChange={(checked) =>
                    setHttpMethod(checked ? "PUT" : "all")
                  }
                  className="mr-2"
                />
                PUT
              </DropdownMenuItem>
              <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                <Checkbox
                  checked={httpMethod === "PATCH"}
                  onCheckedChange={(checked) =>
                    setHttpMethod(checked ? "PATCH" : "all")
                  }
                  className="mr-2"
                />
                PATCH
              </DropdownMenuItem>
              <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                <Checkbox
                  checked={httpMethod === "DELETE"}
                  onCheckedChange={(checked) =>
                    setHttpMethod(checked ? "DELETE" : "all")
                  }
                  className="mr-2"
                />
                DELETE
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <div className="px-2 py-1.5 text-sm font-semibold">
                Status Filter
              </div>
              <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                <Checkbox
                  checked={statusCode === "200"}
                  onCheckedChange={(checked) =>
                    setStatusCode(checked ? "200" : "all")
                  }
                  className="mr-2"
                />
                200 OK
              </DropdownMenuItem>
              <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                <Checkbox
                  checked={statusCode === "201"}
                  onCheckedChange={(checked) =>
                    setStatusCode(checked ? "201" : "all")
                  }
                  className="mr-2"
                />
                201 Created
              </DropdownMenuItem>
              <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                <Checkbox
                  checked={statusCode === "400"}
                  onCheckedChange={(checked) =>
                    setStatusCode(checked ? "400" : "all")
                  }
                  className="mr-2"
                />
                400 Bad Request
              </DropdownMenuItem>
              <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                <Checkbox
                  checked={statusCode === "401"}
                  onCheckedChange={(checked) =>
                    setStatusCode(checked ? "401" : "all")
                  }
                  className="mr-2"
                />
                401 Unauthorized
              </DropdownMenuItem>
              <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                <Checkbox
                  checked={statusCode === "403"}
                  onCheckedChange={(checked) =>
                    setStatusCode(checked ? "403" : "all")
                  }
                  className="mr-2"
                />
                403 Forbidden
              </DropdownMenuItem>
              <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                <Checkbox
                  checked={statusCode === "404"}
                  onCheckedChange={(checked) =>
                    setStatusCode(checked ? "404" : "all")
                  }
                  className="mr-2"
                />
                404 Not Found
              </DropdownMenuItem>
              <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                <Checkbox
                  checked={statusCode === "500"}
                  onCheckedChange={(checked) =>
                    setStatusCode(checked ? "500" : "all")
                  }
                  className="mr-2"
                />
                500 Server Error
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      <CardContent>
        {chartData.length === 0 ? (
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <IconActivity className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">
                No API request data available for this period
              </p>
            </div>
          </div>
        ) : (
          <ChartContainer
            config={{
              requests: {
                label: "API Requests",
                color: isDark ? "hsl(0 0% 85%)" : "hsl(var(--chart-2))",
              },
            }}
            className="h-64 w-full"
          >
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
                minTickGap={32}
              />
              <YAxis tickLine={false} axisLine={false} tickMargin={8} />
              <ChartTooltip
                cursor={false}
                content={<ChartTooltipContent indicator="line" />}
              />
              <Area
                dataKey="requests"
                type="monotone"
                fill="var(--color-requests)"
                fillOpacity={0.4}
                stroke="var(--color-requests)"
                stackId="a"
              />
            </AreaChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  );
}
