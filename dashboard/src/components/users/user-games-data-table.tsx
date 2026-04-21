"use client";

import { GameViewDialog } from "@/components/games/game-view-dialog";
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
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { type EnrichedGame, useUserGames } from "@/hooks/useGames";
import { useIsMobile } from "@/hooks/useMobile";
import type { AdminGame } from "@/lib/types/games.types";
import {
  IconAlertTriangle,
  IconCalendar,
  IconColumns,
  IconDotsVertical,
  IconExternalLink,
  IconEye,
  IconFilter,
  IconRefresh,
  IconTrophy,
} from "@tabler/icons-react";
import {
  type ColumnDef,
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { format, formatDistanceToNow } from "date-fns";
import { enUS } from "date-fns/locale";
import * as React from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router";

interface UserGamesDataTableProps {
  userId: string;
}

function GameIdCell({ id }: { id: string | undefined }) {
  const [isExpanded, setIsExpanded] = React.useState(false);
  if (!id) return <span className="text-muted-foreground">-</span>;
  return (
    <button
      type="button"
      className="font-mono text-xs text-muted-foreground cursor-pointer hover:text-foreground transition-colors text-left"
      onClick={() => setIsExpanded(!isExpanded)}
      title={isExpanded ? "Click to collapse" : "Click to expand"}
    >
      {isExpanded ? id : `${id.slice(0, 8)}...`}
    </button>
  );
}

function GameStatusCell({ status }: { status: string | undefined }) {
  const statusColors: Record<string, string> = {
    waiting:
      "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300",
    ready: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300",
    in_progress:
      "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300",
    completed:
      "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
    cancelled: "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300",
    abandoned: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
  };
  if (!status) return <Badge variant="outline">Unknown</Badge>;
  return (
    <Badge variant="outline" className={statusColors[status] || ""}>
      {status.replace("_", " ")}
    </Badge>
  );
}

function GameModeCell({ mode }: { mode: string | undefined }) {
  const modeColors: Record<string, string> = {
    solo: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300",
    "1v1":
      "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300",
    tournament:
      "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300",
    team: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
  };
  if (!mode) return <Badge variant="outline">Unknown</Badge>;
  return (
    <Badge variant="outline" className={modeColors[mode] || ""}>
      {mode}
    </Badge>
  );
}

export function UserGamesDataTable({ userId }: UserGamesDataTableProps) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const pathname = useLocation().pathname;
  const {
    games,
    loading,
    refreshing,
    error,
    filters,
    refresh,
    setFilter,
    clearFilters,
  } = useUserGames(userId);
  const isMobile = useIsMobile();

  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({
      public_id: false,
      dataset_id: false,
    });
  const [rowSelection, setRowSelection] = React.useState({});
  const [currentPage, setCurrentPage] = React.useState(1);
  const [currentLimit, setCurrentLimit] = React.useState(10);
  const [viewGame, setViewGame] = React.useState<EnrichedGame | null>(null);
  const [viewOpen, setViewOpen] = React.useState(false);
  const [filterDropdownOpen, setFilterDropdownOpen] = React.useState(false);

  const viewId = searchParams.get("gameView");
  React.useEffect(() => {
    if (viewId && games.length > 0) {
      const found = games.find((g) => g.id === viewId);
      if (found) {
        setViewGame(found);
        setViewOpen(true);
      }
    }
  }, [viewId, games]);

  const handleView = React.useCallback(
    (game: EnrichedGame) => {
      setViewGame(game);
      setViewOpen(true);
      const params = new URLSearchParams(searchParams.toString());
      params.set("gameView", game.id);
      navigate(`${pathname}?${params.toString()}`);
    },
    [searchParams, pathname, navigate],
  );

  const columns: ColumnDef<EnrichedGame>[] = React.useMemo(
    () => [
      {
        accessorKey: "id",
        header: "ID",
        cell: ({ row }) => <GameIdCell id={row.getValue("id")} />,
      },
      {
        accessorKey: "public_id",
        header: "Public ID",
        cell: ({ row }) => <GameIdCell id={row.getValue("public_id")} />,
      },
      {
        accessorKey: "mode",
        header: "Mode",
        cell: ({ row }) => <GameModeCell mode={row.getValue("mode")} />,
      },
      {
        accessorKey: "status",
        header: "Status",
        cell: ({ row }) => <GameStatusCell status={row.getValue("status")} />,
      },
      {
        accessorKey: "question_count",
        header: "Questions",
        cell: ({ row }) => (
          <div className="text-sm">{row.getValue("question_count")}</div>
        ),
      },
      {
        id: "score",
        header: "Your Score",
        cell: ({ row }) => {
          const game = row.original;
          return (
            <div
              className={`font-medium ${game._is_winner ? "text-green-600" : ""}`}
            >
              {game._user_score}
              {game._is_winner && (
                <IconTrophy className="inline h-3 w-3 ml-1 text-yellow-500" />
              )}
            </div>
          );
        },
      },
      {
        accessorKey: "created_at",
        header: "Date",
        cell: ({ row }) => {
          const dateValue = row.getValue("created_at") as string;
          if (
            !dateValue ||
            dateValue === "0001-01-01T00:00:00Z" ||
            Number.isNaN(new Date(dateValue).getTime())
          ) {
            return (
              <div className="flex items-center gap-1 text-sm text-muted-foreground">
                <IconCalendar className="h-4 w-4" />
                <span>-</span>
              </div>
            );
          }
          const date = new Date(dateValue);
          return (
            <div className="flex items-center gap-1 text-sm">
              <IconCalendar className="h-4 w-4 text-muted-foreground" />
              <span title={format(date, "PPpp", { locale: enUS })}>
                {formatDistanceToNow(date, { addSuffix: true, locale: enUS })}
              </span>
            </div>
          );
        },
      },
      {
        accessorKey: "dataset_id",
        header: "Dataset",
        cell: ({ row }) => (
          <div className="text-xs text-muted-foreground truncate max-w-[150px]">
            {row.getValue("dataset_id") || "-"}
          </div>
        ),
      },
      {
        id: "actions",
        header: "Actions",
        enableHiding: false,
        cell: ({ row }) => {
          const game = row.original;
          return (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="h-8 w-8 p-0">
                  <span className="sr-only">Open menu</span>
                  <IconDotsVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>Actions</DropdownMenuLabel>
                <DropdownMenuItem onSelect={() => handleView(game)}>
                  <IconEye className="mr-2 h-4 w-4" />
                  View Details
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={() => navigate(`/games/${game.id}`)}>
                  <IconExternalLink className="mr-2 h-4 w-4" />
                  Open Game Page
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => navigator.clipboard.writeText(game.id)}
                >
                  Copy Game ID
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          );
        },
      },
    ],
    [navigate, handleView],
  );

  const table = useReactTable({
    data: games,
    columns,
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    pageCount: Math.ceil(games.length / currentLimit),
    state: {
      sorting,
      columnVisibility,
      rowSelection,
      pagination: { pageIndex: currentPage - 1, pageSize: currentLimit },
    },
    onPaginationChange: (updater) => {
      const newPagination =
        typeof updater === "function"
          ? updater(table.getState().pagination)
          : updater;
      setCurrentPage(newPagination.pageIndex + 1);
      setCurrentLimit(newPagination.pageSize);
    },
  });

  if (loading && !refreshing && games.length === 0) {
    return (
      <div className="space-y-4 py-4">
        <div className="flex items-center justify-between">
          <Skeleton className="h-9 w-64" />
          <Skeleton className="h-9 w-32" />
        </div>
        <div className="rounded-md border">
          <div className="space-y-3 p-4">
            {[0, 1, 2, 3, 4, 5, 6, 7].map((i) => (
              <Skeleton key={`game-skel-${i}`} className="h-10 w-full" />
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <IconAlertTriangle className="h-8 w-8 text-red-500 mx-auto" />
          <p className="mt-2 text-sm text-destructive">{error}</p>
          <Button
            variant="outline"
            size="sm"
            onClick={refresh}
            className="mt-2"
          >
            <IconRefresh className="mr-2 h-4 w-4 text-current" />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle>User Games</CardTitle>
        <CardDescription>View and manage games for this user</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="relative">
          <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
            <div className="flex flex-wrap gap-2 w-full lg:w-auto lg:ml-auto">
              <Button
                variant="outline"
                size="sm"
                onClick={refresh}
                disabled={refreshing}
              >
                <IconRefresh
                  className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`}
                />
              </Button>
              <DropdownMenu open={filterDropdownOpen} onOpenChange={setFilterDropdownOpen}>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" size="sm">
                    <IconFilter className="mr-2 h-4 w-4" />
                    Filter
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-56">
                  <div className="px-2 py-1.5 text-sm font-semibold">
                    Status
                  </div>
                  {["waiting", "ready", "in_progress", "completed", "cancelled", "abandoned"].map(
                    (s) => (
                      <DropdownMenuItem
                        key={`status-${s}`}
                        onSelect={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                        }}
                        onClick={(e) => {
                          e.stopPropagation();
                          setFilter("status", filters.status === s ? "" : s);
                        }}
                      >
                        <Checkbox
                          checked={filters.status === s}
                          onCheckedChange={(checked) => {
                            setFilter("status", checked ? s : "");
                          }}
                          onClick={(e) => e.stopPropagation()}
                          className="mr-2"
                        />
                        <span className="capitalize">{s.replace("_", " ")}</span>
                      </DropdownMenuItem>
                    ),
                  )}
                  <DropdownMenuSeparator />
                  <div className="px-2 py-1.5 text-sm font-semibold">
                    Mode
                  </div>
                  {["solo", "1v1", "tournament", "team"].map((m) => (
                    <DropdownMenuItem
                      key={`mode-${m}`}
                      onSelect={(e) => {
                        e.preventDefault();
                        e.stopPropagation();
                      }}
                      onClick={(e) => {
                        e.stopPropagation();
                        setFilter("mode", filters.mode === m ? "" : m);
                      }}
                    >
                      <Checkbox
                        checked={filters.mode === m}
                        onCheckedChange={(checked) => {
                          setFilter("mode", checked ? m : "");
                        }}
                        onClick={(e) => e.stopPropagation()}
                        className="mr-2"
                      />
                      <span className="capitalize">{m}</span>
                    </DropdownMenuItem>
                  ))}
                  {(filters.status || filters.mode) && (
                    <>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem
                        onSelect={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                        }}
                        onClick={(e) => {
                          e.stopPropagation();
                          clearFilters();
                        }}
                      >
                        Clear Filters
                      </DropdownMenuItem>
                    </>
                  )}
                </DropdownMenuContent>
              </DropdownMenu>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" size="sm">
                    <IconColumns className="mr-2 h-4 w-4" />
                    Columns
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  {table
                    .getAllColumns()
                    .filter((column) => column.getCanHide())
                    .map((column) => (
                      <DropdownMenuCheckboxItem
                        key={column.id}
                        className="capitalize"
                        checked={column.getIsVisible()}
                        onCheckedChange={(value) =>
                          column.toggleVisibility(!!value)
                        }
                      >
                        {column.id}
                      </DropdownMenuCheckboxItem>
                    ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
        </div>

        {isMobile ? (
          <div className="space-y-3">
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => {
                const game = row.original;
                return (
                  <button
                    type="button"
                    key={row.id}
                    className="w-full text-left rounded-lg border bg-card p-4 space-y-3 cursor-pointer hover:bg-muted/50"
                    onClick={() => handleView(game)}
                  >
                    <div className="flex items-center justify-between">
                      <GameModeCell mode={game.mode} />
                      <GameStatusCell status={game.status} />
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-muted-foreground">
                        {formatDistanceToNow(new Date(game.created_at), {
                          addSuffix: true,
                          locale: enUS,
                        })}
                      </span>
                      <span
                        className={`font-medium ${game._is_winner ? "text-green-600" : ""}`}
                      >
                        Score: {game._user_score}
                        {game._is_winner && (
                          <IconTrophy className="inline h-3 w-3 ml-1 text-yellow-500" />
                        )}
                      </span>
                    </div>
                  </button>
                );
              })
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
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <TableHead key={header.id}>
                        {header.isPlaceholder
                          ? null
                          : flexRender(
                              header.column.columnDef.header,
                              header.getContext(),
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
                      data-state={row.getIsSelected() && "selected"}
                      className="cursor-pointer hover:bg-muted/50"
                      onClick={(e) => {
                        const target = e.target as HTMLElement;
                        if (
                          target.closest(
                            'button, a, [role="checkbox"], [role="menuitem"], [data-radix-collection-item]',
                          )
                        )
                          return;
                        handleView(row.original);
                      }}
                    >
                      {row.getVisibleCells().map((cell) => {
                        const isActionCell =
                          cell.column.id === "actions" ||
                          cell.column.id === "select";
                        return (
                          <TableCell
                            key={cell.id}
                            onClick={
                              isActionCell
                                ? (e) => e.stopPropagation()
                                : undefined
                            }
                          >
                            {flexRender(
                              cell.column.columnDef.cell,
                              cell.getContext(),
                            )}
                          </TableCell>
                        );
                      })}
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
          pageCount={table.getPageCount() || 1}
          pageSize={currentLimit}
          totalCount={games.length}
          onPageChange={(page) => table.setPageIndex(page)}
          onPageSizeChange={(newSize) => {
            setCurrentLimit(newSize);
            setCurrentPage(1);
          }}
        />

        {viewGame && (
          <GameViewDialog
            game={viewGame as unknown as AdminGame}
            open={viewOpen}
            onOpenChange={(open) => {
              setViewOpen(open);
              if (!open) {
                setViewGame(null);
                const params = new URLSearchParams(searchParams.toString());
                params.delete("gameView");
                const qs = params.toString();
                navigate(qs ? `${pathname}?${qs}` : pathname);
              }
            }}
          />
        )}
      </CardContent>
    </Card>
  );
}
