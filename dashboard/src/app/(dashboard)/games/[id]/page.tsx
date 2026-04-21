"use client";

import { GameAnswersList } from "@/components/games/game-answers-list";
import { GameEventTimeline } from "@/components/games/game-event-timeline";
import { GameOverviewCard } from "@/components/games/game-overview-card";
import { GameWinnerCard } from "@/components/games/game-winner-card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { apiGet, apiPost } from "@/lib/api-client";
import { API_BASE, AVATAR_ENDPOINTS, GAMES_ENDPOINTS } from "@/lib/api/endpoints";
import type {
  GameAnswer,
  GameAnswerDetail,
  GameDetail,
  GamePlayer,
  GameQuestion,
  GeoQuestionData,
} from "@/lib/types/games.types";
import {
  IconArrowLeft,
  IconCheck,
  IconClock,
  IconCopy,
  IconPlayerStop,
  IconRefresh,
  IconUsers,
} from "@tabler/icons-react";
import * as React from "react";
import { Link, useNavigate, useParams } from "react-router";
import { toast } from "sonner";

function formatDuration(start: string, end: string | null) {
  if (!end) return "Ongoing";
  const ms = new Date(end).getTime() - new Date(start).getTime();
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

function formatTimeSpent(ms: number) {
  if (ms <= 0) return "—";
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

function CompactPlayers({
  game,
  answers,
}: {
  game: GameDetail;
  answers: GameAnswer[];
}) {
  const sortedPlayers = React.useMemo(
    () => [...(game.players ?? [])].sort((a, b) => b.score - a.score),
    [game.players],
  );

  return (
    <Card className="border-0 dark:border">
      <CardHeader className="pb-2 pt-3 px-4">
        <CardTitle className="text-sm font-medium flex items-center gap-1.5">
          <IconUsers className="h-3.5 w-3.5" />
          Players ({game.players.length})
        </CardTitle>
      </CardHeader>
      <CardContent className="px-4 pb-3">
        {sortedPlayers.length === 0 ? (
          <p className="text-xs text-muted-foreground">No players yet</p>
        ) : (
          <div className="space-y-2">
            {sortedPlayers.map((player, idx) => {
              const playerAnswers = answers.filter(
                (a) => a.player_id === player.id,
              );
              const correctCount = playerAnswers.filter((a) => a.is_correct).length;
              const totalTime = playerAnswers.reduce(
                (sum, a) => sum + a.time_spent,
                0,
              );
              const username =
                player.user?.username ?? player.user_public_id.slice(0, 8);

              return (
                <div
                  key={player.id}
                  className="flex items-center justify-between py-0.5"
                >
                  <div className="flex items-center gap-2 min-w-0">
                    <span className="text-xs text-muted-foreground font-mono w-5 shrink-0">
                      #{idx + 1}
                    </span>
                    <Avatar className="h-6 w-6 shrink-0">
                      <AvatarImage
                        src={
                          player.user?.has_avatar
                            ? AVATAR_ENDPOINTS.GET(player.user_public_id)
                            : undefined
                        }
                      />
                      <AvatarFallback className="text-[10px]">
                        {username[0]?.toUpperCase() ?? "?"}
                      </AvatarFallback>
                    </Avatar>
                    <div className="min-w-0">
                      <div className="flex items-center gap-1.5 flex-wrap">
                        <Link
                          to={`/users?view=${player.user_public_id}`}
                          className="text-sm font-medium hover:underline truncate"
                        >
                          {username}
                        </Link>
                        {(player.status === "left" ||
                          player.status === "disconnected") && (
                          <Badge
                            variant="outline"
                            className="text-[10px] h-4 px-1 text-orange-600 border-orange-300 dark:text-orange-400 dark:border-orange-700"
                          >
                            Left
                          </Badge>
                        )}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {correctCount}/{playerAnswers.length} correct
                        {totalTime > 0 && (
                          <span className="ml-1.5">
                            avg{" "}
                            {formatTimeSpent(
                              Math.round(totalTime / (playerAnswers.length || 1)),
                            )}
                            /q
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                  <span className="text-sm font-bold font-mono tabular-nums ml-2 shrink-0">
                    {player.score}
                  </span>
                </div>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default function GameDetailPage() {
  const params = useParams();
  const navigate = useNavigate();
  const gameId = params.id as string;

  const [game, setGame] = React.useState<GameDetail | null>(null);
  const [questions, setQuestions] = React.useState<GameQuestion[]>([]);
  const [enrichedAnswers, setEnrichedAnswers] = React.useState<GameAnswerDetail[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [copiedField, setCopiedField] = React.useState<string | null>(null);
  const [countryNames, setCountryNames] = React.useState<Record<string, string>>({});

  const fetchAll = React.useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const [gameRes, questionsRes, answersRes] = await Promise.all([
        apiGet(GAMES_ENDPOINTS.GET(gameId)),
        apiGet(GAMES_ENDPOINTS.GET_QUESTIONS(gameId)),
        apiGet(GAMES_ENDPOINTS.GET_ANSWERS(gameId)),
      ]);

      if (!gameRes.ok) throw new Error("Failed to fetch game");
      const gameJson = await gameRes.json();
      const gameData: GameDetail = gameJson.data ?? gameJson;
      setGame(gameData);

      if (questionsRes.ok) {
        const questionsJson = await questionsRes.json();
        const raw = questionsJson.data ?? questionsJson;
        setQuestions(Array.isArray(raw) ? raw : []);
      }

      if (answersRes.ok) {
        const answersJson = await answersRes.json();
        const raw = answersJson.data ?? answersJson;
        const sorted = (Array.isArray(raw) ? raw : []) as GameAnswerDetail[];
        sorted.sort(
          (a, b) =>
            new Date(a.answered_at).getTime() - new Date(b.answered_at).getTime(),
        );
        setEnrichedAnswers(sorted);
      }

      if (gameData.category === "geography" || gameData.category === "flags") {
        try {
          const countriesRes = await apiGet(
            `${API_BASE}/geography/countries?limit=300`,
          );
          if (countriesRes.ok) {
            const countriesJson = await countriesRes.json();
            const countries = countriesJson.data ?? countriesJson;
            const lookup: Record<string, string> = {};
            if (Array.isArray(countries)) {
              for (const c of countries) {
                const getCountryName = (nameObj: unknown): string => {
                  if (typeof nameObj === "string") return nameObj;
                  if (typeof nameObj === "object" && nameObj) {
                    const n = nameObj as Record<string, string>;
                    return n.en || n.fr || Object.values(n)[0] || "";
                  }
                  return "";
                };
                const name = getCountryName(c.name);
                if (c.slug) lookup[c.slug.toLowerCase()] = name;
                if (c.iso_alpha2) lookup[c.iso_alpha2.toLowerCase()] = name;
                if (c.iso_alpha3) lookup[c.iso_alpha3.toLowerCase()] = name;
              }
            }
            for (const q of gameData.questions ?? []) {
              const d = q.data as GeoQuestionData | undefined;
              if (d?.target_slug && d?.target_name) {
                const name =
                  d.target_name.en ||
                  d.target_name.fr ||
                  Object.values(d.target_name)[0] ||
                  "";
                lookup[d.target_slug.toLowerCase()] = name;
              }
              if (d?.target_iso2 && d?.target_name) {
                const name =
                  d.target_name.en ||
                  d.target_name.fr ||
                  Object.values(d.target_name)[0] ||
                  "";
                lookup[d.target_iso2.toLowerCase()] = name;
              }
            }
            setCountryNames(lookup);
          }
        } catch {
          // Non-critical
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setLoading(false);
    }
  }, [gameId]);

  React.useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  const handleCancel = async () => {
    if (!confirm("Cancel this game?")) return;
    const prev = game;
    setGame((g) => (g ? { ...g, status: "cancelled" } : g));
    try {
      const res = await apiPost(GAMES_ENDPOINTS.CANCEL(gameId));
      if (!res.ok) throw new Error("Failed to cancel game");
      toast.success("Game cancelled");
    } catch {
      setGame(prev);
      toast.error("Failed to cancel game");
    }
  };

  const copyToClipboard = async (text: string, field: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedField(field);
      setTimeout(() => setCopiedField(null), 2000);
    } catch {
      /* silent */
    }
  };

  const playerMap = React.useMemo(() => {
    const map: Record<string, GamePlayer> = {};
    for (const p of game?.players ?? []) {
      map[p.id] = p;
    }
    return map;
  }, [game?.players]);

  const sortedAnswers = React.useMemo(() => {
    return [...(game?.answers ?? [])].sort(
      (a, b) =>
        new Date(a.answered_at).getTime() - new Date(b.answered_at).getTime(),
    );
  }, [game?.answers]);

  const isRunning =
    game?.status === "waiting" ||
    game?.status === "ready" ||
    game?.status === "in_progress";

  if (loading) {
    return (
      <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
        <div className="px-4 lg:px-6">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-4 w-64 mt-2" />
        </div>
        <div className="px-4 lg:px-6 space-y-4">
          <Skeleton className="h-6 w-96" />
          <Skeleton className="h-32 w-full" />
          <Skeleton className="h-64 w-full" />
        </div>
      </div>
    );
  }

  if (error || !game) {
    return (
      <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
        <div className="px-4 lg:px-6">
          <Button variant="ghost" onClick={() => navigate(-1)} className="mb-4">
            <IconArrowLeft className="h-4 w-4 mr-2" />
            Back to Games
          </Button>
          <div className="text-center py-8">
            <p className="text-muted-foreground">Failed to load game details</p>
            <Button onClick={fetchAll} className="mt-4">
              <IconRefresh className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>
    );
  }

  const winnerPlayer = game.winner_id ? (playerMap[game.winner_id] ?? null) : null;

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      {/* Header */}
      <div className="px-4 lg:px-6">
        <Button variant="ghost" onClick={() => navigate(-1)} className="mb-4">
          <IconArrowLeft className="h-4 w-4 mr-2" />
          Back to Games
        </Button>
        <div className="flex items-center justify-between flex-wrap gap-3">
          <div>
            <h1 className="text-2xl font-bold">Game Details</h1>
            <div className="flex items-center gap-2 mt-1 flex-wrap">
              <span className="text-muted-foreground font-mono text-sm">
                {game.public_id}
              </span>
              <Button
                variant="ghost"
                size="sm"
                className="h-6 w-6 p-0"
                onClick={() => copyToClipboard(game.id, "gameId")}
              >
                {copiedField === "gameId" ? (
                  <IconCheck className="h-3 w-3 text-green-600" />
                ) : (
                  <IconCopy className="h-3 w-3" />
                )}
              </Button>
              <span className="text-muted-foreground text-xs font-mono opacity-50">
                {game.id}
              </span>
            </div>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={fetchAll}>
              <IconRefresh className="h-4 w-4" />
            </Button>
            {isRunning && (
              <Button variant="outline" size="sm" onClick={handleCancel}>
                <IconPlayerStop className="h-4 w-4 mr-2" />
                Cancel
              </Button>
            )}
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6 space-y-4">
        {/* Compact meta strip */}
        <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm">
          <div className="flex items-center gap-1.5">
            <span className="text-xs text-muted-foreground uppercase tracking-wide">
              Mode
            </span>
            <Badge variant="outline" className="capitalize h-5 text-xs px-1.5">
              {game.mode}
            </Badge>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="text-xs text-muted-foreground uppercase tracking-wide">
              Status
            </span>
            <Badge variant="outline" className="capitalize h-5 text-xs px-1.5">
              {game.status.replace("_", " ")}
            </Badge>
          </div>
          {game.category && (
            <div className="flex items-center gap-1.5">
              <span className="text-xs text-muted-foreground uppercase tracking-wide">
                Category
              </span>
              <span className="capitalize">{game.category}</span>
              {game.flag_variant && (
                <span className="text-xs text-muted-foreground">
                  · {game.flag_variant}
                </span>
              )}
            </div>
          )}
          <div className="flex items-center gap-1.5">
            <span className="text-xs text-muted-foreground uppercase tracking-wide">
              Questions
            </span>
            <span className="font-mono tabular-nums">
              {sortedAnswers.length} / {game.question_count}
            </span>
          </div>
          <div className="flex items-center gap-1.5">
            <IconClock className="h-3 w-3 text-muted-foreground" />
            <span>
              {game.started_at
                ? formatDuration(game.started_at, game.completed_at)
                : "Not started"}
            </span>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          <GameOverviewCard game={game} />
          <CompactPlayers game={game} answers={sortedAnswers} />
        </div>

        <GameWinnerCard game={game} winnerPlayer={winnerPlayer} />

        {/* Tabs: Answers + Events */}
        <Tabs defaultValue="answers" className="w-full">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="answers">
              Answers ({enrichedAnswers.length || sortedAnswers.length})
            </TabsTrigger>
            <TabsTrigger value="events">Events</TabsTrigger>
          </TabsList>

          <TabsContent value="answers">
            <GameAnswersList
              enrichedAnswers={enrichedAnswers}
              playerMap={playerMap}
              questions={questions}
              gameCategory={game.category}
              flagVariant={game.flag_variant || undefined}
              countryNames={countryNames}
            />
          </TabsContent>

          <TabsContent value="events" className="pt-4">
            <GameEventTimeline gameId={gameId} />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
