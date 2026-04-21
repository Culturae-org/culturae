"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { AVATAR_ENDPOINTS } from "@/lib/api/endpoints";
import type { AdminUser } from "@/lib/types/user.types";
import { IconCheck, IconCopy, IconUser } from "@tabler/icons-react";

interface UserProfileCardProps {
  user: AdminUser;
  copiedField: string | null;
  onCopy: (text: string, field: string) => void;
}

export function UserProfileCard({
  user,
  copiedField,
  onCopy,
}: UserProfileCardProps) {
  const formatDate = (dateString: string) =>
    new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });

  const getRoleBadgeVariant = (role: string) =>
    role === "administrator" ? "default" : "secondary";

  const getAccountStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      active: "text-green-600 dark:text-green-400",
      suspended: "text-yellow-600 dark:text-yellow-400",
      banned: "text-red-600 dark:text-red-400",
      inactive: "text-gray-600 dark:text-gray-400",
    };
    return colors[status] ?? colors.active;
  };

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <IconUser className="h-5 w-5" />
          Profile Information
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-4">
          <Avatar className="w-16 h-16">
            <AvatarImage
              src={
                user.has_avatar && user.id
                  ? AVATAR_ENDPOINTS.GET(user.id)
                  : undefined
              }
              alt={`${user.username}'s avatar`}
            />
            <AvatarFallback className="bg-primary/10">
              <IconUser className="h-8 w-8 text-primary" />
            </AvatarFallback>
          </Avatar>
          <div>
            <h3 className="font-semibold text-lg">{user.username}</h3>
            <div className="flex items-center gap-2 mt-1">
              <div
                className={`h-2.5 w-2.5 rounded-full ${user.is_online ? "bg-green-500" : "bg-gray-400"}`}
              />
              <span className="text-sm text-muted-foreground">
                {user.is_online ? "Online" : "Offline"}
              </span>
            </div>
          </div>
        </div>

        <Separator />

        <div className="grid grid-cols-2 gap-3 text-sm">
          <div>
            <p className="text-muted-foreground text-xs mb-1">Email</p>
            <p className="font-medium truncate">{user.email}</p>
          </div>
          <div>
            <p className="text-muted-foreground text-xs mb-1">Language</p>
            <p className="font-medium capitalize">{user.language}</p>
          </div>
          <div>
            <p className="text-muted-foreground text-xs mb-1">Role</p>
            <Badge variant={getRoleBadgeVariant(user.role)} className="w-fit">
              {user.role === "administrator" ? "Administrator" : "User"}
            </Badge>
          </div>
          <div>
            <p className="text-muted-foreground text-xs mb-1">Account Status</p>
            <p
              className={`font-medium capitalize ${getAccountStatusColor(user.account_status)}`}
            >
              {user.account_status.replace("_", " ")}
            </p>
          </div>
          <div>
            <p className="text-muted-foreground text-xs mb-1">Created</p>
            <p className="font-medium">{formatDate(user.created_at)}</p>
          </div>
          <div>
            <p className="text-muted-foreground text-xs mb-1">Updated</p>
            <p className="font-medium">{formatDate(user.updated_at)}</p>
          </div>
        </div>

        <Separator />

        <div className="space-y-1">
          <p className="text-muted-foreground text-xs mb-2">IDs</p>
          {[
            { label: "User ID", value: user.id, field: "userId" },
            { label: "Public ID", value: user.public_id, field: "publicId" },
          ].map(({ label, value, field }) => (
            <div key={field} className="flex items-center gap-2">
              <span className="text-xs text-muted-foreground w-20 shrink-0">
                {label}
              </span>
              <code className="text-xs bg-muted px-2 py-1 rounded flex-1 truncate">
                {value}
              </code>
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 shrink-0"
                onClick={() => onCopy(value, field)}
              >
                {copiedField === field ? (
                  <IconCheck className="h-3.5 w-3.5 text-green-600" />
                ) : (
                  <IconCopy className="h-3.5 w-3.5" />
                )}
              </Button>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
