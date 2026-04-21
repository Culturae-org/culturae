"use client";

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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
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
import { useUserFriendRequests } from "@/hooks/useFriends";
import { useIsMobile } from "@/hooks/useMobile";
import type { FriendRequest } from "@/lib/types/friends.types";
import {
  IconAlertTriangle,
  IconColumns,
  IconDotsVertical,
  IconEye,
  IconFilter,
  IconRefresh,
  IconUserPlus,
} from "@tabler/icons-react";
import {
  type ColumnDef,
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { formatDistanceToNow } from "date-fns";
import { enUS } from "date-fns/locale";
import * as React from "react";
import { Link, useLocation, useNavigate, useSearchParams } from "react-router";

const STATUS_COLORS: Record<string, string> = {
  pending:
    "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300",
  accepted: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
  rejected: "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300",
  blocked: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
  cancelled: "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300",
};

function ShortIdCell({ id }: { id: string }) {
  const [expanded, setExpanded] = React.useState(false);
  return (
    <button
      type="button"
      className="font-mono text-xs text-muted-foreground cursor-pointer hover:text-foreground transition-colors bg-transparent border-none p-0 text-left"
      onClick={(e) => {
        e.stopPropagation();
        setExpanded(!expanded);
      }}
      title={expanded ? "Click to collapse" : "Click to expand"}
    >
      {expanded ? id : `${id.slice(0, 8)}...`}
    </button>
  );
}

function UserIdLink({
  id,
  username,
  userId,
}: { id: string; username: string; userId: string }) {
  const isSelf = id === userId;
  return (
    <div className="flex items-center gap-1.5">
      <Link
        to={`/users/${id}`}
        className="hover:underline text-sm font-medium"
        onClick={(e) => e.stopPropagation()}
      >
        {username || `${id.slice(0, 8)}...`}
      </Link>
      <span className="font-mono text-xs text-muted-foreground">
        {id.slice(0, 8)}...
      </span>
      {isSelf && (
        <Badge variant="secondary" className="text-[10px] h-4 px-1 py-0">
          you
        </Badge>
      )}
    </div>
  );
}

function FriendRequestViewDialog({
  request,
  open,
  onOpenChange,
}: {
  request: FriendRequest | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  if (!request) return null;

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
            <IconUserPlus className="h-5 w-5" />
            Friend Request
          </DialogTitle>
          <DialogDescription className="flex items-center gap-2">
            Request details —{" "}
            <Badge
              variant="outline"
              className={STATUS_COLORS[request.status] || ""}
            >
              {request.status}
            </Badge>
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-6">
          <div className="flex flex-col gap-3 p-4 bg-muted/50 rounded-lg">
            <div className="flex items-center gap-2">
              <Badge
                variant="outline"
                className={STATUS_COLORS[request.status] || ""}
              >
                {request.status}
              </Badge>
              <span className="text-sm text-muted-foreground">
                {formatDistanceToNow(new Date(request.created_at), {
                  addSuffix: true,
                  locale: enUS,
                })}
              </span>
            </div>
          </div>
          <div className="space-y-4">
            <h4 className="font-medium flex items-center gap-2">
              <IconUserPlus className="h-4 w-4" />
              Users
            </h4>
            <div className="space-y-3 pl-6">
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  From (Sender)
                </Label>
                <UserIdLink
                  id={request.from_user_id}
                  username={request.from_username || ""}
                  userId={request.from_user_id}
                />
              </div>
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  To (Receiver)
                </Label>
                <UserIdLink
                  id={request.to_user_id}
                  username={request.to_username || ""}
                  userId={request.to_user_id}
                />
              </div>
            </div>
          </div>
          <Separator />
          <div className="space-y-4">
            <h4 className="font-medium flex items-center gap-2">
              <IconUserPlus className="h-4 w-4" />
              Timeline
            </h4>
            <div className="space-y-3 pl-6">
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  Created At
                </Label>
                <div className="text-sm">{formatDate(request.created_at)}</div>
              </div>
              {request.updated_at &&
                request.updated_at !== request.created_at && (
                  <div className="flex flex-col gap-1">
                    <Label className="text-xs text-muted-foreground">
                      Updated At
                    </Label>
                    <div className="text-sm">
                      {formatDate(request.updated_at)}
                    </div>
                  </div>
                )}
            </div>
          </div>
        </div>

        <div className="flex justify-end mt-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

interface UserFriendRequestsDataTableProps {
  userId: string;
}

export function UserFriendRequestsDataTable({
  userId,
}: UserFriendRequestsDataTableProps) {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const pathname = useLocation().pathname;

  const viewId = searchParams.get("requestView");
  const [viewOpen, setViewOpen] = React.useState(false);
  const [viewRequest, setViewRequest] = React.useState<FriendRequest | null>(
    null,
  );

  const [direction, setDirection] = React.useState<"" | "sent" | "received">(
    "",
  );
  const [statusFilter, setStatusFilter] = React.useState<
    "" | "pending" | "accepted" | "rejected" | "cancelled" | "blocked"
  >("");
  const [currentPage, setCurrentPage] = React.useState(1);
  const [pageSize, setPageSize] = React.useState(20);

  const {
    requests: data,
    loading,
    refreshing,
    error,
    total,
    refresh,
  } = useUserFriendRequests(
    userId,
    direction,
    statusFilter,
    currentPage,
    pageSize,
  );

  const isMobile = useIsMobile();

  const totalPages = Math.ceil(total / pageSize) || 1;

  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = React.useState({});

  React.useEffect(() => {
    if (viewId) {
      const req = data.find((r) => r.id === viewId);
      if (req) {
        setViewRequest(req);
        setViewOpen(true);
      }
    }
  }, [viewId, data]);

  const handleView = React.useCallback(
    (request: FriendRequest) => {
      setViewRequest(request);
      setViewOpen(true);
      const params = new URLSearchParams(searchParams.toString());
      params.set("requestView", request.id);
      navigate(`${pathname}?${params.toString()}`);
    },
    [searchParams, pathname, navigate],
  );

  const activeFilterCount = [direction, statusFilter].filter(Boolean).length;

  const columns: ColumnDef<FriendRequest>[] = React.useMemo(
    () => [
      {
        accessorKey: "id",
        header: "ID",
        cell: ({ row }) => <ShortIdCell id={row.getValue("id")} />,
      },
      {
        accessorKey: "type",
        header: "Type",
        cell: ({ row }) => (
          <Badge variant="outline" className="capitalize">
            {row.getValue("type")}
          </Badge>
        ),
      },
      {
        accessorKey: "status",
        header: "Status",
        cell: ({ row }) => (
          <Badge
            variant="outline"
            className={STATUS_COLORS[row.getValue("status") as string] || ""}
          >
            {row.getValue("status")}
          </Badge>
        ),
      },
      {
        accessorKey: "from_user_id",
        header: "From",
        cell: ({ row }) => {
          const item = row.original;
          return (
            <UserIdLink
              id={item.from_user_id}
              username={item.from_username || ""}
              userId={userId}
            />
          );
        },
      },
      {
        accessorKey: "to_user_id",
        header: "To",
        cell: ({ row }) => {
          const item = row.original;
          return (
            <UserIdLink
              id={item.to_user_id}
              username={item.to_username || ""}
              userId={userId}
            />
          );
        },
      },
      {
        accessorKey: "created_at",
        header: "Created",
        cell: ({ row }) => (
          <span className="text-sm text-muted-foreground">
            {formatDistanceToNow(new Date(row.getValue("created_at")), {
              addSuffix: true,
              locale: enUS,
            })}
          </span>
        ),
      },
      {
        id: "actions",
        header: "Actions",
        enableHiding: false,
        cell: ({ row }) => {
          const request = row.original;
          return (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="h-8 w-8 p-0">
                  <IconDotsVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>Actions</DropdownMenuLabel>
                <DropdownMenuItem onClick={() => handleView(request)}>
                  <IconEye className="mr-2 h-4 w-4" />
                  View details
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => navigator.clipboard.writeText(request.id)}
                >
                  Copy ID
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          );
        },
      },
    ],
    [handleView, userId],
  );

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onSortingChange: setSorting,
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnVisibility,
      rowSelection,
    },
    manualPagination: true,
    pageCount: totalPages,
  });

  if (loading && !refreshing) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Friend Requests</CardTitle>
          <CardDescription>
            Manage incoming and outgoing friend requests
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4 py-4">
            <div className="flex items-center justify-between">
              <Skeleton className="h-9 w-64" />
              <Skeleton className="h-9 w-32" />
            </div>
            <div className="rounded-md border">
              <div className="space-y-3 p-4">
                {[0, 1, 2, 3, 4, 5, 6, 7].map((i) => (
                  <Skeleton key={`req-skel-${i}`} className="h-10 w-full" />
                ))}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Friend Requests</CardTitle>
          <CardDescription>
            Manage incoming and outgoing friend requests
          </CardDescription>
        </CardHeader>
        <CardContent>
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
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle>Friend Requests</CardTitle>
        <CardDescription>
          Manage incoming and outgoing friend requests
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
          <div className="flex flex-wrap gap-2 w-full lg:w-auto lg:ml-auto">
            <Button
              variant="outline"
              size="sm"
              onClick={() => refresh()}
              disabled={loading || refreshing}
            >
              <IconRefresh
                className={`h-4 w-4 ${loading || refreshing ? "animate-spin" : ""}`}
              />
            </Button>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <IconFilter className="h-4 w-4 mr-2" />
                  Filter
                  {activeFilterCount > 0 && (
                    <Badge
                      variant="secondary"
                      className="ml-2 h-5 px-1.5 text-xs"
                    >
                      {activeFilterCount}
                    </Badge>
                  )}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <div className="px-2 py-1.5 text-sm font-semibold">
                  Direction
                </div>
                {(["sent", "received"] as const).map((d) => (
                  <DropdownMenuItem
                    key={d}
                    onSelect={(e) => e.preventDefault()}
                  >
                    <Checkbox
                      checked={direction === d}
                      onCheckedChange={(checked) => {
                        setDirection(checked ? d : "");
                        setCurrentPage(1);
                      }}
                      className="mr-2"
                    />
                    {d === "sent" ? "Sent" : "Received"}
                  </DropdownMenuItem>
                ))}
                <DropdownMenuSeparator />
                <div className="px-2 py-1.5 text-sm font-semibold">Status</div>
                {(
                  [
                    "pending",
                    "accepted",
                    "rejected",
                    "cancelled",
                    "blocked",
                  ] as const
                ).map((s) => (
                  <DropdownMenuItem
                    key={s}
                    onSelect={(e) => e.preventDefault()}
                  >
                    <Checkbox
                      checked={statusFilter === s}
                      onCheckedChange={(checked) => {
                        setStatusFilter(checked ? s : "");
                        setCurrentPage(1);
                      }}
                      className="mr-2"
                    />
                    {s.charAt(0).toUpperCase() + s.slice(1)}
                  </DropdownMenuItem>
                ))}
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
                  .filter((c) => c.getCanHide())
                  .map((c) => (
                    <DropdownMenuCheckboxItem
                      key={c.id}
                      className="capitalize"
                      checked={c.getIsVisible()}
                      onCheckedChange={(v) => c.toggleVisibility(!!v)}
                    >
                      {c.id}
                    </DropdownMenuCheckboxItem>
                  ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {activeFilterCount > 0 && (
          <div className="flex flex-wrap gap-2 pb-4">
            {direction && (
              <Badge variant="secondary" className="gap-1">
                {direction === "sent" ? "Sent" : "Received"}
                <button
                  type="button"
                  className="ml-1 hover:text-foreground"
                  onClick={() => {
                    setDirection("");
                    setCurrentPage(1);
                  }}
                >
                  ×
                </button>
              </Badge>
            )}
            {statusFilter && (
              <Badge variant="secondary" className="gap-1 capitalize">
                {statusFilter}
                <button
                  type="button"
                  className="ml-1 hover:text-foreground"
                  onClick={() => {
                    setStatusFilter("");
                    setCurrentPage(1);
                  }}
                >
                  ×
                </button>
              </Badge>
            )}
          </div>
        )}

        {isMobile ? (
          <div className="space-y-3">
            {data.length ? (
              data.map((request) => {
                const isSent = request.from_user_id === userId;
                const otherUsername = isSent
                  ? request.to_username
                  : request.from_username;
                const otherId = isSent
                  ? request.to_user_id
                  : request.from_user_id;
                return (
                  <div
                    key={request.id}
                    className="rounded-lg border bg-card p-4 space-y-3"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Badge variant="outline" className="capitalize">
                          {isSent ? "sent" : "received"}
                        </Badge>
                        <Badge
                          variant="outline"
                          className={STATUS_COLORS[request.status] || ""}
                        >
                          {request.status}
                        </Badge>
                      </div>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" className="h-8 w-8 p-0">
                            <IconDotsVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleView(request)}>
                            <IconEye className="mr-2 h-4 w-4" />
                            View
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() =>
                              navigator.clipboard.writeText(request.id)
                            }
                          >
                            Copy ID
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {isSent ? "To: " : "From: "}
                      {otherUsername || otherId.slice(0, 8)}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {formatDistanceToNow(new Date(request.created_at), {
                        addSuffix: true,
                        locale: enUS,
                      })}
                    </div>
                  </div>
                );
              })
            ) : (
              <div className="rounded-lg border bg-card p-8 text-center text-muted-foreground">
                No friend requests found.
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
                {table.getRowModel().rows.length ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow
                      key={row.id}
                      data-state={row.getIsSelected() && "selected"}
                      className="cursor-pointer hover:bg-muted/50"
                      onClick={(e) => {
                        const target = e.target as HTMLElement;
                        if (
                          target.closest(
                            'button, a, [role="checkbox"], [role="menuitem"]',
                          )
                        )
                          return;
                        handleView(row.original);
                      }}
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
                      No friend requests found.
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
          pageSize={pageSize}
          totalCount={total}
          onPageChange={(page) => setCurrentPage(page + 1)}
          onPageSizeChange={setPageSize}
        />

        <FriendRequestViewDialog
          request={viewRequest}
          open={viewOpen}
          onOpenChange={(open) => {
            setViewOpen(open);
            if (!open) {
              setViewRequest(null);
              const params = new URLSearchParams(searchParams.toString());
              params.delete("requestView");
              navigate(
                params.toString()
                  ? `${pathname}?${params.toString()}`
                  : pathname,
              );
            }
          }}
        />
      </CardContent>
    </Card>
  );
}
