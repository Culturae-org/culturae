"use client";

import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { apiGet } from "@/lib/api-client";
import { LOGS_ENDPOINTS } from "@/lib/api/endpoints";
import type { AdminActionLog } from "@/lib/types/logs.types";
import { IconShield } from "@tabler/icons-react";
import * as React from "react";
import { Link } from "react-router";

function timeAgo(date: string) {
  const s = Math.floor((Date.now() - new Date(date).getTime()) / 1000);
  if (s < 60) return `${s}s ago`;
  if (s < 3600) return `${Math.floor(s / 60)}m ago`;
  if (s < 86400) return `${Math.floor(s / 3600)}h ago`;
  return `${Math.floor(s / 86400)}d ago`;
}

export function RecentAdminActions() {
  const [actions, setActions] = React.useState<AdminActionLog[]>([]);
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    async function load() {
      try {
        const res = await apiGet(
          `${LOGS_ENDPOINTS.ADMIN_ACTIONS}?page=1&limit=5`,
        );
        if (res.ok) {
          const data = await res.json();
          const actionsArray = Array.isArray(data.data)
            ? data.data
            : Array.isArray(data.logs)
              ? data.logs
              : [];
          setActions(actionsArray.slice(0, 5));
        }
      } catch {
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  return (
    <Card className="border-0 dark:border">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2 text-base">
              <IconShield className="h-4 w-4" />
              Recent Admin Actions
            </CardTitle>
            <CardDescription>Last 5 administrative actions</CardDescription>
          </div>
          <Link
            to="/logs"
            className="text-xs text-muted-foreground hover:text-foreground transition-colors"
          >
            View all
          </Link>
        </div>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="space-y-3">
            {[0, 1, 2, 3, 4].map((i) => (
              <Skeleton key={`action-skel-${i}`} className="h-10 w-full" />
            ))}
          </div>
        ) : actions.length === 0 ? (
          <p className="text-sm text-muted-foreground text-center py-4">
            No recent actions
          </p>
        ) : (
          <div className="space-y-3">
            {actions.map((a) => (
              <div
                key={a.ID}
                className="flex items-center justify-between gap-3"
              >
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium truncate">
                      {a.AdminName}
                    </span>
                    <Badge variant="outline" className="text-xs shrink-0">
                      {a.Action.replace(/_/g, " ")}
                    </Badge>
                  </div>
                  <div className="text-xs text-muted-foreground capitalize">
                    {a.Resource}
                  </div>
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  <Badge
                    variant={a.IsSuccess ? "default" : "destructive"}
                    className={`text-xs ${a.IsSuccess ? "bg-green-500" : ""}`}
                  >
                    {a.IsSuccess ? "OK" : "Fail"}
                  </Badge>
                  <span className="text-xs text-muted-foreground w-14 text-right">
                    {timeAgo(a.CreatedAt)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
