"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AVATAR_ENDPOINTS } from "@/lib/api/endpoints";
import type {
  GameAnswer,
  GameDetail,
} from "@/lib/types/games.types";
import { IconUsers } from "@tabler/icons-react";
import * as React from "react";
import { Link } from "react-router";

function formatTimeSpent(ms: number) {
  if (ms <= 0) return "—";
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

interface GamePlayersListProps {
  game: GameDetail;
  answers: GameAnswer[];
}

export function GamePlayersList({ game, answers }: GamePlayersListProps) {
  const sortedPlayers = React.useMemo(() => {
    return [...(game.players ?? [])].sort((a, b) => b.score - a.score);
  }, [game.players]);

  const sortedAnswers = React.useMemo(() => {
    return [...answers].sort(
      (a, b) =>
        new Date(a.answered_at).getTime() - new Date(b.answered_at).getTime(),
    );
  }, [answers]);

  if (sortedPlayers.length === 0) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <IconUsers className="h-4 w-4" />
            Leaderboard
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-center py-6 text-sm">
            No players
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <IconUsers className="h-4 w-4" />
          Leaderboard
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {sortedPlayers.map((player, idx) => {
            const playerAnswers = sortedAnswers.filter(
              (a) => a.player_id === player.id,
            );
            const correctCount = playerAnswers.filter(
              (a) => a.is_correct,
            ).length;
            const totalTime = playerAnswers.reduce(
              (sum, a) => sum + a.time_spent,
              0,
            );
            const username =
              player.user?.username ?? player.user_public_id.slice(0, 8);

            return (
              <div
                key={player.id}
                className={`flex items-center justify-between p-3 border rounded-lg ${idx === 0 && game.status === "completed" ? "border-yellow-300 bg-yellow-50/30 dark:border-yellow-800 dark:bg-yellow-950/20" : ""}`}
              >
                <div className="flex items-center gap-3">
                  <span className="text-muted-foreground font-mono text-sm w-6 text-center">
                    {idx === 0 && game.status === "completed"
                      ? "1."
                      : `#${idx + 1}`}
                  </span>
                  <Avatar className="h-8 w-8">
                    <AvatarImage
                      src={
                        player.user?.has_avatar
                          ? AVATAR_ENDPOINTS.GET(player.user_public_id)
                          : undefined
                      }
                    />
                    <AvatarFallback>
                      {username[0]?.toUpperCase() ?? "?"}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <div className="flex items-center gap-2">
                      <Link
                        to={`/users?view=${player.user_public_id}`}
                        className="font-medium hover:underline text-sm"
                      >
                        {username}
                      </Link>
                      {(player.status === "left" ||
                        player.status === "disconnected") && (
                        <Badge
                          variant="outline"
                          className="text-[10px] h-4 px-1.5 text-orange-600 border-orange-300 dark:text-orange-400 dark:border-orange-700"
                        >
                          Left
                        </Badge>
                      )}
                    </div>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <span>
                        {correctCount}/{playerAnswers.length} correct
                      </span>
                      {totalTime > 0 && (
                        <span>
                          avg{" "}
                          {formatTimeSpent(
                            Math.round(totalTime / playerAnswers.length),
                          )}
                          /q
                        </span>
                      )}
                      <span title={new Date(player.joined_at).toLocaleString()}>
                        joined {new Date(player.joined_at).toLocaleTimeString()}
                      </span>
                    </div>
                  </div>
                </div>
                <div className="text-right">
                  <div className="font-bold text-xl">{player.score}</div>
                  <div className="text-xs text-muted-foreground">points</div>
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
