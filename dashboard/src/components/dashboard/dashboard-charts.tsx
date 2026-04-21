"use client";

import { ApiRequestsChart } from "@/components/api/api-requests-chart";
import { GameChart } from "@/components/dashboard/dashboard-game-chart";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { UsersChart } from "@/components/users/user-chart-area-interactive";
import * as React from "react";

type ChartType = "users" | "api" | "games";

export function DashboardCharts() {
  const [activeChart, setActiveChart] = React.useState<ChartType>("users");

  return (
    <div className="px-4 lg:px-6">
      <Tabs
        value={activeChart}
        onValueChange={(v) => setActiveChart(v as ChartType)}
      >
        <TabsList className="mb-4">
          <TabsTrigger value="users">Users</TabsTrigger>
          <TabsTrigger value="api">API</TabsTrigger>
          <TabsTrigger value="games">Games</TabsTrigger>
        </TabsList>

        {activeChart === "users" && <UsersChart />}
        {activeChart === "api" && <ApiRequestsChart />}
        {activeChart === "games" && <GameChart />}
      </Tabs>
    </div>
  );
}
