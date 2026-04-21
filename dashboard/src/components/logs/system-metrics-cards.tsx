"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { SystemMetrics } from "@/lib/types/metrics.types";
import {
  IconActivity,
  IconClock,
  IconTrendingUp,
  IconUsers,
} from "@tabler/icons-react";

interface SystemMetricsCardsProps {
  systemMetrics: SystemMetrics | null;
  onlineUsers: number;
}

export function SystemMetricsCards({
  systemMetrics,
  onlineUsers,
}: SystemMetricsCardsProps) {
  if (!systemMetrics) return null;

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <Card className="border-0 dark:border">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Total Users</CardTitle>
          <IconUsers className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{systemMetrics.total_users}</div>
          <p className="text-xs text-muted-foreground">{onlineUsers} online</p>
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">
            API Requests (24h)
          </CardTitle>
          <IconTrendingUp className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {systemMetrics.total_api_requests}
          </div>
          <p className="text-xs text-muted-foreground">
            {systemMetrics.error_rate.toFixed(1)}% error rate
          </p>
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Active Sessions</CardTitle>
          <IconActivity className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {systemMetrics.active_sessions}
          </div>
          <p className="text-xs text-muted-foreground">
            of {systemMetrics.total_sessions} total
          </p>
        </CardContent>
      </Card>

      <Card className="border-0 dark:border">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">
            Avg Response Time
          </CardTitle>
          <IconClock className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {systemMetrics.avg_response_time_ms.toFixed(1)}ms
          </div>
          <p className="text-xs text-muted-foreground">last 24 hours</p>
        </CardContent>
      </Card>
    </div>
  );
}
