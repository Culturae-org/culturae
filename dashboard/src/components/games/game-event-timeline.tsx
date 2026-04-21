"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { apiGet } from "@/lib/api-client";
import { GAMES_ENDPOINTS } from "@/lib/api/endpoints";
import type { GameEventLog } from "@/lib/types/games.types";
import { IconAlertTriangle, IconRefresh } from "@tabler/icons-react";
import * as React from "react";

const EVENT_STYLES: Record<string, { color: string; label: string }> = {
  game_created: { color: "bg-blue-500", label: "Created" },
  game_ready: { color: "bg-blue-400", label: "Ready" },
  game_started: { color: "bg-green-500", label: "Started" },
  game_completed: { color: "bg-green-700", label: "Completed" },
  game_cancelled: { color: "bg-red-500", label: "Cancelled" },
  player_joined: { color: "bg-teal-500", label: "Player Joined" },
  player_left: { color: "bg-orange-500", label: "Player Left" },
  player_ready: { color: "bg-teal-400", label: "Player Ready" },
  player_disconnected: { color: "bg-orange-600", label: "Disconnected" },
  player_reconnected: { color: "bg-teal-600", label: "Reconnected" },
  question_sent: { color: "bg-purple-500", label: "Question Sent" },
  answer_received: { color: "bg-indigo-500", label: "Answer" },
  score_updated: { color: "bg-indigo-400", label: "Score" },
  question_timeout: { color: "bg-yellow-500", label: "Timeout" },
  game_error: { color: "bg-red-600", label: "Error" },
};

function getEventStyle(eventType: string) {
  return (
    EVENT_STYLES[eventType] ?? {
      color: "bg-muted-foreground",
      label: eventType,
    }
  );
}

function formatEventData(
  eventType: string,
  data: Record<string, unknown>,
): string {
  if (!data || Object.keys(data).length === 0) return "";

  switch (eventType) {
    case "player_joined":
    case "player_left":
    case "player_disconnected":
    case "player_reconnected":
    case "player_ready": {
      const userId = (data.user_id as string | undefined)?.slice(0, 8);
      const ready =
        data.ready !== undefined
          ? data.ready
            ? "ready"
            : "not ready"
          : undefined;
      return [userId ? `user: ${userId}...` : null, ready]
        .filter(Boolean)
        .join(" · ");
    }
    case "answer_received": {
      const correct = data.is_correct ? "✓ correct" : "✗ wrong";
      const pts = data.points !== undefined ? `+${data.points}pts` : "";
      const ms =
        data.time_spent_ms !== undefined
          ? `${Math.round(data.time_spent_ms as number)}ms`
          : "";
      return [correct, pts, ms].filter(Boolean).join(" · ");
    }
    case "score_updated": {
      return `score: ${data.score}`;
    }
    case "question_sent": {
      const n = data.question_number;
      const total = data.total_questions;
      const limit =
        data.time_limit !== undefined ? `${data.time_limit}s limit` : "";
      return [`Q${n}/${total}`, limit].filter(Boolean).join(" · ");
    }
    case "question_timeout": {
      return `question #${data.question_number}`;
    }
    case "game_error": {
      return String(data.error ?? "");
    }
    default:
      return "";
  }
}

function formatTime(iso: string, reference: Date): string {
  const d = new Date(iso);
  const diffMs = d.getTime() - reference.getTime();
  if (Math.abs(diffMs) < 1000) return "+0s";
  const sign = diffMs >= 0 ? "+" : "-";
  const abs = Math.abs(diffMs);
  const s = Math.floor(abs / 1000);
  if (s < 60) return `${sign}${s}s`;
  const m = Math.floor(s / 60);
  return `${sign}${m}m${s % 60}s`;
}

function formatAbsTime(iso: string): string {
  return new Date(iso).toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });
}

export function GameEventTimeline({ gameId }: { gameId: string }) {
  const [events, setEvents] = React.useState<GameEventLog[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [refreshing, setRefreshing] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

  const fetchEvents = React.useCallback(
    async (isRefresh = false) => {
      try {
        if (isRefresh) setRefreshing(true);
        else setLoading(true);

        const res = await apiGet(GAMES_ENDPOINTS.GET_EVENTS(gameId));
        if (!res.ok) throw new Error("Failed to fetch events");
        const json = await res.json();
        const raw = json.data ?? json;
        setEvents(Array.isArray(raw) ? raw : []);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load events");
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [gameId],
  );

  React.useEffect(() => {
    fetchEvents();
  }, [fetchEvents]);

  const reference = React.useMemo(() => {
    const first =
      events.find((e) => e.event_type === "game_started") ?? events[0];
    return first ? new Date(first.occurred_at) : new Date();
  }, [events]);

  if (loading) {
    return (
      <div className="space-y-3 py-4">
        {[0, 1, 2, 3, 4, 5, 6, 7].map((i) => (
          <div key={`timeline-skel-${i}`} className="flex items-start gap-3">
            <Skeleton className="h-3 w-3 rounded-full mt-1.5 shrink-0" />
            <div className="space-y-1 flex-1">
              <Skeleton className="h-4 w-32" />
              <Skeleton className="h-3 w-64" />
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-48">
        <div className="text-center">
          <IconAlertTriangle className="h-8 w-8 text-red-500 mx-auto" />
          <p className="mt-2 text-sm text-destructive">{error}</p>
          <Button
            variant="outline"
            size="sm"
            onClick={() => fetchEvents(true)}
            className="mt-2"
          >
            <IconRefresh className="mr-2 h-4 w-4" />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  if (events.length === 0) {
    return (
      <div className="text-center text-muted-foreground py-12">
        <p className="text-sm">No events recorded for this game.</p>
        <p className="text-xs mt-1 opacity-60">
          Events are captured from when the game engine runs.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between pb-2">
        <span className="text-xs text-muted-foreground">
          {events.length} events · times relative to game start
        </span>
        <Button
          variant="outline"
          size="sm"
          onClick={() => fetchEvents(true)}
          disabled={refreshing}
        >
          <IconRefresh
            className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`}
          />
        </Button>
      </div>

      <div className="relative">
        {/* Vertical line */}
        <div className="absolute left-[5px] top-2 bottom-2 w-px bg-border" />

        <div className="space-y-0">
          {events.map((event, i) => {
            const style = getEventStyle(event.event_type);
            const detail = formatEventData(event.event_type, event.data ?? {});
            const relTime = formatTime(event.occurred_at, reference);
            const absTime = formatAbsTime(event.occurred_at);

            return (
              <div
                key={event.id ?? i}
                className="flex items-start gap-3 py-2 group"
              >
                {/* Dot */}
                <div
                  className={`w-2.5 h-2.5 rounded-full shrink-0 mt-1.5 relative z-10 ${style.color}`}
                />

                {/* Content */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="text-sm font-medium">{style.label}</span>
                    {detail && (
                      <span className="text-xs text-muted-foreground truncate max-w-xs">
                        {detail}
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-2 mt-0.5">
                    <span className="text-xs text-muted-foreground font-mono">
                      {absTime}
                    </span>
                    <Badge
                      variant="outline"
                      className="text-[10px] px-1 py-0 h-4 font-mono"
                    >
                      {relTime}
                    </Badge>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
