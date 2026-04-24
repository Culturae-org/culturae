"use client";

import { PodStatsCards, PodStatsCardsSkeleton, PodTable, usePods } from "@/components/pods";
import { Button } from "@/components/ui/button";
import { IconRefresh } from "@tabler/icons-react";

export default function PodsPage() {
  const {
    podsDiscovery,
    loading,
    lastUpdated,
    error,
    mainPods,
    headlessPods,
    totalActiveGames,
    totalLoadScore,
    refresh,
  } = usePods(5000);

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold">Pod Instances</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              Real-time monitoring of all running instances
              {lastUpdated && (
                <span className="ml-2 text-xs">· updated {lastUpdated.toLocaleTimeString()}</span>
              )}
            </p>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={refresh}
            disabled={loading}
            className="self-start sm:self-auto"
          >
            <IconRefresh className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
          </Button>
        </div>

        {error && (
          <div className="mt-3 rounded-md border border-red-200 bg-red-50 px-4 py-2 text-sm text-red-700 dark:border-red-900 dark:bg-red-950/40 dark:text-red-400">
            Failed to fetch pods: {error}
          </div>
        )}
      </div>

      {loading && !podsDiscovery ? (
        <PodStatsCardsSkeleton />
      ) : (
        <PodStatsCards
          podsDiscovery={podsDiscovery}
          totalActiveGames={totalActiveGames}
          totalLoadScore={totalLoadScore}
        />
      )}

      <PodTable
        pods={podsDiscovery?.pods ?? []}
        mainPods={mainPods}
        headlessPods={headlessPods}
        loading={loading}
      />
    </div>
  );
}
