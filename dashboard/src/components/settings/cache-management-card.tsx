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
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { apiPost } from "@/lib/api-client";
import { SETTINGS_ENDPOINTS } from "@/lib/api/endpoints";
import { IconRefresh, IconTrash } from "@tabler/icons-react";
import { toast } from "sonner";

interface CacheManagementCardProps {
  loading: boolean;
  onLoadingChange: (loading: boolean) => void;
}

export function CacheManagementCard({
  loading,
  onLoadingChange,
}: CacheManagementCardProps) {
  const clearCache = async () => {
    onLoadingChange(true);
    try {
      const res = await apiPost(SETTINGS_ENDPOINTS.CACHE_CLEAR);
      if (!res.ok) throw new Error("Failed");
      toast.success("Cache cleared successfully");
    } catch {
      toast.error("Failed to clear cache");
    } finally {
      onLoadingChange(false);
    }
  };

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          Cache Management
        </CardTitle>
        <CardDescription>
          Clear the Redis cache. This will flush all cached data including
          sessions, game state, and temporary data.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button variant="destructive" disabled={loading}>
              {loading ? (
                <IconRefresh className="h-4 w-4 mr-2 animate-spin" />
              ) : (
                <IconTrash className="h-4 w-4 mr-2" />
              )}
              Clear All Cache
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Clear all cached data?</AlertDialogTitle>
              <AlertDialogDescription>
                This will flush all cached data including sessions, game state,
                and temporary data. This may temporarily impact performance.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction
                onClick={clearCache}
                className="bg-destructive text-white hover:bg-destructive/90"
              >
                Clear Cache
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </CardContent>
    </Card>
  );
}
