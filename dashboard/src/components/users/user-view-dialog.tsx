"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { useIsMobile } from "@/hooks/useMobile";
import { AVATAR_ENDPOINTS } from "@/lib/api/endpoints";
import type { AdminUser } from "@/lib/types/user.types";
import {
  IconCheck,
  IconCopy,
  IconEdit,
  IconEye,
  IconUser,
} from "@tabler/icons-react";
import * as React from "react";
import { useLocation, useNavigate } from "react-router";

interface UserViewDialogProps {
  user: AdminUser;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  onUserUpdated?: () => void;
  onEditClick?: () => void;
}

export function UserViewDialog({
  user,
  open: controlledOpen,
  onOpenChange,
  onEditClick,
}: UserViewDialogProps) {
  const isMobile = useIsMobile();
  const [internalOpen, setInternalOpen] = React.useState(false);
  const isOpen = controlledOpen ?? internalOpen;
  const setIsOpen = onOpenChange ?? setInternalOpen;
  const navigate = useNavigate();
  const location = useLocation();
  const [copiedField, setCopiedField] = React.useState<string | null>(null);

  const avatarUrl = user.has_avatar ? AVATAR_ENDPOINTS.GET(user.id) : undefined;

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getRoleBadgeVariant = (role: string) => {
    return role === "administrator" ? "default" : "secondary";
  };

  const getLevelColor = (rank: string) => {
    const colors: Record<string, string> = {
      beginner:
        "text-green-600 bg-green-100 dark:text-green-400 dark:bg-green-900",
      intermediate:
        "text-yellow-600 bg-yellow-100 dark:text-yellow-400 dark:bg-yellow-900",
      pro: "text-orange-600 bg-orange-100 dark:text-orange-400 dark:bg-orange-900",
      expert:
        "text-purple-600 bg-purple-100 dark:text-purple-400 dark:bg-purple-900",
      legend:
        "text-amber-600 bg-amber-100 dark:text-amber-400 dark:bg-amber-900",
    };
    return colors[rank?.toLowerCase() || ""] || colors.beginner;
  };

  const getAccountStatusColor = (status: string) => {
    const colors = {
      active: "text-green-600 dark:text-green-400",
      suspended: "text-yellow-600 dark:text-yellow-400",
      banned: "text-red-600 dark:text-red-400",
      inactive: "text-gray-600 dark:text-gray-400",
    };
    return colors[status as keyof typeof colors] || colors.active;
  };

  const handleViewDetails = () => {
    setIsOpen(false);
    navigate(`/users/${user.id}`);
  };

  const handleClose = () => {
    setIsOpen(false);
    navigate(location.pathname, { replace: true });
  };

  const copyToClipboard = async (text: string, field: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedField(field);
      setTimeout(() => setCopiedField(null), 2000);
    } catch (err) {
      console.error("Failed to copy text: ", err);
    }
  };

  return (
    <Drawer
        direction={isMobile ? "bottom" : "right"}
        open={isOpen}
        onOpenChange={(open) => {
          if (!open) {
            handleClose();
          } else {
            setIsOpen(true);
          }
        }}
      >
        {controlledOpen === undefined && (
          <DrawerTrigger asChild>
            <Button variant="ghost" size="sm" onClick={() => setIsOpen(true)}>
              <IconEye className="h-4 w-4" />
            </Button>
          </DrawerTrigger>
        )}
        <DrawerContent>
          <DrawerHeader className="gap-1">
            <div className="flex items-center justify-between">
              <DrawerTitle className="flex items-center gap-2">
                <IconUser className="h-5 w-5" />
                User Profile
              </DrawerTitle>
              {onEditClick && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setIsOpen(false);
                    onEditClick();
                  }}
                >
                  <IconEdit className="h-4 w-4" />
                </Button>
              )}
            </div>
            <DrawerDescription>
              Detailed information for <b>{user.username}</b>
            </DrawerDescription>
          </DrawerHeader>

          <div className="flex flex-col gap-6 overflow-y-auto px-4 pb-4">
            <div className="flex flex-col items-center gap-3 p-4 bg-muted/50 rounded-lg">
              <Avatar className="w-16 h-16">
                <AvatarImage
                  src={avatarUrl}
                  alt={`${user.username}'s avatar`}
                />
                <AvatarFallback className="bg-primary/10">
                  <IconUser className="h-8 w-8 text-primary" />
                </AvatarFallback>
              </Avatar>
              <div className="text-center">
                <h3 className="font-semibold text-lg">{user.username}</h3>
                <div className="flex items-center gap-2 justify-center mt-1">
                  <div
                    className={`h-2 w-2 rounded-full ${
                      user.is_online ? "bg-green-500" : "bg-gray-400"
                    }`}
                  />
                  <span className="text-sm text-muted-foreground">
                    {user.is_online ? "Online" : "Offline"}
                  </span>
                </div>
              </div>
            </div>

            <div className="space-y-4">
              <h4 className="font-medium flex items-center">
                Contact Information
              </h4>
              <div className="space-y-3 pl-6">
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Username
                  </Label>
                  <div className="text-sm flex items-center gap-2">
                    <span className="font-medium">{user.username}</span>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => copyToClipboard(user.username, "username")}
                    >
                      {copiedField === "username" ? (
                        <IconCheck className="h-4 w-4 text-green-500" />
                      ) : (
                        <IconCopy className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">Email</Label>
                  <div className="text-sm flex items-center gap-2">
                    <span>{user.email}</span>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => copyToClipboard(user.email, "email")}
                    >
                      {copiedField === "email" ? (
                        <IconCheck className="h-4 w-4 text-green-500" />
                      ) : (
                        <IconCopy className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Language
                  </Label>
                  <div className="text-sm capitalize">{user.language}</div>
                </div>
              </div>
            </div>

            <Separator />

            <div className="space-y-4">
              <h4 className="font-medium flex items-center">
                Role & Permissions
              </h4>
              <div className="space-y-3 pl-6">
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">Role</Label>
                  <Badge
                    variant={getRoleBadgeVariant(user.role)}
                    className="w-fit"
                  >
                    {user.role === "administrator" ? "Administrator" : "User"}
                  </Badge>
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">Rank</Label>
                  <Badge
                    variant="outline"
                    className={`w-fit ${getLevelColor(user.rank)}`}
                  >
                    {user.rank || "N/A"}
                  </Badge>
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Account Status
                  </Label>
                  <Badge
                    variant="outline"
                    className={`w-fit ${getAccountStatusColor(user.account_status)}`}
                  >
                    {user.account_status.charAt(0).toUpperCase() +
                      user.account_status.slice(1)}
                  </Badge>
                </div>
              </div>
            </div>

            <Separator />

            <div className="space-y-4">
              <h4 className="font-medium flex items-center">
                Account Information
              </h4>
              <div className="space-y-3 pl-6">
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    User ID
                  </Label>
                  <div className="flex items-center gap-2">
                    <div className="text-xs font-mono bg-muted px-2 py-1 rounded flex-1">
                      {user.id}
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => copyToClipboard(user.id, "userId")}
                      className="h-6 w-6 p-0"
                    >
                      {copiedField === "userId" ? (
                        <IconCheck className="h-3 w-3 text-green-600" />
                      ) : (
                        <IconCopy className="h-3 w-3" />
                      )}
                    </Button>
                  </div>
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Created
                  </Label>
                  <div className="text-sm">{formatDate(user.created_at)}</div>
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Last Updated
                  </Label>
                  <div className="text-sm">{formatDate(user.updated_at)}</div>
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Avatar
                  </Label>
                  <Badge variant="outline" className="w-fit">
                    {user.has_avatar ? "✓ Has Avatar" : "✗ No Avatar"}
                  </Badge>
                </div>
              </div>
            </div>
          </div>

          <DrawerFooter className="flex flex-row">
            <Button
              variant="default"
              onClick={handleViewDetails}
              className="flex-1"
            >
              View Details
            </Button>
            <Button variant="outline" className="flex-1" onClick={handleClose}>
              Close
            </Button>
          </DrawerFooter>
        </DrawerContent>
      </Drawer>
  );
}
