"use client";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { DataTablePagination } from "@/components/ui/data-table-pagination";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useGames } from "@/hooks/useGames";
import { useIsMobile } from "@/hooks/useMobile";
import type { AdminGame } from "@/lib/types/games.types";
import {
  IconArchive,
  IconColumns,
  IconDotsVertical,
  IconEye,
  IconFilter,
  IconPlayerStop,
  IconRefresh,
  IconSearch,
  IconX,
} from "@tabler/icons-react";
import {
  type ColumnDef,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import * as React from "react";
import { useNavigate } from "react-router";
import { GameArchiveDialog } from "./game-archive-dialog";
import { GameCancelDialog } from "./game-cancel-dialog";

function GameIdCell({ id }: { id: string }) {
  const [isExpanded, setIsExpanded] = React.useState(false);
  return (
    <button
      type="button"
      className="font-mono text-xs text-muted-foreground cursor-pointer hover:text-foreground transition-colors bg-transparent border-none p-0 text-left"
      onClick={(e) => {
        e.stopPropagation();
        setIsExpanded(!isExpanded);
      }}
      title={isExpanded ? "Click to collapse" : "Click to expand"}
    >
      {isExpanded ? id : `${id.slice(0, 8)}...`}
    </button>
  );
}

function GameActionsCell({
  game,
  updateGameInList,
  onGameDeleted,
}: {
  game: AdminGame;
  updateGameInList: (g: AdminGame) => void;
  onGameDeleted: (id: string) => void;
}) {
  const navigate = useNavigate();
  const [cancelOpen, setCancelOpen] = React.useState(false);
  const [archiveOpen, setArchiveOpen] = React.useState(false);
  const isArchived = !!game.deleted_at;
  const canCancel =
    game.status === "in_progress" ||
    game.status === "active" ||
    game.status === "waiting" ||
    game.status === "ready";

  const handleArchive = (updatedGame: AdminGame) => {
    updateGameInList(updatedGame);
  };

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" className="h-8 w-8 p-0">
            <span className="sr-only">Open menu</span>
            <IconDotsVertical className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuLabel>Actions</DropdownMenuLabel>
          <DropdownMenuItem onSelect={() => navigate(`?view=${game.id}`)}>
            <IconEye className="mr-2 h-4 w-4" />
            View Game
          </DropdownMenuItem>
          {canCancel && (
            <DropdownMenuItem
              onSelect={() => setCancelOpen(true)}
              className="text-destructive"
            >
              <IconPlayerStop className="mr-2 h-4 w-4" />
              Cancel Game
            </DropdownMenuItem>
          )}
          <DropdownMenuItem
            onSelect={() => setArchiveOpen(true)}
            className={isArchived ? "text-green-600" : "text-destructive"}
          >
            <IconArchive className="mr-2 h-4 w-4" />
            {isArchived ? "Unarchive Game" : "Archive Game"}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      {canCancel && (
        <GameCancelDialog
          game={game}
          onGameUpdated={updateGameInList}
          open={cancelOpen}
          onOpenChange={setCancelOpen}
        />
      )}
      <GameArchiveDialog
        game={game}
        isArchived={isArchived}
        onGameArchived={handleArchive}
        open={archiveOpen}
        onOpenChange={setArchiveOpen}
      />
    </>
  );
}

const STATUS_COLORS: Record<string, string> = {
  waiting:
    "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300",
  ready: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300",
  in_progress:
    "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
  active: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
  completed:
    "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300",
  cancelled: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
  abandoned: "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300",
};

export function GamesDataTable() {
  const {
    games,
    loading,
    refreshing,
    error,
    currentPage,
    totalPages,
    totalCount,
    currentLimit,
    filters,
    goToPage,
    setPageSize,
    refresh,
    setFilter,
    clearFilters,
    updateGameInList,
    removeGame,
  } = useGames();

  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({
      public_id: false,
      category: false,
      language: false,
      question_count: false,
    });
  const [removingIds, setRemovingIds] = React.useState<Set<string>>(new Set());

  const navigate = useNavigate();
  const isMobile = useIsMobile();

  const handleGameDeleted = React.useCallback(
    (gameId: string) => {
      setRemovingIds(new Set([gameId]));
      setTimeout(() => {
        removeGame(gameId);
        setRemovingIds(new Set());
      }, 300);
    },
    [removeGame],
  );

  const columns: ColumnDef<AdminGame>[] = React.useMemo(
    () => [
      {
        accessorKey: "id",
        header: "Game ID",
        cell: ({ row }) => <GameIdCell id={row.getValue("id")} />,
      },
      {
        accessorKey: "mode",
        header: "Mode",
        cell: ({ row }) => (
          <Badge variant="outline" className="capitalize">
            {row.getValue("mode") as string}
          </Badge>
        ),
      },
      {
        accessorKey: "status",
        header: "Status",
        cell: ({ row }) => {
          const s = (row.getValue("status") as string) || "unknown";
          return (
            <Badge variant="outline" className={STATUS_COLORS[s] || ""}>
              {s.replace("_", " ").charAt(0).toUpperCase() +
                s.replace("_", " ").slice(1)}
            </Badge>
          );
        },
      },
      {
        id: "player",
        header: "Player",
        cell: ({ row }) => {
          const players = row.original.players || [];
          const names = players.map(
            (p) => p.user?.username || p.username || "Unknown",
          );
          return (
            <div className="text-sm font-medium">
              {names.length > 0 ? names.join(", ") : "—"}
            </div>
          );
        },
        enableHiding: false,
      },
      {
        id: "players_count",
        header: "Players",
        cell: ({ row }) => {
          const game = row.original;
          if (game.mode === "solo")
            return <div className="text-sm text-muted-foreground">1</div>;
          const current = game.current_players ?? game.players?.length ?? 0;
          const max = game.max_players ? `/${game.max_players}` : "";
          return (
            <div className="text-sm text-muted-foreground">
              {current}
              {max}
            </div>
          );
        },
      },
      {
        accessorKey: "public_id",
        header: "Public ID",
        cell: ({ row }) => {
          const val = row.getValue("public_id") as string | undefined;
          return val ? (
            <GameIdCell id={val} />
          ) : (
            <div className="text-sm font-mono text-muted-foreground">—</div>
          );
        },
      },
      {
        accessorKey: "category",
        header: "Category",
        cell: ({ row }) => (
          <div className="text-sm capitalize text-muted-foreground">
            {row.getValue("category") || "—"}
          </div>
        ),
      },
      {
        accessorKey: "language",
        header: "Language",
        cell: ({ row }) => (
          <div className="text-sm uppercase text-muted-foreground">
            {row.getValue("language") || "—"}
          </div>
        ),
      },
      {
        accessorKey: "question_count",
        header: "Questions",
        cell: ({ row }) => (
          <div className="text-sm text-muted-foreground">
            {row.getValue("question_count") || "—"}
          </div>
        ),
      },
      {
        accessorKey: "created_at",
        header: "Created",
        cell: ({ row }) => (
          <div className="text-sm">
            {new Date(row.getValue("created_at")).toLocaleDateString("en-US", {
              month: "short",
              day: "numeric",
              hour: "2-digit",
              minute: "2-digit",
            })}
          </div>
        ),
      },
      {
        accessorKey: "started_at",
        header: "Started",
        cell: ({ row }) => {
          const startedAt = row.getValue("started_at") as string | null;
          if (!startedAt)
            return <div className="text-sm text-muted-foreground">-</div>;
          return (
            <div className="text-sm">
              {new Date(startedAt).toLocaleDateString("en-US", {
                month: "short",
                day: "numeric",
                hour: "2-digit",
                minute: "2-digit",
              })}
            </div>
          );
        },
      },
      {
        id: "actions",
        header: "Actions",
        enableHiding: false,
        cell: ({ row }) => (
          <GameActionsCell
            game={row.original}
            updateGameInList={updateGameInList}
            onGameDeleted={handleGameDeleted}
          />
        ),
      },
    ],
    [updateGameInList, handleGameDeleted],
  );

  const table = useReactTable({
    data: games,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    manualPagination: true,
    pageCount: totalPages,
    state: { columnVisibility },
  });

  const hasActiveFilters =
    filters.status !== "" || filters.mode !== "" || filters.archived !== "";

  if (loading && !refreshing && games.length === 0) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <Skeleton className="h-6 w-32" />
              <Skeleton className="h-4 w-48 mt-1" />
            </div>
            <Skeleton className="h-9 w-32" />
          </div>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border">
            <div className="space-y-3 p-4">
              {[0, 1, 2, 3, 4, 5, 6, 7].map((i) => (
                <Skeleton key={`game-skel-${i}`} className="h-10 w-full" />
              ))}
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="border-0 dark:border">
        <CardContent className="pt-6">
          <div className="flex items-center justify-center h-32">
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <CardTitle>Games</CardTitle>
            <CardDescription>Manage and view game sessions</CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
          <div className="relative flex-1 max-w-sm w-full lg:w-auto">
            <IconSearch className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search by Game ID or Public ID..."
              value={filters.search}
              onChange={(e) => setFilter("search", e.target.value)}
              className="pl-10 pr-10"
            />
            {filters.search && (
              <IconX
                className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground cursor-pointer"
                onClick={() => setFilter("search", "")}
              />
            )}
          </div>
          <div className="flex flex-wrap gap-2 w-full lg:w-auto lg:ml-auto">
            <Button
              variant="outline"
              size="sm"
              onClick={refresh}
              disabled={loading || refreshing}
            >
              <IconRefresh
                className={`h-4 w-4 ${loading || refreshing ? "animate-spin" : ""}`}
              />
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <IconFilter className="h-4 w-4" />
                  Filter
                  {hasActiveFilters && (
                    <Badge
                      variant="secondary"
                      className="ml-1 h-5 px-1 text-xs"
                    >
                      {(filters.status ? 1 : 0) +
                        (filters.mode ? 1 : 0) +
                        (filters.archived ? 1 : 0)}
                    </Badge>
                  )}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <div className="px-2 py-1.5 text-sm font-semibold">
                  Archived
                </div>
                {[
                  { value: "", label: "Active" },
                  { value: "false", label: "Non-archived" },
                  { value: "true", label: "Archived" },
                ].map((a) => (
                  <DropdownMenuItem
                    key={a.value}
                    onSelect={(e) => e.preventDefault()}
                  >
                    <Checkbox
                      checked={filters.archived === a.value}
                      onCheckedChange={(checked) =>
                        setFilter("archived", checked ? a.value : "")
                      }
                      className="mr-2"
                    />
                    {a.label}
                  </DropdownMenuItem>
                ))}
                <DropdownMenuSeparator />
                <div className="px-2 py-1.5 text-sm font-semibold">Status</div>
                {["active", "completed", "abandoned", "cancelled"].map((s) => (
                  <DropdownMenuItem
                    key={s}
                    onSelect={(e) => e.preventDefault()}
                  >
                    <Checkbox
                      checked={filters.status === s}
                      onCheckedChange={(checked) =>
                        setFilter("status", checked ? s : "")
                      }
                      className="mr-2"
                    />
                    {s.charAt(0).toUpperCase() + s.slice(1)}
                  </DropdownMenuItem>
                ))}
                <DropdownMenuSeparator />
                <div className="px-2 py-1.5 text-sm font-semibold">Mode</div>
                {["solo", "1v1", "tournament", "team"].map((m) => (
                  <DropdownMenuItem
                    key={m}
                    onSelect={(e) => e.preventDefault()}
                  >
                    <Checkbox
                      checked={filters.mode === m}
                      onCheckedChange={(checked) =>
                        setFilter("mode", checked ? m : "")
                      }
                      className="mr-2"
                    />
                    {m === "1v1"
                      ? "1v1"
                      : m.charAt(0).toUpperCase() + m.slice(1)}
                  </DropdownMenuItem>
                ))}
                {hasActiveFilters && (
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onSelect={clearFilters}>
                      Clear all filters
                    </DropdownMenuItem>
                  </>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <IconColumns className="h-4 w-4" />
                  Columns
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {table
                  .getAllColumns()
                  .filter((c) => c.getCanHide())
                  .map((c) => (
                    <DropdownMenuCheckboxItem
                      key={c.id}
                      className="capitalize"
                      checked={c.getIsVisible()}
                      onCheckedChange={(value) => c.toggleVisibility(!!value)}
                    >
                      {c.id.replace("_", " ")}
                    </DropdownMenuCheckboxItem>
                  ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {isMobile ? (
          <div className="space-y-3">
            {games.length ? (
              games.map((game) => (
                <button
                  type="button"
                  key={game.id}
                  onClick={(e) => {
                    const target = e.target as HTMLElement;
                    if (
                      target.closest("button") ||
                      target.closest(".checkbox") ||
                      target.closest('[role="menu"]')
                    )
                      return;
                    navigate(`?view=${game.id}`);
                  }}
                  className={`w-full text-left cursor-pointer rounded-lg border bg-card p-4 space-y-3 transition-all duration-300 hover:bg-muted/50 ${removingIds.has(game.id) ? "opacity-0 scale-y-0 h-0" : ""}`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className="capitalize">
                        {game.mode}
                      </Badge>
                      <Badge
                        variant="outline"
                        className={STATUS_COLORS[game.status] || ""}
                      >
                        {game.status.replace("_", " ").charAt(0).toUpperCase() +
                          game.status.replace("_", " ").slice(1)}
                      </Badge>
                    </div>
                    <GameActionsCell
                      game={game}
                      updateGameInList={updateGameInList}
                      onGameDeleted={handleGameDeleted}
                    />
                  </div>
                  <div className="flex flex-col gap-1.5 text-sm mt-2 pt-2 border-t text-muted-foreground">
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-foreground">
                        {game.players
                          ?.map((p) => p.user?.username || p.username)
                          .join(", ") || "—"}
                      </span>
                      <span className="text-xs">
                        {new Date(game.created_at).toLocaleDateString("en-US", {
                          month: "short",
                          day: "numeric",
                          hour: "2-digit",
                          minute: "2-digit",
                        })}
                      </span>
                    </div>
                    <div className="text-xs flex flex-wrap gap-x-3 gap-y-1">
                      <span>
                        {game.mode === "solo"
                          ? "1 Player"
                          : `Players: ${game.current_players ?? game.players?.length ?? 0}${game.max_players ? `/${game.max_players}` : ""}`}
                      </span>
                      {game.category && (
                        <span>
                          • <span className="capitalize">{game.category}</span>
                        </span>
                      )}
                      {game.question_count && (
                        <span>• {game.question_count} Qs</span>
                      )}
                    </div>
                  </div>
                </button>
              ))
            ) : (
              <div className="rounded-lg border bg-card p-8 text-center text-muted-foreground">
                No games found.
              </div>
            )}
          </div>
        ) : (
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                {table.getHeaderGroups().map((hg) => (
                  <TableRow key={hg.id}>
                    {hg.headers.map((h) => (
                      <TableHead key={h.id}>
                        {h.isPlaceholder
                          ? null
                          : flexRender(
                              h.column.columnDef.header,
                              h.getContext(),
                            )}
                      </TableHead>
                    ))}
                  </TableRow>
                ))}
              </TableHeader>
              <TableBody>
                {table.getRowModel().rows?.length ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow
                      key={row.id}
                      onClick={(e) => {
                        const target = e.target as HTMLElement;
                        if (
                          target.closest("button") ||
                          target.closest(".checkbox") ||
                          target.closest('[role="menu"]')
                        )
                          return;
                        navigate(`?view=${row.original.id}`);
                      }}
                      className={`cursor-pointer transition-all duration-300 ${removingIds.has(row.original.id) ? "opacity-0 scale-y-0 h-0" : ""}`}
                    >
                      {row.getVisibleCells().map((cell) => (
                        <TableCell key={cell.id}>
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext(),
                          )}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell
                      colSpan={columns.length}
                      className="h-24 text-center"
                    >
                      No games found.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        )}

        <DataTablePagination
          pageIndex={currentPage - 1}
          pageCount={totalPages}
          pageSize={currentLimit}
          totalCount={totalCount}
          onPageChange={(page) => goToPage(page + 1)}
          onPageSizeChange={setPageSize}
          loading={loading}
        />
      </CardContent>
    </Card>
  );
}
