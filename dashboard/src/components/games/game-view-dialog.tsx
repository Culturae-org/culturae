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
import { Skeleton } from "@/components/ui/skeleton";
import { useIsMobile } from "@/hooks/useMobile";
import { useViewDialog } from "@/hooks/useViewDialog";
import { apiGet } from "@/lib/api-client";
import { GAMES_ENDPOINTS } from "@/lib/api/endpoints";
import type { AdminGame, GameDetail } from "@/lib/types/games.types";
import {
  IconCalendar,
  IconCheck,
  IconCopy,
  IconEye,
  IconFileText,
  IconUsers,
} from "@tabler/icons-react";
import * as React from "react";
import { Link } from "react-router";

// ─── Helpers ────────────────────────────────────────────────────────────────

function formatDate(dateString: string) {
  return new Date(dateString).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function formatDuration(start: string, end: string | null) {
  if (!end) return "Ongoing";
  const ms = new Date(end).getTime() - new Date(start).getTime();
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

const VARIANT_LABELS: Record<string, string> = {
  flag_to_name: "Flag → Name",
  name_to_flag: "Name → Flag",
  flag_to_capital: "Flag → Capital",
  capital_to_flag: "Capital → Flag",
};

// ─── Main Component ─────────────────────────────────────────────────────────

interface GameViewDialogProps {
  game: AdminGame;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export function GameViewDialog({
  game,
  open: controlledOpen,
  onOpenChange,
}: GameViewDialogProps) {
  const isMobile = useIsMobile();
  const viewDialog = useViewDialog("view");
  const open =
    controlledOpen !== undefined ? controlledOpen : viewDialog.isOpen;
  const handleOpenChange = (isOpen: boolean) => {
    if (onOpenChange) {
      onOpenChange(isOpen);
    } else if (!isOpen) {
      viewDialog.close();
    }
  };

  const [loading, setLoading] = React.useState(false);
  const [detail, setDetail] = React.useState<GameDetail | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [copiedField, setCopiedField] = React.useState<string | null>(null);

  const copyToClipboard = async (text: string, field: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedField(field);
      setTimeout(() => setCopiedField(null), 2000);
    } catch (err) {
      console.error("Failed to copy text: ", err);
    }
  };

  const fetchDetails = React.useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const gameRes = await apiGet(GAMES_ENDPOINTS.GET(game.id));
      if (!gameRes.ok) throw new Error("Failed to fetch game details");
      const gameJson = await gameRes.json();
      const gameData: GameDetail = gameJson.data ?? gameJson;
      setDetail(gameData);
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setLoading(false);
    }
  }, [game.id]);

  React.useEffect(() => {
    if (open) fetchDetails();
  }, [open, fetchDetails]);

  const _isSolo = detail?.mode === "solo";

  return (
    <Drawer
      direction={isMobile ? "bottom" : "right"}
      open={open}
      onOpenChange={handleOpenChange}
    >
      {controlledOpen === undefined && (
        <DrawerTrigger asChild>
          <Button variant="ghost" size="sm">
            <IconEye className="h-4 w-4" />
          </Button>
        </DrawerTrigger>
      )}
      <DrawerContent>
        <DrawerHeader className="gap-1">
          <DrawerTitle>Game Details</DrawerTitle>
          <DrawerDescription>
            <span className="font-mono">
              {detail?.public_id ?? `${game.id.slice(0, 8)}...`}
            </span>
          </DrawerDescription>
        </DrawerHeader>

        <div className="flex flex-col gap-4 overflow-y-auto px-4 pb-4">
          {loading ? (
            <div className="space-y-6">
              <div className="space-y-3">
                <Skeleton className="h-4 w-36" />
                <div className="space-y-2 pl-6">
                  <Skeleton className="h-5 w-full" />
                  <Skeleton className="h-5 w-3/4" />
                </div>
              </div>
              <Skeleton className="h-px w-full" />
              <div className="space-y-3">
                <Skeleton className="h-4 w-32" />
                <div className="space-y-2 pl-6">
                  <Skeleton className="h-5 w-1/2" />
                  <Skeleton className="h-5 w-2/3" />
                  <Skeleton className="h-5 w-1/2" />
                </div>
              </div>
              <Skeleton className="h-px w-full" />
              <div className="space-y-3">
                <Skeleton className="h-4 w-24" />
                <div className="space-y-2 pl-6">
                  <Skeleton className="h-5 w-48" />
                  <Skeleton className="h-5 w-48" />
                </div>
              </div>
            </div>
          ) : error ? (
            <div className="flex items-center justify-center py-8 text-destructive text-sm">
              {error}
            </div>
          ) : detail ? (
            <>
              {/* ── Game Info ── */}
              <div className="space-y-4">
                <h4 className="font-medium flex items-center gap-2">
                  <IconFileText className="h-4 w-4" />
                  Game Information
                </h4>
                <div className="space-y-3 pl-6">
                  <div className="flex flex-col gap-1">
                    <Label className="text-xs text-muted-foreground">
                      Status
                    </Label>
                    <div className="text-sm flex gap-2 flex-wrap">
                      <Badge variant="outline" className="capitalize">
                        {detail.mode}
                      </Badge>
                      <Badge variant="outline" className="capitalize">
                        {detail.status.replace("_", " ")}
                      </Badge>
                      {detail.category && (
                        <Badge variant="outline" className="capitalize">
                          {detail.category}
                        </Badge>
                      )}
                      {detail.flag_variant && (
                        <Badge variant="outline" className="capitalize">
                          {VARIANT_LABELS[detail.flag_variant] ||
                            detail.flag_variant}
                        </Badge>
                      )}
                      {detail.language && (
                        <Badge
                          variant="outline"
                          className="uppercase text-[10px]"
                        >
                          {detail.language}
                        </Badge>
                      )}
                    </div>
                  </div>
                  <div className="flex flex-col gap-1">
                    <Label className="text-xs text-muted-foreground">
                      Game ID
                    </Label>
                    <div className="text-sm flex items-center gap-2">
                      <div className="text-xs font-mono bg-muted px-2 py-1 rounded flex-1 truncate">
                        {game.id}
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => copyToClipboard(game.id, "gameId")}
                        className="h-6 w-6 p-0"
                      >
                        {copiedField === "gameId" ? (
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

              {/* ── Overview ── */}
              <div className="space-y-4">
                <h4 className="font-medium flex items-center gap-2">
                  <IconUsers className="h-4 w-4" />
                  Game Overview
                </h4>
                <div className="space-y-3 pl-6">
                  <div className="flex flex-col gap-1">
                    <Label className="text-xs text-muted-foreground">
                      Players
                    </Label>
                    <div className="text-sm font-medium">
                      {detail.players.length}
                    </div>
                  </div>
                  <div className="flex flex-col gap-1">
                    <Label className="text-xs text-muted-foreground">
                      Duration
                    </Label>
                    <div className="text-sm font-medium">
                      {detail.started_at
                        ? formatDuration(detail.started_at, detail.completed_at)
                        : "Not started"}
                    </div>
                  </div>
                  {detail.time_bonus !== undefined && (
                    <div className="flex flex-col gap-1">
                      <Label className="text-xs text-muted-foreground">
                        Score Mode
                      </Label>
                      <div className="text-sm font-medium">
                        {detail.time_bonus ? "Time bonus" : "Classic"}
                      </div>
                    </div>
                  )}
                  {detail.template_id && (
                    <div className="flex flex-col gap-1">
                      <Label className="text-xs text-muted-foreground">
                        Template
                      </Label>
                      <div className="text-sm flex items-center gap-2">
                        <Link
                          to={`/game-templates?view=${detail.template_id}`}
                          className="text-xs font-mono hover:underline opacity-70"
                        >
                          {detail.template_id}
                        </Link>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 w-6 p-0"
                          onClick={() => {
                            const templateId = detail.template_id;
                            if (templateId)
                              copyToClipboard(templateId, "templateId");
                          }}
                        >
                          {copiedField === "templateId" ? (
                            <IconCheck className="h-3 w-3 text-green-600" />
                          ) : (
                            <IconCopy className="h-3 w-3" />
                          )}
                        </Button>
                      </div>
                    </div>
                  )}
                </div>
              </div>

              <Separator />

              {/* ── Timeline ── */}
              <div className="space-y-4">
                <h4 className="font-medium flex items-center gap-2">
                  <IconCalendar className="h-4 w-4" />
                  Timeline
                </h4>
                <div className="space-y-3 pl-6">
                  <div className="flex flex-col gap-1">
                    <Label className="text-xs text-muted-foreground">
                      Created
                    </Label>
                    <div className="text-sm">
                      {formatDate(detail.created_at)}
                    </div>
                  </div>
                  {detail.started_at && (
                    <div className="flex flex-col gap-1">
                      <Label className="text-xs text-muted-foreground">
                        Started
                      </Label>
                      <div className="text-sm">
                        {formatDate(detail.started_at)}
                      </div>
                    </div>
                  )}
                  {detail.completed_at && (
                    <div className="flex flex-col gap-1">
                      <Label className="text-xs text-muted-foreground">
                        Completed
                      </Label>
                      <div className="text-sm">
                        {formatDate(detail.completed_at)}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </>
          ) : null}
        </div>

        <DrawerFooter className="flex flex-row gap-2">
          <DrawerClose asChild>
            <Button variant="outline" className="flex-1">
              Close
            </Button>
          </DrawerClose>
          {detail && (
            <Button variant="outline" asChild className="flex-1">
              <Link to={`/games/${game.id}`}>Full Details</Link>
            </Button>
          )}
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  );
}
