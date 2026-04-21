"use client";

import { useSystemMonitoring, useUsersStats } from "@/hooks";
import type { SystemMetrics } from "@/lib/types/stats.types";
import {
  IconActivity,
  IconAlertTriangle,
  IconCircleCheck,
  IconCircleX,
  IconClock,
  IconCloud,
  IconCrown,
  IconDatabase,
  IconServer,
  IconShield,
  IconTrendingDown,
  IconTrendingUp,
  IconUser,
  IconUsers,
} from "@tabler/icons-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router";

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

function TotalUsersCardSkeleton() {
  return (
    <Card className="@container/card">
      <CardHeader>
        <CardDescription>Total Users</CardDescription>
        <Skeleton className="h-8 w-24" />
        <CardAction>
          <Skeleton className="h-5 w-16" />
        </CardAction>
      </CardHeader>
      <CardFooter className="flex-col items-start gap-1.5 text-sm">
        <Skeleton className="h-4 w-40" />
        <Skeleton className="h-4 w-24" />
        <div className="w-full space-y-2 mt-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Skeleton className="size-4 rounded-full" />
                <Skeleton className="h-4 w-16" />
              </div>
              <div className="flex items-center gap-2">
                <Skeleton className="w-16 h-2 rounded-full" />
                <Skeleton className="h-5 w-9" />
              </div>
            </div>
          ))}
        </div>
      </CardFooter>
    </Card>
  );
}

function UserLevelCardSkeleton() {
  return (
    <Card className="@container/card">
      <CardHeader>
        <Skeleton className="h-8 w-28" />
        <div className="flex items-center gap-2">
          <Skeleton className="size-4 rounded-full" />
          <Skeleton className="h-4 w-24" />
        </div>
        <CardAction>
          <Skeleton className="h-5 w-16" />
        </CardAction>
      </CardHeader>
      <CardFooter className="flex-col items-start gap-2 text-sm">
        <div className="w-full space-y-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Skeleton className="size-4 rounded-full" />
                <Skeleton className="h-4 w-20" />
              </div>
              <div className="flex items-center gap-2">
                <Skeleton className="w-16 h-2 rounded-full" />
                <Skeleton className="h-5 w-9" />
              </div>
            </div>
          ))}
        </div>
      </CardFooter>
    </Card>
  );
}

function SystemServicesCardSkeleton() {
  return (
    <Card className="@container/card">
      <CardHeader>
        <Skeleton className="h-8 w-36" />
        <div className="flex items-center gap-2">
          <Skeleton className="size-4 rounded-full" />
          <Skeleton className="h-4 w-32" />
        </div>
        <CardAction>
          <Skeleton className="h-5 w-20" />
        </CardAction>
      </CardHeader>
      <CardFooter className="flex-col items-start gap-2 text-sm">
        <div className="w-full space-y-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Skeleton className="size-4 rounded-full" />
                <Skeleton className="h-4 w-16" />
              </div>
              <div className="flex items-center gap-2">
                <Skeleton className="h-5 w-16" />
                <Skeleton className="h-4 w-10" />
              </div>
            </div>
          ))}
        </div>
      </CardFooter>
    </Card>
  );
}

function ApiMetricsCardSkeleton() {
  return (
    <Card className="@container/card">
      <CardHeader>
        <Skeleton className="h-8 w-28" />
        <div className="flex items-center gap-2">
          <Skeleton className="size-4 rounded-full" />
          <Skeleton className="h-4 w-36" />
        </div>
        <CardAction>
          <Skeleton className="h-5 w-12" />
        </CardAction>
      </CardHeader>
      <CardFooter className="flex-col items-start gap-2 text-sm">
        <div className="w-full space-y-3">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Skeleton className="size-4 rounded-full" />
                <Skeleton className="h-4 w-28" />
              </div>
              <Skeleton className="h-5 w-16" />
            </div>
          ))}
        </div>
      </CardFooter>
    </Card>
  );
}

function SectionCardsSkeleton() {
  return (
    <div className="dark:*:data-[slot=card]:bg-card grid grid-cols-1 gap-4 px-4 *:data-[slot=card]:shadow-xs lg:px-6 @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
      <TotalUsersCardSkeleton />
      <UserLevelCardSkeleton />
      <SystemServicesCardSkeleton />
      <ApiMetricsCardSkeleton />
    </div>
  );
}

const levelColors = {
  beginner: "bg-green-100 text-green-800 border-green-200",
  intermediate: "bg-blue-100 text-blue-800 border-blue-200",
  advanced: "bg-purple-100 text-purple-800 border-purple-200",
  expert: "bg-orange-100 text-orange-800 border-orange-200",
  master: "bg-red-100 text-red-800 border-red-200",
};

const roleColors = {
  user: "bg-gray-100 text-gray-800 border-gray-200",
  moderator: "bg-yellow-100 text-yellow-800 border-yellow-200",
  admin: "bg-red-100 text-red-800 border-red-200",
  superadmin: "bg-purple-100 text-purple-800 border-purple-200",
};

const getLevelColor = (level: string) =>
  levelColors[level as keyof typeof levelColors] ||
  "bg-gray-100 text-gray-800 border-gray-200";
const getRoleColor = (role: string) =>
  roleColors[role as keyof typeof roleColors] ||
  "bg-gray-100 text-gray-800 border-gray-200";

const getLevelIcon = (level: string) => {
  switch (level.toLowerCase()) {
    case "beginner":
      return <IconUser className="size-4" />;
    case "intermediate":
      return <IconUsers className="size-4" />;
    case "advanced":
      return <IconShield className="size-4" />;
    case "expert":
      return <IconCrown className="size-4" />;
    case "master":
      return <IconCrown className="size-4" />;
    default:
      return <IconUser className="size-4" />;
  }
};

const getRoleIcon = (role: string) => {
  switch (role.toLowerCase()) {
    case "user":
      return <IconUser className="size-4" />;
    case "moderator":
      return <IconShield className="size-4" />;
    case "admin":
      return <IconCrown className="size-4" />;
    case "superadmin":
      return <IconCrown className="size-4" />;
    default:
      return <IconUser className="size-4" />;
  }
};

const getServiceIcon = (serviceName: string) => {
  switch (serviceName.toLowerCase()) {
    case "postgres":
      return <IconDatabase className="size-4" />;
    case "redis":
      return <IconServer className="size-4" />;
    case "minio":
      return <IconCloud className="size-4" />;
    default:
      return <IconServer className="size-4" />;
  }
};

const getServiceStatusColor = (status: string) => {
  switch (status.toLowerCase()) {
    case "healthy":
      return "bg-green-100 text-green-800 border-green-200";
    case "unhealthy":
      return "bg-red-100 text-red-800 border-red-200";
    default:
      return "bg-gray-100 text-gray-800 border-gray-200";
  }
};

const getServiceStatusIcon = (status: string) => {
  switch (status.toLowerCase()) {
    case "healthy":
      return <IconCircleCheck className="size-4" />;
    case "unhealthy":
      return <IconCircleX className="size-4" />;
    default:
      return <IconAlertTriangle className="size-4" />;
  }
};

export function SectionCardsTwo() {
  const navigate = useNavigate();
  const {
    fetchUserCount,
    fetchLevelStats,
    fetchRoleStats,
    fetchCreationDates,
  } = useUsersStats();
  const { fetchServiceStatus, fetchSystemMetrics } = useSystemMonitoring();

  const [loading, setLoading] = useState(true);
  const [userCount, setUserCount] = useState<number>(0);
  const [levelStats, setLevelStats] = useState<Record<string, number>>({});
  const [roleStats, setRoleStats] = useState<Record<string, number>>({});
  const [trendingPercentage, setTrendingPercentage] = useState<number>(0);
  const [isTrendingUp, setIsTrendingUp] = useState<boolean>(true);
  const [serviceStatus, setServiceStatus] = useState<
    Array<{
      service_name: string;
      status: string;
      last_check: string;
      response_time_ms: number;
      error_msg?: string;
    }>
  >([]);
  const [systemMetrics, setSystemMetrics] = useState<SystemMetrics | null>(
    null,
  );

  useEffect(() => {
    const loadData = async () => {
      try {
        const [
          countData,
          levelData,
          roleData,
          serviceData,
          metricsData,
          creationDates,
        ] = await Promise.all([
          fetchUserCount().catch(() => null),
          fetchLevelStats().catch(() => null),
          fetchRoleStats().catch(() => null),
          fetchServiceStatus().catch(() => null),
          fetchSystemMetrics().catch(() => null),
          fetchCreationDates().catch(() => null),
        ]);

        setUserCount(countData || 0);

        if (creationDates && creationDates.length > 0) {
          const now = new Date();
          const sevenDaysAgo = new Date(now);
          sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);
          const fourteenDaysAgo = new Date(now);
          fourteenDaysAgo.setDate(fourteenDaysAgo.getDate() - 14);

          const usersLast7Days = creationDates.filter((dateString: string) => {
            const date = new Date(dateString);
            return date >= sevenDaysAgo && date <= now;
          }).length;

          const usersPrevious7Days = creationDates.filter(
            (dateString: string) => {
              const date = new Date(dateString);
              return date >= fourteenDaysAgo && date < sevenDaysAgo;
            },
          ).length;

          if (usersPrevious7Days > 0) {
            const percentChange =
              ((usersLast7Days - usersPrevious7Days) / usersPrevious7Days) *
              100;
            setTrendingPercentage(Math.round(percentChange * 10) / 10);
            setIsTrendingUp(percentChange >= 0);
          } else if (usersLast7Days > 0) {
            setTrendingPercentage(100);
            setIsTrendingUp(true);
          } else {
            setTrendingPercentage(0);
            setIsTrendingUp(true);
          }
        }

        const levelStatsObj = levelData || {};
        setLevelStats(levelStatsObj);

        const roleStatsObj = roleData || {};
        setRoleStats(roleStatsObj);

        if (serviceData) {
          const servicesArray = Object.entries(serviceData)
            .filter(([key]) => key !== "timestamp")
            .map(([serviceName, serviceInfo]) => ({
              service_name: serviceName,
              status: (serviceInfo as { status: string }).status,
              last_check:
                (serviceInfo as { last_check?: string }).last_check || "",
              response_time_ms:
                (serviceInfo as { response_time_ms?: number })
                  .response_time_ms || 0,
            }));
          setServiceStatus(servicesArray);
        }

        if (metricsData) {
          setSystemMetrics(metricsData);
        }
      } catch (error) {
        console.error("Error loading dashboard data:", error);
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [
    fetchUserCount,
    fetchLevelStats,
    fetchRoleStats,
    fetchServiceStatus,
    fetchSystemMetrics,
    fetchCreationDates,
  ]);

  const totalLevelUsers = Object.values(levelStats).reduce(
    (sum, count) => sum + count,
    0,
  );
  const totalRoleUsers = Object.values(roleStats).reduce(
    (sum, count) => sum + count,
    0,
  );

  if (loading) {
    return <SectionCardsSkeleton />;
  }

  return (
    <div className="dark:*:data-[slot=card]:bg-card grid grid-cols-1 gap-4 px-4 *:data-[slot=card]:shadow-xs lg:px-6 @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
      <Card
        className="@container/card cursor-pointer hover:shadow-md transition-shadow"
        onClick={() => navigate("/users")}
      >
        <CardHeader>
          <CardDescription>Total Users</CardDescription>
          <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
            {userCount.toLocaleString()}
          </CardTitle>
          <CardAction>
            <Badge
              variant="outline"
              className={isTrendingUp ? "text-green-600" : "text-red-600"}
            >
              {isTrendingUp ? <IconTrendingUp /> : <IconTrendingDown />}
              {isTrendingUp ? "+" : ""}
              {trendingPercentage}%
            </Badge>
          </CardAction>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-1.5 text-sm">
          <div className="line-clamp-1 flex gap-2 font-medium">
            {isTrendingUp ? "Trending up" : "Trending down"} this week{" "}
            {isTrendingUp ? (
              <IconTrendingUp className="size-4" />
            ) : (
              <IconTrendingDown className="size-4" />
            )}
          </div>
          <div className="text-muted-foreground">Compared to last week</div>
          <div className="w-full space-y-2">
            {Object.entries(roleStats)
              .sort(([, a], [, b]) => b - a)
              .map(([role, count]) => {
                const percentage =
                  totalRoleUsers > 0 ? (count / totalRoleUsers) * 100 : 0;
                return (
                  <div key={role} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      {getRoleIcon(role)}
                      <span className="font-medium capitalize">{role}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <div className="w-16 h-2 bg-gray-200 rounded-full overflow-hidden">
                        <div
                          className="h-full bg-gray-500 rounded-full transition-all duration-300"
                          style={{ width: `${percentage}%` }}
                        />
                      </div>
                      <Badge className={`text-xs w-9 ${getRoleColor(role)}`}>
                        {count}
                      </Badge>
                    </div>
                  </div>
                );
              })}
          </div>
        </CardFooter>
      </Card>

      <Card
        className="@container/card cursor-pointer hover:shadow-md transition-shadow"
        onClick={() => navigate("/users")}
      >
        <CardHeader>
          <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl ">
            User Level
          </CardTitle>
          <CardDescription className="flex items-center gap-2">
            <IconUsers className="size-4 text-muted-foreground" />
            Users by Level
          </CardDescription>
          <CardAction>
            <Badge variant="outline">
              {Object.keys(levelStats).length} levels
            </Badge>
          </CardAction>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-2 text-sm">
          <div className="w-full space-y-2">
            {Object.entries(levelStats)
              .sort(([, a], [, b]) => b - a)
              .map(([level, count]) => {
                const percentage =
                  totalLevelUsers > 0 ? (count / totalLevelUsers) * 100 : 0;
                return (
                  <div
                    key={level}
                    className="flex items-center justify-between"
                  >
                    <div className="flex items-center gap-2">
                      {getLevelIcon(level)}
                      <span className="font-medium capitalize">{level}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <div className="w-16 h-2 bg-gray-200 rounded-full overflow-hidden">
                        <div
                          className="h-full bg-gray-500 rounded-full transition-all duration-300"
                          style={{ width: `${percentage}%` }}
                        />
                      </div>
                      <Badge className={`text-xs w-9 ${getLevelColor(level)}`}>
                        {count}
                      </Badge>
                    </div>
                  </div>
                );
              })}
          </div>
        </CardFooter>
      </Card>

      <Card
        className="@container/card cursor-pointer hover:shadow-md transition-shadow"
        onClick={() => navigate("/logs")}
      >
        <CardHeader>
          <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
            System Services
          </CardTitle>
          <CardDescription className="flex items-center gap-2">
            <IconServer className="size-4 text-muted-foreground" />
            Service Health Status
          </CardDescription>
          <CardAction>
            <Badge variant="outline">{serviceStatus.length} services</Badge>
          </CardAction>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-2 text-sm">
          <div className="w-full space-y-2">
            {serviceStatus.map((service, index) => (
              <button
                type="button"
                key={service.service_name || `service-${index}`}
                className="w-full flex items-center justify-between hover:bg-muted rounded-md px-2 py-1 transition-colors cursor-pointer text-left"
                onClick={(e) => {
                  e.stopPropagation();
                  navigate(`/logs?service=${service.service_name}`);
                }}
              >
                <div className="flex items-center gap-2">
                  {getServiceIcon(service.service_name)}
                  <span className="font-medium capitalize">
                    {service.service_name}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <Badge
                    className={`text-xs min-w-[90px] flex justify-center ${getServiceStatusColor(service.status)}`}
                  >
                    {getServiceStatusIcon(service.status)}
                    <span className="ml-1">{service.status}</span>
                  </Badge>
                  <span className="text-xs text-muted-foreground min-w-[30px] text-right">
                    {service.response_time_ms}ms
                  </span>
                </div>
              </button>
            ))}
          </div>
        </CardFooter>
      </Card>

      <Card
        className="@container/card cursor-pointer hover:shadow-md transition-shadow"
        onClick={() => navigate("/logs")}
      >
        <CardHeader>
          <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
            API Metrics
          </CardTitle>
          <CardDescription className="flex items-center gap-2">
            <IconActivity className="size-4 text-muted-foreground" />
            Request Performance (24h)
          </CardDescription>
          <CardAction>
            <Badge variant="outline">
              <IconClock className="size-3 mr-1" />
              Live
            </Badge>
          </CardAction>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-2 text-sm">
          {systemMetrics && (
            <div className="w-full space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <span className="font-medium">Total Requests</span>
                </div>
                <Badge variant="outline" className="font-mono">
                  {systemMetrics.total_api_requests?.toLocaleString() || 0}
                </Badge>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <span className="font-medium">Error Rate</span>
                </div>
                <Badge
                  className={`text-xs ${
                    (systemMetrics.error_rate || 0) > 5
                      ? "bg-red-100 text-red-800 border-red-200"
                      : (systemMetrics.error_rate || 0) > 1
                        ? "bg-yellow-100 text-yellow-800 border-yellow-200"
                        : "bg-green-100 text-green-800 border-green-200"
                  }`}
                >
                  {systemMetrics.error_rate?.toFixed(1) || 0}%
                </Badge>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <span className="font-medium">Avg Response Time</span>
                </div>
                <Badge variant="outline" className="font-mono">
                  {systemMetrics.avg_response_time_ms?.toFixed(1) || 0}ms
                </Badge>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <span className="font-medium">Active Sessions</span>
                </div>
                <Badge variant="outline" className="font-mono">
                  {systemMetrics.active_sessions || 0}
                </Badge>
              </div>
            </div>
          )}
        </CardFooter>
      </Card>
    </div>
  );
}
