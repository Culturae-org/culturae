"use client";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { DataTablePagination } from "@/components/ui/data-table-pagination";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useUserFriends } from "@/hooks/useFriends";
import { useIsMobile } from "@/hooks/useMobile";
import type { Friendship } from "@/lib/types/friends.types";
import {
  IconAlertTriangle,
  IconCalendar,
  IconCheck,
  IconDotsVertical,
  IconExternalLink,
  IconRefresh,
  IconUsers,
} from "@tabler/icons-react";
import {
  type ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { formatDistanceToNow } from "date-fns";
import { enUS } from "date-fns/locale";
import * as React from "react";
import { Link, useLocation, useNavigate, useSearchParams } from "react-router";

interface UserFriendsDataTableProps {
  userId: string;
}

function FriendshipViewDialog({
  friendship,
  userId,
  open,
  onOpenChange,
}: {
  friendship: Friendship | null;
  userId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const navigate = useNavigate();

  if (!friendship) return null;

  const friendId =
    friendship.user1_id === userId ? friendship.user2_id : friendship.user1_id;
  const friendUsername =
    friendship.user1_id === userId
      ? friendship.user2_username
      : friendship.user1_username;

  const formatDate = (d: string) =>
    new Date(d).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <IconUsers className="h-5 w-5" />
            Friendship
          </DialogTitle>
          <DialogDescription>
            Friends since{" "}
            {formatDistanceToNow(new Date(friendship.created_at), {
              addSuffix: true,
              locale: enUS,
            })}
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <Label className="text-xs text-muted-foreground">User 1</Label>
            <div className="flex items-center gap-2">
              <div className="flex flex-col flex-1 min-w-0 bg-muted px-2 py-1.5 rounded">
                <span className="text-sm font-medium">
                  {friendship.user1_username || "—"}
                </span>
                <span className="font-mono text-xs text-muted-foreground break-all">
                  {friendship.user1_id}
                </span>
              </div>
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 shrink-0"
                asChild
              >
                <Link to={`/users/${friendship.user1_id}`}>
                  <IconExternalLink className="h-3.5 w-3.5" />
                </Link>
              </Button>
            </div>
          </div>

          <div className="flex flex-col gap-1">
            <Label className="text-xs text-muted-foreground">User 2</Label>
            <div className="flex items-center gap-2">
              <div className="flex flex-col flex-1 min-w-0 bg-muted px-2 py-1.5 rounded">
                <span className="text-sm font-medium">
                  {friendship.user2_username || "—"}
                </span>
                <span className="font-mono text-xs text-muted-foreground break-all">
                  {friendship.user2_id}
                </span>
              </div>
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 shrink-0"
                asChild
              >
                <Link to={`/users/${friendship.user2_id}`}>
                  <IconExternalLink className="h-3.5 w-3.5" />
                </Link>
              </Button>
            </div>
          </div>

          <Separator />

          <div className="flex flex-col gap-1">
            <Label className="text-xs text-muted-foreground">
              Friends Since
            </Label>
            <div className="text-sm">{formatDate(friendship.created_at)}</div>
          </div>
        </div>

        <div className="flex gap-2 mt-2">
          <Button
            variant="default"
            onClick={() => {
              onOpenChange(false);
              navigate(`/users/${friendId}`);
            }}
            className="flex-1"
          >
            <IconExternalLink className="h-4 w-4 mr-2" />
            View {friendUsername || "Friend"}
          </Button>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            className="flex-1"
          >
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

const DATE_PRESETS = [
  { label: "All time", days: 0 },
  { label: "Last 7 days", days: 7 },
  { label: "Last 30 days", days: 30 },
  { label: "Last 90 days", days: 90 },
  { label: "Last year", days: 365 },
];

export function UserFriendsDataTable({ userId }: UserFriendsDataTableProps) {
  const isMobile = useIsMobile();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const pathname = useLocation().pathname;

  const [pageIndex, setPageIndex] = React.useState(0);
  const [pageSize, setPageSize] = React.useState(10);
  const [datePreset, setDatePreset] = React.useState(0);
  const [viewFriendship, setViewFriendship] = React.useState<Friendship | null>(
    null,
  );
  const [viewOpen, setViewOpen] = React.useState(false);

  const { friends, total, loading, refreshing, error, refresh } =
    useUserFriends(userId, pageIndex + 1, pageSize);

  const viewId = searchParams.get("friendView");
  React.useEffect(() => {
    if (viewId && friends.length > 0) {
      const found = friends.find((f) => f.id === viewId);
      if (found) {
        setViewFriendship(found);
        setViewOpen(true);
      }
    }
  }, [viewId, friends]);

  const handleView = React.useCallback(
    (f: Friendship) => {
      setViewFriendship(f);
      setViewOpen(true);
      const params = new URLSearchParams(searchParams.toString());
      params.set("friendView", f.id);
      navigate(`${pathname}?${params.toString()}`);
    },
    [searchParams, pathname, navigate],
  );

  const filteredFriends = React.useMemo(() => {
    if (datePreset === 0) return friends;
    const cutoff = new Date();
    cutoff.setDate(cutoff.getDate() - datePreset);
    return friends.filter((f) => new Date(f.created_at) >= cutoff);
  }, [friends, datePreset]);

  const totalPages = Math.ceil(total / pageSize) || 1;

  const columns: ColumnDef<Friendship>[] = React.useMemo(
    () => [
      {
        id: "name",
        header: "Name",
        cell: ({ row }) => {
          const f = row.original;
          const friendId = f.user1_id === userId ? f.user2_id : f.user1_id;
          const friendUsername =
            f.user1_id === userId ? f.user2_username : f.user1_username;
          return (
            <Link
              to={`/users/${friendId}`}
              className="font-medium hover:underline text-sm"
              onClick={(e) => e.stopPropagation()}
            >
              {friendUsername || "—"}
            </Link>
          );
        },
      },
      {
        id: "friend_id",
        header: "ID",
        cell: ({ row }) => {
          const f = row.original;
          const friendId = f.user1_id === userId ? f.user2_id : f.user1_id;
          return (
            <span className="font-mono text-xs text-muted-foreground">
              {friendId}
            </span>
          );
        },
      },
      {
        accessorKey: "created_at",
        header: "Friends Since",
        cell: ({ row }) => (
          <div className="text-sm text-muted-foreground">
            {formatDistanceToNow(new Date(row.getValue("created_at")), {
              addSuffix: true,
              locale: enUS,
            })}
          </div>
        ),
      },
      {
        id: "actions",
        header: "Actions",
        enableHiding: false,
        cell: ({ row }) => {
          const f = row.original;
          const _friendId = f.user1_id === userId ? f.user2_id : f.user1_id;
          return (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="h-8 w-8 p-0">
                  <IconDotsVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>Actions</DropdownMenuLabel>
                <DropdownMenuItem onSelect={() => handleView(f)}>
                  View details
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          );
        },
      },
    ],
    [userId, handleView],
  );

  const table = useReactTable({
    data: filteredFriends,
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    pageCount: totalPages,
    state: { pagination: { pageIndex, pageSize } },
    onPaginationChange: (updater) => {
      const next =
        typeof updater === "function"
          ? updater({ pageIndex, pageSize })
          : updater;
      setPageIndex(next.pageIndex);
      setPageSize(next.pageSize);
    },
  });

  if (loading && !refreshing) {
    return (
      <div className="space-y-4 py-4">
        <div className="flex items-center justify-between">
          <Skeleton className="h-9 w-64" />
          <Skeleton className="h-9 w-32" />
        </div>
        <div className="rounded-md border">
          <div className="space-y-3 p-4">
            {[0, 1, 2, 3, 4, 5, 6, 7].map((i) => (
              <Skeleton key={`friend-skel-${i}`} className="h-10 w-full" />
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
        <CardTitle>User Friends</CardTitle>
        <CardDescription>View and manage friends for this user</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-wrap gap-2 py-4 justify-between items-center">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm">
                <IconCalendar className="h-4 w-4 mr-2" />
                {DATE_PRESETS.find((p) => p.days === datePreset)?.label ??
                  "All time"}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start">
              {DATE_PRESETS.map((p) => (
                <DropdownMenuItem
                  key={p.days}
                  onSelect={() => {
                    setDatePreset(p.days);
                    setPageIndex(0);
                  }}
                >
                  {datePreset === p.days && (
                    <IconCheck className="h-4 w-4 mr-2" />
                  )}
                  {datePreset !== p.days && <span className="w-6 mr-0" />}
                  {p.label}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
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
        </div>

        {isMobile ? (
          <div className="space-y-3">
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => {
                const f = row.original;
                const friendId =
                  f.user1_id === userId ? f.user2_id : f.user1_id;
                const friendUsername =
                  f.user1_id === userId ? f.user2_username : f.user1_username;
                return (
                  <button
                    type="button"
                    key={row.id}
                    className="w-full text-left rounded-lg border bg-card p-4 space-y-3 cursor-pointer hover:bg-muted/50"
                    onClick={() => handleView(f)}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Link
                          to={`/users/${friendId}`}
                          className="font-medium hover:underline"
                          onClick={(e) => e.stopPropagation()}
                        >
                          {friendUsername || friendId.slice(0, 8)}
                        </Link>
                        <span className="font-mono text-xs text-muted-foreground">
                          {friendId.slice(0, 8)}...
                        </span>
                      </div>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            variant="ghost"
                            className="h-8 w-8 p-0"
                            onClick={(e) => e.stopPropagation()}
                          >
                            <IconDotsVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuLabel>Actions</DropdownMenuLabel>
                          <DropdownMenuItem onSelect={() => handleView(f)}>
                            View details
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                    <div className="text-xs text-muted-foreground">
                      Friends since{" "}
                      {formatDistanceToNow(new Date(f.created_at), {
                        addSuffix: true,
                        locale: enUS,
                      })}
                    </div>
                  </button>
                );
              })
            ) : (
              <div className="rounded-lg border bg-card p-8 text-center text-muted-foreground">
                No friends yet
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
                {table.getRowModel().rows.length ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow
                      key={row.id}
                      className="cursor-pointer hover:bg-muted/50"
                      onClick={(e) => {
                        const target = e.target as HTMLElement;
                        if (target.closest("button, a")) return;
                        handleView(row.original);
                      }}
                    >
                      {row.getVisibleCells().map((cell) => (
                        <TableCell
                          key={cell.id}
                          onClick={
                            cell.column.id === "actions"
                              ? (e) => e.stopPropagation()
                              : undefined
                          }
                        >
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
                      className="h-24 text-center text-muted-foreground"
                    >
                      No friends yet
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        )}

        <DataTablePagination
          pageIndex={pageIndex}
          pageCount={table.getPageCount() || 1}
          pageSize={pageSize}
          totalCount={total}
          onPageChange={(page) => table.setPageIndex(page)}
          onPageSizeChange={(newSize) => table.setPageSize(newSize)}
        />

        <FriendshipViewDialog
          friendship={viewFriendship}
          userId={userId}
          open={viewOpen}
          onOpenChange={(open) => {
            setViewOpen(open);
            if (!open) {
              setViewFriendship(null);
              const params = new URLSearchParams(searchParams.toString());
              params.delete("friendView");
              const qs = params.toString();
              navigate(qs ? `${pathname}?${qs}` : pathname);
            }
          }}
        />
      </CardContent>
    </Card>
  );
}
