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
import { Pie, PieChart } from "recharts";

const STATUS_GROUPS = ["2xx", "3xx", "4xx", "5xx"] as const;
type StatusGroup = (typeof STATUS_GROUPS)[number];

const chartConfig = {
  count: { label: "Requests" },
  "2xx": { label: "2xx Success", color: "var(--chart-1)" },
  "3xx": { label: "3xx Redirect", color: "var(--chart-2)" },
  "4xx": { label: "4xx Client Error", color: "var(--chart-3)" },
  "5xx": { label: "5xx Server Error", color: "var(--chart-4)" },
} satisfies ChartConfig;

export function ApiRequestsByStatusChart() {
  const [chartData, setChartData] = React.useState<
    Array<{ status: StatusGroup; count: number; fill: string }>
  >([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    logsService
      .getAPIRequestStats()
      .then((stats) => {
        const groups: Record<StatusGroup, number> = {
          "2xx": 0,
          "3xx": 0,
          "4xx": 0,
          "5xx": 0,
        };
        
        const requestsByStatus = stats?.requests_by_status;
        
        if (requestsByStatus) {
          for (const [code, count] of Object.entries(requestsByStatus)) {
            const family = `${code[0]}xx` as StatusGroup;
            if (family in groups) groups[family] += Number(count);
          }
        }
        
        const data = STATUS_GROUPS.filter((g) => groups[g] > 0).map((g) => ({
          status: g,
          count: groups[g],
          fill: `var(--color-${g})`,
        }));
        
        setChartData(data);
      })
      .catch((err) => {
        console.error("ApiRequestsByStatusChart:", err);
        setError(err instanceof Error ? err.message : "Failed to load stats");
      })
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="w-full lg:w-1/2">
      <Card className="border-0 dark:border flex flex-col">
        <CardHeader className="items-center pb-0">
          <CardTitle>Requests by Status</CardTitle>
          <CardDescription>Distribution by HTTP status group</CardDescription>
        </CardHeader>
        <CardContent className="flex-1 pb-0">
          {loading ? (
            <div className="mx-auto aspect-square max-h-[300px] flex items-center justify-center">
              <Skeleton className="h-[260px] w-[260px] rounded-full" />
            </div>
          ) : error ? (
            <div className="mx-auto aspect-square max-h-[300px] flex items-center justify-center text-destructive text-sm">
              {error}
            </div>
          ) : chartData.length === 0 ? (
            <div className="mx-auto aspect-square max-h-[300px] flex items-center justify-center text-muted-foreground text-sm">
              No data available
            </div>
          ) : (
            <ChartContainer
              config={chartConfig}
              className="mx-auto aspect-square max-h-[300px] pb-0 [&_.recharts-pie-label-text]:fill-foreground"
            >
              <PieChart>
                <ChartTooltip content={<ChartTooltipContent hideLabel />} />
                <Pie data={chartData} dataKey="count" nameKey="status" label />
              </PieChart>
            </ChartContainer>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
