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
import { useUserActionLogs } from "@/hooks/useLogs";
import { useIsMobile } from "@/hooks/useMobile";
import type { UserActionLog } from "@/lib/types/logs.types";
import {
  IconAlertTriangle,
  IconCalendar,
  IconColumns,
  IconDeviceDesktop,
  IconDotsVertical,
  IconEye,
  IconFilter,
  IconRefresh,
  IconShield,
} from "@tabler/icons-react";
import type {
  ColumnDef,
  SortingState,
  VisibilityState,
} from "@tanstack/react-table";
import {
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
import { UserActionLogViewDialog } from "./user-action-log-view-dialog";

interface UserActionLogsDataTableProps {
  userId: string;
}

function LogIdCell({ id }: { id: string | undefined }) {
  const [isExpanded, setIsExpanded] = React.useState(false);

  if (!id) return <span className="text-muted-foreground">-</span>;

  return (
    <button
      type="button"
      className="font-mono text-xs text-muted-foreground cursor-pointer hover:text-foreground transition-colors text-left"
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

function StatusCell({ isSuccess }: { isSuccess: boolean | undefined }) {
  if (isSuccess === undefined) return <Badge variant="outline">Unknown</Badge>;
  return (
    <Badge variant={isSuccess ? "default" : "destructive"}>
      {isSuccess ? "Success" : "Failed"}
    </Badge>
  );
}

function ActionCell({ action }: { action: string | undefined }) {
  if (!action) return <Badge variant="outline">Unknown</Badge>;

  const actionColors: Record<string, string> = {
    profile_update: "bg-blue-100 text-blue-800",
    password_change: "bg-orange-100 text-orange-800",
    avatar_upload: "bg-purple-100 text-purple-800",
    avatar_delete: "bg-purple-100 text-purple-800",
    id_regenerate: "bg-yellow-100 text-yellow-800",
    account_delete: "bg-red-100 text-red-800",
    login: "bg-green-100 text-green-800",
    logout: "bg-gray-100 text-gray-800",
  };

  return (
    <Badge
      className={`${actionColors[action] || "bg-gray-100 text-gray-800"} border-0`}
    >
      {action.replace(/_/g, " ")}
    </Badge>
  );
}

function IpCell({ ip }: { ip: string | undefined }) {
  if (!ip) return <span className="text-muted-foreground">-</span>;
  const cleanIp = ip.replace(/^::ffff:/, "");
  return (
    <div className="flex items-center gap-1">
      <span className="text-sm font-mono">{cleanIp}</span>
    </div>
  );
}

function DetailsCell({
  details,
}: { details: Record<string, unknown> | undefined }) {
  const [isExpanded, setIsExpanded] = React.useState(false);

  if (!details) return <span className="text-muted-foreground">-</span>;

  return (
    <button
      type="button"
      className="text-xs text-muted-foreground cursor-pointer hover:text-foreground transition-colors font-mono text-left"
      onClick={(e) => {
        e.stopPropagation();
        setIsExpanded(!isExpanded);
      }}
      title={isExpanded ? "Click to collapse" : "Click to expand"}
    >
      {isExpanded
        ? JSON.stringify(details, null, 2)
        : `${JSON.stringify(details).slice(0, 50)}...`}
    </button>
  );
}

export function UserActionLogsDataTable({
  userId,
}: UserActionLogsDataTableProps) {
  const { logs, loading, refreshing, error, refresh } =
    useUserActionLogs(userId);
  const isMobile = useIsMobile();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const pathname = useLocation().pathname;

  const [actionFilter, setActionFilter] = React.useState<string>("all");
  const [categoryFilter, setCategoryFilter] = React.useState<string>("all");
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({
      resource_id: false,
      details: false,
      user_agent: false,
    });
  const [rowSelection, setRowSelection] = React.useState({});
  const [currentPage, setCurrentPage] = React.useState(1);
  const [currentLimit, setCurrentLimit] = React.useState(10);
  const [selectedLog, setSelectedLog] = React.useState<UserActionLog | null>(
    null,
  );
  const [viewOpen, setViewOpen] = React.useState(false);

  const viewId = searchParams.get("logView");
  React.useEffect(() => {
    if (viewId && logs.length > 0) {
      const found = logs.find((l) => l.ID === viewId);
      if (found) {
        setSelectedLog(found);
        setViewOpen(true);
      }
    }
  }, [viewId, logs]);

  const uniqueActions = React.useMemo(() => {
    const actions = new Set(logs.map((log) => log.Action).filter(Boolean));
    return Array.from(actions);
  }, [logs]);

  const filteredData = React.useMemo(() => {
    let data = logs;
    if (actionFilter !== "all")
      data = data.filter((l) => l.Action === actionFilter);
    if (categoryFilter !== "all")
      data = data.filter(
        (l) =>
          (l as unknown as { Category?: string }).Category === categoryFilter,
      );
    return data;
  }, [logs, actionFilter, categoryFilter]);

  const columns: ColumnDef<UserActionLog>[] = React.useMemo(
    () => [
      {
        id: "select",
        header: ({ table }) => (
          <Checkbox
            checked={
              table.getIsAllPageRowsSelected() ||
              (table.getIsSomePageRowsSelected() && "indeterminate")
            }
            onCheckedChange={(value) =>
              table.toggleAllPageRowsSelected(!!value)
            }
            aria-label="Select all"
          />
        ),
        cell: ({ row }) => (
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label="Select row"
          />
        ),
        enableSorting: false,
        enableHiding: false,
      },
      {
        accessorKey: "ID",
        header: "ID",
        cell: ({ row }) => <LogIdCell id={row.getValue("ID")} />,
      },
      {
        accessorKey: "Action",
        header: "Action",
        cell: ({ row }) => <ActionCell action={row.getValue("Action")} />,
      },
      {
        accessorKey: "Resource",
        header: "Resource",
        cell: ({ row }) => {
          const resource = row.getValue("Resource") as string;
          return (
            <div className="flex items-center gap-1 text-sm">
              {resource}
            </div>
          );
        },
      },
      {
        accessorKey: "ResourceID",
        header: "Resource ID",
        cell: ({ row }) => {
          const resourceId = row.getValue("ResourceID") as string | undefined;
          return resourceId ? (
            <LogIdCell id={resourceId} />
          ) : (
            <span className="text-muted-foreground">-</span>
          );
        },
      },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => <StatusCell isSuccess={row.original.IsSuccess} />,
      },
      {
        accessorKey: "IPAddress",
        header: "IP Address",
        cell: ({ row }) => <IpCell ip={row.getValue("IPAddress")} />,
      },
      {
        accessorKey: "UserAgent",
        header: "User Agent",
        cell: ({ row }) => {
          const userAgent = row.getValue("UserAgent") as string;
          return (
            <div className="flex items-center gap-1 text-sm truncate max-w-[200px]">
              <span className="truncate" title={userAgent}>
                {userAgent || "Unknown"}
              </span>
            </div>
          );
        },
      },
      {
        accessorKey: "Details",
        header: "Details",
        cell: ({ row }) => <DetailsCell details={row.getValue("Details")} />,
      },
      {
        accessorKey: "ErrorMsg",
        header: "Error",
        cell: ({ row }) => {
          const errorMsg = row.getValue("ErrorMsg") as string | undefined;
          return errorMsg ? (
            <div
              className="text-sm text-red-600 truncate max-w-[200px]"
              title={errorMsg}
            >
              {errorMsg}
            </div>
          ) : (
            <span className="text-muted-foreground text-sm">-</span>
          );
        },
      },
      {
        accessorKey: "CreatedAt",
        header: "Date",
        cell: ({ row }) => {
          const dateValue = row.getValue("CreatedAt") as string;
          if (
            !dateValue ||
            dateValue === "0001-01-01T00:00:00Z" ||
            Number.isNaN(new Date(dateValue).getTime())
          ) {
            return (
              <div className="flex items-center gap-1 text-sm text-muted-foreground">
                <span>-</span>
              </div>
            );
          }
          const date = new Date(dateValue);
          return (
            <div className="flex items-center gap-1 text-sm">
              <span title={format(date, "PPpp", { locale: enUS })}>
                {formatDistanceToNow(date, { addSuffix: true, locale: enUS })}
              </span>
            </div>
          );
        },
      },
      {
        id: "actions",
        header: "Actions",
        enableHiding: false,
        cell: ({ row }) => {
          const log = row.original;
          return (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  className="h-8 w-8 p-0"
                  onClick={(e) => e.stopPropagation()}
                >
                  <span className="sr-only">Open menu</span>
                  <IconDotsVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>Actions</DropdownMenuLabel>
                <DropdownMenuItem
                  onSelect={() => {
                    setSelectedLog(log);
                    setViewOpen(true);
                    const params = new URLSearchParams(searchParams.toString());
                    params.set("logView", log.ID);
                    navigate(`${pathname}?${params.toString()}`);
                  }}
                >
                  <IconEye className="mr-2 h-4 w-4" />
                  View Details
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={(e) => {
                    e.stopPropagation();
                    navigator.clipboard.writeText(log.ID);
                  }}
                >
                  Copy Log ID
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={(e) => {
                    e.stopPropagation();
                    log.IPAddress &&
                      navigator.clipboard.writeText(log.IPAddress);
                  }}
                >
                  Copy IP Address
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          );
        },
      },
    ],
    [pathname, searchParams, navigate],
  );

  const table = useReactTable({
    data: filteredData,
    columns,
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    pageCount: Math.ceil(filteredData.length / currentLimit),
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

  if (loading && !refreshing && logs.length === 0) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>User Action Logs</CardTitle>
          <CardDescription>
            View and manage action logs for this user
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
                  <Skeleton key={`action-skel-${i}`} className="h-10 w-full" />
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
          <CardTitle>User Action Logs</CardTitle>
          <CardDescription>
            View and manage action logs for this user
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
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
        <CardTitle>User Action Logs</CardTitle>
        <CardDescription>
          View and manage action logs for this user
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="relative">
          <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
            <div className="flex flex-wrap gap-2 w-full lg:w-auto">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" size="sm">
                    <IconFilter className="h-4 w-4" />
                    Filter
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <div className="px-2 py-1.5 text-sm font-semibold">
                    Category Filter
                  </div>
                  {["profile", "avatar", "auth", "game", "friends"].map(
                    (cat) => (
                      <DropdownMenuItem
                        key={cat}
                        onSelect={(e) => e.preventDefault()}
                      >
                        <Checkbox
                          checked={categoryFilter === cat}
                          onCheckedChange={(checked) =>
                            setCategoryFilter(checked ? cat : "all")
                          }
                          className="mr-2"
                        />
                        <span className="capitalize">{cat}</span>
                      </DropdownMenuItem>
                    ),
                  )}
                  <DropdownMenuSeparator />
                  <div className="px-2 py-1.5 text-sm font-semibold">
                    Action Filter
                  </div>
                  <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                    <Checkbox
                      checked={actionFilter === "all"}
                      onCheckedChange={() => setActionFilter("all")}
                      className="mr-2"
                    />
                    All Actions
                  </DropdownMenuItem>
                  {uniqueActions.map((action) => (
                    <DropdownMenuItem
                      key={action}
                      onSelect={(e) => e.preventDefault()}
                    >
                      <Checkbox
                        checked={actionFilter === action}
                        onCheckedChange={(checked) =>
                          setActionFilter(checked ? action : "all")
                        }
                        className="mr-2"
                      />
                      {action.replace(/_/g, " ")}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
              {(categoryFilter !== "all" || actionFilter !== "all") && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setCategoryFilter("all");
                    setActionFilter("all");
                  }}
                >
                  Clear Filters
                </Button>
              )}
            </div>
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
                const log = row.original;
                return (
                  <div
                    key={row.id}
                    className="rounded-lg border bg-card p-4 space-y-3"
                  >
                    <div className="flex items-center justify-between">
                      <ActionCell action={log.Action} />
                      <StatusCell isSuccess={log.IsSuccess} />
                    </div>
                    <div className="text-sm">
                      <span className="text-muted-foreground">Resource: </span>
                      {log.Resource}
                    </div>
                    {log.IPAddress && (
                      <div className="text-sm font-mono">
                        {log.IPAddress.replace(/^::ffff:/, "")}
                      </div>
                    )}
                    <span className="text-xs text-muted-foreground">
                      {formatDistanceToNow(new Date(log.CreatedAt), {
                        addSuffix: true,
                        locale: enUS,
                      })}
                    </span>
                  </div>
                );
              })
            ) : (
              <div className="rounded-lg border bg-card p-8 text-center text-muted-foreground">
                No action logs found.
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
                      className="cursor-pointer"
                      onClick={() => {
                        setSelectedLog(row.original);
                        setViewOpen(true);
                        const params = new URLSearchParams(
                          searchParams.toString(),
                        );
                        params.set("logView", row.original.ID);
                        navigate(`${pathname}?${params.toString()}`);
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
                      No action logs found.
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
          totalCount={filteredData.length}
          onPageChange={(page) => table.setPageIndex(page)}
          onPageSizeChange={(newSize) => {
            setCurrentLimit(newSize);
            setCurrentPage(1);
          }}
        />

        {selectedLog && (
          <UserActionLogViewDialog
            log={selectedLog}
            open={viewOpen}
            onOpenChange={(open) => {
              setViewOpen(open);
              if (!open) {
                const params = new URLSearchParams(searchParams.toString());
                params.delete("logView");
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
