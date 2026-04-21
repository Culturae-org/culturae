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
import { IconRefresh } from "@tabler/icons-react";
import * as React from "react";

interface GameCancelDialogProps {
  game: AdminGame;
  onGameUpdated?: (updatedGame: AdminGame) => void;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export function GameCancelDialog({
  game,
  onGameUpdated,
  open: controlledOpen,
  onOpenChange,
}: GameCancelDialogProps) {
  const [internalOpen, setInternalOpen] = React.useState(false);
  const open = controlledOpen ?? internalOpen;
  const setOpen = onOpenChange ?? setInternalOpen;
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

  const handleCancel = async () => {
    try {
      setLoading(true);
      setError(null);

      const response = await apiPost(GAMES_ENDPOINTS.CANCEL(game.id), {});
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || "Failed to cancel game");
      }

      const updatedGame = { ...game, status: "cancelled" };
      onGameUpdated?.(updatedGame);

      setOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setLoading(false);
    }
  };

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      {controlledOpen === undefined && (
        <AlertDialogTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="text-orange-600 hover:text-orange-700 hover:bg-orange-50"
          />
        </AlertDialogTrigger>
      )}
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2">
            Cancel Game
          </AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to cancel this game? This action cannot be
            undone. All players will be notified and the game will be marked as
            cancelled.
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
              <strong>Players:</strong>{" "}
              {game.current_players ?? game.players?.length ?? 0}
              {game.max_players ? `/${game.max_players}` : ""}
            </div>
            <div>
              <strong>Status:</strong> {game.status}
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
            onClick={handleCancel}
            disabled={loading}
          >
            {loading ? (
              <>
                <IconRefresh className="h-4 w-4 mr-2 animate-spin" />
                Cancelling...
              </>
            ) : (
              <>Cancel Game</>
            )}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
