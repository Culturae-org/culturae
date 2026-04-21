"use client";

import { GameViewDialog } from "@/components/games/game-view-dialog";
import { GamesCharts } from "@/components/games/games-charts";
import { GamesDataTable } from "@/components/games/games-data-table";
import { GamesStatsCards } from "@/components/games/games-stats-cards";
import { MatchmakingSection } from "@/components/games/matchmaking-section";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import type { AdminGame } from "@/lib/types/games.types";
import * as React from "react";
import { useNavigate, useSearchParams } from "react-router";

export default function GamesPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const viewId = searchParams.get("view");

  const [activeTab, setActiveTab] = React.useState("overview");

  React.useEffect(() => {
    if (viewId) setActiveTab("games");
  }, [viewId]);

  const handleViewClose = (open: boolean) => {
    if (!open) {
      const params = new URLSearchParams(searchParams.toString());
      params.delete("view");
      const qs = params.toString();
      navigate(qs ? `/games?${qs}` : "/games");
    }
  };

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <h1 className="text-3xl font-bold">Games Management</h1>
        <p className="text-muted-foreground">
          Monitor and manage all games on your platform
        </p>
      </div>

      <Tabs
        value={activeTab}
        onValueChange={setActiveTab}
        className="px-4 lg:px-6"
      >
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="games">Games</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          <GamesStatsCards />
          <MatchmakingSection />
          <GamesCharts />
        </TabsContent>

        <TabsContent value="games">
          <GamesDataTable />
        </TabsContent>
      </Tabs>

      {viewId && (
        <GameViewDialog
          game={{ id: viewId } as AdminGame}
          open={true}
          onOpenChange={handleViewClose}
        />
      )}
    </div>
  );
}
