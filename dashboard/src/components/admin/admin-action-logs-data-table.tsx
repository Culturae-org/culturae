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
import { useAdminLogs } from "@/hooks/useAdminLogs";
import { useIsMobile } from "@/hooks/useMobile";
import {
  IconColumns,
  IconDotsVertical,
  IconEye,
  IconFilter,
  IconRefresh,
  IconSearch,
  IconX,
} from "@tabler/icons-react";
import type {
  ColumnDef,
  ColumnFiltersState,
  SortingState,
  VisibilityState,
} from "@tanstack/react-table";
import {
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import * as React from "react";
import { useNavigate } from "react-router";

export type { AdminActionLog } from "@/lib/types/logs.types";
import type { AdminActionLog } from "@/lib/types/logs.types";

const formatTimeAgo = (date: string) => {
  const now = new Date();
  const past = new Date(date);
  const diffInSeconds = Math.floor((now.getTime() - past.getTime()) / 1000);

  if (diffInSeconds < 60) return `${diffInSeconds}s ago`;
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
  return `${Math.floor(diffInSeconds / 86400)}d ago`;
};

interface AdminActionLogsDataTableProps {
  onViewRequest?: (log: AdminActionLog) => void;
}

export function AdminActionLogsDataTable({
  onViewRequest,
}: AdminActionLogsDataTableProps = {}) {
  const isMobile = useIsMobile();
  const navigate = useNavigate();
  const { fetchAdminLogs, loading } = useAdminLogs();
  const [data, setData] = React.useState<AdminActionLog[]>([]);
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    [],
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});

  const [currentPage, setCurrentPage] = React.useState(1);
  const [totalPages, setTotalPages] = React.useState(1);
  const [totalCount, setTotalCount] = React.useState(0);
  const [pageSize, setPageSize] = React.useState(50);

  const [actionFilter, setActionFilter] = React.useState<string>("");
  const [resourceFilter, setResourceFilter] = React.useState<string>("all");
  const [statusFilter, setStatusFilter] = React.useState<string>("all");
  const fetchLogs = React.useCallback(
    async (
      page: number,
      limit: number,
      actionType?: string,
      targetType?: string,
      userId?: string,
    ) => {
      try {
        const result = await fetchAdminLogs(
          page,
          limit,
          actionType,
          targetType,
          userId,
        );

        const logs = result.logs || [];
        setData(logs);
        setTotalCount(result.total_count);
        setTotalPages(result.total_pages);
        setCurrentPage(page);
      } catch (err) {
        console.error("Error fetching admin logs:", err);
      }
    },
    [fetchAdminLogs],
  );

  React.useEffect(() => {
    fetchLogs(1, pageSize, actionFilter, resourceFilter);
  }, [fetchLogs, pageSize, actionFilter, resourceFilter]);

  const columns: ColumnDef<AdminActionLog>[] = React.useMemo(
    () => [
      {
        accessorKey: "AdminName",
        header: "Admin",
        cell: ({ row }) => (
          <div className="font-medium">{row.getValue("AdminName")}</div>
        ),
      },
      {
        accessorKey: "Action",
        header: "Action",
        cell: ({ row }) => (
          <Badge variant="outline">{row.getValue("Action")}</Badge>
        ),
      },
      {
        accessorKey: "Resource",
        header: "Resource",
        cell: ({ row }) => (
          <div className="capitalize">{row.getValue("Resource")}</div>
        ),
      },
      {
        accessorKey: "Details",
        header: "Summary",
        cell: ({ row }) => {
          const details = row.original.Details as Record<string, unknown>;
          const summary =
            details?.summary || row.original.Action.replace(/_/g, " ");
          return (
            <div
              className="max-w-xs truncate text-sm text-muted-foreground"
              title={summary as string}
            >
              {summary as string}
            </div>
          );
        },
      },
      {
        accessorKey: "IsSuccess",
        header: "Status",
        cell: ({ row }) => {
          const isSuccess = row.getValue("IsSuccess") as boolean;
          return (
            <Badge
              variant={isSuccess ? "default" : "destructive"}
              className={isSuccess ? "bg-green-500" : ""}
            >
              {isSuccess ? "Success" : "Failed"}
            </Badge>
          );
        },
      },
      {
        accessorKey: "CreatedAt",
        header: "Time",
        cell: ({ row }) => (
          <div className="text-sm text-muted-foreground">
            {formatTimeAgo(row.getValue("CreatedAt"))}
          </div>
        ),
      },
      {
        id: "actions",
        header: "Actions",
        enableHiding: false,
        cell: ({ row }) => (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <span className="sr-only">Open menu</span>
                <IconDotsVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>Actions</DropdownMenuLabel>
              <DropdownMenuItem
                onClick={() => {
                  navigate(`?admin-view=${row.original.ID}`);
                  onViewRequest?.(row.original);
                }}
              >
                View
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        ),
      },
    ],
    [onViewRequest, navigate],
  );

  const table = useReactTable({
    data,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
    },
  });

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle>Admin Actions</CardTitle>
        <CardDescription>Log of all administrative actions</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
          <div className="relative flex-1 max-w-sm w-full lg:w-auto">
            <IconSearch className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search admin action..."
              value={actionFilter}
              onChange={(e) => setActionFilter(e.target.value)}
              className="w-60 pl-10 pr-10"
            />
            {actionFilter && (
              <IconX
                className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground cursor-pointer"
                onClick={() => setActionFilter("")}
              />
            )}
          </div>
          <div className="flex flex-wrap gap-2 ml-auto">
            <DropdownMenu>
              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  fetchLogs(currentPage, pageSize, actionFilter, resourceFilter)
                }
                disabled={loading}
              >
                <IconRefresh
                  className={`h-4 w-4 ${loading ? "animate-spin" : ""}`}
                />
              </Button>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <IconFilter className="h-4 w-4" />
                  Filter
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                <div className="px-2 py-1.5 text-sm font-semibold">
                  Resource Filter
                </div>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Checkbox
                    checked={resourceFilter === "user"}
                    onCheckedChange={(checked) => {
                      setResourceFilter(checked ? "user" : "all");
                      fetchLogs(
                        1,
                        pageSize,
                        actionFilter,
                        checked ? "user" : undefined,
                      );
                    }}
                    className="mr-2"
                  />
                  User
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Checkbox
                    checked={resourceFilter === "avatar"}
                    onCheckedChange={(checked) => {
                      setResourceFilter(checked ? "avatar" : "all");
                      fetchLogs(
                        1,
                        pageSize,
                        actionFilter,
                        checked ? "avatar" : undefined,
                      );
                    }}
                    className="mr-2"
                  />
                  Avatar
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Checkbox
                    checked={resourceFilter === "game"}
                    onCheckedChange={(checked) => {
                      setResourceFilter(checked ? "game" : "all");
                      fetchLogs(
                        1,
                        pageSize,
                        actionFilter,
                        checked ? "game" : undefined,
                      );
                    }}
                    className="mr-2"
                  />
                  Game
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Checkbox
                    checked={resourceFilter === "matchmaking"}
                    onCheckedChange={(checked) => {
                      setResourceFilter(checked ? "matchmaking" : "all");
                      fetchLogs(
                        1,
                        pageSize,
                        actionFilter,
                        checked ? "matchmaking" : undefined,
                      );
                    }}
                    className="mr-2"
                  />
                  Matchmaking
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Checkbox
                    checked={resourceFilter === "dataset"}
                    onCheckedChange={(checked) => {
                      setResourceFilter(checked ? "dataset" : "all");
                      fetchLogs(
                        1,
                        pageSize,
                        actionFilter,
                        checked ? "dataset" : undefined,
                      );
                    }}
                    className="mr-2"
                  />
                  Dataset
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Checkbox
                    checked={resourceFilter === "question"}
                    onCheckedChange={(checked) => {
                      setResourceFilter(checked ? "question" : "all");
                      fetchLogs(
                        1,
                        pageSize,
                        actionFilter,
                        checked ? "question" : undefined,
                      );
                    }}
                    className="mr-2"
                  />
                  Question
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Checkbox
                    checked={resourceFilter === "geography_dataset"}
                    onCheckedChange={(checked) => {
                      setResourceFilter(checked ? "geography_dataset" : "all");
                      fetchLogs(
                        1,
                        pageSize,
                        actionFilter,
                        checked ? "geography_dataset" : undefined,
                      );
                    }}
                    className="mr-2"
                  />
                  Geography
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Checkbox
                    checked={resourceFilter === "friendship"}
                    onCheckedChange={(checked) => {
                      setResourceFilter(checked ? "friendship" : "all");
                      fetchLogs(
                        1,
                        pageSize,
                        actionFilter,
                        checked ? "friendship" : undefined,
                      );
                    }}
                    className="mr-2"
                  />
                  Friendship
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Checkbox
                    checked={resourceFilter === "settings"}
                    onCheckedChange={(checked) => {
                      setResourceFilter(checked ? "settings" : "all");
                      fetchLogs(
                        1,
                        pageSize,
                        actionFilter,
                        checked ? "settings" : undefined,
                      );
                    }}
                    className="mr-2"
                  />
                  Settings
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <div className="px-2 py-1.5 text-sm font-semibold">
                  Status Filter
                </div>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Checkbox
                    checked={statusFilter === "success"}
                    onCheckedChange={(checked) => {
                      setStatusFilter(checked ? "success" : "all");
                    }}
                    className="mr-2"
                  />
                  Success
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Checkbox
                    checked={statusFilter === "failed"}
                    onCheckedChange={(checked) => {
                      setStatusFilter(checked ? "failed" : "all");
                    }}
                    className="mr-2"
                  />
                  Failed
                </DropdownMenuItem>
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
                  .map((column) => {
                    return (
                      <DropdownMenuCheckboxItem
                        key={column.id}
                        className="capitalize"
                        checked={column.getIsVisible()}
                        onCheckedChange={(value) =>
                          column.toggleVisibility(!!value)
                        }
                      >
                        {column.id.replace("_", " ")}
                      </DropdownMenuCheckboxItem>
                    );
                  })}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {isMobile ? (
          <div className="space-y-3">
            {data.length ? (
              data.map((log) => (
                <div
                  key={log.ID}
                  className="rounded-lg border bg-card p-4 space-y-3"
                >
                  <div className="flex items-center justify-between">
                    <span className="font-medium text-sm">{log.AdminName}</span>
                    <div className="flex items-center gap-2">
                      <Badge
                        variant={log.IsSuccess ? "default" : "destructive"}
                        className={log.IsSuccess ? "bg-green-500" : ""}
                      >
                        {log.IsSuccess ? "OK" : "Fail"}
                      </Badge>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          navigate(`?admin-view=${log.ID}`);
                          onViewRequest?.(log);
                        }}
                      >
                        <IconEye className="h-4 w-4 mr-2" />
                        View
                      </Button>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline">{log.Action}</Badge>
                    <span className="capitalize text-sm text-muted-foreground">
                      {log.Resource}
                    </span>
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {formatTimeAgo(log.CreatedAt)}
                  </div>
                </div>
              ))
            ) : (
              <div className="rounded-lg border bg-card p-8 text-center text-muted-foreground">
                {loading ? (
                  <div className="space-y-2">
                    {[1, 2, 3, 4, 5].map((n) => (
                      <Skeleton
                        key={`skeleton-mobile-${n}`}
                        className="h-4 w-full"
                      />
                    ))}
                  </div>
                ) : (
                  "No results."
                )}
              </div>
            )}
          </div>
        ) : (
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => {
                      return (
                        <TableHead key={header.id}>
                          {header.isPlaceholder
                            ? null
                            : flexRender(
                                header.column.columnDef.header,
                                header.getContext(),
                              )}
                        </TableHead>
                      );
                    })}
                  </TableRow>
                ))}
              </TableHeader>
              <TableBody>
                {table.getRowModel().rows?.length ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow
                      key={row.id}
                      data-state={row.getIsSelected() && "selected"}
                      onClick={() => {
                        navigate(`?admin-view=${row.original.ID}`);
                        onViewRequest?.(row.original);
                      }}
                      className="cursor-pointer"
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
                      {loading ? (
                        <div className="space-y-2">
                          {[1, 2, 3, 4, 5].map((n) => (
                            <Skeleton
                              key={`skeleton-table-${n}`}
                              className="h-4 w-full"
                            />
                          ))}
                        </div>
                      ) : (
                        "No results."
                      )}
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
          totalCount={totalCount}
          onPageChange={(newPage) =>
            fetchLogs(newPage + 1, pageSize, actionFilter, resourceFilter)
          }
          onPageSizeChange={(newSize) => {
            setPageSize(newSize);
            setCurrentPage(1);
          }}
          loading={loading}
        />
      </CardContent>
    </Card>
  );
}
