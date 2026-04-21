"use client";

import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { ServiceStatus } from "@/lib/types/stats.types";
import {
  IconAlertCircle,
  IconBucket,
  IconCircleCheck,
  IconCircleX,
  IconDatabase,
  IconServer,
} from "@tabler/icons-react";

interface ServiceHealthCardProps {
  services: ServiceStatus[];
  onServiceClick: (service: ServiceStatus) => void;
}

export function ServiceHealthCard({
  services,
  onServiceClick,
}: ServiceHealthCardProps) {
  const getStatusBadge = (status: string) => {
    switch (status) {
      case "healthy":
        return (
          <Badge variant="default" className="bg-green-500">
            <IconCircleCheck className="h-3 w-3 mr-1" />
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

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle>Service Health</CardTitle>
        <CardDescription>
          Current status of system services - Click on a service for details
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {services.map((service, index) => (
            <button
              type="button"
              key={service.service_name || `service-${index}`}
              className="w-full flex items-center justify-between p-4 border rounded-lg cursor-pointer hover:bg-muted/50 transition-colors text-left"
              onClick={() => onServiceClick(service)}
            >
              <div className="flex items-center space-x-3">
                {service.service_name === "postgres" && (
                  <IconDatabase className="h-5 w-5" />
                )}
                {service.service_name === "redis" && (
                  <IconServer className="h-5 w-5" />
                )}
                {service.service_name === "minio" && (
                  <IconBucket className="h-5 w-5" />
                )}
                <div>
                  <p className="font-medium capitalize">
                    {service.service_name}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    {service.response_time_ms}ms response time
                  </p>
                </div>
              </div>
              <div className="flex items-center space-x-2">
                {getStatusBadge(service.status)}
                {service.error_msg && (
                  <span className="text-sm text-red-500">
                    {service.error_msg}
                  </span>
                )}
              </div>
            </button>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
