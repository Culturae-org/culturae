"use client";

import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { apiPost } from "@/lib/api-client";
import { GAMES_ENDPOINTS } from "@/lib/api/endpoints";
import type { AdminGame } from "@/lib/types/games.types";
import {
  IconArchive,
  IconArchiveOff,
  IconRefresh,
} from "@tabler/icons-react";
import * as React from "react";
import { toast } from "sonner";

interface GameArchiveDialogProps {
  game: AdminGame;
  onGameArchived?: (game: AdminGame) => void;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  isArchived?: boolean;
}

export function GameArchiveDialog({
  game,
  onGameArchived,
  open: controlledOpen,
  onOpenChange,
  isArchived = false,
}: GameArchiveDialogProps) {
  const [internalOpen, setInternalOpen] = React.useState(false);
  const open = controlledOpen ?? internalOpen;
  const setOpen = onOpenChange ?? setInternalOpen;
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

  const handleArchive = async () => {
    try {
      setLoading(true);
      setError(null);

      const endpoint = isArchived
        ? GAMES_ENDPOINTS.UNARCHIVE(game.id)
        : GAMES_ENDPOINTS.ARCHIVE(game.id);

      const response = await apiPost(endpoint, {});
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(
          errorData.error ||
            `Failed to ${isArchived ? "unarchive" : "archive"} game`,
        );
      }

      const updatedGame = {
        ...game,
        deleted_at: isArchived ? null : new Date().toISOString(),
      };
      onGameArchived?.(updatedGame);
      setOpen(false);
      toast.success(
        isArchived
          ? "Game restored successfully"
          : "Game archived successfully",
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setLoading(false);
    }
  };

  const _action = isArchived ? "unarchive" : "archive";
  const ActionIcon = isArchived ? IconArchiveOff : IconArchive;

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      {controlledOpen === undefined && (
        <AlertDialogTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className={
              isArchived
                ? "text-green-600 hover:text-green-700 hover:bg-green-50"
                : "text-orange-600 hover:text-orange-700 hover:bg-orange-50"
            }
          >
            <ActionIcon className="h-4 w-4" />
          </Button>
        </AlertDialogTrigger>
      )}
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2">
            {isArchived ? "Unarchive Game" : "Archive Game"}
          </AlertDialogTitle>
          <AlertDialogDescription>
            {isArchived
              ? "This game will be restored and visible in the games list again."
              : "This game will be archived and hidden from the main list. You can unarchive it later. All game data will be preserved for analytics."}
          </AlertDialogDescription>
        </AlertDialogHeader>

        <div className="py-2">
          <div className="text-sm">
            <div>
              <strong>Game ID:</strong> {game.id}
            </div>
            <div>
              <strong>Mode:</strong> {game.mode}
            </div>
            <div>
              <strong>Status:</strong> {game.status}
            </div>
            <div>
              <strong>Players:</strong>{" "}
              {game.current_players ?? game.players?.length ?? 0}
              {game.max_players ? `/${game.max_players}` : ""}
            </div>
            <div>
              <strong>Created:</strong>{" "}
              {new Date(game.created_at).toLocaleDateString()}
            </div>
          </div>
        </div>

        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <AlertDialogFooter>
          <Button
            variant="outline"
            onClick={() => setOpen(false)}
            disabled={loading}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleArchive}
            disabled={loading}
          >
            {loading ? (
              <>
                <IconRefresh className="h-4 w-4 mr-2 animate-spin" />
                {isArchived ? "Restoring..." : "Archiving..."}
              </>
            ) : (
              isArchived ? "Unarchive Game" : "Archive Game"
            )}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
