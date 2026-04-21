"use client";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useUsers } from "@/hooks/useUsers";
import type { AdminUser } from "@/lib/types/user.types";
import { IconAlertTriangle, IconTrash } from "@tabler/icons-react";
import * as React from "react";

interface UserDeleteDialogProps {
  user: AdminUser;
  onUserDeleted?: (userId: string) => void;
  trigger?: React.ReactNode;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export function UserDeleteDialog({
  user,
  onUserDeleted,
  trigger,
  open: controlledOpen,
  onOpenChange,
}: UserDeleteDialogProps) {
  const { deleteUser } = useUsers();
  const [internalOpen, setInternalOpen] = React.useState(false);
  const isOpen = controlledOpen ?? internalOpen;
  const setIsOpen = onOpenChange ?? setInternalOpen;
  const [isLoading, setIsLoading] = React.useState(false);

  const handleDelete = async () => {
    setIsLoading(true);

    try {
      await deleteUser(user.id);

      setIsOpen(false);

      if (onUserDeleted) {
        onUserDeleted(user.id);
      }
    } catch (error) {
      console.error("Error deleting user:", error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      {controlledOpen === undefined && (
        <DialogTrigger asChild>
          {trigger || (
            <Button
              variant="ghost"
              size="sm"
              className="text-red-600 hover:text-red-700 hover:bg-red-50"
            >
              <IconTrash className="h-4 w-4" />
            </Button>
          )}
        </DialogTrigger>
      )}
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <IconTrash className="h-5 w-5 text-red-600" />
            Delete User
          </DialogTitle>
          <DialogDescription>
            Are you sure you want to delete <strong>{user.username}</strong>?
            This action cannot be undone.
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-center gap-3 p-4 bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-800 rounded-lg">
          <IconAlertTriangle className="h-5 w-5 text-red-600 dark:text-red-400 flex-shrink-0" />
          <div className="text-sm">
            <p className="text-red-800 dark:text-red-200 font-medium">
              This action is permanent
            </p>
            <p className="text-red-700 dark:text-red-300">
              All user data, including profile information, sessions, and
              associated content will be permanently removed.
            </p>
          </div>
        </div>

        <div className="space-y-2">
          <div className="text-sm">
            <span className="font-medium">Username:</span> {user.username}
          </div>
          <div className="text-sm">
            <span className="font-medium">Email:</span> {user.email}
          </div>
          <div className="text-sm">
            <span className="font-medium">Role:</span> {user.role}
          </div>
          <div className="text-sm">
            <span className="font-medium">Account Status:</span>{" "}
            {user.account_status}
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => setIsOpen(false)}
            disabled={isLoading}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={isLoading}
          >
            {isLoading ? "Deleting..." : "Delete User"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
