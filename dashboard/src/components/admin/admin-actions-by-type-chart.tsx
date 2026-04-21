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
import { Cell, Label, Pie, PieChart } from "recharts";

function getSliceColor(index: number, total: number): string {
  const pct = Math.round(100 - (index / Math.max(total - 1, 1)) * 68);
  return `color-mix(in oklch, var(--chart-1) ${pct}%, transparent)`;
}

export function AdminActionsByTypeChart() {
  const [chartData, setChartData] = React.useState<
    Array<{ action: string; count: number }>
  >([]);
  const [totalActions, setTotalActions] = React.useState(0);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    logsService
      .getAdminActionStats()
      .then((stats) => {
        setTotalActions(stats.total_actions as unknown as number);

        const sorted = Object.entries(stats.actions_by_type ?? {})
          .map(([action, count]) => ({
            action: action.replace(/_/g, " "),
            count: count as number,
          }))
          .sort((a, b) => b.count - a.count);

        const top = sorted.slice(0, 5);
        const otherCount = sorted
          .slice(5)
          .reduce((acc, cur) => acc + cur.count, 0);

        const data = [...top];
        if (otherCount > 0) data.push({ action: "other", count: otherCount });

        setChartData(data);
      })
      .catch((err) => {
        console.error("AdminActionsByTypeChart:", err);
        setError(err instanceof Error ? err.message : "Failed to load stats");
      })
      .finally(() => setLoading(false));
  }, []);

  const chartConfig = React.useMemo<ChartConfig>(() => {
    const config: ChartConfig = { count: { label: "Actions" } };
    chartData.forEach((item, i) => {
      config[item.action] = {
        label: item.action,
        color: getSliceColor(i, chartData.length),
      };
    });
    return config;
  }, [chartData]);

  return (
    <div className="w-full lg:w-1/2">
      <Card className="border-0 dark:border flex flex-col">
        <CardHeader>
          <CardTitle>Actions by Type</CardTitle>
          <CardDescription>Most frequent admin action types</CardDescription>
        </CardHeader>
        <CardContent className="flex-1 pb-0">
          {loading ? (
            <div className="mx-auto aspect-square max-h-[460px] flex items-center justify-center">
              <Skeleton className="h-[420px] w-[420px] rounded-full" />
            </div>
          ) : error ? (
            <div className="mx-auto aspect-square max-h-[460px] flex items-center justify-center text-destructive text-sm">
              {error}
            </div>
          ) : chartData.length === 0 ? (
            <div className="mx-auto aspect-square max-h-[460px] flex items-center justify-center text-muted-foreground text-sm">
              No admin actions recorded yet
            </div>
          ) : (
            <ChartContainer
              config={chartConfig}
              className="mx-auto aspect-square max-h-[460px]"
            >
              <PieChart>
                <ChartTooltip
                  cursor={false}
                  content={<ChartTooltipContent hideLabel />}
                />
                <Pie
                  data={chartData}
                  dataKey="count"
                  nameKey="action"
                  innerRadius="40%"
                  strokeWidth={2}
                  stroke="transparent"
                >
                  {chartData.map((item, i) => (
                    <Cell
                      key={item.action ?? i}
                      fill={getSliceColor(i, chartData.length)}
                    />
                  ))}
                  <Label
                    content={({ viewBox }) => {
                      if (viewBox && "cx" in viewBox && "cy" in viewBox) {
                        return (
                          <text
                            x={viewBox.cx}
                            y={viewBox.cy}
                            textAnchor="middle"
                            dominantBaseline="middle"
                          >
                            <tspan
                              x={viewBox.cx}
                              y={viewBox.cy}
                              className="fill-foreground text-3xl font-bold"
                            >
                              {totalActions.toLocaleString()}
                            </tspan>
                            <tspan
                              x={viewBox.cx}
                              y={(viewBox.cy || 0) + 24}
                              className="fill-muted-foreground text-sm"
                            >
                              Actions
                            </tspan>
                          </text>
                        );
                      }
                    }}
                  />
                </Pie>
              </PieChart>
            </ChartContainer>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
