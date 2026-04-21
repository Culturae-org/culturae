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
} from "@/components/ui/alert-dialog";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { apiPost } from "@/lib/api-client";
import { SETTINGS_ENDPOINTS } from "@/lib/api/endpoints";
import * as React from "react";
import { toast } from "sonner";
import { InfoHover } from "./info-hover";

interface MaintenanceModeCardProps {
  maintenanceMode: boolean;
  loading: boolean;
  initialLoading: boolean;
  onMaintenanceModeChange: (value: boolean) => void;
  onLoadingChange: (loading: boolean) => void;
}

export function MaintenanceModeCard({
  maintenanceMode,
  loading,
  initialLoading,
  onMaintenanceModeChange,
  onLoadingChange,
}: MaintenanceModeCardProps) {
  const [confirmOpen, setConfirmOpen] = React.useState(false);

  const confirmToggle = async () => {
    const newValue = !maintenanceMode;
    setConfirmOpen(false);
    onLoadingChange(true);
    try {
      const res = await apiPost(SETTINGS_ENDPOINTS.MAINTENANCE, {
        enabled: newValue,
      });
      if (!res.ok) throw new Error("Failed");
      onMaintenanceModeChange(newValue);
      toast.success(
        newValue ? "Maintenance mode enabled" : "Maintenance mode disabled",
      );
    } catch {
      toast.error("Failed to update maintenance mode");
    } finally {
      onLoadingChange(false);
    }
  };

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          Maintenance Mode
        </CardTitle>
        <CardDescription>
          When enabled, the API will return 503 for non-admin requests. Use
          during deployments or database migrations.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Label>Enable maintenance mode</Label>
              <InfoHover description="When enabled, the API returns 503 for all non-admin requests. Use during deployments, migrations, or emergency situations." />
            </div>
            <p className="text-sm text-muted-foreground">
              {maintenanceMode
                ? "Maintenance mode is active. Only admins can access the API."
                : "The API is operating normally."}
            </p>
          </div>
          <Switch
            checked={maintenanceMode}
            onCheckedChange={() => setConfirmOpen(true)}
            disabled={loading || initialLoading}
          />
        </div>
        <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>
                {maintenanceMode
                  ? "Disable maintenance mode?"
                  : "Enable maintenance mode?"}
              </AlertDialogTitle>
              <AlertDialogDescription>
                {maintenanceMode
                  ? "The API will resume normal operation. All users will be able to access the platform again."
                  : "The API will return 503 for all non-admin requests. Users will not be able to access the platform until maintenance mode is disabled."}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction
                onClick={confirmToggle}
                className={
                  maintenanceMode
                    ? ""
                    : "bg-destructive text-white hover:bg-destructive/90"
                }
              >
                {maintenanceMode ? "Disable" : "Enable"}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </CardContent>
    </Card>
  );
}
