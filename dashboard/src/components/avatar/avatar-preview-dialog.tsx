"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface AvatarPreviewDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  user: {
    username: string;
    avatarUrl: string;
  } | null;
}

export function AvatarPreviewDialog({
  open,
  onOpenChange,
  user,
}: AvatarPreviewDialogProps) {
  if (!user) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[90vw] sm:max-w-[40vw] max-h-[80vh] flex flex-col p-4 sm:p-6 rounded-3xl">
        <DialogHeader className="sr-only">
          <DialogTitle>{user.username}&apos;s Avatar</DialogTitle>
        </DialogHeader>
        <div className="flex items-center justify-center p-1 flex-1 min-h-[300px]">
          <Avatar className="w-auto h-auto max-w-full max-h-[60vh] rounded-3xl border-1 border-background overflow-visible">
            <AvatarImage
              src={user.avatarUrl}
              alt={user.username}
              className="object-contain w-auto h-auto max-w-full max-h-[60vh] rounded-3xl aspect-auto"
            />
            <AvatarFallback className="rounded-3xl text-[10vw] w-64 h-64 aspect-square">
              {user.username.charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
        </div>
      </DialogContent>
    </Dialog>
  );
}
