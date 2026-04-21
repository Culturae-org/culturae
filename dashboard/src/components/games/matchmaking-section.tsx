"use client";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { apiDelete, apiGet } from "@/lib/api-client";
import { MATCHMAKING_ENDPOINTS } from "@/lib/api/endpoints";
import { IconRefresh, IconTrash, IconUsers } from "@tabler/icons-react";
import * as React from "react";
import { toast } from "sonner";

type QueueStats = Record<string, number>;

export function MatchmakingSection() {
  const [stats, setStats] = React.useState<QueueStats | null>(null);
  const [loading, setLoading] = React.useState(true);

  const fetchStats = React.useCallback(async () => {
    try {
      setLoading(true);
      const response = await apiGet(MATCHMAKING_ENDPOINTS.STATS);
      if (!response.ok) {
        throw new Error("Failed to fetch matchmaking stats");
      }
      const json = await response.json();
      setStats(json.data ?? json);
    } catch (err) {
      console.error("Failed to fetch stats:", err);
      toast.error("Failed to load matchmaking stats");
    } finally {
      setLoading(false);
    }
  }, []);

  const clearQueue = async (mode: string) => {
    try {
      const response = await apiDelete(MATCHMAKING_ENDPOINTS.CLEAR_QUEUE(mode));

      if (!response.ok) {
        throw new Error("Failed to clear queue");
      }

      toast.success(`Cleared ${mode} queue`);
      fetchStats();
    } catch (err) {
      console.error("Failed to clear queue:", err);
      toast.error("Failed to clear queue");
    }
  };

  React.useEffect(() => {
    fetchStats();

    const interval = setInterval(fetchStats, 5000);
    return () => clearInterval(interval);
  }, [fetchStats]);

  return (
    <div className="px-4 lg:px-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold">Matchmaking Queues</h2>
        <Button onClick={fetchStats} variant="outline" size="sm">
          <IconRefresh
            className={`h-4 w-4 ${loading ? " animate-spin" : ""}`}
          />
        </Button>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <Card className="border-0 dark:border">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Total Players in Queue
            </CardTitle>
            <IconUsers className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {loading && !stats ? (
                <Skeleton className="h-8 w-20" />
              ) : (
                Object.values(stats ?? {}).reduce((a, b) => a + b, 0)
              )}
            </div>
          </CardContent>
        </Card>

        {["1v1"].map((mode) => (
          <Card key={mode} className="border-0 dark:border">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium capitalize">
                {mode}
              </CardTitle>
              <Badge variant="secondary">
                {loading && !stats ? (
                  <Skeleton className="h-4 w-8 inline-block" />
                ) : (
                  (stats?.[mode] ?? 0)
                )}{" "}
                players
              </Badge>
            </CardHeader>
            <CardContent>
              <div className="flex justify-end">
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={!stats?.[mode]}
                      className="text-destructive hover:text-destructive hover:bg-destructive/10 border-destructive/30"
                    >
                      <IconTrash className="h-3.5 w-3.5 mr-1.5" />
                      Clear Queue
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>Clear {mode} queue?</AlertDialogTitle>
                      <AlertDialogDescription>
                        This will remove all {stats?.[mode] ?? 0} waiting player
                        {(stats?.[mode] ?? 0) > 1 ? "s" : ""} from the {mode}{" "}
                        matchmaking queue. They will be notified and returned to
                        the main menu.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>Cancel</AlertDialogCancel>
                      <AlertDialogAction
                        onClick={() => clearQueue(mode)}
                        className="bg-destructive text-white hover:bg-destructive/90"
                      >
                        Clear Queue
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
