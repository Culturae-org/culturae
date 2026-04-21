"use client";

import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";

interface PlatformInfoCardProps {
  version: string | null;
  buildTime: string | null;
  environment: string | null;
}

export function PlatformInfoCard({
  version,
  buildTime,
  environment,
}: PlatformInfoCardProps) {
  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          Platform Information
        </CardTitle>
        <CardDescription>
          Current backend version and environment. A notification is displayed
          automatically when a new version is deployed.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {version ? (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="space-y-1">
              <Label className="text-muted-foreground">Environment</Label>
              <div className="flex items-center gap-2">
                <Badge
                  variant={
                    environment === "development" || environment === "dev"
                      ? "secondary"
                      : "default"
                  }
                >
                  {environment === "development" || environment === "dev"
                    ? "Development"
                    : "Production"}
                </Badge>
              </div>
            </div>
            <div className="space-y-1">
              <Label className="text-muted-foreground">Version</Label>
              <p className="text-sm font-mono">{version}</p>
            </div>
            <div className="space-y-1">
              <Label className="text-muted-foreground">Build Time</Label>
              <p className="text-sm">
                {buildTime && buildTime !== "unknown"
                  ? new Date(buildTime).toLocaleString()
                  : "Unknown"}
              </p>
            </div>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            Loading platform information...
          </p>
        )}
      </CardContent>
    </Card>
  );
}
