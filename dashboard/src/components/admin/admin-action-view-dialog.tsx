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
  DrawerTrigger,
} from "@/components/ui/drawer";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { useIsMobile } from "@/hooks/useMobile";
import { useViewDialog } from "@/hooks/useViewDialog";
import { usersService } from "@/lib/services/users.service";
import type { AdminUser } from "@/lib/types/user.types";
import {
  IconArrowRight,
  IconEye,
  IconShield,
  IconTarget,
  IconUser,
} from "@tabler/icons-react";
import * as React from "react";

import type { AdminActionLog } from "@/lib/types/logs.types";

interface AdminActionViewDialogProps {
  action: AdminActionLog;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export function AdminActionViewDialog({
  action,
  open: controlledOpen,
  onOpenChange,
}: AdminActionViewDialogProps) {
  const isMobile = useIsMobile();
  const viewDialog = useViewDialog("view");
  const isOpen =
    controlledOpen !== undefined ? controlledOpen : viewDialog.isOpen;
  const handleIsOpenChange = (open: boolean) => {
    if (onOpenChange) {
      onOpenChange(open);
    } else if (!open) {
      viewDialog.close();
    }
  };
  const [affectedUser, setAffectedUser] = React.useState<AdminUser | null>(
    null,
  );
  const [loadingUser, setLoadingUser] = React.useState(false);
  const hasFetchedUser = React.useRef(false);

  React.useEffect(() => {
    if (
      isOpen &&
      action.Resource === "user" &&
      action.ResourceID &&
      !hasFetchedUser.current
    ) {
      hasFetchedUser.current = true;
      setLoadingUser(true);
      usersService
        .getUserById(action.ResourceID)
        .then((user) => setAffectedUser(user))
        .catch((error) => {
          console.error("Failed to fetch affected user:", error);
          hasFetchedUser.current = false;
        })
        .finally(() => setLoadingUser(false));
    }
  }, [isOpen, action.Resource, action.ResourceID]);

  React.useEffect(() => {
    if (!isOpen) {
      setAffectedUser(null);
      setLoadingUser(false);
      hasFetchedUser.current = false;
    }
  }, [isOpen]);

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

  const getActionDescription = (actionType: string, resource: string) => {
    const descriptions = {
      user_status_active: "Restored user account access and functionality",
      user_status_suspended: "Temporarily restricted user account access",
      user_status_banned: "Permanently blocked user account access",
      user_deactivate: "Disabled user account without deleting data",
      user_activate: "Enabled previously deactivated user account",
      user_delete: "Permanently removed user account and all associated data",
      user_role_update: "Modified user role and associated permissions",
      user_level_update: "Changed user experience level",
      bulk_user_role_update: "Updated roles for multiple users simultaneously",
      bulk_user_level_update: "Modified experience levels for multiple users",
      bulk_user_status_update: "Changed account status for multiple users",
      bulk_user_delete: "Permanently removed multiple user accounts",
      avatar_upload: "Uploaded or updated user profile picture",
      avatar_delete: "Removed user profile picture",
      avatar_delete_self: "User removed their own profile picture",
      session_logout: "Terminated active user session",
      session_logout_all: "Ended all active sessions for the user",
      delete_game: "Permanently removed game and all associated data",
      cancel_game: "Stopped active game and notified participants",
      cleanup_abandoned_games: "Removed inactive games that were abandoned",
      run_game_maintenance: "Executed maintenance tasks on game system",
      cancel_game_invite: "Revoked pending game invitation",
      default: `Performed ${actionType.replace(/_/g, " ")} on ${resource}`,
    };

    return (
      descriptions[actionType as keyof typeof descriptions] ||
      descriptions.default
    );
  };

  const renderDetails = (details: Record<string, unknown>) => {
    if (!details || Object.keys(details).length === 0) {
      return (
        <div className="text-sm text-muted-foreground">
          No additional details
        </div>
      );
    }

    if (action.Resource === "user" && details.account_status) {
      return (
        <div className="space-y-3">
          <div className="flex items-center gap-2 p-3 rounded-lg border">
            <IconTarget className="h-4 w-4 text-muted-foreground" />
            <div>
              <div className="text-sm font-medium">Account Status Changed</div>
              <div className="text-sm text-muted-foreground">
                New status:{" "}
                <Badge className="ml-1" variant="outline">
                  {String(details.account_status)}
                </Badge>
              </div>
            </div>
          </div>
        </div>
      );
    }

    if (action.Resource === "user" && details.role) {
      return (
        <div className="space-y-3">
          <div className="flex items-center gap-2 p-3 rounded-lg border">
            <IconShield className="h-4 w-4 text-muted-foreground" />
            <div>
              <div className="text-sm font-medium">Role Updated</div>
              <div className="text-sm text-muted-foreground">
                New role:{" "}
                <Badge className="ml-1" variant="outline">
                  {String(details.role)}
                </Badge>
              </div>
            </div>
          </div>
        </div>
      );
    }

    return (
      <div className="space-y-2">
        {Object.entries(details).map(([key, value]) => (
          <div key={key} className="flex flex-col gap-1">
            <Label className="text-xs text-muted-foreground capitalize">
              {key.replace(/_/g, " ")}
            </Label>
            <div className="text-sm bg-muted px-2 py-1 rounded">
              {typeof value === "object"
                ? JSON.stringify(value, null, 2)
                : String(value)}
            </div>
          </div>
        ))}
      </div>
    );
  };

  const formatValue = (value: unknown): string => {
    if (typeof value === "boolean") return value ? "true" : "false";
    if (value === null || value === undefined) return "N/A";
    return String(value);
  };

  const renderChanges = (
    changes: Record<string, { from: unknown; to: unknown }>,
  ) => {
    if (!changes || Object.keys(changes).length === 0) {
      return (
        <div className="text-sm text-muted-foreground">No changes recorded</div>
      );
    }

    return (
      <div className="space-y-3">
        {Object.entries(changes).map(([field, change]) => (
          <div
            key={field}
            className="flex items-center gap-3 p-3 rounded-lg border"
          >
            <IconArrowRight className="h-4 w-4 text-muted-foreground" />
            <div className="flex-1">
              <div className="text-sm font-medium capitalize">
                {field.replace(/_/g, " ")}
              </div>
              <div className="flex items-center gap-2 mt-1 flex-wrap">
                <Badge variant="outline" className="font-mono text-xs">
                  {formatValue(change.from)}
                </Badge>
                <IconArrowRight className="h-3 w-3 text-muted-foreground flex-shrink-0" />
                <Badge variant="outline" className="font-mono text-xs">
                  {formatValue(change.to)}
                </Badge>
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  };

  const renderChangeSummary = (
    action: AdminActionLog,
    affectedUser: AdminUser | null,
  ) => {
    if (
      action.Details?.changes &&
      typeof action.Details?.changes === "object"
    ) {
      const changes = action.Details?.changes as Record<
        string,
        { from: unknown; to: unknown }
      >;
      const summary =
        typeof action.Details?.summary === "string"
          ? action.Details?.summary
          : null;

      return (
        <div className="space-y-4">
          {summary && (
            <div className="p-3 rounded-lg border">
              <div className="text-sm font-medium">{summary}</div>
            </div>
          )}

          {renderChanges(changes)}
        </div>
      );
    }

    if (
      action.Resource === "user" &&
      action.Details?.account_status &&
      affectedUser
    ) {
      const newStatus = String(action.Details?.account_status);
      const currentStatus = affectedUser.account_status;

      return (
        <div className="space-y-3">
          <div className="flex items-center gap-3 p-3 rounded-lg border">
            <IconArrowRight className="h-4 w-4 text-muted-foreground" />
            <div className="flex-1">
              <div className="text-sm font-medium">Account Status Change</div>
              <div className="flex items-center gap-2 mt-1">
                <Badge variant="outline">{currentStatus}</Badge>
                <IconArrowRight className="h-3 w-3 text-muted-foreground" />
                <Badge variant="outline">{newStatus}</Badge>
              </div>
            </div>
          </div>
        </div>
      );
    }

    if (action.Resource === "user" && action.Details?.role && affectedUser) {
      const newRole = String(action.Details?.role);
      const currentRole = affectedUser.role;

      return (
        <div className="space-y-3">
          <div className="flex items-center gap-3 p-3 rounded-lg border">
            <IconArrowRight className="h-4 w-4 text-muted-foreground" />
            <div className="flex-1">
              <div className="text-sm font-medium">Role Change</div>
              <div className="flex items-center gap-2 mt-1">
                <Badge variant="outline">{currentRole}</Badge>
                <IconArrowRight className="h-3 w-3 text-muted-foreground" />
                <Badge variant="outline">{newRole}</Badge>
              </div>
            </div>
          </div>
        </div>
      );
    }

    if (action.Action.startsWith("bulk_")) {
      return (
        <div className="space-y-3">
          <div className="p-3 rounded-lg border">
            <div className="text-sm font-medium">Bulk Operation</div>
            <div className="text-sm text-muted-foreground mt-1">
              This action affected multiple resources simultaneously
            </div>
          </div>
          {renderDetails(action.Details ?? {})}
        </div>
      );
    }

    return renderDetails(action.Details ?? {});
  };

  return (
    <Drawer
      direction={isMobile ? "bottom" : "right"}
      open={isOpen}
      onOpenChange={handleIsOpenChange}
    >
      {controlledOpen === undefined && (
        <DrawerTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handleIsOpenChange(true)}
          >
            <IconEye className="h-4 w-4 mr-2" />
            View
          </Button>
        </DrawerTrigger>
      )}
      <DrawerContent className="max-w-md">
        <DrawerHeader className="gap-1">
          <DrawerTitle className="flex items-center">
            Admin Action Details
          </DrawerTitle>
          <DrawerDescription>
            Detailed information for admin action {action.ID}
          </DrawerDescription>
        </DrawerHeader>

        <div className="flex flex-col gap-6 overflow-y-auto px-4 pb-4">
          {/* Action Overview */}
          <div className="flex flex-col gap-3 p-4 bg-muted/50 rounded-lg">
            <div className="flex items-center gap-2">
              <Badge variant="outline">
                {action.Action.replace(/_/g, " ")}
              </Badge>
              <Badge variant="outline">{action.Resource}</Badge>
            </div>
            <div className="text-center">
              <h3 className="font-semibold text-lg">{action.AdminName}</h3>
              <div className="flex items-center gap-2 justify-center mt-1">
                <div className="h-2 w-2 rounded-full bg-muted-foreground/40" />
                <span className="text-sm text-muted-foreground">
                  {action.IsSuccess ? "Success" : "Failed"}
                </span>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="font-medium flex items-center">Action Details</h4>
            <div className="space-y-3 pl-6">
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  Action ID
                </Label>
                <div className="text-xs font-mono bg-muted px-2 py-1 rounded w-fit">
                  {action.ID}
                </div>
              </div>
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  Action Type
                </Label>
                <div className="text-sm font-medium">
                  {action.Action.replace(/_/g, " ")}
                </div>
                <div className="text-xs text-muted-foreground mt-1">
                  {getActionDescription(action.Action, action.Resource)}
                </div>
              </div>
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  Resource Type
                </Label>
                <div className="text-sm">{action.Resource}</div>
              </div>
              {action.ResourceID && (
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Resource ID
                  </Label>
                  <div className="text-xs font-mono bg-muted px-2 py-1 rounded w-fit">
                    {action.ResourceID}
                  </div>
                </div>
              )}
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  Execution Status
                </Label>
                <Badge variant="outline" className="w-fit">
                  {action.IsSuccess ? "Success" : "Failed"}
                </Badge>
              </div>
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <h4 className="font-medium flex items-center">Administrator</h4>
            <div className="space-y-3 pl-6">
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  Admin Name
                </Label>
                <div className="text-sm font-medium">{action.AdminName}</div>
              </div>
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  Admin ID
                </Label>
                <div className="text-xs font-mono bg-muted px-2 py-1 rounded w-fit">
                  {action.AdminID}
                </div>
              </div>
            </div>
          </div>

          <Separator />

          {action.Resource === "user" && action.ResourceID && (
            <div className="space-y-4">
              <h4 className="font-medium flex items-center">Affected User</h4>
              <div className="space-y-3 pl-6">
                {loadingUser ? (
                  <div className="text-sm text-muted-foreground">
                    Loading user information...
                  </div>
                ) : affectedUser ? (
                  <div className="space-y-3">
                    <div className="flex items-center gap-3 p-3 rounded-lg border">
                      <IconUser className="h-5 w-5 text-muted-foreground" />
                      <div>
                        <div className="text-sm font-medium">
                          {affectedUser.username}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {affectedUser.email}
                        </div>
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-3 text-sm">
                      <div className="flex flex-col gap-1">
                        <Label className="text-xs text-muted-foreground">
                          User ID
                        </Label>
                        <div className="text-xs font-mono bg-muted px-2 py-1 rounded">
                          {affectedUser.id.slice(0, 8)}...
                        </div>
                      </div>
                      <div className="flex flex-col gap-1">
                        <Label className="text-xs text-muted-foreground">
                          Current Status
                        </Label>
                        <Badge variant="outline" className="w-fit">
                          {affectedUser.account_status}
                        </Badge>
                      </div>
                      <div className="flex flex-col gap-1">
                        <Label className="text-xs text-muted-foreground">
                          Role
                        </Label>
                        <Badge variant="outline" className="w-fit">
                          {affectedUser.role}
                        </Badge>
                      </div>
                      <div className="flex flex-col gap-1">
                        <Label className="text-xs text-muted-foreground">
                          Level
                        </Label>
                        <Badge variant="outline" className="w-fit">
                          {affectedUser.level}
                        </Badge>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="text-sm text-muted-foreground">
                    User information not available (ID:{" "}
                    {action.ResourceID.slice(0, 8)}...)
                  </div>
                )}
              </div>
            </div>
          )}

          <Separator />
          <div className="space-y-4">
            <h4 className="font-medium flex items-center">
              Client Information
            </h4>
            <div className="space-y-3 pl-6">
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  IP Address
                </Label>
                <div className="text-sm font-mono">
                  {action.IPAddress || "N/A"}
                </div>
              </div>
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  User Agent
                </Label>
                <div className="text-sm break-all">
                  {action.UserAgent || "N/A"}
                </div>
              </div>
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <h4 className="font-medium flex items-center">Change Summary</h4>
            <div className="space-y-3 pl-6">
              {renderChangeSummary(action, affectedUser)}
            </div>
          </div>

          {action.ErrorMsg && (
            <>
              <Separator />
              <div className="space-y-4">
                <h4 className="font-medium flex items-center">Error Details</h4>
                <div className="space-y-3 pl-6">
                  <div className="flex flex-col gap-1">
                    <Label className="text-xs text-muted-foreground">
                      Error Message
                    </Label>
                    <div className="text-sm bg-muted p-3 rounded border">
                      {action.ErrorMsg}
                    </div>
                  </div>
                </div>
              </div>
            </>
          )}

          <Separator />

          <div className="space-y-4">
            <h4 className="font-medium flex items-center">Timestamp</h4>
            <div className="space-y-3 pl-6">
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  Executed At
                </Label>
                <div className="text-sm">{formatDate(action.CreatedAt)}</div>
              </div>
            </div>
          </div>
        </div>

        <DrawerFooter>
          <DrawerClose asChild>
            <Button variant="outline">Close</Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  );
}
