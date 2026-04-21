"use client";

import { Alert, AlertDescription } from "@/components/ui/alert";
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
import { useViewDialog } from "@/hooks/useViewDialog";
import type { UserActionLog } from "@/lib/types/logs.types";
import {
  IconCheck,
  IconCopy,
} from "@tabler/icons-react";
import * as React from "react";

interface UserActionLogViewDialogProps {
  log: UserActionLog;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

function StatusCell({ isSuccess }: { isSuccess: boolean | undefined }) {
  if (isSuccess === undefined) {
    return <Badge variant="outline">Unknown</Badge>;
  }

  return (
    <Badge variant={isSuccess ? "default" : "destructive"}>
      {isSuccess ? "Success" : "Failed"}
    </Badge>
  );
}

function ActionCell({ action }: { action: string | undefined }) {
  if (!action) {
    return <Badge variant="outline">Unknown</Badge>;
  }

  const actionColors: Record<string, string> = {
    profile_update: "bg-blue-100 text-blue-800",
    password_change: "bg-orange-100 text-orange-800",
    avatar_upload: "bg-purple-100 text-purple-800",
    avatar_delete: "bg-purple-100 text-purple-800",
    id_regenerate: "bg-yellow-100 text-yellow-800",
    account_delete: "bg-red-100 text-red-800",
    login: "bg-green-100 text-green-800",
    logout: "bg-gray-100 text-gray-800",
  };

  const colorClass = actionColors[action] || "bg-gray-100 text-gray-800";

  return (
    <Badge className={`${colorClass} border-0`}>
      {action.replace(/_/g, " ")}
    </Badge>
  );
}

export function UserActionLogViewDialog({
  log,
  open: controlledOpen,
  onOpenChange,
}: UserActionLogViewDialogProps) {
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
  const [copiedField, setCopiedField] = React.useState<string | null>(null);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
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
        onOpenChange={handleIsOpenChange}
      >
        <DrawerContent>
          <DrawerHeader className="gap-1">
            <DrawerTitle className="flex items-center">
              Action Details
            </DrawerTitle>
            <DrawerDescription>
              Detailed information for this action
            </DrawerDescription>
          </DrawerHeader>

          <div className="flex flex-col gap-6 overflow-y-auto px-4 pb-4">
            <div className="space-y-4">
              <h4 className="font-medium flex items-center">
                Action Information
              </h4>
              <div className="space-y-3 pl-6">
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Log ID
                  </Label>
                  <div className="flex items-center gap-2">
                    <div className="text-xs font-mono bg-muted px-2 py-1 rounded flex-1 overflow-x-auto">
                      {log.ID || "-"}
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => copyToClipboard(log.ID, "id")}
                      className="h-6 w-6 p-0"
                    >
                      {copiedField === "id" ? (
                        <IconCheck className="h-3 w-3 text-green-600" />
                      ) : (
                        <IconCopy className="h-3 w-3" />
                      )}
                    </Button>
                  </div>
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Action
                  </Label>
                  <ActionCell action={log.Action} />
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Resource
                  </Label>
                  <div className="text-sm">{log.Resource || "-"}</div>
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Resource ID
                  </Label>
                  <div className="flex items-center gap-2">
                    <div className="text-xs font-mono bg-muted px-2 py-1 rounded flex-1 overflow-x-auto">
                      {log.ResourceID || "-"}
                    </div>
                    {log.ResourceID && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() =>
                          log.ResourceID &&
                          copyToClipboard(log.ResourceID, "resourceId")
                        }
                        className="h-6 w-6 p-0"
                      >
                        {copiedField === "resourceId" ? (
                          <IconCheck className="h-3 w-3 text-green-600" />
                        ) : (
                          <IconCopy className="h-3 w-3" />
                        )}
                      </Button>
                    )}
                  </div>
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Status
                  </Label>
                  <StatusCell isSuccess={log.IsSuccess} />
                </div>
              </div>
            </div>

            <Separator />

            <div className="space-y-4">
              <h4 className="font-medium flex items-center">
                User Information
              </h4>
              <div className="space-y-3 pl-6">
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Username
                  </Label>
                  <div className="text-sm">{log.Username || "-"}</div>
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    User ID
                  </Label>
                  <div className="flex items-center gap-2">
                    <div className="text-xs font-mono bg-muted px-2 py-1 rounded flex-1">
                      {log.UserID || "-"}
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => copyToClipboard(log.UserID, "userId")}
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
              </div>
            </div>

            <Separator />

            <div className="space-y-4">
              <h4 className="font-medium flex items-center">
                Connection Information
              </h4>
              <div className="space-y-3 pl-6">
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    IP Address
                  </Label>
                  <div className="flex items-center gap-2">
                    <div className="text-sm font-mono bg-muted px-2 py-1 rounded flex-1 overflow-x-auto">
                      {log.IPAddress || "-"}
                    </div>
                    {log.IPAddress && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() =>
                          log.IPAddress &&
                          copyToClipboard(log.IPAddress, "ipAddress")
                        }
                        className="h-6 w-6 p-0"
                      >
                        {copiedField === "ipAddress" ? (
                          <IconCheck className="h-3 w-3 text-green-600" />
                        ) : (
                          <IconCopy className="h-3 w-3" />
                        )}
                      </Button>
                    )}
                  </div>
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    User Agent
                  </Label>
                  <div className="text-sm text-muted-foreground break-all">
                    {log.UserAgent || "-"}
                  </div>
                </div>
              </div>
            </div>

            <Separator />

            <div className="space-y-4">
              <h4 className="font-medium flex items-center">
                Timestamps
              </h4>
              <div className="space-y-3 pl-6">
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Created At
                  </Label>
                  <div className="text-sm">{formatDate(log.CreatedAt)}</div>
                </div>
              </div>
            </div>

            {log.Details && (
              <>
                <Separator />
                <div className="space-y-4">
                  <h4 className="font-medium">Details</h4>
                  <div className="bg-muted p-3 rounded-lg overflow-x-auto">
                    <pre className="text-xs font-mono whitespace-pre-wrap">
                      {JSON.stringify(log.Details, null, 2)}
                    </pre>
                  </div>
                </div>
              </>
            )}

            {log.ErrorMsg && (
              <>
                <Separator />
                <div className="space-y-4">
                  <Alert variant="destructive">
                    <AlertDescription>
                      <pre className="text-sm whitespace-pre-wrap">
                        {log.ErrorMsg}
                      </pre>
                    </AlertDescription>
                  </Alert>
                </div>
              </>
            )}
          </div>

          <DrawerFooter>
            <DrawerClose asChild>
              <Button variant="outline" className="w-full">
                Close
              </Button>
            </DrawerClose>
          </DrawerFooter>
        </DrawerContent>
      </Drawer>
  );
}
