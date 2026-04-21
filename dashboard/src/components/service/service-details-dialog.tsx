"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { useIsMobile } from "@/hooks/useMobile";
import type {
  DatabaseDetails,
  MinioDetails,
  RedisDetails,
  ServiceStatus,
} from "@/lib/types/stats.types";
import {
  IconAlertCircle,
  IconCircleCheck,
  IconCircleX,
  IconRefresh,
} from "@tabler/icons-react";
import * as React from "react";

interface ServiceDetailsDialogProps {
  service: ServiceStatus | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onRefresh?: () => Promise<void>;
}

export function ServiceDetailsDialog({
  service,
  open,
  onOpenChange,
  onRefresh,
}: ServiceDetailsDialogProps) {
  const isMobile = useIsMobile();
  const [refreshing, setRefreshing] = React.useState(false);

  if (!service) return null;

  const handleRefresh = async () => {
    if (onRefresh) {
      setRefreshing(true);
      try {
        await onRefresh();
      } finally {
        setRefreshing(false);
      }
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  };

  const formatUptime = (seconds: string) => {
    const sec = Number.parseInt(seconds, 10);
    const days = Math.floor(sec / 86400);
    const hours = Math.floor((sec % 86400) / 3600);
    const minutes = Math.floor((sec % 3600) / 60);

    if (days > 0) return `${days}d ${hours}h ${minutes}m`;
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  };

  const formatBytes = (bytes: string | number) => {
    const b = typeof bytes === "string" ? Number.parseInt(bytes, 10) : bytes;
    if (b === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(b) / Math.log(k));
    return `${Number.parseFloat((b / k ** i).toFixed(2))} ${sizes[i]}`;
  };

  const getStatusBadge = () => {
    switch (service.status) {
      case "healthy":
        return (
          <Badge variant="default" className="bg-green-500">
            <IconCircleCheck className="w-3 h-3 mr-1" />
            Healthy
          </Badge>
        );
      case "unhealthy":
        return (
          <Badge variant="destructive">
            <IconCircleX className="w-3 h-3 mr-1" />
            Unhealthy
          </Badge>
        );
      default:
        return (
          <Badge variant="secondary">
            <IconAlertCircle className="w-3 h-3 mr-1" />
            Unknown
          </Badge>
        );
    }
  };

  const renderDatabaseDetails = (details: DatabaseDetails) => (
    <div className="space-y-4">
      <h4 className="font-medium flex items-center">Connections</h4>
      <div className="grid grid-cols-2 gap-4 pl-6">
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">
            Open Connections
          </Label>
          <div className="text-sm font-medium">
            {details.open_connections}
          </div>
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">
            Idle Connections
          </Label>
          <div className="text-sm font-medium">
            {details.idle_connections}
          </div>
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">
            In Use Connections
          </Label>
          <div className="text-sm font-medium">
            {details.in_use_connections}
          </div>
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">
            Total Connections
          </Label>
          <div className="text-sm font-medium">
            {details.total_connections}
          </div>
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">
            Max Open Connections
          </Label>
          <div className="text-sm font-medium">
            {details.max_open_connections}
          </div>
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">
            Wait Count
          </Label>
          <div className="text-sm font-medium">{details.wait_count}</div>
        </div>
      </div>

      <Separator />

      <h4 className="font-medium flex items-center">Server Info</h4>
      <div className="pl-6">
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">Wait Duration</Label>
          <div className="text-sm font-medium">{details.wait_duration}</div>
        </div>
      </div>
    </div>
  );

  const renderRedisDetails = (details: RedisDetails) => (
    <div className="space-y-4">
      <h4 className="font-medium flex items-center">Connections</h4>
      <div className="grid grid-cols-2 gap-4 pl-6">
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">
            Connected Clients
          </Label>
          <div className="text-sm font-medium">{details.connected_clients}</div>
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">
            Total Connections
          </Label>
          <div className="text-sm font-medium">{details.total_connections}</div>
        </div>
      </div>

      <Separator />

      <h4 className="font-medium flex items-center">Cache Performance</h4>
      <div className="grid grid-cols-2 gap-4 pl-6">
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">
            Cache Hit Rate
          </Label>
          <div className="text-sm font-medium">
            {details.cache_hit_rate != null
              ? `${details.cache_hit_rate.toFixed(2)}%`
              : "N/A"}
          </div>
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">Keyspace Hits</Label>
          <div className="text-sm font-medium">{details.keyspace_hits}</div>
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">
            Keyspace Misses
          </Label>
          <div className="text-sm font-medium">{details.keyspace_misses}</div>
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">
            Total Commands
          </Label>
          <div className="text-sm font-medium">{details.total_commands}</div>
        </div>
      </div>

      <Separator />

      <h4 className="font-medium flex items-center">Memory & Keys</h4>
      <div className="grid grid-cols-2 gap-4 pl-6">
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">Used Memory</Label>
          <div className="text-sm font-medium">
            {formatBytes(details.used_memory_bytes)}
          </div>
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">Expired Keys</Label>
          <div className="text-sm font-medium">{details.expired_keys}</div>
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">Evicted Keys</Label>
          <div className="text-sm font-medium">{details.evicted_keys}</div>
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">Uptime</Label>
          <div className="text-sm font-medium">
            {formatUptime(details.uptime_seconds)}
          </div>
        </div>
      </div>

      <Separator />

      <h4 className="font-medium flex items-center">Server Info</h4>
      <div className="pl-6">
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">Version</Label>
          <div className="text-sm font-medium">Redis {details.version}</div>
        </div>
      </div>
    </div>
  );

  const renderMinioDetails = (details: MinioDetails) => (
    <div className="space-y-4">
      <h4 className="font-medium flex items-center">Bucket Information</h4>
      <div className="grid grid-cols-2 gap-4 pl-6">
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">Bucket Name</Label>
          <div className="text-sm font-medium">{details.bucket_name}</div>
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">Location</Label>
          <div className="text-sm font-medium">{details.location}</div>
        </div>
      </div>

      <Separator />

      <h4 className="font-medium flex items-center">Storage</h4>
      <div className="grid grid-cols-2 gap-4 pl-6">
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">Total Objects</Label>
          <div className="text-sm font-medium">{details.total_objects}</div>
        </div>
        <div className="flex flex-col gap-1">
          <Label className="text-xs text-muted-foreground">Total Size</Label>
          <div className="text-sm font-medium">{details.total_size_human}</div>
        </div>
      </div>
    </div>
  );

  const renderDetails = () => {
    if (!service.details) {
      return (
        <div className="text-center text-muted-foreground py-4">
          No detailed information available
        </div>
      );
    }

    switch (service.service_name) {
      case "postgres":
        return renderDatabaseDetails(service.details as DatabaseDetails);
      case "redis":
        return renderRedisDetails(service.details as RedisDetails);
      case "minio":
        return renderMinioDetails(service.details as MinioDetails);
      default:
        return (
          <div className="space-y-2 pl-6">
            {Object.entries(service.details).map(([key, value]) => (
              <div key={key} className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground capitalize">
                  {key.replace(/_/g, " ")}
                </Label>
                <div className="text-sm font-medium">{String(value)}</div>
              </div>
            ))}
          </div>
        );
    }
  };

  return (
    <Drawer
      direction={isMobile ? "bottom" : "right"}
      open={open}
      onOpenChange={onOpenChange}
    >
      <DrawerContent className="max-w-lg">
        <DrawerHeader className="gap-1">
          <div className="flex items-start justify-between gap-4">
            <div className="flex flex-col gap-1 flex-1">
              <DrawerTitle className="flex items-center gap-2">
                <span className="capitalize">{service.service_name}</span>{" "}
                Details
              </DrawerTitle>
              <DrawerDescription>
                Detailed health information and metrics
              </DrawerDescription>
            </div>
            {onRefresh && (
              <Button
                variant="outline"
                onClick={handleRefresh}
                disabled={refreshing}
                size="icon"
              >
                <IconRefresh
                  className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`}
                />
              </Button>
            )}
          </div>
        </DrawerHeader>

        <div className="flex flex-col gap-6 overflow-y-auto px-4 pb-4">
          <div className="flex items-center justify-between p-4 bg-muted/50 rounded-lg">
            <div className="flex items-center gap-3">
              <div>
                <p className="font-medium capitalize">{service.service_name}</p>
                <p className="text-sm text-muted-foreground">
                  {service.response_time_ms}ms response time
                </p>
              </div>
            </div>
            {getStatusBadge()}
          </div>

          {service.error_msg && (
            <div className="p-3 bg-red-100 dark:bg-red-900/20 text-red-800 dark:text-red-200 rounded-lg">
              <p className="text-sm font-medium">Error</p>
              <p className="text-sm">{service.error_msg}</p>
            </div>
          )}

          <div className="space-y-2">
            <h4 className="font-medium flex items-center">Last Health Check</h4>
            <div className="pl-6 text-sm text-muted-foreground">
              {service.last_check ? formatDate(service.last_check) : "N/A"}
            </div>
          </div>

          <Separator />

          {renderDetails()}
        </div>

        <DrawerFooter className="flex flex-row gap-2">
          <DrawerClose asChild>
            <Button variant="outline" className="flex-1">
              Close
            </Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  );
}
