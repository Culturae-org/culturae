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
import { PolarAngleAxis, PolarGrid, Radar, RadarChart } from "recharts";

const chartConfig = {
  actions: {
    label: "Actions",
    color: "var(--chart-1)",
  },
} satisfies ChartConfig;

export function AdminActionsByResourceChart() {
  const [chartData, setChartData] = React.useState<
    Array<{ resource: string; actions: number }>
  >([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    logsService
      .getAdminActionStats()
      .then((stats) => {
        const GEO_RESOURCES = new Set([
          "geography_dataset",
          "country",
          "continent",
          "region",
        ]);
        const merged: Record<string, number> = {};
        for (const [resource, count] of Object.entries(
          stats.actions_by_resource ?? {},
        )) {
          const key = GEO_RESOURCES.has(resource)
            ? "Geography"
            : resource
                .replace(/_/g, " ")
                .replace(/\b\w/g, (c) => c.toUpperCase());
          merged[key] = (merged[key] ?? 0) + (count as number);
        }
        const data = Object.entries(merged)
          .map(([resource, actions]) => ({ resource, actions }))
          .sort((a, b) => b.actions - a.actions);
        setChartData(data);
      })
      .catch((err) => {
        console.error("AdminActionsByResourceChart:", err);
        setError(err instanceof Error ? err.message : "Failed to load stats");
      })
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="w-full lg:w-1/2">
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Actions by Resource</CardTitle>
          <CardDescription>
            Distribution of admin actions across resource types
          </CardDescription>
        </CardHeader>
        <CardContent>
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
              <RadarChart
                data={chartData}
                margin={{ top: 16, right: 48, bottom: 16, left: 48 }}
                outerRadius="65%"
              >
                <ChartTooltip
                  cursor={false}
                  content={<ChartTooltipContent />}
                />
                <PolarAngleAxis dataKey="resource" tick={{ fontSize: 11 }} />
                <PolarGrid />
                <Radar
                  dataKey="actions"
                  fill="var(--color-actions)"
                  fillOpacity={0.6}
                />
              </RadarChart>
            </ChartContainer>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
