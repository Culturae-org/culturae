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
import { logsService } from "@/lib/services/logs.service";
import * as React from "react";
import { Bar, BarChart, XAxis } from "recharts";

const METHODS = ["GET", "POST", "PUT", "PATCH", "DELETE"] as const;

const chartConfig = {
  GET: { label: "GET", color: "var(--chart-1)" },
  POST: { label: "POST", color: "var(--chart-2)" },
  PUT: { label: "PUT", color: "var(--chart-3)" },
  PATCH: { label: "PATCH", color: "var(--chart-4)" },
  DELETE: { label: "DELETE", color: "var(--chart-5)" },
} satisfies ChartConfig;

type DayEntry = {
  date: string;
  GET: number;
  POST: number;
  PUT: number;
  PATCH: number;
  DELETE: number;
};

function bucketByDay(timestamps: string[]): Map<string, number> {
  const map = new Map<string, number>();
  for (const ts of timestamps) {
    const day = new Date(ts).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
    });
    map.set(day, (map.get(day) ?? 0) + 1);
  }
  return map;
}

function last7DayLabels(): string[] {
  const labels: string[] = [];
  for (let i = 6; i >= 0; i--) {
    const d = new Date();
    d.setDate(d.getDate() - i);
    labels.push(
      d.toLocaleDateString("en-US", { month: "short", day: "numeric" }),
    );
  }
  return labels;
}

export function ApiRequestsByMethodChart() {
  const [chartData, setChartData] = React.useState<DayEntry[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    const start = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000);
    const end = new Date();
    const params = {
      start_date: start.toISOString(),
      end_date: end.toISOString(),
    };

    Promise.all(
      METHODS.map((method) =>
        logsService.getAPIRequestTimestamps({ ...params, method }),
      ),
    )
      .then((results) => {
        const days = last7DayLabels();
        const byMethod = Object.fromEntries(
          METHODS.map((method, i) => [method, bucketByDay(results[i])]),
        ) as Record<(typeof METHODS)[number], Map<string, number>>;

        const data: DayEntry[] = days.map((day) => ({
          date: day,
          GET: byMethod.GET.get(day) ?? 0,
          POST: byMethod.POST.get(day) ?? 0,
          PUT: byMethod.PUT.get(day) ?? 0,
          PATCH: byMethod.PATCH.get(day) ?? 0,
          DELETE: byMethod.DELETE.get(day) ?? 0,
        }));
        setChartData(data);
      })
      .catch((err) => {
        console.error("ApiRequestsByMethodChart:", err);
        setError(err instanceof Error ? err.message : "Failed to load stats");
      })
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="w-full lg:w-1/2">
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Requests by Method</CardTitle>
          <CardDescription>
            Last 7 days — breakdown by HTTP method
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <Skeleton className="h-[300px] w-full" />
          ) : error ? (
            <div className="h-[300px] flex items-center justify-center text-destructive text-sm">
              {error}
            </div>
          ) : (
            <ChartContainer config={chartConfig} className="h-[300px] w-full">
              <BarChart accessibilityLayer data={chartData}>
                <XAxis
                  dataKey="date"
                  tickLine={false}
                  tickMargin={10}
                  axisLine={false}
                  tickFormatter={(value) => {
                    const parts = value.split(" ");
                    const date = new Date(
                      `${parts[0]} ${parts[1]}, ${new Date().getFullYear()}`,
                    );
                    return date.toLocaleDateString("en-US", {
                      weekday: "short",
                    });
                  }}
                />
                <Bar
                  dataKey="GET"
                  stackId="a"
                  fill="var(--color-GET)"
                  radius={[0, 0, 4, 4]}
                />
                <Bar
                  dataKey="POST"
                  stackId="a"
                  fill="var(--color-POST)"
                  radius={[0, 0, 0, 0]}
                />
                <Bar
                  dataKey="PUT"
                  stackId="a"
                  fill="var(--color-PUT)"
                  radius={[0, 0, 0, 0]}
                />
                <Bar
                  dataKey="PATCH"
                  stackId="a"
                  fill="var(--color-PATCH)"
                  radius={[0, 0, 0, 0]}
                />
                <Bar
                  dataKey="DELETE"
                  stackId="a"
                  fill="var(--color-DELETE)"
                  radius={[4, 4, 0, 0]}
                />
                <ChartTooltip
                  cursor={false}
                  defaultIndex={1}
                  content={
                    <ChartTooltipContent
                      hideLabel
                      className="w-[180px]"
                      formatter={(value, name, item, index) => (
                        <>
                          <div
                            className="h-2.5 w-2.5 shrink-0 rounded-[2px] bg-(--color-bg)"
                            style={
                              {
                                "--color-bg": `var(--color-${name})`,
                              } as React.CSSProperties
                            }
                          />
                          {chartConfig[name as keyof typeof chartConfig]
                            ?.label || name}
                          <div className="ml-auto flex items-baseline gap-0.5 font-mono font-medium text-foreground tabular-nums">
                            {value}
                          </div>
                          {index === METHODS.length - 1 && (
                            <div className="mt-1.5 flex basis-full items-center border-t pt-1.5 text-xs font-medium text-foreground">
                              Total
                              <div className="ml-auto flex items-baseline gap-0.5 font-mono font-medium text-foreground tabular-nums">
                                {item.payload.GET +
                                  item.payload.POST +
                                  item.payload.PUT +
                                  item.payload.PATCH +
                                  item.payload.DELETE}
                              </div>
                            </div>
                          )}
                        </>
                      )}
                    />
                  }
                />
              </BarChart>
            </ChartContainer>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
