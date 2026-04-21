"use client";

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
import { useDashboardOverview } from "@/hooks/useDashboardOverview";
import {
  IconAlertTriangle,
  IconCheck,
  IconChartBar,
  IconFileText,
  IconGoGame,
  IconPlayerPlay,
  IconTrendingDown,
} from "@tabler/icons-react";
import { useNavigate } from "react-router";

function GamesCardSkeleton() {
  return (
    <Card className="@container/card">
      <CardHeader>
        <Skeleton className="h-8 w-28" />
        <Skeleton className="h-4 w-24" />
        <CardAction>
          <Skeleton className="h-5 w-16" />
        </CardAction>
      </CardHeader>
      <CardFooter className="flex-col items-start gap-2 text-sm">
        <div className="w-full space-y-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex items-center justify-between">
              <Skeleton className="h-4 w-20" />
              <Skeleton className="h-5 w-12" />
            </div>
          ))}
        </div>
      </CardFooter>
    </Card>
  );
}

function DatasetsCardSkeleton() {
  return (
    <Card className="@container/card">
      <CardHeader>
        <Skeleton className="h-8 w-32" />
        <Skeleton className="h-4 w-36" />
        <CardAction>
          <Skeleton className="h-5 w-12" />
        </CardAction>
      </CardHeader>
      <CardFooter className="flex-col items-start gap-2 text-sm">
        <div className="w-full space-y-2">
          {[1, 2].map((i) => (
            <div key={i} className="flex items-center justify-between">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-5 w-8" />
            </div>
          ))}
        </div>
      </CardFooter>
    </Card>
  );
}

function ReportsCardSkeleton() {
  return (
    <Card className="@container/card">
      <CardHeader>
        <Skeleton className="h-8 w-24" />
        <Skeleton className="h-4 w-28" />
        <CardAction>
          <Skeleton className="h-5 w-14" />
        </CardAction>
      </CardHeader>
      <CardFooter className="flex-col items-start gap-2 text-sm">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-10 w-full" />
      </CardFooter>
    </Card>
  );
}

function DashboardOverviewSkeleton() {
  return (
    <div className="dark:*:data-[slot=card]:bg-card grid grid-cols-1 gap-4 px-4 *:data-[slot=card]:shadow-xs lg:px-6 @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
      <GamesCardSkeleton />
      <DatasetsCardSkeleton />
      <ReportsCardSkeleton />
      <ReportsCardSkeleton />
    </div>
  );
}

export function SectionCardsOne() {
  const navigate = useNavigate();
  const { data, loading } = useDashboardOverview();
  const { games, datasets, reports } = data;

  if (loading) {
    return <DashboardOverviewSkeleton />;
  }

  return (
    <div className="dark:*:data-[slot=card]:bg-card grid grid-cols-1 gap-4 px-4 *:data-[slot=card]:shadow-xs lg:px-6 @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
      <Card
        className="@container/card cursor-pointer hover:shadow-md transition-shadow"
        onClick={() => navigate("/games")}
      >
        <CardHeader>
          <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
            Games
          </CardTitle>
          <CardDescription className="flex items-center gap-2">
            <IconGoGame className="size-4 text-muted-foreground" />
            Game Statistics
          </CardDescription>
          <CardAction>
            <Badge variant="outline">{games.popularMode || "Games"}</Badge>
          </CardAction>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-2 text-sm">
          <div className="w-full space-y-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span className="font-medium">Total Games</span>
              </div>
              <Badge variant="outline" className="font-mono">
                {games.total.toLocaleString()}
              </Badge>
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <IconPlayerPlay className="size-4 text-muted-foreground" />
                <span className="font-medium">Active</span>
              </div>
              <Badge className="bg-gray-100 text-gray-800 border-gray-200">
                {games.active.toLocaleString()}
              </Badge>
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <IconCheck className="size-4 text-muted-foreground" />
                <span className="font-medium">Completed</span>
              </div>
              <Badge className="bg-gray-100 text-gray-800 border-gray-200">
                {games.completed.toLocaleString()}
              </Badge>
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <IconTrendingDown className="size-4 text-muted-foreground" />
                <span className="font-medium">Abandoned</span>
              </div>
              <Badge className="bg-gray-100 text-gray-800 border-gray-200">
                {games.abandoned.toLocaleString()}
              </Badge>
            </div>
          </div>
        </CardFooter>
      </Card>

      <Card
        className="@container/card cursor-pointer hover:shadow-md transition-shadow"
        onClick={() => navigate("/datasets")}
      >
        <CardHeader>
          <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
            Datasets
          </CardTitle>
          <CardDescription className="flex items-center gap-2">
            <IconFileText className="size-4 text-muted-foreground" />
            Question & Geography
          </CardDescription>
          <CardAction>
            <Badge variant="outline">
              {datasets.questions + datasets.geography} total
            </Badge>
          </CardAction>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-2 text-sm">
          <div className="w-full space-y-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center">
                <span className="font-medium">Question Datasets</span>
              </div>
              <Badge className="bg-gray-100 text-gray-800 border-gray-200">
                {datasets.questions}
              </Badge>
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center">
                <span className="font-medium">Geography Datasets</span>
              </div>
              <Badge className="bg-gray-100 text-gray-800 border-gray-200">
                {datasets.geography}
              </Badge>
            </div>
          </div>
        </CardFooter>
      </Card>

      <Card
        className="@container/card cursor-pointer hover:shadow-md transition-shadow"
        onClick={() => navigate("/reports")}
      >
        <CardHeader>
          <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
            Reports
          </CardTitle>
          <CardDescription className="flex items-center gap-2">
            <IconAlertTriangle className="size-4 text-muted-foreground" />
            User Reports
          </CardDescription>
          <CardAction>
            <Badge variant="outline">
              <IconAlertTriangle className="size-3 mr-1" />
              Needs attention
            </Badge>
          </CardAction>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-2 text-sm">
          <div className="w-full">
            <div className="flex items-center justify-between mb-2">
              <span className="font-medium">Pending Reviews</span>
              <Badge className="bg-gray-100 text-gray-800 border-gray-200">
                {reports.pending}
              </Badge>
            </div>
            <div className="w-full bg-muted rounded-full h-2">
              <div
                className="bg-gray-500 h-2 rounded-full"
                style={{ width: reports.pending > 0 ? "100%" : "0%" }}
              />
            </div>
          </div>
        </CardFooter>
      </Card>

      <Card
        className="@container/card cursor-pointer hover:shadow-md transition-shadow"
        onClick={() => navigate("/analytics")}
      >
        <CardHeader>
          <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
            Analytics
          </CardTitle>
          <CardDescription className="flex items-center gap-2">
            <IconChartBar className="size-4 text-muted-foreground" />
            Platform Overview
          </CardDescription>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-2 text-sm">
          <p className="text-muted-foreground text-sm">
            Users, API metrics, service health and activity charts.
          </p>
        </CardFooter>
      </Card>
    </div>
  );
}
