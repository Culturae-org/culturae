import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardAction,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import type { PodsDiscovery } from "@/lib/types/metrics.types";

interface PodStatsCardsProps {
  podsDiscovery: PodsDiscovery | null;
  totalActiveGames: number;
  totalLoadScore: number;
}

export function PodStatsCards({ podsDiscovery, totalActiveGames, totalLoadScore }: PodStatsCardsProps) {
  return (
    <div className="dark:*:data-[slot=card]:bg-card grid grid-cols-2 gap-4 px-4 *:data-[slot=card]:shadow-xs lg:px-6 @xl/main:grid-cols-4">
      <Card className="@container/card">
        <CardHeader>
          <CardDescription>Total Pods</CardDescription>
          <CardTitle className="text-3xl font-semibold tabular-nums">
            {podsDiscovery?.meta.total_pods ?? 0}
          </CardTitle>
          <CardAction>
            <Badge variant="outline">
              {podsDiscovery?.meta.total_pods === 1 ? "instance" : "instances"}
            </Badge>
          </CardAction>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-1 text-sm">
          <div className="flex items-center justify-between w-full">
            <span className="text-muted-foreground text-xs">Main</span>
            <Badge variant="outline" className="font-mono text-xs">
              {podsDiscovery?.meta.main_pods ?? 0}
            </Badge>
          </div>
          <div className="flex items-center justify-between w-full">
            <span className="text-muted-foreground text-xs">Headless</span>
            <Badge variant="outline" className="font-mono text-xs">
              {podsDiscovery?.meta.headless_pods ?? 0}
            </Badge>
          </div>
        </CardFooter>
      </Card>

      <Card className="@container/card">
        <CardHeader>
          <CardDescription>Online Users</CardDescription>
          <CardTitle className="text-3xl font-semibold tabular-nums">
            {podsDiscovery?.meta.total_users ?? 0}
          </CardTitle>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-1 text-sm">
          <div className="flex items-center justify-between w-full">
            <span className="text-muted-foreground text-xs">WS connections</span>
            <Badge variant="outline" className="font-mono text-xs">
              {podsDiscovery?.meta.total_clients ?? 0}
            </Badge>
          </div>
        </CardFooter>
      </Card>

      <Card className="@container/card">
        <CardHeader>
          <CardDescription>Active Games</CardDescription>
          <CardTitle className="text-3xl font-semibold tabular-nums">
            {totalActiveGames}
          </CardTitle>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-1 text-sm">
          <p className="text-muted-foreground text-xs">Games in progress across all pods</p>
        </CardFooter>
      </Card>

      <Card className="@container/card">
        <CardHeader>
          <CardDescription>Cluster Load</CardDescription>
          <CardTitle className="text-3xl font-semibold tabular-nums">
            {totalLoadScore}
          </CardTitle>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-1 text-sm">
          <p className="text-muted-foreground text-xs">
            Routing score — clients + games × 2
          </p>
        </CardFooter>
      </Card>
    </div>
  );
}

export function PodStatsCardsSkeleton() {
  return (
    <div className="dark:*:data-[slot=card]:bg-card grid grid-cols-2 gap-4 px-4 *:data-[slot=card]:shadow-xs lg:px-6 @xl/main:grid-cols-4">
      {[1, 2, 3, 4].map((i) => (
        <Card key={i} className="@container/card">
          <CardHeader>
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-8 w-16" />
          </CardHeader>
          <CardFooter>
            <Skeleton className="h-3 w-32" />
          </CardFooter>
        </Card>
      ))}
    </div>
  );
}
