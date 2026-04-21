"use client";

import { Card, CardContent } from "@/components/ui/card";
import type { GameDetail, GamePlayer } from "@/lib/types/games.types";

interface GameWinnerCardProps {
  game: GameDetail;
  winnerPlayer: GamePlayer | null;
}

export function GameWinnerCard({ game, winnerPlayer }: GameWinnerCardProps) {
  if (!game.winner_id) return null;

  const winnerName =
    winnerPlayer?.user?.username ?? game.winner_id?.slice(0, 8);

  return (
    <Card>
      <CardContent className="pt-4 pb-4 flex items-center gap-3">
        <div>
          <div className="text-xs text-muted-foreground mb-0.5">Winner</div>
          <div className="font-semibold">{winnerName}</div>
        </div>
        {winnerPlayer && (
          <div className="ml-auto text-right">
            <div className="text-xs text-muted-foreground mb-0.5">
              Final Score
            </div>
            <div className="font-bold text-lg">{winnerPlayer.score}</div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
