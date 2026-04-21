"use client";

import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { UserActionLogsDataTable } from "@/components/users/user-action-logs-data-table";
import { UserActiveSessionsDataTable } from "@/components/users/user-active-sessions-data-table";
import { UserConnectionLogsDataTable } from "@/components/users/user-connection-logs-data-table";
import { UserDeleteDialog } from "@/components/users/user-delete-dialog";
import { UserEditDialog } from "@/components/users/user-edit-dialog";
import { UserFriendRequestsDataTable } from "@/components/users/user-friend-requests-data-table";
import { UserFriendsDataTable } from "@/components/users/user-friends-data-table";
import { UserGamesDataTable } from "@/components/users/user-games-data-table";
import { UserPrivacyCard } from "@/components/users/user-privacy-card";
import { UserProfileCard } from "@/components/users/user-profile-card";
import { UserProgressionCard } from "@/components/users/user-progression-card";
import { UserStatsCards } from "@/components/users/user-stats-cards";
import { apiGet } from "@/lib/api-client";
import { USERS_ENDPOINTS } from "@/lib/api/endpoints";
import { useUser } from "@/lib/stores";
import type { AdminUser } from "@/lib/types/user.types";
import { IconRefresh } from "@tabler/icons-react";
import * as React from "react";
import { useNavigate, useParams } from "react-router";

export default function UserDetailPage() {
  const params = useParams();
  const navigate = useNavigate();
  const userId = params.id as string;
  const { userId: currentUserId, refetchProfile } = useUser();

  const [user, setUser] = React.useState<AdminUser | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [copiedField, setCopiedField] = React.useState<string | null>(null);

  const fetchUser = React.useCallback(async () => {
    try {
      setLoading(true);
      const response = await apiGet(USERS_ENDPOINTS.GET(userId));
      if (!response.ok) throw new Error("Failed to fetch user details");
      const data = await response.json();
      setUser(data.data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setLoading(false);
    }
  }, [userId]);

  React.useEffect(() => {
    fetchUser();
  }, [fetchUser]);

  const copyToClipboard = async (text: string, field: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedField(field);
      setTimeout(() => setCopiedField(null), 2000);
    } catch {
      /* ignore */
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
        <div className="px-4 lg:px-6 space-y-2">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-4 w-64" />
        </div>
        <div className="px-4 lg:px-6 grid grid-cols-2 md:grid-cols-4 gap-4">
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
        </div>
        <div className="px-4 lg:px-6">
          <Skeleton className="h-64 w-full" />
        </div>
      </div>
    );
  }

  if (error || !user) {
    return (
      <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
        <div className="px-4 lg:px-6 text-center py-8">
          <p className="text-muted-foreground">Failed to load user details</p>
          <Button onClick={fetchUser} className="mt-4">
            <IconRefresh className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">User Details</h1>
            <p className="text-muted-foreground">
              Detailed information for {user.username}
            </p>
          </div>
          <div className="flex gap-2">
            <UserEditDialog
              user={user}
              onUserUpdated={() => {
                fetchUser();
                if (currentUserId === userId) refetchProfile();
              }}
            />
            <UserDeleteDialog
              user={user}
              onUserDeleted={() => navigate("/users")}
            />
          </div>
        </div>
      </div>



      <div className="px-4 lg:px-6">
        <Tabs defaultValue="overview" className="w-full">
          <TabsList className="grid w-full grid-cols-4">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="games">Games</TabsTrigger>
            <TabsTrigger value="activity">Activity</TabsTrigger>
            <TabsTrigger value="friends">Friends</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <UserProfileCard
                user={user}
                copiedField={copiedField}
                onCopy={copyToClipboard}
              />
              <UserProgressionCard user={user} />
            </div>

            <div className="px-4 lg:px-6">
              <UserStatsCards user={user} />
            </div>

            <UserPrivacyCard user={user} />
          </TabsContent>

          <TabsContent value="games" className="space-y-6">
            <UserGamesDataTable userId={userId} />
          </TabsContent>

          <TabsContent value="activity" className="space-y-6">
            <Tabs defaultValue="logs" className="w-full">
              <TabsList>
                <TabsTrigger value="logs">Connection Logs</TabsTrigger>
                <TabsTrigger value="sessions">Active Sessions</TabsTrigger>
                <TabsTrigger value="actions">Profile Actions</TabsTrigger>
              </TabsList>

              <TabsContent value="logs" className="space-y-4">
                <UserConnectionLogsDataTable userId={userId} />
              </TabsContent>

              <TabsContent value="sessions" className="space-y-4">
                <UserActiveSessionsDataTable userId={userId} />
              </TabsContent>

              <TabsContent value="actions" className="space-y-4">
                <UserActionLogsDataTable userId={userId} />
              </TabsContent>
            </Tabs>
          </TabsContent>

          <TabsContent value="friends" className="space-y-6">
            <Tabs defaultValue="friends-list" className="w-full">
              <TabsList>
                <TabsTrigger value="friends-list">Friends</TabsTrigger>
                <TabsTrigger value="requests">Requests</TabsTrigger>
              </TabsList>

              <TabsContent value="friends-list" className="space-y-4">
                <UserFriendsDataTable userId={userId} />
              </TabsContent>

              <TabsContent value="requests" className="space-y-4">
                <UserFriendRequestsDataTable userId={userId} />
              </TabsContent>
            </Tabs>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
